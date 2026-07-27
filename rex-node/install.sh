#!/usr/bin/env bash
# REX One-Line Super Installer (Linux / Ubuntu / Debian / CentOS)
# Usage: curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash

set -e

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

# Detect public IP address safely with fallback methods
SERVER_IP=$(curl -s -4 --connect-timeout 3 ifconfig.me || curl -s -4 --connect-timeout 3 api.ipify.org || hostname -I | awk '{print $1}' || echo "UNKNOWN_IP")

mkdir -p /usr/local/bin /etc/rex

# Stop existing running daemon if updating
systemctl stop rex-node 2>/dev/null || true
pkill -f rex-node 2>/dev/null || true
rm -f /usr/local/bin/rex-node.tmp

# Download pre-compiled binary safely via GitHub Release redirects
DOWNLOAD_URL="https://github.com/T4wroot/rex/releases/download/v1.0.1/${BINARY_NAME}"
curl -L -f -s -S "$DOWNLOAD_URL" -o /usr/local/bin/rex-node.tmp
chmod +x /usr/local/bin/rex-node.tmp
mv -f /usr/local/bin/rex-node.tmp /usr/local/bin/rex-node

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
SERVER_IP=$(curl -s -4 --connect-timeout 3 ifconfig.me || hostname -I | awk '{print $1}')
case "$1" in
  status)
    systemctl status rex-node --no-pager
    ;;
  start)
    systemctl start rex-node
    echo "✅ REX Node started."
    ;;
  stop)
    systemctl stop rex-node
    echo "🛑 REX Node stopped."
    ;;
  restart)
    systemctl restart rex-node
    echo "🔄 REX Node restarted."
    ;;
  logs)
    journalctl -u rex-node -f
    ;;
  config)
    cat /etc/rex/config.yaml
    ;;
  token)
    grep "token:" /etc/rex/config.yaml | awk '{print $2}'
    ;;
  ip)
    echo "$SERVER_IP"
    ;;
  info)
    TOKEN=$(grep "token:" /etc/rex/config.yaml | awk '{print $2}')
    echo "=========================================="
    echo " 📍 REX Node Info"
    echo "=========================================="
    echo " Server IP: $SERVER_IP"
    echo " Port:      7443"
    echo " Token:     $TOKEN"
    echo "=========================================="
    ;;
  uninstall|remove)
    echo "⚠️ Removing REX Node from server..."
    systemctl stop rex-node 2>/dev/null || true
    systemctl disable rex-node 2>/dev/null || true
    rm -f /etc/systemd/system/rex-node.service
    systemctl daemon-reload
    rm -rf /usr/local/bin/rex-node /usr/local/bin/rex /etc/rex
    echo "🗑️ REX Node has been completely uninstalled and removed."
    ;;
  *)
    echo "REX CLI Management Tool"
    echo "Usage: rex {status|start|stop|restart|logs|info|token|ip|uninstall}"
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

# Open firewall port 7443 if ufw or iptables exist
ufw allow 7443/tcp 2>/dev/null || iptables -A INPUT -p tcp --dport 7443 -j ACCEPT 2>/dev/null || true

echo ""
echo "===================================================="
echo " ✅ REX Node installed and running successfully!"
echo "===================================================="
echo " ├─ Server IP: ${SERVER_IP}"
echo " ├─ Port:      ${PORT}"
echo " └─ Token:     ${TOKEN}"
echo "===================================================="
echo " 💡 Useful Commands:"
echo "    • rex status    (Check node status)"
echo "    • rex stop      (Stop node)"
echo "    • rex restart   (Restart node)"
echo "    • rex uninstall (Completely remove REX)"
echo "    • rex info      (Print IP & Token)"
echo "===================================================="
