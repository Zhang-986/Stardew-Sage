#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
core_dir="$repo_root/echofarm-core"
temp_root="$(mktemp -d)"
core_log="$temp_root/core.log"
relay_log="$temp_root/relay.log"
wrong_relay_log="$temp_root/wrong-relay.log"

pick_port() {
  python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
}

cleanup() {
  for pid in "${wrong_relay_pid:-}" "${relay_pid:-}" "${core_pid:-}"; do
    if [[ -n "$pid" ]]; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
    fi
  done
  rm -rf "$temp_root"
}
trap cleanup EXIT

core_port="$(pick_port)"
relay_port="$(pick_port)"
wrong_relay_port="$(pick_port)"
identity_output="$(ECHOFARM_LAN_DIR="$temp_root/lan" "$repo_root/scripts/lan/Initialize-EchoFarmLan.sh" --host 127.0.0.1)"
token="$(sed -n 's/^LAN_TOKEN=//p' <<<"$identity_output")"
fingerprint="$(sed -n 's/^CERT_SHA256=//p' <<<"$identity_output")"
[[ "$token" =~ ^[0-9a-f]{64}$ && "$fingerprint" =~ ^[0-9a-f]{64}$ ]]

(
  cd "$core_dir"
  go build -trimpath -o "$temp_root/echofarm" ./cmd/echofarm
  go build -trimpath -o "$temp_root/echofarm-relay" ./cmd/echofarm-relay
)

ECHOFARM_MODEL_MODE=fixture \
ECHOFARM_ADDRESS="127.0.0.1:$core_port" \
ECHOFARM_DATABASE_PATH="$temp_root/echofarm.db" \
ECHOFARM_ALLOW_LAN=true \
ECHOFARM_LAN_TOKEN="$token" \
ECHOFARM_TLS_CERT_FILE="$temp_root/lan/server.crt" \
ECHOFARM_TLS_KEY_FILE="$temp_root/lan/server.key" \
"$temp_root/echofarm" >"$core_log" 2>&1 &
core_pid=$!

ECHOFARM_RELAY_ADDRESS="127.0.0.1:$relay_port" \
ECHOFARM_RELAY_UPSTREAM_ADDRESS="127.0.0.1:$core_port" \
ECHOFARM_RELAY_TOKEN="$token" \
ECHOFARM_RELAY_CERT_SHA256="$fingerprint" \
ECHOFARM_RELAY_TIMEOUT_SECONDS=10 \
"$temp_root/echofarm-relay" >"$relay_log" 2>&1 &
relay_pid=$!

base_url="http://127.0.0.1:$relay_port"
ready=false
for _ in {1..120}; do
  if curl --silent --fail "$base_url/healthz" >/dev/null; then
    ready=true
    break
  fi
  if ! kill -0 "$core_pid" 2>/dev/null || ! kill -0 "$relay_pid" 2>/dev/null; then
    break
  fi
  sleep 0.25
done
if [[ "$ready" != true ]]; then
  sed -n '1,120p' "$core_log" >&2
  sed -n '1,120p' "$relay_log" >&2
  exit 1
fi

curl --silent --fail \
  -D "$temp_root/teach.headers" \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/morning-teaching.json" \
  "$base_url/v1/demonstrations/learn" >"$temp_root/learn.json"
jq -e '.playerModel.revision == 1 and .skill.name == "morning-farm-routine"' "$temp_root/learn.json" >/dev/null

curl --silent --fail "$base_url/v1/echo/memory?saveId=demo-farm" >"$temp_root/memory.json"
jq -e '.modelRevision == 1 and .saveId == "demo-farm"' "$temp_root/memory.json" >/dev/null

request_id="$(awk 'BEGIN{IGNORECASE=1} /^X-EchoFarm-Request-ID:/ {gsub("\r", "", $2); print $2}' "$temp_root/teach.headers")"
[[ "$request_id" =~ ^[0-9a-f]{32}$ ]]
grep -q "request_id=$request_id" "$relay_log"
grep -q "request_id=$request_id" "$core_log"

wrong_token="$(printf 'b%.0s' {1..64})"
ECHOFARM_RELAY_ADDRESS="127.0.0.1:$wrong_relay_port" \
ECHOFARM_RELAY_UPSTREAM_ADDRESS="127.0.0.1:$core_port" \
ECHOFARM_RELAY_TOKEN="$wrong_token" \
ECHOFARM_RELAY_CERT_SHA256="$fingerprint" \
ECHOFARM_RELAY_TIMEOUT_SECONDS=5 \
"$temp_root/echofarm-relay" >"$wrong_relay_log" 2>&1 &
wrong_relay_pid=$!

wrong_status="000"
for _ in {1..40}; do
  wrong_status="$(curl --silent --output /dev/null --write-out '%{http_code}' "http://127.0.0.1:$wrong_relay_port/healthz" || true)"
  [[ "$wrong_status" != "000" ]] && break
  sleep 0.1
done
[[ "$wrong_status" == "401" ]]

if grep -Fq "$token" "$core_log" "$relay_log" || grep -Fq 'demo-farm' "$core_log" "$relay_log"; then
  echo "LAN logs leaked a token or gameplay payload." >&2
  exit 1
fi

jq '{modelRevision, stableTraits: (.stableTraits | length)}' "$temp_root/memory.json"
echo "EchoFarm TLS Kitex/Thrift LAN relay demo passed."
