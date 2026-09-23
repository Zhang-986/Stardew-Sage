# EchoFarm SMAPI Bridge Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Build the tested .NET bridge foundation that records normal Stardew play, exchanges strict contracts with the Go/Eino core, and drives a safe Echo session state machine ready for binding to SMAPI.

**Architecture:** A game-independent .NET library owns wire contracts, event compression, HTTP transport, correlation validation, and the Echo lifecycle. A thin SMAPI project adapts game events and rendering to those tested interfaces. Because Stardew Valley and SMAPI are not installed on this machine, the reusable bridge library and tests must compile locally; the concrete mod project is packaged as a clearly marked integration target whose final build requires a real game installation.

**Tech Stack:** C# 12, .NET 8 SDK, `net6.0` bridge library, `System.Text.Json`, `HttpClient`, xUnit; SMAPI 4.1+ via `Pathoschild.Stardew.ModBuildConfig` 4.4.0.

---

## File Map

- `global.json`: pins a compatible .NET 8 SDK family.
- `stardew-echo-mod/EchoFarm.sln`: build entry point for game-independent bridge tests.
- `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/*.cs`: DTOs that serialize exactly like the Go contracts.
- `stardew-echo-mod/src/EchoFarm.Bridge/Recording/*.cs`: teaching-session state and movement compression.
- `stardew-echo-mod/src/EchoFarm.Bridge/Transport/*.cs`: strict localhost API client.
- `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/*.cs`: Echo lifecycle, correlation guard, and bridge-facing ports.
- `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/*.cs`: executable contract and behavior tests.
- `stardew-echo-mod/src/EchoFarm.Mod/`: SMAPI entry point, manifest, game adapter, and translucent Echo renderer.
- `stardew-echo-mod/README.md`: local build, game path, installation, and manual smoke procedure.

### Task 1: Bootstrap an Isolated .NET Solution

**Files:**
- Create: `global.json`
- Create: `stardew-echo-mod/EchoFarm.sln`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/EchoFarm.Bridge.csproj`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/EchoFarm.Bridge.Tests.csproj`
- Modify: `.gitignore`

- [x] **Step 1: Install the .NET 8 SDK locally**

Install into `.tools/dotnet` using Microsoft's `dotnet-install.sh`; add `.tools/`, `bin/`, and `obj/` entries to `.gitignore`. Do not modify system-wide runtime configuration.

- [x] **Step 2: Create the solution and test project**

The bridge library targets `net6.0` for Stardew 1.6 compatibility; tests target `net8.0`. Add `Microsoft.NET.Test.Sdk`, `xunit`, `xunit.runner.visualstudio`, and `coverlet.collector` only to the test project.

- [x] **Step 3: Verify the empty baseline**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: PASS with the generated smoke test, proving the local SDK and NuGet restore work.

- [x] **Step 4: Commit**

```bash
git add .gitignore global.json stardew-echo-mod
git commit -m "build(echofarm): bootstrap dotnet bridge solution"
```

### Task 2: Mirror Go Contracts in C#

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/WorldSnapshot.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/Demonstration.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/PlayerMemory.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/Actions.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/JsonOptions.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Contracts/JsonContractTests.cs`

- [x] **Step 1: Write failing golden JSON tests**

Deserialize `demo/fixtures/morning-teaching.json`, `changed-rainy-farm.json`, and `empty-can-farm.json` with unknown-member rejection. Re-serialize them and assert camel-case properties such as `saveId`, `snapshotVersion`, `wateringCan`, and `targetId`; assert unknown action kinds fail enum conversion.

- [x] **Step 2: Run the contract tests and verify red**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln --filter FullyQualifiedName~JsonContractTests`

Expected: FAIL because the DTOs and serializer configuration do not exist.

- [x] **Step 3: Implement contracts and strict JSON options**

Use string-valued enums with exact snake-case wire values. Define `Position`, `Crop`, `WaterSource`, `Chest`, `ToolState`, `InventorySummary`, `WorldSnapshot`, `StateDelta`, `DemonstrationEvent`, `Demonstration`, `PlayerModel`, `SkillProgram`, `HighLevelAction`, `ActionResult`, `LearnResponse`, and `ActionResponse`. Capture unmapped fields through a shared `JsonExtensionData` base and reject any non-empty extension map after deserialization so the behavior also works on `net6.0`.

- [x] **Step 4: Run tests and commit**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: PASS.

```bash
git add stardew-echo-mod
git commit -m "feat(echofarm-mod): share strict bridge contracts"
```

### Task 3: Record a Teaching Session Without Frame Noise

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Recording/TeachingRecorder.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Recording/ObservedGameEvent.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Recording/TeachingRecorderTests.cs`

- [x] **Step 1: Write failing recorder tests**

Assert that `Start` creates a unique teaching session, movement observations are coalesced to one event per changed tile, semantic actions preserve before/after deltas and failures, `Stop` returns an immutable demonstration, and calls outside the recording state are ignored or rejected deterministically.

- [x] **Step 2: Verify red**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln --filter FullyQualifiedName~TeachingRecorderTests`

Expected: FAIL because `TeachingRecorder` is missing.

- [x] **Step 3: Implement the recorder**

Expose `Start(saveId, tick)`, `Observe(ObservedGameEvent)`, `Stop(tick)`, and `Cancel()`. Generate IDs locally, copy collections on output, and never retain game object references. Coalesce consecutive moves to the latest changed tile while preserving semantic actions exactly.

- [x] **Step 4: Run tests and commit**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: PASS.

```bash
git add stardew-echo-mod
git commit -m "feat(echofarm-mod): record player demonstrations"
```

### Task 4: Implement the Local Go-Core Client

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/IEchoFarmClient.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmClient.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmException.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Transport/EchoFarmClientTests.cs`

- [x] **Step 1: Write failing HTTP tests**

Use a fake `HttpMessageHandler` to verify exact paths and JSON for learn, next-action, and action-result requests. Assert 503 becomes `ModelUnavailableException`, malformed JSON is rejected, and responses with stale save/session/snapshot fields raise `StaleActionException`.

- [x] **Step 2: Verify red**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln --filter FullyQualifiedName~EchoFarmClientTests`

Expected: FAIL because the client does not exist.

- [x] **Step 3: Implement transport and correlation checks**

Use one injected `HttpClient`, a 35-second cancellation timeout, `application/json`, strict serializer options, `EnsureSuccess` logic that never includes provider response bodies, and request-specific response correlation before returning an action.

- [x] **Step 4: Run tests and commit**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: PASS.

```bash
git add stardew-echo-mod
git commit -m "feat(echofarm-mod): connect to local Go core"
```

### Task 5: Build the Echo Lifecycle and Safety Gate

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoSession.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/ActionSafetyGate.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/IGamePort.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoSessionTests.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/ActionSafetyGateTests.cs`

- [x] **Step 1: Write failing lifecycle tests**

Assert transitions `Idle -> Recording -> Learning -> Ready -> Acting -> AwaitingResult`, rejection of double starts, stop-on-save/title/timeout, one in-flight action at a time, and replan after `path_blocked`.

- [x] **Step 2: Write failing safety tests**

Reject stale snapshots, unknown targets, watering during rain, watering with an empty can, harvesting immature crops, refilling a full can, and depositing into a missing chest. Accept the matching valid cases.

- [x] **Step 3: Verify red, then implement minimal runtime**

`EchoSession.TickAsync` obtains a fresh `WorldSnapshot`, asks the Go core for one action, passes it through `ActionSafetyGate`, executes through `IGamePort`, captures a new snapshot, and reports the result. Cancellation always returns to `Idle` without propagating into the game loop.

- [x] **Step 4: Run tests and commit**

Run: `.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: PASS.

```bash
git add stardew-echo-mod
git commit -m "feat(echofarm-mod): coordinate safe Echo sessions"
```

### Task 6: Add the Thin SMAPI Adapter Skeleton

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Mod/EchoFarm.Mod.csproj`
- Create: `stardew-echo-mod/src/EchoFarm.Mod/manifest.json`
- Create: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Mod/StardewGamePort.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Mod/WorldSnapshotMapper.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Mod/EchoRenderer.cs`

- [x] **Step 1: Create the SMAPI project metadata**

Target `net6.0`, reference `EchoFarm.Bridge`, and reference `Pathoschild.Stardew.ModBuildConfig` 4.4.0 with deploy/zip disabled by default. `manifest.json` declares minimum SMAPI 4.1.0 and the unique ID `Zhang986.EchoFarm`.

- [x] **Step 2: Implement the adapter boundary**

`ModEntry` owns hotkeys and SMAPI event subscription. `StardewGamePort` maps crop/water/chest IDs, gathers snapshots, and schedules every mutation on the update tick. `EchoRenderer` draws a translucent player-like sprite without inserting a social NPC. No model prompts, preferences, or planning rules belong in this project.

- [x] **Step 3: Document the unavailable local integration gate**

Do not claim the SMAPI project compiles on this machine. Record the exact missing prerequisites: a legal Stardew Valley 1.6 installation, SMAPI 4.1+, and `GamePath` in `~/stardewvalley.targets`. Keep it outside the default solution until those references exist, while keeping all reusable bridge code compiled and tested.

- [x] **Step 4: Commit**

```bash
git add stardew-echo-mod
git commit -m "feat(echofarm-mod): scaffold the SMAPI game adapter"
```

### Task 7: Verify .NET-to-Go Compatibility

**Files:**
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Integration/GoCoreContractTests.cs`
- Create: `stardew-echo-mod/README.md`
- Modify: `README.md`

- [x] **Step 1: Write a cross-process integration test**

Start the already-built Go core in fixture mode on an ephemeral loopback port and temporary SQLite path. Submit the C# teaching DTO, then submit rainy and empty-can snapshots. Assert `harvest_target` and `refill_can`, respectively. Always terminate the child process.

- [x] **Step 2: Run complete verification**

Run:

```bash
.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln
(cd echofarm-core && go test -race ./... && go vet ./...)
./demo/run-core-demo.sh
```

Expected: all commands pass. The SMAPI adapter remains an explicit game-machine build gate until a Stardew installation is supplied.

- [x] **Step 3: Document install and live-game smoke steps**

Document building `EchoFarm.Mod.csproj` with a real `GamePath`, copying the generated mod folder, pressing F7 to record, pressing F8 to summon, verifying translucent rendering, and testing rain/new-crop/empty-can behavior without using a personal save.

- [x] **Step 4: Commit**

```bash
git add README.md stardew-echo-mod docs/superpowers/plans/2026-09-24-smapi-bridge-foundation.md
git commit -m "test(echofarm-mod): verify Go bridge compatibility"
```
