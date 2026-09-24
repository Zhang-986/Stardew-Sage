# EchoFarm 0.4.0 Reflective Policy build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 125 tests passed, 0 failed, including capability gates, setup readiness, endpoint conflict detection, diagnostic redaction, model-usage contracts, semantic activity classification, and F9 usage presentation.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Windows workflow`: 9 self-tests passed for doctor results, package allowlisting, atomic install, `config.json` preservation, safe uninstall, and evidence redaction.
- `Action capability boundary`: the Mod emits explicit capabilities; Go skips disabled candidates; C# rejects disabled harvesting again before mutation. Harvesting defaults off.
- `Model operations`: every generation reserves a durable request ledger row with purpose and safe status; provider token metadata remains explicitly known/unknown; concurrent call limits do not overshoot; exhausted action/intent budgets persist `stop_session`.
- `Model data boundary`: the usage schema contains no prompt, response, API-key, or provider-error-body column. Only bounded error classes are persisted, and no currency estimate is produced.
- `Semantic activity learning`: version-2 events classify multi-tick tree/rock work, mine-floor transitions, and caught/escaped fishing outcomes; fixture-mode cross-process learning produces stable activity-order, resource, mining, and fishing traits while keeping every new category learn-only.
- `Reflective policy`: failures and player corrections produce bounded, evidence-linked experiences; duplicate evidence is idempotent; matching decisions expose calibrated confidence and validated alternatives.
- `Structured generation`: OpenAI-compatible calls request `json_object` response format; prompts require materially applied experience citations, counterfactual alternatives, and the smallest evidence-supported causal rule.
- `Durable reflection`: the first canonical failed result atomically enqueues one SQLite job; expiring leases prevent simultaneous processing, transient failures return to pending, and restart recovery preserves once-only experience persistence.
- `Concurrent result safety`: an optimistic SQLite compare-and-swap elects exactly one canonical action result; identical retries cannot enqueue duplicate work and conflicting outcomes cannot overwrite it. Sixteen-way result and lease races each produced one winner across twenty stress-test runs.
- `Experience effectiveness`: feedback for materially applied experiences commits atomically with the canonical action result; deterministic attribution separates success, resource contradiction, and neutral world churn. Reads project effective confidence without overwriting semantic confidence, and repeatedly contradicted experience is cooled below the policy threshold.
- `Correction boundary`: F10 capture is represented by a twenty-second state machine, accepts one successful supported action, stays paused for a retry after model failure, and validates save/session/snapshot/target correlation in C# and Go.
- `Cross-process demos`: the core, four-day Continuum, six-stage reflective, and two-day semantic-activity demos passed. The reflective demo restarted the Go process twice; the activity demo persisted stable lifestyle traits without adding unsupported actions to the execution catalog.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the rebuilt 0.4.0 macOS arm64 core started in fixture mode and returned HTTP 200 with `{"status":"ok"}` from `/healthz`.
- `Reproducibility`: a fresh second clean cross-build was byte-for-byte identical for all four sidecars after the production-foundation changes.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.4.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 24 MB | `53b739692461c842a5953f2776064d684bd555d4075c86a6b768121334737ff3` |
| `echofarm-core-macos-arm64` | 21 MB | `57426ddc3f31b5d383837c843983185437efd0217d9e3bc7e6471f610315b8b9` |
| `echofarm-core-macos-x64` | 25 MB | `10d5a523cd86d4b75618e0ace013edd263b727f172f5215715a142873e4d6a08` |
| `echofarm-core-windows-x64.exe` | 25 MB | `64662202cc8628ccfb9cc05d608aaa49a209b55545470b7b99df75dff672cc10` |

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
