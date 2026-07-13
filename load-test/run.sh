#!/usr/bin/env bash
#
# Octopus 负载测试统一执行入口
#
# 用法:
#   ./run.sh [chat|embeddings|stats|all]   # 默认 all
#
# 环境变量:
#   BASE_URL   - Octopus 网关地址(默认 http://localhost:1088)
#   API_KEY    - 中继 API Key(chat / embeddings 必填)
#   JWT_TOKEN  - 管理 API JWT Token(stats 必填)
#   K6_OUT     - k6 输出格式(默认空,可选 json://results/xxx.json)
#
# 输出:
#   - 实时控制台日志
#   - JSON 结果保存到 load-test/results/<scenario>-<timestamp>.json
#   - 文本汇总报告保存到 load-test/results/summary-<timestamp>.txt

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_DIR="${SCRIPT_DIR}/results"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"

BASE_URL="${BASE_URL:-http://localhost:1088}"
API_KEY="${API_KEY:-}"
JWT_TOKEN="${JWT_TOKEN:-}"

SCENARIO="${1:-all}"

# 颜色输出(非 TTY 时禁用)
if [[ -t 1 ]]; then
    COLOR_INFO='\033[36m'
    COLOR_WARN='\033[33m'
    COLOR_OK='\033[32m'
    COLOR_ERR='\033[31m'
    COLOR_RESET='\033[0m'
else
    COLOR_INFO=''; COLOR_WARN=''; COLOR_OK=''; COLOR_ERR=''; COLOR_RESET=''
fi

info()  { printf "${COLOR_INFO}[INFO]${COLOR_RESET} %s\n" "$*" >&2; }
warn()  { printf "${COLOR_WARN}[WARN]${COLOR_RESET} %s\n" "$*" >&2; }
ok()    { printf "${COLOR_OK}[ OK ]${COLOR_RESET} %s\n" "$*" >&2; }
err()   { printf "${COLOR_ERR}[ERR ]${COLOR_RESET} %s\n" "$*" >&2; }

# 1. 检查 k6 是否安装
check_k6() {
    if ! command -v k6 >/dev/null 2>&1; then
        err "未检测到 k6,请先安装:"
        cat >&2 <<'EOF'

  macOS (Homebrew):
    brew install k6

  Debian/Ubuntu:
    sudo apt install -y k6
    # 或使用官方源:
    # sudo gpg -k && sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415B614F6BAF778CEC4F369C49E8
    # echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
    # sudo apt update && sudo apt install k6

  Docker:
    docker run --rm -i grafana/k6 run - < load-test/k6-chat-completions.js

  Windows (Chocolatey):
    choco install k6

  其他平台: https://k6.io/docs/getting-started/installation/
EOF
        return 1
    fi
    local ver
    ver="$(k6 version 2>&1 | head -n1 || echo unknown)"
    info "已检测到 k6: ${ver}"
}

# 2. 校验凭据
check_credentials() {
    local scenario="$1"
    case "$scenario" in
        chat|embeddings)
            if [[ -z "$API_KEY" ]]; then
                err "运行 ${scenario} 需要环境变量 API_KEY(中继 API Key)"
                err "示例: API_KEY=sk-xxxx ./run.sh ${scenario}"
                return 1
            fi
            ;;
        stats)
            if [[ -z "$JWT_TOKEN" ]]; then
                err "运行 stats 需要环境变量 JWT_TOKEN(管理 API JWT)"
                err "获取方式: curl -sX POST ${BASE_URL}/api/v1/user/login -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"...\"}'"
                return 1
            fi
            ;;
        all)
            if [[ -z "$API_KEY" ]]; then
                warn "API_KEY 未设置,chat/embeddings 将失败"
            fi
            if [[ -z "$JWT_TOKEN" ]]; then
                warn "JWT_TOKEN 未设置,stats 将失败"
            fi
            ;;
        *)
            err "未知场景: ${scenario},应为 chat|embeddings|stats|all"
            return 1
            ;;
    esac
}

# 3. 执行单个 k6 脚本
run_k6() {
    local script_name="$1"
    local script_path="${SCRIPT_DIR}/${script_name}"
    local scenario_name="${script_name%.js}"
    local result_json="${RESULTS_DIR}/${scenario_name}-${TIMESTAMP}.json"

    if [[ ! -f "$script_path" ]]; then
        err "脚本不存在: ${script_path}"
        return 1
    fi

    mkdir -p "$RESULTS_DIR"

    info "开始执行 ${scenario_name} -> ${result_json}"

    # 同时输出到控制台 + JSON 文件
    if ! k6 run \
            -e BASE_URL="$BASE_URL" \
            -e API_KEY="$API_KEY" \
            -e JWT_TOKEN="$JWT_TOKEN" \
            --out "json=${result_json}" \
            --summary-export "${RESULTS_DIR}/${scenario_name}-summary-${TIMESTAMP}.json" \
            "$script_path"; then
        err "${scenario_name} 执行失败"
        return 1
    fi

    ok "${scenario_name} 完成,结果: ${result_json}"
}

# 4. 从 k6 JSON 结果聚合汇总(从 --summary-export 输出读取)
# k6 summary JSON 结构包含 metrics.*.{values.{p(50),p(95),p(99),rate},values.count} 字段
extract_metric() {
    local summary_file="$1"
    local metric_name="$2"
    local field="$3"

    # 使用 python3 解析 JSON;若无 python3 则退化为 grep
    if command -v python3 >/dev/null 2>&1; then
        python3 - "$summary_file" "$metric_name" "$field" <<'PY' 2>/dev/null || echo "N/A"
import json, sys
try:
    with open(sys.argv[1], 'r', encoding='utf-8') as f:
        data = json.load(f)
    m = data.get('metrics', {}).get(sys.argv[2], {})
    vals = m.get('values', {})
    print(vals.get(sys.argv[3], 'N/A'))
except Exception:
    print('N/A')
PY
    else
        echo "N/A (python3 缺失)"
    fi
}

# 5. 汇总报告
write_summary() {
    local summary_file="${RESULTS_DIR}/summary-${TIMESTAMP}.txt"
    : > "$summary_file"

    {
        echo "=============================================="
        echo " Octopus 负载测试汇总报告"
        echo " 时间: $(date '+%Y-%m-%d %H:%M:%S %z')"
        echo " BASE_URL: ${BASE_URL}"
        echo " 场景: ${SCENARIO}"
        echo "=============================================="
        echo

        for script in k6-chat-completions k6-embeddings k6-stats; do
            local scenario_name="${script#k6-}"
            local sfile="${RESULTS_DIR}/${script}-summary-${TIMESTAMP}.json"
            if [[ ! -f "$sfile" ]]; then
                continue
            fi
            echo "----------------------------------------------"
            echo " 场景: ${scenario_name}"
            echo "----------------------------------------------"
            echo " RPS(请求总数):        $(extract_metric "$sfile" http_reqs count)"
            echo " P50 延迟(ms):          $(extract_metric "$sfile" 'http_req_duration' 'p(50)')"
            echo " P95 延迟(ms):          $(extract_metric "$sfile" 'http_req_duration' 'p(95)')"
            echo " P99 延迟(ms):          $(extract_metric "$sfile" 'http_req_duration' 'p(99)')"
            echo " 错误率:                $(extract_metric "$sfile" 'http_req_failed' 'rate')"
            echo " 迭代次数:              $(extract_metric "$sfile" 'iterations' 'count')"
            echo " 峰值 VU:               $(extract_metric "$sfile" 'vus_max' 'value')"
            echo
        done

        echo "=============================================="
        echo " 详细 JSON:"
        ls -1 "${RESULTS_DIR}"/*-"${TIMESTAMP}".json 2>/dev/null | sed 's/^/   /'
        echo "=============================================="
    } >> "$summary_file"

    cat "$summary_file"
    ok "汇总报告已保存: ${summary_file}"
}

# 6. 主流程
main() {
    info "Octopus 负载测试启动"
    info "BASE_URL=${BASE_URL}"
    info "场景=${SCENARIO}"

    check_k6 || exit 1
    check_credentials "$SCENARIO" || exit 1

    local rc=0

    case "$SCENARIO" in
        chat)
            run_k6 k6-chat-completions.js || rc=1
            ;;
        embeddings)
            run_k6 k6-embeddings.js || rc=1
            ;;
        stats)
            run_k6 k6-stats.js || rc=1
            ;;
        all)
            run_k6 k6-chat-completions.js || rc=1
            run_k6 k6-embeddings.js || rc=1
            run_k6 k6-stats.js || rc=1
            ;;
    esac

    write_summary || true

    if [[ $rc -ne 0 ]]; then
        err "部分场景执行失败,请检查上方日志"
        exit $rc
    fi

    ok "全部场景执行完成"
}

main "$@"
