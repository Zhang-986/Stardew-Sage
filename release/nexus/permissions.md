# Permissions, privacy, and disclosure

## Permissions

- You may use, modify, and redistribute EchoFarm, including in mod collections, under the repository's MIT license and notice requirements.
- Please distinguish unofficial builds from releases published by the project owner, and never bundle API credentials.
- Stardew Valley, SMAPI, model providers, and third-party Go/.NET dependencies retain their own licenses and trademarks.

## Privacy

- The Go service binds to `127.0.0.1` by default.
- Demonstrated gameplay events, player-model records, and skills are stored locally in SQLite.
- When real-model mode is enabled, the current structured world snapshot and demonstration data are sent to the OpenAI-compatible endpoint configured by the player.
- EchoFarm does not intentionally collect analytics, account data, chat messages, or save files.
- API keys are read from `ECHOFARM_MODEL_API_KEY` and are not written to Mod configuration or logs.

## Generative AI

EchoFarm uses generative AI at runtime as a core gameplay mechanic. The model infers intent from demonstrations and selects from a strict action allowlist. No AI-generated images, character art, voice, music, or other media are included in the release.
