#!/usr/bin/env bash
# REX Node Installer (Remote EXecution Protocol)
# Usage: curl -fsSL https://your-host/install.sh | bash -s -- --token YOUR_TOKEN

set -euo pipefail

REX_VERSION="1.0.0"
REX_PORT="7443"
REX_TOKEN=""
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/rex"

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --token) REX_TOKEN="$2"; shift 2 ;;
    --port)  REX_PORT="$2";  shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

if [[ -z "$REX_TOKEN" ]]; then
  echo "Error: --token is required"
  echo "Usage: $0 --token YOUR_SECRET_TOKEN [--port 7443]"
  exit 1
fi

echo "=== REX Node Installer v${REX_VERSION} ==="

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac
OS="linux"

echo "[1/5] Architecture: ${OS}/${ARCH}"

echo "[2/5] Creating config directory..."
mkdir -p "$CONFIG_DIR"

echo "[3/5] Writing config..."
cat > "${CONFIG_DIR}/config.yaml" << EOF
token: "${REX_TOKEN}"
port: ${REX_PORT}
tls: false
allowlist: ${CONFIG_DIR}/allowlist.yaml
log_level: info
EOF

if [[ ! -f "${CONFIG_DIR}/allowlist.yaml" ]]; then
  echo "[4/5] Writing default allowlist..."
  cat > "${CONFIG_DIR}/allowlist.yaml" << 'ALLOWLIST'
allowed_commands:
  - "systemctl restart xray"
  - "systemctl start xray"
  - "systemctl stop xray"
  - "systemctl status xray"
  - "systemctl restart x-ui"
  - "systemctl status x-ui"
  - "docker restart *"
  - "docker ps"
  - "df -h"
  - "free -m"
  - "uptime"
  - "journalctl -u xray --no-pager -n *"

allowed_paths:
  - /var/log/xray/
  - /var/log/nginx/

denied_commands:
  - "rm -rf"
  - "passwd"
  - "shutdown"
  - "reboot"
ALLOWLIST
else
  echo "[4/5] Keeping existing allowlist."
fi

echo "[5/5] Creating systemd service..."
cat > /etc/systemd/system/rex-node.service << EOF
[Unit]
Description=REX Node - Remote EXecution Protocol Daemon
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/rex-node --config ${CONFIG_DIR}/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable rex-node

echo ""
echo "=== Installation complete ==="
echo "Binary:  ${INSTALL_DIR}/rex-node"
echo "Config:  ${CONFIG_DIR}/config.yaml"
echo "Port:    ${REX_PORT}"
echo ""
echo "Start:   systemctl start rex-node"
echo "Status:  systemctl status rex-node"
echo "Logs:    journalctl -u rex-node -f"
