#!/usr/bin/env bash
# Build procpane.app + .dmg from a native macOS GUI binary (adhoc-signed with stable bundle id; CI release).
# Developer ID signing and notarization are out of scope for this script.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
readme_src="$root/scripts/release/macos-install-README.txt"
version_go="$root/internal/version/version.go"
binary="${1:-$root/dist/release-app/procpane-app}"
out_dir="${2:-$root/dist/release-app}"
app_name="procpane.app"
dmg_name="procpane-app-darwin-arm64.dmg"
zip_name="procpane-app-darwin-arm64.zip"
bundle_id="com.zx06.procpane"

if [ ! -f "$binary" ]; then
  echo "binary not found: $binary" >&2
  exit 1
fi
if [ ! -f "$readme_src" ]; then
  echo "install README not found: $readme_src" >&2
  exit 1
fi
if [ ! -f "$version_go" ]; then
  echo "version source not found: $version_go" >&2
  exit 1
fi

procpane_version="$(grep -E '^\s*const Version = ' "$version_go" | sed -n 's/.*"\([^"]*\)".*/\1/p')"
if [ -z "$procpane_version" ]; then
  echo "could not read Version from $version_go" >&2
  exit 1
fi

mkdir -p "$out_dir"
binary="$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary")"
out_dir="$(cd "$out_dir" && pwd)"
staging="$(mktemp -d)"
trap 'rm -rf "$staging"' EXIT

app_root="$staging/$app_name"
mkdir -p "$app_root/Contents/MacOS" "$app_root/Contents/Resources"
sed "s/__PROCPANE_VERSION__/$procpane_version/g" \
  "$root/scripts/release/Info.plist.template" > "$app_root/Contents/Info.plist"
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

if ! command -v codesign >/dev/null 2>&1; then
  echo "codesign not found (required for macOS GUI packaging)" >&2
  exit 1
fi
codesign --force --deep --sign - --identifier "$bundle_id" "$app_root"

dmg_staging="$staging/dmg-root"
mkdir -p "$dmg_staging"
cp -R "$app_root" "$dmg_staging/"
cp "$readme_src" "$dmg_staging/README.txt"
ln -s /Applications "$dmg_staging/Applications"

rm -f "$out_dir/$dmg_name"
hdiutil create -volname "procpane" -srcfolder "$dmg_staging" -ov -format UDZO "$out_dir/$dmg_name"

cp "$readme_src" "$staging/README.txt"
rm -f "$out_dir/$zip_name"
( cd "$staging" && zip -9 -r "$out_dir/$zip_name" "$app_name" README.txt )

echo "macOS GUI packages in $out_dir: $dmg_name, $zip_name (version $procpane_version, adhoc-signed as $bundle_id)"
