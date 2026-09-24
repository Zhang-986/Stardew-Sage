# EchoFarm Windows x64 production-candidate smoke checklist

This checklist is the final gate for Stardew Valley 1.6 + SMAPI 4.1+ single-player. Use a copied disposable save. Do not enable experimental harvesting and do not test first on the only copy of a personal save.

## Evidence header

Record these values in `dist/windows/EchoFarm.evidence.json` before gameplay:

- [ ] source commit is the exact tested commit;
- [ ] Windows edition/build and x64 architecture;
- [ ] Stardew Valley version;
- [ ] SMAPI version;
- [ ] package SHA-256 matches `EchoFarm.sha256`;
- [ ] `EnableExperimentalHarvest` is `false`.

The build script fills machine-verifiable fields and leaves gameplay checks `pending` and unsigned. The tester changes a check to `passed` only after observing it in game.

## Install and first launch

- [ ] `Install-EchoFarm.ps1 -Doctor` reports `Ready: True`.
- [ ] `-Build` creates `EchoFarm.zip`, `EchoFarm.sha256`, and `EchoFarm.evidence.json`.
- [ ] `-Install` produces `Mods/EchoFarm/manifest.json` without an extra nested directory.
- [ ] A pre-existing `Mods/EchoFarm/config.json` survives reinstall byte-for-byte.
- [ ] SMAPI loads EchoFarm without red errors.
- [ ] The bundled `echofarm-core.exe` starts and `/healthz` reports healthy.
- [ ] F9 opens before any routine has been taught.
- [ ] F9 shows `DEMO`, Core state, session state, database path, and `Harvest: disabled`.
- [ ] No API key is required or requested in fixture mode.

## Normal gameplay loop

- [ ] F7 starts teaching and the player remains in control.
- [ ] Record watering, watering-can refill, and a chest deposit; F7 completes teaching.
- [ ] F8 summons a visible translucent Echo only after learned memory exists.
- [ ] Echo walks tile-by-tile to a valid target without teleporting the player.
- [ ] Echo waters a dry crop and does not water during rain or storm.
- [ ] An empty Echo watering can causes refill before another watering action.
- [ ] Echo deposits only items accepted by the demonstrated chest.
- [ ] A full chest retains rejected items and causes a bounded safe stop.
- [ ] A blocked route returns `path_blocked`, replans once, and does not loop forever.
- [ ] Removing/changing a target after planning returns `target_changed` and does not mutate another target.
- [ ] A model/fixture timeout stops Echo while the game remains responsive.
- [ ] F10 accepts one successful corrective action and later matching situations use that correction.
- [ ] F9 shows current intent, last decision, confidence, safe alternatives, and evidence-backed memory.

## Lifecycle and persistence

- [ ] Saving during an active Echo action stops safely and the save completes.
- [ ] Reloading the save restores learned memory from the same database.
- [ ] Starting a new day resets transient activity while retaining learned traits.
- [ ] Returning to title cancels work and hides Echo without hanging SMAPI.
- [ ] Exiting the game terminates only the sidecar process owned by this Mod.
- [ ] Reinstall preserves `config.json` and the local SQLite database.
- [ ] Default uninstall removes `Mods/EchoFarm` but preserves `%LOCALAPPDATA%\EchoFarm`.
- [ ] `-DeleteLocalData` removes local memory only after explicit invocation.

## Semantic activity sensors — learn-only

These checks certify observation and AI learning only. They must not enable autonomous chopping, mining, mine traversal, or fishing.

- [ ] During F7 teaching, repeated axe hits followed by one felled tree produce one `chop_tree` episode with wood/item gains.
- [ ] During F7 teaching, a pickaxe hit that removes a rock produces one `break_rock` episode; an unchanged target does not report success.
- [ ] Moving between two mine floors produces one `enter_mine_floor` event with the correct signed floor delta.
- [ ] A completed catch produces `fish_caught` with a positive fish item delta; leaving fishing without a catch produces `fish_escaped` or a bounded cancellation.
- [ ] Saving, changing day, returning to title, or leaving the location clears pending multi-tick activity state.
- [ ] After two matching teaching days, F9 shows stable activity-order/resource/fishing or mine habits without raw evidence IDs.
- [ ] F9 labels chopping, mining, mine floors, and fishing as `learn-only`; none appears in the executable action catalog.

## Real-model opt-in

- [ ] Set `ModelMode` to `openai`, configure an absolute HTTP(S) `ModelBaseUrl` and `ModelName`, and provide `ECHOFARM_MODEL_API_KEY` only in the SMAPI process environment.
- [ ] Missing URL, model name, or key prevents sidecar launch and F9 shows the matching stable issue code.
- [ ] A non-loopback Core URL is rejected before any process starts.
- [ ] A port occupied by a non-Echo service is reported as `unhealthy_core`.
- [ ] F9 no longer shows `DEMO` after a healthy real-model startup.
- [ ] Model failure never falls back silently to fixture decisions.
- [ ] Search the package, SMAPI log, Core log, and evidence JSON for the API key; there must be zero matches.

## Harvest certification — separate gate

Do not enable `EnableExperimentalHarvest` for the first candidate. Native harvest certification remains unsigned until an adapter preserves yield, quality, regrowth, experience, sounds/events, special crops, and relevant SMAPI hooks.

- [ ] Native harvest semantics implemented.
- [ ] Special/regrowing/multi-yield crop matrix passed.
- [ ] SMAPI compatibility hooks observed.
- [ ] Harvest capability deliberately enabled for a new candidate.

## Sign-off

- Tester:
- UTC timestamp:
- Evidence JSON path:
- Disposable save identifier:
- Result: `pending` / `passed` / `failed`
- Notes and screenshots:
