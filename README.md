# REX — Remote EXecution Protocol (RXP/1.0)
> **Ultra-fast, zero-overhead agent-to-server infrastructure control protocol for AI Agents.**  
> *Sub-5ms execution latency, persistent WebSocket communication, 3 security tiers, terminal-free operation.*

---

[![GitHub Release](https://img.shields.io/github/v/release/T4wroot/rex?style=flat-square&color=blue)](https://github.com/T4wroot/rex/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/T4wroot/rex/build.yml?branch=master&style=flat-square&label=build)](https://github.com/T4wroot/rex/actions)
[![Docker Image](https://img.shields.io/badge/Docker-ghcr.io%2Ft4wroot%2Frex%2Frex--node-blue?style=flat-square&logo=docker)](https://github.com/T4wroot/rex/pkgs/container/rex%2Frex-node)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

---

## ⚡ Quick Start & One-Line Installation

### 1. Server Installation (`rex-node`)

Run this single zero-config command on any Linux server (Ubuntu, Debian, CentOS, etc.):

```bash
curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash
```

*(This automatically generates a secure token, downloads the binary, sets up `/etc/rex/config.yaml`, and starts a `systemd` background service on port 7443).*

Optionally, you can specify your own token:
```bash
curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash -s -- --token YOUR_TOKEN
```

---

### 2. Client Installation for AI Agents (`rex-client`)

Install the REX Python SDK in your AI Agent environment:

```bash
pip install git+https://github.com/T4wroot/rex.git#subdirectory=rex-client
```

---

### 3. Run via Official Docker Image (`ghcr.io`)

```bash
docker run -d \
  --name rex-node \
  --restart always \
  -p 7443:7443 \
  ghcr.io/t4wroot/rex/rex-node:latest
```

---

## 🚀 How AI Agents Use REX (Zero Terminal Overhead)

```python
import asyncio
from rex_client import REXPersistentClient

async def main():
    # Connect ONCE at agent startup (Background WebSocket Channel)
    agent = REXPersistentClient(host="167.172.102.14", token="YOUR_TOKEN", port=7443)
    await agent.start()

    # Instant execution (<5ms background RPC)
    res = await agent.exec_fast("systemctl restart xray")
    print(f"Status: {res.exit_code}, Output: {res.stdout}")

    # Fetch Hardware Metrics
    info = await agent.sysinfo_fast()
    print(f"RAM: {info.mem_used_gb}GB / {info.mem_total_gb}GB | Load: {info.load_1m}")

    await agent.stop()

asyncio.run(main())
```

---

## 🛡️ 3-Tier Security Levels (`/etc/rex/allowlist.yaml`)

REX supports 3 operational security modes:

```yaml
# Modes: "autonomous" | "review" | "allowlist"
mode: "autonomous"

# Protected Rules — ALWAYS enforced in ALL modes (including autonomous)
denied_commands:
  - "rm -rf /"
  - "chmod -R 777 /"
  - "mkfs"
```

1. **`autonomous` (Default):** Full agent freedom — runs commands instantly while enforcing absolute safety bans.
2. **`review`:** Read-only inspection commands execute automatically; dangerous operations flag for human approval.
3. **`allowlist`:** Strict mode — only pre-configured commands execute.

---

## 📄 License
This project is licensed under the **MIT License**.
