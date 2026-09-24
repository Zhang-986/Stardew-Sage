Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function New-EchoFarmIssue {
    param(
        [Parameter(Mandatory)][string]$Code,
        [Parameter(Mandatory)][string]$Message,
        [Parameter(Mandatory)][string]$Correction
    )
    [pscustomobject]@{
        Code       = $Code
        Message    = $Message
        Correction = $Correction
    }
}

function New-EchoFarmCheckResult {
    param(
        [string]$GamePath,
        [object[]]$Issues
    )
    [pscustomobject]@{
        Ready    = $Issues.Count -eq 0
        GamePath = $GamePath
        Issues   = @($Issues)
    }
}

function Resolve-EchoFarmGamePath {
    [CmdletBinding()]
    param(
        [string]$GamePath,
        [string[]]$SteamRoots
    )

    if (-not [string]::IsNullOrWhiteSpace($GamePath)) {
        return [System.IO.Path]::GetFullPath($GamePath)
    }

    if ($null -eq $SteamRoots -or $SteamRoots.Count -eq 0) {
        $roots = [System.Collections.Generic.List[string]]::new()
        if (-not [string]::IsNullOrWhiteSpace(${env:ProgramFiles(x86)})) {
            $roots.Add((Join-Path ${env:ProgramFiles(x86)} 'Steam'))
        }
        if (-not [string]::IsNullOrWhiteSpace($env:ProgramFiles)) {
            $roots.Add((Join-Path $env:ProgramFiles 'Steam'))
        }
        if (-not [string]::IsNullOrWhiteSpace($env:ProgramW6432)) {
            $roots.Add((Join-Path $env:ProgramW6432 'Steam'))
        }
        $SteamRoots = $roots.ToArray()
    }

    foreach ($root in $SteamRoots) {
        if ([string]::IsNullOrWhiteSpace($root)) {
            continue
        }
        $candidate = Join-Path $root 'steamapps/common/Stardew Valley'
        if (Test-Path -LiteralPath $candidate -PathType Container) {
            return [System.IO.Path]::GetFullPath($candidate)
        }
    }
    return $null
}

function Test-EchoFarmPrerequisites {
    [CmdletBinding()]
    param(
        [string]$GamePath,
        [string[]]$SteamRoots,
        [switch]$SkipHostValidation,
        [switch]$SkipToolchainValidation
    )

    $issues = [System.Collections.Generic.List[object]]::new()
    $resolved = Resolve-EchoFarmGamePath -GamePath $GamePath -SteamRoots $SteamRoots
    if ([string]::IsNullOrWhiteSpace($resolved) -or -not (Test-Path -LiteralPath $resolved -PathType Container)) {
        $issues.Add((New-EchoFarmIssue 'game_not_found' 'Stardew Valley was not found.' 'Pass -GamePath with the Stardew Valley 1.6 installation directory.'))
        return New-EchoFarmCheckResult -GamePath $resolved -Issues $issues.ToArray()
    }

    if (-not $SkipHostValidation) {
        if ([System.Environment]::OSVersion.Platform -ne [System.PlatformID]::Win32NT) {
            $issues.Add((New-EchoFarmIssue 'unsupported_platform' 'This owner workflow targets Windows.' 'Run the script on the Windows x64 game machine.'))
        }
        if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -ne [System.Runtime.InteropServices.Architecture]::X64) {
            $issues.Add((New-EchoFarmIssue 'unsupported_architecture' 'This release candidate targets Windows x64.' 'Use a Windows x64 machine.'))
        }
    }

    if (-not (Test-Path -LiteralPath (Join-Path $resolved 'Stardew Valley.dll') -PathType Leaf)) {
        $issues.Add((New-EchoFarmIssue 'missing_game_dll' 'Stardew Valley.dll is missing.' 'Select the game directory that contains Stardew Valley.dll.'))
    }
    if (-not (Test-Path -LiteralPath (Join-Path $resolved 'StardewModdingAPI.dll') -PathType Leaf)) {
        $issues.Add((New-EchoFarmIssue 'missing_smapi' 'SMAPI is not installed in the selected game directory.' 'Install SMAPI 4.1 or newer, then run the doctor again.'))
    }

    $modsPath = Join-Path $resolved 'Mods'
    if (-not (Test-Path -LiteralPath $modsPath -PathType Container)) {
        $issues.Add((New-EchoFarmIssue 'missing_mods_directory' 'The Stardew Valley Mods directory is missing.' 'Launch SMAPI once or create the Mods directory, then retry.'))
    }
    else {
        $attributes = [System.IO.File]::GetAttributes($modsPath)
        if (($attributes -band [System.IO.FileAttributes]::ReadOnly) -ne 0) {
            $issues.Add((New-EchoFarmIssue 'mods_not_writable' 'The Mods directory is marked read-only.' 'Grant the current user write access to the Mods directory.'))
        }
    }

    if (-not $SkipToolchainValidation) {
        if ($null -eq (Get-Command go -ErrorAction SilentlyContinue)) {
            $issues.Add((New-EchoFarmIssue 'missing_go' 'Go is required for a source build.' 'Install Go 1.24.1 or newer, then reopen PowerShell.'))
        }
        if ($null -eq (Get-Command dotnet -ErrorAction SilentlyContinue)) {
            $issues.Add((New-EchoFarmIssue 'missing_dotnet' '.NET SDK is required for a source build.' 'Install the .NET 8 SDK, then reopen PowerShell.'))
        }
    }

    return New-EchoFarmCheckResult -GamePath $resolved -Issues $issues.ToArray()
}

function Get-EchoFarmRelativeFileList {
    param([Parameter(Mandatory)][string]$PackagePath)
    $root = [System.IO.Path]::GetFullPath($PackagePath).TrimEnd(
        [System.IO.Path]::DirectorySeparatorChar,
        [System.IO.Path]::AltDirectorySeparatorChar
    )
    @(
        Get-ChildItem -LiteralPath $root -File -Recurse -Force | ForEach-Object {
            $_.FullName.Substring($root.Length).TrimStart('\', '/').Replace('\', '/')
        }
    )
}

function Test-EchoFarmPackage {
    [CmdletBinding()]
    param([Parameter(Mandatory)][string]$PackagePath)

    $issues = [System.Collections.Generic.List[object]]::new()
    if (-not (Test-Path -LiteralPath $PackagePath -PathType Container)) {
        $issues.Add((New-EchoFarmIssue 'package_not_found' 'The EchoFarm package directory was not found.' 'Build or extract the EchoFarm package, then pass its path.'))
        return [pscustomobject]@{ Ready = $false; PackagePath = $PackagePath; Issues = @($issues.ToArray()) }
    }

    $required = @(
        'manifest.json',
        'EchoFarm.Mod.dll',
        'EchoFarm.Bridge.dll',
        'LICENSE',
        'README.md',
        'core/echofarm-core.exe'
    )
    $allowed = @{}
    foreach ($path in $required) {
        $allowed[$path] = $true
    }
    $files = @(Get-EchoFarmRelativeFileList -PackagePath $PackagePath)
    foreach ($requiredPath in $required) {
        if ($files -notcontains $requiredPath) {
            $issues.Add((New-EchoFarmIssue 'missing_package_file' "Package file '$requiredPath' is missing." 'Rebuild the Windows package from a clean checkout.'))
        }
    }
    foreach ($file in $files) {
        if (-not $allowed.ContainsKey($file)) {
            $issues.Add((New-EchoFarmIssue 'forbidden_package_file' "Package file '$file' is not allowed." 'Remove secrets, logs, databases, symbols, game files, and other unexpected files.'))
        }
    }

    [pscustomobject]@{
        Ready       = $issues.Count -eq 0
        PackagePath = [System.IO.Path]::GetFullPath($PackagePath)
        Issues      = @($issues.ToArray())
    }
}

function Copy-EchoFarmDirectoryContents {
    param(
        [Parameter(Mandatory)][string]$Source,
        [Parameter(Mandatory)][string]$Destination
    )
    New-Item -ItemType Directory -Path $Destination -Force | Out-Null
    foreach ($item in Get-ChildItem -LiteralPath $Source -Force) {
        Copy-Item -LiteralPath $item.FullName -Destination $Destination -Recurse -Force
    }
}

function Move-EchoFarmDirectoryAtomically {
    param(
        [Parameter(Mandatory)][string]$StagingPath,
        [Parameter(Mandatory)][string]$DestinationPath
    )
    $parent = Split-Path -Parent $DestinationPath
    $backup = Join-Path $parent ('.' + (Split-Path -Leaf $DestinationPath) + '.previous')
    if (Test-Path -LiteralPath $backup) {
        Remove-Item -LiteralPath $backup -Recurse -Force
    }
    $hadExisting = Test-Path -LiteralPath $DestinationPath -PathType Container
    try {
        if ($hadExisting) {
            Move-Item -LiteralPath $DestinationPath -Destination $backup
        }
        Move-Item -LiteralPath $StagingPath -Destination $DestinationPath
        if (Test-Path -LiteralPath $backup) {
            Remove-Item -LiteralPath $backup -Recurse -Force
        }
    }
    catch {
        if (-not (Test-Path -LiteralPath $DestinationPath) -and (Test-Path -LiteralPath $backup)) {
            Move-Item -LiteralPath $backup -Destination $DestinationPath
        }
        throw
    }
}

function Protect-EchoFarmDiagnosticText {
    param([string]$Text)
    if ([string]::IsNullOrWhiteSpace($Text)) {
        return $Text
    }
    $safe = $Text -replace '(?i)(authorization\s*:\s*bearer\s+)[^\s,;]+', '$1[REDACTED]'
    $safe = $safe -replace '(?i)(api[_-]?key\s*[=:]\s*)[^\s,;]+', '$1[REDACTED]'
    return $safe -replace '(?i)sk-[a-z0-9_-]+', '[REDACTED]'
}

function Write-EchoFarmEvidenceReport {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$OutputPath,
        [Parameter(Mandatory)][string]$SourceCommit,
        [Parameter(Mandatory)][string]$GameVersion,
        [Parameter(Mandatory)][string]$SmapiVersion,
        [Parameter(Mandatory)][string]$Architecture,
        [Parameter(Mandatory)][ValidatePattern('^[a-fA-F0-9]{64}$')][string]$PackageSha256,
        [Parameter(Mandatory)][object[]]$Checks,
        [object[]]$Diagnostics = @()
    )

    $safeDiagnostics = @(
        foreach ($diagnostic in $Diagnostics) {
            [pscustomobject][ordered]@{
                code    = [string]$diagnostic.Code
                message = Protect-EchoFarmDiagnosticText ([string]$diagnostic.Message)
            }
        }
    )
    $safeChecks = @(
        foreach ($check in $Checks) {
            [pscustomobject][ordered]@{
                name   = [string]$check.Name
                status = [string]$check.Status
                signed = [bool]$check.Signed
            }
        }
    )
    $report = [pscustomobject][ordered]@{
        schemaVersion = 1
        generatedAtUtc = [DateTime]::UtcNow.ToString('o')
        sourceCommit = $SourceCommit
        platform = 'windows'
        architecture = $Architecture
        gameVersion = $GameVersion
        smapiVersion = $SmapiVersion
        packageSha256 = $PackageSha256.ToLowerInvariant()
        checks = $safeChecks
        diagnostics = $safeDiagnostics
    }
    $parent = Split-Path -Parent ([System.IO.Path]::GetFullPath($OutputPath))
    New-Item -ItemType Directory -Path $parent -Force | Out-Null
    $report | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $OutputPath -Encoding utf8
    return $report
}

function Get-EchoFarmFileVersion {
    param([Parameter(Mandatory)][string]$Path)
    $info = [System.Diagnostics.FileVersionInfo]::GetVersionInfo($Path)
    if (-not [string]::IsNullOrWhiteSpace($info.ProductVersion)) { return $info.ProductVersion }
    if (-not [string]::IsNullOrWhiteSpace($info.FileVersion)) { return $info.FileVersion }
    return 'unknown'
}

function Build-EchoFarmPackage {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$GamePath,
        [string]$OutputPath,
        [string]$ModBuildDirectory,
        [string]$CoreExecutablePath,
        [switch]$SkipCompilation
    )

    $repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
    $prerequisites = Test-EchoFarmPrerequisites -GamePath $GamePath -SkipHostValidation:$SkipCompilation -SkipToolchainValidation:$SkipCompilation
    if (-not $prerequisites.Ready) {
        $codes = ($prerequisites.Issues | ForEach-Object Code) -join ', '
        throw "EchoFarm prerequisites failed: $codes"
    }
    $GamePath = $prerequisites.GamePath
    if ([string]::IsNullOrWhiteSpace($OutputPath)) {
        $OutputPath = Join-Path $repoRoot 'dist/windows/EchoFarm'
    }
    $OutputPath = [System.IO.Path]::GetFullPath($OutputPath)

    $temporaryRoot = Join-Path ([System.IO.Path]::GetTempPath()) ('echofarm-build-' + [Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $temporaryRoot -Force | Out-Null
    try {
        if (-not $SkipCompilation) {
            $dotnet = (Get-Command dotnet -ErrorAction Stop).Source
            & $dotnet restore (Join-Path $repoRoot 'stardew-echo-mod/EchoFarm.sln')
            if ($LASTEXITCODE -ne 0) { throw 'dotnet restore failed.' }
            & $dotnet build (Join-Path $repoRoot 'stardew-echo-mod/src/EchoFarm.Mod/EchoFarm.Mod.csproj') '--configuration' 'Release' "-p:GamePath=$GamePath" '-p:EnableModDeploy=false' '-p:EnableModZip=false'
            if ($LASTEXITCODE -ne 0) { throw 'EchoFarm Mod build failed.' }
            $modDll = Get-ChildItem -LiteralPath (Join-Path $repoRoot 'stardew-echo-mod/src/EchoFarm.Mod/bin/Release') -Filter 'EchoFarm.Mod.dll' -File -Recurse | Select-Object -First 1
            if ($null -eq $modDll) { throw 'EchoFarm.Mod.dll was not produced.' }
            $ModBuildDirectory = $modDll.DirectoryName
            $CoreExecutablePath = Join-Path $temporaryRoot 'echofarm-core.exe'
            Push-Location (Join-Path $repoRoot 'echofarm-core')
            try {
                $env:CGO_ENABLED = '0'
                $env:GOOS = 'windows'
                $env:GOARCH = 'amd64'
                & go build -trimpath -ldflags '-s -w -buildid=' -o $CoreExecutablePath ./cmd/echofarm
                if ($LASTEXITCODE -ne 0) { throw 'EchoFarm Windows core build failed.' }
            }
            finally {
                Pop-Location
            }
        }

        foreach ($file in @('EchoFarm.Mod.dll', 'EchoFarm.Bridge.dll')) {
            if ([string]::IsNullOrWhiteSpace($ModBuildDirectory) -or -not (Test-Path -LiteralPath (Join-Path $ModBuildDirectory $file) -PathType Leaf)) {
                throw "Missing built Mod file: $file"
            }
        }
        if ([string]::IsNullOrWhiteSpace($CoreExecutablePath) -or -not (Test-Path -LiteralPath $CoreExecutablePath -PathType Leaf)) {
            throw 'Missing Windows core executable.'
        }

        $stagedPackage = Join-Path $temporaryRoot 'EchoFarm'
        New-Item -ItemType Directory -Path (Join-Path $stagedPackage 'core') -Force | Out-Null
        Copy-Item -LiteralPath (Join-Path $ModBuildDirectory 'EchoFarm.Mod.dll') -Destination $stagedPackage
        Copy-Item -LiteralPath (Join-Path $ModBuildDirectory 'EchoFarm.Bridge.dll') -Destination $stagedPackage
        Copy-Item -LiteralPath (Join-Path $repoRoot 'stardew-echo-mod/src/EchoFarm.Mod/manifest.json') -Destination $stagedPackage
        Copy-Item -LiteralPath (Join-Path $repoRoot 'LICENSE') -Destination $stagedPackage
        Copy-Item -LiteralPath (Join-Path $repoRoot 'release/nexus/PLAYER_README.md') -Destination (Join-Path $stagedPackage 'README.md')
        Copy-Item -LiteralPath $CoreExecutablePath -Destination (Join-Path $stagedPackage 'core/echofarm-core.exe')

        $validation = Test-EchoFarmPackage -PackagePath $stagedPackage
        if (-not $validation.Ready) {
            $codes = ($validation.Issues | ForEach-Object Code) -join ', '
            throw "EchoFarm package validation failed: $codes"
        }

        $parent = Split-Path -Parent $OutputPath
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
        $outputStaging = Join-Path $parent ('.' + (Split-Path -Leaf $OutputPath) + '.building')
        if (Test-Path -LiteralPath $outputStaging) {
            Remove-Item -LiteralPath $outputStaging -Recurse -Force
        }
        Copy-EchoFarmDirectoryContents -Source $stagedPackage -Destination $outputStaging
        Move-EchoFarmDirectoryAtomically -StagingPath $outputStaging -DestinationPath $OutputPath

        $archivePath = "$OutputPath.zip"
        if (Test-Path -LiteralPath $archivePath) { Remove-Item -LiteralPath $archivePath -Force }
        Compress-Archive -LiteralPath $OutputPath -DestinationPath $archivePath -CompressionLevel Optimal
        $checksum = (Get-FileHash -LiteralPath $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
        Set-Content -LiteralPath "$OutputPath.sha256" -Value "$checksum  $([System.IO.Path]::GetFileName($archivePath))"
        $sourceCommit = 'unknown'
        if ($null -ne (Get-Command git -ErrorAction SilentlyContinue)) {
            $candidateCommit = (& git -C $repoRoot rev-parse HEAD 2>$null | Select-Object -First 1)
            if (-not [string]::IsNullOrWhiteSpace($candidateCommit)) {
                $sourceCommit = $candidateCommit.Trim()
            }
        }
        $checks = @(
            [pscustomobject]@{ Name = 'prerequisites'; Status = 'passed'; Signed = $true },
            [pscustomobject]@{ Name = 'mod_build'; Status = $(if ($SkipCompilation) { 'fixture' } else { 'passed' }); Signed = -not $SkipCompilation },
            [pscustomobject]@{ Name = 'sidecar_build'; Status = $(if ($SkipCompilation) { 'fixture' } else { 'passed' }); Signed = -not $SkipCompilation },
            [pscustomobject]@{ Name = 'package_allowlist'; Status = 'passed'; Signed = $true },
            [pscustomobject]@{ Name = 'gameplay_smoke'; Status = 'pending'; Signed = $false }
        )
        $evidencePath = "$OutputPath.evidence.json"
        Write-EchoFarmEvidenceReport `
            -OutputPath $evidencePath `
            -SourceCommit $sourceCommit `
            -GameVersion (Get-EchoFarmFileVersion (Join-Path $GamePath 'Stardew Valley.dll')) `
            -SmapiVersion (Get-EchoFarmFileVersion (Join-Path $GamePath 'StardewModdingAPI.dll')) `
            -Architecture 'x64' `
            -PackageSha256 $checksum `
            -Checks $checks | Out-Null
        return [pscustomobject]@{
            Ready       = $true
            PackagePath = $OutputPath
            ArchivePath = $archivePath
            Checksum    = $checksum
            EvidencePath = $evidencePath
            Issues      = @()
        }
    }
    finally {
        Remove-Item -LiteralPath $temporaryRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Install-EchoFarm {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)][string]$GamePath,
        [Parameter(Mandatory)][string]$PackagePath
    )

    $prerequisites = Test-EchoFarmPrerequisites -GamePath $GamePath -SkipHostValidation -SkipToolchainValidation
    if (-not $prerequisites.Ready) {
        $codes = ($prerequisites.Issues | ForEach-Object Code) -join ', '
        throw "EchoFarm install prerequisites failed: $codes"
    }
    $validation = Test-EchoFarmPackage -PackagePath $PackagePath
    if (-not $validation.Ready) {
        $codes = ($validation.Issues | ForEach-Object Code) -join ', '
        throw "EchoFarm package validation failed: $codes"
    }

    $modsPath = Join-Path $prerequisites.GamePath 'Mods'
    $destination = Join-Path $modsPath 'EchoFarm'
    $staging = Join-Path $modsPath '.EchoFarm.installing'
    if (Test-Path -LiteralPath $staging) {
        Remove-Item -LiteralPath $staging -Recurse -Force
    }
    try {
        Copy-EchoFarmDirectoryContents -Source $validation.PackagePath -Destination $staging
        $configPath = Join-Path $destination 'config.json'
        if (Test-Path -LiteralPath $configPath -PathType Leaf) {
            Copy-Item -LiteralPath $configPath -Destination (Join-Path $staging 'config.json') -Force
        }
        Move-EchoFarmDirectoryAtomically -StagingPath $staging -DestinationPath $destination
    }
    catch {
        Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue
        throw
    }
    [pscustomobject]@{ Installed = $true; InstallPath = $destination; ConfigPreserved = Test-Path -LiteralPath (Join-Path $destination 'config.json') }
}

function Uninstall-EchoFarm {
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter(Mandatory)][string]$GamePath,
        [string]$LocalDataPath,
        [switch]$DeleteLocalData
    )

    $resolvedGame = [System.IO.Path]::GetFullPath($GamePath)
    $installPath = Join-Path $resolvedGame 'Mods/EchoFarm'
    if ($PSCmdlet.ShouldProcess($installPath, 'Remove EchoFarm program files') -and (Test-Path -LiteralPath $installPath)) {
        Remove-Item -LiteralPath $installPath -Recurse -Force
    }

    if ([string]::IsNullOrWhiteSpace($LocalDataPath)) {
        $LocalDataPath = Join-Path ([System.Environment]::GetFolderPath([System.Environment+SpecialFolder]::LocalApplicationData)) 'EchoFarm'
    }
    $resolvedData = [System.IO.Path]::GetFullPath($LocalDataPath)
    if ($DeleteLocalData) {
        if ((Split-Path -Leaf $resolvedData) -ne 'EchoFarm') {
            throw "Refusing to delete local data outside an EchoFarm directory: $resolvedData"
        }
        if ($PSCmdlet.ShouldProcess($resolvedData, 'Delete EchoFarm local AI memory and logs') -and (Test-Path -LiteralPath $resolvedData)) {
            Remove-Item -LiteralPath $resolvedData -Recurse -Force
        }
    }
    [pscustomobject]@{
        Uninstalled        = -not (Test-Path -LiteralPath $installPath)
        InstallPath        = $installPath
        LocalDataPath      = $resolvedData
        LocalDataPreserved = Test-Path -LiteralPath $resolvedData
    }
}

Export-ModuleMember -Function Resolve-EchoFarmGamePath, Test-EchoFarmPrerequisites, Test-EchoFarmPackage, Write-EchoFarmEvidenceReport, Build-EchoFarmPackage, Install-EchoFarm, Uninstall-EchoFarm
