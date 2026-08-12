"""
rex_client — Remote EXecution Protocol (RXP) Python Client

Usage (RXP/2.0 Direct TCP Server Runtime):
    from rex_client import RXPDirectClient

    async with RXPDirectClient("1.2.3.4", token="my-token", port=7444) as client:
        session = await client.spawn_pty()
        await session.send("cd /var/www\n")
        await session.send("pwd\n")
        print(await session.read_output())
"""

from rex_client.client import REXClient
from rex_client.direct_client import RXPDirectClient, ServerPTYSession
from rex_client.models import ExecResult, SysinfoResult, HandshakeInfo
from rex_client.exceptions import REXError, REXAuthError, REXCommandDenied, REXTimeout

__version__ = "2.0.0"
__all__ = [
    "REXClient",
    "RXPDirectClient",
    "ServerPTYSession",
    "ExecResult",
    "SysinfoResult",
    "HandshakeInfo",
    "REXError",
    "REXAuthError",
    "REXCommandDenied",
    "REXTimeout",
]
