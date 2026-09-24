# EchoFarm Continuum Consistency Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make learning, action selection, result progression, trait inference, and memory polling deterministic under retries and concurrency.

**Architecture:** SQLite remains the source of truth. A service may generate a candidate, but after an idempotent write it must re-read and return the canonical stored record. Result feedback must always advance to a newer world snapshot, model observations must be unambiguous per key/context slot, and the SMAPI adapter must use separate command and read client profiles.

**Tech Stack:** Go 1.24, modernc SQLite, Eino-backed intelligence interfaces, C#/.NET 6 bridge, xUnit, SMAPI adapter.

---

### Task 1: Canonical learning outcome

**Files:**
- Modify: `echofarm-core/internal/learning/service_test.go`
- Modify: `echofarm-core/internal/learning/service.go`

- [x] **Step 1: Write the failing concurrency-winner test**

Add `TestTeachOutcomeReturnsCanonicalPersistedWinner`. Configure the store stub so the initial read returns `memory.ErrNotFound`, `SaveLearningOutcome` simulates another request having stored a different valid outcome, and the post-save read returns that winner. Assert the service returns the winner rather than the local inference.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/learning -run TestTeachOutcomeReturnsCanonicalPersistedWinner -count=1
```

Expected: FAIL because `TeachOutcome` returns the locally generated outcome.

- [x] **Step 3: Return the committed record**

After `SaveLearningOutcome`, call:

```go
stored, err := s.store.GetLearningOutcome(ctx, demonstration.SaveID, demonstration.ID)
if err != nil {
    return domain.LearningOutcome{}, fmt.Errorf("reload persisted learning result: %w", err)
}
return stored, nil
```

Make the store stub persist either the supplied test winner or the locally supplied outcome so existing tests model the real SQLite contract.

- [x] **Step 4: Verify GREEN**

Run `go test ./internal/learning -count=1` and expect PASS.

### Task 2: Canonical action decision

**Files:**
- Modify: `echofarm-core/internal/policy/service_test.go`
- Modify: `echofarm-core/internal/policy/service.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`
- Modify: `echofarm-core/internal/memory/sqlite.go`

- [x] **Step 1: Write the failing decision-winner test**

Add `TestNextActionReturnsCanonicalPersistedWinner`. Let the actor produce a water action while `SaveDecision` simulates a concurrent writer retaining a harvest action for the same `(save, session, snapshotVersion)`. Assert the returned action is the persisted harvest action.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/policy -run TestNextActionReturnsCanonicalPersistedWinner -count=1
```

Expected: FAIL because `NextAction` returns its locally generated final action.

- [x] **Step 3: Reload after the idempotent save**

Change `saveDecision` to return `(domain.HighLevelAction, error)`. After `SaveDecision`, load the same ledger key with `GetDecision` and return `stored.FinalAction`. Route both normal and stop-session paths in `NextAction` and `HandleResult` through that returned canonical value. In the SQLite transaction, derive `echo_sessions.status` from the canonical decision rather than a conflicting local candidate.

- [x] **Step 4: Verify GREEN**

Run `go test ./internal/policy ./internal/memory -count=1` and expect PASS, including `TestSQLiteDecisionConflictKeepsSessionStatusFromCanonicalWinner`.

### Task 3: Strict result snapshot progression

**Files:**
- Modify: `echofarm-core/internal/policy/service_test.go`
- Modify: `echofarm-core/internal/policy/service.go`

- [x] **Step 1: Write the failing progression test**

Add `TestHandleResultRequiresNewerSnapshotThanExecutedAction` with a valid result and a current snapshot at the same version. Assert an error containing `newer` is returned and the result is not attached.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/policy -run TestHandleResultRequiresNewerSnapshotThanExecutedAction -count=1
```

Expected: FAIL because equal versions are currently accepted.

- [x] **Step 3: Enforce a strictly newer snapshot**

Replace the permissive comparison with:

```go
if snapshot.SnapshotVersion <= result.SnapshotVersion {
    return domain.HighLevelAction{}, errors.New("current snapshot must be newer than action result")
}
```

Update result-handling tests so `result.Action` belongs to the executed snapshot and replanned actions belong to the incremented current snapshot.

- [x] **Step 4: Verify GREEN**

Run `go test ./internal/policy -count=1` and expect PASS.

### Task 4: Reject ambiguous trait observations

**Files:**
- Modify: `echofarm-core/internal/modeling/merger_test.go`
- Modify: `echofarm-core/internal/modeling/merger.go`

- [x] **Step 1: Write the failing ambiguity test**

Add `TestMergeRejectsMultipleValuesForSameTraitContext` with two valid observations that share `PreferenceTaskOrder` and `TraitContextSunny` but use different values. Assert `Merge` rejects them.

- [x] **Step 2: Verify RED**

Run:

```bash
go test ./internal/modeling -run TestMergeRejectsMultipleValuesForSameTraitContext -count=1
```

Expected: FAIL because the current duplicate identity includes the value.

- [x] **Step 3: Key inference uniqueness by trait slot**

Track seen observations by `key + context`, preserving validation of every observation. Reject a second value for an occupied slot before mutating the cloned model, so inference order cannot affect confidence decay or projection.

- [x] **Step 4: Verify GREEN**

Run `go test ./internal/modeling -count=1` and expect PASS.

### Task 5: Separate command and read timeout profiles

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmClientFactory.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Transport/EchoFarmClientFactoryTests.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmClient.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`

- [x] **Step 1: Write the failing factory-boundary test**

Add `CreateUsesIndependentCommandAndReadProfiles`, asserting the bridge factory creates distinct clients for the same loopback endpoint, with a 35-second command timeout and a 2-second read timeout.

- [x] **Step 2: Verify RED**

Run:

```bash
../.tools/dotnet/dotnet test EchoFarm.sln --no-restore --filter EchoFarmClientFactoryTests
```

Expected: FAIL to compile because the factory boundary does not exist.

- [x] **Step 3: Implement and wire the profiles**

Expose the immutable configured timeout on `EchoFarmClient`. Add an `EchoFarmClientFactory` returning a pair named `Commands` and `Reads`, backed by separate `HttpClient` instances. Keep command timeout at 35 seconds and read timeout at 2 seconds. In `ModEntry`, pass `Commands` into `EchoSession`, and use `Reads` only for restore and F9 memory polling.

- [x] **Step 4: Verify GREEN**

Run `../.tools/dotnet/dotnet test EchoFarm.sln --no-restore` and expect all bridge tests to pass.

### Task 6: Full verification and reproducible artifacts

**Files:**
- Modify: `release/nexus/BUILD-EVIDENCE.md`
- Rebuild ignored outputs under: `artifacts/echofarm-core-0.3.0/`

- [x] **Step 1: Verify all source paths**

Run:

```bash
go test -race -count=1 ./...
go vet ./...
../.tools/dotnet/dotnet test EchoFarm.sln --no-restore
./demo/run-core-demo.sh
./demo/run-continuum-demo.sh
./scripts/tests/package-nexus-smoke.sh
```

- [x] **Step 2: Rebuild four deterministic sidecars**

Cross-compile `./cmd/echofarm` with `CGO_ENABLED=0`, `-trimpath`, and `-ldflags='-s -w -buildid='` for linux/amd64, darwin/amd64, darwin/arm64, and windows/amd64 into `artifacts/echofarm-core-0.3.0/`.

- [x] **Step 3: Native smoke and evidence**

Start the macOS arm64 sidecar in fixture mode against a temporary SQLite database, require HTTP 200 from `/healthz`, stop it, calculate all four SHA-256 hashes, and replace the counts/hashes in `release/nexus/BUILD-EVIDENCE.md`.

- [x] **Step 4: Review and commit without pushing**

Inspect `git diff --check`, `git diff`, and `git status`. Commit the coherent hardening changes on `codex/living-valley-director`; do not push or claim a playable Nexus build without the legal game DLLs.
