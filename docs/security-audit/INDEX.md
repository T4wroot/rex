# REX Security Audit Index

**Scope:** source review of commit `c73af70` (`v2.0.0`), plus local build, vet, race-instrumented test, Python compilation, shell syntax, and RXP/2.0 loopback smoke checks.

## Repository map

| Area | Guide | Primary trust boundary |
|---|---|---|
| `rex-node/` wire protocol, TCP/WebSocket server, PTY, native handlers | [protocol-server-guide.md](protocol-server-guide.md) | Authenticated network peer to privileged server process |
| `rex-node/` policy, config, installer, container/service deployment | [auth-deployment-guide.md](auth-deployment-guide.md) | Token/config/install supply chain to daemon privileges |
| `rex-client/`, examples, CI, Docker, release/docs claims | [client-release-guide.md](client-release-guide.md) | Agent SDK and distribution artifacts |

Top-level files reviewed: `README.md`, `README_FA.md`, `CHANGELOG.md`, `LICENSE`, `Dockerfile`, `apt-setup.sh`, `.github/workflows/build.yml`, `.gitignore`, and `debian-itp-submission.txt`.

## Executive verdict

**REX is not safe for production deployment as advertised.** It is a functional prototype for authenticated remote command/file control, but the token is effectively the only authorization boundary. TLS is disabled by default, the installer defaults to autonomous execution, PTY and native file operations bypass the configured command/path policy, and the daemon has no privilege drop or meaningful resource limits. A stolen/intercepted token can become arbitrary remote code execution and arbitrary file modification/deletion, especially if the daemon runs as root.

## Severity summary

- **Critical:** authenticated peer receives unrestricted PTY shell and unrestricted native file read/write/delete; token transport is plaintext by default.
- **High:** unsafe autonomous/default credentials and deployment; installer executes unsigned remote content and downloads an unverified binary; no service hardening; origin checks disabled; unauthenticated resource exhaustion.
- **Medium:** path-prefix authorization bypass and symlink weakness; release/version/port drift; SDK error/cleanup edge cases; no tests or benchmark evidence.

## Verification performed

- Go build: passed with the installed Go toolchain.
- `go vet ./...`: passed.
- `go test -race ./...`: passed but reported no test files.
- Python `compileall`: passed.
- `bash -n rex-node/install.sh apt-setup.sh`: passed.
- Local RXP/2.0 smoke test: authentication, guide, sysinfo, native `stat`, and PTY command execution passed.

Passing these checks validates basic operation, not security.
