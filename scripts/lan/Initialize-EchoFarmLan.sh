#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $0 --host <mac-lan-ip-or-hostname> [--force]" >&2
  exit 2
}

host=""
force=false
while (($#)); do
  case "$1" in
    --host)
      (($# >= 2)) || usage
      host="$2"
      shift 2
      ;;
    --force)
      force=true
      shift
      ;;
    *) usage ;;
  esac
done
[[ -n "$host" ]] || usage
[[ "$host" != *"/"* && "$host" != *" "* ]] || { echo "Invalid LAN host." >&2; exit 2; }

lan_dir="${ECHOFARM_LAN_DIR:-$HOME/Library/Application Support/EchoFarm/LAN}"
certificate="$lan_dir/server.crt"
private_key="$lan_dir/server.key"
token_file="$lan_dir/relay.token"

umask 077
mkdir -p "$lan_dir"
chmod 700 "$lan_dir"
if [[ "$force" != true ]] && [[ -e "$certificate" || -e "$private_key" || -e "$token_file" ]]; then
  echo "EchoFarm LAN identity already exists at $lan_dir; use --force to rotate it." >&2
  exit 1
fi

san="DNS:$host"
if [[ "$host" =~ ^[0-9a-fA-F:.]+$ ]]; then
  san="IP:$host"
fi

openssl req -x509 -newkey rsa:3072 -sha256 -nodes \
  -keyout "$private_key" \
  -out "$certificate" \
  -days 365 \
  -subj "/CN=EchoFarm LAN" \
  -addext "subjectAltName=$san" \
  -addext "keyUsage=digitalSignature,keyEncipherment" \
  -addext "extendedKeyUsage=serverAuth" \
  >/dev/null 2>&1
openssl rand -hex 32 >"$token_file"
chmod 600 "$certificate" "$private_key" "$token_file"

fingerprint="$(openssl x509 -in "$certificate" -noout -fingerprint -sha256 | awk -F= '{print tolower($2)}' | tr -d ':')"
token="$(tr -d '\r\n' <"$token_file")"

echo "EchoFarm LAN identity created outside the repository: $lan_dir"
echo "CERT_SHA256=$fingerprint"
echo "LAN_TOKEN=$token"
echo "Copy the fingerprint and token to the Windows setup, then clear this terminal."
