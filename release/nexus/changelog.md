# 0.4.0 — Reflective Policy

- Added bounded Eino reflection for failed actions and explicit player corrections.
- Added persistent, save-scoped policy experiences with deterministic merge, contradiction handling, and idempotent evidence revisions.
- Added ranked action proposals with calibrated policy confidence and up to two validated fallback actions.
- Added F10 correction capture: Echo pauses, learns from the next successful farm action, and safely retries temporary model failures.
- Expanded F9 memory with decision confidence, fallback count, and failure/player-correction evidence.
- Added a five-stage cross-process demo proving proactive failure avoidance and corrected chest selection across restarts.
- Enforced OpenAI-compatible JSON mode and tightened causal, evidence-bounded proposal/reflection prompts.
- Made concurrent action-result attachment first-write-wins so retries cannot overwrite the canonical outcome.
- Added leased SQLite reflection jobs so transient model failures and process restarts no longer discard pending failure learning.
- Added an append-only effectiveness ledger for experiences actually used by executed actions, with atomic idempotent attribution, explainable success/contradiction/neutral counts, effective-confidence ranking, and cooling after repeated contradictions.
- Extended the reflective cross-process demo with a sixth stage proving that a successful corrected action strengthens the exact experience that drove it.

# 0.3.0 — Continuum co-player

- Added multi-day, evidence-backed player traits with deterministic confidence growth and contradiction handling.
- Added weather-scoped habits so a rainy-day routine does not erase sunny-day behavior.
- Added live player-intent inference and target claims so Echo takes complementary work instead of competing with the player.
- Added an idempotent learning revision store and per-action execution ledger.
- Added the F9 Echo Memory panel with trait confidence and latest decision reasoning.
- Added a reproducible four-day cross-process collaboration demo.

# 0.2.0 — Echo vertical slice

- Added player-demonstration recording and structured player memory.
- Added the Go + CloudWeGo Eino learning, decision, and failure-replanning graphs.
- Added a translucent in-world Echo with bounded tile pathfinding.
- Added watering and watering-can refill actions.
- Added a separate 12-slot Echo inventory.
- Added mature crop harvesting and lossless partial chest deposits.
- Added `inventory_full` recovery through the learned preferred chest and safe stop on `chest_full`.
- Added strict C#/Go JSON contracts, stale-action rejection, save isolation, and localhost-only networking.
- Added automatic bundled-core startup with external-process ownership protection.
