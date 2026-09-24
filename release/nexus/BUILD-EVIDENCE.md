# EchoFarm 0.2.0 build evidence

Generated on 2026-09-24 from branch `codex/living-valley-director`.

## Verified locally

- `.NET bridge`: 70 tests passed, 0 failed.
- `Go core`: all packages passed with `-race -count=1`; `go vet ./...` passed.
- `Cross-process demo`: learning, rainy-layout harvest, empty-can refill, and full-inventory deposit recovery all returned the expected structured actions.
- `Nexus packaging smoke`: four platform ZIPs were built with fixture Mod DLLs, validated, and deleted after the test. These smoke archives are not playable release files.
- `Native sidecar smoke`: the built macOS arm64 core started in fixture mode and returned HTTP 200 from `/healthz`.

## Cross-compiled sidecars

Generated files are ignored by Git under `artifacts/echofarm-core-0.2.0/`.

| File | Approximate size | SHA-256 |
| --- | ---: | --- |
| `echofarm-core-linux-x64` | 23 MB | `6f8964c6dbf68e361efa9c4a1649dd443405ad9ef78e206596cf87be90454496` |
| `echofarm-core-macos-arm64` | 20 MB | `b9ebd95cc9db5a796979e18def0098ce7202ecb3c0ea1e0e02bb598777a2a7f0` |
| `echofarm-core-macos-x64` | 24 MB | `19f56c86445f6cacbbcf87ef600fbb09febf51bcd1289242cc3bbd362c59fc00` |
| `echofarm-core-windows-x64.exe` | 24 MB | `49fdb765b8bbd7d750de54252183ee17c9902e62db47003f76410d4f01769e92` |

## External publication gates

The current machine has no discoverable `Stardew Valley.dll`. The Release build stops in `Pathoschild.Stardew.ModBuildConfig` with `The mod build package can't find your game folder`; therefore no playable Mod DLL or public Nexus archive has been claimed.

No `NEXUS_API_KEY` is present and no Nexus mod ID has been assigned. No external page or file was created.

When both prerequisites are available, continue with:

```bash
./scripts/package-nexus.sh \
  --version 0.2.0 \
  --game-path "/absolute/path/to/Stardew Valley" \
  --nexus-mod-id 12345
```

Then perform the disposable-save smoke checklist and upload only the validated files under `dist/nexus/0.2.0/`.
