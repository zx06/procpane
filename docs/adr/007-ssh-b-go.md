# ADR 007: SSH HostSource (B-go)

## Status

Accepted (2026-11-04) — [issue #21](https://github.com/zx06/procpane/issues/21)

## Context

Remote LISTEN via SSH without a persistent agent on the host.

## Decision

**B-go**: connect with `golang.org/x/crypto/ssh`, resolve hosts via OpenSSH config (`github.com/kevinburke/ssh_config` + `~/.ssh/config`).

Remote collection fallback: **`ss` → `lsof` → `netstat`**. Process fields enriched best-effort (`ps`, `/proc/pid/cmdline`).

Auth: **SSH agent and key files only** — **no password prompts**.

Enable gate: only aliases toggled in `~/.config/procpane/state.yaml` (XDG) are listed by `procpane ssh ls` and the GUI **SSH** tab.

Kill is **not** supported on SSH rows.

## Consequences

- Host/port/user are **not** edited in procpane settings (read from ssh_config).
- Match blocks and exotic directives may be unsupported by the parser (documented limitation).
- Host key verification uses insecure ignore for MVP; production hardening deferred.
