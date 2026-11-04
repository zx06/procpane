#!/usr/bin/env bash
# CLI unit tests + smoke (Linux CI / headless). Does not build Wails GUI.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
export CGO_ENABLED=0

echo "== go test (internal + CLI) =="
go test ./internal/... ./cmd/procpane/...

echo "== build CLI =="
go build -o bin/procpane ./cmd/procpane

echo "== list --json =="
./bin/procpane list --json | jq -e '.listeners | type' >/dev/null

echo "== kill 1 refused =="
if ./bin/procpane kill 1 2>/dev/null; then
  echo "FAIL: kill 1 should refuse" >&2
  exit 1
fi

PORT="${PROCPANE_TEST_PORT:-18765}"
python3 -m http.server "$PORT" --bind 127.0.0.1 >/dev/null 2>&1 &
srv=$!
trap 'kill "$srv" 2>/dev/null || true' EXIT
sleep 0.4

echo "== find :$PORT =="
./bin/procpane find ":$PORT" | grep -q ":$PORT"

pid=$(pgrep -f "http.server $PORT" | head -1)
echo "== show $pid =="
./bin/procpane show "$pid" | jq -e ".pid == $pid" >/dev/null

echo "CLI smoke OK"
