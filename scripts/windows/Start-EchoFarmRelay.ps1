[CmdletBinding()]
param(
    [string]$DataRoot = (Join-Path $env:LOCALAPPDATA 'EchoFarm\relay'),
    [Security.SecureString]$Token,
    [ValidateRange(1, 60)]
    [int]$HealthTimeoutSeconds = 15
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$DataRoot = [IO.Path]::GetFullPath($DataRoot)
$executable = Join-Path $DataRoot 'echofarm-relay.exe'
$configPath = Join-Path $DataRoot 'relay.json'
$pidPath = Join-Path $DataRoot 'relay.pid'
if (-not (Test-Path -LiteralPath $executable -PathType Leaf) -or -not (Test-Path -LiteralPath $configPath -PathType Leaf)) {
    throw 'Relay is not installed. Run Install-EchoFarmRelay.ps1 first.'
}

if (Test-Path -LiteralPath $pidPath) {
    $existingPid = (Get-Content -LiteralPath $pidPath -Raw).Trim()
    if ($existingPid -match '^\d+$' -and (Get-Process -Id ([int]$existingPid) -ErrorAction SilentlyContinue)) {
        throw "EchoFarm relay is already running with PID $existingPid."
    }
    Remove-Item -LiteralPath $pidPath -Force
}

$config = Get-Content -LiteralPath $configPath -Raw | ConvertFrom-Json
if ($null -eq $Token) {
    $Token = Read-Host 'LAN relay token (not stored)' -AsSecureString
}

$bstr = [IntPtr]::Zero
$plainToken = $null
$process = $null
try {
    $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($Token)
    $plainToken = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr)
    if ($plainToken -notmatch '^[0-9a-fA-F]{64}$') {
        throw 'LAN relay token must contain exactly 64 hexadecimal characters.'
    }

    $env:ECHOFARM_RELAY_ADDRESS = '127.0.0.1:18471'
    $env:ECHOFARM_RELAY_UPSTREAM_ADDRESS = [string]$config.upstreamAddress
    $env:ECHOFARM_RELAY_CERT_SHA256 = [string]$config.certSha256
    $env:ECHOFARM_RELAY_TIMEOUT_SECONDS = [string]$config.timeoutSeconds
    $env:ECHOFARM_RELAY_MAX_IN_FLIGHT = [string]$config.maxInFlight
    $env:ECHOFARM_RELAY_TOKEN = $plainToken
    $process = Start-Process -FilePath $executable -WorkingDirectory $DataRoot -PassThru
}
finally {
    Remove-Item Env:ECHOFARM_RELAY_TOKEN -ErrorAction SilentlyContinue
    Remove-Item Env:ECHOFARM_RELAY_ADDRESS -ErrorAction SilentlyContinue
    Remove-Item Env:ECHOFARM_RELAY_UPSTREAM_ADDRESS -ErrorAction SilentlyContinue
    Remove-Item Env:ECHOFARM_RELAY_CERT_SHA256 -ErrorAction SilentlyContinue
    Remove-Item Env:ECHOFARM_RELAY_TIMEOUT_SECONDS -ErrorAction SilentlyContinue
    Remove-Item Env:ECHOFARM_RELAY_MAX_IN_FLIGHT -ErrorAction SilentlyContinue
    $plainToken = $null
    if ($bstr -ne [IntPtr]::Zero) {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
    }
}

if ($null -eq $process) {
    throw 'EchoFarm relay did not start.'
}
Set-Content -LiteralPath $pidPath -Value $process.Id -Encoding ASCII

$deadline = [DateTime]::UtcNow.AddSeconds($HealthTimeoutSeconds)
do {
    if ($process.HasExited) {
        Remove-Item -LiteralPath $pidPath -Force -ErrorAction SilentlyContinue
        throw "EchoFarm relay exited before its health check (exit $($process.ExitCode))."
    }
    try {
        $health = Invoke-RestMethod -Uri 'http://127.0.0.1:18471/healthz' -TimeoutSec 2
        if ($health.status -eq 'ok') {
            [pscustomobject]@{
                Ready    = $true
                ProcessId = $process.Id
                Endpoint = 'http://127.0.0.1:18471'
            }
            return
        }
    }
    catch {
        Start-Sleep -Milliseconds 250
    }
} while ([DateTime]::UtcNow -lt $deadline)

Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath $pidPath -Force -ErrorAction SilentlyContinue
throw 'EchoFarm relay could not reach the authenticated Mac server.'
