# ADR 006: Docker HostSource (A1 published ports)

## Status

Accepted (2026-11-04) — [issue #21](https://github.com/zx06/procpane/issues/21)

## Context

v0.2 adds Docker as an independent HostSource. Options included netns/sidecar (A2) vs Engine API published ports (A1).

## Decision

Use **A1**: `ContainerList` and each container's **published** port mappings (`Ports` / host `PublicPort`). No netns inspection, no sidecar.

Implementation: official Moby API client (`github.com/moby/moby/client`), honoring `DOCKER_HOST` and default unix socket.

## Consequences

- Rows reflect **published** ports only (not every in-container LISTEN).
- Kill is **not** supported on Docker rows.
- CLI: `procpane docker ls` · GUI: **Docker** tab with empty/error when Engine unreachable.
- No Docker daemon does not affect local list/GUI tab (graceful empty + hint).
