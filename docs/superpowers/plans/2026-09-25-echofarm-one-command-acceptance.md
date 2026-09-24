# EchoFarm One-Command Windows Acceptance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give a Windows owner one command that builds the real SMAPI Mod, packages and installs it, proves the bundled Go/Eino sidecar is healthy, and emits an honest machine-readable handoff for the remaining in-game checks.

**Architecture:** Extend the existing PowerShell setup module instead of adding another installer. A bounded core-health probe owns temporary loopback startup and shutdown; a candidate orchestrator composes the existing build and atomic install functions, validates the installed tree, and writes a separate acceptance report whose public-release state remains false until native gameplay evidence is signed.

**Tech Stack:** PowerShell 5.1+, Go sidecar HTTP `/healthz`, JSON evidence, GitHub Actions Windows runner.

---

### Task 1: Add the candidate acceptance contract

**Files:**
- Modify: `scripts/windows/Test-EchoFarmSetup.ps1`
- Modify: `scripts/windows/EchoFarm.Setup.psm1`

- [x] **Step 1: Write the failing orchestration test**

Add a fixture test that calls `New-EchoFarmCandidate` with fixture Mod/Core files and an injected successful health probe. Assert that the package is installed, `ReadyForDisposableSave` is true, `ReadyForPublicRelease` is false, the native gameplay stage is `pending`, and `EchoFarm.acceptance.json` exists without secret-bearing content.

- [x] **Step 2: Verify RED**

Run:

```bash
.tools/pwsh/pwsh -NoLogo -NoProfile -File scripts/windows/Test-EchoFarmSetup.ps1
```

Expected: FAIL because the candidate orchestration function does not exist.

- [x] **Step 3: Implement the bounded health probe**

Add `Test-EchoFarmCoreHealth` to `EchoFarm.Setup.psm1`. It must choose a free loopback port, launch only the supplied sidecar with fixture-mode environment variables through `System.Diagnostics.ProcessStartInfo`, poll `/healthz` until it receives `{ "status": "ok" }`, terminate only the process it owns, delete its temporary database, and return stable issue codes instead of leaking process output.

- [x] **Step 4: Implement candidate orchestration and evidence**

Add `New-EchoFarmCandidate` to compose `Build-EchoFarmPackage`, `Install-EchoFarm`, `Test-EchoFarmPackage` against the installed directory, and `Test-EchoFarmCoreHealth`. Write `EchoFarm.acceptance.json` with automated stages `build`, `package`, `install`, and `sidecar_health` passed, manual stages `smapi_launch` and `semantic_activity_gameplay` pending, `readyForDisposableSave: true`, and `readyForPublicRelease: false`. Export both new functions.

- [x] **Step 5: Verify GREEN and commit**

Run the PowerShell suite and require all tests to pass, then commit:

```bash
git add scripts/windows/EchoFarm.Setup.psm1 scripts/windows/Test-EchoFarmSetup.ps1
git commit -m "feat(release): add Windows candidate acceptance gate"
```

### Task 2: Expose one command and require health verification in CI

**Files:**
- Modify: `scripts/windows/Install-EchoFarm.ps1`
- Modify: `.github/workflows/ci.yml`
- Modify: `README.md`
- Modify: `release/nexus/README.md`
- Modify: `release/nexus/PLAYER_README.md`
- Modify: `docs/echofarm/windows-smoke-checklist.md`

- [ ] **Step 1: Write the failing entry-point assertion**

Extend `Test-EchoFarmSetup.ps1` to parse `Install-EchoFarm.ps1` and assert that `-Prepare` is one of the mutually exclusive operations and dispatches to `Prepare-EchoFarmCandidate` without a `PackagePath` requirement.

- [ ] **Step 2: Verify RED**

Run the PowerShell suite. Expected: FAIL because the entry point does not expose `-Prepare`.

- [ ] **Step 3: Add `-Prepare` and CI health coverage**

Add the switch to the entry point and operation-count guard. Dispatch it with `GamePath` and optional `OutputPath`. Replace the duplicated Windows workflow health-smoke body with a call to `Test-EchoFarmCoreHealth` against the freshly built executable and fail unless `Ready` is true.

- [ ] **Step 4: Document the exact player path**

Make this the primary owner command in all three readmes:

```powershell
.\scripts\windows\Install-EchoFarm.ps1 -Prepare -GamePath $game
```

Document the resulting acceptance JSON and state explicitly that `readyForDisposableSave` is an automated handoff, while `readyForPublicRelease` remains false until the SMAPI launch and semantic activity checklist are signed on a real disposable save.

- [ ] **Step 5: Run full verification and commit**

Run the PowerShell suite, full .NET Release suite, Go race/vet, all four cross-process demos, Nexus package smoke, tracked-source secret scan, and `git diff --check`. Commit and push the exact SHA, then require Linux `verify` and Windows `windows-sidecar` to pass.
