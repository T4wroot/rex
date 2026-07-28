# 🇮🇷 راهنمای فارسی پروتکل REX (Remote EXecution Protocol)

> **لایه ارتباطی امن، فوق‌العاده سریع و بدون وابستگی به ترمینال بین AI Agentها و سرورها**  
> *زمان پاسخگویی زیر ۵ میلی‌ثانیه، اتصال زنده پس‌زمینه WebSocket، ۳ سطح دسترسی امنیتی، کاملاً بدون ترمینال.*

---

<div dir="rtl" style="font-family: 'Vazirmatn', sans-serif; line-height: 2.0; text-align: right;">

## 💡 پروتکل REX چیست؟

**REX** یک پروتکل ارتباطی اپن‌سورس و نیتیو برای **AI Agentها** (مانند هرمس، آنتی‌گریویتی و سایر ایجنت‌های هوشمند) است که به آن‌ها اجازه می‌دهد بدون نیاز به SSH، اختصاص TTY یا ترمینال‌های سنتی، سرورهای لینوکس را در پس‌زمینه با سرعت بالا مشاهده و مدیریت کنند.

### ⚡ چرا REX به جای SSH؟
- **سرعت پاسخگویی زیر ۵ میلی‌ثانیه:** حذف دست دادن‌های سنگین SSH و TTY shell.
- **ارتباط زنده پس‌زمینه (Persistent Channel):** اتصال یک‌باره برقرار می‌ماند و پینگ/پونگ پس‌زمینه آن را زنده نگه می‌دارد.
- **۳ سطح دسترسی امنیتی دائم و آنی:** پشتیبانی از مودهای `autonomous` (خودمختار)، `review` (بازبینی) و `allowlist` (لیست سفید).

---

## ⚡ نصب و راه‌اندازی سریع

### ۱. نصب روی سرور لینوکس (`rex-node`)

تنها با اجرای این تک‌دستور روی سرور (Ubuntu, Debian, CentOS):

```bash
curl -fsSL https://raw.githubusercontent.com/T4wroot/rex/master/rex-node/install.sh | bash
```

---

### ۲. ابزار مدیریت خط فرمان روی سرور (`rex`)

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
<td style="padding:10px; border:1px solid #444;">تغییر آنی سطح دسترسی به خودمختار کامل (آزادی عمل ایجنت با مسدودی دستورات مخرب)</td>
</tr>
<tr>
<td style="padding:10px; border:1px solid #444;"><code>rex mode review</code></td>
<td style="padding:10px; border:1px solid #444;">تغییر به حالت بازبینی (دستورات خواندنی آزاد، تغییردهنده مسدود)</td>
</tr>
<tr>
<td style="padding:10px; border:1px solid #444;"><code>rex mode allowlist</code></td>
<td style="padding:10px; border:1px solid #444;">تغییر به حالت لیست سفید سخت‌گیرانه</td>
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

### ۳. نصب برای ایجنت پایتون (`rex-client`)

```bash
pip install git+https://github.com/T4wroot/rex.git#subdirectory=rex-client
```

---

## 🚀 نمونه کد پایتون در ایجنت

```python
import asyncio
from rex_client import REXPersistentClient

async def main():
    # 1. اتصال زنده در پس‌زمینه
    agent = REXPersistentClient(host="80.253.254.207", token="YOUR_TOKEN", port=7443)
    await agent.start()

    # 2. اجرای آنی دستور بدون ترمینال (<5ms)
    res = await agent.exec_fast("systemctl restart xray")
    print(f"خروجی: {res.stdout}")

    # 3. دریافت سخت‌افزار سرور
    info = await agent.sysinfo_fast()
    print(f"حافظه: {info.mem_used_gb}GB / {info.mem_total_gb}GB")

    await agent.stop()

asyncio.run(main())
```

---

## 🤝 مشارکت و توسعه (Contributing)

پروژه **REX** به صورت ۱۰۰٪ اپن‌سورس توسعه داده می‌شود. از تمام توسعه‌دهندگان، ایده‌پردازان و علاقه‌مندان به دنیای AI Agents و زیرساخت دعوت می‌کنیم تا در ارتقای این پروتکل همراه ما باشند!

### چطور مشارکت کنید؟
- **پیشنهاد ایده یا گزارش باگ:** از بخش [GitHub Issues](https://github.com/T4wroot/rex/issues) اقدام کنید.
- **توسعه کد:** پروژه را Fork کرده و Pull Request ارسال کنید.
- **توسعه SDK زبان‌های دیگر:** ایجاد کلاینت‌های Node.js / Rust / Go برای REX.

</div>
