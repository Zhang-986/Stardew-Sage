# EchoFarm Experience Effectiveness Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn canonical game action results into idempotent, auditable effectiveness feedback that changes later experience ranking without model-driven scoring.

**Architecture:** The result transaction appends one feedback row per materially applied experience. Base semantic confidence stays in `policy_experiences`; reads aggregate the append-only ledger into success/failure/neutral counts and an order-independent effective confidence. The matcher ranks projected confidence and cools repeatedly contradicted experience.

**Tech Stack:** Go 1.24, modernc SQLite, CloudWeGo Eino structured inputs, C#/.NET strict JSON contracts, SMAPI F9 presenter, Bash/JQ cross-process demo.

---

### Task 1: Define attribution and effective confidence

**Files:**
- Create: `echofarm-core/internal/domain/experience_feedback.go`
- Create: `echofarm-core/internal/domain/experience_feedback_test.go`
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/domain/validate.go`

- [ ] **Step 1: Write failing attribution tests**

Cover successful results, water/out-of-water, harvest/inventory-full, deposit/chest-full, deposit/inventory-empty, path-blocked, target-changed, unsupported-crop, and unknown errors. Require unknown diagnostics to normalize to `other`.

- [ ] **Step 2: Write failing confidence and cooling tests**

Require:

```go
EffectiveExperienceConfidence(0.65, 0, 0) == 0.65
EffectiveExperienceConfidence(0.65, 1, 0) == 0.72
EffectiveExperienceConfidence(0.65, 0, 1) == 0.52
```

Verify order independence, invalid count rejection, and cooling only when `failureCount >= 3 && effectiveConfidence < 0.40`.

- [ ] **Step 3: Verify RED**

Run `go test ./internal/domain -run 'Test(Classify|Effective|ExperienceCooling)' -count=1`; expected compilation failure because the feedback policy does not exist.

- [ ] **Step 4: Implement minimal domain policy**

Add `ExperienceFeedbackOutcome` values `succeeded`, `contradicted`, and `neutral`. Extend `PolicyExperience` with optional `effectiveConfidence`, `successCount`, `failureCount`, and `neutralCount`. Implement the fixed-prior formula `(confidence*4+success)/(4+success+failure)`, classification catalog, normalized diagnostic, and cooling predicate. Extend validation for finite projected confidence and non-negative counts.

- [ ] **Step 5: Verify GREEN and commit**

Run `go test -race ./internal/domain -count=1` and commit with `git commit -m "feat(echofarm): define experience effectiveness policy"`.

### Task 2: Persist feedback atomically with action results

**Files:**
- Modify: `echofarm-core/internal/memory/sqlite.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`

- [ ] **Step 1: Write failing transaction tests**

Create a persisted decision whose selected primary proposal applies two real experience IDs. Attach a successful result and require two `succeeded` ledger rows. Repeat the result and require no additional rows. Attach a conflicting result and require the original rows unchanged.

Add cases proving alternatives/stops produce no feedback, a missing referenced experience rolls back result attachment, and sixteen concurrent writers leave one canonical result plus one row per experience.

- [ ] **Step 2: Verify RED**

Run `go test -race ./internal/memory -run 'TestSQLite.*ExperienceFeedback' -count=1`; expected FAIL because `experience_feedback` does not exist.

- [ ] **Step 3: Implement schema and atomic append**

Add the table from the design. In `AttachDecisionResult`, after the canonical CAS succeeds and before commit, check eligibility, verify every experience exists for the save, classify the result, normalize its error code, and insert rows keyed by save/session/snapshot/experience. Any validation or insert error rolls back the result and reflection job together.

- [ ] **Step 4: Verify GREEN under race**

Run `go test -race ./internal/memory -run 'TestSQLite.*(ExperienceFeedback|CanonicalResult)' -count=20`.

- [ ] **Step 5: Commit**

Commit with `git commit -m "feat(echofarm): record atomic experience feedback"`.

### Task 3: Project feedback and change matching

**Files:**
- Modify: `echofarm-core/internal/memory/sqlite.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`
- Modify: `echofarm-core/internal/experience/matcher.go`
- Modify: `echofarm-core/internal/experience/matcher_test.go`

- [ ] **Step 1: Write failing projection tests**

After successful, contradicted, and neutral feedback across decisions, close/reopen SQLite and require `ListPolicyExperiences` to return exact counts and effective confidence while the stored base confidence remains unchanged. Save a later reflection outcome and prove feedback counts survive rather than being overwritten.

- [ ] **Step 2: Write failing matcher tests**

Require effective confidence to outrank semantic confidence and filter an experience with three failures below `0.40`. Require legacy experiences with zero projected fields to rank by semantic confidence.

- [ ] **Step 3: Verify RED**

Run `go test ./internal/memory ./internal/experience -run 'Test.*(FeedbackProjection|EffectiveRanking|Cooled)' -count=1`; expected assertion failures against base-only reads and sorting.

- [ ] **Step 4: Implement projection and cooling**

Close the base experience rows before querying aggregated feedback because SQLite is configured with one connection. Reset projection-only fields on each decoded base object, aggregate by outcome, calculate effective confidence, and reject malformed outcome values. Strip projection fields before `SaveExperienceOutcome` marshals base JSON. Update matcher sorting and filtering through domain helpers.

- [ ] **Step 5: Verify GREEN and commit**

Run `go test -race ./internal/memory ./internal/experience ./internal/policy -count=1` and commit with `git commit -m "feat(echofarm): rank experience by observed outcomes"`.

### Task 4: Extend strict contracts and F9 explanation

**Files:**
- Create: `contracts/policy-experience.schema.json`
- Modify: `contracts/echo-memory.schema.json`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/PlayerMemory.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoMemoryPresenter.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Contracts/JsonContractTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/EchoMemoryPresenterTests.cs`

- [ ] **Step 1: Write failing strict JSON and presenter tests**

Deserialize a policy experience containing `effectiveConfidence`, `successCount`, `failureCount`, and `neutralCount`. Require F9 to render `有效 78% · 成功 3 / 失败 1`. Keep old JSON valid with zero/default fields.

- [ ] **Step 2: Verify RED**

Run `../.tools/dotnet/dotnet test EchoFarm.sln --no-restore --filter 'JsonContractTests|EchoMemoryPresenterTests'`; expected strict unknown-member failure or missing property compilation errors.

- [ ] **Step 3: Implement contracts, schema, and presentation**

Add the four C# properties, define a strict policy-experience schema with optional projected fields, reference it from `echo-memory.schema.json`, and append the concise effectiveness text only when projected confidence is positive.

- [ ] **Step 4: Verify GREEN and commit**

Run the full .NET suite and commit with `git commit -m "feat(echofarm): explain experience effectiveness in game"`.

### Task 5: Demonstrate feedback and rebuild release evidence

**Files:**
- Modify: `demo/run-reflective-demo.sh`
- Modify: `README.md`
- Modify: `echofarm-core/README.md`
- Modify: `docs/echofarm/smapi-bridge-contract.md`
- Modify: `release/nexus/BUILD-EVIDENCE.md`
- Modify: `release/nexus/changelog.md`
- Modify: `release/nexus/description.md`
- Rebuild ignored: `artifacts/echofarm-core-0.4.0/`

- [ ] **Step 1: Extend the cross-process demo**

After the corrected day-six deposit decision, report that exact action as succeeded against snapshot version two. Query memory and require the correction experience to have `successCount == 1` and `effectiveConfidence > confidence`. Print the projected experience as demo evidence.

- [ ] **Step 2: Document the feedback boundary**

Explain outcome attribution, neutral failures, append-only idempotency, effective confidence, and cooling without claiming reward-model or reinforcement-learning infrastructure.

- [ ] **Step 3: Run complete verification**

Run Go race/vet, the full .NET suite, all three demos, and Nexus packaging smoke. Stress the result/feedback concurrency tests for twenty repetitions.

- [ ] **Step 4: Rebuild deterministic sidecars and smoke native health**

Rebuild linux/amd64, darwin/amd64, darwin/arm64, and windows/amd64 with the existing reproducible flags. Require byte-for-byte equality on a second build, update checksums and build evidence, then require macOS arm64 `/healthz` to return HTTP 200.

- [ ] **Step 5: Review and commit locally**

Run `git diff --check`, inspect the diff and status, preserve `.superpowers/`, do not push, and commit with `git commit -m "feat(echofarm): close the experience outcome loop"`.
