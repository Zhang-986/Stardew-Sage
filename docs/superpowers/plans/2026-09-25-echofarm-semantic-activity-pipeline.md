# EchoFarm Semantic Activity Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a versioned, evidence-backed activity pipeline that classifies tree chopping, rock breaking, mine-floor traversal, and fishing outcomes and turns them into persistent AI player traits visible in F9.

**Architecture:** Keep Stardew-specific inspection in `EchoFarm.Mod`, move multi-tick episode assembly into a pure `EchoFarm.Bridge` state machine, and serialize one strict cross-language event contract. Go validates and segments the events before Eino sees them; deterministic model merging owns confidence, while autonomous mutations remain behind existing capability gates.

**Tech Stack:** C#/.NET 6 bridge and SMAPI adapter, xUnit, Go 1.24.1, CloudWeGo Eino, SQLite, JSON Schema, Bash/jq cross-process fixtures.

---

### Task 1: Version the cross-language activity contract

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/Demonstration.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Contracts/JsonContractTests.cs`
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/domain/validate.go`
- Modify: `echofarm-core/internal/domain/validate_test.go`
- Modify: `contracts/demonstration.schema.json`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Recording/TeachingRecorder.cs`

- [x] **Step 1: Write failing C# and Go contract tests**

Add JSON round-trip coverage for schema version 2 and all five new event kinds. Assert that `location`, `timeOfDay`, `targetKind`, `durationTicks`, `healthDelta`, `mineFloorDelta`, and signed item deltas survive serialization. Add Go validation cases that reject a new event in a version-1 demonstration, zero item quantities, duplicate item IDs, more than 32 item deltas, negative duration, and unsupported target kinds.

- [x] **Step 2: Verify RED**

Run:

```bash
cd stardew-echo-mod
dotnet test EchoFarm.sln --filter 'FullyQualifiedName~JsonContractTests'
cd ../echofarm-core
go test ./internal/domain -run 'TestDemonstration.*V2|TestDemonstrationRejectsInvalidActivityEvidence' -count=1
```

Expected: compile failures because the new event kinds and fields do not exist.

- [x] **Step 3: Add the strict contracts**

Add these cross-language shapes with the existing camel-case JSON policy:

```go
type ItemDelta struct {
    ItemID string `json:"itemId"`
    Name string `json:"name,omitempty"`
    Quantity int `json:"quantity"`
}

type StateDelta struct {
    EnergyDelta int `json:"energyDelta"`
    WaterDelta int `json:"waterDelta"`
    InventoryDelta int `json:"inventoryDelta"`
    HealthDelta int `json:"healthDelta,omitempty"`
    MineFloorDelta int `json:"mineFloorDelta,omitempty"`
}
```

Extend `Demonstration` with `SchemaVersion`; treat `0` as legacy version 1 during Go validation. Extend `DemonstrationEvent` with `Location`, `TimeOfDay`, `TargetKind`, `DurationTicks`, and `ItemDeltas`. Add event kinds `chop_tree`, `break_rock`, `enter_mine_floor`, `fish_caught`, and `fish_escaped`. The C# recorder emits `SchemaVersion = 2`.

- [x] **Step 4: Implement bounded validation and schema rules**

Accept only schema versions 1 and 2. Require version 2 for new event kinds, `durationTicks >= 0`, `timeOfDay` in `0..2600` when present, at most 32 item deltas, non-empty unique item IDs, non-zero quantities, and target kinds `crop`, `water`, `chest`, `tree`, `rock`, `mine_floor`, `fish`, or empty for movement. Keep existing version-1 fixtures valid.

- [x] **Step 5: Verify and commit**

Run both targeted suites and commit:

```bash
git commit -m "feat(echofarm): version semantic activity events"
```

### Task 2: Add a pure multi-tick activity classifier

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Recording/SemanticActivityTracker.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Recording/SemanticActivityTrackerTests.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Recording/ObservedGameEvent.cs`

- [x] **Step 1: Write failing classifier tests**

Cover tree and rock episodes that aggregate repeated progress signals into one event; target disappearance as success; target replacement, cancellation, and a 600-tick timeout as bounded failures; confirmed mine-floor transition; caught fish with item evidence; and escaped fish without a fabricated item. Verify a second concurrent episode is rejected and `Reset` discards pending state without emitting evidence.

- [x] **Step 2: Verify RED**

Run:

```bash
cd stardew-echo-mod
dotnet test EchoFarm.sln --filter 'FullyQualifiedName~SemanticActivityTrackerTests'
```

Expected: compile failure because `SemanticActivityTracker` and its signal contracts do not exist.

- [x] **Step 3: Implement the classifier state machine**

Expose this game-neutral API:

```csharp
public enum SemanticActivityFamily { TreeChopping, RockBreaking, Fishing }

public sealed record ActivitySample(
    long Tick,
    string Location,
    int TimeOfDay,
    Position Position,
    string TargetId,
    string TargetKind,
    string Tool,
    GameStateSample State,
    IReadOnlyList<ItemDelta> ItemDeltas,
    bool TargetPresent
);

public sealed class SemanticActivityTracker
{
    public bool TryBegin(SemanticActivityFamily family, ActivitySample sample);
    public ObservedGameEvent? Observe(ActivitySample sample);
    public ObservedGameEvent? CompleteFishing(ActivitySample sample, bool caught);
    public ObservedGameEvent? RecordMineTransition(ActivitySample before, ActivitySample after, int floorDelta);
    public ObservedGameEvent? Cancel(long tick, string errorCode);
    public void Reset();
}
```

Normalize item deltas by item ID, use the first sample as the episode baseline, and emit only the bounded codes `activity_timeout`, `target_changed`, `location_changed`, `activity_cancelled`, and `fish_escaped`.

- [x] **Step 4: Verify and commit**

Run the full bridge suite and commit:

```bash
git commit -m "feat(mod): classify multi-tick player activities"
```

### Task 3: Teach Go to segment and learn the new activities

**Files:**
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/domain/validate.go`
- Modify: `echofarm-core/internal/domain/validate_test.go`
- Modify: `echofarm-core/internal/trace/segmenter.go`
- Modify: `echofarm-core/internal/trace/segmenter_test.go`
- Modify: `echofarm-core/internal/intelligence/prompts.go`
- Modify: `echofarm-core/internal/intelligence/fixture_generator.go`
- Modify: `echofarm-core/internal/intelligence/learning_graph_test.go`
- Modify: `echofarm-core/internal/modeling/merger.go`
- Modify: `echofarm-core/internal/modeling/merger_test.go`

- [x] **Step 1: Write failing segment and trait tests**

Prove `chop_tree` maps to one `woodcutting` segment, `break_rock` plus mine transitions map to `mining` and `mine_traversal`, and caught/escaped outcomes map to `fishing`. Assert segment totals preserve health, floor, duration, and normalized item deltas. Add trait-validation and deterministic merge cases for `activity_order`, `resource_priority`, `mine_exit_policy`, and `fishing_context` with real evidence IDs.

- [x] **Step 2: Verify RED**

Run:

```bash
cd echofarm-core
go test ./internal/trace ./internal/modeling ./internal/intelligence \
  -run 'TestSegment.*Activity|TestMerge.*Lifestyle|TestLearningGraph.*Activity' -count=1
```

Expected: compile failures for missing behavior and preference kinds.

- [x] **Step 3: Extend deterministic segmentation**

Add behavior kinds `woodcutting`, `mining`, `mine_traversal`, and `fishing`. Extend `BehaviorSegment` with `DurationTicks`, `HealthDelta`, `MineFloorDelta`, and normalized `ItemDeltas`. Movement remains route context; adjacent events merge only when their behavior kind matches, and item quantities merge by stable item ID in lexical order.

- [x] **Step 4: Extend bounded AI learning**

Add the four preference keys and update the learning prompt to describe their allowed meaning. The prompt must require evidence IDs and forbid conclusions from movement alone. Update the fixture generator to emit activity order, repeated resource priority, mine exit policy, and fishing context only when corresponding segments exist; it must not add unsupported executable skill steps.

- [x] **Step 5: Verify and commit**

Run Go race tests for the affected packages and commit:

```bash
git commit -m "feat(ai): learn evidence-backed lifestyle activities"
```

### Task 4: Expose learned activity evidence in F9

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/PlayerMemory.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoMemoryPresenter.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoMemoryPresenterTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Contracts/JsonContractTests.cs`

- [x] **Step 1: Write failing presenter tests**

Create a memory view containing stable `activity_order`, `resource_priority`, `mine_exit_policy`, and `fishing_context` traits. Assert F9 renders compact Chinese labels, confidence, and no raw evidence IDs, item IDs, prompt text, or provider content. Keep the ten-line panel bound.

- [x] **Step 2: Verify RED**

Run:

```bash
cd stardew-echo-mod
dotnet test EchoFarm.sln --filter 'FullyQualifiedName~EchoMemoryPresenterTests|FullyQualifiedName~JsonContractTests'
```

Expected: enum deserialization or label assertions fail.

- [x] **Step 3: Add the C# trait contracts and labels**

Map the new keys to `活动顺序`, `资源偏好`, `下矿退出习惯`, and `钓鱼场景`. Continue to show only the two strongest stable traits so the overlay remains bounded.

- [x] **Step 4: Verify and commit**

Run the full .NET suite and commit:

```bash
git commit -m "feat(mod): show learned lifestyle traits in F9"
```

### Task 5: Wire Stardew sensors without enabling mutations

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/StardewGamePort.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/WorldSnapshotMapper.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Recording/TeachingRecorder.cs`
- Modify: `docs/echofarm/windows-smoke-checklist.md`

- [x] **Step 1: Add sensor adapter tests where game-neutral behavior permits**

Keep all episode state in `SemanticActivityTracker`. The Mod translates axe/tree, pickaxe/world-object, mine warp, and fishing lifecycle observations into `ActivitySample`; no new `ActionKind` or execution path is added.

- [x] **Step 2: Wire lifecycle events**

On teaching-mode tool input, start or advance a tree/rock episode. On each update tick, poll the pending target and complete it only on a proven state transition. Subscribe to player warp events for mine-floor transitions. Start fishing on rod use and complete only from a confirmed catch/escape signal. Reset every pending activity on save, day change, title return, or Mod exit.

- [x] **Step 3: Keep unsupported execution explicit**

Do not add chop, mine, or fish to `ActionKind`. Add the four observation categories to F9 setup diagnostics as `learn-only`; keep `Harvest: disabled` unchanged. Add native signal checks to the Windows disposable-save checklist.

- [ ] **Step 4: Build on a legal game installation when available**

Run:

```powershell
$game = "C:\Program Files (x86)\Steam\steamapps\common\Stardew Valley"
.\scripts\windows\Install-EchoFarm.ps1 -Doctor -GamePath $game
.\scripts\windows\Install-EchoFarm.ps1 -Build -GamePath $game
```

Expected: the real `EchoFarm.Mod.dll` builds and the evidence JSON leaves the four native activity checks pending until observed in game. On machines without the legal assemblies, record the external gate and do not claim native certification.

- [x] **Step 5: Commit**

```bash
git commit -m "feat(mod): observe extended Stardew activities"
```

### Task 6: Prove the cross-process learning path

**Files:**
- Create: `demo/fixtures/activity-day-1.json`
- Create: `demo/fixtures/activity-day-2.json`
- Create: `demo/run-activity-learning-demo.sh`
- Modify: `.github/workflows/ci.yml`
- Modify: `README.md`
- Modify: `release/nexus/BUILD-EVIDENCE.md`

- [x] **Step 1: Write the failing demo**

Teach two version-2 demonstrations containing woodcutting, mine traversal, and fishing evidence. Read `/v1/player-model` and `/v1/echo/memory`; require stable lifestyle traits to cite both days and appear in the memory payload. Assert no unsupported action kind is returned by `/v1/echo/next-action`.

- [x] **Step 2: Verify RED then make fixtures deterministic**

Run:

```bash
./demo/run-activity-learning-demo.sh
```

Expected before Tasks 1–4: schema or trait assertions fail. After implementation: `EchoFarm semantic activity learning demo passed.`

- [x] **Step 3: Add the demo to CI and document the boundary**

Run the demo in Linux `verify`. Document the actual supported observation categories, the `learn-only` execution boundary, version-2 payload privacy, and the requirement for Windows native signal certification.

- [x] **Step 4: Run the complete release matrix**

Run Go race/vet, the .NET Release suite, PowerShell setup tests, all four demos, package smoke, twenty iterations of the existing SQLite concurrency races, deterministic four-platform sidecar rebuild, native macOS health smoke, `git diff --check`, and the tracked-source secret scan.

- [x] **Step 5: Commit, push, and verify exact-SHA CI**

```bash
git commit -m "docs(echofarm): verify semantic activity learning"
git push origin codex/living-valley-director
```

Require both `verify` and `windows-sidecar` to pass for the exact final SHA. Do not mark native Stardew activity checks passed without owner-machine evidence.
