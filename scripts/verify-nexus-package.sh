#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "usage: $0 <archive.zip> <runtime-id> <version>" >&2
  exit 2
fi

archive="$1"
runtime_id="$2"
version="$3"

test -s "$archive" || { echo "archive is missing or empty: $archive" >&2; exit 1; }
command -v unzip >/dev/null || { echo "unzip is required" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

listing="$(unzip -Z1 "$archive")"
while IFS= read -r entry; do
  [[ "$entry" == EchoFarm/* ]] || { echo "archive entry is outside EchoFarm/: $entry" >&2; exit 1; }
  [[ "$entry" != *../* ]] || { echo "archive contains an unsafe path: $entry" >&2; exit 1; }
done <<<"$listing"

required=(
  "EchoFarm/manifest.json"
  "EchoFarm/EchoFarm.Mod.dll"
  "EchoFarm/EchoFarm.Bridge.dll"
  "EchoFarm/LICENSE"
  "EchoFarm/README.md"
)
for entry in "${required[@]}"; do
  grep -qx "$entry" <<<"$listing" || { echo "archive is missing $entry" >&2; exit 1; }
done

expected_core="EchoFarm/core/echofarm-core"
if [[ "$runtime_id" == windows-* ]]; then
  expected_core+=".exe"
fi
grep -qx "$expected_core" <<<"$listing" || { echo "archive is missing $expected_core" >&2; exit 1; }

core_count="$(grep -Ec '^EchoFarm/core/echofarm-core(\.exe)?$' <<<"$listing")"
[[ "$core_count" -eq 1 ]] || { echo "archive must contain exactly one core executable" >&2; exit 1; }

if grep -Eqi '(^|/)(\.env([^/]*)?|[^/]*\.(db|sqlite|log|pdb)|Stardew Valley\.dll|StardewModdingAPI\.dll)$' <<<"$listing"; then
  echo "archive contains a forbidden secret, runtime-data, debug, or game file" >&2
  exit 1
fi

manifest="$(unzip -p "$archive" EchoFarm/manifest.json)"
jq -e --arg version "$version" '
  .Name == "EchoFarm" and
  .Version == $version and
  .UniqueID == "Zhang986.EchoFarm" and
  .EntryDll == "EchoFarm.Mod.dll"
' <<<"$manifest" >/dev/null

echo "verified $(basename "$archive")"
