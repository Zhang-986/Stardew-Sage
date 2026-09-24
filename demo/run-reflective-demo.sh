#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
core_dir="$repo_root/echofarm-core"
port="${ECHOFARM_REFLECTIVE_PORT:-18473}"
base_url="http://127.0.0.1:$port"
temp_root="$(mktemp -d "${TMPDIR:-/tmp}/echofarm-reflective.XXXXXX")"
database="$temp_root/echo.db"
core_binary="$temp_root/echofarm-core"
server_log="$temp_root/server.log"

cleanup() {
  stop_server
  rm -rf "$temp_root"
}

stop_server() {
  if [[ -n "${server_pid:-}" ]]; then
    kill "$server_pid" 2>/dev/null || true
    wait "$server_pid" 2>/dev/null || true
    server_pid=""
  fi
}

start_server() {
  : >"$server_log"
  ECHOFARM_MODEL_MODE=fixture \
  ECHOFARM_ADDRESS="127.0.0.1:$port" \
  ECHOFARM_DATABASE_PATH="$database" \
    "$core_binary" >"$server_log" 2>&1 &
  server_pid=$!
  for _ in {1..120}; do
    if curl --silent --fail "$base_url/healthz" >/dev/null; then
      return
    fi
    if ! kill -0 "$server_pid" 2>/dev/null; then
      sed -n '1,160p' "$server_log" >&2
      exit 1
    fi
    sleep 0.25
  done
  sed -n '1,160p' "$server_log" >&2
  exit 1
}

trap cleanup EXIT

(cd "$core_dir" && go build -o "$core_binary" ./cmd/echofarm)
start_server

echo "1) Teach the baseline player routine"
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/morning-teaching.json" \
  "$base_url/v1/demonstrations/learn" >"$temp_root/learn.json"
jq -e '.playerModel.revision == 1' "$temp_root/learn.json" >/dev/null

echo "2) Let the first full-inventory harvest fail and reflect once"
jq '.snapshot | .snapshotVersion = 1' "$repo_root/demo/fixtures/full-inventory-result.json" |
  curl --silent --fail \
    -H 'Content-Type: application/json' \
    --data-binary @- \
    "$base_url/v1/echo/next-action" >"$temp_root/attempt.json"
jq -e '.action.kind == "harvest_target"' "$temp_root/attempt.json" >/dev/null
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/full-inventory-result.json" \
  "$base_url/v1/echo/action-result" >"$temp_root/recovery.json"
jq -e '.action.kind == "deposit_items" and .action.targetId == "shipping-chest" and (.appliedExperiences | length) == 1' "$temp_root/recovery.json" >/dev/null

echo "3) Restart the process and prove proactive cross-session learning"
stop_server
start_server
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/day-5-reflective-farm.json" \
  "$base_url/v1/echo/next-action" >"$temp_root/day5.json"
jq -e '
  .action.kind == "deposit_items" and
  .action.targetId == "shipping-chest" and
  .confidence >= 0.75 and
  ([.alternatives[].kind] | index("harvest_target")) != null and
  ([.alternatives[].kind] | index("stop_session")) != null and
  (.appliedExperiences | length) == 1
' "$temp_root/day5.json" >/dev/null

echo "4) Submit the player's explicit chest correction"
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$repo_root/demo/fixtures/player-chest-correction.json" \
  "$base_url/v1/echo/corrections" >"$temp_root/correction.json"
jq -e '.experience.source == "correction" and .experience.preferredTargetId == "artisan-chest" and .experience.confidence >= 0.8' "$temp_root/correction.json" >/dev/null

echo "5) Restart again and prove the corrected target wins"
stop_server
start_server
jq '.sessionId = "echo-day-6" | .snapshotVersion = 1 | .tick = 6000 | .day = 6' \
  "$repo_root/demo/fixtures/day-5-reflective-farm.json" |
  curl --silent --fail \
    -H 'Content-Type: application/json' \
    --data-binary @- \
    "$base_url/v1/echo/next-action" >"$temp_root/day6.json"
jq -e '.action.kind == "deposit_items" and .action.targetId == "artisan-chest" and .confidence >= 0.9 and (.appliedExperiences | length) == 1' "$temp_root/day6.json" >/dev/null

curl --silent --fail "$base_url/v1/echo/memory?saveId=demo-farm" >"$temp_root/memory.json"
jq -e '
  ([.experiences[].evidenceRefs[]] | index("decision:echo-day-4:1")) != null and
  ([.experiences[].evidenceRefs[]] | index("correction-chest-west-1")) != null and
  .lastDecision.finalAction.targetId == "artisan-chest"
' "$temp_root/memory.json" >/dev/null

echo "6) Report the corrected action succeeding and verify effectiveness feedback"
applied_experience_id="$(jq -r '.appliedExperiences[0]' "$temp_root/day6.json")"
jq -n \
  --slurpfile world "$repo_root/demo/fixtures/day-5-reflective-farm.json" \
  --slurpfile decision "$temp_root/day6.json" '
  {
    saveId: "demo-farm",
    snapshot: ($world[0] |
      .sessionId = "echo-day-6" |
      .snapshotVersion = 2 |
      .tick = 6100 |
      .day = 6 |
      .inventory = {freeSlots: 12, items: []}),
    result: {
      saveId: "demo-farm",
      sessionId: "echo-day-6",
      snapshotVersion: 1,
      action: $decision[0].action,
      status: "succeeded"
    }
  }' >"$temp_root/day6-success.json"
curl --silent --fail \
  -H 'Content-Type: application/json' \
  --data-binary "@$temp_root/day6-success.json" \
  "$base_url/v1/echo/action-result" >"$temp_root/day6-next.json"
curl --silent --fail "$base_url/v1/echo/memory?saveId=demo-farm" >"$temp_root/feedback-memory.json"
jq -e --arg id "$applied_experience_id" '
  any(.experiences[];
    .id == $id and
    .successCount == 1 and
    .effectiveConfidence > .confidence)
' "$temp_root/feedback-memory.json" >/dev/null

jq --arg id "$applied_experience_id" '{decision: .lastDecision, feedback: (.experiences[] | select(.id == $id))}' "$temp_root/feedback-memory.json"
echo "EchoFarm reflective cross-process demo passed."
