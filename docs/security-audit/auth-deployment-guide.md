# Authentication, Policy, and Deployment Guide

## Coverage

Reviewed `rex-node/allowlist.go`, `config.go`, `config.yaml`, `rex_config.yaml`, `allowlist.yaml`, `install.sh`, `apt-setup.sh`, `Dockerfile`, `main.go`, `server.go`, and service/deployment references.

## Findings

### Critical — the policy modes do not protect the most dangerous capabilities

`allowlist.go:50-94` only gates `Executor.Run`, used by the legacy WebSocket `exec` path. The raw TCP PTY path in `server.go:125-149` and native file path in `server.go:154-158` never consult this policy. Therefore “allowlist” and “review” do not mean allowlist/review for the advertised RXP/2.0 runtime.

### Critical — default deployment is autonomous and TLS-disabled

`install.sh:7-10` defaults to `MODE="autonomous"`; lines 52-67 write `tls: false`; `rex-node/allowlist.yaml:7` is also autonomous. `config.go:26-30` defaults TLS false. The README presents autonomous mode as full agent freedom, which is incompatible with a safe-by-default claim.

### High — static/shared example secrets are present

`rex-node/rex_config.yaml:1` contains `rex-agent-secret-token`; `rex-client/get_server_status.py:11` uses the same token; `example_rxp2.py:12` contains another static token. These are repository-visible bearer credentials and should be treated as compromised. `config.yaml:6` also ships a placeholder that may be deployed unchanged.

### High — bearer authentication has no confidentiality, rotation, expiry, scope, or replay resistance

`server.go:110-123` compares a raw token from the wire. Both transports can run without TLS. There is no per-operation authorization, token expiry, rotation, audience, nonce, or audit identity. Compromise of one token grants all capabilities on that node.

### High — installer executes remote content and downloads an unverified binary

`install.sh:45-49` downloads a release binary with `curl -L` and verifies neither checksum nor signature. `apt-setup.sh:14-20` downloads the installer and creates a helper that later pipes raw GitHub content directly into `bash`. This is a supply-chain and tag/branch compromise risk.

### High — installer can stop/kill services and writes a root service without hardening

`install.sh:41-43` stops the existing service and runs `pkill -f rex-node`; lines 157-173 create and enable a systemd service with no `User=`, `NoNewPrivileges=`, capability bounding, filesystem protection, resource limits, or restrictive umask. The daemon therefore commonly runs as root with host-wide authority.

### High — deployment artifacts are inconsistent and expose the wrong protocol

`install.sh:8` defaults to port 7443, downloads v1.0.1 (`:46`), and only opens 7443 (`:175-176`), while RXP/2.0 listens on 7444 in `config.go:27` and `server.go:60-69`. `Dockerfile:20` exposes only 7443 and copies a TLS-disabled placeholder config. `main.go:9-14` still reports v1.0.0/RXP/1.0. This creates version drift and can leave the intended control plane unreachable or unexpectedly exposed.

### High — WebSocket origin policy is allow-all

`server.go:20-22` returns true for every origin. This is especially dangerous for a browser-reachable control endpoint.

### Medium — command parsing is not a shell parser

`executor.go:85-112` splits only on spaces and handles quotes simplistically. It does not implement escapes, tabs, shell operators, or nested quoting. This creates behavior drift between legacy `exec` and PTY execution and can cause policy review to reason about a different command than the OS receives. The PTY remains the more serious bypass.

### Medium — path authorization uses unsafe prefix matching

`allowlist.go:102-108` uses a string prefix and does not resolve symlinks. An allowed `/var/log` can match `/var/logs`, and a symlink can escape the configured tree. Native file operations bypass this check entirely.

## Deployment baseline required before production

Use a dedicated unprivileged account, mTLS or a hardened TLS listener on both ports, default-deny policy, no static/example tokens, signed/version-pinned artifacts with checksums, systemd sandboxing and resource limits, host firewall restricted to management networks, and explicit authorization for each capability. Do not use `curl | bash`.
