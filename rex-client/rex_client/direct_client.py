"""
RXPDirectClient — High-performance Direct TCP Client for RXP/2.0 protocol.
Runs zero local subprocesses; communicates directly with the REX server runtime.
"""

import asyncio
import json
import logging
import struct
from typing import AsyncIterator, Callable, Dict, Optional

from rex_client.protocol import (
    OP_AGENT_GUIDE,
    OP_AUTH_HANDSHAKE,
    OP_ERROR,
    OP_NATIVE_FILE_OP,
    OP_NATIVE_SYSINFO,
    OP_PING,
    OP_PONG,
    OP_PTY_CLOSE,
    OP_PTY_DATA,
    OP_PTY_RESIZE,
    OP_PTY_SPAWN,
    VERSION_2,
    RXPFrame,
    read_rxp_frame,
)

logger = logging.getLogger("rex_direct_client")


class RXPDirectClient:
    """
    Direct TCP client for REX RXP/2.0 Protocol.

    Usage:
        async with RXPDirectClient("1.2.3.4", token="secret", port=7444) as client:
            session = await client.spawn_pty()
            await session.send("cd /var/www\n")
            await session.send("pwd\n")
            output = await session.read_output()
            print(output)
    """

    def __init__(
        self,
        host: str,
        token: str,
        port: int = 7444,
        timeout: float = 30.0,
    ):
        self.host = host
        self.port = port
        self.token = token
        self.timeout = timeout

        self._reader: Optional[asyncio.StreamReader] = None
        self._writer: Optional[asyncio.StreamWriter] = None
        self._next_stream_id = 1
        self._pty_listeners: Dict[int, asyncio.Queue] = {}
        self._pending_rpc: Dict[int, asyncio.Future] = {}
        self._listen_task: Optional[asyncio.Task] = None
        self._is_connected = False

    @property
    def is_connected(self) -> bool:
        return self._is_connected and self._writer is not None

    async def connect(self):
        """Establish direct TCP connection and authenticate via RXP/2.0."""
        if self.is_connected:
            return

        logger.info(f"Connecting to REX Server runtime at {self.host}:{self.port} (RXP/2.0)")
        self._reader, self._writer = await asyncio.open_connection(self.host, self.port)

        # Send AUTH_HANDSHAKE frame
        frame = RXPFrame(
            version=VERSION_2,
            opcode=OP_AUTH_HANDSHAKE,
            stream_id=0,
            payload=self.token.encode("utf-8"),
        )
        self._writer.write(frame.encode())
        await self._writer.drain()

        # Read Auth response frame
        res = await asyncio.wait_for(read_rxp_frame(self._reader), timeout=10.0)
        if res is None or res.opcode != OP_AUTH_HANDSHAKE:
            err_msg = res.payload.decode("utf-8", errors="ignore") if res else "No response"
            self._writer.close()
            await self._writer.wait_closed()
            raise ConnectionError(f"RXP/2.0 Auth failed: {err_msg}")

        self._is_connected = True
        self._listen_task = asyncio.create_task(self._read_loop())
        logger.info(f"Connected and authenticated to REX Server runtime at {self.host}:{self.port}")

    async def disconnect(self):
        """Close connection and background listeners."""
        self._is_connected = False
        if self._listen_task:
            self._listen_task.cancel()
            try:
                await self._listen_task
            except asyncio.CancelledError:
                pass
            self._listen_task = None

        if self._writer:
            self._writer.close()
            await self._writer.wait_closed()
            self._writer = None
            self._reader = None
        logger.info("Disconnected from REX Server runtime")

    async def __aenter__(self) -> "RXPDirectClient":
        await self.connect()
        return self

    async def __aexit__(self, *_):
        await self.disconnect()

    async def spawn_pty(self, cols: int = 80, rows: int = 24) -> "ServerPTYSession":
        """Spawn a persistent server-side PTY shell session."""
        stream_id = self._next_stream_id
        self._next_stream_id += 1

        queue: asyncio.Queue[bytes] = asyncio.Queue()
        self._pty_listeners[stream_id] = queue

        payload = struct.pack("!HH", cols, rows)
        fut = asyncio.Future()
        self._pending_rpc[stream_id] = fut

        frame = RXPFrame(
            version=VERSION_2,
            opcode=OP_PTY_SPAWN,
            stream_id=stream_id,
            payload=payload,
        )
        self._writer.write(frame.encode())
        await self._writer.drain()

        await asyncio.wait_for(fut, timeout=self.timeout)
        return ServerPTYSession(self, stream_id, queue)

    async def native_file_op(self, action: str, path: str, content: bytes = b"", mode: int = 0o644) -> dict:
        """Execute a direct file operation via native OS syscalls on the server."""
        stream_id = self._next_stream_id
        self._next_stream_id += 1

        req = {
            "action": action,
            "path": path,
            "mode": mode,
        }
        if content:
            req["content"] = list(content)

        payload = json.dumps(req).encode("utf-8")
        fut = asyncio.Future()
        self._pending_rpc[stream_id] = fut

        frame = RXPFrame(
            version=VERSION_2,
            opcode=OP_NATIVE_FILE_OP,
            stream_id=stream_id,
            payload=payload,
        )
        self._writer.write(frame.encode())
        await self._writer.drain()

        res_frame = await asyncio.wait_for(fut, timeout=self.timeout)
        return json.loads(res_frame.payload.decode("utf-8"))

    async def sysinfo(self) -> dict:
        """Fetch server hardware and OS status directly."""
        stream_id = self._next_stream_id
        self._next_stream_id += 1

        fut = asyncio.Future()
        self._pending_rpc[stream_id] = fut

        frame = RXPFrame(
            version=VERSION_2,
            opcode=OP_NATIVE_SYSINFO,
            stream_id=stream_id,
            payload=b"",
        )
        self._writer.write(frame.encode())
        await self._writer.drain()

        res_frame = await asyncio.wait_for(fut, timeout=self.timeout)
        return json.loads(res_frame.payload.decode("utf-8"))

    async def get_agent_guide(self) -> str:
        """Fetch the fundamental REX protocol rule directive."""
        stream_id = self._next_stream_id
        self._next_stream_id += 1

        fut = asyncio.Future()
        self._pending_rpc[stream_id] = fut

        frame = RXPFrame(
            version=VERSION_2,
            opcode=OP_AGENT_GUIDE,
            stream_id=stream_id,
            payload=b"",
        )
        self._writer.write(frame.encode())
        await self._writer.drain()

        res_frame = await asyncio.wait_for(fut, timeout=self.timeout)
        return res_frame.payload.decode("utf-8")

    async def _send_frame(self, frame: RXPFrame):
        if not self.is_connected or not self._writer:
            raise ConnectionError("Client is not connected")
        self._writer.write(frame.encode())
        await self._writer.drain()

    async def _read_loop(self):
        try:
            while self._is_connected and self._reader:
                frame = await read_rxp_frame(self._reader)
                if frame is None:
                    break

                if frame.stream_id in self._pending_rpc:
                    fut = self._pending_rpc.pop(frame.stream_id)
                    if not fut.done():
                        fut.set_result(frame)

                elif frame.opcode == OP_PTY_DATA:
                    if frame.stream_id in self._pty_listeners:
                        await self._pty_listeners[frame.stream_id].put(frame.payload)

                elif frame.opcode == OP_ERROR:
                    logger.error(f"Server error stream {frame.stream_id}: {frame.payload.decode('utf-8', 'ignore')}")

        except Exception as exc:
            if self._is_connected:
                logger.warning(f"RXP read loop ended: {exc}")


class ServerPTYSession:
    """Represents a persistent interactive PTY shell running on the remote server."""

    def __init__(self, client: RXPDirectClient, stream_id: int, queue: asyncio.Queue):
        self.client = client
        self.stream_id = stream_id
        self.queue = queue

    async def send(self, data: str):
        """Send input command/data directly into the server PTY shell STDIN."""
        frame = RXPFrame(
            version=VERSION_2,
            opcode=OP_PTY_DATA,
            stream_id=self.stream_id,
            payload=data.encode("utf-8"),
        )
        await self.client._send_frame(frame)

    async def read_output(self, timeout: float = 1.0) -> str:
        """Read accumulated stdout/stderr bytes from the server PTY shell."""
        chunks = []
        try:
            while True:
                chunk = await asyncio.wait_for(self.queue.get(), timeout=timeout)
                chunks.append(chunk)
        except asyncio.TimeoutError:
            pass

        return b"".join(chunks).decode("utf-8", errors="replace")

    async def close(self):
        """Close the server-side PTY shell session."""
        frame = RXPFrame(
            version=VERSION_2,
            opcode=OP_PTY_CLOSE,
            stream_id=self.stream_id,
            payload=b"",
        )
        await self.client._send_frame(frame)
        self.client._pty_listeners.pop(self.stream_id, None)
