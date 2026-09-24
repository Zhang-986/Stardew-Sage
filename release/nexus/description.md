# EchoFarm — Teach an AI by Playing

EchoFarm adds a second, translucent version of your farmer to Stardew Valley. You teach it by playing a normal farm routine once. It observes the actions and their outcomes, builds a structured player model, and later performs the same goal against the farm as it exists now.

It is not a chat window and it is not a coordinate macro. Move a crop, add a new one, make it rain, or empty the watering can: Echo reads the current world state and chooses a new high-level action.

## Current features

- F7 records a normal farm routine and learns task order plus preferred storage.
- F8 summons a translucent player-like Echo without taking control away from you.
- Echo walks tile by tile, waters dry crops, refills its own can, harvests mature crops into its own inventory, and deposits them into the chest you taught it to use.
- Rain, changed layouts, empty watering cans, blocked paths, full inventory, and full chests produce explicit replanning or safe-stop behavior.
- Player models and learned skills are isolated by save ID in a local SQLite database.
- A second deterministic safety gate rejects stale, invented, or invalid model actions before game state changes.

## AI architecture

The local sidecar is written in Go and uses CloudWeGo Eino for structured learning and action graphs. The C#/SMAPI Mod is deliberately thin: it observes the game, validates actions, animates Echo, and mutates the world only on the game thread. No RAG or external knowledge base is used.

The service binds only to localhost. Model credentials stay in the process environment and are not stored in the Mod configuration, logs, archive, or save file.

## Requirements and installation

Install Stardew Valley 1.6 and SMAPI 4.1+, then download the file matching your platform. Extract the single `EchoFarm` folder into `Stardew Valley/Mods` and follow the included README to configure an OpenAI-compatible endpoint.

This is an early technical preview. Use a disposable save until the listed live-game smoke checks have been completed for your platform.

The bundled native sidecar is not currently code-signed. macOS Gatekeeper or Windows security software may require explicit approval; verify the published SHA-256 checksum before allowing it.

## Generative AI disclosure

This project intentionally uses a generative language model at runtime to infer player intent and choose constrained gameplay actions. It contains no AI-generated visual or audio assets. Fixture mode is deterministic and is identified as a non-learning demo mode.
