# EchoFarm Core

EchoFarm Core is the Go + Eino brain for a Stardew Valley AI echo that learns from normal play and later participates in the farm alongside the player.

This module owns behavior interpretation, multi-day evidence merging, player-intent inference, reflective policy learning, ranked next-action selection, decision auditing, and safety validation. It does not control game memory directly. A thin SMAPI bridge translates high-level actions into game-thread operations.

## Run the offline smoke demo

The fixture model is deterministic and only verifies the complete local pipeline. It is not used as a fallback when a real model fails.

```bash
cd ..
./demo/run-core-demo.sh
```

The demo learns from one recorded morning, then shows changed-world decisions:

- rain plus a newly positioned mature crop produces `harvest_target` rather than coordinate replay;
- an empty watering can plus a dry new crop produces `refill_can` before watering;
- an `inventory_full` harvest failure produces `deposit_items` at the learned preferred chest.

The fixture replanner also routes `inventory_full` to the player's demonstrated chest and stops safely on `chest_full`, matching the game bridge's lossless partial-deposit behavior.

## Run the four-day Continuum demo

```bash
cd ..
./demo/run-continuum-demo.sh
```

This scenario teaches two consistent sunny routines and one rainy routine, then starts a co-play day where the player is already watering two crops. It proves that stable sunny habits survive the weather-context change, the player intent is inferred as `watering`, and Echo selects an unclaimed harvest instead of competing for the same targets. The final output is the persisted memory and decision view.

## Run the reflective-policy demo

```bash
cd ..
./demo/run-reflective-demo.sh
```

This scenario restarts the real Go process twice while retaining one SQLite database. It proves that a failed full-inventory harvest creates bounded policy experience, that the next session deposits before harvesting, and that an explicit player chest correction overrides the earlier target on a later decision. It then reports that corrected action as successful and verifies that the applied experience gains one success plus a higher effective confidence. Responses expose calibrated confidence, safe alternatives, and the exact experience ID applied.

Failed-action reflection is backed by a durable SQLite job. Result attachment and enqueue commit together; one lease holder performs inference, transient failures return the job to pending, and a later policy request can recover it after restart. Experience persistence remains idempotent by source ID. The model call itself is at-least-once across a crash boundary and may be recomputed, but it cannot apply the learned experience twice.

Applied-experience feedback is deterministic rather than model-scored. A canonical successful result counts as `succeeded`; known resource contradictions such as `out_of_water`, `inventory_full`, `inventory_empty`, and `chest_full` count as `contradicted` for the relevant action; route churn and unknown failures stay `neutral`. The append-only feedback row commits atomically with the first accepted action result and is projected on reads as success/failure/neutral counts plus `(baseConfidence*4+success)/(4+success+failure)`. An experience is cooled only after at least three contradictions drive that effective confidence below `0.40`.

## Run with a real model

EchoFarm uses Eino's OpenAI-compatible chat-model component. All settings come from environment variables; no credential is stored in the repository.

The configured endpoint must support OpenAI JSON mode. EchoFarm sends `response_format: {"type":"json_object"}` for every structured generation, then applies domain and current-world validation before accepting the response.

```bash
export ECHOFARM_MODEL_MODE=openai
export ECHOFARM_MODEL_BASE_URL=https://your-openai-compatible-endpoint/v1
export ECHOFARM_MODEL_API_KEY=your-key
export ECHOFARM_MODEL_NAME=your-model
go run ./cmd/echofarm
```

Optional settings:

- `ECHOFARM_ADDRESS` defaults to `127.0.0.1:18471` and must remain loopback-only.
- `ECHOFARM_DATABASE_PATH` defaults to `echofarm.db`.
- `ECHOFARM_MAX_MODEL_CALLS_PER_SESSION` defaults to `32` and must be positive.
- `ECHOFARM_MAX_REPORTED_TOKENS_PER_SESSION` defaults to `100000` and must be positive.

Every fixture or real-model invocation reserves one durable SQLite ledger row before execution. The ledger stores a random local request ID, purpose (`learning`, `intent`, `action`, `recovery`, or `reflection`), save/session/day identity, status, latency, bounded error class, and provider-reported token counts when present. It never stores prompts, responses, API keys, or provider error bodies. Once either session budget is reached, no new model call starts. Runtime policy requests persist a `stop_session`; learning and correction requests return HTTP 429 with `model_budget_exhausted`.

## API

- `GET /healthz`
- `POST /v1/demonstrations/learn`
- `POST /v1/echo/next-action`
- `POST /v1/echo/action-result`
- `POST /v1/echo/corrections`
- `GET /v1/player-model?saveId=...`
- `GET /v1/skills/morning-farm-routine?saveId=...`
- `GET /v1/echo/memory?saveId=...`
- `GET /v1/model-usage?saveId=...&sessionId=...`

All model outputs are validated against the action catalog and current snapshot before they can cross into the game bridge. If the model is unavailable, Echo stops and returns HTTP 503; it does not silently switch to scripted behavior.
