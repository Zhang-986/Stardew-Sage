# Nexus upload checklist

## Before creating the page

- [ ] Build against a legal Stardew Valley 1.6 installation.
- [ ] Complete the disposable-save smoke checklist on every platform marked supported.
- [ ] Confirm the author name, support URL, and MIT license text.
- [ ] Capture the required screenshots with no credentials or personal paths visible.
- [ ] Select the Stardew Valley game page and a gameplay/utility category.
- [ ] Apply the platform's current generative-AI disclosure/tag because AI inference is a core runtime feature.

## Page fields

- **Name:** EchoFarm — Teach an AI by Playing
- **Version:** 0.4.0
- **Summary:** Train a persistent AI echo that learns habits, reflects on failures, accepts in-game corrections, and changes later decisions.
- **Requirements:** Stardew Valley 1.6, SMAPI 4.1+, an OpenAI-compatible endpoint for real learning
- **Source:** https://github.com/Zhang-986/Stardew-Sage
- **Description:** copy `description.md`
- **Changelog:** copy `changelog.md`
- **Permissions:** copy `permissions.md`

## Files

- [ ] Upload `EchoFarm-0.4.0-windows-x64.zip` as the main file.
- [ ] Upload Linux x64, macOS x64, and macOS arm64 as optional files.
- [ ] Verify each Nexus download hash against `SHA256SUMS.txt` after upload.
- [ ] Do not upload the fixture-DLL smoke-test output.

## After Nexus assigns the mod ID

- [ ] Re-run `scripts/package-nexus.sh` with `--nexus-mod-id <id>`.
- [ ] Confirm staged `manifest.json` contains exactly `"UpdateKeys": ["Nexus:<id>"]`.
- [ ] Replace the initial files with the rebuilt archives if the first upload had no update key.
- [ ] Install the downloaded archive into a clean `Mods` directory and repeat the live smoke test.
