# EchoFarm Nexus Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a Nexus-ready EchoFarm distribution whose SMAPI mod safely supervises a bundled Go/Eino core and whose platform archives are reproducible and validated.

**Architecture:** `EchoFarm.Bridge` owns a dependency-injected core process supervisor and its tests. `EchoFarm.Mod` supplies configuration, paths, environment variables, and SMAPI logging. A repository script builds the native Go sidecars, consumes a real Release build of the Mod, validates staged contents, and emits one ZIP per platform without secrets or game assemblies.

**Tech Stack:** C# 12/.NET 6, xUnit, Go 1.24, CloudWeGo Eino, Bash, `zip`, SHA-256 tooling, SMAPI 4.1+.

---

## File Map

- `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/CoreProcessSupervisor.cs`: health-first process ownership state machine.
- `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/SystemCoreProcess.cs`: real HTTP probe and child-process adapters.
- `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/CoreProcessSupervisorTests.cs`: lifecycle, timeout, and ownership tests.
- `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`: safe sidecar and model configuration.
- `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`: starts the core at launch and terminates only an owned child at process exit.
- `scripts/package-nexus.sh`: Release build, cross-compilation, staging, validation, ZIP, and checksums.
- `scripts/verify-nexus-package.sh`: reusable archive-structure and secret-exclusion checks.
- `scripts/tests/package-nexus-smoke.sh`: fixture-DLL packaging smoke test that needs no game installation.
- `release/nexus/*.md`: listing copy, install/configuration guide, permissions, screenshots, and upload checklist.
- `.github/workflows/ci.yml`: repeatable Go, .NET, demo, and packaging-smoke verification.

### Task 1: Build a Tested Core Process Supervisor

**Files:**
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/CoreProcessSupervisorTests.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/CoreProcessSupervisor.cs`
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/SystemCoreProcess.cs`

- [x] **Step 1: Write failing lifecycle tests**

Test that a healthy pre-existing core is reused without launching, disabled auto-start reports unavailable, an owned child becomes ready after polling, a child exit reports its exit code, timeout terminates the owned child, repeated starts do not launch twice, and `Stop` never terminates an external process.

- [x] **Step 2: Verify red**

Run:

```bash
./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln --filter FullyQualifiedName~CoreProcessSupervisorTests
```

Expected: compilation fails because the supervisor contracts do not exist.

- [x] **Step 3: Implement minimal lifecycle types**

Add `CoreHostState`, `CoreLaunchOptions`, `ICoreHealthProbe`, `ICoreProcess`, `ICoreProcessLauncher`, and `IAsyncDelay`. `StartAsync` probes first, owns only a process it launched, polls a bounded number of times, and terminates that child on timeout. `Stop` is idempotent.

- [x] **Step 4: Add real adapters**

`HttpCoreHealthProbe` performs a bounded `/healthz` request. `SystemCoreProcessLauncher` uses `UseShellExecute=false`, redirects output, passes an explicit environment map, and kills the complete owned process tree.

- [x] **Step 5: Verify and commit**

Run the filtered tests and the complete .NET suite, then commit with `feat(echofarm-mod): supervise bundled Go core`.

### Task 2: Wire Sidecar Startup Into SMAPI

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/manifest.json`
- Modify: `stardew-echo-mod/README.md`

- [x] **Step 1: Add safe configuration**

Add auto-start, executable override, startup timeout, model mode, model URL/name, and database path. Reject non-loopback `CoreUrl`, invalid timeouts, and any model mode other than `openai` or explicitly selected `fixture`. Do not add an API-key field.

- [x] **Step 2: Resolve the launch environment**

Use `core/echofarm-core.exe` on Windows and `core/echofarm-core` elsewhere unless overridden. Put the default SQLite file below the OS local application-data directory. Pass non-empty model settings as `ECHOFARM_*` variables and preserve an inherited `ECHOFARM_MODEL_API_KEY`.

- [x] **Step 3: Start and stop safely**

Start asynchronously on `GameLaunched`, await the same startup task before restoring learned state, stream child logs through `IMonitor`, and register process-exit cleanup. A failed sidecar startup disables Echo while leaving the game loop usable.

- [x] **Step 4: Verify and commit**

Run all game-independent tests and the expected SMAPI build gate, then commit with `feat(echofarm-mod): auto-start local AI core`.

### Task 3: Build and Validate Nexus Archives

**Files:**
- Create: `scripts/package-nexus.sh`
- Create: `scripts/verify-nexus-package.sh`
- Create: `scripts/tests/package-nexus-smoke.sh`
- Modify: `.gitignore`

- [x] **Step 1: Write a failing package smoke test**

Create temporary fixture Mod DLLs and invoke package mode without a game build. Assert four ZIPs exist, each has one `EchoFarm/` root, contains the manifest, both DLLs, exactly one correctly named native core, license, and player README, and contains no `.env`, database, log, PDB, or game DLL.

- [x] **Step 2: Verify red**

Run `bash scripts/tests/package-nexus-smoke.sh`.

Expected: FAIL because the package scripts do not exist.

- [x] **Step 3: Implement the validator and packager**

Support `--version`, `--game-path`, `--mod-build-dir`, `--output-dir`, and optional `--nexus-mod-id`. In normal mode, run tests and build `EchoFarm.Mod.csproj`; in fixture mode, consume caller-provided DLLs. Cross-compile `./cmd/echofarm` for `windows/amd64`, `linux/amd64`, `darwin/amd64`, and `darwin/arm64`, stage each archive, validate it, and emit `SHA256SUMS.txt`.

- [x] **Step 4: Verify green and commit**

Run the smoke test, inspect all ZIP listings, and commit with `build(echofarm): package Nexus release archives`.

### Task 4: Prepare Nexus Listing and CI Evidence

**Files:**
- Create: `release/nexus/README.md`
- Create: `release/nexus/description.md`
- Create: `release/nexus/changelog.md`
- Create: `release/nexus/permissions.md`
- Create: `release/nexus/screenshot-checklist.md`
- Create: `.github/workflows/ci.yml`
- Modify: `README.md`

- [x] **Step 1: Write the player and publisher material**

State prerequisites, one-folder installation, OpenAI-compatible configuration, explicit fixture-demo behavior, privacy boundaries, AI-use disclosure, known limitations, supported platforms, required screenshots, and the exact Nexus upload fields. Do not claim live-game validation before it occurs.

- [x] **Step 2: Add CI for redistributable checks**

Run Go tests with race detection, Go vet, .NET bridge tests, the cross-process demo, and the fixture packaging smoke test. Do not download or redistribute Stardew assemblies in CI.

- [x] **Step 3: Verify and commit**

Run the complete local verification matrix and commit with `docs(echofarm): prepare Nexus release`.

### Task 5: Produce the Overnight Handoff

**Files:**
- Modify: `docs/superpowers/plans/2026-09-24-nexus-release.md`

- [ ] **Step 1: Build local Go release binaries**

Cross-compile all four supported targets into a temporary artifact directory and record sizes and SHA-256 hashes. Do not commit generated binaries.

- [ ] **Step 2: Re-run release gates**

Run:

```bash
./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln
(cd echofarm-core && go test -race -count=1 ./... && go vet ./...)
./demo/run-core-demo.sh
bash scripts/tests/package-nexus-smoke.sh
```

Attempt the real SMAPI build and record the exact external failure if no game path is available.

- [ ] **Step 3: Audit publication blockers**

Confirm whether a Nexus API key/mod ID and legal game installation exist. If absent, leave the validated archives unuploaded and provide the exact one-command continuation; never publish a fixture or source-only archive as a playable release.
