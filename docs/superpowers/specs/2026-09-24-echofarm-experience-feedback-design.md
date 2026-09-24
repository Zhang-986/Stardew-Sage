# EchoFarm Experience Effectiveness Feedback Design

## Goal

Close the AI learning loop with real gameplay outcomes. When Echo executes a primary action whose ranking materially used policy experience, the canonical action result becomes durable feedback for those experience IDs. Later decisions rank experience by both semantic evidence and observed effectiveness instead of trusting every reflection forever.

## Chosen approach

EchoFarm will use an append-only SQLite feedback ledger plus deterministic projection. Directly mutating `confidence` was rejected because concurrent outcome order would change the answer and erase audit history. A contextual bandit was rejected because exploration would make the game demo nondeterministic and introduce policy risk without enough observations.

No model call is needed for outcome feedback. Eino still creates and applies semantic policy experience; Go owns attribution, idempotency, scoring, cooling, persistence, and presentation.

## Feedback eligibility

A decision creates experience feedback only when all of these are true:

- the persisted decision has a structured proposal;
- the primary candidate was selected (`selectedCandidate == 0`);
- the final action is not `stop_session`;
- `appliedExperienceIds` is non-empty;
- the first canonical result is being attached.

If policy safety selected an alternative or stopped Echo, proposal-level experience cannot be reliably attributed to the executed action and receives no feedback.

Every eligible applied experience ID receives the same outcome for that canonical decision. Proposal validation already guarantees that these IDs were supplied by the matcher, are unique, and exist for the current save.

## Deterministic attribution

Feedback has three closed outcomes:

- `succeeded`: the action result succeeded;
- `contradicted`: the selected strategy failed for a precondition the strategy should have anticipated;
- `neutral`: execution or world churn failed without proving the experience wrong.

The first attribution catalog is deliberately narrow:

| Executed action | Result | Feedback |
| --- | --- | --- |
| any supported action | `succeeded` | `succeeded` |
| `water_target` | `out_of_water` | `contradicted` |
| `harvest_target` | `inventory_full` | `contradicted` |
| `deposit_items` | `chest_full` or `inventory_empty` | `contradicted` |
| any action | `path_blocked`, `target_changed`, `unsupported_crop`, or another failure | `neutral` |

This prevents route changes and stale world state from poisoning a sound high-level rule.

## Data model

Add an append-only table:

```sql
CREATE TABLE IF NOT EXISTS experience_feedback (
  save_id TEXT NOT NULL,
  session_id TEXT NOT NULL,
  snapshot_version INTEGER NOT NULL,
  experience_id TEXT NOT NULL,
  outcome TEXT NOT NULL,
  error_code TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (save_id, session_id, snapshot_version, experience_id)
);
```

The result attachment transaction inserts feedback rows after installing the canonical result and before commit. `error_code` stores a recognized game failure code or the literal `other`, never arbitrary input. A missing referenced experience is a consistency error and rolls back both result and feedback. Identical retries see the stored result and cannot add rows; conflicting retries remain rejected.

The base `policy_experiences` JSON remains semantic evidence only. Feedback is projected when experiences are read, so a reflection write that began before an action result cannot overwrite newer effectiveness data.

## Effective confidence

`PolicyExperience.confidence` remains semantic evidence confidence. Projected experience adds:

- `effectiveConfidence`;
- `successCount`;
- `failureCount`;
- `neutralCount`.

With a fixed prior weight of four:

```text
effectiveConfidence =
  (confidence * 4 + successCount) /
  (4 + successCount + failureCount)
```

Neutral outcomes are visible but do not change the score. The formula is bounded, deterministic, order-independent, and exactly reproduces semantic confidence before attributed outcomes exist.

An experience is cooled and excluded from matching when it has at least three contradicted outcomes and `effectiveConfidence < 0.40`. New failure reflection or an explicit player correction may change its semantic confidence or create a replacement experience, allowing future policy recovery without random exploration.

## Components

### Domain feedback policy

A focused domain module defines the outcome enum, attribution catalog, effective-confidence formula, and cooling predicate. It rejects invalid counts and non-finite confidence.

### Atomic feedback ledger

`memory.SQLite.AttachDecisionResult` inspects the canonical `DecisionRecord` inside its existing transaction. For eligible decisions it verifies each applied experience belongs to the save and inserts one immutable feedback row before committing the result.

### Experience projection and ranking

`ListPolicyExperiences` loads base experience JSON, aggregates feedback counts by experience ID, and returns projected copies. The matcher sorts by effective confidence, then semantic evidence count, then ID, and filters cooled experience. Reflection persistence strips projection-only fields before updating base JSON.

### Game-facing explanation

The Go/C# contract gains optional effectiveness fields. F9 renders concise evidence such as `有效 78% · 成功 3 / 失败 1`; neutral results remain auditable without lowering the score. Old databases and older stored JSON deserialize with zero counters and fall back to semantic confidence.

## Error and concurrency behavior

- Feedback insertion is part of canonical result commit; there is no result-without-feedback crash window.
- The composite primary key makes feedback idempotent per decision and experience.
- All experience IDs are checked against the current save before insertion.
- Feedback projection never rewrites the ledger.
- A malformed historical feedback row fails the read instead of silently changing ranking.
- Result diagnostics never contain prompts, arbitrary input, or model credentials; only recognized game error codes or `other` are stored.

## Verification

- Domain tests cover every attribution row, formula bounds/order independence, and cooling threshold.
- SQLite tests cover atomic result-plus-feedback commit, duplicates, conflicts, multiple applied IDs, non-primary exclusion, save isolation, and projection across reopen.
- Concurrency tests retain one canonical result and one feedback row per experience under sixteen writers.
- Matcher tests prove effective ordering and cooled filtering.
- C# strict JSON and F9 presenter tests cover the new fields.
- The reflective cross-process demo reports effectiveness before and after a successful experience-guided deposit.
- Full Go race/vet, .NET, three demos, Nexus packaging, deterministic sidecar builds, checksums, and native health smoke remain release gates.

## Compatibility and non-goals

The migration is additive through `CREATE TABLE IF NOT EXISTS`. Existing experience JSON remains valid; projected fields are optional. This phase does not add reward-model calls, random exploration, long-horizon credit assignment, player scoring, or autonomous modification of the action allowlist.
