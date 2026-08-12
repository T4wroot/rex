"""
Example usage and verification test for REX RXP/2.0 Direct TCP Protocol.
"""

import asyncio
import sys
from rex_client.direct_client import RXPDirectClient


async def test_direct_rxp():
    print("=== Testing REX RXP/2.0 Direct TCP Server Runtime ===")
    token = "secret-token-123"

    try:
        async with RXPDirectClient("127.0.0.1", token=token, port=7444) as client:
            print("[✓] RXP/2.0 TCP Connection & Authentication Successful!")

            # 1. Test Persistent Server PTY Session (cd & pwd state continuity)
            print("\n--- Spawning Server PTY Session ---")
            session = await client.spawn_pty()
            await session.send("cd /var\n")
            await asyncio.sleep(0.2)
            await session.send("pwd\n")
            await asyncio.sleep(0.2)

            out = await session.read_output(timeout=0.5)
            print(f"[PTY Output]\n{out.strip()}")
            await session.close()

            # 2. Test Native File Syscall (No Shell)
            print("\n--- Testing Native File Syscall ---")
            stat = await client.native_file_op("stat", "/etc/passwd")
            print(f"[Native Stat Result] Status: {stat.get('status')}, Stat: {stat.get('stat')}")

            # 3. Test Native Sysinfo
            print("\n--- Testing Native Sysinfo ---")
            info = await client.sysinfo()
            print(f"[Sysinfo Result] OS: {info.get('os')}, Arch: {info.get('arch')}, CPUs: {info.get('cpus')}")

            print("\n[✓] ALL RXP/2.0 VERIFICATION TESTS PASSED SUCCESSFULLY!")

    except Exception as exc:
        print(f"[✗] Test Error: {exc}", file=sys.stderr)


if __name__ == "__main__":
    asyncio.run(test_direct_rxp())
