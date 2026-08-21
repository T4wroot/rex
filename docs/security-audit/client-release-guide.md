# Client, Release, and Claims Guide

## Coverage

Reviewed all files under `rex-client/`, examples, `pyproject.toml`, `README.md`, `README_FA.md`, `CHANGELOG.md`, `Dockerfile`, `apt-setup.sh`, `rex-node/install.sh`, and `.github/workflows/build.yml`.

## Findings

### High — Python SDK defaults to insecure transport

`rex-client/client.py:41-56` defaults `tls=False`; `direct_client.py:45-55` has no TLS support at all for port 7444. The server's raw TCP listener is always plaintext (`server.go:60-79`). The token is therefore sent as a bearer credential over cleartext unless an external secure tunnel is used.

### High — SDK exposes a capability model broader than the legacy policy

`RXPDirectClient` exposes `spawn_pty`, `native_file_op`, and `sysinfo` (`direct_client.py:124-215`). These map to server handlers that do not enforce the command/path policy. The SDK's “zero local subprocess” property is real for the direct client, but it shifts all trust to the remote daemon and does not make the operation safe.

### Medium — direct-client response handling can mis-handle errors

`direct_client.py:223-240` resolves any pending future matching `stream_id` before checking opcode. A server `OP_ERROR` for a pending operation is delivered as if it were a successful response; `spawn_pty` (`:145-146`) does not validate the response opcode or payload, so a failed spawn can return a session object. Error futures and pending requests are not comprehensively failed when the reader loop exits.

### Medium — malformed server frames can leave client operations hanging

`protocol.py:47-64` uses `readexactly` and raises on truncated input. `direct_client.py:223-245` catches the exception but does not resolve all pending futures with a connection error. Callers can wait until their outer timeout, and cleanup behavior depends on the caller.

### Medium — SDK lacks bounds and backpressure for remote output

`ServerPTYSession.read_output` accumulates arbitrary chunks in a queue (`direct_client.py:265-275`). There is no maximum output size, queue bound, or server-side PTY output quota. A remote process can fill memory or stall the connection.

### High — release and installer artifacts are stale or contradictory

The repository tag/README describes RXP/2.0 and port 7444, but `main.go:9-14` reports v1.0.0/RXP/1.0. The GitHub workflow (`.github/workflows/build.yml:32-36`) builds binaries without tests; its image tags (`:49-56`) use `v1.0.0`. `install.sh:45-49` downloads v1.0.1 and configures only 7443. `Dockerfile:20` exposes only 7443. This makes it unclear which artifact corresponds to the documented protocol.

### High — distribution has no artifact integrity verification

`install.sh:45-49` follows a GitHub release redirect and installs the result without checksum/signature verification. `apt-setup.sh:14-20` further encourages downloading executable shell content from the mutable `master` branch and piping it to Bash.

### Medium — container is not hardened

`Dockerfile:11-22` runs with no `USER`, capabilities, read-only root filesystem, resource limits, or explicit 7444 exposure. It copies `config.yaml` with TLS disabled and a placeholder token (`rex-node/config.yaml:6,12`). If run with host privileges, the remote execution service becomes a container breakout/high-impact target.

### Medium — advertised performance/security claims are not reproducible

README claims sub-2 ms execution, specific CPU/memory numbers, TLS 1.3 transport, and zero local footprint. The repository contains no benchmark program, protocol test suite, integration test, or CI security test. The zero-local-subprocess claim is supported by the direct Python client implementation, but the latency/resource/TLS claims are not verified and TLS is disabled by default.

## Protocol compatibility map

| Client | Server path | Encoding | Default port | TLS |
|---|---|---|---:|---|
| `REXClient` | `/rex` WebSocket | JSON/RXP 1.0 | 7443 | Optional, off |
| `RXPDirectClient` | raw TCP | 8-byte RXP 2.0 frames | 7444 | None in SDK/server |

## Required release gates

Add protocol/auth fuzz tests, integration tests for every authorization mode, bounded-output tests, TLS/mTLS tests, race tests with real assertions, SBOM/dependency scanning, signed artifacts/checksums, version generated from one source, and a deployment smoke test that verifies both documented ports and protocols.
