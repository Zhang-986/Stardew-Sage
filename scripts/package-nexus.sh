#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version=""
game_path=""
mod_build_dir=""
output_dir=""
nexus_mod_id=""

usage() {
  echo "usage: $0 --version <semver> [--game-path <path> | --mod-build-dir <path>] [--output-dir <path>] [--nexus-mod-id <id>]" >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) version="${2:-}"; shift 2 ;;
    --game-path) game_path="${2:-}"; shift 2 ;;
    --mod-build-dir) mod_build_dir="${2:-}"; shift 2 ;;
    --output-dir) output_dir="${2:-}"; shift 2 ;;
    --nexus-mod-id) nexus_mod_id="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $1" >&2; usage; exit 2 ;;
  esac
done

[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?$ ]] || { echo "--version must be semantic version text" >&2; exit 2; }
if [[ -z "$game_path" && -z "$mod_build_dir" ]]; then
  echo "provide --game-path for a real build or --mod-build-dir for a packaging fixture" >&2
  exit 2
fi
if [[ -n "$game_path" && -n "$mod_build_dir" ]]; then
  echo "--game-path and --mod-build-dir are mutually exclusive" >&2
  exit 2
fi
if [[ -n "$nexus_mod_id" && ! "$nexus_mod_id" =~ ^[0-9]+$ ]]; then
  echo "--nexus-mod-id must be numeric" >&2
  exit 2
fi

for tool in go zip unzip jq; do
  command -v "$tool" >/dev/null || { echo "$tool is required" >&2; exit 1; }
done

dotnet="$repo_root/.tools/dotnet/dotnet"
if [[ ! -x "$dotnet" ]]; then
  dotnet="$(command -v dotnet || true)"
fi
[[ -x "$dotnet" ]] || { echo "dotnet SDK is required" >&2; exit 1; }

if [[ -n "$output_dir" && "$output_dir" != /* ]]; then
  output_dir="$repo_root/$output_dir"
fi
output_dir="${output_dir:-$repo_root/dist/nexus/$version}"

"$dotnet" test "$repo_root/stardew-echo-mod/EchoFarm.sln" --no-restore
(
  cd "$repo_root/echofarm-core"
  go test ./...
  go vet ./...
)

if [[ -n "$game_path" ]]; then
  "$dotnet" build "$repo_root/stardew-echo-mod/src/EchoFarm.Mod/EchoFarm.Mod.csproj" \
    --configuration Release \
    --property:"GamePath=$game_path" \
    --property:EnableModDeploy=false \
    --property:EnableModZip=false
  mod_dll="$(find "$repo_root/stardew-echo-mod/src/EchoFarm.Mod/bin/Release" -type f -name EchoFarm.Mod.dll -print -quit)"
  [[ -n "$mod_dll" ]] || { echo "EchoFarm.Mod.dll was not produced" >&2; exit 1; }
  mod_build_dir="$(dirname "$mod_dll")"
elif [[ "$mod_build_dir" != /* ]]; then
  mod_build_dir="$repo_root/$mod_build_dir"
fi

for file in EchoFarm.Mod.dll EchoFarm.Bridge.dll; do
  test -s "$mod_build_dir/$file" || { echo "missing built file: $mod_build_dir/$file" >&2; exit 1; }
done

mkdir -p "$output_dir"
temp_root="$(mktemp -d "${TMPDIR:-/tmp}/echofarm-package.XXXXXX")"
trap 'rm -rf "$temp_root"' EXIT

targets=(
  "windows-x64 windows amd64 echofarm-core.exe"
  "linux-x64 linux amd64 echofarm-core"
  "macos-x64 darwin amd64 echofarm-core"
  "macos-arm64 darwin arm64 echofarm-core"
)

for target in "${targets[@]}"; do
  read -r runtime_id goos goarch executable <<<"$target"
  core_output="$temp_root/core/$runtime_id/$executable"
  mkdir -p "$(dirname "$core_output")"
  (
    cd "$repo_root/echofarm-core"
    CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
      go build -trimpath -ldflags='-s -w -buildid=' -o "$core_output" ./cmd/echofarm
  )
  if [[ "$goos" != "windows" ]]; then
    chmod 755 "$core_output"
  fi

  package_root="$temp_root/package-$runtime_id/EchoFarm"
  mkdir -p "$package_root/core"
  cp "$mod_build_dir/EchoFarm.Mod.dll" "$package_root/"
  cp "$mod_build_dir/EchoFarm.Bridge.dll" "$package_root/"
  cp "$repo_root/LICENSE" "$package_root/LICENSE"
  cp "$repo_root/stardew-echo-mod/README.md" "$package_root/README.md"
  cp "$core_output" "$package_root/core/$executable"

  if [[ -n "$nexus_mod_id" ]]; then
    jq --arg version "$version" --arg update_key "Nexus:$nexus_mod_id" \
      '.Version = $version | .UpdateKeys = [$update_key]' \
      "$repo_root/stardew-echo-mod/src/EchoFarm.Mod/manifest.json" >"$package_root/manifest.json"
  else
    jq --arg version "$version" '.Version = $version' \
      "$repo_root/stardew-echo-mod/src/EchoFarm.Mod/manifest.json" >"$package_root/manifest.json"
  fi

  find "$package_root" -exec touch -t 202001010000 {} +
  archive="$output_dir/EchoFarm-$version-$runtime_id.zip"
  rm -f "$archive"
  (
    cd "$(dirname "$package_root")"
    find EchoFarm -type f -print | LC_ALL=C sort | zip -X -q "$archive" -@
  )
  bash "$repo_root/scripts/verify-nexus-package.sh" "$archive" "$runtime_id" "$version"
done

(
  cd "$output_dir"
  rm -f SHA256SUMS.txt
  if command -v sha256sum >/dev/null; then
    sha256sum EchoFarm-"$version"-*.zip >SHA256SUMS.txt
  else
    shasum -a 256 EchoFarm-"$version"-*.zip >SHA256SUMS.txt
  fi
)

echo "Nexus packages written to $output_dir"
