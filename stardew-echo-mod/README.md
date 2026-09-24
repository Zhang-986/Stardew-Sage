# EchoFarm .NET / SMAPI bridge

This directory contains the game-side half of EchoFarm.

## What is runnable here

`EchoFarm.Bridge` targets `net6.0` and has no dependency on Stardew Valley. It provides:

- C# DTOs matching the Go JSON contracts;
- strict JSON parsing and enum handling;
- teaching-session recording with duplicate movement suppression;
- the localhost Go-core client with timeout and stale-response rejection;
- a second action safety gate;
- the Echo session state machine and single-action-in-flight guarantee.

Run the tests from the repository root:

```bash
./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln
```

The integration suite starts the Go service in fixture mode and proves that C# can teach a routine and receive `harvest_target` and `refill_can` decisions for changed farm states.

## SMAPI adapter status

`src/EchoFarm.Mod` contains the real SMAPI-facing adapter:

- F7 starts/stops teaching;
- F8 summons the learned Echo;
- F9 opens the in-game Echo memory panel;
- F10 pauses Echo and treats the next successful farm action as an explicit correction;
- player movement and semantic tool/action events are recorded;
- farm crops, chests, water sources, weather, time, and Echo resources are mapped into `WorldSnapshot`;
- Go actions are queued back onto the game update thread and use bounded grid pathfinding;
- the Echo is rendered as a translucent copy of the player's appearance;
- movement is animated tile by tile; watering and watering-can refill are wired;
- mature crops are collected into a bounded, quality-aware Echo inventory;
- Echo inventory is deposited into the learned target chest without touching the player's backpack;
- full inventories and full chests produce explicit replanning failures instead of losing items.
- recent player actions are sent as a short-lived intent window, so Echo avoids targets the player is already handling;
- the memory panel shows stable cross-day traits, confidence, current inferred intent, and the latest division-of-work decision.
- reflective decisions retain calibrated confidence, safe alternatives, and the policy experience that influenced them.

The adapter is intentionally outside `EchoFarm.sln` on machines without the game. `Pathoschild.Stardew.ModBuildConfig` needs legal Stardew Valley assemblies before it can compile.

## Build against a real game installation

Install Stardew Valley 1.6 and SMAPI 4.1 or newer. Create `~/stardewvalley.targets`:

```xml
<Project>
  <PropertyGroup>
    <GamePath>/absolute/path/to/Stardew Valley</GamePath>
  </PropertyGroup>
</Project>
```

Then build:

```bash
./.tools/dotnet/dotnet build stardew-echo-mod/src/EchoFarm.Mod/EchoFarm.Mod.csproj
```

For a normal Steam installation on macOS, the game path is usually:

```text
~/Library/Application Support/Steam/steamapps/common/Stardew Valley/Contents/MacOS
```

Copy the generated EchoFarm mod folder into the game's `Mods` directory if automatic deployment is disabled.

The adapter targets the Stardew 1.6 API surface. The checked calls include `Game1.tileSize`, `FarmerRenderer.draw`, `GameLocation.isTilePassable`, `Character.GetToolLocation`, Net collection `Pairs`, `Crop.GetData`, `HoeDirt.destroyCrop`, `ItemRegistry.Create`, and `Chest.addItem`. This source-level audit does not replace compiling and running against a legally installed game.

## Run with the Go core

Release archives bundle the correct Go service binary and start it automatically. Set the model provider in SMAPI's generated `config.json`:

```json
{
  "CoreUrl": "http://127.0.0.1:18471",
  "AutoStartCore": true,
  "CoreStartupTimeoutSeconds": 10,
  "ModelMode": "openai",
  "ModelBaseUrl": "https://your-endpoint/v1",
  "ModelName": "your-model",
  "MaxModelCallsPerSession": 32,
  "MaxReportedTokensPerSession": 100000
}
```

Set `ECHOFARM_MODEL_API_KEY` in the environment used to launch SMAPI. EchoFarm intentionally has no API-key field and never writes the key into `config.json`. `DatabasePath` and `CoreExecutablePath` can be overridden; empty values use the OS-local application-data directory and bundled `core/echofarm-core[.exe]` respectively.

F9 shows current-session calls, reported tokens, failures, and recent latency. If the provider omits token metadata it shows `unknown` instead of estimating. Reaching either configured session limit stops Echo before another model call; currency cost is intentionally not calculated from mutable external pricing.

For source development, start the Go service before launching SMAPI:

```bash
cd echofarm-core
ECHOFARM_MODEL_MODE=fixture go run ./cmd/echofarm
```

Fixture mode is only for transport and gameplay-loop smoke testing. For actual learning, configure the OpenAI-compatible Eino model variables described in `echofarm-core/README.md`.

## Live smoke checklist

Use a disposable test save:

1. Start the Go core and launch the game through SMAPI.
2. Press F7, water crops, harvest one ripe parsnip, put it in the intended chest, and press F7 again.
3. Confirm the SMAPI console says the routine was learned.
4. Sleep, place a new mature parsnip at a different tile, keep the player's backpack count visible, and press F8.
5. Confirm a cyan translucent Echo appears and acts while the player remains controllable.
6. Confirm Echo walks to the new crop, harvests it without changing the player's backpack, and deposits it into the demonstrated chest.
7. Fill that chest before another run and confirm Echo retains any rejected items and stops instead of deleting them or looping.
8. Empty the Echo watering state and confirm it requests a refill before more watering.
9. Save or return to title during an action and confirm Echo stops without blocking the game.
10. Teach the same routine on two sunny days, then teach a rainy-day harvest; press F9 and confirm the sunny routine remains stable while the rainy behavior is stored separately.
11. Summon Echo, water two crops yourself, and confirm the panel reports your watering intent while Echo selects an unclaimed harvest target.
12. Let a full-inventory harvest fail, start a new Echo session, and confirm it deposits before attempting another harvest.
13. Press F10 while a chest decision is pending, deposit into another chest within 20 seconds, then confirm a later matching decision uses that corrected chest and F9 identifies the player-correction evidence.

Do not use a personal save until the live-game checks pass. The adapter never writes save files directly.
