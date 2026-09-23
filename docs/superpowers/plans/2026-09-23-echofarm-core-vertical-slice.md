# EchoFarm Core Vertical Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a runnable Go + Eino core that turns one recorded morning-farm demonstration into an evidence-backed player model and reusable skill, then selects and replans safe high-level actions for a changed farm state.

**Architecture:** This first implementation slice deliberately stops at the local HTTP boundary. Domain contracts and validators protect the game from model output; Eino orchestrates the learning and acting graphs behind a small `Intelligence` interface; SQLite persists artifacts by save ID; HTTP endpoints make the slice immediately consumable by the later SMAPI bridge. A recorded fixture proves the complete `trace -> profile/skill -> action -> feedback/replan` path before any game rendering work begins.

**Tech Stack:** Go 1.24+, CloudWeGo Eino, Eino OpenAI-compatible model adapter, SQLite (`modernc.org/sqlite`), standard `net/http`, JSON Schema documents, Go testing.

---

## Scope Boundary

This plan implements the independently testable AI core. It does not pretend that the game bridge exists: Echo rendering, SMAPI event capture, pathfinding, and main-thread action execution belong to the next plan after this HTTP contract is stable. The server can run with either a real OpenAI-compatible model or a deterministic fixture model, so CI and the demo do not require a paid API.

## File Map

- `contracts/*.schema.json`: language-neutral wire contracts shared with the future C# bridge.
- `echofarm-core/internal/domain/*.go`: canonical Go types, enums, and trust-boundary validation.
- `echofarm-core/internal/trace/segmenter.go`: compresses low-level demonstration events into meaningful behavior segments.
- `echofarm-core/internal/intelligence/*.go`: model-neutral prompts/results and Eino graph assembly.
- `echofarm-core/internal/learning/service.go`: orchestrates segmentation, AI inference, validation, and persistence.
- `echofarm-core/internal/policy/service.go`: orchestrates next-action selection and recoverable-failure replanning.
- `echofarm-core/internal/memory/sqlite.go`: save-scoped persistence for profiles, skills, demonstrations, and sessions.
- `echofarm-core/internal/httpapi/*.go`: localhost API used by fixtures now and SMAPI later.
- `echofarm-core/cmd/echofarm/main.go`: configuration and process lifecycle.
- `demo/fixtures/*.json`: reproducible recorded teaching and changed-world scenarios.
- `demo/run-core-demo.sh`: one-command proof of the vertical slice.

### Task 1: Establish Domain and Wire Contracts

**Files:**
- Create: `echofarm-core/go.mod`
- Create: `echofarm-core/internal/domain/types.go`
- Create: `echofarm-core/internal/domain/validate.go`
- Create: `echofarm-core/internal/domain/validate_test.go`
- Create: `contracts/world-snapshot.schema.json`
- Create: `contracts/demonstration.schema.json`
- Create: `contracts/player-model.schema.json`
- Create: `contracts/skill-program.schema.json`
- Create: `contracts/high-level-action.schema.json`

- [ ] **Step 1: Write failing validation tests**

Cover valid snapshots, duplicate target IDs, out-of-range confidence, missing skill success conditions, and rejection of action names outside:

```go
var AllowedActionKinds = map[ActionKind]struct{}{
    ActionMoveTo: {}, ActionEquipTool: {}, ActionWaterTarget: {},
    ActionRefillCan: {}, ActionHarvestTarget: {}, ActionDepositItems: {},
    ActionStopSession: {},
}
```

Use a table test whose invalid cases require a non-nil error and whose valid case requires `Validate()` to return nil.

- [ ] **Step 2: Run the test and verify it fails**

Run: `cd echofarm-core && go test ./internal/domain`

Expected: FAIL because the domain types and validators do not exist.

- [ ] **Step 3: Implement minimal typed contracts and validators**

Define stable JSON field names for `Position`, `Crop`, `WaterSource`, `Chest`, `ToolState`, `InventorySummary`, `WorldSnapshot`, `StateDelta`, `DemonstrationEvent`, `Demonstration`, `ObservedPreference`, `PlayerModel`, `SkillStep`, `RecoveryStrategy`, `SkillProgram`, `HighLevelAction`, and `ActionResult`.

Validation must enforce:

```go
func (a HighLevelAction) Validate() error {
    if _, ok := AllowedActionKinds[a.Kind]; !ok {
        return fmt.Errorf("unsupported action kind %q", a.Kind)
    }
    if a.Kind != ActionStopSession && a.TargetID == "" && a.Destination == nil {
        return errors.New("action requires a target or destination")
    }
    return nil
}
```

Represent confidence as a JSON number constrained to `[0,1]`. Every entity addressable by an action must have a stable ID rather than relying on yesterday's tile coordinate.

- [ ] **Step 4: Add matching JSON Schemas**

Schemas must use draft 2020-12, `additionalProperties: false` at object boundaries, required identity/session fields, action `enum` values, and numeric confidence bounds. Keep schema property names identical to Go JSON tags.

- [ ] **Step 5: Run domain tests**

Run: `cd echofarm-core && go test ./internal/domain`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add contracts echofarm-core/go.mod echofarm-core/internal/domain
git commit -m "feat(echofarm): define trusted game contracts"
```

### Task 2: Segment Demonstrations Into Behaviors

**Files:**
- Create: `echofarm-core/internal/trace/segmenter.go`
- Create: `echofarm-core/internal/trace/segmenter_test.go`

- [ ] **Step 1: Write failing segmenter tests**

Build an event sequence containing repeated movement, watering three crops, refilling the can, harvesting, and depositing. Assert that:

```go
segments := Segment(events)
require.Equal(t, []domain.BehaviorKind{
    domain.BehaviorWatering,
    domain.BehaviorRefilling,
    domain.BehaviorHarvesting,
    domain.BehaviorDepositing,
}, kinds(segments))
require.Equal(t, []string{"crop-1", "crop-2", "crop-3"}, segments[0].TargetIDs)
```

Also assert that an unsuccessful tool action is retained as evidence and that empty demonstrations are rejected by the caller.

- [ ] **Step 2: Run the test and verify it fails**

Run: `cd echofarm-core && go test ./internal/trace`

Expected: FAIL because `Segment` does not exist.

- [ ] **Step 3: Implement deterministic event compression**

Use deterministic code only for noise reduction, never for intent inference:

```go
func Segment(events []domain.DemonstrationEvent) []domain.BehaviorSegment {
    // Ignore pure movement events, start a segment when the semantic action kind
    // changes, and preserve target order, state deltas, timing, and failures.
}
```

Adjacent semantic actions of the same kind belong to one segment. Movement paths are summarized as start/end positions on the following semantic segment so the model can infer route preferences without receiving frame-by-frame input.

- [ ] **Step 4: Run segmenter and full tests**

Run: `cd echofarm-core && go test ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/internal/trace
git commit -m "feat(echofarm): segment gameplay demonstrations"
```

### Task 3: Build the Eino Learning Graph

**Files:**
- Create: `echofarm-core/internal/intelligence/contracts.go`
- Create: `echofarm-core/internal/intelligence/prompts.go`
- Create: `echofarm-core/internal/intelligence/learning_graph.go`
- Create: `echofarm-core/internal/intelligence/learning_graph_test.go`

- [ ] **Step 1: Write a failing graph test with a fake structured model**

The fake returns a `LearningResult` containing a morning routine and two observed preferences. Assert that invoking the compiled graph receives segments plus the existing profile, validates the result, and preserves evidence IDs. Add cases for malformed JSON and an unsupported skill step.

The production boundary is:

```go
type Intelligence interface {
    Learn(ctx context.Context, input LearningInput) (LearningResult, error)
    ChooseAction(ctx context.Context, input ActionInput) (domain.HighLevelAction, error)
    Replan(ctx context.Context, input ReplanInput) (domain.HighLevelAction, error)
}

type StructuredGenerator interface {
    GenerateJSON(ctx context.Context, systemPrompt string, input any, output any) error
}
```

- [ ] **Step 2: Run the graph test and verify it fails**

Run: `cd echofarm-core && go test ./internal/intelligence -run Learning`

Expected: FAIL because the learning graph is missing.

- [ ] **Step 3: Implement the Eino graph**

Build a typed Eino graph with these nodes:

```text
LearningInput -> infer structured LearningResult -> validate evidence and actions -> LearningResult
```

The system prompt must say that the model infers goals, not coordinates; every preference must cite demonstration event IDs; only allow the morning-care action catalog; and output no prose outside the structured result. Compile the graph once at startup.

- [ ] **Step 4: Add the OpenAI-compatible generator adapter**

Configure the adapter only from environment variables:

```text
ECHOFARM_MODEL_BASE_URL
ECHOFARM_MODEL_API_KEY
ECHOFARM_MODEL_NAME
```

No real key or default remote endpoint may be committed. Make model timeout configurable and return typed `ErrModelUnavailable` and `ErrInvalidModelOutput` errors.

- [ ] **Step 5: Run intelligence tests**

Run: `cd echofarm-core && go test ./internal/intelligence`

Expected: PASS without network access.

- [ ] **Step 6: Commit**

```bash
git add echofarm-core/go.mod echofarm-core/go.sum echofarm-core/internal/intelligence
git commit -m "feat(echofarm): compile demonstrations with Eino"
```

### Task 4: Persist Evidence-Backed Player Memory

**Files:**
- Create: `echofarm-core/internal/memory/store.go`
- Create: `echofarm-core/internal/memory/sqlite.go`
- Create: `echofarm-core/internal/memory/sqlite_test.go`

- [ ] **Step 1: Write failing SQLite tests**

Open a temporary database, save a demonstration, player model, and skill for `save-a`, then prove they load after reopening. Save different data for `save-b` and prove no cross-save records appear. Assert a newer profile revision replaces only the same save's previous revision.

- [ ] **Step 2: Run the test and verify it fails**

Run: `cd echofarm-core && go test ./internal/memory`

Expected: FAIL because `OpenSQLite` and repository methods are absent.

- [ ] **Step 3: Implement migrations and transactions**

Create tables:

```sql
CREATE TABLE IF NOT EXISTS demonstrations (
  save_id TEXT NOT NULL, demonstration_id TEXT NOT NULL,
  payload_json BLOB NOT NULL, created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, demonstration_id)
);
CREATE TABLE IF NOT EXISTS player_models (
  save_id TEXT PRIMARY KEY, revision INTEGER NOT NULL,
  payload_json BLOB NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS skills (
  save_id TEXT NOT NULL, skill_name TEXT NOT NULL, revision INTEGER NOT NULL,
  payload_json BLOB NOT NULL, updated_at TEXT NOT NULL,
  PRIMARY KEY (save_id, skill_name)
);
```

Use `modernc.org/sqlite` to avoid CGO. Store validated JSON payloads and execute each learning-result write in one transaction.

- [ ] **Step 4: Run memory and race tests**

Run: `cd echofarm-core && go test -race ./internal/memory`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/go.mod echofarm-core/go.sum echofarm-core/internal/memory
git commit -m "feat(echofarm): persist save-scoped player memory"
```

### Task 5: Orchestrate Learning, Action Selection, and Replanning

**Files:**
- Create: `echofarm-core/internal/learning/service.go`
- Create: `echofarm-core/internal/learning/service_test.go`
- Create: `echofarm-core/internal/policy/service.go`
- Create: `echofarm-core/internal/policy/service_test.go`

- [ ] **Step 1: Write failing learning-service tests**

With fake memory and intelligence ports, assert that `Teach` validates the demonstration, segments it, loads the current profile, calls AI once, validates every model-produced artifact, and atomically persists the demonstration/profile/skill. Assert invalid AI output results in zero writes.

- [ ] **Step 2: Implement the learning service**

Expose:

```go
func (s *Service) Teach(ctx context.Context, demo domain.Demonstration) (domain.PlayerModel, domain.SkillProgram, error)
```

Do not add rule-based fallback inference. If AI is unavailable, return a typed error and preserve the last valid memory.

- [ ] **Step 3: Write failing policy tests for changed worlds**

Cover these states:

1. Rain plus mature crops selects `harvest_target`, never `water_target`.
2. A new dry crop ID absent from the teaching trace can be selected for watering.
3. Empty watering can selects `refill_can` before watering.
4. A recoverable `path_blocked` result calls `Replan` with the latest snapshot.
5. Low energy produces `stop_session` when the profile's reserve threshold would be violated.
6. Any invented action kind or stale target ID is rejected.

- [ ] **Step 4: Implement policy trust boundaries**

Expose:

```go
func (s *Service) NextAction(ctx context.Context, saveID string, snapshot domain.WorldSnapshot) (domain.HighLevelAction, error)
func (s *Service) HandleResult(ctx context.Context, saveID string, snapshot domain.WorldSnapshot, result domain.ActionResult) (domain.HighLevelAction, error)
```

AI chooses the high-level action. Deterministic code only validates that the action is in the catalog, its target exists in the latest snapshot, weather/tool preconditions hold, and the profile's safety threshold is respected.

- [ ] **Step 5: Run orchestration tests**

Run: `cd echofarm-core && go test ./internal/learning ./internal/policy`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add echofarm-core/internal/learning echofarm-core/internal/policy
git commit -m "feat(echofarm): learn and replan player-like routines"
```

### Task 6: Expose a Localhost API and Reproducible Demo

**Files:**
- Create: `echofarm-core/internal/httpapi/handler.go`
- Create: `echofarm-core/internal/httpapi/handler_test.go`
- Create: `echofarm-core/cmd/echofarm/main.go`
- Create: `demo/fixtures/morning-teaching.json`
- Create: `demo/fixtures/changed-rainy-farm.json`
- Create: `demo/fixtures/empty-can-farm.json`
- Create: `demo/run-core-demo.sh`
- Create: `echofarm-core/README.md`

- [ ] **Step 1: Write failing HTTP contract tests**

Test with `httptest.Server`:

```text
GET  /healthz
POST /v1/demonstrations/learn
POST /v1/echo/next-action
POST /v1/echo/action-result
GET  /v1/player-model?saveId=...
GET  /v1/skills/morning-farm-routine?saveId=...
```

Assert JSON content type, request-size limit, unknown-field rejection, save/session consistency, typed 400/422/503 responses, and no stack traces or model credentials in responses.

- [ ] **Step 2: Implement handlers and graceful shutdown**

Bind to `127.0.0.1:18471` by default. Use `http.Server` timeouts, a 2 MiB request limit, strict JSON decoding, structured error codes, SIGINT/SIGTERM shutdown, and close SQLite after the server stops.

- [ ] **Step 3: Add fixtures and demo runner**

The teaching fixture must include watering, a refill, harvest, and deposit evidence. The changed rainy fixture must contain a mature crop and moved/new crops. The empty-can fixture must make refilling a valid action.

`demo/run-core-demo.sh` starts the server in fixture-model mode, posts teaching data, queries the player model, requests actions for both changed states, prints JSON with `jq`, and always stops the server with a shell trap.

- [ ] **Step 4: Document two runtime modes**

Document:

```bash
# Offline deterministic smoke demo
./demo/run-core-demo.sh

# Real Eino-backed inference
export ECHOFARM_MODEL_BASE_URL=...
export ECHOFARM_MODEL_API_KEY=...
export ECHOFARM_MODEL_NAME=...
go run ./cmd/echofarm
```

State clearly that fixture mode verifies plumbing but the real model is required for learning novel behavior.

- [ ] **Step 5: Verify the full slice**

Run:

```bash
cd echofarm-core
go test -race ./...
go vet ./...
cd ..
./demo/run-core-demo.sh
```

Expected: tests and vet pass; the demo prints a persisted player model, `harvest_target` for the rainy changed farm, and `refill_can` for the empty-can farm.

- [ ] **Step 6: Commit**

```bash
git add echofarm-core demo
git commit -m "feat(echofarm): expose the AI core demo API"
```

### Task 7: Document the Game-Bridge Handoff

**Files:**
- Create: `docs/echofarm/smapi-bridge-contract.md`
- Modify: `README.md`

- [ ] **Step 1: Write the bridge contract**

For each endpoint, include request/response examples, correlation fields (`saveId`, `sessionId`, `snapshotVersion`), timeout behavior, retry safety, and the rule that action validation happens again on the game main thread. Document that a 503 or timeout stops Echo but never blocks game saving.

- [ ] **Step 2: Replace the root README entry point**

Lead with EchoFarm's product sentence and a short demo flow. Mark the existing Java/Vue folders as legacy experiments outside the new path; do not delete them in this slice. Do not copy any credentials from existing files into documentation.

- [ ] **Step 3: Check docs and repository hygiene**

Run:

```bash
rg -n "FIXME|CHANGEME|sk-[A-Za-z0-9_-]+|password:" docs/echofarm echofarm-core/README.md README.md
git status --short
```

Expected: no placeholders or credentials; only intended changes appear.

- [ ] **Step 4: Commit**

```bash
git add README.md docs/echofarm
git commit -m "docs(echofarm): define SMAPI bridge handoff"
```

## Completion Gate

Before calling this slice complete:

1. Every model-produced object passes domain validation before persistence or execution.
2. No test needs network access or a real model key.
3. The demo proves target-based generalization rather than coordinate replay.
4. SQLite data is isolated by `save_id`.
5. The HTTP server only listens on localhost by default.
6. Model failure disables Echo cleanly without inventing a rule-based substitute.
7. The next implementation plan starts from `docs/echofarm/smapi-bridge-contract.md` and delivers the in-game translucent Echo plus the watering/refill/harvest loop.
