"""REX exception hierarchy."""


class REXError(Exception):
    """Base class for all REX errors."""


class REXAuthError(REXError):
    """Raised when the server rejects the authentication token."""


class REXCommandDenied(REXError):
    """Raised when the server denies the command due to allowlist policy."""


class REXTimeout(REXError):
    """Raised when a request times out."""


class REXConnectionError(REXError):
    """Raised when the WebSocket connection fails or drops."""


class REXProtocolError(REXError):
    """Raised when an unexpected or malformed message is received."""
