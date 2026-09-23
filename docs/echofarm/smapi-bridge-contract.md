# EchoFarm SMAPI Bridge Contract

This document freezes the boundary between the Go/Eino core and the future C#/SMAPI mod. The bridge is a sensor and actuator; it must not infer player intent or choose goals.

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

Success returns the new `playerModel` and `skill`. The bridge can render these as a short “Echo memory” page but must not edit them.

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
  "obstacles": []
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

## Read learned memory

```http
GET /v1/player-model?saveId=farm-123
GET /v1/skills/morning-farm-routine?saveId=farm-123
```

These endpoints support the in-game Echo memory panel and diagnostics. A missing profile or skill returns 404.

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
