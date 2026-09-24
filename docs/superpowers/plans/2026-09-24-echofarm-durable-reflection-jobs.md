# EchoFarm Durable Reflection Jobs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist failed-action reflection work atomically, lease it to one processor at a time, and retry transient failures after later requests or process restarts without duplicating durable experience.

**Architecture:** SQLite remains the only source of truth. `AttachDecisionResult` stores the canonical failed result and one pending job in a transaction; the experience service leases one job, runs the existing idempotent reflection path, then completes or releases the lease. The policy opportunistically processes one job before preparing a decision and never lets reflection failure block safe gameplay.

**Tech Stack:** Go 1.24, CloudWeGo Eino, modernc SQLite, `database/sql`, Go race detector, C#/.NET bridge regression tests, Bash/JQ cross-process demos.

---

### Task 1: Define the durable job boundary

**Files:**
- Create: `echofarm-core/internal/memory/reflection_job.go`
- Modify: `echofarm-core/internal/memory/store.go:25-40`
- Test: `echofarm-core/internal/memory/sqlite_test.go`

- [x] **Step 1: Write the failing compile-time and validation tests**

Add tests that construct a valid `memory.ReflectionJobLease`, reject an empty save ID or non-positive lease duration, and require `DecisionStore.AttachDecisionResult` to receive the newer `WorldSnapshot`.

The desired public storage boundary is:

```go
type ReflectionJob struct {
	SaveID       string
	SourceID     string
	Snapshot     domain.WorldSnapshot
	Result       domain.ActionResult
	AttemptCount int
}

type ReflectionJobLease struct {
	Job   ReflectionJob
	Token string
}

type ReflectionJobStore interface {
	ClaimReflectionJob(context.Context, string, time.Duration) (ReflectionJobLease, bool, error)
	CompleteReflectionJob(context.Context, ReflectionJobLease) error
	ReleaseReflectionJob(context.Context, ReflectionJobLease, string) error
}
```

- [x] **Step 2: Run the focused test and verify RED**

Run:

```bash
cd echofarm-core
go test ./internal/memory -run 'TestSQLite(RejectsInvalidReflectionClaim|Attaches)' -count=1
```

Expected: compilation fails because the job types/methods and new attachment signature do not exist.

- [x] **Step 3: Add the minimal types and interfaces**

Create the two job structs above, add `ErrReflectionLeaseLost`, validate save ID/token/duration at the store boundary, and change the decision method to:

```go
AttachDecisionResult(context.Context, domain.ActionResult, domain.WorldSnapshot) (bool, error)
```

- [x] **Step 4: Run the focused package test**

Run `go test ./internal/memory -count=1`; expected PASS after callers are updated only enough to compile.

- [x] **Step 5: Commit**

```bash
git add echofarm-core/internal/memory/reflection_job.go echofarm-core/internal/memory/store.go echofarm-core/internal/memory/sqlite_test.go
git commit -m "feat(echofarm): define durable reflection jobs"
```

### Task 2: Attach results and enqueue reflection atomically

**Files:**
- Modify: `echofarm-core/internal/memory/sqlite.go:15-101,281-342`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`
- Modify: `echofarm-core/internal/policy/service.go:112-128`
- Modify: `echofarm-core/internal/policy/service_test.go:97-110`

- [x] **Step 1: Write failing SQLite transaction tests**

Add tests with a real temporary database that assert:

```go
first, err := store.AttachDecisionResult(ctx, failed, currentSnapshot)
// first == true, err == nil, one pending job exists

first, err = store.AttachDecisionResult(ctx, failed, currentSnapshot)
// first == false, err == nil, still one job

_, err = store.AttachDecisionResult(ctx, conflicting, currentSnapshot)
// err contains "different content", canonical result and job are unchanged
```

Retain the 16-writer race test and assert that exactly one canonical result wins and no more than one failed-result job exists.

- [x] **Step 2: Verify RED**

Run:

```bash
cd echofarm-core
go test -race ./internal/memory -run 'TestSQLite(Attaches|Enqueues|RejectsResult|AttachesOnly)' -count=1
```

Expected: failures because `reflection_jobs` is absent and attachment is not transactional with enqueue.

- [x] **Step 3: Implement the schema and transaction**

Add the `reflection_jobs` table from the design. Marshal this private payload:

```go
type reflectionJobPayload struct {
	Snapshot domain.WorldSnapshot `json:"snapshot"`
	Result   domain.ActionResult   `json:"result"`
}
```

Validate snapshot/result identity and strict version progression. In one `BeginTx`, read the decision, compare actions, update `decision_records` with `WHERE payload_json=?`, insert the pending job for a failed winner using `ON CONFLICT DO NOTHING`, and commit. Reload conflict state through the same transaction so `SetMaxOpenConns(1)` cannot deadlock.

- [x] **Step 4: Update policy and test stubs for the new argument**

Pass the already validated current snapshot into attachment:

```go
attached, err := s.store.AttachDecisionResult(ctx, result, snapshot)
```

The stub records both values and preserves first-write-wins behavior.

- [x] **Step 5: Verify GREEN and commit**

Run:

```bash
cd echofarm-core
go test -race ./internal/memory ./internal/policy -count=1
```

Then commit:

```bash
git add echofarm-core/internal/memory/sqlite.go echofarm-core/internal/memory/sqlite_test.go echofarm-core/internal/policy/service.go echofarm-core/internal/policy/service_test.go
git commit -m "feat(echofarm): enqueue reflection with failed results"
```

### Task 3: Implement lease lifecycle and crash recovery

**Files:**
- Modify: `echofarm-core/internal/memory/reflection_job.go`
- Modify: `echofarm-core/internal/memory/sqlite.go`
- Modify: `echofarm-core/internal/memory/sqlite_test.go`

- [x] **Step 1: Write failing lease tests**

Use a controllable `store.now` clock and assert:

- the first claim returns a non-empty opaque token;
- a second concurrent claim returns `ok=false` while the lease is live;
- `ReleaseReflectionJob` changes the job back to pending and increments `AttemptCount`;
- an expired processing lease can be reclaimed with a different token;
- only the owning token can complete or release a job;
- completed jobs cannot be reclaimed;
- a pending job survives closing and reopening SQLite.

- [x] **Step 2: Verify RED**

Run `go test -race ./internal/memory -run 'TestSQLiteReflectionJob' -count=1`; expected FAIL because lease methods are not implemented.

- [x] **Step 3: Implement claim, complete, and release**

Add `now func() time.Time` to `SQLite`, default it to `time.Now`, and generate lease tokens from 16 bytes of `crypto/rand` encoded with hex. Claim the oldest eligible row and conditionally update it to `processing` with a UTC expiry. Complete/release with `WHERE lease_token=? AND status='processing'`; return `ErrReflectionLeaseLost` when `RowsAffected` is zero.

Permit only these stored failure codes:

```go
const (
	ReflectionFailureModelUnavailable = "model_unavailable"
	ReflectionFailureCanceled         = "canceled"
	ReflectionFailureInternal         = "reflection_failed"
)
```

No provider error text, prompt, or credential is persisted.

- [x] **Step 4: Verify GREEN, stress the lease race, and commit**

Run:

```bash
cd echofarm-core
go test -race ./internal/memory -run 'TestSQLiteReflectionJob' -count=20
```

Then commit the memory files with `git commit -m "feat(echofarm): lease persistent reflection work"`.

### Task 4: Process pending reflection idempotently

**Files:**
- Modify: `echofarm-core/internal/experience/service.go`
- Modify: `echofarm-core/internal/experience/service_test.go`

- [x] **Step 1: Write failing worker tests**

Extend the real behavior tests to prove:

```go
processed, err := service.ProcessPending(ctx, "farm-1")
// first model failure: processed == true, error wraps ErrModelUnavailable,
// lease released and attempt count becomes 1

processed, err = service.ProcessPending(ctx, "farm-1")
// second attempt succeeds, experience is durable, job is completed

processed, err = service.ProcessPending(ctx, "farm-1")
// processed == false, reflector call count remains 2
```

Also test that a pre-existing canonical experience lets a reclaimed job complete without another model call.

- [x] **Step 2: Verify RED**

Run `go test ./internal/experience -run 'TestProcessPending' -count=1`; expected compilation failure because `ProcessPending` does not exist.

- [x] **Step 3: Implement one-job processing**

Use a 45-second lease. After claiming, call the existing `learn` method with the job's persisted source ID, snapshot, and result. On success call `CompleteReflectionJob`; on failure classify the error as `model_unavailable`, `canceled`, or `reflection_failed`, release the lease, and return the original error joined with any release error.

- [x] **Step 4: Verify GREEN and commit**

Run `go test -race ./internal/experience -count=1`, then commit with `git commit -m "feat(echofarm): retry durable reflection work"`.

### Task 5: Recover pending work through policy requests

**Files:**
- Modify: `echofarm-core/internal/policy/service.go`
- Modify: `echofarm-core/internal/policy/service_test.go`
- Test: `echofarm-core/cmd/echofarm/main_test.go`

- [x] **Step 1: Write failing policy tests**

Replace the direct learner stub with:

```go
type reflectionProcessorStub struct {
	calls int
	err   error
}

func (s *reflectionProcessorStub) ProcessPending(context.Context, string) (bool, error) {
	s.calls++
	return true, s.err
}
```

Assert that `HandleResultDecision` passes the newer snapshot to atomic attachment, invokes the processor before policy preparation, and continues replanning when processing fails. Assert that `NextDecision` invokes the processor before loading experiences, which recovers work after restart. A request processes at most one job.

- [x] **Step 2: Verify RED**

Run `go test ./internal/policy -run 'Test(HandleResult|NextDecision).*Reflection' -count=1`; expected FAIL against direct `LearnFromResult` orchestration.

- [x] **Step 3: Implement opportunistic recovery**

Change the policy dependency to:

```go
type reflectionProcessor interface {
	ProcessPending(context.Context, string) (bool, error)
}
```

Call it once near the start of `NextDecision` and once after action-result attachment. Ignore its error only at this gameplay boundary; the worker has already returned the lease to durable pending state. Do not loop.

- [x] **Step 4: Verify GREEN and commit**

Run:

```bash
cd echofarm-core
go test -race ./internal/policy ./cmd/echofarm -count=1
```

Then commit with `git commit -m "feat(echofarm): recover reflection jobs during play"`.

### Task 6: Document and verify the production path

**Files:**
- Modify: `echofarm-core/README.md`
- Modify: `docs/echofarm/smapi-bridge-contract.md`
- Modify: `release/nexus/BUILD-EVIDENCE.md`
- Modify: `release/nexus/changelog.md`
- Rebuild ignored outputs: `artifacts/echofarm-core-0.4.0/`

- [x] **Step 1: Document precise delivery semantics**

State that failed results enqueue durable reflection, leases prevent simultaneous processing, transient failures retry on later policy requests, persisted experience is idempotent, and model inference itself is at-least-once across crash boundaries.

- [x] **Step 2: Run complete verification**

Run:

```bash
cd echofarm-core
go test -race -count=1 ./...
go vet ./...
cd ../stardew-echo-mod
../.tools/dotnet/dotnet test EchoFarm.sln --no-restore
cd ..
./demo/run-core-demo.sh
./demo/run-continuum-demo.sh
./demo/run-reflective-demo.sh
./scripts/tests/package-nexus-smoke.sh
```

Expected: every command exits 0 and .NET reports 87 tests unless new bridge-only tests intentionally raise the count.

- [x] **Step 3: Rebuild and verify artifacts**

Cross-build linux/amd64, darwin/amd64, darwin/arm64, and windows/amd64 with `CGO_ENABLED=0`, `-trimpath`, and `-ldflags='-s -w -buildid='`. Rebuild a second time into a temporary directory and require `cmp` equality. Recompute `SHA256SUMS.txt`, update `BUILD-EVIDENCE.md`, and require the macOS arm64 binary to return HTTP 200 from `/healthz` in fixture mode.

- [x] **Step 4: Review and commit locally**

Run `git diff --check`, inspect `git diff`, and stage only the files listed above. Preserve `.superpowers/`, do not push, and commit with:

```bash
git commit -m "feat(echofarm): make AI reflection crash-recoverable"
```
