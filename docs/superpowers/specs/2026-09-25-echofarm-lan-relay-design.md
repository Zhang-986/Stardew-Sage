# EchoFarm LAN Relay Design

**Date:** 2026-09-25

## Goal

Run the AI brain, player memory, model credential, and diagnostic state on the owner's Mac while a Windows Stardew Valley installation sends game events and receives validated Echo actions through a small Go relay. The Windows machine must not receive the model API key, Mac filesystem paths, database contents, provider errors, or unrestricted access to the Mac service.

## Non-goals

- Exposing EchoFarm to the public internet.
- Remote shell, file browsing, log download, or arbitrary HTTP proxying.
- Storing the provider API key in the repository, Mod configuration, relay configuration, logs, or SQLite.
- Offline replay of live action requests. A stale action must never be executed after connectivity returns.
- Replacing the existing C# gameplay capture and action executor.

## Considered Approaches

### Direct C# to Mac

The Mod could call the Mac URL directly. This uses fewer processes, but it couples TLS, authentication, retry behavior, and network diagnostics to the game Mod and weakens the existing loopback-only boundary.

### Windows Go relay to Mac server — selected

The Mod continues to call `127.0.0.1`. A purpose-built relay adds authentication, TLS certificate pinning, route filtering, timeouts, and correlated metadata logs before forwarding to the Mac. This keeps gameplay code unchanged and leaves all provider and memory state on the Mac.

### Full core on Windows

This is the existing packaged topology. It is convenient for ordinary players but does not provide centralized Mac-side observation and tuning.

## Architecture

```text
Stardew Valley / SMAPI
        |
        | HTTP on Windows loopback only
        v
echofarm-relay.exe : 127.0.0.1:18471
        |
        | certificate-pinned TLS + Kitex/Thrift RPC over private LAN
        v
echofarm-core on Mac : 0.0.0.0:18472
        |                         |
        |                         +--> local SQLite memory and usage ledger
        +--> configured model provider
```

The relay is transport-only. It does not interpret gameplay, cache model responses, store request bodies, or possess the provider credential. The Mac core remains the sole source of player-model, policy, and model-usage state.

## Mac Server Mode

Loopback remains the default. A non-loopback bind is rejected unless all LAN requirements are explicitly configured:

- `ECHOFARM_ALLOW_LAN=true`
- `ECHOFARM_LAN_TOKEN` as exactly 64 hexadecimal characters generated from 32 random bytes
- `ECHOFARM_TLS_CERT_FILE`
- `ECHOFARM_TLS_KEY_FILE`
- an `ECHOFARM_ADDRESS` whose port is explicitly chosen for the LAN service

The server uses CloudWeGo Kitex with generated Thrift contracts over TLS for every LAN request. Nine typed RPC methods mirror the fixed HTTP contract; there is no generic proxy method. It authenticates the relay token using a constant-time comparison before parsing request bodies. `Health` also requires authentication so unauthenticated devices cannot fingerprint the service.

The server exposes only the existing EchoFarm API routes. It never exposes filesystem, database, configuration, environment, diagnostics-download, or generic proxy endpoints. Existing response redaction remains in force: provider response bodies and internal errors are never returned to the relay.

## Windows Relay

`cmd/echofarm-relay` builds as a static `windows/amd64` executable and:

- listens only on `127.0.0.1:18471`;
- accepts only the exact EchoFarm route and HTTP-method allowlist;
- rejects request bodies larger than 2 MiB;
- maps each accepted route to one generated Kitex/Thrift method;
- adds the dedicated LAN token only to the typed RPC request;
- validates the Mac TLS certificate against a configured SHA-256 fingerprint;
- applies bounded connect, read, and total request timeouts;
- forwards status and JSON response bodies without exposing upstream transport details;
- limits concurrent upstream requests to prevent accidental model-call storms;
- logs only request ID, route, status, byte counts, and latency.

The relay returns a locally generated request ID in `X-EchoFarm-Request-ID`. The Mac server preserves that ID in its sanitized access log, allowing one Windows request to be correlated without logging gameplay payloads or credentials.

The relay does not maintain a durable queue. Teaching and correction calls receive one bounded transport retry because their IDs make them idempotent. Live next-action and action-result calls are never replayed because their snapshots become stale.

## Secret Handling and Pairing

There are two unrelated credentials:

1. The model API key exists only in the Mac server process environment. It is never transmitted to Windows.
2. The LAN relay token authorizes only the fixed EchoFarm HTTP API. It cannot access the provider or other Mac resources.

All persistent player data remains on the Mac. When a cloud model is selected, the minimum semantic gameplay payload required for inference is necessarily sent from the Mac to that configured provider; host paths, database contents, logs, credentials, and unrelated Mac data are never included. A strict zero-egress deployment must use an OpenAI-compatible model running locally on the Mac instead of a cloud API.

Mac setup creates a self-signed server certificate and a random relay token outside the repository under the user's local application-data directory with owner-only permissions. The initialization command shows the LAN token once on its controlling terminal for manual pairing with Windows; it never places the token in command history. Generated keys, certificates, tokens, databases, logs, and `.env` files are excluded from Git packaging.

Windows setup stores only the non-secret upstream address, certificate fingerprint, and timeout settings under `%LOCALAPPDATA%\EchoFarm\relay`. The launcher requests the token as a `SecureString` on every start, passes it through the relay child process environment, clears the launcher environment immediately, and never writes the plaintext token to disk.

Because the design uses TLS plus exact leaf-certificate pinning, another LAN device cannot read game payloads or impersonate the Mac merely by learning its IP address. The server port must be allowed only on the private Windows/Mac LAN profile and must not be forwarded by the router.

## Data Flow

### Teaching

1. The C# Mod records semantic events and posts one bounded demonstration to localhost.
2. The relay maps the request to a typed Thrift call and forwards it over pinned TLS to the Mac.
3. The core validates the contract, invokes the model under configured budgets, and stores the resulting model and skill in Mac SQLite.
4. The validated learning response returns through the relay to the Mod.

### Echo action

1. The Mod sends the current world snapshot to localhost.
2. The relay forwards it once; it never retries the POST automatically.
3. The Mac core chooses and validates a high-level action.
4. The C# bridge performs its existing second safety check before mutating the game.

### Debugging

The Windows and Mac logs share a request ID. Operational logs show transport timing and status only. The Mac SQLite database retains the existing structured demonstrations, decisions, player memory, and model-usage ledger; it does not retain prompts, model responses, provider error bodies, or credentials.

## Failure Behavior

- Invalid or missing LAN token: HTTP 401 without parsing the body.
- Unpinned or changed Mac certificate: relay rejects the connection.
- Route outside the allowlist: HTTP 404 from the relay without contacting the Mac.
- Oversized body: HTTP 413 without contacting the Mac.
- Mac unavailable or request timeout: HTTP 503 with a bounded local error body.
- Model unavailable or budget exhausted: existing redacted core error passes through.
- Any transport failure during an Echo session: the Mod stops Echo safely while Stardew Valley continues.

## Packaging and Operation

The repository produces `echofarm-relay-windows-amd64.exe` separately from the normal bundled core. A Windows PowerShell setup command configures the upstream address and certificate pin without storing the token; the start command prompts for the token and verifies its authenticated upstream health check. The Mod uses `CoreUrl=http://127.0.0.1:18471`, `ConnectionMode=relay`, and `AutoStartCore=false`; the relay is started before SMAPI.

The Mac start command loads the model key and LAN token without echo, starts the TLS core on port `18472`, and prints only the listening address and certificate fingerprint. Stopping the command removes credentials from the process environment.

## Verification

Automated tests must prove:

- non-loopback core binding fails closed without explicit LAN mode, TLS files, and a strong token;
- missing or incorrect tokens receive 401 before request parsing;
- the relay maps every supported route to the correct generated Thrift method and rejects unsupported routes;
- arbitrary HTTP headers cannot cross the typed RPC boundary;
- certificate pin mismatch fails closed;
- payload limits, timeouts, and concurrency bounds work;
- logs contain correlation metadata but no token, authorization header, provider key, prompt, or response body;
- a cross-process fixture test completes `C# contract -> Windows relay binary -> Mac core -> SQLite`;
- `GOOS=windows GOARCH=amd64` produces the relay executable;
- repository and release-package secret scans find no generated credential or local database.

Manual acceptance on the private LAN must prove:

1. Windows localhost relay reports authenticated Mac health.
2. F7 teaching reaches the Mac and appears in the Mac memory view.
3. F8 action selection returns to Windows and retains existing safety validation.
4. Disconnecting Wi-Fi or stopping the Mac core stops Echo without blocking or corrupting the game.
5. A wrong token and wrong certificate fingerprint both fail closed.

## Acceptance Boundary

Passing automated tests produces a LAN-demo candidate, not a public production release. Public readiness still requires a Windows Stardew/SMAPI run on a disposable save, measured end-to-end latency, and confirmation that the private-network firewall rule is scoped correctly.
