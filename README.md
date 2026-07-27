# REX — Remote EXecution Protocol (RXP/1.0)
> **Ultra-fast, zero-overhead agent-to-server infrastructure control protocol for AI Agents.**  
> *Sub-5ms execution latency, persistent WebSocket communication, 3 security tiers, terminal-free operation.*

---

[![GitHub Release](https://img.shields.io/github/v/release/T4wroot/rex?style=flat-square&color=blue)](https://github.com/T4wroot/rex/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/T4wroot/rex/build.yml?branch=master&style=flat-square&label=build)](https://github.com/T4wroot/rex/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/T4wroot/rex?filename=rex-node%2Fgo.mod&style=flat-square)](https://go.dev)
[![Python Version](https://img.shields.io/badge/Python-3.10%2B-blue?style=flat-square&logo=python)](https://python.org)

---

## 💡 What is REX?

**REX** (Remote EXecution Protocol) is a high-craft, agent-native infrastructure control protocol designed specifically for AI Agents (like Hermes, Antigravity, and autonomous LLM agents) to manage remote Linux servers **without relying on SSH sessions, TTY allocation, or terminal emulators.**

### ❌ The Problem with SSH for AI Agents
Standard SSH was designed for interactive human shell sessions. Every time an agent executes a command via SSH, it pays a heavy penalty:
- Full TLS/SSH handshake overhead (800ms - 1500ms)
- Interactive PTY/TTY allocation delays
- Raw unstructured stdout/stderr parsing
- Binary security model (All root or nothing)

### ✅ The REX Architecture
```
+-------------------------------------------------------------+
|                     AI Agent (Hermes/AGY)                   |
|                   rex_client / REXPersistentClient           |
+------------------------------+------------------------------+
                               |
                               |  Persistent WebSocket Channel (RXP/1.0)
                               |  Latency: ~3ms | Pure In-Memory RPC
                               |
+------------------------------v------------------------------+
|                     rex-node (Go Daemon)                    |
|                Security Mode: Autonomous / Review           |
+------------------------------+------------------------------+
                               |
                               | Internal OS Calls & Systemd
                               v
                     +-------------------+
                     |   Linux Server    |
                     +-------------------+
```

---

## ✨ Key Features

- **⚡ Sub-5ms Execution Latency:** Eliminates SSH shell creation overhead. Commands execute directly via Go native subprocesses in ~3-5ms.
- **🔄 Persistent WebSocket Channel (`REXPersistentClient`):** Connects ONCE upon agent startup, maintaining a zero-delay background connection with automatic keep-alive ping/pong.
- **🛡️ 3-Tier Security Levels (`Mode`):**
  1. `autonomous`: Full agent freedom — runs commands instantly while enforcing absolute safety bans (`rm -rf /`).
  2. `review`: Safe read-only commands execute automatically; dangerous operations flag for human approval.
  3. `allowlist`: Strict mode where only pre-configured commands execute.
- **📊 Native System Metrics:** Fetches RAM, CPU load, disk usage, and uptime structured without calling `top` or `free`.
- **📡 Real-Time Log Streaming:** Stream `/var/log/xray/access.log` line-by-line asynchronously over WebSocket.

---

## ⚡ Quick Start

### 1. Server Installation (`rex-node`)

Run this single command on your remote Linux server:

```bash
curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash -s -- --token YOUR_SECRET_TOKEN --port 7443
```

*(This automatically downloads the binary, writes `/etc/rex/config.yaml`, and creates a `systemd` background service).*

### 2. Client Installation (`rex-client`)

Install the REX Python SDK in your AI Agent environment:

```bash
pip install git+https://github.com/T4wroot/rex.git#subdirectory=rex-client
```

---

## 🚀 Usage in AI Agents

### Persistent Connection (Recommended for Agents)

```python
import asyncio
from rex_client import REXPersistentClient

async def main():
    # 1. Connect ONCE at agent startup
    agent = REXPersistentClient(host="167.172.102.14", token="YOUR_SECRET_TOKEN", port=7443)
    await agent.start()

    # 2. Instant execution (<5ms, background, no terminal)
    res = await agent.exec_fast("systemctl restart xray")
    print(f"Status: {res.exit_code}, Output: {res.stdout}")

    # 3. Query System Metrics instantly
    info = await agent.sysinfo_fast()
    print(f"RAM: {info.mem_used_gb}GB / {info.mem_total_gb}GB | Load: {info.load_1m}")

    await agent.stop()

asyncio.run(main())
```

---

## 📡 Protocol Specification (RXP/1.0)

REX communicates using bidirectional JSON frames over WebSocket:

| Frame Type | Direction | Description |
|---|---|---|
| `handshake` | Client → Node | Token authentication and client identification |
| `handshake_ack` | Node → Client | Server metadata, OS arch, and capabilities |
| `exec` | Client → Node | Command execution frame with optional timeout |
| `exec_result` | Node → Client | Exit code, stdout, stderr, and duration in `ms` |
| `sysinfo` | Client → Node | Request structured system hardware metrics |
| `sysinfo_result` | Node → Client | JSON payload containing RAM, CPU, Load, Uptime |
| `stream` | Client → Node | Request asynchronous tail log streaming |
| `stream_line` | Node → Client | Live log line frame |
| `ping` / `pong` | Both | Keep-alive heartbeat frames |

---

## 🛡️ Security Configuration (`/etc/rex/allowlist.yaml`)

```yaml
# Options: "autonomous" | "review" | "allowlist"
mode: "autonomous"

# Protected Rules — Always enforced in ALL modes including autonomous
denied_commands:
  - "rm -rf /"
  - "chmod -R 777 /"
  - "mkfs"
  - "dd if=/dev/zero"
```

---

## 📄 License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.
