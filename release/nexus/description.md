# EchoFarm — Teach an AI by Playing

EchoFarm adds a second, translucent version of your farmer to Stardew Valley. You teach it by playing normal farm routines across multiple days. It turns those observations into evidence-backed habits, separates stable preferences from weather-specific behavior, and later works beside you against the farm as it exists now.

It is not a chat window and it is not a coordinate macro. Move a crop, add a new one, make it rain, or empty the watering can: Echo reads the current world state and chooses a new high-level action.

## Current features

- F7 records a normal farm routine and learns task order plus preferred storage.
- F8 summons a translucent player-like Echo without taking control away from you.
- F9 opens Echo Memory to show learned traits, confidence, current player intent, and the latest division-of-work decision.
- F10 pauses Echo and turns your next successful farm action into a high-signal correction.
- Echo walks tile by tile, waters dry crops, refills its own can, harvests mature crops into its own inventory, and deposits them into the chest you taught it to use.
- Rain, changed layouts, empty watering cans, blocked paths, full inventory, and full chests produce explicit replanning or safe-stop behavior.
- Player models and learned skills are isolated by save ID in a local SQLite database.
- Repeated teaching strengthens consistent habits; one unusual day cannot silently overwrite a stable preference.
- While you and Echo work together, a short-lived activity window lets Echo avoid crops and chests you are already handling.
- Learning revisions and runtime decisions are recorded in a structured local ledger for explainability.
- Failed actions are reflected into bounded policy experiences, so a later matching session can avoid the failure before trying the same action.
- Outcomes from actions that actually used an experience feed an auditable local effectiveness score; successful strategies rise, repeated resource contradictions cool them, and transient world changes do not unfairly punish them.
- Each AI proposal carries calibrated confidence and up to two safe alternatives; low-confidence or invalid plans stop instead of touching the world.
- A second deterministic safety gate rejects stale, invented, or invalid model actions before game state changes.

## AI architecture

The local sidecar is written in Go and uses CloudWeGo Eino for trait extraction, player-intent inference, ranked action proposals, and bounded reflection. Deterministic Go code owns evidence and experience merging, durable leased reflection work, idempotent outcome attribution, effective-confidence projection, target claims, and validation. The C#/SMAPI Mod is deliberately thin: it observes the game, captures explicit corrections, validates actions, animates Echo, and mutates the world only on the game thread. No RAG or external knowledge base is used.

The service binds only to localhost. Model credentials stay in the process environment and are not stored in the Mod configuration, logs, archive, or save file.

## Requirements and installation

Install Stardew Valley 1.6 and SMAPI 4.1+, then download the file matching your platform. Extract the single `EchoFarm` folder into `Stardew Valley/Mods` and follow the included README to configure an OpenAI-compatible endpoint.

This is an early technical preview. Use a disposable save until the listed live-game smoke checks have been completed for your platform.

The bundled native sidecar is not currently code-signed. macOS Gatekeeper or Windows security software may require explicit approval; verify the published SHA-256 checksum before allowing it.

## Generative AI disclosure

This project intentionally uses a generative language model at runtime to infer player intent and choose constrained gameplay actions. It contains no AI-generated visual or audio assets. Fixture mode is deterministic and is identified as a non-learning demo mode.
