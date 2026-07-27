#!/usr/bin/env bash
# REX One-Line Super Installer (Linux / Ubuntu / Debian)
# Usage: curl -fsSL https://rexprotocol.dev/install.sh | bash -s -- --token YOUR_TOKEN

set -euo pipefail

TOKEN=""
PORT="7443"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --token) TOKEN="$2"; shift 2 ;;
    --port)  PORT="$2";  shift 2 ;;
    *) shift ;;
  esac
done

if [[ -z "$TOKEN" ]]; then
  TOKEN=$(openssl rand -hex 16 2>/dev/null || echo "rex_secret_token_$(date +%s)")
fi

echo "⚡ Installing REX Node Daemon..."

# Check architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  BINARY_NAME="rex-node-linux-amd64" ;;
  aarch64) BINARY_NAME="rex-node-linux-arm64" ;;
  *) echo "Unsupported Architecture: $ARCH"; exit 1 ;;
esac

# 1. Download pre-compiled binary directly from Releases
mkdir -p /usr/local/bin /etc/rex
curl -fsSL "https://github.com/T4wroot/rex/releases/download/v1.0.0/${BINARY_NAME}" -o /usr/local/bin/rex-node
chmod +x /usr/local/bin/rex-node

# 2. Write Config
cat > /etc/rex/config.yaml << EOF
token: "${TOKEN}"
port: ${PORT}
tls: false
allowlist: /etc/rex/allowlist.yaml
log_level: info
EOF

# 3. Write Default Allowlist (Autonomous Mode)
cat > /etc/rex/allowlist.yaml << 'EOF'
mode: "autonomous"
denied_commands:
  - "rm -rf /"
  - "chmod -R 777 /"
  - "mkfs"
EOF

# 4. Setup Systemd Service
cat > /etc/systemd/system/rex-node.service << EOF
[Unit]
Description=REX Node Daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/rex-node --config /etc/rex/config.yaml
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now rex-node

echo ""
echo "✅ REX Node installed and running successfully!"
echo " ├─ Status:  active (running) on port ${PORT}"
echo " └─ Token:   ${TOKEN}"
