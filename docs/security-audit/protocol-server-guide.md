# Protocol and Server Security Guide

## Coverage and data flow

`rex-node/main.go` loads YAML config and allowlist, then `server.go` starts a WebSocket listener on `port` and a raw TCP listener on `tcp_port`. TCP frames are decoded by `protocol.go`; authenticated frames dispatch to `pty_runtime.go`, `native_engine.go`, and `sysinfo.go`. WebSocket JSON frames dispatch to `executor.go`, `stream.go`, and sysinfo. The daemon runs with the OS privileges of its service process.

## Findings

### Critical — PTY bypasses command policy and grants arbitrary remote shell

**Evidence:** `server.go:125-149` accepts `OpPTYSpawn`, `OpPTYData`, resize, and close after only the connection-level token check. `pty_runtime.go:55-70` starts `$SHELL`/`/bin/bash`; `pty_runtime.go:106` writes attacker-controlled bytes into it. No `Allowlist.IsCommandAllowed` call occurs.

**Impact:** any holder of the token gets an interactive shell, including shell operators, pipelines, redirections, background jobs, and commands that autonomous/review/allowlist mode would reject. If the daemon is root, compromise is host-wide.

**Fix:** make PTY an explicitly separate capability, disabled by default; enforce an authorization policy before spawn and per command/session; run under a dedicated unprivileged account, namespace/sandbox the session, cap CPU/memory/processes, and kill the process group on close/timeout.

### Critical — native file RPC bypasses path policy and permits arbitrary read/write/delete

**Evidence:** `server.go:154-158` calls `HandleNativeFileOp` directly. `native_engine.go:47-105` accepts attacker-selected paths and implements `os.ReadFile`, `os.WriteFile`, and `os.RemoveAll`. There is no `IsPathAllowed` check. The only path policy use found is `stream.go:33`.

**Impact:** token holder can exfiltrate secrets, overwrite configuration/binaries, or recursively delete arbitrary paths accessible to the daemon. Symlink and race protections are absent.

**Fix:** centralize authorization before every file operation; use an allowlisted root with `openat2`/no-follow semantics or equivalent, reject traversal and symlink escapes, constrain write/delete actions separately, enforce size limits, and never run the service as root.

### High — plaintext raw TCP and optional plaintext WebSocket expose bearer tokens and commands

**Evidence:** `config.go:26-30` defaults TLS false; `server.go:54-57` uses `ListenAndServe` unless TLS fields are configured; `server.go:110-123` sends the token directly in RXP payload; `rex_config.yaml:4` sets `tls: false`. `StartTCP` has no TLS path at all (`server.go:60-79`).

**Impact:** network observers can capture the bearer token and replay it to obtain shell/file control. README's TLS 1.3 badge is not true for the default deployment and does not cover port 7444.

**Fix:** require TLS 1.3 or a mutually authenticated tunnel for both protocols; fail closed when certificates are absent; use short-lived, scoped credentials with rotation and replay protection.

### High — unauthenticated resource exhaustion

**Evidence:** `server.go:71-77` accepts unlimited TCP connections and launches one goroutine per connection; after authentication, each connection can spawn multiple PTYs (`pty_runtime.go:35-40`) and `StartTCP` has no connection/read deadline. WebSocket upgrade/server paths likewise have no rate, body, connection, or session limits.

**Impact:** memory, file-descriptor, process, PTY, and goroutine exhaustion. A remote unauthenticated attacker can hold sockets; an authenticated attacker can amplify the impact with PTYs and streams.

**Fix:** listener limits, accept throttling, handshake deadlines, per-IP and per-token quotas, maximum PTYs/streams, bounded queues, idle timeouts, output limits, and OS-level cgroups/rlimits.

### High — WebSocket origin protection is disabled

**Evidence:** `server.go:20-22` sets `CheckOrigin` to always return true.

**Impact:** if the WebSocket endpoint is reachable from a browser context or exposed through a proxy, cross-site requests are not rejected. This is defense-in-depth failure around a powerful control endpoint.

**Fix:** allow only configured origins, or remove browser-facing WebSocket support and require a non-browser client with mTLS.

### Medium — path-prefix check is not a directory containment check

**Evidence:** `allowlist.go:97-109` authorizes with `strings.HasPrefix(cleanPath, allowedDir)`. `/var/logs/x` matches an allowed `/var/log` prefix. It also does not resolve symlinks.

**Impact:** stream authorization can expose files outside the intended directory; symlink swaps can escape after authorization.

**Fix:** use canonical root-relative path checks and no-follow file opening, preferably descriptor-based.

### Medium — malformed/version-invalid frames are insufficiently rejected

**Evidence:** `protocol.go:50-71` validates magic but does not reject an unsupported `Version`; it allocates the declared payload up to 65535 bytes. `server.go` does not impose read deadlines or per-client frame budgets.

**Impact:** protocol confusion and low-cost repeated allocation/connection pressure. This is not the primary RCE issue but harms reliability.

**Fix:** require exact supported version, validate opcode/stream state, set deadlines, cap aggregate bytes and outstanding requests, and close on protocol violations.

## Reliability notes

- `pty_runtime.go:168-185` treats PTY EOF/EIO as a logged read error and calls `Close`; lifecycle behavior is not covered by tests.
- `stream.go:59-85` ignores scanner errors, including the default 64 KiB token limit, so long log lines can silently terminate streaming.
- There are no Go test files, protocol fuzz tests, or authorization tests.
