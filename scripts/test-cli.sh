#!/usr/bin/env bash
# CLI unit tests + smoke (all platforms). Does not build Wails GUI.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
export CGO_ENABLED=0

echo "== go test (internal + CLI) =="
go test ./internal/... ./cmd/procpane/...

echo "== build CLI =="
cli_bin=bin/procpane
case "$(uname -s 2>/dev/null || echo)" in
  MINGW* | MSYS* | CYGWIN*) cli_bin=bin/procpane.exe ;;
esac
go build -o "$cli_bin" ./cmd/procpane
export PROCPANE_BIN="$cli_bin"

echo "== native smoke =="
go run ./scripts/smoke
