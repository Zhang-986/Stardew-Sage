#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
core_dir="$repo_root/echofarm-core"
port="${ECHOFARM_ACTIVITY_PORT:-18475}"
base_url="http://127.0.0.1:$port"
database="${TMPDIR:-/tmp}/echofarm-activity-$$.db"
server_log="${TMPDIR:-/tmp}/echofarm-activity-$$.log"
result_dir="${TMPDIR:-/tmp}/echofarm-activity-results-$$"
mkdir -p "$result_dir"

cleanup() {
  if [[ -n "${server_pid:-}" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
  fi
  rm -f "$database" "$server_log"
  rm -rf "$result_dir"
}
trap cleanup EXIT

(
  cd "$core_dir"
  ECHOFARM_MODEL_MODE=fixture \
  ECHOFARM_ADDRESS="127.0.0.1:$port" \
  ECHOFARM_DATABASE_PATH="$database" \
  go run ./cmd/echofarm >"$server_log" 2>&1
) &
server_pid=$!

for _ in {1..120}; do
  if curl --silent --fail "$base_url/healthz" >/dev/null; then
    break
  fi
  if ! kill -0 "$server_pid" 2>/dev/null; then
    sed -n '1,160p' "$server_log" >&2
    exit 1
  fi
  sleep 0.25
done
curl --silent --fail "$base_url/healthz" >/dev/null

teach() {
  curl --silent --fail \
    -H 'Content-Type: application/json' \
    --data-binary "@$repo_root/demo/fixtures/$1" \
    "$base_url/v1/demonstrations/learn" >"$2"
}

echo "Day 1: classify woodcutting, mining, mine traversal, and fishing"
teach activity-day-1.json "$result_dir/day1.json"
jq -e '.playerModel.revision == 1 and ([.playerModel.traits[] | select(.key == "activity_order")] | length == 1)' "$result_dir/day1.json" >/dev/null

echo "Day 2: repeated evidence becomes stable player memory"
teach activity-day-2.json "$result_dir/day2.json"
jq -e '.playerModel.revision == 2 and ([.playerModel.traits[] | select(.key == "activity_order" and .observationCount == 2)] | length == 1)' "$result_dir/day2.json" >/dev/null
jq -e '([.playerModel.traits[] | select(.key == "resource_priority" and .value == "Wood" and .observationCount == 2)] | length == 1)' "$result_dir/day2.json" >/dev/null

curl --silent --fail "$base_url/v1/echo/memory?saveId=activity-farm" >"$result_dir/memory.json"
jq -e '.modelRevision == 2 and ([.stableTraits[] | select(.key == "activity_order")] | length == 1)' "$result_dir/memory.json" >/dev/null

curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/activity-runtime.json" \
  "$base_url/v1/echo/next-action" >"$result_dir/action.json"
jq -e '.action.kind | IN("move_to", "equip_tool", "water_target", "refill_can", "harvest_target", "deposit_items", "stop_session")' "$result_dir/action.json" >/dev/null
jq -e '.action.kind | IN("chop_tree", "break_rock", "enter_mine_floor", "fish_caught", "fish_escaped") | not' "$result_dir/action.json" >/dev/null

jq '{revision: .modelRevision, stableTraits}' "$result_dir/memory.json"
echo "EchoFarm semantic activity learning demo passed."
