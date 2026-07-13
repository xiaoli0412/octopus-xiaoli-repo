# Octopus 企业级部署 Runbook

> 本 Runbook 覆盖 Octopus(LLM API 聚合与负载均衡服务)的部署、扩缩容、配置、故障排查、回滚与监控告警。
>
> 所有操作步骤均以可执行为原则,如遇与本文档不一致的实际行为,以代码与 `AGENTS.md` 为准并回头修订本文档。
>
> 关键参考:
> - 项目文档: `AGENTS.md`
> - 配置定义: `internal/conf/config.go`
> - 指标定义: `internal/observability/metrics.go`
> - 服务器路由: `internal/server/server.go`
> - 安装脚本: `scripts/install.sh`
> - 发布流程: `.github/workflows/release.yaml`

---

## 目录

1. [部署](#1-部署)
    - 1.1 [Docker 部署](#11-docker-部署)
    - 1.2 [二进制部署](#12-二进制部署linux-服务器)
    - 1.3 [Kubernetes 部署](#13-kubernetes-部署)
    - 1.4 [健康检查配置](#14-健康检查配置)
2. [扩缩容](#2-扩缩容)
    - 2.1 [水平扩容](#21-水平扩容)
    - 2.2 [垂直扩容](#22-垂直扩容)
    - 2.3 [渠道/分组扩容](#23-渠道分组扩容)
    - 2.4 [数据库扩容与迁移](#24-数据库扩容与迁移)
3. [配置](#3-配置)
    - 3.1 [环境变量完整列表](#31-环境变量完整列表)
    - 3.2 [配置文件](#32-配置文件)
    - 3.3 [密钥管理](#33-密钥管理)
    - 3.4 [可观测性配置](#34-可观测性配置)
4. [故障排查](#4-故障排查)
    - 4.1 [数据库连接失败](#41-数据库连接失败)
    - 4.2 [渠道全部熔断](#42-渠道全部熔断)
    - 4.3 [OOM(内存溢出)](#43-oom内存溢出)
    - 4.4 [内存泄漏](#44-内存泄漏)
    - 4.5 [SSE 连接泄漏](#45-sse-连接泄漏)
    - 4.6 [上游同步失败](#46-上游同步失败)
5. [回滚](#5-回滚)
    - 5.1 [版本回滚](#51-版本回滚)
    - 5.2 [数据库回滚](#52-数据库回滚)
    - 5.3 [配置回滚](#53-配置回滚)
6. [监控告警](#6-监控告警)
    - 6.1 [Prometheus 指标列表](#61-prometheus-指标列表)
    - 6.2 [Grafana 仪表板建议](#62-grafana-仪表板建议)
    - 6.3 [告警规则](#63-告警规则prometheus-alertmanager)

---

## 1. 部署

### 1.1 Docker 部署

Octopus 官方镜像托管在 GitHub Container Registry(GHCR),支持 `linux/amd64`、`linux/386`、`linux/arm64`、`linux/arm/v7` 多架构。提供两种 Dockerfile flavor:
- **debian**(默认,tag 后缀空):`ghcr.io/xiaoli0412/octopus-xiaoli-repo:latest`
- **alpine**(更小,tag 后缀 `-alpine`):`ghcr.io/xiaoli0412/octopus-xiaoli-repo:latest-alpine`

#### 1.1.1 docker-compose(推荐)

仓库根目录提供 `docker-compose.yml`:

```bash
cd /path/to/octopus_repo
docker compose up -d
```

覆盖默认配置:

```bash
OCTOPUS_PORT=1088 \
OCTOPUS_DATA_DIR=/data/octopus \
OCTOPUS_CONTAINER_NAME=octopus \
OCTOPUS_IMAGE=ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.20.0 \
docker compose up -d
```

`docker-compose.yml` 关键安全加固(已内置,无需手动配置):
- `read_only: true`(根文件系统只读)
- `tmpfs: /tmp:rw,noexec,nosuid,size=64m`
- `security_opt: no-new-privileges:true`
- `cap_drop: ALL` + 仅保留 `CHOWN/SETGID/SETUID`
- 健康检查:`/app/octopus healthcheck --url http://127.0.0.1:1088/healthz`

验证部署:

```bash
docker compose ps
curl -fsS http://127.0.0.1:1088/healthz   # 期望 {"status":"ok"}
curl -fsS http://127.0.0.1:1088/readyz     # 期望 200,失败返回 503
curl -fsS http://127.0.0.1:1088/metrics | head -n 20
```

#### 1.1.2 docker run

```bash
docker run -d \
  --name octopus \
  --restart unless-stopped \
  -p 1088:1088 \
  -e OCTOPUS_SERVER_HOST=0.0.0.0 \
  -e OCTOPUS_SERVER_PORT=1088 \
  -e OCTOPUS_DATABASE_TYPE=sqlite \
  -e OCTOPUS_DATABASE_PATH=/app/data/data.db \
  -e OCTOPUS_LOG_LEVEL=info \
  -e DATA_DIR=/app/data \
  -e PUID=10001 \
  -e PGID=10001 \
  -v /data/octopus:/app/data \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m \
  --security-opt no-new-privileges:true \
  --cap-drop ALL \
  --cap-add CHOWN \
  --cap-add SETGID \
  --cap-add SETUID \
  ghcr.io/xiaoli0412/octopus-xiaoli-repo:latest
```

> **注意**:Docker Hub(`docker.io/xiaoli0412/octopus-xiaoli-repo`)已废弃,安装脚本会拒绝该来源。仅接受 GHCR 官方镜像或显式指定的私有/镜像仓库地址。

### 1.2 二进制部署(Linux 服务器)

仓库提供 `scripts/install.sh`,支持一键安装(优先 GHCR 拉取,失败回退源码 Docker 构建,最终可由本地二进制兜底):

```bash
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | bash
```

自定义端口:

```bash
curl -fsSL https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh | OCTOPUS_PORT=1088 bash
```

`raw.githubusercontent.com` 不可达时,两步执行:

```bash
curl -fsSL -o install-octopus.sh https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/main/scripts/install.sh
bash install-octopus.sh
```

GHCR 拉取失败且源码 Docker 构建仍受阻时,使用已知可用二进制兜底:

```bash
OCTOPUS_BINARY_PATH=/root/octopus-linux-amd64 bash install-octopus.sh
```

安装脚本会:
1. 检查 Docker 守护进程可用性
2. 解析外部端口(默认 1088,被占用时交互式或自动选择回退端口 `18080/18086/28080`)
3. 复用既有数据目录(检测容器挂载或 `/root/octopus-data/data.db`)
4. 拉取镜像 → 启动容器 → 验证存活
5. 输出 UI 地址、容器名、数据目录、镜像信息

### 1.3 Kubernetes 部署

> **完整部署清单**已提取至 [`deploy/k8s/`](../deploy/k8s/) 目录,包含 namespace、configmap、secret、pvc、deployment、service、ingress 七个独立 YAML 文件。
>
> 详细部署步骤、配置说明、升级与卸载流程请参考 [`deploy/k8s/README.md`](../deploy/k8s/README.md)。

**关键要点**:

- **namespace**:独立 `octopus` namespace,带 `app.kubernetes.io/name` 标签
- **副本数**:SQLite 模式下必须为 **1**(ReadWriteOnce PVC);多副本需切换 MySQL/PostgreSQL(见 [2.4](#24-数据库扩容与迁移))
- **探针**:`livenessProbe` → `/healthz`(进程存活),`readinessProbe` → `/readyz`(依赖就绪,未就绪自动摘除流量)
- **资源配额**:默认 `requests: { cpu: 100m, memory: 128Mi }`,`limits: { cpu: 1000m, memory: 512Mi }`;生产建议调大
- **安全加固**:`runAsNonRoot`、`readOnlyRootFilesystem`、`cap_drop: ALL`、`seccompProfile: RuntimeDefault`
- **存储**:PVC `octopus-data`(10Gi,ReadWriteOnce)挂载 `/app/data`;`emptyDir`(Memory,64Mi)挂载 `/tmp`
- **Ingress**:Nginx Ingress,`proxy-body-size: 10m` 支持大请求体,`proxy-buffering: off` 支持 SSE 长连接

以下为各清单的快速参考(完整文件见 `deploy/k8s/`):

#### 1.3.1 Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: octopus-secret
  namespace: octopus
type: Opaque
stringData:
  # 数据库连接串(以 MySQL 为例,生产推荐外部托管 MySQL)
  DATABASE_DSN: "octopus:CHANGE_ME@tcp(mysql.octopus.svc:3306)/octopus?charset=utf8mb4&parseTime=True&loc=Local"
  # OpenTelemetry OTLP 端点(可选)
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://otel-collector.observability.svc:4318"
---
apiVersion: v1
kind: Secret
metadata:
  name: octopus-admin-secret
  namespace: octopus
type: Opaque
# 用于镜像拉取(如使用私有镜像仓库)
data:
  .dockerconfigjson: CHANGE_ME_BASE64
```

#### 1.3.2 ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: octopus-config
  namespace: octopus
data:
  OCTOPUS_SERVER_HOST: "0.0.0.0"
  OCTOPUS_SERVER_PORT: "1088"
  OCTOPUS_DATABASE_TYPE: "mysql"        # 生产推荐 mysql / postgresql
  OCTOPUS_LOG_LEVEL: "info"
  OCTOPUS_SHUTDOWN_TIMEOUT: "30s"
  OCTOPUS_SHUTDOWN_TASK_TIMEOUT: "10s"
  OCTOPUS_HTTP_MAX_IDLE_CONNS: "200"
  OCTOPUS_HTTP_MAX_CONNS_PER_HOST: "100"
  OCTOPUS_HTTP_MAX_IDLE_CONNS_PER_HOST: "100"
  OCTOPUS_HTTP_IDLE_CONN_TIMEOUT: "90s"
  OCTOPUS_HTTP_TLS_HANDSHAKE_TIMEOUT: "10s"
  OCTOPUS_HTTP_EXPECT_CONTINUE_TIMEOUT: "1s"
  OCTOPUS_HTTP_RESPONSE_HEADER_TIMEOUT: "30s"
```

#### 1.3.3 Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: octopus
  namespace: octopus
  labels:
    app: octopus
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: octopus
  template:
    metadata:
      labels:
        app: octopus
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "1088"
        prometheus.io/path: "/metrics"
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 10001
        runAsGroup: 10001
        fsGroup: 10001
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: octopus
          image: ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.25.0
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: 1088
              protocol: TCP
          envFrom:
            - configMapRef:
                name: octopus-config
            - secretRef:
                name: octopus-secret
          env:
            - name: DATA_DIR
              value: /app/data
            - name: PUID
              value: "10001"
            - name: PGID
              value: "10001"
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "2000m"
              memory: "2Gi"
          # 注意: read-only 文件系统需 tmpfs 给 /tmp
          securityContext:
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
              add: ["CHOWN", "SETGID", "SETUID"]
          volumeMounts:
            - name: tmp
              mountPath: /tmp
            - name: data
              mountPath: /app/data
          startupProbe:
            httpGet:
              path: /healthz
              port: http
            failureThreshold: 30
            periodSeconds: 5
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /readyz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
      volumes:
        - name: tmp
          emptyDir:
            medium: Memory
            sizeLimit: 64Mi
        - name: data
          persistentVolumeClaim:
            claimName: octopus-data
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: octopus-data
  namespace: octopus
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  # storageClassName: <your-storage-class>
```

> **SQLite 注意**:使用 `ReadWriteOnce` PVC 时,副本数必须为 1。多副本场景请改用 MySQL/PostgreSQL(见 [2.4](#24-数据库扩容与迁移))。

#### 1.3.4 Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: octopus
  namespace: octopus
  labels:
    app: octopus
spec:
  type: ClusterIP
  selector:
    app: octopus
  ports:
    - name: http
      port: 1088
      targetPort: http
      protocol: TCP
---
# 可选:Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: octopus
  namespace: octopus
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "300"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "300"
    # SSE 长连接需关闭 buffering
    nginx.ingress.kubernetes.io/proxy-buffering: "off"
spec:
  ingressClassName: nginx
  rules:
    - host: octopus.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: octopus
                port:
                  name: http
```

> **完整部署流程**(含 Secret 创建、一键部署、验证、升级、卸载)请参考 [`deploy/k8s/README.md`](../deploy/k8s/README.md)。

#### 1.3.5 ServiceMonitor(Prometheus Operator)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: octopus
  namespace: octopus
  labels:
    release: prometheus
spec:
  selector:
    matchLabels:
      app: octopus
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
      scrapeTimeout: 10s
```

### 1.4 健康检查配置

Octopus 暴露三个端点(定义见 `internal/server/server.go`):

| 端点 | 用途 | 成功响应 | 失败响应 |
|------|------|----------|----------|
| `GET /healthz` | liveness(进程存活) | `200 {"status":"ok"}` | 不主动返回 5xx,进程挂掉时探针失败 |
| `GET /readyz` | readiness(依赖就绪) | `200` + 原因列表 | `503` + 未就绪原因(DB/缓存/后台任务) |
| `GET /metrics` | Prometheus 抓取 | Prometheus 文本格式 | — |

**Docker / docker-compose** 已内置:

```yaml
healthcheck:
  test: ["CMD", "/app/octopus", "healthcheck", "--url", "http://127.0.0.1:1088/healthz"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 20s
```

**Kubernetes** 探针配置见 [1.3.3](#133-deployment),关键点:
- `livenessProbe` 用 `/healthz`(只看进程存活,避免误杀)
- `readinessProbe` 用 `/readyz`(检查 DB/缓存,未就绪时从 Service endpoints 摘除)
- `startupProbe` 用 `/healthz`(给启动慢的实例宽限期,避免 liveness 在初始化阶段误判)

> **注意**:不要把 `livenessProbe` 指向 `/readyz`。`/readyz` 在依赖故障时返回 503,若用于 liveness 会触发循环重启,无法自愈。

---

## 2. 扩缩容

### 2.1 水平扩容

**适用场景**:RPS 接近上限、CPU 使用率持续 > 70%、P95 延迟上升。

**操作步骤**:

1. **确认数据库类型**:水平扩容要求多实例共享状态,SQLite 不支持。先按 [2.4](#24-数据库扩容与迁移) 迁移到 MySQL/PostgreSQL。
2. **增加副本数**:
   - k8s:`kubectl scale deployment octopus -n octopus --replicas=5`
   - docker-compose:启动多个 compose service,前置 Nginx/HAProxy 做 LB
3. **验证**:`curl` 各实例 `/readyz` 均 200,通过 LB 访问 `/v1/chat/completions` 正常。

**注意事项**:

- **Session 亲和性**:Octopus 默认无状态(状态写共享 DB),可放心轮询。但若启用了会话亲和性(粘性会话保持渠道选择),LB 层需配置 sticky session 或在 Octopus 内将会话状态外置。检查方式:在管理界面查看渠道设置中的「会话亲和性」开关。
- **缓存共享**:Octopus 的内存缓存(channel/group/model 缓存)在每个实例独立。修改配置后需等待所有实例的缓存 TTL 过期或手动重启。生产建议通过管理 API 触发缓存刷新,而非依赖重启。
- **SSE 推送**:`/api/v1/stream/stats` 的 SSE 连接是实例本地的,客户端连到哪个实例就只能看该实例的实时统计。跨实例聚合需通过共享 DB 查询 `/api/v1/stats/total`。
- **资源配额**:扩容前确认节点资源、IP 池、LB 连接数上限。

### 2.2 垂直扩容

**适用场景**:单实例 CPU/内存瓶颈,且暂时不便水平扩容(如 SQLite 单机)。

**操作步骤**:

1. 调整资源限制:
   - k8s:修改 `resources.limits` 与 `resources.requests`
   - docker-compose:无显式限制时,受宿主机资源约束
2. 调整出站 HTTP 连接池(见 [3.1](#31-环境变量完整列表)):
   ```bash
   OCTOPUS_HTTP_MAX_IDLE_CONNS=500
   OCTOPUS_HTTP_MAX_CONNS_PER_HOST=200
   ```
3. 滚动重启使配置生效。
4. 压测验证(`load-test/run.sh`),确认 P95 下降且无 OOM。

**经验值**:
- 单实例 2C4G + SQLite + mock 上游,可承载 ~500 QPS(chat-completions)
- 单实例 4C8G + MySQL + mock 上游,可承载 ~1500 QPS
- 超过单实例上限应转为水平扩容

### 2.3 渠道/分组扩容

**适用场景**:上游限流、单渠道配额不足、需要更高上游吞吐。

**操作步骤**(通过管理 API 或 Web UI):

1. **新增渠道**:`POST /api/v1/channel`,配置 `base_url`、`api_key`(支持多 key 轮换)、`provider_type`。
2. **同步模型**:`POST /api/v1/channel/{id}/sync`,自动拉取上游模型列表。
3. **加入分组**:`POST /api/v1/group/{id}/items`,设置 `priority` 与 `weight`。
4. **选择负载均衡策略**(在 group 配置中):
    - `round_robin`:按顺序轮询
    - `random`:随机
    - `failover`:按 priority 故障转移(主备场景)
    - `weighted`:按 weight 比例分配(差异化配额场景)
5. **验证**:`GET /api/v1/stats/channel` 查看各渠道请求分布是否符合预期。

**注意事项**:
- 新渠道接入后,熔断器初始为 closed,首个失败请求会开始计数。建议先用低权重接入观察 10 分钟。
- 多 key 轮换时,每个 key 独立追踪成本与限流状态,单 key 限流不会拖垮整个渠道。

### 2.4 数据库扩容与迁移

#### 2.4.1 SQLite → MySQL/PostgreSQL 迁移

> **前提**:Octopus 使用 GORM AutoMigrate,新表会自动创建。但 **GORM AutoMigrate 不自动回退**,迁移前必须备份。

**迁移步骤**:

1. **备份 SQLite 数据**(见 [5.2](#52-数据库回滚)):
   ```bash
   curl -H "Authorization: Bearer $JWT" http://localhost:1088/api/v1/backup/download -o backup-$(date +%Y%m%d).tar.gz
   ```

2. **准备目标数据库**:
   ```sql
   -- MySQL
   CREATE DATABASE octopus CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   CREATE USER 'octopus'@'%' IDENTIFIED BY 'CHANGE_ME';
   GRANT ALL ON octopus.* TO 'octopus'@'%';
   ```
   ```sql
   -- PostgreSQL
   CREATE DATABASE octopus ENCODING 'UTF8';
   CREATE USER octopus PASSWORD 'CHANGE_ME';
   GRANT ALL ON DATABASE octopus TO octopus;
   ```

3. **切换配置**:
   ```bash
   OCTOPUS_DATABASE_TYPE=mysql
   # MySQL: 通过 DATABASE_DSN 或 environment 直连
   # PostgreSQL:
   OCTOPUS_DATABASE_TYPE=postgres
   ```

   > **注意**:具体 DSN 环境变量名以 `internal/conf/config.go` 与 `internal/model/` 实际实现为准。当前 `Database` struct 仅暴露 `type` 与 `path`,MySQL/PostgreSQL 连接串可能通过 `path` 字段或独立环境变量传递。迁移前请先在测试环境验证连接参数格式。

4. **启动新实例**:Octopus 启动时会 AutoMigrate 创建所有表。

5. **导入数据**:使用 `pgloader`(SQLite → PostgreSQL)或 `sqlite3 dump` + `mysql` 导入(SQLite → MySQL)。Octopus 的备份/恢复 API(`/api/v1/backup/*`)可能跨数据库兼容,优先尝试。

6. **验证**:对比记录数、触发一次完整请求、查看 stats 是否正常累计。

#### 2.4.2 数据库垂直扩容

- MySQL/PostgreSQL:调整实例规格(CPU/内存/IOPS)、连接数上限
- 连接池:Octopus 通过 GORM 默认连接池,生产建议显式设置(若配置项暴露)
- 索引:`relay_log`、`stats_*` 表在大数据量下查询变慢,建议按 `time` 列建索引

---

## 3. 配置

### 3.1 环境变量完整列表

Octopus 使用 Viper 配置,环境变量前缀为 `OCTOPUS_`,分隔符 `.` 映射为 `_`(例如 `server.host` → `OCTOPUS_SERVER_HOST`)。

以下环境变量列表源自 `internal/conf/config.go` 的 `setDefaults()` 与 struct 定义:

#### Server

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OCTOPUS_SERVER_HOST` | `0.0.0.0` | 监听地址 |
| `OCTOPUS_SERVER_PORT` | `1088` | 监听端口 |
| `OCTOPUS_SERVER_STATIC_DIR` | `static/out` | 前端静态文件目录(嵌入二进制时无需修改) |

#### Database

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OCTOPUS_DATABASE_TYPE` | `sqlite` | 数据库类型:`sqlite` / `mysql` / `postgresql` |
| `OCTOPUS_DATABASE_PATH` | `data/data.db` | SQLite 文件路径;MySQL/PG 时为连接串或 DSN |

#### Log

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OCTOPUS_LOG_LEVEL` | `info` | 日志级别:`debug` / `info` / `warn` / `error` |

#### Shutdown(优雅停机)

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OCTOPUS_SHUTDOWN_TIMEOUT` | `30s` | 停机总超时 |
| `OCTOPUS_SHUTDOWN_TASK_TIMEOUT` | `10s` | 单个清理任务超时 |

#### HTTP Client(出站连接池调优)

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OCTOPUS_HTTP_MAX_IDLE_CONNS` | `200` | 全局空闲连接上限 |
| `OCTOPUS_HTTP_MAX_CONNS_PER_HOST` | `100` | 单上游 host 最大连接数 |
| `OCTOPUS_HTTP_MAX_IDLE_CONNS_PER_HOST` | `100` | 单上游 host 最大空闲连接数 |
| `OCTOPUS_HTTP_IDLE_CONN_TIMEOUT` | `90s` | 空闲连接超时 |
| `OCTOPUS_HTTP_TLS_HANDSHAKE_TIMEOUT` | `10s` | TLS 握手超时 |
| `OCTOPUS_HTTP_EXPECT_CONTINUE_TIMEOUT` | `1s` | Expect-100-continue 超时 |
| `OCTOPUS_HTTP_RESPONSE_HEADER_TIMEOUT` | `30s` | 等待响应头超时(过长延长故障检测,过短误判慢上游) |

#### 容器运行时(docker-compose / install.sh 使用)

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `DATA_DIR` | `/app/data` | 容器内数据目录(非 Octopus 配置,由 entrypoint 使用) |
| `PUID` | `10001` | 容器运行用户 UID |
| `PGID` | `10001` | 容器运行用户 GID |

#### 安装脚本(`scripts/install.sh`)

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OCTOPUS_PORT` | `1088` | 宿主机外部端口 |
| `OCTOPUS_DATA_DIR` | `./data` | 宿主机数据目录 |
| `OCTOPUS_CONTAINER_NAME` | `octopus` | 容器名 |
| `OCTOPUS_IMAGE` | `ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.20.0` | 镜像地址 |
| `OCTOPUS_BINARY_PATH` | _(空)_ | 兜底本地二进制路径 |
| `OCTOPUS_REPO_URL` | `https://github.com/xiaoli0412/octopus-xiaoli-repo.git` | 源码回退仓库 |
| `OCTOPUS_REPO_REF` | 与镜像 tag 同 | 源码回退 checkout ref |

### 3.2 配置文件

除环境变量外,Octopus 支持通过 JSON 配置文件配置。启动时查找顺序:
1. `--config` 参数指定的路径
2. `data/config.json`(默认查找路径)

首次启动若 `data/config.json` 不存在,Octopus 会自动创建包含默认值的配置文件。

示例 `data/config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 1088,
    "static_dir": "static/out"
  },
  "database": {
    "type": "sqlite",
    "path": "data/data.db"
  },
  "log": {
    "level": "info"
  },
  "shutdown": {
    "timeout": "30s",
    "task_timeout": "10s"
  },
  "http": {
    "max_idle_conns": 200,
    "max_conns_per_host": 100,
    "max_idle_conns_per_host": 100,
    "idle_conn_timeout": "90s",
    "tls_handshake_timeout": "10s",
    "expect_continue_timeout": "1s",
    "response_header_timeout": "30s"
  }
}
```

> **优先级**:环境变量 > 配置文件 > 默认值。
>
> **BOM 兼容**:若配置文件含 UTF-8 BOM,Viper 首次读取失败后会自动剥除 BOM 重试(见 `internal/conf/config.go` 的 `readConfigWithBOMFallback`)。

### 3.3 密钥管理

Octopus 涉及两类敏感数据:

#### 3.3.1 JWT 签名密钥

- **存储位置**:数据库 `Setting` 表,key 为 `auth_token_secret`(见 `internal/server/auth/auth.go` 与 `internal/model/setting.go`)。
- **生成方式**:首次启动自动生成;修改管理员密码会轮换密钥(旧 JWT 立即失效)。
- **生产建议**:
  - 不要将 JWT 密钥以环境变量明文传递(当前实现不支持环境变量覆盖)。
  - 数据库文件/卷需加密(SQLite 文件加密、MySQL/PG at-rest 加密)。
  - 多实例共享同一数据库时,JWT 密钥自动共享,无需额外同步。
- **轮换流程**:通过管理 API 修改管理员密码即可触发密钥轮换,所有活跃 JWT 立即失效。

#### 3.3.2 客户端 API Keys

- **存储位置**:数据库 `APIKey` 表,以哈希存储(非明文)。
- **管理方式**:通过 `POST /api/v1/apikey` 创建,返回明文 key 仅一次,后续无法找回(需重新生成)。
- **生产建议**:
  - 创建时立即妥善保存到密钥管理系统(如 Vault、AWS Secrets Manager、k8s Secret)。
  - 不要在日志、监控标签、RelayLog 中暴露完整 key。
  - 定期轮换,通过管理 API 禁用旧 key、创建新 key,灰度切换客户端。

#### 3.3.3 上游渠道 API Keys

- **存储位置**:数据库 `ChannelKey` 表。
- **生产建议**:同上,使用密钥管理系统托管,避免明文写入 docker-compose env 或 ConfigMap。

#### 3.3.4 k8s 部署的密钥管理

```bash
# 创建含数据库凭据的 Secret
kubectl create secret generic octopus-secret \
  -n octopus \
  --from-literal=DATABASE_DSN='octopus:CHANGE_ME@tcp(mysql:3306)/octopus' \
  --dry-run=client -o yaml | kubectl apply -f -

# 通过 envFrom 注入(见 1.3.3 Deployment)
```

> **禁止行为**:不要把密钥写进 ConfigMap、Dockerfile、git 仓库、镜像 label。

### 3.4 可观测性配置

#### 3.4.1 Prometheus 指标

- **端点**:`GET /metrics`(Prometheus 文本格式,由 `promhttp.Handler()` 暴露)
- **启用方式**:默认启用,**无需任何配置**。当前不存在 `OCTOPUS_METRICS_ENABLED` 开关;只要服务在运行,`/metrics` 即可抓取。
- **抓取配置**:
  - 静态配置:`scrape_configs` 指向 `http://<host>:1088/metrics`
  - k8s:通过 ServiceMonitor(见 [1.3.5](#135-servicemonitorprometheus-operator))或 Pod annotation `prometheus.io/scrape: "true"`
- **采集建议**:`interval: 30s`,`scrapeTimeout: 10s`

#### 3.4.2 OpenTelemetry Tracing

- **配置项**:`OTEL_EXPORTER_OTLP_ENDPOINT`(**注意**:此为标准 OTEL 环境变量,无 `OCTOPUS_` 前缀)
- **默认行为**:未设置时为 noop tracer,零开销(见 `internal/observability/trace.go`)
- **启用示例**:
  ```bash
  OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
  ```
- **后端建议**:接入 Tempo / Jaeger / Honeycomb,采样率生产建议 0.1-0.5。

#### 3.4.3 日志

- **格式**:Zap 结构化日志(JSON 或 console,取决于构建配置)
- **级别**:通过 `OCTOPUS_LOG_LEVEL` 控制,生产建议 `info`,排障时临时调到 `debug`
- **聚合建议**:容器 stdout → Fluent Bit / Filebeat → Loki / ELK

---

## 4. 故障排查

### 4.1 数据库连接失败

**症状**:
- `/readyz` 返回 503,reasons 含 `database`
- 启动日志出现 `failed to connect database`、`dial tcp: connect: connection refused`
- 管理界面无法登录、stats 不更新

**诊断步骤**:

1. 查看 `/readyz` 输出:
   ```bash
   curl -s http://localhost:1088/readyz | jq .
   ```
2. 查看容器/进程日志:
   ```bash
   docker logs octopus --tail 100 | grep -iE 'database|sql|gorm'
   # 或二进制:journalctl -u octopus --since '10 min ago' | grep -iE 'database|sql'
   ```
3. 检查配置:
   ```bash
   docker exec octopus env | grep OCTOPUS_DATABASE
   ```
4. 验证数据库连通性:
   - SQLite:`ls -l /app/data/data.db`,确认文件存在且可写
   - MySQL:`mysql -h <host> -u <user> -p<pass> -e "SELECT 1"`
   - PostgreSQL:`psql -h <host> -U <user> -d octopus -c "SELECT 1"`

**处置措施**:

- **SQLite 文件不存在/不可写**:检查挂载路径、PUID/PGID 权限。容器内 `ls -ld /app/data` 应显示 `10001` 属主。
- **MySQL/PG 网络不通**:检查 SecurityGroup/NetworkPolicy、DNS 解析(`nslookup mysql.octopus.svc`)。
- **凭据错误**:重新生成密码,更新 `OCTOPUS_DATABASE_PATH` 或 Secret,滚动重启。
- **数据库满**:检查磁盘空间 `df -h`,清理或扩容。

**预防建议**:
- `/readyz` 接入 readiness 探针,故障时自动摘除流量。
- 数据库配置变更先在测试环境验证。
- 数据库监控:连接数、慢查询、磁盘使用率。

---

### 4.2 渠道全部熔断

**症状**:
- 所有中继请求返回 5xx,错误信息含 `no available channel`、`circuit breaker open`
- `/metrics` 中 `octopus_circuit_breaker_state` 多为 1(open)或 2(half-open)
- `octopus_channel_health` 多为 0
- 管理界面渠道状态显示不健康

**诊断步骤**:

1. 查看熔断器状态:
   ```bash
   curl -s http://localhost:1088/metrics | grep octopus_circuit_breaker_state
   # 0=closed(正常), 1=open(熔断), 2=half-open(探测)
   ```
2. 查看渠道健康:
   ```bash
   curl -s http://localhost:1088/metrics | grep octopus_channel_health
   ```
3. 查看近期失败请求:
   ```bash
   curl -s -H "Authorization: Bearer $JWT" \
     "http://localhost:1088/api/v1/stats/token-breakdown?window=1h" | jq .
   ```
4. 检查上游可用性(从 Octopus 主机直接 curl 上游 base_url)。
5. 检查上游 API Key 有效性、配额、限流。

**处置措施**:

- **上游临时故障**:等待熔断器 half-open 自动探测恢复(无需手动干预,冷却时间由熔断器配置决定)。
- **上游 API Key 失效**:更新渠道 key,触发同步 `POST /api/v1/channel/{id}/sync`。
- **上游限流**:降低 group weight 或临时禁用部分客户端 API Key,缓解上游压力。
- **手动重置熔断器**:通过管理 API 或重启实例重置熔断状态(具体 API 以代码实现为准;若管理 API 未暴露,重启实例可使熔断器状态清零)。
- **切换 group 策略**:临时切换为 `failover`,优先用健康渠道。

**预防建议**:
- 配置多渠道 + 多 key,避免单点。
- 设置合理的熔断阈值(连续失败次数)与冷却时间,避免抖动。
- 告警:`circuit_breaker_open_ratio > 50%`(见 [6.3](#63-告警规则prometheus-alertmanager))。

---

### 4.3 OOM(内存溢出)

**症状**:
- 容器被 OOMKilled(`docker inspect` 显示 `OOMKilled: true`)
- k8s Pod 重启,事件含 `OOMKilled`
- 进程突然消失,无 panic 日志

**诊断步骤**:

1. 查看容器/Pod 重启历史:
   ```bash
   docker inspect octopus | jq '.[0].State.OOMKilled'
   kubectl describe pod -n octopus -l app=octopus | grep -A5 'Last State'
   ```
2. 查看资源使用趋势(Grafana / cAdvisor):
   - 内存是否持续上涨(泄漏)还是突发峰值(大请求)
   - 是否与某个大模型/大上下文请求相关
3. 检查资源限制是否过低:
   ```bash
   kubectl get pod -n octopus -l app=octopus -o jsonpath='{.items[0].spec.containers[0].resources}'
   ```

**处置措施**:

- **临时缓解**:提高内存 limit,重启实例。
- **大请求导致**:检查是否有客户端发送超大 context(如完整代码库),通过 API Key 限流或模型 `max_tokens` 限制。
- **缓存膨胀**:检查 channel/group/model 缓存大小,确认缓存 TTL 配置合理。
- **SSE 连接堆积**:见 [4.5](#45-sse-连接泄漏)。
- **goroutine 泄漏**:见 [4.4](#44-内存泄漏)。

**预防建议**:
- 设置合理的 `resources.limits.memory`(起步 2Gi)。
- 配置 HPA 基于 CPU/内存自动扩容。
- 监控内存使用率告警(> 80% 持续 5 分钟)。

---

### 4.4 内存泄漏

**症状**:
- 内存使用率随时间单调上升,重启后恢复正常
- goroutine 数量持续增长(`/metrics` 无直接指标,通过 pprof 看)
- 长时间运行后响应延迟上升(GC 压力)

**诊断步骤**:

1. 启用 pprof(若构建包含 `net/http/pprof`,Octopus 通过 debug 路由暴露;具体路径以 `internal/server/server.go` 为准):
   ```bash
   # heap profile
   curl -s http://localhost:1088/debug/pprof/heap > heap.pprof
   go tool pprof -top heap.pprof

   # goroutine
   curl -s http://localhost:1088/debug/pprof/goroutine > goroutine.pprof
   go tool pprof -top goroutine.pprof
   ```
   > **注意**:若 pprof 路由未暴露,需在 `conf.Debug` 启用或构建 debug 版本。生产环境临时排障可使用二进制 + `--debug` 启动。

2. 对比不同时间点的 heap profile,定位增长点:
   ```bash
   go tool pprof -base heap-09h.pprof heap-10h.pprof
   ```

3. 检查 goroutine 数量趋势(通过 pprof goroutine count)。

**处置措施**:

- **临时缓解**:定期滚动重启实例(治标),配合 HPA 横向扩容吸收流量。
- **根因修复**:根据 pprof 定位的分配热点修复代码,优先排查:
  - SSE 流式响应未正确关闭(见 [4.5](#45-sse-连接泄漏))
  - 大 context 请求的内存缓冲未释放
  - 缓存无 TTL 或 TTL 过长
  - goroutine 在 context 取消后未退出

**预防建议**:
- CI 加入 `go test -race`(仓库 release 流程已包含,见 `.github/workflows/release.yaml`)。
- 长稳测试:用 `load-test/run.sh all` 跑 1 小时,观察内存趋势。
- 监控 goroutine 数量(若暴露指标)与内存使用率。

---

### 4.5 SSE 连接泄漏

**症状**:
- goroutine 数量持续增长
- 内存使用率上升
- `/api/v1/stream/stats` 连接数持续增长不下降
- 客户端断开后服务端仍持有连接

**诊断步骤**:

1. 查看 SSE 相关 goroutine:
   ```bash
   curl -s http://localhost:1088/debug/pprof/goroutine?debug=2 | grep -c 'sse\|stream\|Flusher'
   ```
2. 查看活跃 SSE 连接(若管理 API 暴露):
   ```bash
   curl -s -H "Authorization: Bearer $JWT" http://localhost:1088/api/v1/stats/total | jq .
   ```
3. 复现:用 `curl -N http://localhost:1088/api/v1/stream/stats` 建立连接,Ctrl+C 断开后观察服务端 goroutine 是否减少。
4. 检查客户端断开检测:Octopus 应通过 `c.Request.Context().Done()` 感知客户端断开并退出推送 goroutine。

**处置措施**:

- **临时缓解**:重启实例。
- **根因修复**:
   - 确认 SSE handler 正确监听 `ctx.Done()`,客户端断开时立即返回。
   - 确认 `sse.Source` 的订阅者在 handler 退出时被正确取消订阅(避免订阅列表无限增长)。
   - 检查 `tmaxmax/go-sse` 库的使用是否符合最佳实践(订阅者缓冲区满时的丢弃策略)。

**预防建议**:
- 长稳测试包含 SSE 场景(建立/断开循环 1000 次,验证 goroutine 回收)。
- 监控 SSE 连接数(若有指标)与 goroutine 总数。
- 客户端实现健壮的重连逻辑(带 backoff),避免短时间大量重连。

---

### 4.6 上游同步失败

**症状**:
- 管理界面渠道「同步模型」失败
- 日志含 `failed to sync models`、`failed to fetch pricing`
- 模型列表不更新、定价信息过期

**诊断步骤**:

1. 手动触发同步,查看详细错误:
   ```bash
   curl -s -X POST -H "Authorization: Bearer $JWT" \
     http://localhost:1088/api/v1/channel/{id}/sync | jq .
   ```
2. 检查上游连通性(从 Octopus 主机):
   ```bash
   curl -v -H "Authorization: Bearer <upstream-key>" \
     https://api.openai.com/v1/models
   ```
3. 检查上游 API Key 有效性、配额。
4. 检查渠道 `base_url` 配置是否正确(末尾斜杠、路径)。
5. 查看价格同步(models.dev):
   ```bash
   curl -v https://models.dev/api/v1/pricing
   ```

**处置措施**:

- **API Key 失效**:更新渠道 key,重新同步。
- **网络问题**:检查出站代理、DNS、防火墙。Octopus 容器需能访问外网(或配置 `extra_hosts` / proxy)。
- **base_url 错误**:修正渠道配置(注意 OpenAI 格式为 `https://api.openai.com`,Anthropic 为 `https://api.anthropic.com`)。
- **models.dev 不可达**:价格同步是后台任务,失败不影响中继功能,可延后处理。
- **同步频率**:避免过高频率同步(建议每日一次),触发上游限流。

**预防建议**:
- 同步任务失败时告警(若有相关指标或日志告警)。
- 渠道配置变更后手动触发一次同步验证。
- 多渠道时,单渠道同步失败不影响其他渠道。

---

## 5. 回滚

### 5.1 版本回滚

#### 5.1.1 Docker 镜像回滚

```bash
# 1. 确认当前版本
docker inspect octopus | jq '.[0].Config.Image'

# 2. 回滚到上一个 tag(例如从 v1.20.0 回到 v1.19.8)
docker rm -f octopus
docker run -d \
  --name octopus \
  --restart unless-stopped \
  -p 1088:1088 \
  -e OCTOPUS_DATABASE_TYPE=sqlite \
  -e OCTOPUS_DATABASE_PATH=/app/data/data.db \
  -e OCTOPUS_LOG_LEVEL=info \
  -e DATA_DIR=/app/data \
  -e PUID=10001 \
  -e PGID=10001 \
  -v /data/octopus:/app/data \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m \
  ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.19.8

# 3. 验证
curl -fsS http://127.0.0.1:1088/healthz
curl -fsS http://127.0.0.1:1088/readyz
```

docker-compose 方式:

```bash
OCTOPUS_IMAGE=ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.19.8 \
docker compose up -d
```

#### 5.1.2 k8s 回滚

```bash
# 查看发布历史
kubectl rollout history deployment/octopus -n octopus

# 回滚到上一版本
kubectl rollout undo deployment/octopus -n octopus

# 回滚到指定版本
kubectl rollout undo deployment/octopus -n octopus --to-revision=3

# 监控回滚状态
kubectl rollout status deployment/octopus -n octopus
```

#### 5.1.3 二进制回滚

```bash
# 1. 停止当前服务
systemctl stop octopus   # 或 docker stop octopus

# 2. 替换二进制(从 GitHub Release 下载旧版本)
curl -L -o /opt/octopus/octopus \
  https://github.com/xiaoli0412/octopus-xiaoli-repo/releases/download/v1.19.8/octopus-linux-amd64
chmod +x /opt/octopus/octopus

# 3. 启动
systemctl start octopus
```

> **注意**:回滚后,数据库 schema 可能已被新版本的 AutoMigrate 升级。GORM AutoMigrate **不会自动回退** schema,见 [5.2](#52-数据库回滚)。多数情况下新增字段/表对旧版本兼容(旧代码忽略新字段),但删除/重命名字段可能导致旧版本报错。回滚前务必在测试环境验证。

---

### 5.2 数据库回滚

> **关键**:GORM AutoMigrate 只增不减,**不提供回滚迁移**。数据库回滚依赖**备份恢复**。

#### 5.2.1 升级前备份(必备)

```bash
# 通过管理 API 备份(推荐)
curl -s -H "Authorization: Bearer $JWT" \
  http://localhost:1088/api/v1/backup/download \
  -o backup-pre-upgrade-$(date +%Y%m%d-%H%M%S).tar.gz

# 或直接备份 SQLite 文件(停机备份)
docker stop octopus
cp /data/octopus/data.db /data/octopus/data.db.pre-upgrade-$(date +%Y%m%d)
docker start octopus
```

#### 5.2.2 回滚步骤

1. **停止 Octopus**:
   ```bash
   docker stop octopus   # 或 systemctl stop octopus
   ```
2. **恢复备份**:
   - SQLite:替换 `data.db` 文件
     ```bash
     cp /data/octopus/data.db.pre-upgrade-20260711 /data/octopus/data.db
     chown 10001:10001 /data/octopus/data.db
     ```
   - MySQL/PG:使用 `mysql` / `pg_restore` 恢复 dump
3. **启动旧版本二进制/镜像**(见 [5.1](#51-版本回滚))。
4. **验证**:登录、触发一次中继请求、查看 stats。

#### 5.2.3 无法回滚 schema 的情况

若新版本已执行 AutoMigrate 且无法用旧备份恢复:
- **新增字段**:旧版本忽略,通常兼容。
- **删除字段**:旧版本查询会报错,需手动 `ALTER TABLE` 恢复列(从备份导出该列数据)。
- **重命名字段**:同上,手动 `ALTER TABLE ... RENAME COLUMN`。
- **新增表**:旧版本忽略,无影响。

> **强烈建议**:重大版本升级前,先在测试环境跑完整流程(升级 → 验证 → 回滚),确认 schema 兼容性。

---

### 5.3 配置回滚

#### 5.3.1 环境变量回滚

```bash
# 1. 记录当前配置
docker inspect octopus | jq '.[0].Config.Env' > octopus-env-$(date +%Y%m%d).json

# 2. 修改 docker-compose.yml 或 docker run 命令,恢复旧值
# 3. 重新部署
docker compose up -d   # 或 docker rm -f octopus && docker run ...
```

k8s:

```bash
# 回滚 ConfigMap(需有版本控制,如 GitOps)
kubectl apply -f octopus-config-v1.19.yaml

# 重启 Pod 使配置生效
kubectl rollout restart deployment/octopus -n octopus
```

#### 5.3.2 配置文件回滚

```bash
# 1. 备份当前配置
cp /data/octopus/config.json /data/octopus/config.json.bak

# 2. 恢复旧配置
cp /data/octopus/config.json.v1.19 /data/octopus/config.json

# 3. 重启
docker restart octopus
```

#### 5.3.3 数据库内配置回滚

部分配置存储在 `Setting` 表(如 JWT 密钥、动态路由开关)。回滚方式:
- 通过管理 API 逐项修改
- 直接 SQL 更新(谨慎,需确认 key 名):
  ```sql
  UPDATE settings SET value = '<old_value>' WHERE key = '<setting_key>';
  ```

> **建议**:生产环境配置变更全部通过 GitOps(IaC)管理,环境变量与 ConfigMap 版本化,便于回滚。

---

## 6. 监控告警

### 6.1 Prometheus 指标列表

以下指标定义见 `internal/observability/metrics.go`,通过 `GET /metrics` 暴露:

| 指标名 | 类型 | Labels | 说明 |
|--------|------|--------|------|
| `octopus_relay_requests_total` | Counter | `channel_id`, `model`, `provider_type`, `status` | 中继请求总数(status 含 success/fail) |
| `octopus_relay_duration_seconds` | Histogram | `channel_id`, `model`, `provider_type` | 中继请求耗时(秒) |
| `octopus_channel_health` | Gauge | `channel_id` | 渠道健康状态(1=健康, 0=不健康) |
| `octopus_circuit_breaker_state` | Gauge | `channel_id`, `model` | 熔断器状态(0=closed, 1=open, 2=half-open) |
| `octopus_token_throughput_total` | Counter | `channel_id`, `model`, `type` | token 吞吐总量(type 含 input/output/cache_read/cache_write) |
| `octopus_http_client_pool_idle` | Gauge | `client_type` | 出站 HTTP 客户端空闲连接数 |
| `octopus_http_requests_total` | Counter | `method`, `path`, `status` | HTTP 请求总数(网关自身入站) |
| `octopus_http_request_duration_seconds` | Histogram | `method`, `path` | HTTP 请求耗时(秒,网关自身入站) |

> **抓取建议**:`interval: 30s`,Histogram 默认使用 Prometheus DefBuckets(`0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10`)。

### 6.2 Grafana 仪表板建议

| 面板 | 指标 | PromQL 示例 | 面板类型 |
|------|------|-------------|----------|
| **RPS(中继请求速率)** | `octopus_relay_requests_total` | `sum(rate(octopus_relay_requests_total[1m]))` | Time series |
| **延迟分布(P50/P95/P99)** | `octopus_relay_duration_seconds` | `histogram_quantile(0.95, sum(rate(octopus_relay_duration_seconds_bucket[5m])) by (le))` | Time series |
| **错误率** | `octopus_relay_requests_total` | `sum(rate(octopus_relay_requests_total{status="fail"}[5m])) / sum(rate(octopus_relay_requests_total[5m]))` | Stat / Gauge |
| **渠道健康** | `octopus_channel_health` | `octopus_channel_health` (按 channel_id 分组) | Status panel / Table |
| **熔断器状态** | `octopus_circuit_breaker_state` | `count(octopus_circuit_breaker_state == 1) by (channel_id)` | Stacked bar |
| **Token 吞吐** | `octopus_token_throughput_total` | `sum(rate(octopus_token_throughput_total[1m])) by (type)` | Time series |
| **HTTP 连接池** | `octopus_http_client_pool_idle` | `octopus_http_client_pool_idle` | Stat |
| **网关入站 RPS** | `octopus_http_requests_total` | `sum(rate(octopus_http_requests_total[1m])) by (path)` | Time series |
| **网关入站延迟** | `octopus_http_request_duration_seconds` | `histogram_quantile(0.95, sum(rate(octopus_http_request_duration_seconds_bucket[5m])) by (le, path))` | Time series |

**仪表板组织建议**:
- Row 1: 总览(RPS、错误率、P95、活跃渠道数)
- Row 2: 中继详情(按 model/channel 分组)
- Row 3: 渠道健康与熔断器
- Row 4: 资源与连接池
- Row 5: 网关入站 HTTP

### 6.3 告警规则(Prometheus AlertManager)

```yaml
# alerting-rules.yaml
groups:
  - name: octopus
    rules:
      # 1. 中继错误率 > 5% 持续 5 分钟
      - alert: OctopusRelayHighErrorRate
        expr: |
          (
            sum(rate(octopus_relay_requests_total{status="fail"}[5m]))
            /
            clamp_min(sum(rate(octopus_relay_requests_total[5m])), 1)
          ) > 0.05
        for: 5m
        labels:
          severity: critical
          service: octopus
        annotations:
          summary: "Octopus 中继错误率过高 ({{ $value | humanizePercentage }})"
          description: "Octopus 中继请求错误率超过 5% 已持续 5 分钟,请检查渠道健康与上游可用性。"

      # 2. 中继 P95 延迟 > 5s 持续 5 分钟
      - alert: OctopusRelayHighLatencyP95
        expr: |
          histogram_quantile(0.95,
            sum(rate(octopus_relay_duration_seconds_bucket[5m])) by (le)
          ) > 5
        for: 5m
        labels:
          severity: warning
          service: octopus
        annotations:
          summary: "Octopus 中继 P95 延迟过高 ({{ $value }}s)"
          description: "中继请求 P95 延迟超过 5s 已持续 5 分钟,请检查上游延迟、连接池配置与并发负载。"

      # 3. 渠道全部不健康(channel_health == 0 持续 10 分钟)
      - alert: OctopusAllChannelsUnhealthy
        expr: |
          count(octopus_channel_health == 1) == 0
          and
          count(octopus_channel_health) > 0
        for: 10m
        labels:
          severity: critical
          service: octopus
        annotations:
          summary: "Octopus 所有渠道均不健康"
          description: "所有已配置渠道的健康状态均为 0 已持续 10 分钟,中继功能完全不可用。立即检查上游与渠道配置。"

      # 4. 熔断器开启比例 > 50% 持续 5 分钟
      - alert: OctopusCircuitBreakerHighOpenRatio
        expr: |
          (
            count(octopus_circuit_breaker_state == 1)
            /
            clamp_min(count(octopus_circuit_breaker_state), 1)
          ) > 0.5
        for: 5m
        labels:
          severity: warning
          service: octopus
        annotations:
          summary: "Octopus 熔断器开启比例过高 ({{ $value | humanizePercentage }})"
          description: "超过 50% 的熔断器处于 open 状态已持续 5 分钟,可能存在上游批量故障或配置问题。"

      # 5. (补充)实例宕机
      - alert: OctopusDown
        expr: up{job="octopus"} == 0
        for: 2m
        labels:
          severity: critical
          service: octopus
        annotations:
          summary: "Octopus 实例宕机 ({{ $labels.instance }})"
          description: "Prometheus 抓取失败已持续 2 分钟,实例可能已宕机或 /metrics 不可达。"

      # 6. (补充)HTTP 连接池耗尽
      - alert: OctopusHTTPPoolExhausted
        expr: octopus_http_client_pool_idle < 5
        for: 5m
        labels:
          severity: warning
          service: octopus
        annotations:
          summary: "Octopus HTTP 连接池空闲数过低 ({{ $value }})"
          description: "出站 HTTP 连接池空闲连接数低于 5 已持续 5 分钟,考虑调大 max_idle_conns 或排查连接泄漏。"
```

**告警收敛建议**:
- 同一 `channel_id` 的告警合并(避免渠道级抖动刷屏)
- critical 告警通过 PagerDuty / 电话通知,warning 通过 Slack / 飞书
- 告警规则中 `for` 持续时间避免过短(建议 ≥ 5m),减少瞬时抖动误报
- `clamp_min` 防止分母为 0 导致告警异常

**Alertmanager 路由示例**:

```yaml
route:
  receiver: default
  group_by: ['alertname', 'service', 'channel_id']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  routes:
    - match:
        severity: critical
      receiver: oncall-pager
    - match:
        severity: warning
      receiver: slack-octopus
```

---

## 附录:常用命令速查

```bash
# 健康检查
curl -fsS http://localhost:1088/healthz
curl -fsS http://localhost:1088/readyz

# 查看指标
curl -s http://localhost:1088/metrics | grep octopus_

# 登录获取 JWT
curl -s -X POST http://localhost:1088/api/v1/user/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"CHANGE_ME"}'

# 查看渠道健康
curl -s -H "Authorization: Bearer $JWT" http://localhost:1088/api/v1/stats/token-breakdown | jq .

# 备份
curl -s -H "Authorization: Bearer $JWT" http://localhost:1088/api/v1/backup/download -o backup.tar.gz

# 触发渠道同步
curl -s -X POST -H "Authorization: Bearer $JWT" http://localhost:1088/api/v1/channel/{id}/sync

# 压测
cd load-test
BASE_URL=http://localhost:1088 API_KEY=sk-xxxx JWT_TOKEN=eyJ... ./run.sh all

# pprof(若启用)
curl -s http://localhost:1088/debug/pprof/heap > heap.pprof
go tool pprof -top heap.pprof
```
