# EchoFarm Reflective Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a bounded Eino reflection loop that learns reusable policy experience from failures and explicit player corrections, then returns explainable confidence and fallback actions on later decisions.

**Architecture:** Extend the existing Go domain with structured proposals, correction evidence, and policy experiences. Eino performs semantic proposal/reflection, while deterministic Go validation, matching, merging, confidence calculation, and SQLite idempotency control all durable state and game-facing actions. The C# bridge captures an explicit F10 correction without allowing model output to bypass the existing safety gate.

**Tech Stack:** Go 1.24, CloudWeGo Eino, modernc SQLite, net6.0 bridge, xUnit, SMAPI, strict JSON over loopback HTTP.

---

### Task 1: Define bounded reflective-policy contracts

**Files:**
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/domain/validate.go`
- Modify: `echofarm-core/internal/domain/validate_test.go`

- [ ] **Step 1: Write failing validation tests**

Add tests covering a valid `ActionProposal`, confidence outside `0..1`, more than two alternatives, duplicate candidates, an unknown uncertainty code, a `PlayerCorrection` whose preferred action belongs to another snapshot, and an `ExperienceObservation` with an unsupported signal.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/domain -count=1`; compilation must fail because the new contracts do not exist.

- [ ] **Step 3: Add the contracts and strict validation**

Introduce these bounded shapes:

```go
type ActionProposal struct {
    Primary              HighLevelAction   `json:"primary"`
    Alternatives         []HighLevelAction `json:"alternatives,omitempty"`
    ModelConfidence      float64           `json:"modelConfidence"`
    UncertaintyCodes     []UncertaintyCode `json:"uncertaintyCodes,omitempty"`
    AppliedExperienceIDs []string          `json:"appliedExperienceIds,omitempty"`
}

type ActionDecision struct {
    Action               HighLevelAction   `json:"action"`
    Confidence           float64           `json:"confidence"`
    Alternatives         []HighLevelAction `json:"alternatives,omitempty"`
    AppliedExperienceIDs []string          `json:"appliedExperienceIds,omitempty"`
}
```

Add closed enums for uncertainty, experience trigger, situation signal, and evidence source. Add `ExperienceObservation`, `PolicyExperience`, `ExperienceOutcome`, and `PlayerCorrection`. Validate IDs, finite confidence, current snapshot correlation, maximum candidate counts, uniqueness, supported enums, and evidence references.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/domain -count=1` and expect PASS.

### Task 2: Make Eino produce proposals and reflections

**Files:**
- Modify: `echofarm-core/internal/intelligence/contracts.go`
- Modify: `echofarm-core/internal/intelligence/action_graph.go`
- Modify: `echofarm-core/internal/intelligence/action_graph_test.go`
- Create: `echofarm-core/internal/intelligence/reflection_graph.go`
- Create: `echofarm-core/internal/intelligence/reflection_graph_test.go`
- Modify: `echofarm-core/internal/intelligence/prompts.go`
- Modify: `echofarm-core/internal/intelligence/fixture_generator.go`
- Modify: `echofarm-core/internal/intelligence/fixture_generator_test.go`

- [ ] **Step 1: Write failing proposal and reflection tests**

Require `ActionGraph.ProposeAction` and `Replan` to return strict `ActionProposal` values. Verify unknown experience IDs and duplicate alternatives are rejected. Require `ReflectionGraph.Reflect` to reject fabricated decision/correction evidence and unsupported conditions.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/intelligence -count=1`; expect compilation failures for the new graph and proposal API.

- [ ] **Step 3: Implement one-shot structured graphs**

Extend `ActionInput` with `ApplicableExperiences []domain.PolicyExperience`. Compile Eino chains whose output is `domain.ActionProposal`. Add `ReflectionInput` with exactly one of `Result` or `Correction`, and compile a Reflection Graph that returns one `domain.ExperienceObservation`. Validate model output before returning it.

Update the fixture generator so a full-inventory experience proposes `deposit_items` before `harvest_target`, includes `stop_session` as a fallback, and emits a deterministic reflection for `inventory_full` and `player_correction`.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/intelligence -count=1` and expect PASS.

### Task 3: Merge and match policy experience

**Files:**
- Create: `echofarm-core/internal/experience/merger.go`
- Create: `echofarm-core/internal/experience/merger_test.go`
- Create: `echofarm-core/internal/experience/matcher.go`
- Create: `echofarm-core/internal/experience/matcher_test.go`

- [ ] **Step 1: Write failing merger and matcher tests**

Cover deterministic IDs, `0.65` failure confidence cap, `0.85` correction cap, repeated evidence strengthening, same-scope contradiction decay, twelve-reference bounding, input immutability, exact context isolation, full-inventory signal matching, and stable top-three ordering.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/experience -count=1`; expect compilation failure because the package does not exist.

- [ ] **Step 3: Implement deterministic experience logic**

Derive signals only from `WorldSnapshot`, normalize and sort signal/action fields before hashing an experience ID, merge by semantic identity, and rank applicable experiences by confidence, observation count, then ID. Never evaluate generated free-form predicates.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/experience -count=1` and expect PASS.

### Task 4: Persist experience and corrections atomically

**Files:**
- Modify: `echofarm-core/internal/memory/store.go`
- Modify: `echofarm-core/internal/memory/sqlite.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`

- [ ] **Step 1: Write failing persistence tests**

Verify an experience outcome survives reopen, duplicate source IDs return the first canonical outcome, a correction and its experience revision commit together, and experiences stay isolated by save ID.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/memory -count=1`; expect failures because the new tables and methods are absent.

- [ ] **Step 3: Add append-only evidence and current projections**

Add `policy_experiences`, `experience_revisions`, and `player_corrections`. Implement `SaveExperienceOutcome`, `GetExperienceOutcome`, and `ListPolicyExperiences` in one transaction using `ON CONFLICT ... DO NOTHING`, then reload canonical outcomes after save.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/memory -count=1` and expect PASS.

### Task 5: Orchestrate reflection idempotently

**Files:**
- Create: `echofarm-core/internal/experience/service.go`
- Create: `echofarm-core/internal/experience/service_test.go`

- [ ] **Step 1: Write failing service tests**

Require duplicate failure/correction evidence to skip the model, require valid observations to merge against stored experiences, require canonical post-save reload, and verify model unavailability leaves durable experience unchanged.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/experience -run 'TestLearn' -count=1`; expect missing-service failures.

- [ ] **Step 3: Implement the reflection application service**

Create `LearnFromResult` and `LearnFromCorrection`. Load existing experience, call the reflector only for unseen evidence, merge deterministically, save atomically, and return the canonical persisted `ExperienceOutcome`.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/experience -count=1` and expect PASS.

### Task 6: Select safe candidates and expose confidence

**Files:**
- Modify: `echofarm-core/internal/policy/service.go`
- Modify: `echofarm-core/internal/policy/service_test.go`
- Modify: `echofarm-core/internal/domain/types.go`

- [ ] **Step 1: Write failing adaptive policy tests**

Require applicable experiences in `ActionInput`, fallback to the first safe alternative, rejection of fabricated experience IDs, deterministic confidence bonuses/penalties, low-confidence stop, reflection after a failed action, and canonical retry behavior.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/policy -count=1`; expect failures because policy still accepts a single action.

- [ ] **Step 3: Implement proposal selection**

Keep compatibility wrappers returning `HighLevelAction`, and add decision-returning methods for HTTP. Validate every candidate against the snapshot and player claims, preserve valid nonselected candidates as alternatives, calculate policy confidence with the documented formula, persist the complete proposal in `DecisionRecord`, and invoke reflection only after a newly attached failed result.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/policy -count=1` and expect PASS.

### Task 7: Add correction and decision metadata APIs

**Files:**
- Modify: `echofarm-core/internal/httpapi/handler.go`
- Modify: `echofarm-core/internal/httpapi/handler_test.go`
- Modify: `echofarm-core/internal/memoryview/service.go`
- Modify: `echofarm-core/internal/memoryview/service_test.go`
- Modify: `echofarm-core/cmd/echofarm/main.go`
- Modify: `echofarm-core/cmd/echofarm/main_test.go`

- [ ] **Step 1: Write failing HTTP and read-model tests**

Assert action responses contain confidence, ordered alternatives, and applied experience IDs. Assert `POST /v1/echo/corrections` rejects bad correlation, returns the persisted experience, and maps model outages safely. Assert memory view returns stable experiences and last-decision metadata.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/httpapi ./internal/memoryview ./cmd/echofarm -count=1`; expect failures for missing endpoints and wiring.

- [ ] **Step 3: Wire services and API**

Build the Reflection Graph and experience service in `buildHandler`, inject them into policy/HTTP, preserve existing response field `action`, and expose only validated structured summaries through the memory endpoint.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/httpapi ./internal/memoryview ./cmd/echofarm -count=1` and expect PASS.

### Task 8: Add C# correction capture and metadata contracts

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/Actions.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/PlayerMemory.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/IEchoFarmClient.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmClient.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoSession.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/CorrectionCapture.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/EchoMemoryOverlay.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoMemoryPresenter.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Contracts/JsonContractTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Transport/EchoFarmClientTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoSessionTests.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/CorrectionCaptureTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoMemoryPresenterTests.cs`

- [ ] **Step 1: Write failing bridge tests**

Add strict JSON tests for proposal/correction/experience fields, transport tests for correction correlation, session tests for arming/canceling/completing one correction, and presenter tests for confidence, alternatives, and experience evidence.

- [ ] **Step 2: Verify RED**

Run `../.tools/dotnet/dotnet test EchoFarm.sln --no-restore`; expect compilation/test failures for the new contracts and methods.

- [ ] **Step 3: Implement bridge and SMAPI wiring**

Keep execution on `ActionResponse.Action`; retain response metadata for the HUD. Add `CorrectAsync`, `EchoSessionState.Correcting`, a twenty-second correction deadline, last-action correlation, and conversion of the next successful `ObservedGameEvent` into `PlayerCorrection`. Bind `CorrectionKey` to `F10` and clear correction state on timeout, save change, title return, or cancellation.

- [ ] **Step 4: Verify GREEN**

Run `../.tools/dotnet/dotnet test EchoFarm.sln --no-restore` and expect all bridge tests to pass. Record that the Mod project itself still requires legal game assemblies.

### Task 9: Prove cross-session learning and package it

**Files:**
- Create: `demo/fixtures/day-5-reflective-farm.json`
- Create: `demo/fixtures/player-chest-correction.json`
- Create: `demo/run-reflective-demo.sh`
- Modify: `README.md`
- Modify: `echofarm-core/README.md`
- Modify: `导学-EchoFarm.md`
- Modify: `面经-EchoFarm.md`
- Modify: `release/nexus/BUILD-EVIDENCE.md`
- Modify: version metadata under `stardew-echo-mod/src/EchoFarm.Mod/manifest.json`

- [ ] **Step 1: Write the failing cross-process demo**

The script must start a real Go process in fixture mode and fail until it can assert failure-derived experience, proactive next-session deposit, nonempty alternatives, confidence, player-corrected chest selection, and memory-view evidence.

- [ ] **Step 2: Verify RED, then implement fixtures and fixture inference**

Run `./demo/run-reflective-demo.sh`; observe the first missing endpoint/field assertion, then add only the fixtures and fixture behavior required to satisfy each assertion.

- [ ] **Step 3: Run the complete verification matrix**

Run:

```bash
go test -race -count=1 ./...
go vet ./...
../.tools/dotnet/dotnet test EchoFarm.sln --no-restore
./demo/run-core-demo.sh
./demo/run-continuum-demo.sh
./demo/run-reflective-demo.sh
./scripts/tests/package-nexus-smoke.sh
```

- [ ] **Step 4: Update documentation and reproducible binaries**

Document the reflective loop as version `0.4.0`, rebuild deterministic linux-x64, windows-x64, macos-x64, and macos-arm64 sidecars under `artifacts/echofarm-core-0.4.0/`, verify macOS arm64 `/healthz`, and record exact SHA-256 values without claiming a playable Nexus build.

- [ ] **Step 5: Review and commit without pushing**

Run `git diff --check`, inspect the complete diff and repository status, commit coherent checkpoints on `codex/living-valley-director`, and leave the branch local.
