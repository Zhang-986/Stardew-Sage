# Echo Inventory, Harvest, and Deposit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the in-game Echo harvest mature crops into its own bounded inventory and deposit those items into the selected Stardew chest without touching the player's inventory.

**Architecture:** Keep inventory and crop-lifecycle rules in the game-independent `EchoFarm.Bridge` assembly so they can be driven by xUnit. The SMAPI adapter converts real `Crop`, `Item`, and `Chest` objects at the boundary, mutates them only on the game thread, and reports explicit failure codes when a target changes or storage is full.

**Tech Stack:** C# 12, .NET 6 bridge library, xUnit on .NET 8, Stardew Valley 1.6.8 APIs, SMAPI 4.1+.

---

## File Map

- `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoInventory.cs`: bounded, quality-aware Echo-owned item stacks.
- `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/HarvestTransfer.cs`: mature-crop checks and atomic crop-to-Echo transfer result.
- `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/DepositTransfer.cs`: partial-safe Echo-to-chest transfer accounting.
- `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoInventoryTests.cs`: capacity, stacking, removal, and immutable snapshot tests.
- `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/HarvestTransferTests.cs`: immature, dead, full-inventory, one-shot, and regrowing crop tests.
- `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/DepositTransferTests.cs`: empty, complete, partial, and invalid chest acceptance tests.
- `stardew-echo-mod/src/EchoFarm.Mod/EchoAvatarState.cs`: owns the tested inventory instead of a raw dictionary.
- `stardew-echo-mod/src/EchoFarm.Mod/StardewGamePort.cs`: maps real crop yields and chests to the tested transfer model.
- `stardew-echo-mod/src/EchoFarm.Mod/WorldSnapshotMapper.cs`: exposes Echo inventory capacity and contents to the Go policy.
- `stardew-echo-mod/README.md`: documents the live-game harvest/deposit smoke scenario and integration gate.

### Task 1: Add a Bounded Echo Inventory

**Files:**
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoInventoryTests.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoInventory.cs`

- [x] **Step 1: Write failing inventory tests**

Cover these concrete behaviors: a new `(itemId, quality)` consumes one slot, the same key stacks without consuming another slot, a different quality consumes another slot, a full inventory rejects the entire incoming stack, `Remove` decrements or deletes a stack, and mutating a returned snapshot cannot mutate inventory state.

- [x] **Step 2: Verify red**

Run:

```bash
./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln --filter FullyQualifiedName~EchoInventoryTests
```

Expected: compilation fails because `EchoInventory` and `EchoItemStack` do not exist.

- [x] **Step 3: Implement the minimal inventory**

Expose `Capacity`, `FreeSlots`, `IsEmpty`, `TryAdd(EchoItemStack)`, `Remove(itemId, quality, quantity)`, and `Snapshot()`. Validate positive capacity and quantity, use ordinal item IDs, key stacks by item ID plus quality, and return copied stack values from `Snapshot()`.

- [x] **Step 4: Verify green and commit**

Run the filtered test command, then the full .NET solution. Commit only the inventory files with `feat(echofarm-mod): add Echo-owned inventory`.

### Task 2: Make Harvest Transfer Atomic and Testable

**Files:**
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/HarvestTransferTests.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/HarvestTransfer.cs`

- [x] **Step 1: Write failing harvest-transfer tests**

Use explicit crop states to assert that dead and immature crops return `target_changed`, a full inventory returns `inventory_full` without changing crop or inventory state, a one-shot crop returns `RemoveCrop = true`, and a regrowing crop returns `RemoveCrop = false`, `FullyGrown = true`, and `DayOfCurrentPhase = RegrowDays`.

- [x] **Step 2: Verify red**

Run:

```bash
./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln --filter FullyQualifiedName~HarvestTransferTests
```

Expected: compilation fails because the harvest transfer types do not exist.

- [x] **Step 3: Implement the minimal transfer**

Define immutable `CropGrowthState`, `CropGrowthTransition`, and `HarvestTransferResult` values. `HarvestTransfer.TryCollect` must validate maturity before calling `EchoInventory.TryAdd`, and only return a crop transition after the full item stack is accepted.

- [x] **Step 4: Verify green and commit**

Run the filtered test command, then the full .NET solution. Commit the transfer and tests with `feat(echofarm-mod): model atomic crop collection`.

### Task 3: Bind Real Stardew Harvesting

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/EchoAvatarState.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/StardewGamePort.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/WorldSnapshotMapper.cs`

- [x] **Step 1: Replace the raw dictionary**

Construct `EchoInventory` with 12 slots in `EchoAvatarState`. Map `Snapshot()` entries to bridge `InventoryItem` values and use `FreeSlots` directly.

- [x] **Step 2: Execute `harvest_target` on the game thread**

Resolve the `HoeDirt`, reject changed/dead/immature targets, create an Echo stack from `crop.indexOfHarvest`, call `HarvestTransfer.TryCollect`, then apply the returned transition by either `dirt.destroyCrop(false)` or setting the regrow fields. Play harvest feedback only after the transfer succeeds.

- [x] **Step 3: Audit the adapter against Stardew 1.6.8 source**

Confirm `Game1.tileSize`, `FarmerRenderer.draw`, `GameLocation.isTilePassable`, `Character.GetToolLocation`, `terrainFeatures.Pairs`, `Objects.Pairs`, `Crop.GetData`, `HoeDirt.destroyCrop`, and `ItemRegistry.Create` signatures. Record the game-install build gate instead of claiming a local SMAPI build.

- [x] **Step 4: Run all game-independent tests and commit**

Run the full .NET test suite. Commit the adapter binding with `feat(echofarm-mod): harvest crops into Echo inventory`.

### Task 4: Deposit Echo Inventory Into a Real Chest

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/StardewGamePort.cs`

- [x] **Step 1: Implement partial-safe chest transfer**

For each Echo stack, create the matching Stardew item and call `Chest.addItem`. Compute the inserted count from the returned remainder, then remove only that inserted count from Echo. Return `inventory_empty` for an empty Echo, `target_changed` for a missing chest, and `chest_full` when any remainder stays in Echo.

- [x] **Step 2: Keep execution semantics observable**

Play chest feedback only when at least one item moved. Report success only when every Echo stack was deposited so the Go policy sees a truthful post-action snapshot and can replan on a full chest.

- [x] **Step 3: Run all game-independent tests and commit**

Run the full .NET test suite. Commit the adapter binding with `feat(echofarm-mod): deposit Echo inventory into chests`.

### Task 5: Document and Verify the Vertical Slice

**Files:**
- Modify: `stardew-echo-mod/README.md`
- Modify: `docs/superpowers/plans/2026-09-24-echo-inventory-harvest-deposit.md`

- [x] **Step 1: Document the live smoke path**

Describe teaching harvest and deposit actions, summoning Echo after adding a mature crop, checking that the player's backpack is unchanged, and checking that the harvested item appears in the target chest. State that a legal Stardew/SMAPI install is required for this check.

- [x] **Step 2: Run fresh full verification**

Run:

```bash
./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln
(cd echofarm-core && go test -race ./... && go vet ./...)
./demo/run-core-demo.sh
```

Expected: all game-independent checks pass. Also run the SMAPI project build and preserve its expected missing-game-path error as the remaining external integration gate.

- [x] **Step 3: Mark checkboxes and commit**

Update this plan with completed checkboxes and commit documentation with `docs(echofarm): add harvest and deposit demo runbook`.
