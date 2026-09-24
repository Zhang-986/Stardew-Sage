[CmdletBinding()]
param(
    [switch]$Doctor,
    [switch]$Build,
    [switch]$Install,
    [switch]$Uninstall,
    [string]$GamePath,
    [string]$PackagePath,
    [string]$OutputPath,
    [string]$LocalDataPath,
    [switch]$DeleteLocalData,
    [switch]$AsJson,
    [switch]$SkipHostValidation,
    [switch]$SkipToolchainValidation
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSScriptRoot 'EchoFarm.Setup.psm1') -Force

$selected = @($Doctor, $Build, $Install, $Uninstall | Where-Object { $_ }).Count
if ($selected -ne 1) {
    throw 'Choose exactly one operation: -Doctor, -Build, -Install, or -Uninstall.'
}

if ($Doctor) {
    $result = Test-EchoFarmPrerequisites `
        -GamePath $GamePath `
        -SkipHostValidation:$SkipHostValidation `
        -SkipToolchainValidation:$SkipToolchainValidation
}
elseif ($Build) {
    if ([string]::IsNullOrWhiteSpace($GamePath)) {
        throw '-Build requires -GamePath.'
    }
    $result = Build-EchoFarmPackage -GamePath $GamePath -OutputPath $OutputPath
}
elseif ($Install) {
    if ([string]::IsNullOrWhiteSpace($GamePath) -or [string]::IsNullOrWhiteSpace($PackagePath)) {
        throw '-Install requires -GamePath and -PackagePath.'
    }
    $result = Install-EchoFarm -GamePath $GamePath -PackagePath $PackagePath
}
else {
    if ([string]::IsNullOrWhiteSpace($GamePath)) {
        throw '-Uninstall requires -GamePath.'
    }
    $result = Uninstall-EchoFarm `
        -GamePath $GamePath `
        -LocalDataPath $LocalDataPath `
        -DeleteLocalData:$DeleteLocalData
}

if ($AsJson) {
    $result | ConvertTo-Json -Depth 8
}
else {
    $result | Format-List | Out-Host
    if ($result.PSObject.Properties.Name -contains 'Issues') {
        foreach ($issue in $result.Issues) {
            Write-Host "[$($issue.Code)] $($issue.Message)"
            Write-Host "  Fix: $($issue.Correction)"
        }
    }
}

if ($result.PSObject.Properties.Name -contains 'Ready' -and -not $result.Ready) {
    throw 'EchoFarm setup checks failed. Apply the listed fixes and rerun the command.'
}
