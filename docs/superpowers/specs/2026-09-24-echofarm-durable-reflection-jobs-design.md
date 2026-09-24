# EchoFarm Durable Reflection Jobs Design

## Goal

Make failure-derived AI reflection durable without adding an external queue: a failed action and its pending reflection job commit atomically, only one worker may actively process the job, transient model failures remain retryable, and a restart cannot silently lose unfinished learning.

## Scope

This change covers automatic reflection from failed Echo actions. F10 player corrections keep their existing retryable game-side state machine. It does not add RAG, background cloud services, arbitrary tools, or a general-purpose job framework.

## Guarantees

- The canonical action result remains first-write-wins for `(saveId, sessionId, snapshotVersion)`.
- The first canonical failed result and one reflection job are written in the same SQLite transaction.
- Identical result retries observe the same canonical result and never create another job.
- Conflicting result retries cannot replace the canonical result or its job.
- A job has at most one live lease. An expired lease can be reclaimed after a process crash.
- A failed reflection attempt returns the job to `pending` and records a bounded diagnostic; normal gameplay remains fail-safe.
- A completed experience outcome is idempotent by `sourceId`, so retrying work cannot apply confidence twice.
- The external model call is at-least-once around a crash boundary: if the process dies after the model responds but before completion is committed, the call may be repeated. EchoFarm promises deterministic, once-only persistence, not impossible distributed exactly-once inference.

## Data model

Add a forward-only `reflection_jobs` table:

```sql
CREATE TABLE IF NOT EXISTS reflection_jobs (
  save_id TEXT NOT NULL,
  source_id TEXT NOT NULL,
  payload_json BLOB NOT NULL,
  status TEXT NOT NULL,
  lease_token TEXT,
  lease_until TEXT,
  attempt_count INTEGER NOT NULL,
  last_error TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (save_id, source_id)
);
```

The payload contains the post-action `WorldSnapshot` and canonical failed `ActionResult`. The source ID uses the existing `decision:<sessionId>:<snapshotVersion>` format. Status is limited by code to `pending`, `processing`, or `completed`.

## Components

### Atomic result attachment

`memory.SQLite.AttachDecisionResult` receives both the action result and the newer snapshot. Inside one transaction it compares the recorded action, atomically installs the first result, and inserts a pending job only when that canonical result failed. Duplicate identical results return `firstWrite=false`; conflicting content returns an error.

### Lease-based job store

The memory package exposes three narrow operations:

- claim the oldest pending or expired job for a save with a fresh opaque lease token;
- complete the job only when the caller still owns that token;
- release the job to pending after failure, incrementing its attempt count and storing a sanitized diagnostic.

Claims use a conditional SQLite update, so concurrent HTTP requests cannot both own the same live job. Lease timestamps are UTC RFC3339Nano values. The worker uses a fixed bounded lease longer than the configured model timeout.

### Reflection worker

The experience package owns a `ProcessPending` operation. It claims at most one job, invokes the existing idempotent `LearnFromResult` path, and completes the lease after the canonical experience outcome is durable. On an error it releases the lease and returns the error to the policy layer.

The policy layer attempts one pending job after result attachment and before preparing the next action. It also attempts one job at the beginning of a later `NextDecision`, which is the restart recovery path. Reflection failure is deliberately non-fatal to gameplay: Echo still computes a safe next action while the durable job remains retryable.

## Data flow

```text
SMAPI reports failed action + newer snapshot
  -> SQLite transaction: attach canonical result + insert pending job
  -> worker conditionally leases one job
  -> Eino produces bounded causal observation
  -> deterministic merge + idempotent experience persistence
  -> worker marks lease completed
  -> policy ranks the next action using the new experience
```

If inference fails, the worker releases the job to `pending`. The next action request, duplicate result request, or later restarted session retries it before normal policy preparation.

## Failure handling

- Invalid result/snapshot correlation is rejected before the transaction.
- Missing decisions return the existing not-found error and enqueue nothing.
- Lease loss prevents a stale worker from completing or releasing another worker's job.
- Stored diagnostics are truncated and never include model credentials or raw prompts.
- Repeated model failure does not block saving, quitting, or deterministic action safety checks.
- No unbounded retry loop runs inside one request; each policy request processes at most one job.

## Verification

- SQLite tests prove atomic enqueue, duplicate/conflict behavior, concurrent single-winner attachment, lease exclusion, release/retry, expiry recovery, completion, and reopen persistence.
- Experience tests prove a model failure releases the job and a later attempt completes it without duplicate experience application.
- Policy tests prove failed-result handling attempts pending reflection once per request and that reflection failure does not prevent safe replanning.
- Full Go race tests, Go vet, all C# bridge tests, three cross-process demos, Nexus packaging smoke, deterministic sidecar rebuild, checksum verification, and native macOS health smoke remain release gates.

## Compatibility

`CREATE TABLE IF NOT EXISTS` migrates existing databases without rewriting current records. Existing decision and experience JSON remain readable. The localhost HTTP and C# contracts do not change because the newer snapshot is already part of the action-result request.
