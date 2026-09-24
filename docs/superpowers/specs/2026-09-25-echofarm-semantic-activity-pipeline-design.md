# EchoFarm Semantic Activity Pipeline Design

## Goal

Turn player gameplay beyond the original farm loop into evidence that EchoFarm can classify, learn from, explain, and safely use. The first expansion covers tree chopping, rock breaking, mine-floor traversal, and fishing outcomes while preserving the existing watering, refill, harvest, and deposit behavior.

This milestone makes the observation-to-memory path general enough for a public demo. It does not claim that Echo can autonomously perform every observed activity. New mutations remain disabled until each game adapter has its own capability gate and native-game certification.

## Current limitation

The current C# adapter classifies only movement, watering, watering-can refill, crop harvest, and chest deposit. It infers those actions from the selected tool or action button plus one before/after sample. Go then groups the events into four behavior kinds and the model can emit only four player-trait keys.

That is a valid farm vertical slice, but not a general player-model system. Tree chopping is a multi-hit episode, fishing is a multi-stage interaction, mine progress is a location transition, and resource gathering needs item-level deltas rather than a single inventory-count difference. Treating all four as another one-frame button mapping would produce misleading evidence.

## Public delivery standard

An activity is considered supported for learning only when all of the following are true:

1. The Mod emits a versioned semantic event with stable save/session identity, game context, target classification, and an explicit outcome.
2. The event is derived from observable before/after game state rather than from a button press alone.
3. C# and Go reject unknown enum values, impossible deltas, mismatched identities, and malformed evidence.
4. Go groups adjacent events into a bounded activity episode and preserves the evidence IDs used by AI.
5. The model returns only bounded profile observations and cites supplied episode evidence.
6. Deterministic merging handles repetition and contradiction across days.
7. F9 shows what was learned and the supporting activity category.
8. Fixture-mode cross-process tests prove the complete observation payload through persisted memory without a paid model.

An activity is considered supported for autonomous execution only after the separate action capability, deterministic validation, latest-snapshot recheck, result feedback, and Windows disposable-save certification all pass. Observation support never implicitly enables execution.

## Approaches considered

### Add more conditionals inside `StardewGamePort`

This is the smallest patch, but it keeps classification coupled to SMAPI input handling and assumes every activity completes in one tick. It does not scale to fishing or repeated axe strikes and cannot be unit tested without game assemblies.

### Send raw frames or logs to the model

This avoids writing classifiers but increases token use, leaks irrelevant game state, makes outcomes hard to audit, and lets the model invent semantics. It is rejected.

### Versioned semantic activity protocol

This is the selected approach. Game-specific sensors create normalized observations; a pure C# classifier owns episode state and emits strict contracts; Go validates, segments, and learns from those contracts. The model receives compact named facts rather than raw frames, and unsupported execution stays visibly disabled.

## Architecture

The data path is:

```text
SMAPI game state
  -> game-specific sensor samples
  -> pure C# semantic activity classifier
  -> DemonstrationEvent v2
  -> Go validation and episode segmentation
  -> Eino bounded trait inference
  -> deterministic multi-day merge
  -> SQLite evidence and F9 memory view
```

The existing one-frame observation remains valid for watering, refill, crop harvest, and deposit. New activity sensors use explicit lifecycles:

- tree chopping starts on an axe hit against a tree and completes when the target falls, changes identity, the player leaves the context, or a bounded timeout expires;
- rock breaking starts on a pickaxe hit against a breakable world object and completes when the target disappears or the attempt times out;
- mine traversal is emitted from a confirmed location/floor transition rather than an input guess;
- fishing starts when a rod cast begins and completes on caught, escaped, cancelled, location changed, or timeout.

The pure classifier accepts game-neutral samples, so all state transitions and timeout behavior can be tested in `EchoFarm.Bridge.Tests`. `EchoFarm.Mod` remains a thin translator from Stardew types into those samples.

## Contract

`Demonstration` gains `schemaVersion`. Existing payloads that omit it are read as version 1; the Mod emits version 2 after this change, and every new activity kind requires version 2. This gives stored demonstrations and API clients an explicit migration boundary without invalidating existing local memory.

`DemonstrationEvent` keeps its existing identity, kind, tick, position, target, tool, delta, success, and error fields. It gains:

- `location` and `timeOfDay` for context-sensitive routines;
- `targetKind` for stable semantic target categories;
- `itemDeltas`, each containing item ID, display name, and signed quantity;
- optional `healthDelta` and `mineFloorDelta` in `StateDelta`;
- optional `durationTicks` for multi-tick episodes.

New event kinds are:

- `chop_tree`;
- `break_rock`;
- `enter_mine_floor`;
- `fish_caught`;
- `fish_escaped`.

New behavior segments are:

- `woodcutting`;
- `mining`;
- `mine_traversal`;
- `fishing`.

Item deltas are bounded to 32 entries per event, zero quantities are forbidden, duplicate item IDs are merged deterministically, and display names are informational only. Stable item IDs, not localized names, are evidence keys.

## Classification rules

The classifier does not infer broad player intent. It emits a semantic event only when state proves the outcome:

- a tree event succeeds when the same target transitions from present to felled/removed;
- a rock event succeeds when the same target transitions from present to removed;
- a mine-floor event succeeds when the location remains a mine and the floor changes to a valid positive floor;
- a fishing event succeeds only when the fishing lifecycle reports a caught item; an escape is a separate unsuccessful event;
- timeout, cancellation, target replacement, and location changes produce bounded error codes.

Repeated tool hits belonging to one target are one episode, not separate learned actions. Movement remains route context and is not promoted to a lifestyle trait on its own.

## AI learning

The model does not see raw key presses or frames. It receives validated episodes such as “six axe hits felled one oak tree, consumed 10 energy, yielded 14 wood, sunny Farm morning” or “entered mine floor 45, left at 22:40 with low health and 12 coal.”

The bounded trait vocabulary expands with:

- `activity_order` for ordering farm, woodcutting, mining, and fishing blocks;
- `resource_priority` for repeatedly pursued item categories;
- `mine_exit_policy` for observed health, time, or inventory stop patterns;
- `fishing_context` for repeated location, weather, and time preferences.

These are observations, not executable commands. Each must cite real event IDs, and one day cannot produce a fully trusted habit. Existing deterministic confidence and contradiction merging remains authoritative.

The learned traits are supplied to the runtime action policy immediately. In this milestone they may change prioritization, explanations, and safe stopping among already certified actions. They do not unlock chopping, mining, or fishing actions. A later work-plan milestone can use the same profile to generate a day-level division of labor.

## Player experience

F7 records every supported semantic activity during the teaching window. When learning completes, F9 shows the strongest stable lifestyle traits alongside existing farm traits. Fixture mode exposes the same path without an API key and labels it as demo data.

If the player performs an activity whose sensor is unavailable, the Mod ignores it rather than sending an `unknown` event to the model. Unsupported observation and unsupported execution are shown separately in diagnostics so the UI never implies that Echo can perform something it has only observed.

## Privacy and reliability

The protocol stores semantic activity facts, not screenshots, chat, raw input streams, save-file contents, prompt bodies, or model responses. In real-model mode, the configured provider receives the compact demonstration, current player model, and behavior segments required for inference. Fixture mode sends nothing externally.

All event lists and strings are bounded before model invocation. Existing per-session model budgets, local request ledger, timeout handling, and fail-closed behavior continue to apply. Invalid model output does not alter the saved player model.

## Verification

Automated acceptance requires:

- C# unit tests for successful, failed, cancelled, target-changed, and timeout episodes for all four new activity families;
- cross-language JSON round-trip tests for every new field and enum;
- Go validation tests for bounds and impossible data;
- segmenter tests proving multi-hit tree/rock activity is not over-counted;
- learning tests proving new traits cite supplied evidence and merge across days;
- a fixture cross-process demo that teaches at least woodcutting, mine traversal, and fishing, then reads the persisted traits through the memory API;
- all existing Go race/vet, .NET, PowerShell, demo, package, and Windows sidecar CI gates remain green.

The owner-machine Windows gate additionally proves that actual Stardew/SMAPI signals map to the expected semantic events. Until that signed evidence exists, the repository may claim protocol and classifier support but not universal native-game certification.
