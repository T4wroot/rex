# REX — Remote EXecution Protocol (RXP/1.0)
> **Ultra-fast, zero-overhead agent-to-server infrastructure control protocol for AI Agents.**  
> *Sub-5ms execution latency, persistent WebSocket communication, 3 security tiers, terminal-free operation.*

---

[![GitHub Release](https://img.shields.io/github/v/release/T4wroot/rex?style=flat-square&color=blue)](https://github.com/T4wroot/rex/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/T4wroot/rex/build.yml?branch=master&style=flat-square&label=build)](https://github.com/T4wroot/rex/actions)
[![Docker Image](https://img.shields.io/badge/Docker-ghcr.io%2Ft4wroot%2Frex%2Frex--node-blue?style=flat-square&logo=docker)](https://github.com/T4wroot/rex/pkgs/container/rex%2Frex-node)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

---

> 🇮🇷 **راهنمای فارسی:** برای مطالعه راهنمای کامل به زبان فارسی، به برگه [README_FA.md](README_FA.md) مراجعه کنید.

---

## 💡 What is REX?

**REX** (Remote EXecution Protocol) is an open-source, agent-native infrastructure control protocol designed specifically for AI Agents (such as Hermes, Antigravity, and autonomous LLM agents) to observe, manage, and execute commands on remote servers **without relying on SSH sessions, terminal emulators, or interactive shell overhead.**

### ⚡ Why REX over SSH?
- **Sub-5ms Latency:** Zero SSH handshake or TTY allocation overhead. Commands run via native in-memory RPC frames.
- **Persistent Channel (`REXPersistentClient`):** Keeps a single background WebSocket connection open with automatic keep-alive ping/pong.
- **3-Tier Security Modes:** Choose between `autonomous`, `review`, and `allowlist` security levels dynamically.

---

## ⚡ Quick Start & Installation

### 1. Server Installation (`rex-node`)

Run this single zero-config command on any Linux server:

```bash
curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash
```

### 2. Manage Security Modes via CLI (`rex mode`)

```bash
# Check current security mode
rex mode

# Switch to Autonomous Mode (Full Agent Freedom)
rex mode autonomous

# Switch to Review Mode (Read-only Auto, Dangerous Blocked)
rex mode review

# Switch to Strict Allowlist Mode
rex mode allowlist
```

---

### 3. Client Installation for AI Agents (`rex-client`)

```bash
pip install git+https://github.com/T4wroot/rex.git#subdirectory=rex-client
```

---

## 🚀 Python Usage in AI Agents

```python
import asyncio
from rex_client import REXPersistentClient

async def main():
    agent = REXPersistentClient(host="80.253.254.207", token="YOUR_TOKEN", port=7443)
    await agent.start()

    # Instant Execution (<5ms)
    res = await agent.exec_fast("systemctl restart xray")
    print(f"Status: {res.exit_code}, Output: {res.stdout}")

    # Fetch Hardware Metrics
    info = await agent.sysinfo_fast()
    print(f"RAM: {info.mem_used_gb}GB / {info.mem_total_gb}GB | Load: {info.load_1m}")

    await agent.stop()

asyncio.run(main())
```

---

## 🤝 Contributing & Community Proposals

We welcome contributions, feature proposals, and feedback from the community!

### How to Contribute:
- **Propose Ideas & Report Issues:** Open a new issue in [GitHub Issues](https://github.com/T4wroot/rex/issues).
- **Code Contributions:** Fork the repository, create a branch, and submit a Pull Request.
- **Language SDKs:** Help build client SDKs for Node.js, Rust, or Go.

---

## 📄 License
This project is licensed under the **MIT License**.
