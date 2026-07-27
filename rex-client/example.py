"""
Example: REX Client (Remote EXecution Protocol) usage in AI Agents.
"""

import asyncio
import logging
from rex_client import REXClient

logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s: %(message)s")


async def main():
    async with REXClient(
        host="YOUR_SERVER_IP",
        token="YOUR_SECRET_TOKEN",
        port=7443,
    ) as node:

        print(f"Connected to node: {node.server_info.node_id}")
        print(f"OS: {node.server_info.os}")
        print(f"Protocol: {node.server_info.protocol}")
        print()

        # Ping
        rtt = await node.ping()
        print(f"Ping: {rtt}ms")
        print()

        # System Metrics
        info = await node.sysinfo()
        print(f"Metrics:")
        print(f"  CPUs:    {info.cpus}")
        print(f"  Memory:  {info.mem_used_gb}GB / {info.mem_total_gb}GB")
        print(f"  Load:    {info.load_1m:.2f} {info.load_5m:.2f} {info.load_15m:.2f}")
        print(f"  Uptime:  {info.uptime_hours}h")
        print()

        # Command Execution
        cmds = [
            "df -h",
            "uptime",
            "systemctl status xray",
        ]
        for cmd in cmds:
            result = await node.exec(cmd, timeout=15)
            status = "✓" if result.ok else "✗"
            print(f"[{status}] {cmd}")
            if result.stdout:
                print(f"    stdout: {result.stdout[:200]}")
            if result.stderr:
                print(f"    stderr: {result.stderr[:200]}")
            print()

        # Stream Logs
        print("Streaming log sample:")
        count = 0
        async for line in node.stream("/var/log/xray/access.log", tail=10):
            print(f"  {line}")
            count += 1
            if count >= 10:
                break


if __name__ == "__main__":
    asyncio.run(main())
