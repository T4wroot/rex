"""
REX 2.0 Live Server Status Fetcher via RXPDirectClient (Raw TCP Socket)
"""

import asyncio
import json
import os
from rex_client.direct_client import RXPDirectClient


async def fetch_status():
    token = os.environ.get("REX_TOKEN")
    if not token:
        raise RuntimeError("Set REX_TOKEN to a deployment-specific token before running this example")
    host = "127.0.0.1"
    port = 7444

    status_report = {}

    print(f"Connecting to REX 2.0 Server Runtime at {host}:{port} via RXP/2.0 TCP Binary Protocol...")
    async with RXPDirectClient(host, token=token, port=port) as client:
        print("[✓] Connected & Authenticated via RXP/2.0 Binary Header!")

        # 1. Native Sysinfo RPC
        sys_info = await client.sysinfo()
        status_report["native_sysinfo"] = sys_info

        # 2. Native File Syscall (read /proc/loadavg)
        loadavg_res = await client.native_file_op("read", "/proc/loadavg")
        if loadavg_res.get("status") == "ok":
            import base64
            raw = loadavg_res.get("content", "")
            if isinstance(raw, str):
                content_bytes = base64.b64decode(raw)
            else:
                content_bytes = bytes(raw)
            status_report["load_avg"] = content_bytes.decode("utf-8").strip()

        # 3. Direct Server PTY Session Status Commands
        session = await client.spawn_pty()
        status_report["pty_session_id"] = session.stream_id

        # Execute uname -a
        await session.send("uname -a\n")
        await asyncio.sleep(0.3)
        uname_out = await session.read_output(timeout=0.5)

        # Execute free -h
        await session.send("free -h\n")
        await asyncio.sleep(0.3)
        free_out = await session.read_output(timeout=0.5)

        # Execute df -h /
        await session.send("df -h /\n")
        await asyncio.sleep(0.3)
        df_out = await session.read_output(timeout=0.5)

        await session.close()

        status_report["pty_commands"] = {
            "uname": uname_out.strip(),
            "free": free_out.strip(),
            "df": df_out.strip(),
        }

    print("\n=== LIVE REX 2.0 SERVER STATUS REPORT ===")
    print(json.dumps(status_report, indent=2, ensure_ascii=False))

    with open("server_status_result.json", "w", encoding="utf-8") as f:
        json.dump(status_report, f, indent=2, ensure_ascii=False)


if __name__ == "__main__":
    asyncio.run(fetch_status())
