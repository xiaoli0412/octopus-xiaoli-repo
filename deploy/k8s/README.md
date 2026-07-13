# Octopus Kubernetes 部署指南

本目录包含 Octopus(LLM API 聚合与负载均衡服务)的 Kubernetes 部署清单模板。

> 完整运维文档请参考 [`docs/RUNBOOK.md`](../../docs/RUNBOOK.md)。

## 清单文件概览

| 文件 | 说明 |
|------|------|
| `namespace.yaml` | 创建 `octopus` namespace |
| `configmap.yaml` | 非敏感配置(Server、Database、Log、HTTP 连接池等) |
| `secret.yaml` | 敏感配置占位符(管理员密码、JWT 密钥、数据库 DSN) |
| `pvc.yaml` | 数据持久化卷(默认 10Gi) |
| `deployment.yaml` | 工作负载定义(单副本、探针、资源配额、安全上下文) |
| `service.yaml` | ClusterIP Service |
| `ingress.yaml` | Nginx Ingress(含 SSE 长连接优化注解) |

## 1. 前置条件

- **Kubernetes 集群**:v1.22+(需支持 `networking.k8s.io/v1` Ingress API)
- **kubectl**:已配置集群访问凭据
- **StorageClass**:集群需有可用的 StorageClass(用于 PVC 动态供给)

  ```bash
  kubectl get storageclass
  ```

- **Nginx Ingress Controller**(可选,使用 Ingress 时需要)
- **Prometheus Operator**(可选,使用 ServiceMonitor 时需要)

## 2. 部署步骤

### 2.1 创建 namespace

```bash
kubectl apply -f namespace.yaml
```

### 2.2 创建 Secret(使用真实凭据)

**请勿直接 apply `secret.yaml` 中的占位符!** 使用以下命令创建包含真实凭据的 Secret:

```bash
kubectl create secret generic octopus-secret -n octopus \
  --from-literal=OCTOPUS_ADMIN_PASSWORD='your-password' \
  --from-literal=OCTOPUS_JWT_SECRET='your-jwt-secret' \
  --from-literal=DATABASE_DSN='octopus:password@tcp(mysql:3306)/octopus'
```

> SQLite 模式下 `DATABASE_DSN` 可传空字符串或省略。

### 2.3 应用其余配置

```bash
kubectl apply -f configmap.yaml -f pvc.yaml -f deployment.yaml -f service.yaml -f ingress.yaml
```

### 2.4 一键部署(跳过 Secret 占位符)

若已通过 2.2 创建 Secret,可一键应用全部清单(Secret 已存在时 apply 不会覆盖真实值):

```bash
kubectl apply -f deploy/k8s/
```

> **注意**:如果直接 `kubectl apply -f secret.yaml`,其中的 `<change-me>` 占位符会被写入。请务必在 apply 后通过 2.2 的命令覆盖,或编辑 Secret 替换占位符。

## 3. 验证部署

```bash
# 查看 Pod 状态
kubectl get pods -n octopus

# 查看日志
kubectl logs -n octopus -l app=octopus

# 查看所有资源
kubectl get all -n octopus

# 验证健康检查
kubectl exec -n octopus deployment/octopus -- curl -s http://127.0.0.1:1088/healthz
# 期望: {"status":"ok"}

kubectl exec -n octopus deployment/octopus -- curl -s http://127.0.0.1:1088/readyz
# 期望: 200,失败返回 503

# 端口转发本地访问
kubectl port-forward -n octopus svc/octopus 1088:1088
# 浏览器打开 http://127.0.0.1:1088
```

## 4. 配置说明

### 4.1 ConfigMap(`octopus-config`)

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `OCTOPUS_SERVER_HOST` | `0.0.0.0` | 监听地址 |
| `OCTOPUS_SERVER_PORT` | `1088` | 监听端口 |
| `OCTOPUS_DATABASE_TYPE` | `sqlite` | 数据库类型:`sqlite` / `mysql` / `postgresql` |
| `OCTOPUS_DATABASE_PATH` | `/app/data/data.db` | SQLite 文件路径 |
| `OCTOPUS_LOG_LEVEL` | `info` | 日志级别:`debug` / `info` / `warn` / `error` |
| `OCTOPUS_SHUTDOWN_TIMEOUT` | `30s` | 停机总超时 |
| `OCTOPUS_SHUTDOWN_TASK_TIMEOUT` | `10s` | 单个清理任务超时 |
| `OCTOPUS_HTTP_MAX_IDLE_CONNS` | `200` | 全局空闲连接上限 |
| `OCTOPUS_HTTP_MAX_CONNS_PER_HOST` | `100` | 单上游 host 最大连接数 |
| `OCTOPUS_HTTP_MAX_IDLE_CONNS_PER_HOST` | `100` | 单上游 host 最大空闲连接数 |
| `OCTOPUS_HTTP_IDLE_CONN_TIMEOUT` | `90s` | 空闲连接超时 |
| `OCTOPUS_HTTP_TLS_HANDSHAKE_TIMEOUT` | `10s` | TLS 握手超时 |
| `OCTOPUS_HTTP_EXPECT_CONTINUE_TIMEOUT` | `1s` | Expect-100-continue 超时 |
| `OCTOPUS_HTTP_RESPONSE_HEADER_TIMEOUT` | `30s` | 等待响应头超时 |
| `DATA_DIR` | `/app/data` | 容器内数据目录(entrypoint 使用) |
| `PUID` | `10001` | 容器运行用户 UID |
| `PGID` | `10001` | 容器运行用户 GID |

### 4.2 Secret(`octopus-secret`)

| 环境变量 | 说明 |
|----------|------|
| `OCTOPUS_ADMIN_PASSWORD` | 管理员初始密码(首次启动初始化) |
| `OCTOPUS_JWT_SECRET` | JWT 签名密钥 |
| `DATABASE_DSN` | MySQL/PostgreSQL 连接串(SQLite 模式可留空) |

### 4.3 修改配置

编辑 ConfigMap 后需重启 Pod 使配置生效:

```bash
kubectl edit configmap octopus-config -n octopus
kubectl rollout restart deployment/octopus -n octopus
```

## 5. 升级步骤

```bash
# 方法一:修改 deployment.yaml 中的镜像 tag 后重新 apply
# 方法二:直接更新镜像
kubectl set image deployment/octopus octopus=ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1.25.0 -n octopus

# 方法三:滚动重启(相同 tag,用于拉取最新镜像)
kubectl rollout restart deployment/octopus -n octopus

# 监控滚动更新状态
kubectl rollout status deployment/octopus -n octopus

# 查看发布历史
kubectl rollout history deployment/octopus -n octopus

# 回滚到上一版本
kubectl rollout undo deployment/octopus -n octopus
```

> **数据库兼容**:GORM AutoMigrate 只增不减,不提供回滚迁移。升级前务必备份数据。详见 [`docs/RUNBOOK.md`](../../docs/RUNBOOK.md) 5.2 节。

## 6. 卸载

```bash
# 删除所有资源(保留 namespace)
kubectl delete -f deploy/k8s/

# 彻底删除 namespace(含所有资源与 PVC 数据)
kubectl delete namespace octopus
```

> **警告**:`kubectl delete namespace octopus` 会删除 PVC 及其数据,操作前请确认已备份。

## 7. 生产建议

1. **数据库**:SQLite 仅适用于单副本测试/小规模部署。生产环境请切换到 MySQL 或 PostgreSQL,支持多副本水平扩容。
2. **HPA(水平 Pod 自动扩缩)**:切换到 MySQL/PG 后,配置 HPA 基于 CPU/内存自动扩容:

   ```yaml
   apiVersion: autoscaling/v2
   kind: HorizontalPodAutoscaler
   metadata:
     name: octopus
     namespace: octopus
   spec:
     scaleTargetRef:
       apiVersion: apps/v1
       kind: Deployment
       name: octopus
     minReplicas: 2
     maxReplicas: 10
     metrics:
       - type: Resource
         resource:
           name: cpu
           target:
             type: Utilization
             averageUtilization: 70
   ```

3. **StorageClass**:显式指定适合集群的 StorageClass(如 SSD-backed),而非依赖默认。
4. **Ingress TLS**:取消 `ingress.yaml` 中 TLS 配置的注释,配合 cert-manager 自动签发证书:

   ```bash
   # 使用 cert-manager 自动签发 Let's Encrypt 证书
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
   ```

5. **Prometheus 监控**:Deployment 已包含 `prometheus.io/scrape` 注解,可直接被 Prometheus 抓取。若使用 Prometheus Operator,补充 ServiceMonitor(见 [`docs/RUNBOOK.md`](../../docs/RUNBOOK.md) 1.3.5 节)。
6. **密钥管理**:使用 sealed-secrets、external-secrets 或 Vault 管理敏感配置,避免明文 Secret。
7. **资源配额**:根据实际负载调整 `resources.requests` 与 `resources.limits`,生产建议起步 `requests: { cpu: 500m, memory: 512Mi }`。
8. **网络策略**:配置 NetworkPolicy 限制出入站流量,仅允许 Ingress Controller 与必要的上游 API 访问。
