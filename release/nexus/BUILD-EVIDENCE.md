# EchoFarm 0.4.0 Reflective Policy build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 114 tests passed, 0 failed, including capability gates, setup readiness, endpoint conflict detection, diagnostic redaction, model-usage contracts, and F9 usage presentation.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Windows workflow`: 9 self-tests passed for doctor results, package allowlisting, atomic install, `config.json` preservation, safe uninstall, and evidence redaction.
- `Action capability boundary`: the Mod emits explicit capabilities; Go skips disabled candidates; C# rejects disabled harvesting again before mutation. Harvesting defaults off.
- `Model operations`: every generation reserves a durable request ledger row with purpose and safe status; provider token metadata remains explicitly known/unknown; concurrent call limits do not overshoot; exhausted action/intent budgets persist `stop_session`.
- `Model data boundary`: the usage schema contains no prompt, response, API-key, or provider-error-body column. Only bounded error classes are persisted, and no currency estimate is produced.
- `Reflective policy`: failures and player corrections produce bounded, evidence-linked experiences; duplicate evidence is idempotent; matching decisions expose calibrated confidence and validated alternatives.
- `Structured generation`: OpenAI-compatible calls request `json_object` response format; prompts require materially applied experience citations, counterfactual alternatives, and the smallest evidence-supported causal rule.
- `Durable reflection`: the first canonical failed result atomically enqueues one SQLite job; expiring leases prevent simultaneous processing, transient failures return to pending, and restart recovery preserves once-only experience persistence.
- `Concurrent result safety`: an optimistic SQLite compare-and-swap elects exactly one canonical action result; identical retries cannot enqueue duplicate work and conflicting outcomes cannot overwrite it. Sixteen-way result and lease races each produced one winner across twenty stress-test runs.
- `Experience effectiveness`: feedback for materially applied experiences commits atomically with the canonical action result; deterministic attribution separates success, resource contradiction, and neutral world churn. Reads project effective confidence without overwriting semantic confidence, and repeatedly contradicted experience is cooled below the policy threshold.
- `Correction boundary`: F10 capture is represented by a twenty-second state machine, accepts one successful supported action, stays paused for a retry after model failure, and validates save/session/snapshot/target correlation in C# and Go.
- `Cross-process demos`: the core, four-day Continuum, and six-stage reflective demos passed. The reflective demo restarted the Go process twice and proved proactive deposit, corrected-chest selection, and successful-outcome strengthening from persisted SQLite experience.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the rebuilt 0.4.0 macOS arm64 core started in fixture mode and returned HTTP 200 with `{"status":"ok"}` from `/healthz`.
- `Reproducibility`: a fresh second clean cross-build was byte-for-byte identical for all four sidecars after the production-foundation changes.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.4.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 24 MB | `a0f4e6df81983d7d46b9983e96b8d34fb6e3daadb3b9965d8ec88b2aa84570ac` |
| `echofarm-core-macos-arm64` | 21 MB | `71ffb1178b170c165f8bf4e2033501150288269a465edcd13c74c7a452275bde` |
| `echofarm-core-macos-x64` | 25 MB | `b499aa98c6b0f042f9101ffce402613057aeae72cfd5d4196649eadfa707b553` |
| `echofarm-core-windows-x64.exe` | 25 MB | `d29739c15645df59830c53a334c6e8262c3ac446bbce8994df234b3f9d6638a3` |

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

## Machine-readable Windows evidence

`Install-EchoFarm.ps1 -Build` writes `dist/windows/EchoFarm.evidence.json` next to the staged directory, ZIP, and SHA-256 file. It records:

- source commit;
- Windows/x64 target;
- game and SMAPI file versions;
- package SHA-256;
- signed automated checks;
- unsigned `pending` gameplay checks;
- redacted diagnostics only.

The JSON is evidence input, not automatic game certification. Complete [the Windows smoke checklist](../../docs/echofarm/windows-smoke-checklist.md) on a disposable save and sign the pending checks before calling an archive playable.
