#!/usr/bin/env bash
# Verify toolchain pins documented in README (run on 2026-11-04 baseline).
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
grep -q 'go 1.27.3' "$root/go.mod"
grep -q 'toolchain go1.27.3' "$root/go.mod"
grep -q '"svelte": "5.57.3"' "$root/frontend/package.json"
grep -q '"@wailsio/runtime": "3.1.0"' "$root/frontend/package.json"
grep -q 'v4.26.10' "$root/go.mod"
grep -q 'v1.10.2' "$root/go.mod"
echo "pins OK"
