# EchoFarm 0.4.0 Reflective Policy build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 109 tests passed, 0 failed, including capability gates, setup readiness, endpoint conflict detection, and diagnostic redaction.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Windows workflow`: 9 self-tests passed for doctor results, package allowlisting, atomic install, `config.json` preservation, safe uninstall, and evidence redaction.
- `Action capability boundary`: the Mod emits explicit capabilities; Go skips disabled candidates; C# rejects disabled harvesting again before mutation. Harvesting defaults off.
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
| `echofarm-core-linux-x64` | 24 MB | `81879f5626dc1fef7561336f9f063a220507d2138cb1bc2abc89aa37d7d22c2f` |
| `echofarm-core-macos-arm64` | 21 MB | `0738be3e00ed00437f4a5956f689bca3c2e7b2ea6ff8ed9725bb0d440fa9c5a9` |
| `echofarm-core-macos-x64` | 25 MB | `37f13437e9cc74117d1db7957aa50e9a7c4f57a17c57e1c91ac25f6fec2ad370` |
| `echofarm-core-windows-x64.exe` | 25 MB | `b58be55c904dbf6964ac1c4141b91af31c1987fd52b0f3d4fa79668b6d47ac3c` |

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
