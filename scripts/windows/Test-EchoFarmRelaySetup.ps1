[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) { throw $Message }
}

$tempRoot = Join-Path ([IO.Path]::GetTempPath()) ('echofarm-relay-test-' + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
try {
    $fixture = Join-Path $tempRoot 'fixture-relay.exe'
    Set-Content -LiteralPath $fixture -Value 'fixture executable' -Encoding ASCII
    $install = Join-Path $PSScriptRoot 'Install-EchoFarmRelay.ps1'
    $start = Join-Path $PSScriptRoot 'Start-EchoFarmRelay.ps1'
    $secret = 'a' * 64

    foreach ($scriptPath in @($install, $start)) {
        $tokens = $null
        $parseErrors = $null
        [void][Management.Automation.Language.Parser]::ParseFile($scriptPath, [ref]$tokens, [ref]$parseErrors)
        Assert-True ($parseErrors.Count -eq 0) "PowerShell syntax errors in $scriptPath`: $($parseErrors.Message -join '; ')"
    }

    $result = & $install `
        -UpstreamAddress '192.168.1.20:18472' `
        -CertSha256 ('b' * 64) `
        -DataRoot $tempRoot `
        -RelayExecutablePath $fixture

    Assert-True $result.Ready 'Relay installer did not report ready.'
    Assert-True (-not $result.TokenStored) 'Relay installer claimed that it stored the token.'
    Assert-True (Test-Path -LiteralPath $result.ExecutablePath) 'Relay executable was not installed.'
    $raw = Get-Content -LiteralPath $result.ConfigPath -Raw
    $config = $raw | ConvertFrom-Json
    Assert-True ($config.upstreamAddress -eq '192.168.1.20:18472') 'Upstream address was not saved.'
    Assert-True ($config.certSha256 -eq ('b' * 64)) 'Certificate fingerprint was not saved.'
    Assert-True (-not $raw.Contains($secret)) 'Configuration contains token-like secret material.'

    $startSource = Get-Content -LiteralPath $start -Raw
    Assert-True ($startSource.Contains("Read-Host 'LAN relay token (not stored)' -AsSecureString")) 'Launcher does not securely prompt for the token.'
    Assert-True ($startSource.Contains('Remove-Item Env:ECHOFARM_RELAY_TOKEN')) 'Launcher does not clear its token environment variable.'
    Assert-True (-not $startSource.Contains('ECHOFARM_MODEL_API_KEY')) 'Windows launcher handles model credentials.'
    Write-Host 'PASS EchoFarm relay setup stores no plaintext token.'
}
finally {
    Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
}
