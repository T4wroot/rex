# REX (Remote EXecution Protocol)

REX یک لایه ارتباط امن و سبک بین **AI Agentها** و **سرورها** است که جایگزین وابستگی به SSH و ترمینال‌های سنتی می‌شود.

---

## ساختار پروژه

```
rex/
├── rex-node/          ← Go daemon روی سرورها (پورت 7443)
└── rex-client/        ← Python SDK برای Agentها
```

---

## نصب روی سرور (rex-node)

### ۱. Build باینری

```bash
cd rex-node
go mod tidy
GOOS=linux GOARCH=amd64 go build -o rex-node .
```

### ۲. نصب سریع با Script

```bash
# انتقال باینری به سرور
scp rex-node root@YOUR_SERVER:/usr/local/bin/

# اجرای اسکریپت نصب روی سرور
ssh root@YOUR_SERVER "bash -s" < install.sh --token YOUR_SECRET_TOKEN --port 7443
```

---

## نصب برای Agent (rex-client)

```bash
cd rex-client
pip install -e .
```

---

## نمونه کد استفاده در Agent

```python
import asyncio
from rex_client import REXClient

async def main():
    async with REXClient("1.2.3.4", token="YOUR_SECRET_TOKEN") as node:
        # ۱. دریافت اطلاعات سیستم
        info = await node.sysinfo()
        print(f"CPUs: {info.cpus}, RAM: {info.mem_total_gb}GB, Load: {info.load_1m}")

        # ۲. اجرای دستور (محدود به allowlist سرور)
        result = await node.exec("systemctl restart xray")
        if result.ok:
            print("Xray Service Restarted!")

        # ۳. دریافت آنلاین لاگ‌ها (Streaming)
        async for line in node.stream("/var/log/xray/access.log", tail=20):
            print(line)

asyncio.run(main())
```

---

## مشخصات پروتکل RXP/1.0

ارتباط روی WebSocket دوطرفه (مزیت: Persistent Connection بدون Overhead دست دادن SSH):

| Message Type   | Direction      | توضیح                        |
|----------------|----------------|------------------------------|
| `handshake`    | Client → Node  | احراز هویت با Token          |
| `handshake_ack`| Node → Client  | تأیید اتصال و ارسال مشخصات  |
| `exec`         | Client → Node  | اجرای دستور                 |
| `exec_result`  | Node → Client  | پاسخ خروجی، ExitCode و زمان  |
| `sysinfo`      | Client → Node  | استعلام وضعیت RAM/CPU/Uptime |
| `sysinfo_result`| Node → Client | اطلاعات ساختاریافته سیستم   |
| `stream`       | Client → Node  | شروع استریم لاگ              |
| `stream_line`  | Node → Client  | ارسال هر خط لاگ             |
| `ping`/`pong`  | Both           | تست Latency و زنده بودن       |
