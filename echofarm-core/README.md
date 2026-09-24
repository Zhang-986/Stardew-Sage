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

This scenario restarts the real Go process twice while retaining one SQLite database. It proves that a failed full-inventory harvest creates bounded policy experience, that the next session deposits before harvesting, and that an explicit player chest correction overrides the earlier target on a later decision. Responses expose calibrated confidence, safe alternatives, and the exact experience ID applied.

## Run with a real model

EchoFarm uses Eino's OpenAI-compatible chat-model component. All settings come from environment variables; no credential is stored in the repository.

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

## API

- `GET /healthz`
- `POST /v1/demonstrations/learn`
- `POST /v1/echo/next-action`
- `POST /v1/echo/action-result`
- `POST /v1/echo/corrections`
- `GET /v1/player-model?saveId=...`
- `GET /v1/skills/morning-farm-routine?saveId=...`
- `GET /v1/echo/memory?saveId=...`

All model outputs are validated against the action catalog and current snapshot before they can cross into the game bridge. If the model is unavailable, Echo stops and returns HTTP 503; it does not silently switch to scripted behavior.
