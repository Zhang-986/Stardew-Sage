# EchoFarm 0.4.0 — Reflective Policy

EchoFarm learns a farm routine from how you play, then creates a translucent AI-controlled Echo that performs the learned goal in the current world instead of replaying yesterday's coordinates.

## Requirements

- Stardew Valley 1.6
- SMAPI 4.1 or newer
- Internet access to an OpenAI-compatible model endpoint, unless using an explicitly selected local endpoint or offline fixture demo

## Install

1. Choose the archive matching your operating system and CPU.
2. Extract it into the Stardew Valley `Mods` directory. The result must be `Mods/EchoFarm/manifest.json`, not an extra nested folder.
3. Launch the game through SMAPI. The first run defaults to deterministic fixture mode and requires no API key.
4. Load a disposable save and press F9. Confirm the panel shows `DEMO`, a healthy Core, the database location, and `Harvest: disabled`.

The bundled local Go service starts automatically. It listens only on `127.0.0.1` and stores learned memory under your OS local application-data directory. API keys are inherited from the launch environment and are never written by the Mod.

On Windows, repository owners can use the checked workflow instead of copying files manually:

```powershell
$game = "C:\Program Files (x86)\Steam\steamapps\common\Stardew Valley"
.\scripts\windows\Install-EchoFarm.ps1 -Doctor -GamePath $game
.\scripts\windows\Install-EchoFarm.ps1 -Build -GamePath $game
.\scripts\windows\Install-EchoFarm.ps1 -Install -GamePath $game -PackagePath .\dist\windows\EchoFarm
```

On macOS, the Mod restores the executable permission if the unzip tool removed it. If Gatekeeper still quarantines the unsigned preview binary, remove quarantine from this Mod folder only after verifying the downloaded SHA-256 checksum:

```bash
xattr -dr com.apple.quarantine "/path/to/Stardew Valley/Contents/MacOS/Mods/EchoFarm"
```

## Play

- Press F7, perform a normal morning farm routine, then press F7 again to teach Echo.
- Press F8 on a later day to summon the learned Echo.
- Teach the routine on multiple days so Echo can distinguish stable habits from one-off choices.
- During teaching, Echo can classify tree chopping, rock breaking, mine-floor changes, and fishing outcomes into lifestyle memory. These categories are visibly `learn-only` and do not grant autonomous game mutations.
- Press F9 to inspect Echo's evidence-backed memory, confidence, inferred player intent, and current division of work.
- F9 also shows session call/token budgets, provider-reported usage, model failures, and recent latency. Missing provider token metadata is shown as `unknown`.
- Press F10 while Echo has a pending decision, then perform one successful farm action within 20 seconds to teach a better choice. Press F10 again to cancel, or retry after a temporary model outage.
- Echo currently supports walking, watering, refilling its can, and depositing items into the chest you demonstrated. Experimental harvesting remains disabled by default until it passes native-game certification.
- While Echo is active, it watches only recent semantic farm actions and avoids crops or chests you are already handling.
- Failed actions and explicit corrections become bounded, auditable policy experiences; later matching situations can change the first action before another failure occurs.

Use a disposable save for this preview release. If the model or sidecar is unavailable, Echo stops and Stardew Valley continues normally.

## Offline demo mode

Fixture mode is the first-run default and is visibly labeled `DEMO`; it proves the transport and gameplay loop without making paid model calls. To enable real AI, set `ModelMode` to `openai`, set `ModelBaseUrl` and `ModelName` in `config.json`, and provide `ECHOFARM_MODEL_API_KEY` only in the environment used to launch SMAPI. If any value is missing, Echo fails closed and F9 shows the exact corrective action.

The default per-session safeguards are `MaxModelCallsPerSession: 32` and `MaxReportedTokensPerSession: 100000`. Crossing either limit prevents another model request and stops Echo safely. The local ledger never stores prompts, responses, keys, or provider error bodies, and it does not guess currency cost.

Source and issue tracker: https://github.com/Zhang-986/Stardew-Sage
