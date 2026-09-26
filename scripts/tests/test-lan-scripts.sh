#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
initialize="$repo_root/scripts/lan/Initialize-EchoFarmLan.sh"
start="$repo_root/scripts/lan/Start-EchoFarmLanServer.sh"
windows_install="$repo_root/scripts/windows/Install-EchoFarmRelay.ps1"
windows_start="$repo_root/scripts/windows/Start-EchoFarmRelay.ps1"
temp_root="$(mktemp -d)"
trap 'rm -rf "$temp_root"' EXIT

lan_dir="$temp_root/private-lan"
output="$(ECHOFARM_LAN_DIR="$lan_dir" "$initialize" --host 127.0.0.1)"

test -s "$lan_dir/server.crt"
test -s "$lan_dir/server.key"
test -s "$lan_dir/relay.token"
test "$(wc -c <"$lan_dir/relay.token" | tr -d ' ')" = "65"
grep -Eq '^CERT_SHA256=[0-9a-f]{64}$' <<<"$output"
grep -Eq '^LAN_TOKEN=[0-9a-f]{64}$' <<<"$output"

if stat -f '%Lp' "$lan_dir/server.key" >/dev/null 2>&1; then
  key_mode="$(stat -f '%Lp' "$lan_dir/server.key")"
else
  key_mode="$(stat -c '%a' "$lan_dir/server.key")"
fi
test "$key_mode" = "600"

bash -n "$initialize" "$start"
if grep -Eq -- '--(api-key|model-api-key)' "$start"; then
  echo "model API key must not be accepted as a command-line argument" >&2
  exit 1
fi

test -s "$windows_install"
test -s "$windows_start"
grep -q 'Read-Host.*AsSecureString' "$windows_start"
if grep -Eq 'MODEL_(API_KEY|BASE_URL|NAME)' "$windows_install" "$windows_start"; then
  echo "Windows relay scripts must not handle model provider configuration" >&2
  exit 1
fi

echo "EchoFarm LAN setup scripts passed."
