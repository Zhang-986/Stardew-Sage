# EchoFarm 0.4.0 Reflective Policy build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 87 tests passed, 0 failed.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Reflective policy`: failures and player corrections produce bounded, evidence-linked experiences; duplicate evidence is idempotent; matching decisions expose calibrated confidence and validated alternatives.
- `Correction boundary`: F10 capture is represented by a twenty-second state machine, accepts one successful supported action, stays paused for a retry after model failure, and validates save/session/snapshot/target correlation in C# and Go.
- `Cross-process demos`: the core, four-day Continuum, and five-stage reflective demos passed. The reflective demo restarted the Go process twice and proved proactive deposit plus corrected-chest selection from persisted SQLite experience.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the 0.4.0 macOS arm64 core started in fixture mode and returned HTTP 200 from `/healthz`.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.4.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 24 MB | `218e7a1219b5ed83a38422cc607df7c82bc34fe068914884e28a5f9449cefe6a` |
| `echofarm-core-macos-arm64` | 21 MB | `028ceb6d6da6f333191684b0d68a153c6b84c54dcede12b170b5e17bf11cff10` |
| `echofarm-core-macos-x64` | 25 MB | `be1272d66e174de7a9f4647e3855006ac458b29179bedfb91bff38658f6cba15` |
| `echofarm-core-windows-x64.exe` | 25 MB | `4c5b556f945a46e66d98dd5dc01dca5d94401c180f740a5dcf76f6383dd4f798` |

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
