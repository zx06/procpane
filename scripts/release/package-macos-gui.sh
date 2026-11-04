#!/usr/bin/env bash
# Build procpane.app + .dmg from a native macOS GUI binary (unsigned; CI release).
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
binary="${1:-$root/dist/release-app/procpane-app}"
out_dir="${2:-$root/dist/release-app}"
app_name="procpane.app"
dmg_name="procpane-app-darwin-arm64.dmg"
zip_name="procpane-app-darwin-arm64.zip"

if [ ! -f "$binary" ]; then
  echo "binary not found: $binary" >&2
  exit 1
fi

mkdir -p "$out_dir"
binary="$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary")"
out_dir="$(cd "$out_dir" && pwd)"
staging="$(mktemp -d)"
trap 'rm -rf "$staging"' EXIT

app_root="$staging/$app_name"
mkdir -p "$app_root/Contents/MacOS" "$app_root/Contents/Resources"
cp "$root/scripts/release/Info.plist.template" "$app_root/Contents/Info.plist"
install -m 755 "$binary" "$app_root/Contents/MacOS/procpane-app"

if command -v iconutil >/dev/null 2>&1 && command -v sips >/dev/null 2>&1; then
  iconset="$staging/AppIcon.iconset"
  mkdir -p "$iconset"
  src="$root/assets/icons/procpane-app-256.png"
  for size in 16 32 128 256 512; do
    sips -z "$size" "$size" "$src" --out "$iconset/icon_${size}x${size}.png" >/dev/null
    dbl=$((size * 2))
    sips -z "$dbl" "$dbl" "$src" --out "$iconset/icon_${size}x${size}@2x.png" >/dev/null
  done
  iconutil -c icns "$iconset" -o "$app_root/Contents/Resources/AppIcon.icns"
  /usr/libexec/PlistBuddy -c 'Add :CFBundleIconFile string AppIcon' "$app_root/Contents/Info.plist" 2>/dev/null \
    || /usr/libexec/PlistBuddy -c 'Set :CFBundleIconFile AppIcon' "$app_root/Contents/Info.plist"
fi

dmg_staging="$staging/dmg-root"
mkdir -p "$dmg_staging"
cp -R "$app_root" "$dmg_staging/"
ln -s /Applications "$dmg_staging/Applications"

rm -f "$out_dir/$dmg_name"
hdiutil create -volname "procpane" -srcfolder "$dmg_staging" -ov -format UDZO "$out_dir/$dmg_name"

rm -f "$out_dir/$zip_name"
( cd "$staging" && zip -9 -r "$out_dir/$zip_name" "$app_name" )

echo "macOS GUI packages in $out_dir: $dmg_name, $zip_name"
