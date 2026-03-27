#!/usr/bin/env bash
#
# 将 gistlike 项目部署到 HostDare VPS。
# 流程：本地 rsync 同步 → 远程清理旧部署 → Docker 构建 → 启动应用 + Caddy → 健康检查。
#
# 环境变量：
#   DEPLOY_HOST    — SSH 主机别名（默认 hostdare，需在 ~/.ssh/config 中配置）
#   DEPLOY_APP_DIR — 远程部署目录（默认 /home/gistlike）

set -euo pipefail

# ---------- 配置 ----------
DEPLOY_HOST="${DEPLOY_HOST:-hostdare}"
DEPLOY_APP_DIR="${DEPLOY_APP_DIR:-/home/gistlike}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ---------- 前置检查 ----------
if ! command -v rsync >/dev/null 2>&1; then
  echo "rsync is required but not installed." >&2
  exit 1
fi

if ! command -v ssh >/dev/null 2>&1; then
  echo "ssh is required but not installed." >&2
  exit 1
fi

# ---------- 1. 同步代码到远程 ----------
echo "==> Syncing repository to ${DEPLOY_HOST}:${DEPLOY_APP_DIR}"
ssh "${DEPLOY_HOST}" "mkdir -p '${DEPLOY_APP_DIR}'"

# --delete 保证远程与本地一致；排除运行时数据和无关文件
rsync -az --delete \
  --exclude '.git/' \
  --exclude 'data/' \
  --exclude '.DS_Store' \
  --exclude '.codex/' \
  "${REPO_ROOT}/" "${DEPLOY_HOST}:${DEPLOY_APP_DIR}/"

# ---------- 2. 远程执行部署（通过 heredoc 传入脚本） ----------
echo "==> Deploying on ${DEPLOY_HOST}"
ssh "${DEPLOY_HOST}" /bin/bash <<EOF
set -euo pipefail

APP_DIR="${DEPLOY_APP_DIR}"

# 停掉指定目录下的 docker compose 服务（忽略错误）
cleanup_compose_dir() {
  local dir="\$1"
  if [ -f "\${dir}/docker-compose.yml" ]; then
    echo "==> docker compose down in \${dir}"
    (cd "\${dir}" && docker compose down) || true
  fi
}

# 清理历史遗留的旧部署路径
cleanup_compose_dir "/home/caddy"
cleanup_compose_dir "/root/gistlike"
cleanup_compose_dir "/opt/gistlike"

echo "==> Removing stale deployment directories"
rm -rf /home/caddy /root/gistlike /opt/gistlike

# 初始化数据目录；UID 1000 对应容器内非 root 用户
mkdir -p "\${APP_DIR}/data"
rm -rf "\${APP_DIR}/.git"
chown -R 1000:1000 "\${APP_DIR}/data"

# ---------- 构建 & 启动 ----------
echo "==> Building gistlike image"
docker build -t gistlike-gistlike "\${APP_DIR}"

echo "==> Starting gistlike"
(cd "\${APP_DIR}" && docker compose up -d --no-build)

# Caddy 作为反向代理 / TLS 终端，独立 compose 文件
echo "==> Starting Caddy"
(cd "\${APP_DIR}/deploy/caddy" && docker compose up -d)

# ---------- 健康检查：最多等 20 秒 ----------
echo "==> Waiting for Caddy to accept traffic"
for _ in {1..20}; do
  if curl -fsSI -H 'Host: gist.xyzizz.xyz' http://127.0.0.1 >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

# ---------- 部署结果摘要 ----------
echo "==> Deployment status"
docker ps --format 'table {{.Names}}\t{{.Ports}}\t{{.Status}}'
echo "---"
ss -lntp | egrep ':80 |:443 |:8080 |:2020 ' || true
echo "---"
curl -I -H 'Host: gist.xyzizz.xyz' http://127.0.0.1
EOF
