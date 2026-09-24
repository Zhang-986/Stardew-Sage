# EchoFarm 0.3.0 — Continuum

EchoFarm learns a farm routine from how you play, then creates a translucent AI-controlled Echo that performs the learned goal in the current world instead of replaying yesterday's coordinates.

## Requirements

- Stardew Valley 1.6
- SMAPI 4.1 or newer
- Internet access to an OpenAI-compatible model endpoint, unless using an explicitly selected local endpoint or offline fixture demo

## Install

1. Choose the archive matching your operating system and CPU.
2. Extract it into the Stardew Valley `Mods` directory. The result must be `Mods/EchoFarm/manifest.json`, not an extra nested folder.
3. Set `ECHOFARM_MODEL_API_KEY` in the environment used to launch SMAPI.
4. Launch once, then edit `Mods/EchoFarm/config.json` with `ModelBaseUrl` and `ModelName`.

The bundled local Go service starts automatically. It listens only on `127.0.0.1` and stores learned memory under your OS local application-data directory. API keys are inherited from the launch environment and are never written by the Mod.

On macOS, the Mod restores the executable permission if the unzip tool removed it. If Gatekeeper still quarantines the unsigned preview binary, remove quarantine from this Mod folder only after verifying the downloaded SHA-256 checksum:

```bash
xattr -dr com.apple.quarantine "/path/to/Stardew Valley/Contents/MacOS/Mods/EchoFarm"
```

## Play

- Press F7, perform a normal morning farm routine, then press F7 again to teach Echo.
- Press F8 on a later day to summon the learned Echo.
- Teach the routine on multiple days so Echo can distinguish stable habits from one-off choices.
- Press F9 to inspect Echo's evidence-backed memory, confidence, inferred player intent, and current division of work.
- Echo currently supports walking, watering, refilling its can, harvesting mature crops into its own inventory, and depositing them into the chest you demonstrated.
- While Echo is active, it watches only recent semantic farm actions and avoids crops or chests you are already handling.

Use a disposable save for this preview release. If the model or sidecar is unavailable, Echo stops and Stardew Valley continues normally.

## Offline demo mode

Set `ModelMode` to `fixture` in `config.json` only for a deterministic demo without a model provider. Fixture mode proves the transport and gameplay loop; it does not perform real learning.

Source and issue tracker: https://github.com/Zhang-986/Stardew-Sage
