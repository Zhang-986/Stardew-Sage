#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 --address <mac-lan-ip:port>" >&2
  echo "Model URL, name, and API key are read interactively or from the current environment." >&2
  exit 2
}

address=""
while (($#)); do
  case "$1" in
    --address)
      (($# >= 2)) || usage
      address="$2"
      shift 2
      ;;
    *) usage ;;
  esac
done
[[ -n "$address" ]] || usage

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
core_dir="$repo_root/echofarm-core"
data_root="${ECHOFARM_DATA_DIR:-$HOME/Library/Application Support/EchoFarm}"
lan_dir="${ECHOFARM_LAN_DIR:-$data_root/LAN}"
certificate="$lan_dir/server.crt"
private_key="$lan_dir/server.key"
token_file="$lan_dir/relay.token"

for required in "$certificate" "$private_key" "$token_file"; do
  [[ -s "$required" ]] || { echo "Missing LAN identity file: $required" >&2; exit 1; }
done

model_mode="${ECHOFARM_MODEL_MODE:-openai}"
if [[ "$model_mode" == "openai" ]]; then
  if [[ -z "${ECHOFARM_MODEL_BASE_URL:-}" ]]; then
    read -r -p "Model base URL: " ECHOFARM_MODEL_BASE_URL </dev/tty
  fi
  if [[ -z "${ECHOFARM_MODEL_NAME:-}" ]]; then
    read -r -p "Model name/deployment: " ECHOFARM_MODEL_NAME </dev/tty
  fi
  if [[ -z "${ECHOFARM_MODEL_API_KEY:-}" ]]; then
    read -r -s -p "Model API key (not stored): " ECHOFARM_MODEL_API_KEY </dev/tty
    echo >/dev/tty
  fi
elif [[ "$model_mode" != "fixture" ]]; then
  echo "ECHOFARM_MODEL_MODE must be openai or fixture." >&2
  exit 2
fi

mkdir -p "$data_root/bin"
chmod 700 "$data_root" "$data_root/bin"
binary="$data_root/bin/echofarm-core"
(
  cd "$core_dir"
  go build -trimpath -o "$binary" ./cmd/echofarm
)
chmod 700 "$binary"

export ECHOFARM_ADDRESS="$address"
export ECHOFARM_ALLOW_LAN=true
export ECHOFARM_LAN_TOKEN="$(tr -d '\r\n' <"$token_file")"
export ECHOFARM_TLS_CERT_FILE="$certificate"
export ECHOFARM_TLS_KEY_FILE="$private_key"
export ECHOFARM_DATABASE_PATH="${ECHOFARM_DATABASE_PATH:-$data_root/echofarm.db}"
export ECHOFARM_MODEL_MODE="$model_mode"
export ECHOFARM_MODEL_BASE_URL="${ECHOFARM_MODEL_BASE_URL:-}"
export ECHOFARM_MODEL_NAME="${ECHOFARM_MODEL_NAME:-}"
export ECHOFARM_MODEL_API_KEY="${ECHOFARM_MODEL_API_KEY:-}"
export ECHOFARM_MODEL_TIMEOUT_SECONDS="${ECHOFARM_MODEL_TIMEOUT_SECONDS:-90}"

echo "Starting private EchoFarm Kitex/Thrift server at $address"
echo "Database and model credentials remain under the Mac user account."
exec "$binary"
