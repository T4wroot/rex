"""
rex_client — Remote EXecution Protocol (RXP) Python Client

Usage:
    from rex_client import REXClient

    async with REXClient("1.2.3.4", token="my-token") as node:
        result = await node.exec("systemctl status xray")
        print(result.stdout)
"""

from rex_client.client import REXClient
from rex_client.models import ExecResult, SysinfoResult, HandshakeInfo
from rex_client.exceptions import REXError, REXAuthError, REXCommandDenied, REXTimeout

__version__ = "1.0.0"
__all__ = [
    "REXClient",
    "ExecResult",
    "SysinfoResult",
    "HandshakeInfo",
    "REXError",
    "REXAuthError",
    "REXCommandDenied",
    "REXTimeout",
]
