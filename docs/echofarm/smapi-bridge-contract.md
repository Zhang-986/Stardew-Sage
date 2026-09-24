# EchoFarm SMAPI Bridge Contract

This document freezes the boundary between the Go/Eino core and the C#/SMAPI mod. The bridge is a sensor and actuator; it captures recent semantic activity, while the Go/Eino core infers player intent and chooses goals.

## Connection and lifecycle

- Base URL: `http://127.0.0.1:18471` by default.
- Connect and read timeout from the bridge: 35 seconds for AI endpoints, 2 seconds for `/healthz`.
- Start a new `sessionId` for each teaching recording and each summoned Echo work session.
- Increment `snapshotVersion` whenever relevant game state changes.
- Stop Echo after a timeout, HTTP 503, malformed response, or failed local validation. Never block game save, sleep, location change, or exit.
- Never retry an action request automatically unless the same idempotency tuple (`saveId`, `sessionId`, `snapshotVersion`) is still current. Learning requests may be retried with the same demonstration ID.

## Correlation rules

Every runtime response must echo the exact `saveId`, `sessionId`, and `snapshotVersion` from the submitted snapshot. The bridge rejects stale responses even when their action is otherwise valid.

Stable target IDs are derived from save ID, location, entity kind, and the entity's current game identifier. Tile coordinates are attributes, not identity. Moving or adding a crop therefore changes the current snapshot without turning a learned skill into a coordinate macro.

## Health

```http
GET /healthz
```

```json
{"status":"ok"}
```

The bridge may call this before enabling the “Record Echo” and “Summon Echo” controls.

## Submit a teaching demonstration

```http
POST /v1/demonstrations/learn
Content-Type: application/json
```

The body follows `contracts/demonstration.schema.json`. Movement events should already be coalesced by the bridge; semantic action events must include action outcome and state delta.

```json
{
  "id": "demo-day-1",
  "saveId": "farm-123",
  "sessionId": "teaching-1",
  "day": 1,
  "weather": "sunny",
  "startedAt": 100,
  "endedAt": 400,
  "events": [
    {
      "id": "event-water-1",
      "kind": "water_target",
      "tick": 200,
      "position": {"x": 10, "y": 8},
      "targetId": "crop-abc",
      "tool": "Watering Can",
      "delta": {"energyDelta": -2, "waterDelta": -1, "inventoryDelta": 0},
      "success": true
    }
  ]
}
```

Success returns the new `playerModel`, `skill`, and `learningChange`. Repeating the same demonstration ID is idempotent and returns its stored outcome without increasing confidence again.

## Ask for the next action

```http
POST /v1/echo/next-action
Content-Type: application/json
```

The body follows `contracts/world-snapshot.schema.json` and contains only decision-relevant state.

```json
{
  "saveId": "farm-123",
  "sessionId": "echo-day-2",
  "snapshotVersion": 17,
  "tick": 3600,
  "day": 2,
  "timeOfDay": 620,
  "weather": "sunny",
  "location": "Farm",
  "playerPosition": {"x": 4, "y": 8},
  "energy": 250,
  "maxEnergy": 270,
  "inventory": {"freeSlots": 10, "items": []},
  "wateringCan": {"name": "Watering Can", "level": 1, "water": 20, "capacity": 40},
  "crops": [],
  "waterSources": [],
  "chests": [],
  "obstacles": [],
  "capabilities": {"harvest": false},
  "recentPlayerActions": [
    {"kind": "water_target", "targetId": "crop-north-1", "tick": 3500, "success": true}
  ]
}
```

```json
{
  "action": {
    "saveId": "farm-123",
    "sessionId": "echo-day-2",
    "snapshotVersion": 17,
    "kind": "water_target",
    "targetId": "crop-new-7",
    "reason": "water a currently dry crop"
  }
}
```

Before execution, the bridge must rebuild or inspect the newest game state and recheck:

- correlation fields still match;
- action kind is in the shared allowlist;
- target exists and is interactable;
- rain/tool/energy preconditions still permit the action;
- all game mutations will occur on the SMAPI game thread.
- the target is not claimed by the player's recent activity window.
- the action capability is explicitly enabled; legacy snapshots are accepted only for compatibility tests, while the Mod always sends capabilities.

## Read model usage

```http
GET /v1/model-usage?saveId=farm-123&sessionId=echo-day-2
```

The response contains current-session and same-day call counts, success/failure counts, provider-reported token totals, known/unknown token state, recent latency, configured limits, and whether the session is exhausted. It never contains prompts, model responses, API keys, provider bodies, or a derived currency cost. The same bounded summary is embedded in `/v1/echo/memory` for F9.

## Report an action result

```http
POST /v1/echo/action-result
Content-Type: application/json
```

```json
{
  "saveId": "farm-123",
  "snapshot": {
    "saveId": "farm-123",
    "sessionId": "echo-day-2",
    "snapshotVersion": 18,
    "day": 2,
    "timeOfDay": 630,
    "weather": "sunny",
    "location": "Farm",
    "playerPosition": {"x": 4, "y": 8},
    "energy": 248,
    "maxEnergy": 270,
    "inventory": {"freeSlots": 10, "items": []},
    "wateringCan": {"name": "Watering Can", "level": 1, "water": 19, "capacity": 40}
  },
  "result": {
    "saveId": "farm-123",
    "sessionId": "echo-day-2",
    "snapshotVersion": 17,
    "action": {
      "saveId": "farm-123",
      "sessionId": "echo-day-2",
      "snapshotVersion": 17,
      "kind": "water_target",
      "targetId": "crop-new-7",
      "reason": "water a currently dry crop"
    },
    "status": "failed",
    "errorCode": "path_blocked"
  }
}
```

A failed result invokes the Eino replanning graph with the failure plus latest snapshot. A successful result asks for the next action against the latest snapshot. Both return the same `ActionResponse` shape used by `/next-action`.

When the recorded primary proposal actually applied one or more persisted experiences, accepting its first canonical result also appends deterministic effectiveness feedback for those experience IDs. Success is `succeeded`; action/resource contradictions (`water_target/out_of_water`, `harvest_target/inventory_full`, and `deposit_items/chest_full|inventory_empty`) are `contradicted`; route churn, changed targets, unsupported crops, and unknown failures are `neutral`. Alternatives, safety-stop substitutions, and results that do not exactly match the recorded final action receive no attribution.

Game-side failure codes used by the vertical slice are:

- `target_changed`: the crop, chest, or water source no longer matches the snapshot;
- `path_blocked`: no bounded passable route reaches an adjacent interaction tile;
- `out_of_water`: Echo must replan through a water source;
- `inventory_full`: Echo must deposit before harvesting another distinct stack;
- `inventory_empty`: a deposit action is no longer necessary;
- `chest_full`: the chest accepted only part or none of the Echo inventory;
- `unsupported_crop`: the crop has no safe harvest item mapping.

Harvested items live in Echo's own bounded inventory. A chest transfer removes only the quantity accepted by `Chest.addItem`; a failed or partial deposit never silently deletes the remainder and never routes it through the player's backpack.

## Read learned memory

```http
GET /v1/player-model?saveId=farm-123
GET /v1/skills/morning-farm-routine?saveId=farm-123
GET /v1/echo/memory?saveId=farm-123
```

The memory endpoint projects stable multi-day traits, the latest learning change, active Echo session, and last decision record for the F9 panel. Policy experiences may additionally expose `effectiveConfidence`, `successCount`, `failureCount`, and `neutralCount`; zero-valued projection fields may be omitted for backward compatibility. A missing profile or skill returns 404.

## Persistence and idempotency

- A learning revision, derived player model, skill, and source demonstration commit in one SQLite transaction.
- Reusing a demonstration ID returns the original outcome without another model call.
- Runtime decisions use `(saveId, sessionId, snapshotVersion)` as their idempotency key.
- An action result is accepted only when its full action payload matches the recorded final action for that key.
- Result attachment is an atomic first-write-wins operation. A canonical failed result and its pending reflection job commit together; identical retries cannot enqueue another job, while a later result with different content is rejected.
- Feedback for materially applied experience IDs commits in that same transaction and is keyed by the canonical decision plus experience ID. Identical retries cannot double-count an outcome, and a missing referenced experience rolls the whole result transaction back.
- Base semantic confidence remains part of the learned experience. Memory reads project the append-only feedback ledger into effective confidence using `(confidence * 4 + successes) / (4 + successes + contradictions)`; neutral outcomes remain auditable but do not change the score. Three or more contradictions cool an experience only when the projected score falls below `0.40`.
- Reflection jobs use expiring SQLite leases. A model failure releases the job for a later policy request, and process restart preserves pending work; the eventual experience revision remains idempotent by source ID.
- Model candidates rejected by player target claims are retained in the ledger beside the deterministic `stop_session` decision for diagnosis.

## Error contract

```json
{
  "code": "model_unavailable",
  "message": "Echo is unavailable; the game may continue normally"
}
```

Relevant statuses:

- `400`: malformed JSON, unknown fields, missing IDs, or correlation mismatch;
- `404`: this save has no learned Echo memory;
- `422`: a valid JSON request or model action failed domain/safety validation;
- `503`: model connection, timeout, or provider failure.

Responses never include API keys, provider response bodies, stack traces, or local filesystem paths.

## Bridge execution state machine

```text
Idle -> Recording -> Learning -> Ready -> Acting -> AwaitingResult
          |            |                    ^            |
          |            +-- failure -> Idle  |            |
          +-- cancel -> Idle                +-- replan --+

Acting/AwaitingResult -- save, sleep, warp, timeout, 503 --> Idle
```

The bridge owns animation and path progress. The Go core receives one result per high-level action, not per frame.
