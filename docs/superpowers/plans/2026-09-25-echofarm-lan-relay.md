# EchoFarm Secure LAN Relay Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep the model credential, player memory, and AI processing on the Mac while a Windows-only Go relay securely carries the existing C# Mod protocol over a private LAN.

**Architecture:** The C# Mod continues to use loopback HTTP. A new `echofarm-relay.exe` forwards an exact route allowlist over certificate-pinned HTTPS with a separate bearer token to an explicitly enabled LAN mode in `echofarm-core`. Mac-side middleware authenticates before parsing, adds correlated metadata-only access logs, and never exposes provider or host internals.

**Tech Stack:** Go 1.24 standard library HTTP/TLS/crypto packages, existing Go/Eino core and SQLite store, C#/.NET 6 bridge, PowerShell 5.1 with Windows DPAPI, zsh/OpenSSL setup scripts, GitHub Actions.

---

## File Structure

- `echofarm-core/internal/lanserver/middleware.go`: bearer authentication, request-ID validation/generation, and redacted access logging for LAN mode.
- `echofarm-core/internal/lanserver/middleware_test.go`: auth-before-body, correlation, and secret-redaction tests.
- `echofarm-core/cmd/echofarm/main.go`: strict LAN/TLS configuration and HTTPS server selection.
- `echofarm-core/cmd/echofarm/main_test.go`: loopback defaults and fail-closed LAN configuration tests.
- `echofarm-core/internal/relay/config.go`: relay environment parsing and validation.
- `echofarm-core/internal/relay/proxy.go`: route allowlist, bounded proxying, header filtering, TLS pinning, concurrency, and metadata logging.
- `echofarm-core/internal/relay/config_test.go`: invalid address, URL, token, fingerprint, limit, and timeout tests.
- `echofarm-core/internal/relay/proxy_test.go`: forwarding, rejection, TLS pin, timeout, size, concurrency, and log-redaction tests.
- `echofarm-core/cmd/echofarm-relay/main.go`: Windows-friendly relay lifecycle and loopback HTTP server.
- `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`: explicit `ConnectionMode` and LAN-friendly command timeout.
- `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`: pass relay mode into readiness and client creation.
- `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/SetupReadiness.cs`: skip provider-key requirements only when a healthy local relay is selected.
- `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmClientFactory.cs`: configurable command timeout while preserving the short read timeout.
- matching C# tests under `stardew-echo-mod/tests/EchoFarm.Bridge.Tests`.
- `scripts/lan/Initialize-EchoFarmLan.sh`: generate a local certificate/key, token, and fingerprint outside the repository.
- `scripts/lan/Start-EchoFarmLanServer.sh`: hidden model-key prompt and Mac TLS-core startup.
- `scripts/windows/Install-EchoFarmRelay.ps1`: build/copy the relay and save the LAN token with current-user DPAPI.
- `scripts/windows/Start-EchoFarmRelay.ps1`: decrypt the token into child-process memory and start/probe the relay.
- `scripts/windows/Test-EchoFarmRelaySetup.ps1`: deterministic PowerShell setup tests without real secrets.
- `.github/workflows/ci.yml`: Windows relay build and cross-process relay smoke.
- `docs/echofarm/lan-relay-runbook.md`: exact Mac and Windows commands, firewall boundary, rotation, and troubleshooting.

### Task 1: Fail-closed Mac LAN configuration

**Files:**
- Modify: `echofarm-core/cmd/echofarm/main.go`
- Modify: `echofarm-core/cmd/echofarm/main_test.go`

- [ ] **Step 1: Write failing configuration tests**

Add table tests proving non-loopback addresses fail without LAN opt-in, LAN mode fails without a token/certificate/key, tokens that are not exactly 64 hexadecimal characters fail, and a complete LAN configuration produces `AllowLAN=true` with the supplied TLS paths.

```go
func TestLoadConfigRejectsIncompleteLANMode(t *testing.T) {
    base := map[string]string{
        "ECHOFARM_MODEL_MODE": "fixture",
        "ECHOFARM_ADDRESS": "0.0.0.0:18472",
    }
    _, err := loadConfig(mapLookup(base))
    if err == nil || !strings.Contains(err.Error(), "ECHOFARM_ALLOW_LAN") {
        t.Fatalf("loadConfig() error = %v", err)
    }
}

func TestLoadConfigAcceptsCompleteLANMode(t *testing.T) {
    cfg, err := loadConfig(mapLookup(map[string]string{
        "ECHOFARM_MODEL_MODE": "fixture",
        "ECHOFARM_ADDRESS": "0.0.0.0:18472",
        "ECHOFARM_ALLOW_LAN": "true",
        "ECHOFARM_LAN_TOKEN": strings.Repeat("a", 64),
        "ECHOFARM_TLS_CERT_FILE": "/private/cert.pem",
        "ECHOFARM_TLS_KEY_FILE": "/private/key.pem",
    }))
    if err != nil || !cfg.AllowLAN { t.Fatalf("config = %+v, err = %v", cfg, err) }
}
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `cd echofarm-core && go test ./cmd/echofarm -run 'TestLoadConfig.*LAN' -count=1`

Expected: FAIL because `AllowLAN` and LAN variables do not exist.

- [ ] **Step 3: Implement the minimal configuration**

Add `AllowLAN`, `LANToken`, `TLSCertFile`, and `TLSKeyFile` to `config`. Parse `ECHOFARM_ALLOW_LAN` strictly with `strconv.ParseBool`. Preserve loopback HTTP as the default; require explicit LAN mode plus all secrets and TLS paths for any non-loopback address.

```go
if !loopback && !result.AllowLAN {
    return config{}, errors.New("non-loopback address requires ECHOFARM_ALLOW_LAN=true")
}
if result.AllowLAN && (!validHexSecret(result.LANToken, 32) || result.TLSCertFile == "" || result.TLSKeyFile == "") {
    return config{}, errors.New("LAN mode requires a strong token and TLS certificate/key")
}
```

- [ ] **Step 4: Run focused and package tests**

Run: `cd echofarm-core && go test ./cmd/echofarm -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/cmd/echofarm/main.go echofarm-core/cmd/echofarm/main_test.go
git commit -m "feat(core): add fail-closed LAN configuration"
```

### Task 2: Authenticate and audit Mac requests

**Files:**
- Create: `echofarm-core/internal/lanserver/middleware.go`
- Create: `echofarm-core/internal/lanserver/middleware_test.go`
- Modify: `echofarm-core/cmd/echofarm/main.go`

- [ ] **Step 1: Write failing middleware tests**

Use `httptest` to prove missing/wrong tokens return 401 without calling the wrapped handler, valid bearer tokens reach it, malformed request IDs are replaced, and captured logs do not contain authorization values or request bodies.

```go
func TestMiddlewareAuthenticatesBeforeReadingBody(t *testing.T) {
    called := false
    next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
    handler, _ := lanserver.New(next, strings.Repeat("a", 64), log.New(io.Discard, "", 0))
    request := httptest.NewRequest(http.MethodPost, "/v1/demonstrations/learn", strings.NewReader(`{"private":"game-data"}`))
    response := httptest.NewRecorder()
    handler.ServeHTTP(response, request)
    if response.Code != http.StatusUnauthorized || called { t.Fatalf("code=%d called=%v", response.Code, called) }
}
```

- [ ] **Step 2: Run the tests and verify RED**

Run: `cd echofarm-core && go test ./internal/lanserver -count=1`

Expected: FAIL because the package is missing.

- [ ] **Step 3: Implement middleware and HTTPS startup**

Implement `New(next http.Handler, token string, logger *log.Logger) (http.Handler, error)`. Hash configured and presented tokens with SHA-256 and compare through `subtle.ConstantTimeCompare`. Accept only `Authorization: Bearer <token>`. Generate a 16-byte random hexadecimal request ID unless `X-EchoFarm-Request-ID` already contains exactly 32 hexadecimal characters. Log only request ID, method, escaped path, status, response bytes, and latency.

In `run`, wrap the existing handler only in LAN mode and call:

```go
if cfg.AllowLAN {
    server.Handler, err = lanserver.New(server.Handler, cfg.LANToken, log.Default())
    if err != nil { return err }
    err = server.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
} else {
    err = server.ListenAndServe()
}
```

- [ ] **Step 4: Run middleware and core tests**

Run: `cd echofarm-core && go test -race ./internal/lanserver ./cmd/echofarm -count=1`

Expected: PASS with no sensitive text in captured logs.

- [ ] **Step 5: Commit**

```bash
git add echofarm-core/internal/lanserver echofarm-core/cmd/echofarm/main.go
git commit -m "feat(core): secure LAN API with TLS and bearer auth"
```

### Task 3: Build the loopback-only Windows Go relay

**Files:**
- Create: `echofarm-core/internal/relay/config.go`
- Create: `echofarm-core/internal/relay/config_test.go`
- Create: `echofarm-core/internal/relay/proxy.go`
- Create: `echofarm-core/internal/relay/proxy_test.go`
- Create: `echofarm-core/cmd/echofarm-relay/main.go`

- [ ] **Step 1: Write failing relay configuration tests**

Define and test this public configuration contract:

```go
type Config struct {
    ListenAddress     string
    UpstreamURL       *url.URL
    LANToken          string
    CertSHA256        [32]byte
    RequestTimeout    time.Duration
    MaxRequestBytes   int64
    MaxInFlight       int
}

func LoadConfig(lookup func(string) (string, bool)) (Config, error)
```

Require `ECHOFARM_RELAY_ADDRESS` to be loopback, `ECHOFARM_RELAY_UPSTREAM` to use HTTPS, a 64-hex-character LAN token representing 32 random bytes, a 64-hex-character certificate fingerprint, positive timeouts/body limits, and `MaxInFlight` from 1 through 8.

- [ ] **Step 2: Run configuration tests and verify RED**

Run: `cd echofarm-core && go test ./internal/relay -run TestLoadConfig -count=1`

Expected: FAIL because relay configuration is missing.

- [ ] **Step 3: Implement configuration, pinned TLS, and proxy tests**

Create tests around `httptest.NewTLSServer` for the exact method/path allowlist, token insertion, request-ID propagation, header stripping, 2 MiB body rejection, concurrency rejection, upstream timeout, certificate mismatch, and bounded 503 errors. The allowlist is:

```go
var allowed = map[string]map[string]bool{
    http.MethodGet: {
        "/healthz": true,
        "/v1/player-model": true,
        "/v1/skills/morning-farm-routine": true,
        "/v1/echo/memory": true,
        "/v1/model-usage": true,
    },
    http.MethodPost: {
        "/v1/demonstrations/learn": true,
        "/v1/echo/next-action": true,
        "/v1/echo/action-result": true,
        "/v1/echo/corrections": true,
    },
}
```

- [ ] **Step 4: Run proxy tests and verify RED**

Run: `cd echofarm-core && go test ./internal/relay -count=1`

Expected: FAIL because the proxy implementation is missing.

- [ ] **Step 5: Implement the relay and executable**

Build an `http.Client` that disables redirects and uses a TLS config which verifies the exact SHA-256 leaf-certificate pin plus the certificate validity window. Read each request through `http.MaxBytesReader`, copy only `Content-Type` and `Accept`, add bearer auth and request ID, acquire a bounded semaphore, and return only upstream status/content type/body. Generate local errors as fixed JSON codes without including upstream error strings.

`cmd/echofarm-relay/main.go` loads configuration, handles SIGINT/SIGTERM, applies server timeouts, and logs only the loopback address and upstream host.

- [ ] **Step 6: Run relay tests and cross-build Windows**

Run:

```bash
cd echofarm-core
go test -race ./internal/relay ./cmd/echofarm-relay -count=1
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags='-s -w -buildid=' -o /tmp/echofarm-relay.exe ./cmd/echofarm-relay
```

Expected: tests pass and `/tmp/echofarm-relay.exe` exists.

- [ ] **Step 7: Commit**

```bash
git add echofarm-core/internal/relay echofarm-core/cmd/echofarm-relay
git commit -m "feat(relay): add pinned Windows LAN proxy"
```

### Task 4: Add explicit relay mode to the C# bridge

**Files:**
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModConfig.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Mod/ModEntry.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/SetupReadiness.cs`
- Modify: `stardew-echo-mod/src/EchoFarm.Bridge/Transport/EchoFarmClientFactory.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Runtime/SetupReadinessTests.cs`
- Modify: `stardew-echo-mod/tests/EchoFarm.Bridge.Tests/Transport/EchoFarmClientFactoryTests.cs`

- [ ] **Step 1: Write failing relay-mode readiness tests**

Add `ConnectionMode` to setup input. Prove `local` retains existing provider checks, `relay` requires `AutoStartCore=false`, skips Windows provider/key requirements, and still requires a loopback `CoreUrl`. Add factory tests proving a configured 100-second command timeout and unchanged 2-second read timeout.

```csharp
SetupReadinessReport report = SetupReadiness.Evaluate(Input(
    connectionMode: "relay",
    modelBaseUrl: null,
    modelName: null,
    apiKeyPresent: false,
    autoStartCore: false,
    endpointStatus: CoreEndpointStatus.Healthy));
Assert.Equal(SetupIssueCodes.Ready, report.Code);
```

- [ ] **Step 2: Run C# tests and verify RED**

Run: `./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln --filter 'SetupReadinessTests|EchoFarmClientFactoryTests'`

Expected: FAIL because relay mode and configurable timeout are missing.

- [ ] **Step 3: Implement relay mode**

Add these Mod settings:

```csharp
public string ConnectionMode { get; set; } = "local";
public int CommandTimeoutSeconds { get; set; } = 100;
```

Validate `ConnectionMode` as `local` or `relay`. In relay mode, require `AutoStartCore=false`, continue requiring loopback `CoreUrl`, and do not require model provider fields or `ECHOFARM_MODEL_API_KEY` on Windows. Change `EchoFarmClientFactory.Create` to accept the validated command timeout.

- [ ] **Step 4: Run the full bridge suite**

Run: `./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln`

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add stardew-echo-mod/src stardew-echo-mod/tests
git commit -m "feat(mod): support secure LAN relay mode"
```

### Task 5: Add secret-safe Mac and Windows launchers

**Files:**
- Create: `scripts/lan/Initialize-EchoFarmLan.sh`
- Create: `scripts/lan/Start-EchoFarmLanServer.sh`
- Create: `scripts/windows/Install-EchoFarmRelay.ps1`
- Create: `scripts/windows/Start-EchoFarmRelay.ps1`
- Create: `scripts/windows/Test-EchoFarmRelaySetup.ps1`
- Modify: `.gitignore`

- [ ] **Step 1: Write failing setup-script tests**

The PowerShell test runs against a temporary local-data root and proves configuration validation, DPAPI-backed token export on Windows, no plaintext token in generated JSON, deterministic relay build paths, and cleanup. Shell checks prove Mac output paths are outside the repository, generated files receive owner-only permissions, and no model credential is accepted as a command-line flag.

- [ ] **Step 2: Run tests and verify RED**

Run on macOS/Linux: `bash -n scripts/lan/Initialize-EchoFarmLan.sh scripts/lan/Start-EchoFarmLanServer.sh`

Run on Windows CI: `pwsh -NoProfile -File scripts/windows/Test-EchoFarmRelaySetup.ps1`

Expected before implementation: commands fail because the scripts are absent.

- [ ] **Step 3: Implement Mac initialization and startup**

`Initialize-EchoFarmLan.sh --host <LAN-IP>` uses `openssl req -x509` with an IP subject alternative name, writes certificate/key/token below `${HOME}/Library/Application Support/EchoFarm/LAN`, applies `chmod 600`, and prints the public certificate SHA-256 fingerprint plus the newly generated LAN token once to the controlling terminal for manual Windows pairing. It never places either value in shell history or repository files. `Start-EchoFarmLanServer.sh` reads the LAN token from that protected file, prompts for the provider key with terminal echo disabled, exports the required variables to the core child only, and cleans the shell variables on exit.

- [ ] **Step 4: Implement Windows DPAPI configuration and startup**

`Install-EchoFarmRelay.ps1` builds `echofarm-relay.exe`, validates an HTTPS server URL and 64-character fingerprint, prompts with `Read-Host -AsSecureString`, and saves the secure string through `Export-Clixml` under `%LOCALAPPDATA%\EchoFarm\relay`. It writes only URL/fingerprint/timeouts to JSON. `Start-EchoFarmRelay.ps1` imports the secure string, converts it into the child environment, zeroes the BSTR in `finally`, launches the relay, and requires localhost `/healthz` success.

- [ ] **Step 5: Run syntax/setup tests and scan generated outputs**

Run:

```bash
bash -n scripts/lan/Initialize-EchoFarmLan.sh scripts/lan/Start-EchoFarmLanServer.sh
pwsh -NoProfile -File scripts/windows/Test-EchoFarmRelaySetup.ps1
```

Expected: all checks pass; fixture output contains no plaintext secret.

- [ ] **Step 6: Commit**

```bash
git add .gitignore scripts/lan scripts/windows
git commit -m "feat(ops): add secret-safe LAN pairing scripts"
```

### Task 6: Prove cross-process behavior and package the relay

**Files:**
- Create: `demo/run-lan-relay-demo.sh`
- Modify: `.github/workflows/ci.yml`
- Create: `docs/echofarm/lan-relay-runbook.md`
- Modify: `README.md`
- Modify: `stardew-echo-mod/README.md`

- [ ] **Step 1: Write the cross-process smoke**

The script creates a temporary certificate and tokens, starts the fixture core in LAN/TLS mode, starts the relay on loopback, submits `demo/fixtures/morning-teaching.json`, reads memory through the relay, tests a wrong token directly against the Mac listener, and removes all temporary certificates, keys, logs, and databases in a trap.

- [ ] **Step 2: Run the smoke and verify RED before wiring CI**

Run: `./demo/run-lan-relay-demo.sh`

Expected before the script is complete: FAIL at the first missing LAN/relay behavior.

- [ ] **Step 3: Complete smoke assertions and CI**

Require HTTP 200 for authenticated teaching and memory, HTTP 401 for a wrong token, and a relay log containing the same 32-character request ID as the Mac log. Reject the run if either log contains the generated token or fixture body. Add the smoke to the Linux verify job and add this Windows build command:

```powershell
go build -trimpath -ldflags "-s -w -buildid=" -o "$env:RUNNER_TEMP\echofarm-relay.exe" ./cmd/echofarm-relay
```

Upload the Windows relay binary as a CI artifact without configuration, certificates, tokens, databases, or logs.

- [ ] **Step 4: Write the runbook**

Document Mac initialization/start, Windows install/start, private-firewall rules, `ConnectionMode=relay`, `AutoStartCore=false`, F7/F8/F9 acceptance, request-ID correlation, credential rotation, certificate re-pairing, and the explicit prohibition on router port forwarding.

- [ ] **Step 5: Run full verification**

Run:

```bash
cd echofarm-core && go test -race -count=1 ./... && go vet ./...
cd .. && ./.tools/dotnet/dotnet test stardew-echo-mod/EchoFarm.sln
./demo/run-core-demo.sh
./demo/run-continuum-demo.sh
./demo/run-reflective-demo.sh
./demo/run-activity-learning-demo.sh
./demo/run-lan-relay-demo.sh
./scripts/tests/package-nexus-smoke.sh
git diff --check
```

Expected: every command exits 0.

- [ ] **Step 6: Perform a repository secret scan**

Scan tracked files and the release/artifact allowlists for private keys, bearer tokens, model endpoints, model deployment IDs, API-key assignments, databases, and logs. The scan must report no generated secrets or user-specific provider configuration before commit or push.

- [ ] **Step 7: Commit and push**

```bash
git add .github/workflows/ci.yml demo/run-lan-relay-demo.sh docs/echofarm/lan-relay-runbook.md README.md stardew-echo-mod/README.md
git commit -m "feat(echofarm): deliver secure Mac to Windows relay"
git push origin codex/living-valley-director
```
