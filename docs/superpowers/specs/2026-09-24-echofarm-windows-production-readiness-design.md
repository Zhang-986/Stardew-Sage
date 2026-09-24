# EchoFarm Windows Production Candidate Design

## Goal

Turn the existing EchoFarm prototype into a release candidate that a player can install and evaluate on Windows x64 with Stardew Valley 1.6 and SMAPI 4.1+ in single-player mode. The release candidate must fail closed, protect real saves from unsupported game mutations, expose model usage and actionable setup errors, and produce reproducible evidence from both automation and a real Windows game installation.

This is the first production boundary, not a claim of universal support. Multiplayer, Android, consoles, older Stardew/SMAPI versions, and public Nexus publication remain outside this milestone.

## Current evidence and gaps

The Go/Eino core, SQLite memory, cross-language bridge, deterministic safety policy, reflection jobs, and experience-effectiveness loop have automated coverage. Local verification has passed Go race tests, Go vet, 87 .NET bridge tests, three cross-process demos, concurrency stress, deterministic sidecar builds, packaging shape checks, and a macOS arm64 health smoke.

The current GitHub run for commit `b1a3168` is red on Linux because `github.com/bytedance/sonic/loader v0.5.0` cannot link against the selected Go runtime. The workflow stops before the Go suite and demos. The actual `EchoFarm.Mod` project is not part of `EchoFarm.sln`, because compiling it requires legal game and SMAPI assemblies that CI does not possess. Generated sidecars and packages are intentionally ignored, so the branch contains source but no playable Windows archive.

The game adapter also performs crop harvest through a simplified synthetic transfer: it creates one harvest item and directly removes or resets the crop. That is not sufficient evidence for native Stardew yield, quality, experience, special-crop, multiplayer, or other-Mod semantics. It is unsafe for a valuable save until replaced or disabled.

## Chosen delivery strategy

Use a gated release-candidate pipeline rather than a large rewrite or an AI-first expansion.

1. Restore deterministic CI and add every offline behavior demo to the required workflow.
2. Make unsupported game mutation fail closed, with harvesting disabled by default until the real-game adapter passes certification.
3. Provide a Windows build/install/doctor workflow for the owner machine that has legal game assemblies.
4. Add first-run configuration diagnostics and in-game status so a player does not need to interpret raw SMAPI logs.
5. Add local model-usage accounting and hard budgets before prolonged play.
6. Run a disposable-save Windows certification matrix and use its evidence to decide whether harvesting can be enabled by default.

A big-bang public release was rejected because game behavior cannot be validated on the current machine. Continuing to add autonomous gameplay before the safety, cost, and installation gates was rejected because it expands the failure surface without making the current vertical slice trustworthy.

## Supported product boundary

The first candidate supports:

- Windows x64;
- Stardew Valley 1.6;
- SMAPI 4.1 or newer within the tested 4.x range;
- single-player saves;
- one local player and one visual Echo;
- walking, watering, refilling, depositing, and safe stopping;
- harvesting only behind a capability flag until native-game certification passes;
- fixture mode for installation smoke and OpenAI-compatible JSON mode for real AI;
- local SQLite memory isolated by save ID.

Every unsupported environment must produce a visible, actionable disabled state rather than attempting partial execution.

## Build and CI gates

The Go toolchain and Eino/Sonic dependency set will be made explicit and tested on GitHub-hosted Linux and Windows runners. The Sonic linker failure must be reproduced in a focused CI-compatible command, fixed through a supported dependency/toolchain combination, and covered by a workflow job that builds and starts the native sidecar.

The required workflow will run:

- Go race tests and vet on Linux;
- .NET bridge tests;
- Go-to-.NET contract integration;
- core, Continuum, and reflective cross-process demos;
- Windows x64 sidecar build and `/healthz` smoke on a Windows runner;
- four-platform deterministic cross-build and package-shape verification.

The workflow cannot legally compile the game-backed Mod without game assemblies. Instead, the owner Windows machine becomes the signed-off game-adapter gate. Its script must record game version, SMAPI version, source commit, build result, package checksum, and smoke checklist result without copying either game DLL into the repository or archive.

## Safe game adapter

The model remains restricted to a closed high-level action catalog. Go validates proposal identity, confidence, available targets, and player claims; C# validates the response again against the latest snapshot before scheduling work on the SMAPI game thread.

For production safety, each action also has a declared capability. Unsupported or uncertified capabilities are absent from the world snapshot and rejected by the C# gate even if a model invents them. `EnableExperimentalHarvest` defaults to `false` for the first candidate. Watering, refill, and deposit retain their existing bounded implementations; deposit continues to remove only quantities accepted by the chest.

Harvest certification requires replacing the synthetic one-item mutation with a game-backed adapter that preserves native yield, quality, regrowth, experience, sounds/events, special crops, and relevant SMAPI hooks. If the target game API cannot provide those semantics without impersonating the player, harvesting remains disabled rather than emulated.

World changes, save, return to title, stale snapshots, model timeout, invalid JSON, path failure, and exhausted model budget stop Echo without blocking the game. The production smoke starts with a disposable copied save; no test runs first on the player's only save.

## Windows installation and diagnostics

Add a PowerShell entry point with three explicit operations:

- `Test-EchoFarmPrerequisites`: locate or accept the Stardew directory, verify `Stardew Valley.dll`, `StardewModdingAPI.dll`, Windows x64, Go, .NET, and a writable Mods directory;
- `Build-EchoFarmPackage`: build the real Mod against the local game, build the Windows sidecar, stage the exact `EchoFarm/` tree, validate forbidden files, and create a checksum;
- `Install-EchoFarm`: preserve an existing `config.json`, install through a temporary directory plus rename, and support removal of program files without deleting the SQLite memory database.

The script must never copy game assemblies, API keys, databases, logs, or PDB files into the package. Every failure message includes the failed prerequisite and the exact corrective action. A dry-run/doctor mode performs no installation.

For an eventual public release, players receive the validated ZIP and do not need Go or the .NET SDK. Source builds remain a developer/owner workflow only.

## First-run and in-game experience

The Mod validates configuration before launching the sidecar. Missing model URL, model name, API key, unsupported mode, occupied port, missing executable, and unhealthy sidecar become distinct setup states. F9 displays a compact status panel containing mode, core health, model readiness, database path, current session state, and the latest safe error.

Fixture mode is visibly labeled `DEMO` and never masquerades as real AI. OpenAI mode remains fail closed and never silently falls back to fixture behavior. The API key stays environment-only; configuration and logs may state whether it is present but never display or persist its value.

Player-facing messages explain the lifecycle:

- F7 starts or completes teaching;
- F8 is enabled only after memory exists and the core is healthy;
- F9 always opens status, including setup failures;
- F10 is available only when a decision can be corrected;
- stopping, timeout, and budget exhaustion explain what happened and how to resume.

## Model usage and cost controls

Every model request receives a local request ID and purpose: `learning`, `intent`, `action`, `recovery`, or `reflection`. The generator records start/end time, status, provider-reported prompt/completion/total tokens when present, and a redacted error class. Prompt or response bodies, API keys, and provider error bodies are not stored in the usage ledger.

Usage is appended to SQLite and exposed through a local read endpoint plus the F9 panel. Counters include session/day calls, reported tokens, failures, and recent latency. Missing provider token metadata is shown as unknown, never estimated as fact.

The first candidate enforces configurable per-session call and reported-token budgets. Crossing either budget prevents another model call and returns a typed safe-stop result. Idempotent demonstrations, decisions, corrections, and results continue to reuse stored outcomes and therefore consume no duplicate calls. Cost in currency is not calculated because provider/model pricing is external and mutable.

## AI utilization after the production foundation

The first candidate preserves the current per-high-level-action decision loop because it is observable and bounded. Once Windows certification supplies latency, token, and action-success measurements, the next design can maximize AI use without maximizing API waste:

- let the model produce a short intent-level work plan, while deterministic code expands repeated crop actions;
- replan only when the world invalidates the plan, the player claims a target, or an action fails;
- route cheap intent classification and routine actions to a smaller model while reserving a stronger model for learning and reflection;
- combine player profile, applicable experience, current goals, and effectiveness evidence into explicit decision explanations;
- extend actions only through new capability adapters with their own safety and certification gates.

This keeps AI responsible for generalization, prioritization, coordination, and reflection while local code owns repetitive execution and irreversible game mutations.

## Verification and release decision

Automated release gates:

- all required GitHub jobs green on the exact release commit;
- race-enabled Go suite and full .NET bridge suite pass;
- all three cross-process demos pass;
- duplicate/canonical-result stress tests pass twenty repetitions;
- Windows sidecar starts and returns HTTP 200 from `/healthz`;
- package validation proves one root folder, one sidecar, expected DLLs, no secrets/game assemblies, and a matching SHA-256 checksum;
- configuration, budget exhaustion, model timeout, malformed model output, stale action, and process restart tests all fail closed.

Real Windows game gates:

- Release build succeeds against the installed Stardew/SMAPI versions;
- fixture-mode teaching, summon, status, correction, save, reload, day change, and title return complete without an unhandled exception;
- supported actions are checked on an expendable save, including empty can, blocked route, full Echo inventory, full chest, and target disappearance;
- real-model mode completes teaching and a bounded work session while recording calls, reported tokens, latency, and failures;
- the generated smoke report and SMAPI log contain no API key or provider body;
- uninstall leaves the game save intact and removes only EchoFarm program files.

The candidate is `No-Go` for a valuable save or public Nexus upload until every automated gate is green and the real Windows checklist is recorded. It is `Conditional Go` for a disposable-save owner test once CI is green, harvesting is disabled by default, and the Windows doctor passes.

## Rollback and support boundary

EchoFarm memory is external to the game save. Before testing, the installer identifies the save directory and instructs the owner to copy the selected save. Runtime abort never deletes a game save. Uninstall removes the Mod folder but preserves the local EchoFarm database unless the player explicitly requests a full data reset.

The first candidate does not promise multiplayer correctness, compatibility with every crop/content Mod, cloud synchronization, automatic provider billing, or unattended public rollout. Those claims require separate evidence after the Windows single-player gate is complete.
