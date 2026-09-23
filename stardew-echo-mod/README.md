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
- player movement and semantic tool/action events are recorded;
- farm crops, chests, water sources, weather, time, and Echo resources are mapped into `WorldSnapshot`;
- Go actions are queued back onto the game update thread and use bounded grid pathfinding;
- the Echo is rendered as a translucent copy of the player's appearance;
- movement is animated tile by tile; watering and watering-can refill are wired;
- harvest and deposit currently return explicit recoverable failures instead of mutating the player's inventory.

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

## Run with the Go core

Start the Go service before launching SMAPI:

```bash
cd echofarm-core
ECHOFARM_MODEL_MODE=fixture go run ./cmd/echofarm
```

Fixture mode is only for transport and gameplay-loop smoke testing. For actual learning, configure the OpenAI-compatible Eino model variables described in `echofarm-core/README.md`.

## Live smoke checklist

Use a disposable test save:

1. Start the Go core and launch the game through SMAPI.
2. Press F7, walk through a small crop patch, water it, and press F7 again.
3. Confirm the SMAPI console says the routine was learned.
4. Sleep, alter the crop layout, and press F8.
5. Confirm a cyan translucent Echo appears and acts while the player remains controllable.
6. Empty the Echo watering state and confirm it requests a refill before more watering.
7. Save or return to title during an action and confirm Echo stops without blocking the game.

Do not use a personal save until the live-game checks pass. The adapter never writes save files directly.
