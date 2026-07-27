#!/usr/bin/env bash
# REX One-Line Super Installer (Linux / Ubuntu / Debian / CentOS)
# Usage: curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash

set -eo pipefail

TOKEN=""
PORT="7443"

# Parse optional args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --token) TOKEN="$2"; shift 2 ;;
    --port)  PORT="$2";  shift 2 ;;
    *) shift ;;
  esac
done

if [[ -z "$TOKEN" ]]; then
  TOKEN=$(openssl rand -hex 16 2>/dev/null || date +%s | md5sum | head -c 32)
fi

echo "⚡ Installing REX Node Daemon..."

# Check architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  BINARY_NAME="rex-node-linux-amd64" ;;
  aarch64) BINARY_NAME="rex-node-linux-arm64" ;;
  *) echo "Unsupported Architecture: $ARCH"; exit 1 ;;
esac

mkdir -p /usr/local/bin /etc/rex

# Stop existing running daemon if updating
systemctl stop rex-node 2>/dev/null || true

# Download pre-compiled binary
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
if [[ ! -f /etc/rex/allowlist.yaml ]]; then
cat > /etc/rex/allowlist.yaml << 'EOF'
mode: "autonomous"
denied_commands:
  - "rm -rf /"
  - "chmod -R 777 /"
  - "mkfs"
EOF
fi

# 4. Setup CLI helper tool `/usr/local/bin/rex`
cat > /usr/local/bin/rex << 'EOF'
#!/bin/bash
case "$1" in
  status)  systemctl status rex-node --no-pager ;;
  start)   systemctl start rex-node ;;
  stop)    systemctl stop rex-node ;;
  restart) systemctl restart rex-node ;;
  logs)    journalctl -u rex-node -f ;;
  config)  cat /etc/rex/config.yaml ;;
  token)   grep "token:" /etc/rex/config.yaml | awk '{print $2}' ;;
  *)
    echo "REX CLI Management Tool"
    echo "Usage: rex {status|start|stop|restart|logs|config|token}"
    ;;
esac
EOF
chmod +x /usr/local/bin/rex

# 5. Setup Systemd Service
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
echo "===================================================="
echo " ✅ REX Node installed and running successfully!"
echo "===================================================="
echo " ├─ Status:  active (running) on port ${PORT}"
echo " └─ Token:   ${TOKEN}"
echo "===================================================="
echo " 💡 Useful Commands:"
echo "    • rex status   (Check node status)"
echo "    • rex logs     (Stream live logs)"
echo "    • rex token    (Print authentication token)"
echo "===================================================="
