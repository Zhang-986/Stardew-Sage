#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
core_dir="$repo_root/echofarm-core"
database="${TMPDIR:-/tmp}/echofarm-demo-$$.db"
server_log="${TMPDIR:-/tmp}/echofarm-demo-$$.log"

cleanup() {
  if [[ -n "${server_pid:-}" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -f "$database" "$server_log"
}
trap cleanup EXIT

(
  cd "$core_dir"
  ECHOFARM_MODEL_MODE=fixture \
  ECHOFARM_DATABASE_PATH="$database" \
  go run ./cmd/echofarm >"$server_log" 2>&1
) &
server_pid=$!

for _ in {1..120}; do
  if curl --silent --fail http://127.0.0.1:18471/healthz >/dev/null; then
    break
  fi
  if ! kill -0 "$server_pid" 2>/dev/null; then
    sed -n '1,160p' "$server_log" >&2
    exit 1
  fi
  sleep 0.25
done

if ! curl --silent --fail http://127.0.0.1:18471/healthz >/dev/null; then
  sed -n '1,160p' "$server_log" >&2
  exit 1
fi

echo "1) Teach Echo from a normal morning routine"
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/morning-teaching.json" \
  http://127.0.0.1:18471/v1/demonstrations/learn | jq '{playerModel, skill: {name: .skill.name, goal: .skill.goal, preferredOrder: .skill.preferredOrder}}'

echo "2) Change the layout and make it rain: Echo skips watering and harvests a new crop"
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/changed-rainy-farm.json" \
  http://127.0.0.1:18471/v1/echo/next-action | jq .

echo "3) Empty the can on a sunny day: Echo replans to refill first"
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/empty-can-farm.json" \
  http://127.0.0.1:18471/v1/echo/next-action | jq .

echo "4) Fill Echo's inventory: it replans the failed harvest into a deposit at the learned chest"
jq '.snapshot | .snapshotVersion = 1' "$repo_root/demo/fixtures/full-inventory-result.json" | \
  curl --silent --fail \
    -H 'Content-Type: application/json' \
    --data-binary @- \
    http://127.0.0.1:18471/v1/echo/next-action >/dev/null
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/full-inventory-result.json" \
  http://127.0.0.1:18471/v1/echo/action-result | jq .
