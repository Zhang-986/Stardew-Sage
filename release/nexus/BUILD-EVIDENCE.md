# EchoFarm 0.3.0 Continuum build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 78 tests passed, 0 failed.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Concurrency and inference hardening`: learning and action APIs return the canonical SQLite winner after idempotent writes, action results require a newer world snapshot, and conflicting values for one trait/context slot are rejected.
- `SMAPI transport boundary`: command traffic retains a 35-second model budget while save restore and F9 memory polling use an independent 2-second read client.
- `Cross-process demo`: learning, rainy-layout harvest, empty-can refill, and full-inventory deposit recovery all returned the expected structured actions.
- `Four-day Continuum demo`: model revisions advanced from 1 to 3, repeated sunny habits became stable, rainy behavior remained context-scoped, live watering intent was inferred, claimed crops were avoided, and the decision was visible through the memory API.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the built macOS arm64 core started in fixture mode and returned HTTP 200 from `/healthz`.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.3.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 24 MB | `eead699a3a5a3c2fd6abe2936a24efde72c8e1cf08dca0716f3de6a51272d15f` |
| `echofarm-core-macos-arm64` | 21 MB | `4c3584567ecdd18038de4725b44d4439342d2e04985c0cd3357be9097e236c7b` |
| `echofarm-core-macos-x64` | 25 MB | `d8812a65915b94d5791a67a9dc8994adef0f7b5f6db9a05df5a6b1058cf6f88d` |
| `echofarm-core-windows-x64.exe` | 25 MB | `c38ddb1f08b7cb3cb970541707f4b11c0e21741fae8a457c9eb7ed61a4ba434e` |

## External publication gates

The current machine has no discoverable `Stardew Valley.dll`. The Release build stops in `Pathoschild.Stardew.ModBuildConfig` with `The mod build package can't find your game folder`; therefore no playable Mod DLL or public Nexus archive has been claimed.

No `NEXUS_API_KEY` is present and no Nexus mod ID has been assigned. No external page or file was created.

When both prerequisites are available, continue with:

```bash
./scripts/package-nexus.sh \
  --version 0.3.0 \
  --game-path "/absolute/path/to/Stardew Valley" \
  --nexus-mod-id 12345
```

Then perform the disposable-save smoke checklist and upload only the validated files under `dist/nexus/0.3.0/`.
