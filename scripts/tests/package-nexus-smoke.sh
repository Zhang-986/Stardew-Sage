#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
version="0.3.0"
temp_root="$(mktemp -d "${TMPDIR:-/tmp}/echofarm-package-test.XXXXXX")"
trap 'rm -rf "$temp_root"' EXIT

mkdir -p "$temp_root/mod"
printf 'fixture mod binary\n' >"$temp_root/mod/EchoFarm.Mod.dll"
printf 'fixture bridge binary\n' >"$temp_root/mod/EchoFarm.Bridge.dll"

bash "$repo_root/scripts/package-nexus.sh" \
  --version "$version" \
  --mod-build-dir "$temp_root/mod" \
  --output-dir "$temp_root/out" \
  --nexus-mod-id 12345

expected=(
  "EchoFarm-$version-windows-x64.zip"
  "EchoFarm-$version-linux-x64.zip"
  "EchoFarm-$version-macos-x64.zip"
  "EchoFarm-$version-macos-arm64.zip"
)

for archive in "${expected[@]}"; do
  test -s "$temp_root/out/$archive"
done

test "$(wc -l <"$temp_root/out/SHA256SUMS.txt" | tr -d ' ')" = "4"

windows_listing="$(unzip -Z1 "$temp_root/out/EchoFarm-$version-windows-x64.zip")"
grep -qx 'EchoFarm/core/echofarm-core.exe' <<<"$windows_listing"
if grep -qx 'EchoFarm/core/echofarm-core' <<<"$windows_listing"; then
  echo "Windows archive contains a Unix core executable" >&2
  exit 1
fi

linux_listing="$(unzip -Z1 "$temp_root/out/EchoFarm-$version-linux-x64.zip")"
grep -qx 'EchoFarm/core/echofarm-core' <<<"$linux_listing"
if grep -qx 'EchoFarm/core/echofarm-core.exe' <<<"$linux_listing"; then
  echo "Linux archive contains a Windows core executable" >&2
  exit 1
fi

unzip -p "$temp_root/out/EchoFarm-$version-windows-x64.zip" EchoFarm/manifest.json \
  | jq -e '.UpdateKeys == ["Nexus:12345"]' >/dev/null

echo "Nexus package smoke test passed."
