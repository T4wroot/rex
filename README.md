# REX — Remote EXecution Protocol (RXP/1.0)
> **Ultra-fast, zero-overhead agent-to-server infrastructure control protocol for AI Agents.**  
> *Sub-5ms execution latency, persistent WebSocket communication, 3 security tiers, terminal-free operation.*

---

[![GitHub Release](https://img.shields.io/github/v/release/T4wroot/rex?style=flat-square&color=blue)](https://github.com/T4wroot/rex/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/T4wroot/rex/build.yml?branch=master&style=flat-square&label=build)](https://github.com/T4wroot/rex/actions)
[![Docker Image](https://img.shields.io/badge/Docker-ghcr.io%2Ft4wroot%2Frex%2Frex--node-blue?style=flat-square&logo=docker)](https://github.com/T4wroot/rex/pkgs/container/rex%2Frex-node)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

---

## ⚡ Quick Start & Installation

### Option A: Install via APT Package (`apt install`)

To install `rex-node` on any Ubuntu/Debian server using APT:

```bash
# 1. Download and install the pre-built APT package (.deb)
wget https://github.com/T4wroot/rex/releases/download/v1.0.0/rex-node_1.0.0_amd64.deb
sudo apt install ./rex-node_1.0.0_amd64.deb
```

Or run the single-line automated APT installer:

```bash
curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash
```

---

### Option B: Run Official Docker Container (`ghcr.io`)

Pull and run the official Docker image:

```bash
docker pull ghcr.io/t4wroot/rex/rex-node:latest

docker run -d \
  --name rex-node \
  --restart always \
  -p 7443:7443 \
  ghcr.io/t4wroot/rex/rex-node:latest
```

---

### Option C: Python SDK for AI Agents (`pip install`)

Install the client SDK in your Python Agent:

```bash
pip install git+https://github.com/T4wroot/rex.git#subdirectory=rex-client
```

---

## 🚀 Usage in AI Agents

```python
import asyncio
from rex_client import REXPersistentClient

async def main():
    agent = REXPersistentClient(host="167.172.102.14", token="YOUR_SECRET_TOKEN", port=7443)
    await agent.start()

    # Fast execution (<5ms background RPC)
    res = await agent.exec_fast("systemctl restart xray")
    print(f"Result: {res.stdout}")

    # Fetch Hardware Metrics
    info = await agent.sysinfo_fast()
    print(f"RAM: {info.mem_used_gb}GB / {info.mem_total_gb}GB")

    await agent.stop()

asyncio.run(main())
```

---

## 📄 License
This project is licensed under the **MIT License**.
