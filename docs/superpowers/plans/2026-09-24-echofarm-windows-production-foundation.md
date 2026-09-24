# EchoFarm Windows Production Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a CI-green, fail-closed Windows single-player release-candidate foundation with explicit action capabilities, actionable startup diagnostics, and a repeatable local game build/install workflow.

**Architecture:** Keep the game-independent C# Bridge and Go service fully testable in CI. Carry an explicit capability contract in each world snapshot so Go cannot propose an uncertified action and C# rejects one again before mutation. A PowerShell owner workflow performs the only game-backed build, stages the Mod atomically, preserves user configuration, and emits local evidence without redistributing game assemblies.

**Tech Stack:** Go 1.24.1+, CloudWeGo Eino, SQLite, C#/.NET 6 and 8, SMAPI 4.1+, PowerShell 5.1+, GitHub Actions.

---

### Task 1: Restore CI and make the full offline suite mandatory

**Files:**
- Modify: `echofarm-core/go.mod`
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Capture the existing CI failure**

Record GitHub run `36013202880` as the red baseline. The expected failure is `github.com/bytedance/sonic/loader: invalid reference to runtime.lastmoduledatap` while `setup-go` resolves the `go 1.24.0` directive. Upstream Sonic issue 771 identifies Go 1.24.0's linker defect and requires Go 1.24.1 or newer.

- [ ] **Step 2: Pin the minimum fixed toolchain**

Change the module directive to:

```go
go 1.24.1
```

Keep `actions/setup-go` reading `go-version-file` so the local contract and CI runtime cannot drift independently.

- [ ] **Step 3: Extend the required workflow**

Add the missing reflective demo after the Continuum demo:

```yaml
- name: Run reflective experience demo
  run: ./demo/run-reflective-demo.sh
```

Add a `windows-sidecar` job on `windows-latest` that installs the same Go version, runs `go test ./...`, builds `echofarm-core.exe`, starts it in fixture mode on loopback with a temporary database, polls `/healthz`, and always terminates the process.

- [ ] **Step 4: Verify locally and on GitHub**

Run:

```bash
cd echofarm-core
go test -race -count=1 ./...
go vet ./...
cd ..
./demo/run-reflective-demo.sh
```

Push the task commit and require both Linux `verify` and Windows `windows-sidecar` jobs to pass on that exact SHA.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/go.mod .github/workflows/ci.yml
git commit -m "ci(echofarm): restore cross-platform release gates"
```

### Task 2: Carry explicit safe-action capabilities end to end

**Files:**
- Modify: `contracts/world-snapshot.schema.json`
- Modify: `echofarm-core/internal/domain/types.go`
- Modify: `echofarm-core/internal/domain/validate.go`
- Modify: `echofarm-core/internal/domain/validate_test.go`
- Modify: `echofarm-core/internal/policy/service.go`
- Modify: `echofarm-core/internal/policy/service_test.go`
- Modify: `echofarm-core/internal/intelligence/fixture_generator.go`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Contracts/WorldSnapshot.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/ActionSafetyGate.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/ActionSafetyGateTests.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/WorldSnapshotMapper.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`

- [ ] **Step 1: Write failing Go capability tests**

Add an optional snapshot capability object:

```go
type ActionCapabilities struct {
    Harvest bool `json:"harvest"`
}

type WorldSnapshot struct {
    // existing fields
    Capabilities *ActionCapabilities `json:"capabilities,omitempty"`
}
```

Tests must prove that a legacy `nil` capability preserves fixture compatibility, while an explicit `harvest:false` rejects `harvest_target` before decision persistence. `stop_session`, watering, refill, and deposit remain available.

- [ ] **Step 2: Run Go tests and verify RED**

Run:

```bash
cd echofarm-core
go test ./internal/domain ./internal/policy ./internal/intelligence -run 'Test.*Capabilit' -count=1
```

Expected: compilation or assertion failure because capability-aware validation does not exist.

- [ ] **Step 3: Implement Go capability filtering and validation**

Add one domain helper:

```go
func ActionEnabled(snapshot WorldSnapshot, kind ActionKind) bool {
    return snapshot.Capabilities == nil || kind != ActionHarvestTarget || snapshot.Capabilities.Harvest
}
```

The policy must reject a disabled model candidate and choose a safe enabled alternative when present; otherwise it returns `stop_session`. The fixture generator must not generate harvesting when the explicit capability is disabled.

- [ ] **Step 4: Write failing C# safety tests**

Add `ActionCapabilities` to the strict snapshot contract and require `ActionSafetyGate.EnsureSafe` to reject `HarvestTarget` when `Harvest` is false. Retain a legacy constructor/default path only in game-independent tests; the Mod must always send explicit capabilities.

- [ ] **Step 5: Implement the C# and Mod capability boundary**

Add this default-disabled setting:

```csharp
public bool EnableExperimentalHarvest { get; set; } = false;
```

`WorldSnapshotMapper` receives the configured value and always serializes `Capabilities = new ActionCapabilities { Harvest = enableExperimentalHarvest }`. `ActionSafetyGate` validates the response against the same snapshot value immediately before execution.

- [ ] **Step 6: Verify and commit**

Run the targeted Go and .NET tests, then the full Go race and .NET suites. Commit:

```bash
git commit -m "feat(echofarm): gate uncertified game actions"
```

### Task 3: Make first-run failures actionable in game

**Files:**
- Create: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/SetupReadiness.cs`
- Create: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/SetupReadinessTests.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/EchoMemoryOverlay.cs`

- [ ] **Step 1: Write failing readiness tests**

Cover fixture-ready, OpenAI-ready, missing endpoint, missing model, missing API key, missing bundled executable, invalid loopback URL, and an occupied/unhealthy endpoint. Require stable issue codes and redacted messages that never contain the key value.

- [ ] **Step 2: Verify RED**

Run:

```bash
cd stardew-echo-mod
../.tools/dotnet/dotnet test EchoFarm.sln --no-restore --filter SetupReadinessTests
```

Expected: compilation failure because `SetupReadiness` does not exist.

- [ ] **Step 3: Implement readiness evaluation**

Introduce a pure bridge component returning one of `ready`, `demo_mode`, `missing_model_url`, `missing_model_name`, `missing_api_key`, `missing_core`, or `invalid_core_url`. It accepts only an `ApiKeyPresent` boolean and therefore cannot log the key.

Change the initial Mod configuration to `ModelMode = "fixture"`. Before process startup, evaluate readiness; refuse invalid OpenAI configuration without spawning a process. Keep fixture behavior explicit and expose a `DEMO` marker.

- [ ] **Step 4: Expose status through F9 and player messages**

F9 must open even when no memory exists. The overlay starts with setup mode, core health, session state, database location, harvesting capability, and the latest safe error. Startup failures include a corrective instruction rather than only `Core unavailable`.

- [ ] **Step 5: Verify and commit**

Run the full .NET suite and commit:

```bash
git commit -m "feat(echofarm): add actionable first-run diagnostics"
```

### Task 4: Add a Windows owner build, install, and doctor workflow

**Files:**
- Create: `scripts/windows/EchoFarm.Setup.psm1`
- Create: `scripts/windows/Install-EchoFarm.ps1`
- Create: `scripts/windows/Test-EchoFarmSetup.ps1`
- Modify: `.github/workflows/ci.yml`
- Modify: `release/nexus/README.md`
- Modify: `release/nexus/PLAYER_README.md`

- [ ] **Step 1: Write failing PowerShell module tests**

Use temporary directories containing fixture `Stardew Valley.dll`, `StardewModdingAPI.dll`, manifest, Mod DLLs, and sidecar. Verify path discovery, missing-prerequisite issue codes, package allowlist, preservation of an existing `config.json`, atomic directory replacement, and uninstall that leaves the configured data directory untouched.

- [ ] **Step 2: Run tests and verify RED**

On `windows-latest`, run:

```powershell
pwsh -NoProfile -File scripts/windows/Test-EchoFarmSetup.ps1
```

Expected: failure because the setup module and commands do not exist.

- [ ] **Step 3: Implement doctor and package staging**

`Test-EchoFarmPrerequisites` accepts `-GamePath` or checks the standard Steam location, then returns a structured result without changing disk state. `Build-EchoFarmPackage` invokes `dotnet restore`, a Release Mod build with `GamePath`, and a Windows x64 Go build before validating the staged file allowlist.

- [ ] **Step 4: Implement safe install and uninstall**

`Install-EchoFarm` stages into `Mods/.EchoFarm.installing`, preserves an existing `config.json`, then renames into place. `Uninstall-EchoFarm` removes only `Mods/EchoFarm`; it does not remove `%LOCALAPPDATA%/EchoFarm` unless `-DeleteLocalData` is explicitly supplied.

- [ ] **Step 5: Add Windows CI and documentation**

Run module tests on `windows-latest`. Document one copy-paste path for `-Doctor`, `-Build`, `-Install`, and `-Uninstall`, with fixture mode as the first smoke and environment-only API key configuration for real AI.

- [ ] **Step 6: Verify and commit**

Run PowerShell tests, package-shape smoke, and source secret scans. Commit:

```bash
git commit -m "feat(echofarm): add Windows production setup workflow"
```

### Task 5: Produce the owner-machine acceptance handoff

**Files:**
- Create: `docs/echofarm/windows-smoke-checklist.md`
- Modify: `README.md`
- Modify: `release/nexus/BUILD-EVIDENCE.md`

- [ ] **Step 1: Define the disposable-save matrix**

The checklist records exact versions and verifies startup, fixture marker, F7 teaching, F8 summon, F9 status, F10 correction, watering, refill, deposit, safe stop, blocked route, target disappearance, save, reload, day change, title return, process cleanup, uninstall, and log secret scan. Harvest remains disabled unless a separate native-semantics section passes.

- [ ] **Step 2: Add an evidence report template**

The Windows script emits JSON containing source commit, game/SMAPI version strings, architecture, package checksum, each check name/status, and redacted diagnostics. Human gameplay checks remain explicitly unsigned until the owner records them.

- [ ] **Step 3: Run complete offline verification**

Run Go race/vet, the full .NET suite, three demos, twenty-run result races, the PowerShell self-tests in the Windows CI job, four-platform package smoke, deterministic sidecar rebuild, and native health smoke.

- [ ] **Step 4: Review and commit**

```bash
git commit -m "docs(echofarm): define Windows release acceptance"
```

- [ ] **Step 5: Push and verify GitHub Actions**

Push the feature branch and require all jobs green for the exact final SHA. Do not create a release or claim real-game certification.

## Deferred production work

Model usage accounting, hard request/token budgets, F9 usage display, and decision-call reduction are specified in the approved production design but will be implemented in a separate plan immediately after this foundation. Native harvest enablement remains a separate game-backed certification task because it cannot be proved without the owner's legal Windows game installation.
