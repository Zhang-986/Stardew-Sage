# EchoFarm Nexus Release Design

## Goal

Ship EchoFarm as a normal SMAPI mod whose AI brain is a bundled Go/Eino sidecar, with reproducible platform archives and enough release material to upload to Nexus Mods once a legal Stardew installation and Nexus account are available.

## Chosen distribution model

EchoFarm will publish one archive per supported runtime:

- `EchoFarm-<version>-windows-x64.zip`
- `EchoFarm-<version>-linux-x64.zip`
- `EchoFarm-<version>-macos-x64.zip`
- `EchoFarm-<version>-macos-arm64.zip`

Each archive has one top-level `EchoFarm/` directory containing the SMAPI manifest, `EchoFarm.Mod.dll`, `EchoFarm.Bridge.dll`, the platform's `echofarm-core` executable, the license, and a short player README. This follows the normal “extract one folder into `Mods`” Stardew installation experience while avoiding a universal archive containing three unusable native binaries.

A separate manually started server was rejected because it makes the demo feel like two unrelated projects. A hosted server was rejected because save-derived gameplay data and provider credentials should remain local. A universal archive remains possible later, but platform-specific files are clearer on Nexus and substantially smaller.

## Runtime lifecycle

The SMAPI adapter owns only a process supervisor, never AI logic:

1. Validate that `CoreUrl` is loopback-only.
2. On `GameLaunched`, probe `/healthz`.
3. If a healthy service already exists, reuse it and mark it externally managed.
4. Otherwise, when `AutoStartCore` is enabled, resolve the bundled executable for the current OS and architecture, start it without a shell, and poll health until the configured timeout.
5. Stream sidecar stdout and stderr into the SMAPI monitor without exposing secrets.
6. On process exit, terminate only a child process started by this mod. Never terminate a pre-existing service.

The process supervisor and executable resolver live in the game-independent bridge assembly behind small interfaces, so ownership, timeout, and no-double-start rules are covered by unit tests. The Mod project only supplies filesystem paths, configuration, and SMAPI logging.

## Configuration and secrets

`config.json` exposes:

- `AutoStartCore` (default `true`);
- `CoreUrl` (default `http://127.0.0.1:18471`);
- `CoreExecutablePath` (empty means the bundled executable);
- `CoreStartupTimeoutSeconds` (default `10`);
- `ModelMode` (`openai` by default, `fixture` only for an explicit offline demo);
- `ModelBaseUrl` and `ModelName` (optional overrides);
- `DatabasePath` (empty means an OS-local application-data path).

The API key is never accepted in `config.json`. The child inherits `ECHOFARM_MODEL_API_KEY` from the launch environment. Existing `ECHOFARM_*` values win over empty Mod settings, which keeps local providers and shell-based secrets usable.

## Build and packaging

`scripts/package-nexus.sh` will:

1. require a semantic version and a legal Stardew `GamePath`;
2. run Go tests, .NET bridge tests, and the SMAPI Release build;
3. cross-compile the Go command with `CGO_ENABLED=0` for the four supported runtimes;
4. stage only runtime files under a single `EchoFarm/` directory;
5. validate `manifest.json`, required files, archive root, and forbidden secret/database patterns;
6. create deterministic ZIP files and SHA-256 checksums under `dist/nexus/<version>/`.

The script may accept `--mod-build-dir` for CI/package tests using fixture DLLs, but a public release is valid only when the DLLs came from a successful build against a legal game installation. No game assemblies are copied into any archive.

## Nexus publication material

The repository will contain a release title, short summary, full description, installation steps, configuration guide, permissions statement, AI-use disclosure, changelog template, screenshot checklist, and upload checklist. The initial archive keeps `UpdateKeys` empty; after Nexus assigns a mod ID, packaging accepts that ID and injects `Nexus:<id>` into staged manifests.

Actual Nexus upload is an external state change and requires the owner's authenticated account, game page, mod ID, screenshots, and confirmation of the final archive. Automation will stop at a validated upload-ready directory when those are unavailable.

## Failure handling

- A healthy external core is never replaced or killed.
- A missing bundled executable produces one actionable SMAPI error and leaves Stardew playable.
- A child that exits before health succeeds reports its exit code and leaves Echo disabled.
- A startup timeout kills only the owned child and leaves Echo disabled.
- Missing OpenAI settings fail closed; the Mod never silently switches to fixture behavior.
- Save data, databases, logs, environment files, API keys, and game assemblies are excluded from archives.

## Verification

- Unit tests cover executable resolution, external-service reuse, owned-process startup, timeout cleanup, and shutdown ownership.
- Existing Go race tests, Go vet, .NET tests, and Go↔C# integration tests remain green.
- A packaging smoke test uses fixture DLLs to prove archive shape and secret exclusion without pretending to be a playable build.
- A real release requires a successful SMAPI build and the documented disposable-save smoke run on each claimed platform.
