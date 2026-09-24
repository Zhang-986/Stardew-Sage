# EchoFarm Model Operations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every EchoFarm model call attributable, durable, budget-bounded, and visible to the player without storing prompts, responses, API keys, or provider error bodies.

**Architecture:** Wrap the existing `StructuredGenerator` with a usage-aware decorator. It derives request purpose and save/session/day identity from the typed graph input, atomically reserves a call in SQLite before invoking the model, captures provider-reported Eino token metadata when available, and completes the ledger with sanitized status, error class, and latency. SQLite aggregates current-session and current-day usage for a read endpoint and the existing F9 memory view; the policy converts budget exhaustion into a persisted `stop_session` decision while learning/correction endpoints return a typed fail-closed error.

**Tech Stack:** Go 1.24.1, CloudWeGo Eino, modernc SQLite, C#/.NET 6 bridge contracts, xUnit, SMAPI status overlay.

---

### Task 1: Capture provider usage without storing model content

**Files:**
- Modify: `echofarm-core/internal/intelligence/contracts.go`
- Modify: `echofarm-core/internal/intelligence/generator.go`
- Modify: `echofarm-core/internal/intelligence/generator_test.go`

- [x] **Step 1: Write failing generator metadata tests**

Add tests in `generator_test.go` proving that a response with `schema.ResponseMeta.Usage` returns prompt, completion, and total tokens through a new optional `UsageReportingGenerator` interface, while a response without metadata returns `Reported=false`. Verify provider errors remain classifiable through `ErrModelUnavailable` without exposing response content.

- [x] **Step 2: Verify RED**

Run:

```bash
cd echofarm-core
go test ./internal/intelligence -run 'TestChatGeneratorReports|TestChatGeneratorLeaves' -count=1
```

Expected: compile failure because `GenerationUsage` and `GenerateJSONWithUsage` do not exist.

- [x] **Step 3: Implement minimal usage extraction**

Define:

```go
type GenerationUsage struct {
    Reported         bool
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}

type UsageReportingGenerator interface {
    GenerateJSONWithUsage(context.Context, string, any, any) (GenerationUsage, error)
}
```

Refactor `ChatGenerator.GenerateJSON` to delegate to `GenerateJSONWithUsage`. Read only `response.ResponseMeta.Usage`; do not retain request/response bodies in the returned metadata.

- [x] **Step 4: Verify and commit**

Run the targeted tests and `go test ./internal/intelligence -count=1`, then commit:

```bash
git commit -m "feat(ai): expose provider-reported model usage"
```

### Task 2: Add the durable call ledger and atomic budgets

**Files:**
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/domain/validate.go`
- Modify: `echofarm-core/internal/memory/store.go`
- Modify: `echofarm-core/internal/memory/sqlite.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`
- Create: `echofarm-core/internal/intelligence/tracked_generator.go`
- Create: `echofarm-core/internal/intelligence/tracked_generator_test.go`

- [x] **Step 1: Write failing SQLite ledger tests**

Cover one atomic reservation per request ID, exact call-budget enforcement under concurrency, completion with reported tokens, unknown-token persistence, session/day aggregation, and restart persistence. Records contain only request ID, purpose, save/session/day, timestamps, status, optional token counts, latency, and a bounded error class.

- [x] **Step 2: Verify RED**

Run:

```bash
cd echofarm-core
go test ./internal/memory -run 'TestSQLiteModelUsage|TestSQLiteConcurrentModelBudget' -count=1
```

Expected: compile failure because the model usage store methods do not exist.

- [x] **Step 3: Implement schema and store methods**

Add a `model_usage` table keyed by `request_id`. `ReserveModelCall` uses one transaction to count already-started/completed calls and sum only provider-reported tokens for the same save/session before insertion. `CompleteModelCall` changes exactly one `started` row to `succeeded` or `failed`. `GetModelUsageSummary` returns both current-session and same-day totals; unknown token metadata remains explicitly unknown.

Use these contracts:

```go
type ModelCallRecord struct {
    RequestID, SaveID, SessionID string
    Day int
    Purpose ModelCallPurpose
    StartedAt, FinishedAt time.Time
    Status ModelCallStatus
    PromptTokens, CompletionTokens, TotalTokens *int
    LatencyMS int64
    ErrorClass string
    CallBudget, TokenBudget int
}

type ModelUsageStore interface {
    ReserveModelCall(context.Context, ModelCallRecord) error
    CompleteModelCall(context.Context, ModelCallRecord) error
    GetModelUsageSummary(context.Context, string, string, int) (ModelUsageSummary, error)
}
```

- [x] **Step 4: Write failing tracked-generator tests**

Test purpose inference for `learning`, `intent`, `action`, `recovery`, and `reflection`; success/failure recording; redacted error classes; no delegate invocation after a call or reported-token limit is exhausted; and a typed `ErrModelBudgetExceeded`.

- [x] **Step 5: Implement the decorator**

Create `TrackingGenerator` around any `StructuredGenerator`. Generate a cryptographically random local request ID, reserve before the delegate call, use `UsageReportingGenerator` when available, measure elapsed time from an injectable clock, map errors to the closed set `model_unavailable`, `invalid_model_output`, `deadline_exceeded`, `cancelled`, or `internal`, and complete the ledger even when generation fails.

The constructor and sentinel are:

```go
var ErrModelBudgetExceeded = errors.New("model budget exhausted")

func NewTrackingGenerator(
    delegate StructuredGenerator,
    store ModelUsageStore,
    limits ModelBudgetLimits,
) (*TrackingGenerator, error)
```

- [x] **Step 6: Verify and commit**

Run:

```bash
go test -race ./internal/memory ./internal/intelligence -count=1
```

Commit:

```bash
git commit -m "feat(ai): persist model usage and enforce budgets"
```

### Task 3: Fail closed at the policy and HTTP boundaries

**Files:**
- Modify: `echofarm-core/internal/policy/service.go`
- Modify: `echofarm-core/internal/policy/service_test.go`
- Modify: `echofarm-core/internal/experience/service.go`
- Modify: `echofarm-core/internal/httpapi/handler.go`
- Modify: `echofarm-core/internal/httpapi/handler_test.go`
- Modify: `echofarm-core/cmd/echofarm/main.go`
- Create: `echofarm-core/cmd/echofarm/main_test.go`

- [x] **Step 1: Write failing budget-boundary tests**

Prove that exhausted action or intent budgets produce and persist a correlated `stop_session` action whose reason contains `model budget exhausted`; learning and correction map to HTTP `429` with code `model_budget_exhausted`; reflection jobs return to pending with the stable `budget_exhausted` failure class; and invalid/non-positive budget environment values are rejected.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/policy ./internal/httpapi ./cmd/echofarm -run 'Test.*Budget' -count=1
```

Expected: failures because budget errors are not handled or configured.

- [x] **Step 3: Wire the tracked generator**

Add positive integer settings `ECHOFARM_MAX_MODEL_CALLS_PER_SESSION` (default `32`) and `ECHOFARM_MAX_REPORTED_TOKENS_PER_SESSION` (default `100000`). Wrap fixture and OpenAI generators once, before graph construction. Existing idempotency checks stay before generator invocation, so duplicate demonstrations, decisions, corrections, and results consume no new reservation.

- [x] **Step 4: Implement fail-closed mappings**

Handle `ErrModelBudgetExceeded` in policy preparation/proposal paths by persisting a `stop_session` decision. Map it to HTTP 429 for operations that cannot return an action. Add `budget_exhausted` to the reflection job's bounded failure codes without persisting arbitrary model messages.

The HTTP error body remains the existing strict shape:

```json
{"code":"model_budget_exhausted","message":"Echo stopped because the configured model budget was exhausted"}
```

- [x] **Step 5: Verify and commit**

Run full Go race/vet and commit:

```bash
git commit -m "feat(ai): stop safely when model budgets are exhausted"
```

### Task 4: Expose usage through the API and F9

**Files:**
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/memoryview/service.go`
- Modify: `echofarm-core/internal/memoryview/service_test.go`
- Modify: `echofarm-core/internal/httpapi/handler.go`
- Modify: `echofarm-core/internal/httpapi/handler_test.go`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/PlayerMemory.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Contracts/JsonContractTests.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoMemoryPresenter.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoMemoryPresenterTests.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`

- [x] **Step 1: Write failing API/view/contract tests**

Add `GET /v1/model-usage?saveId=...&sessionId=...` tests plus memory-view projection tests. C# round-trip tests must preserve known-vs-unknown token state, and presenter tests must show session calls/limit, tokens/limit or `unknown`, failure count, and recent latency without displaying costs.

- [x] **Step 2: Verify RED**

Run targeted Go and .NET tests. Expect missing usage contracts and endpoint methods.

- [x] **Step 3: Implement API and memory projection**

Expose the SQLite summary directly and include the latest summary in `EchoMemoryView`. Keep the endpoint loopback-only through the existing server and reject missing save IDs. Do not return prompts, responses, provider bodies, or API-key fields.

Use a bounded aggregate rather than raw call rows:

```go
type ModelUsageSummary struct {
    SaveID, SessionID string
    Day int
    Session ModelUsageTotals
    DayTotals ModelUsageTotals
    CallBudget, TokenBudget int
    BudgetExhausted bool
}
```

- [x] **Step 4: Wire Mod settings and F9 rendering**

Add `MaxModelCallsPerSession = 32` and `MaxReportedTokensPerSession = 100000` to `ModConfig`, pass them as child-process environment values through `CoreLaunchOptionsFactory`, and render concise AI usage lines in F9. `0`, negative, and malformed limits must fail before launching the child.

- [x] **Step 5: Verify and commit**

Run all Go/.NET/PowerShell tests and the three demos, then commit:

```bash
git commit -m "feat(echofarm): expose bounded AI usage in F9"
```

### Task 5: Release verification

**Files:**
- Modify: `README.md`
- Modify: `release/nexus/PLAYER_README.md`
- Modify: `release/nexus/BUILD-EVIDENCE.md`

- [ ] **Step 1: Document budgets and privacy boundary**

Document defaults, environment/config override mapping, typed exhaustion behavior, unknown provider token metadata, and the explicit decision not to calculate currency cost.

- [ ] **Step 2: Run complete verification**

Run Go race/vet, the .NET Release suite, PowerShell setup tests, all three demos, package smoke, and the two concurrency tests twenty times. Confirm tracked source contains no provider-token-shaped value.

- [ ] **Step 3: Commit, push, and verify exact-SHA CI**

Commit:

```bash
git commit -m "docs(echofarm): document model operations controls"
```

Push `codex/living-valley-director` and require both GitHub Actions jobs to pass on the final SHA.
