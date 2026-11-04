#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
CGO_ENABLED=0 go build -o bin/procpane ./cmd/procpane
echo "built bin/procpane"
