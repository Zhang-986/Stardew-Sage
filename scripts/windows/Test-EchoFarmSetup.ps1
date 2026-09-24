[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$modulePath = Join-Path $PSScriptRoot 'EchoFarm.Setup.psm1'
Import-Module $modulePath -Force

$script:Passed = 0
$script:Failed = 0

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) {
        throw $Message
    }
}

function Invoke-Test {
    param([string]$Name, [scriptblock]$Body)
    try {
        & $Body
        $script:Passed++
        Write-Host "PASS $Name"
    }
    catch {
        $script:Failed++
        Write-Host "FAIL $Name`: $($_.Exception.Message)" -ForegroundColor Red
    }
}

function New-FixtureGame {
    param([string]$Root)
    $gamePath = Join-Path $Root 'steamapps/common/Stardew Valley'
    $modsPath = Join-Path $gamePath 'Mods'
    New-Item -ItemType Directory -Path $modsPath -Force | Out-Null
    Set-Content -LiteralPath (Join-Path $gamePath 'Stardew Valley.dll') -Value 'fixture-game'
    Set-Content -LiteralPath (Join-Path $gamePath 'StardewModdingAPI.dll') -Value 'fixture-smapi'
    return $gamePath
}

function New-FixturePackage {
    param([string]$Root, [string]$ModText = 'new-mod')
    $package = Join-Path $Root 'EchoFarm'
    New-Item -ItemType Directory -Path (Join-Path $package 'core') -Force | Out-Null
    Set-Content -LiteralPath (Join-Path $package 'EchoFarm.Mod.dll') -Value $ModText
    Set-Content -LiteralPath (Join-Path $package 'EchoFarm.Bridge.dll') -Value 'bridge'
    Set-Content -LiteralPath (Join-Path $package 'manifest.json') -Value '{"Name":"EchoFarm"}'
    Set-Content -LiteralPath (Join-Path $package 'LICENSE') -Value 'license'
    Set-Content -LiteralPath (Join-Path $package 'README.md') -Value 'readme'
    Set-Content -LiteralPath (Join-Path $package 'core/echofarm-core.exe') -Value 'core'
    return $package
}

$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("echofarm-setup-tests-" + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null
try {
    Invoke-Test 'discovers a standard Steam game path' {
        $steamRoot = Join-Path $tempRoot 'steam'
        $expected = New-FixtureGame $steamRoot
        $actual = Resolve-EchoFarmGamePath -SteamRoots @($steamRoot)
        Assert-True ($actual -eq $expected) "Expected '$expected', got '$actual'."
    }

    Invoke-Test 'reports a ready fixture installation' {
        $gamePath = New-FixtureGame (Join-Path $tempRoot 'ready')
        $result = Test-EchoFarmPrerequisites -GamePath $gamePath -SkipHostValidation -SkipToolchainValidation
        Assert-True $result.Ready "Expected prerequisites to pass: $($result.Issues | ConvertTo-Json -Compress)"
        Assert-True ($result.Issues.Count -eq 0) 'Ready result unexpectedly contained issues.'
    }

    Invoke-Test 'uses stable issue codes for missing prerequisites' {
        $gamePath = Join-Path $tempRoot 'missing/Stardew Valley'
        New-Item -ItemType Directory -Path (Join-Path $gamePath 'Mods') -Force | Out-Null
        $result = Test-EchoFarmPrerequisites -GamePath $gamePath -SkipHostValidation -SkipToolchainValidation
        $codes = @($result.Issues | ForEach-Object Code)
        Assert-True ($codes -contains 'missing_game_dll') 'missing_game_dll was not reported.'
        Assert-True ($codes -contains 'missing_smapi') 'missing_smapi was not reported.'
    }

    Invoke-Test 'rejects files outside the package allowlist' {
        $package = New-FixturePackage (Join-Path $tempRoot 'allowlist')
        $valid = Test-EchoFarmPackage -PackagePath $package
        Assert-True $valid.Ready "Valid package was rejected: $($valid.Issues | ConvertTo-Json -Compress)"
        Set-Content -LiteralPath (Join-Path $package 'debug.pdb') -Value 'symbols'
        $invalid = Test-EchoFarmPackage -PackagePath $package
        Assert-True (-not $invalid.Ready) 'Package with a PDB unexpectedly passed.'
        Assert-True (@($invalid.Issues | ForEach-Object Code) -contains 'forbidden_package_file') 'Forbidden-file issue code was not reported.'
    }

    Invoke-Test 'build staging accepts fixture binaries and emits checksum' {
        $root = Join-Path $tempRoot 'build'
        $gamePath = New-FixtureGame (Join-Path $root 'game')
        $modBuild = Join-Path $root 'mod-build'
        New-Item -ItemType Directory -Path $modBuild -Force | Out-Null
        Set-Content -LiteralPath (Join-Path $modBuild 'EchoFarm.Mod.dll') -Value 'mod'
        Set-Content -LiteralPath (Join-Path $modBuild 'EchoFarm.Bridge.dll') -Value 'bridge'
        $core = Join-Path $root 'echofarm-core.exe'
        Set-Content -LiteralPath $core -Value 'core'
        $output = Join-Path $root 'out/EchoFarm'

        $result = Build-EchoFarmPackage -GamePath $gamePath -OutputPath $output -ModBuildDirectory $modBuild -CoreExecutablePath $core -SkipCompilation

        Assert-True $result.Ready "Fixture package build failed: $($result.Issues | ConvertTo-Json -Compress)"
        Assert-True (Test-Path -LiteralPath (Join-Path $output 'core/echofarm-core.exe')) 'Core executable was not staged.'
        Assert-True (Test-Path -LiteralPath "$output.sha256") 'Package checksum file was not emitted.'
    }

    Invoke-Test 'install atomically replaces program files and preserves config' {
        $root = Join-Path $tempRoot 'install'
        $gamePath = New-FixtureGame (Join-Path $root 'game')
        $package = New-FixturePackage (Join-Path $root 'package')
        $installed = Join-Path $gamePath 'Mods/EchoFarm'
        New-Item -ItemType Directory -Path $installed -Force | Out-Null
        Set-Content -LiteralPath (Join-Path $installed 'config.json') -Value '{"keep":true}'
        Set-Content -LiteralPath (Join-Path $installed 'obsolete.dll') -Value 'old'

        Install-EchoFarm -GamePath $gamePath -PackagePath $package | Out-Null

        Assert-True ((Get-Content -LiteralPath (Join-Path $installed 'config.json') -Raw).Trim() -eq '{"keep":true}') 'Existing config.json was not preserved.'
        Assert-True ((Get-Content -LiteralPath (Join-Path $installed 'EchoFarm.Mod.dll') -Raw).Trim() -eq 'new-mod') 'New Mod DLL was not installed.'
        Assert-True (-not (Test-Path -LiteralPath (Join-Path $installed 'obsolete.dll'))) 'Obsolete program file survived replacement.'
        Assert-True (-not (Test-Path -LiteralPath (Join-Path $gamePath 'Mods/.EchoFarm.installing'))) 'Staging directory was not cleaned up.'
    }

    Invoke-Test 'uninstall preserves local data unless explicitly requested' {
        $root = Join-Path $tempRoot 'uninstall'
        $gamePath = New-FixtureGame (Join-Path $root 'game')
        $installed = Join-Path $gamePath 'Mods/EchoFarm'
        New-Item -ItemType Directory -Path $installed -Force | Out-Null
        Set-Content -LiteralPath (Join-Path $installed 'manifest.json') -Value '{}'
        $localData = Join-Path $root 'local-data/EchoFarm'
        New-Item -ItemType Directory -Path $localData -Force | Out-Null
        Set-Content -LiteralPath (Join-Path $localData 'echofarm.db') -Value 'memory'

        Uninstall-EchoFarm -GamePath $gamePath -LocalDataPath $localData | Out-Null
        Assert-True (-not (Test-Path -LiteralPath $installed)) 'Installed Mod directory was not removed.'
        Assert-True (Test-Path -LiteralPath (Join-Path $localData 'echofarm.db')) 'Local AI memory was removed without consent.'

        New-Item -ItemType Directory -Path $installed -Force | Out-Null
        Uninstall-EchoFarm -GamePath $gamePath -LocalDataPath $localData -DeleteLocalData | Out-Null
        Assert-True (-not (Test-Path -LiteralPath $localData)) 'Explicit local-data deletion did not run.'
    }

    Invoke-Test 'command entry point supports doctor JSON output' {
        $gamePath = New-FixtureGame (Join-Path $tempRoot 'entry')
        $entryScript = Join-Path $PSScriptRoot 'Install-EchoFarm.ps1'

        $json = (& $entryScript -Doctor -GamePath $gamePath -AsJson -SkipHostValidation -SkipToolchainValidation | Out-String)
        $result = $json | ConvertFrom-Json

        Assert-True $result.Ready 'Doctor entry point did not return a ready result.'
        Assert-True ($result.GamePath -eq $gamePath) 'Doctor entry point returned the wrong game path.'
    }
}
finally {
    Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
}

if ($script:Failed -gt 0) {
    throw "$script:Failed EchoFarm setup test(s) failed; $script:Passed passed."
}
Write-Host "EchoFarm setup tests passed: $script:Passed"
