# EchoFarm 0.4.0 Reflective Policy build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 87 tests passed, 0 failed.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Reflective policy`: failures and player corrections produce bounded, evidence-linked experiences; duplicate evidence is idempotent; matching decisions expose calibrated confidence and validated alternatives.
- `Structured generation`: OpenAI-compatible calls request `json_object` response format; prompts require materially applied experience citations, counterfactual alternatives, and the smallest evidence-supported causal rule.
- `Durable reflection`: the first canonical failed result atomically enqueues one SQLite job; expiring leases prevent simultaneous processing, transient failures return to pending, and restart recovery preserves once-only experience persistence.
- `Concurrent result safety`: an optimistic SQLite compare-and-swap elects exactly one canonical action result; identical retries cannot enqueue duplicate work and conflicting outcomes cannot overwrite it. Sixteen-way result and lease races each produced one winner across twenty stress-test runs.
- `Correction boundary`: F10 capture is represented by a twenty-second state machine, accepts one successful supported action, stays paused for a retry after model failure, and validates save/session/snapshot/target correlation in C# and Go.
- `Cross-process demos`: the core, four-day Continuum, and five-stage reflective demos passed. The reflective demo restarted the Go process twice and proved proactive deposit plus corrected-chest selection from persisted SQLite experience.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the rebuilt 0.4.0 macOS arm64 core started in fixture mode and returned HTTP 200 with `{"status":"ok"}` from `/healthz`.
- `Reproducibility`: a second clean cross-build was byte-for-byte identical for all four sidecars.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.4.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 24 MB | `3765a4420633433884f22d0935fcc96f104e2f136321b0e781b3242331868afe` |
| `echofarm-core-macos-arm64` | 21 MB | `c1d7d41615aa053b1a773b250c4f45471dfc4b3ceb7cbcafb2544489d41d7f48` |
| `echofarm-core-macos-x64` | 25 MB | `fc1671dceeffb41acee4a502a4ed1a166f216f2b1b77132686b1a3f2a97e76dc` |
| `echofarm-core-windows-x64.exe` | 25 MB | `e74b4113278827af21e81ab80b589cafa8764247c018defb231c99f367e36956` |

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
