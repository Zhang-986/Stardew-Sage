# EchoFarm 0.3.0 Continuum build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 77 tests passed, 0 failed.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Cross-process demo`: learning, rainy-layout harvest, empty-can refill, and full-inventory deposit recovery all returned the expected structured actions.
- `Four-day Continuum demo`: model revisions advanced from 1 to 3, repeated sunny habits became stable, rainy behavior remained context-scoped, live watering intent was inferred, claimed crops were avoided, and the decision was visible through the memory API.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the built macOS arm64 core started in fixture mode and returned HTTP 200 from `/healthz`.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.3.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 24 MB | `70ede390ee308c7d127b1a8cff355a4375bc05957ac83407f8eefbcd0693790a` |
| `echofarm-core-macos-arm64` | 21 MB | `18b7150b8d135390f37330d5b31cfb2f7710f8659799a5f056284990e3075f9c` |
| `echofarm-core-macos-x64` | 25 MB | `bc2001e791753a3fd002a1b4d14aa37c4b583d911f3a1da2f7e1c753ac2eafaa` |
| `echofarm-core-windows-x64.exe` | 25 MB | `6bd3e831cb6a964c15d03541ec732ad40556e5329e9b5d7477651d253feac053` |

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
