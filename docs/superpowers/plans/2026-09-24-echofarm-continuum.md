# EchoFarm Continuum Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a four-day, end-to-end EchoFarm demo in which evidence-backed player traits strengthen across demonstrations, context prevents false contradictions, Echo avoids the player's live targets, and every decision is explainable from a persisted ledger.

**Architecture:** Eino extracts bounded trait observations and high-level actions; deterministic Go services merge long-term memory, filter player-claimed targets, and persist revision/decision evidence in SQLite. The C# bridge adds recent-player activity and a read-only memory presenter while preserving the existing game-thread safety boundary.

**Tech Stack:** Go 1.24, CloudWeGo Eino, SQLite, HTTP/JSON, C#/.NET 8 tests, SMAPI/.NET 6 adapter, Bash/JQ demo automation.

---

### Task 1: Add versioned trait and collaboration contracts

**Files:**
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/domain/validate.go`
- Modify: `echofarm-core/internal/domain/validate_test.go`
- Modify: `contracts/demonstration.schema.json`
- Modify: `contracts/world-snapshot.schema.json`
- Modify: `contracts/player-model.schema.json`
- Create: `contracts/echo-memory.schema.json`

- [ ] **Step 1: Write failing domain validation tests**

Add table tests proving that a demonstration accepts positive `day` plus a known weather, trait observations reject unknown keys, strength outside `[0,1]`, and evidence IDs not supplied by the caller, while player activity rejects empty targets and negative ticks.

```go
observation := domain.TraitObservation{
    Key: domain.PreferenceTaskOrder, Value: "watering,harvesting,depositing",
    Context: domain.TraitContextSunny, SupportingEventIDs: []string{"water-1"}, Strength: 0.7,
}
if err := observation.Validate(map[string]struct{}{"water-1": {}}); err != nil { t.Fatal(err) }
```

- [ ] **Step 2: Run the domain tests and confirm failure**

Run: `cd echofarm-core && go test ./internal/domain`

Expected: compilation fails because the new contract types do not exist.

- [ ] **Step 3: Implement the domain types and validation**

Add `TraitContext`, `TraitObservation`, `TraitMemory`, `LearningChange`, `PlayerActivity`, `PlayerIntent`, `CoordinationContext`, `DecisionRecord`, and `EchoMemoryView`. Extend `Demonstration` with optional `day` and `weather`, `PlayerModel` with `learnedThroughDay` and `traits`, and `WorldSnapshot` with `recentPlayerActions`.

Validation permits only existing `PreferenceKey` values, contexts `any|sunny|rainy|storm|snow`, intents `unknown|watering|harvesting|depositing`, finite strengths/confidences from zero through one, and non-empty scoped evidence.

- [ ] **Step 4: Update the JSON schemas and run tests**

Run: `cd echofarm-core && go test ./internal/domain`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/internal/domain contracts
git commit -m "feat(echofarm): add continuum memory contracts"
```

### Task 2: Build the deterministic long-term model merger

**Files:**
- Create: `echofarm-core/internal/modeling/merger.go`
- Create: `echofarm-core/internal/modeling/merger_test.go`

- [ ] **Step 1: Write failing merger tests**

Cover first observation capped at `0.75`, repeated evidence using `1-(1-old)*(1-strength*0.5)`, same-context contradiction decay, rainy/sunny isolation, twelve-reference retention, legacy-field projection, and no mutation of the input model.

```go
next, change, err := modeling.Merge(nil, demo, []domain.TraitObservation{observation})
if err != nil { t.Fatal(err) }
if next.Revision != 1 || next.Traits[0].ObservationCount != 1 { t.Fatalf("unexpected model: %+v", next) }
if change.Kind != domain.LearningChangeAdded { t.Fatalf("unexpected change: %+v", change) }
```

- [ ] **Step 2: Run the test and confirm failure**

Run: `cd echofarm-core && go test ./internal/modeling`

Expected: package or `Merge` is missing.

- [ ] **Step 3: Implement merge and projection**

`Merge(existing *PlayerModel, demonstration Demonstration, observations []TraitObservation)` copies the old model, scopes event evidence as `<demo-id>:<event-id>`, applies the confidence rules, advances one revision, sorts traits deterministically, derives `CommonTaskOrder`, `PreferredChestID`, `EnergyReserve`, and `RouteStyle`, and returns a compact `LearningChange`.

- [ ] **Step 4: Run focused and domain tests**

Run: `cd echofarm-core && go test ./internal/modeling ./internal/domain`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/internal/modeling
git commit -m "feat(echofarm): merge multi-day player evidence"
```

### Task 3: Change Eino learning from model replacement to trait extraction

**Files:**
- Modify: `echofarm-core/internal/intelligence/contracts.go`
- Modify: `echofarm-core/internal/intelligence/prompts.go`
- Modify: `echofarm-core/internal/intelligence/learning_graph.go`
- Modify: `echofarm-core/internal/intelligence/learning_graph_test.go`
- Modify: `echofarm-core/internal/intelligence/fixture_generator.go`
- Modify: `echofarm-core/internal/intelligence/fixture_generator_test.go`

- [ ] **Step 1: Write failing graph tests**

Assert the graph returns `LearningInference{Observations, Skill}`, rejects unknown preference keys, forged evidence IDs and invalid strengths, and accepts a rainy contextual observation without requiring a replacement player model.

- [ ] **Step 2: Run intelligence tests and confirm failure**

Run: `cd echofarm-core && go test ./internal/intelligence`

Expected: tests fail against the old `LearningResult` contract.

- [ ] **Step 3: Implement bounded trait extraction**

Replace the model-authored `PlayerModel` with:

```go
type LearningInference struct {
    Observations []domain.TraitObservation `json:"observations"`
    Skill domain.SkillProgram `json:"skill"`
}
```

The prompt instructs the model to infer only allowed keys, distinguish weather context, cite current demonstration events, and leave long-term confidence to deterministic code. Update the fixture generator to emit day/weather-aware observations and preserve the current skill behavior.

- [ ] **Step 4: Run intelligence tests**

Run: `cd echofarm-core && go test ./internal/intelligence`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/internal/intelligence
git commit -m "feat(echofarm): extract bounded player traits with Eino"
```

### Task 4: Persist learning revisions and make teaching idempotent

**Files:**
- Modify: `echofarm-core/internal/memory/store.go`
- Modify: `echofarm-core/internal/memory/sqlite.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`
- Modify: `echofarm-core/internal/learning/service.go`
- Modify: `echofarm-core/internal/learning/service_test.go`

- [ ] **Step 1: Write failing persistence and service tests**

Prove an old database migrates, learning revision plus model plus skill commit atomically, the revision can be loaded by demonstration ID, and submitting the same demonstration twice returns the stored outcome without invoking Eino or increasing observation counts.

- [ ] **Step 2: Run tests and confirm failure**

Run: `cd echofarm-core && go test ./internal/memory ./internal/learning`

Expected: missing revision APIs and old learner result shape.

- [ ] **Step 3: Add the revision table and store methods**

Add `learning_revisions(save_id, demonstration_id, model_revision, change_json, created_at)` with a composite primary key. Introduce `LearningCommit` and `GetLearningOutcome`; extend `SaveLearning` so the demonstration, model, skill and change are one transaction.

- [ ] **Step 4: Integrate `modeling.Merge` into `learning.Service`**

The service first checks idempotency, asks Eino for observations, validates them, merges deterministically, validates the derived model and skill, commits once, and returns `LearningOutcome{PlayerModel, Skill, Change}`.

- [ ] **Step 5: Run focused tests**

Run: `cd echofarm-core && go test ./internal/memory ./internal/learning`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add echofarm-core/internal/memory echofarm-core/internal/learning
git commit -m "feat(echofarm): persist idempotent learning revisions"
```

### Task 5: Add live intent inference and deterministic target claims

**Files:**
- Create: `echofarm-core/internal/coordination/service.go`
- Create: `echofarm-core/internal/coordination/service_test.go`
- Modify: `echofarm-core/internal/intelligence/contracts.go`
- Modify: `echofarm-core/internal/intelligence/action_graph.go`
- Modify: `echofarm-core/internal/intelligence/action_graph_test.go`
- Modify: `echofarm-core/internal/intelligence/fixture_generator.go`

- [ ] **Step 1: Write failing coordinator tests**

Cover dominant recent intent, stale activity removal, target de-duplication, filtering a model-selected claimed crop, and safe stop when every actionable target is claimed.

- [ ] **Step 2: Run tests and confirm failure**

Run: `cd echofarm-core && go test ./internal/coordination ./internal/intelligence`

Expected: coordination package and intent inference contracts are missing.

- [ ] **Step 3: Implement intent graph and coordination context**

Add `InferIntent(ctx, IntentInput) (PlayerIntent, error)` to the intelligence boundary and compile an Eino graph with a strict structured output. `coordination.Service.Prepare` keeps successful activities within the 20-second tick window, derives claimed targets, calls intent inference, and exposes filtered actionable goals.

- [ ] **Step 4: Enforce claims after AI selection**

`coordination.Service.ValidateChoice` rejects any action targeting a player-claimed ID. The fixture actor chooses an unclaimed mature crop before unclaimed dry crops and stops with a collaboration reason when no target remains.

- [ ] **Step 5: Run focused tests**

Run: `cd echofarm-core && go test ./internal/coordination ./internal/intelligence`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add echofarm-core/internal/coordination echofarm-core/internal/intelligence
git commit -m "feat(echofarm): coordinate Echo with live player intent"
```

### Task 6: Add decision/session ledger and memory view

**Files:**
- Modify: `echofarm-core/internal/memory/store.go`
- Modify: `echofarm-core/internal/memory/sqlite.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`
- Modify: `echofarm-core/internal/policy/service.go`
- Modify: `echofarm-core/internal/policy/service_test.go`
- Create: `echofarm-core/internal/memoryview/service.go`
- Create: `echofarm-core/internal/memoryview/service_test.go`

- [ ] **Step 1: Write failing ledger tests**

Assert one next-action request records model revision, inferred intent, claimed targets, AI candidate and final action; action results attach exactly once; a memory view returns stable traits, recent learning change and latest decision.

- [ ] **Step 2: Run focused tests and confirm failure**

Run: `cd echofarm-core && go test ./internal/memory ./internal/policy ./internal/memoryview`

Expected: ledger and view methods are absent.

- [ ] **Step 3: Add SQLite tables and repository methods**

Create `echo_sessions` and `decision_records` with unique `(save_id, session_id, snapshot_version)`. Persist structured JSON rather than free-form model text and make result attachment idempotent.

- [ ] **Step 4: Integrate coordinator and ledger into policy**

Policy prepares collaboration context before the action graph, validates target claims after normal snapshot validation, and records both the candidate and accepted action. Failed or rejected AI output never crosses into the bridge; a rejected claimed target becomes a deterministic `stop_session` decision with an explanation.

- [ ] **Step 5: Implement the memory view projection**

Return traits sorted by stable status, confidence and key; include the latest learning change and decision without exposing prompts or credentials.

- [ ] **Step 6: Run tests and commit**

Run: `cd echofarm-core && go test ./internal/memory ./internal/policy ./internal/memoryview`

```bash
git add echofarm-core/internal/memory echofarm-core/internal/policy echofarm-core/internal/memoryview
git commit -m "feat(echofarm): explain decisions from an execution ledger"
```

### Task 7: Expose the Continuum HTTP API

**Files:**
- Modify: `echofarm-core/internal/httpapi/handler.go`
- Modify: `echofarm-core/internal/httpapi/handler_test.go`
- Modify: `echofarm-core/cmd/echofarm/main.go`
- Modify: `echofarm-core/cmd/echofarm/main_test.go`

- [ ] **Step 1: Write failing HTTP tests**

Verify learn returns `learningChange`, next-action accepts recent player activity, action-result persists a result before selecting again, and `GET /v1/echo/memory` validates `saveId` and returns a structured view.

- [ ] **Step 2: Run tests and confirm failure**

Run: `cd echofarm-core && go test ./internal/httpapi ./cmd/echofarm`

Expected: response and route assertions fail.

- [ ] **Step 3: Wire services and strict JSON handlers**

Keep the existing size limit and unknown-field rejection. Construct intelligence, coordination, policy, learning and memory-view services once at startup and register the new read endpoint.

- [ ] **Step 4: Run tests and commit**

Run: `cd echofarm-core && go test ./internal/httpapi ./cmd/echofarm`

```bash
git add echofarm-core/internal/httpapi echofarm-core/cmd/echofarm
git commit -m "feat(echofarm): expose continuum memory API"
```

### Task 8: Build the four-day cross-process demo

**Files:**
- Create: `demo/fixtures/day-2-sunny-teaching.json`
- Create: `demo/fixtures/day-3-rainy-teaching.json`
- Create: `demo/fixtures/day-4-coplay-farm.json`
- Create: `demo/run-continuum-demo.sh`
- Modify: `demo/run-core-demo.sh`
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add fixtures and a failing assertion script**

The script starts the real Go binary in fixture mode with a temporary database, submits three demonstrations, checks revisions `1,2,3`, verifies sunny task-order observations remain uncontradicted after rain, requests a day-four action with claimed watering targets, and checks the memory view explains an unclaimed harvest decision.

- [ ] **Step 2: Run the script and confirm failure**

Run: `./demo/run-continuum-demo.sh`

Expected: it fails until all API responses and fixture behavior are connected.

- [ ] **Step 3: Complete fixture behavior and CI entry**

Keep `run-core-demo.sh` as the quick single-day smoke and add Continuum as the deeper product demo executed by CI.

- [ ] **Step 4: Run the demo and commit**

Run: `./demo/run-continuum-demo.sh`

Expected: all four day assertions pass and the script deletes its temporary database/process.

```bash
git add demo .github/workflows/ci.yml
git commit -m "test(echofarm): demonstrate four-day co-player growth"
```

### Task 9: Extend the .NET bridge for player activity and memory presentation

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/Demonstration.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/WorldSnapshot.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/PlayerMemory.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/JsonOptions.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/PlayerActivityWindow.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoMemoryPresenter.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/IEchoFarmClient.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmClient.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoSession.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/PlayerActivityWindowTests.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoMemoryPresenterTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Contracts/JsonContractTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Transport/EchoFarmClientTests.cs`

- [ ] **Step 1: Write failing bridge tests**

Cover JSON round trips, twenty-second expiry, same-target de-duplication, save reset, stable trait ordering, concise Chinese presentation lines and the memory endpoint client call.

- [ ] **Step 2: Run tests and confirm failure**

Run: `./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: new types are missing.

- [ ] **Step 3: Implement bridge contracts, window and presenter**

Keep all types independent of Stardew assemblies. `PlayerActivityWindow.Snapshot(nowTick)` returns immutable recent events; `EchoMemoryPresenter.BuildLines(view)` returns at most eight bounded lines suitable for a HUD.

- [ ] **Step 4: Connect session snapshots and HTTP client**

Inject recent player activities into snapshots before next-action calls and add `GetMemoryAsync(saveId)` without coupling the bridge to drawing code.

- [ ] **Step 5: Run tests and commit**

Run: `./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: all existing and new tests pass.

```bash
git add stardew-echo-mod/src/EchoFarm.Bridge stardew-echo-mod/tests/EchoFarm.Bridge.Tests
git commit -m "feat(echofarm-mod): surface live player intent and memory"
```

### Task 10: Add the SMAPI memory HUD and live activity feed

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/StardewGamePort.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Mod/EchoMemoryOverlay.cs`
- Modify: `stardew-echo-mod/README.md`

- [ ] **Step 1: Add `MemoryKey` and semantic activity forwarding**

Use `F9` by default. While Echo is acting, feed successful player water/harvest/deposit observations into the activity window; reset the window on save/title transitions.

- [ ] **Step 2: Implement the read-only overlay**

On F9, fetch the memory view asynchronously, format it through the bridge presenter, cache only display lines, and draw a translucent HUD panel during `RenderedHud`. Network failure updates one status line and never blocks the update thread.

- [ ] **Step 3: Perform source-level SMAPI API audit**

Check all drawing and input calls against APIs already used by the repository. Because the machine lacks legal game assemblies, record the real-game build and play test as an external gate rather than substituting fixture DLLs.

- [ ] **Step 4: Commit**

```bash
git add stardew-echo-mod/src/EchoFarm.Mod stardew-echo-mod/README.md
git commit -m "feat(echofarm-mod): add in-game Echo memory HUD"
```

### Task 11: Verify and refresh delivery evidence

**Files:**
- Modify: `README.md`
- Modify: `echofarm-core/README.md`
- Modify: `release/nexus/description.md`
- Modify: `release/nexus/PLAYER_README.md`
- Modify: `release/nexus/BUILD-EVIDENCE.md`

- [ ] **Step 1: Run all automated verification**

```bash
cd echofarm-core && go test -race -count=1 ./... && go vet ./...
cd .. && ./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln
./demo/run-core-demo.sh
./demo/run-continuum-demo.sh
./scripts/tests/package-nexus-smoke.sh
```

Expected: all commands pass.

- [ ] **Step 2: Update product and release documentation**

Document the four-day scenario, architecture boundaries, memory HUD, exact commands and remaining real-game/Nexus gates. Update test counts from actual output and do not claim a playable SMAPI ZIP or Nexus publication.

- [ ] **Step 3: Verify repository hygiene**

Run: `git diff --check && git status --short && rg -n "API_KEY=|sk-[A-Za-z0-9]" --glob '!go.sum' .`

Expected: no whitespace errors, generated database/log/archive files, or committed credentials.

- [ ] **Step 4: Commit**

```bash
git add README.md echofarm-core/README.md release/nexus
git commit -m "docs(echofarm): document continuum demo evidence"
```
