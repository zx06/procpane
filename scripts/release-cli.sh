#!/usr/bin/env bash
# Cross-compile CLI binaries (CGO_ENABLED=0) for release matrix targets.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
out="${1:-$root/dist/release-cli}"
mkdir -p "$out"

build_one() {
  local goos=$1 goarch=$2 name=$3
  echo "==> ${goos}/${goarch} -> ${name}"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$out/$name" ./cmd/procpane
}

build_one linux amd64 procpane-linux-amd64
build_one linux arm64 procpane-linux-arm64
build_one darwin amd64 procpane-darwin-amd64
build_one darwin arm64 procpane-darwin-arm64
build_one windows amd64 procpane-windows-amd64.exe
build_one windows arm64 procpane-windows-arm64.exe

(
  cd "$out"
  sha256sum ./* > SHA256SUMS
)
echo "CLI release artifacts in $out (see SHA256SUMS)"
