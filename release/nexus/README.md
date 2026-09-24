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

Before the Nexus page exists:

```bash
./scripts/package-nexus.sh \
  --version 0.3.0 \
  --game-path "/absolute/path/to/Stardew Valley"
```

After Nexus assigns the mod ID, rebuild so SMAPI update checks work:

```bash
./scripts/package-nexus.sh \
  --version 0.3.0 \
  --game-path "/absolute/path/to/Stardew Valley" \
  --nexus-mod-id 12345
```

The upload files and `SHA256SUMS.txt` are written to `dist/nexus/0.3.0/`. Upload `windows-x64` as the main file and the other three archives as optional platform files. Keep their names unchanged.

Use [description.md](description.md) for the page body, [changelog.md](changelog.md) for the first file changelog, [permissions.md](permissions.md) for permissions and disclosure, [screenshot-checklist.md](screenshot-checklist.md) for the gallery, and [upload-checklist.md](upload-checklist.md) for the final publication pass.
