# EchoFarm Nexus release kit

This directory contains the copy and checklists for publishing EchoFarm on the Stardew Valley section of Nexus Mods.

## Current publication gate

The repository can build and validate all four Go/Eino sidecars without Stardew Valley. A playable archive additionally requires:

1. a legal Stardew Valley 1.6 installation with SMAPI 4.1 or newer;
2. a successful Release build of `EchoFarm.Mod.dll` against that installation;
3. a Nexus Mods account and manual confirmation of the final listing;
4. the Nexus mod ID assigned after creating the page.

No source-only or fixture-DLL archive should be uploaded as a playable file.

## Build upload-ready files

### Windows owner workflow (recommended for the first production candidate)

Open PowerShell in the repository and replace the example game path with the directory that contains both `Stardew Valley.dll` and `StardewModdingAPI.dll`:

```powershell
$game = "C:\Program Files (x86)\Steam\steamapps\common\Stardew Valley"

.\scripts\windows\Install-EchoFarm.ps1 -Doctor -GamePath $game
.\scripts\windows\Install-EchoFarm.ps1 -Build -GamePath $game
.\scripts\windows\Install-EchoFarm.ps1 -Install -GamePath $game -PackagePath .\dist\windows\EchoFarm
```

The build command compiles the real SMAPI Mod against the local legal game installation, cross-builds the Windows x64 Go sidecar, validates an exact file allowlist, and writes `dist\windows\EchoFarm.zip`, its SHA-256 file, and `EchoFarm.evidence.json`. Installation stages into `Mods\.EchoFarm.installing`, preserves an existing `Mods\EchoFarm\config.json`, and then replaces the program directory.

Uninstall removes only the Mod program directory by default, preserving the learned SQLite memory under local application data:

```powershell
.\scripts\windows\Install-EchoFarm.ps1 -Uninstall -GamePath $game
```

Deleting learned memory requires the explicit `-DeleteLocalData` switch.

### Cross-platform release workflow

Before the Nexus page exists:

```bash
./scripts/package-nexus.sh \
  --version 0.4.0 \
  --game-path "/absolute/path/to/Stardew Valley"
```

After Nexus assigns the mod ID, rebuild so SMAPI update checks work:

```bash
./scripts/package-nexus.sh \
  --version 0.4.0 \
  --game-path "/absolute/path/to/Stardew Valley" \
  --nexus-mod-id 12345
```

The upload files and `SHA256SUMS.txt` are written to `dist/nexus/0.4.0/`. Upload `windows-x64` as the main file and the other three archives as optional platform files. Keep their names unchanged.

Use [description.md](description.md) for the page body, [changelog.md](changelog.md) for the first file changelog, [permissions.md](permissions.md) for permissions and disclosure, [screenshot-checklist.md](screenshot-checklist.md) for the gallery, and [upload-checklist.md](upload-checklist.md) for the final publication pass.
