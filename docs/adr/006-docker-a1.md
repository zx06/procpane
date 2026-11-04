# ADR 006: Docker HostSource (A1 published ports)

## Status

Accepted (2026-11-04) — [issue #21](https://github.com/zx06/procpane/issues/21)

## Context

v0.2 adds Docker as an independent HostSource. Options included netns/sidecar (A2) vs Engine API published ports (A1).

## Decision

Use **A1**: `ContainerList` and each container's **published** port mappings (`Ports` / host `PublicPort`). No netns inspection, no sidecar.

Implementation: official Engine API client (`github.com/docker/docker/client`). Endpoint selection:

- **No procpane context** in `state.yaml` → legacy `DOCKER_HOST` env, else `unix:///var/run/docker.sock` (unchanged).
- **Explicit procpane context** → parse local Docker context store (`github.com/docker/cli/cli/context/store`, same layout as `docker context`), dial that endpoint (incl. TLS).

GUI Settings exposes a context dropdown; CLI: `procpane config docker-contexts`, `procpane config docker-context`.

## Consequences

- Rows reflect **published** ports only (not every in-container LISTEN).
- Kill is **not** supported on Docker rows.
- CLI: `procpane docker ls` · GUI: **Docker** tab with empty/error when Engine unreachable.
- No Docker daemon does not affect local list/GUI tab (graceful empty + hint).
