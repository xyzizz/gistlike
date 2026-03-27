#!/usr/bin/env bash

set -euo pipefail

DEPLOY_HOST="${DEPLOY_HOST:-hostdare}"
DEPLOY_APP_DIR="${DEPLOY_APP_DIR:-/home/gistlike}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if ! command -v rsync >/dev/null 2>&1; then
  echo "rsync is required but not installed." >&2
  exit 1
fi

if ! command -v ssh >/dev/null 2>&1; then
  echo "ssh is required but not installed." >&2
  exit 1
fi

echo "==> Syncing repository to ${DEPLOY_HOST}:${DEPLOY_APP_DIR}"
ssh "${DEPLOY_HOST}" "mkdir -p '${DEPLOY_APP_DIR}'"

rsync -az --delete \
  --exclude '.git/' \
  --exclude 'data/' \
  --exclude '.DS_Store' \
  --exclude '.codex/' \
  "${REPO_ROOT}/" "${DEPLOY_HOST}:${DEPLOY_APP_DIR}/"

echo "==> Deploying on ${DEPLOY_HOST}"
ssh "${DEPLOY_HOST}" /bin/bash <<EOF
set -euo pipefail

APP_DIR="${DEPLOY_APP_DIR}"

cleanup_compose_dir() {
  local dir="\$1"
  if [ -f "\${dir}/docker-compose.yml" ]; then
    echo "==> docker compose down in \${dir}"
    (cd "\${dir}" && docker compose down) || true
  fi
}

cleanup_compose_dir "/home/caddy"
cleanup_compose_dir "/root/gistlike"
cleanup_compose_dir "/opt/gistlike"

echo "==> Removing stale deployment directories"
rm -rf /home/caddy /root/gistlike /opt/gistlike

mkdir -p "\${APP_DIR}/data"
rm -rf "\${APP_DIR}/.git"
chown -R 1000:1000 "\${APP_DIR}/data"

echo "==> Building gistlike image"
docker build -t gistlike-gistlike "\${APP_DIR}"

echo "==> Starting gistlike"
(cd "\${APP_DIR}" && docker compose up -d --no-build)

echo "==> Starting Caddy"
(cd "\${APP_DIR}/deploy/caddy" && docker compose up -d)

echo "==> Waiting for Caddy to accept traffic"
for _ in {1..20}; do
  if curl -fsSI -H 'Host: gist.xyzizz.xyz' http://127.0.0.1 >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "==> Deployment status"
docker ps --format 'table {{.Names}}\t{{.Ports}}\t{{.Status}}'
echo "---"
ss -lntp | egrep ':80 |:443 |:8080 |:2020 ' || true
echo "---"
curl -I -H 'Host: gist.xyzizz.xyz' http://127.0.0.1
EOF
