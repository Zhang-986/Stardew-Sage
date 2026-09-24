# EchoFarm 0.4.0 Reflective Policy build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 87 tests passed, 0 failed.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Reflective policy`: failures and player corrections produce bounded, evidence-linked experiences; duplicate evidence is idempotent; matching decisions expose calibrated confidence and validated alternatives.
- `Structured generation`: OpenAI-compatible calls request `json_object` response format; prompts require materially applied experience citations, counterfactual alternatives, and the smallest evidence-supported causal rule.
- `Durable reflection`: the first canonical failed result atomically enqueues one SQLite job; expiring leases prevent simultaneous processing, transient failures return to pending, and restart recovery preserves once-only experience persistence.
- `Concurrent result safety`: an optimistic SQLite compare-and-swap elects exactly one canonical action result; identical retries cannot enqueue duplicate work and conflicting outcomes cannot overwrite it. Sixteen-way result and lease races each produced one winner across twenty stress-test runs.
- `Experience effectiveness`: feedback for materially applied experiences commits atomically with the canonical action result; deterministic attribution separates success, resource contradiction, and neutral world churn. Reads project effective confidence without overwriting semantic confidence, and repeatedly contradicted experience is cooled below the policy threshold.
- `Correction boundary`: F10 capture is represented by a twenty-second state machine, accepts one successful supported action, stays paused for a retry after model failure, and validates save/session/snapshot/target correlation in C# and Go.
- `Cross-process demos`: the core, four-day Continuum, and six-stage reflective demos passed. The reflective demo restarted the Go process twice and proved proactive deposit, corrected-chest selection, and successful-outcome strengthening from persisted SQLite experience.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the rebuilt 0.4.0 macOS arm64 core started in fixture mode and returned HTTP 200 with `{"status":"ok"}` from `/healthz`.
- `Reproducibility`: a second clean cross-build was byte-for-byte identical for all four sidecars.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.4.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 24 MB | `376c566625ac122059c52984d4c2abda1fc5f9e2e3a756b34d3dea1875c7e2f8` |
| `echofarm-core-macos-arm64` | 21 MB | `f4dfebf8fbf1f9af6e4e5310bbc12a9c93fc8715907a8c5fd114ef586b9b2e81` |
| `echofarm-core-macos-x64` | 25 MB | `517c1074f87ec41d3ddf7a009be81bd2b613f34fef4c543ce8ceefb6705eb153` |
| `echofarm-core-windows-x64.exe` | 25 MB | `5aa8d5f64a00c734432399949b0fcf88cb6e01d64f04d4f876b334492c98e16b` |

## External publication gates

The current machine has no discoverable Stardew Valley installation, `Stardew Valley.dll`, or `StardewModdingAPI.dll`. The Release build stops in `Pathoschild.Stardew.ModBuildConfig` with `The mod build package can't find your game folder`; therefore no playable Mod DLL or public Nexus archive has been claimed.

No Nexus mod ID has been assigned and no external page or file was created.

When both prerequisites are available, continue with:

```bash
./scripts/package-nexus.sh \
  --version 0.4.0 \
  --game-path "/absolute/path/to/Stardew Valley" \
  --nexus-mod-id 12345
```

Then perform the disposable-save smoke checklist and upload only the validated files under `dist/nexus/0.4.0/`.
