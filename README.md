# REX — Remote EXecution Protocol (RXP/1.0)
> **Ultra-fast, zero-overhead agent-to-server infrastructure control protocol for AI Agents.**  
> *Sub-5ms execution latency, persistent WebSocket communication, 3 security tiers, terminal-free operation.*

---

[![GitHub Release](https://img.shields.io/github/v/release/T4wroot/rex?style=flat-square&color=blue)](https://github.com/T4wroot/rex/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/T4wroot/rex/build.yml?branch=master&style=flat-square&label=build)](https://github.com/T4wroot/rex/actions)
[![Docker Image](https://img.shields.io/badge/Docker-ghcr.io%2Ft4wroot%2Frex%2Frex--node-blue?style=flat-square&logo=docker)](https://github.com/T4wroot/rex/pkgs/container/rex%2Frex-node)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](LICENSE)

---

## 🌐 زبان / Language
- [English](#english-documentation)
- [فارسی (Persian)](#راهنمای-فارسی)

---

<a name="english-documentation"></a>
# 🇬🇧 English Documentation

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

<hr />

<a name="راهنمای-فارسی"></a>
# 🇮🇷 راهنمای فارسی (Persian Documentation)

<div dir="rtl" style="font-family: 'Vazirmatn', sans-serif; line-height: 2.0; text-align: right;">

## 💡 پروتکل REX چیست؟

**REX** یک پروتکل ارتباطی امن و فوق‌العاده سریع بین **AI Agentها** (مانند هرمس، آنتی‌گریویتی و سایر ایجنت‌های هوشمند) و **سرورهای لینوکس** است که جایگزین وابستگی به SSH و ترمینال‌های سنتی می‌شود.

### ⚡ چرا REX به جای SSH؟
- **سرعت پاسخگویی زیر ۵ میلی‌ثانیه:** حذف دست دادن‌های سنگین SSH و TTY shell.
- **ارتباط زنده پس‌زمینه (Persistent Channel):** اتصال یک‌باره برقرار می‌ماند و پینگ/پونگ پس‌زمینه آن را زنده نگه می‌دارد.
- **۳ سطح دسترسی امنیتی دائم و آنی:** پشتیبانی از مودهای `autonomous` (خودمختار)، `review` (بازبینی) و `allowlist` (لیست سفید).

---

## ⚡ نصب و راه‌اندازی سریع

### ۱. نصب روی سرور لینوکس (`rex-node`)

تنها با اجرای این تک‌دستور روی سرور (Ubuntu, Debian, CentOS):

<div dir="ltr" style="background:#161b22; color:#e6edf3; padding:12px; border-radius:8px; border-left:4px solid #58a6ff; font-family:Consolas, monospace; font-size:15px; text-align:left;">
curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash
</div>

---

### ۲. تغییر آنی سطح دسترسی با ابزار CLI (`rex mode`)

روی سرور می‌توانید سطح دسترسی ایجنت را آنی تغییر دهید:

<div dir="ltr" style="background:#161b22; color:#e6edf3; padding:12px; border-radius:8px; border-left:4px solid #58a6ff; font-family:Consolas, monospace; font-size:15px; text-align:left;">

# مشاهده مود فعلی
rex mode

# تغییر به حالت خودمختار کامل (آزادی عمل ایجنت با مسدودی دستورات مخرب)
rex mode autonomous

# تغییر به حالت بازبینی (دستورات خواندنی آزاد، تغییردهنده مسدود)
rex mode review

# تغییر به حالت لیست سفید سخت‌گیرانه
rex mode allowlist

</div>

---

### ۳. نصب برای ایجنت پایتون (`rex-client`)

<div dir="ltr" style="background:#161b22; color:#e6edf3; padding:12px; border-radius:8px; border-left:4px solid #58a6ff; font-family:Consolas, monospace; font-size:15px; text-align:left;">
pip install git+https://github.com/T4wroot/rex.git#subdirectory=rex-client
</div>

---

## 🛠️ دستورات کاربردی ابزار CLI در سرور (`rex`)

<table dir="rtl" style="width:100%; border-collapse:collapse; text-align:right; direction:rtl;">
<thead>
<tr style="background:#2d2d2d;">
<th style="padding:10px; border:1px solid #444;">دستور</th>
<th style="padding:10px; border:1px solid #444;">توضیح عملکرد</th>
</tr>
</thead>
<tbody>
<tr>
<td style="padding:10px; border:1px solid #444;"><code>rex info</code></td>
<td style="padding:10px; border:1px solid #444;">چاپ مشخصات كامل IP سرور، پورت، مود و توکن امنیتی</td>
</tr>
<tr>
<td style="padding:10px; border:1px solid #444;"><code>rex mode autonomous</code></td>
<td style="padding:10px; border:1px solid #444;">تغییر آنی سطح دسترسی به خودمختار کامل</td>
</tr>
<tr>
<td style="padding:10px; border:1px solid #444;"><code>rex status</code></td>
<td style="padding:10px; border:1px solid #444;">مشاهده وضعیت زنده دیمون سرور</td>
</tr>
<tr>
<td style="padding:10px; border:1px solid #444;"><code>rex logs</code></td>
<td style="padding:10px; border:1px solid #444;">مشاهده آنلاین لاگ‌های اتصال ایجنت</td>
</tr>
<tr>
<td style="padding:10px; border:1px solid #444;"><code>rex uninstall</code></td>
<td style="padding:10px; border:1px solid #444;">حذف کامل و بی‌اثر ساختن REX از روی سرور</td>
</tr>
</tbody>
</table>

---

## 📄 لایسنس
پروژه REX تحت لایسنس **MIT** منتشر شده است.

</div>
