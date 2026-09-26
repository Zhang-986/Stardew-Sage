[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$UpstreamAddress,
    [Parameter(Mandatory = $true)]
    [string]$CertSha256,
    [string]$DataRoot = (Join-Path $env:LOCALAPPDATA 'EchoFarm\relay'),
    [string]$RelayExecutablePath,
    [ValidateRange(1, 300)]
    [int]$TimeoutSeconds = 100,
    [ValidateRange(1, 8)]
    [int]$MaxInFlight = 2
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

if ($UpstreamAddress -notmatch '^.+:\d+$') {
    throw 'UpstreamAddress must be the private Mac host and port, for example 192.168.1.20:18472.'
}
$fingerprint = ($CertSha256 -replace ':', '').ToLowerInvariant()
if ($fingerprint -notmatch '^[0-9a-f]{64}$') {
    throw 'CertSha256 must contain exactly 64 hexadecimal characters.'
}

$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
$coreRoot = Join-Path $repoRoot 'echofarm-core'
$DataRoot = [IO.Path]::GetFullPath($DataRoot)
$target = Join-Path $DataRoot 'echofarm-relay.exe'
$configPath = Join-Path $DataRoot 'relay.json'
New-Item -ItemType Directory -Path $DataRoot -Force | Out-Null

if ([string]::IsNullOrWhiteSpace($RelayExecutablePath)) {
    Push-Location $coreRoot
    try {
        & go build -trimpath -ldflags '-s -w -buildid=' -o $target ./cmd/echofarm-relay
        if ($LASTEXITCODE -ne 0) {
            throw "Go failed to build echofarm-relay.exe (exit $LASTEXITCODE)."
        }
    }
    finally {
        Pop-Location
    }
}
else {
    $source = [IO.Path]::GetFullPath($RelayExecutablePath)
    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
        throw "Relay executable was not found: $source"
    }
    Copy-Item -LiteralPath $source -Destination $target -Force
}

[ordered]@{
    upstreamAddress = $UpstreamAddress
    certSha256      = $fingerprint
    timeoutSeconds  = $TimeoutSeconds
    maxInFlight     = $MaxInFlight
} | ConvertTo-Json | Set-Content -LiteralPath $configPath -Encoding UTF8

[pscustomobject]@{
    Ready          = $true
    ExecutablePath = $target
    ConfigPath     = $configPath
    TokenStored    = $false
    Next           = "Run Start-EchoFarmRelay.ps1 and enter the one-time LAN token."
}
