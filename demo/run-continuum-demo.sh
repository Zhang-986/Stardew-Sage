#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
core_dir="$repo_root/echofarm-core"
port="${ECHOFARM_CONTINUUM_PORT:-18472}"
base_url="http://127.0.0.1:$port"
database="${TMPDIR:-/tmp}/echofarm-continuum-$$.db"
server_log="${TMPDIR:-/tmp}/echofarm-continuum-$$.log"
result_dir="${TMPDIR:-/tmp}/echofarm-continuum-results-$$"
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
  local fixture="$1"
  local output="$2"
  curl --silent --fail \
    -H 'Content-Type: application/json' \
    --data-binary "@$repo_root/demo/fixtures/$fixture" \
    "$base_url/v1/demonstrations/learn" >"$output"
}

echo "Day 1: Echo forms candidate habits from the first sunny routine"
teach morning-teaching.json "$result_dir/day1.json"
jq -e '.playerModel.revision == 1 and .playerModel.learnedThroughDay == 1' "$result_dir/day1.json" >/dev/null

echo "Day 2: repeated sunny behavior strengthens the model"
teach day-2-sunny-teaching.json "$result_dir/day2.json"
jq -e '.playerModel.revision == 2 and ([.playerModel.traits[] | select(.key == "task_order" and .context == "sunny" and .observationCount == 2)] | length == 1)' "$result_dir/day2.json" >/dev/null

echo "Day 3: rainy behavior is learned as context, not a contradiction"
teach day-3-rainy-teaching.json "$result_dir/day3.json"
jq -e '.playerModel.revision == 3 and .playerModel.learnedThroughDay == 3' "$result_dir/day3.json" >/dev/null
jq -e '[.playerModel.traits[] | select(.key == "task_order" and .context == "sunny" and .observationCount == 2 and .contradictionCount == 0)] | length == 1' "$result_dir/day3.json" >/dev/null

echo "Day 4: player waters the north plot while Echo takes the unclaimed harvest"
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/day-4-coplay-farm.json" \
  "$base_url/v1/echo/next-action" >"$result_dir/day4.json"
jq -e '.action.kind == "harvest_target" and .action.targetId == "crop-south-ripe"' "$result_dir/day4.json" >/dev/null

curl --silent --fail "$base_url/v1/echo/memory?saveId=demo-farm" >"$result_dir/memory.json"
jq -e '.modelRevision == 3 and .learnedThroughDay == 3' "$result_dir/memory.json" >/dev/null
jq -e '.lastDecision.inferredIntent == "watering" and (.lastDecision.playerClaimedTargets | length) == 2 and .lastDecision.finalAction.targetId == "crop-south-ripe"' "$result_dir/memory.json" >/dev/null

jq '{modelRevision, learnedThroughDay, stableTraits, collaboration: .lastDecision}' "$result_dir/memory.json"
echo "EchoFarm Continuum four-day demo passed."
