# EchoFarm Mac AI Server + Windows Game Runbook

This mode keeps the model credential, SQLite memory, prompts, and diagnostic state on the Mac. Windows runs Stardew Valley, the C# Mod, and a small Go relay only.

```text
Stardew / C# Mod
  -> HTTP 127.0.0.1:18471
  -> echofarm-relay.exe
  -> pinned TLS + Kitex/Thrift
  -> Mac EchoFarm core
  -> Eino + SQLite + configured model
```

## Before starting

- Put both machines on the same trusted private LAN.
- Do not forward port `18472` on the router.
- Allow the Mac server port only on the private LAN firewall profile.
- Rotate any model API key that has previously appeared in chat, shell history, screenshots, or logs.
- A cloud model necessarily receives the minimized semantic gameplay input used for inference. Use a Mac-local OpenAI-compatible model for strict zero-egress operation.

## 1. Pair the Mac server

Find the Mac LAN address, for example `192.168.1.20`, then run from the repository root:

```bash
./scripts/lan/Initialize-EchoFarmLan.sh --host 192.168.1.20
```

The command writes the private key, certificate, and random relay token below `~/Library/Application Support/EchoFarm/LAN`, never inside the repository. Copy the printed `CERT_SHA256` and `LAN_TOKEN` once to the Windows machine, then clear the terminal.

Start the server:

```bash
./scripts/lan/Start-EchoFarmLanServer.sh --address 192.168.1.20:18472
```

When model variables are not already present, the script asks for the OpenAI-compatible base URL, model/deployment name, and hidden API key. It builds the core below the Mac application-data directory and starts the TLS Kitex/Thrift gateway. To verify the transport without paid model calls:

```bash
ECHOFARM_MODEL_MODE=fixture \
  ./scripts/lan/Start-EchoFarmLanServer.sh --address 192.168.1.20:18472
```

## 2. Install and start the Windows relay

Open PowerShell in the cloned repository and use the fingerprint printed on the Mac:

```powershell
.\scripts\windows\Install-EchoFarmRelay.ps1 `
  -UpstreamAddress "192.168.1.20:18472" `
  -CertSha256 "<64-hex-certificate-fingerprint>"
```

This builds `echofarm-relay.exe` into `%LOCALAPPDATA%\EchoFarm\relay`. Its JSON configuration contains only the Mac address, certificate fingerprint, timeout, and concurrency limit.

Start it:

```powershell
.\scripts\windows\Start-EchoFarmRelay.ps1
```

Paste the 64-character LAN token at the secure prompt. The token is passed to the relay process and is not saved to disk. A successful command returns `Ready=True` and `Endpoint=http://127.0.0.1:18471`.

## 3. Configure the Mod

In the installed EchoFarm Mod `config.json`, set:

```json
{
  "CoreUrl": "http://127.0.0.1:18471",
  "ConnectionMode": "relay",
  "AutoStartCore": false,
  "CoreStartupTimeoutSeconds": 15,
  "CommandTimeoutSeconds": 100
}
```

Windows does not need `ModelBaseUrl`, `ModelName`, `DatabasePath`, or `ECHOFARM_MODEL_API_KEY` in relay mode. Start the relay before launching SMAPI.

## 4. Play the demo

Use a disposable save first:

1. Press `F7`, perform a normal routine such as watering, harvesting, chopping, mining, or fishing, then press `F7` again.
2. Wait for the learned confirmation. A real reasoning model may take tens of seconds; the default command timeout is 100 seconds.
3. Press `F8` to summon Echo. Live actions are not transport-retried, so stale decisions cannot execute after a disconnect.
4. Press `F9` to inspect learned traits, recent decisions, model usage, and safe errors.
5. Press `F10`, then perform one successful action to correct Echo's current choice.

The current executable action whitelist covers movement, tool selection, watering/refill, harvesting, depositing, and safe stop. Chopping, mining, mine-floor traversal, and fishing are captured as semantic learning signals but remain learn-only until their game mutations receive the same native safety verification.

## Failure checks

- `invalid certificate`: rerun Mac initialization only when intentionally rotating the identity, then reinstall the Windows relay with the new fingerprint.
- `401`: the relay token does not match; restart the Windows relay and enter the current token.
- `503 upstream_unavailable`: confirm both machines are on the same LAN, the Mac process is running, and the private firewall allows the selected port.
- F9 says `unhealthy_core` in relay mode: `echofarm-relay.exe` is not healthy on Windows loopback.
- Model timeout: increase `ECHOFARM_MODEL_TIMEOUT_SECONDS` on Mac and keep `CommandTimeoutSeconds` on Windows slightly higher.

Rotate the certificate and token with:

```bash
./scripts/lan/Initialize-EchoFarmLan.sh --host 192.168.1.20 --force
```

Then reinstall/restart the Windows relay with the new fingerprint and token.

## Automated proof and acceptance boundary

Run the local cross-process proof with:

```bash
./demo/run-lan-relay-demo.sh
```

It verifies a real TLS handshake, certificate pin, generated Kitex/Thrift RPC calls, teaching, Mac SQLite memory, request-ID correlation, rejection of a wrong token, and metadata-only logs. Passing this proof produces a LAN demo candidate. Final public-release acceptance still requires a Windows Stardew 1.6 + SMAPI 4.1+ run on a disposable save and measured gameplay latency.
