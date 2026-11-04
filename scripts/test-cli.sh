#!/usr/bin/env bash
# CLI unit tests + smoke (all platforms). Does not build Wails GUI.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
export CGO_ENABLED=0

echo "== go test (internal + CLI) =="
go test ./internal/... ./cmd/procpane/...

echo "== build CLI =="
go build -o bin/procpane ./cmd/procpane

echo "== native smoke =="
go run ./scripts/smoke
