[CmdletBinding()]
param(
    [string]$GoArch,
    [string]$OutputPath,
    [switch]$ForceFrontendInstall,
    [switch]$SkipFrontendBuild,
    [switch]$SkipFrontendSync,
    [switch]$CheckOnly,
    [switch]$EnableRust,
    [string]$RustTarget = ''
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ""
    Write-Host ("== " + $Message + " ==")
}

function Write-Info {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ("[INFO] " + $Message)
}

function Write-Success {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ("[OK] " + $Message)
}

function Fail-Build {
    param([Parameter(Mandatory = $true)][string]$Message)

    throw $Message
}

function Get-RequiredCommand {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$InstallHint
    )

    $command = Get-Command -Name $Name -ErrorAction SilentlyContinue
    if ($null -eq $command) {
        Fail-Build "Missing required command '$Name'. $InstallHint"
    }

    return $command
}

function Get-NormalizedPath {
    param([Parameter(Mandatory = $true)][string]$Path)

    return [System.IO.Path]::GetFullPath($Path)
}

function Test-PathWithinRoot {
    param(
        [Parameter(Mandatory = $true)][string]$RootPath,
        [Parameter(Mandatory = $true)][string]$CandidatePath
    )

    $normalizedRoot = Get-NormalizedPath -Path $RootPath
    $normalizedCandidate = Get-NormalizedPath -Path $CandidatePath
    $rootWithSeparator = $normalizedRoot.TrimEnd('\') + '\'

    return $normalizedCandidate.Equals($normalizedRoot, [System.StringComparison]::OrdinalIgnoreCase) -or
        $normalizedCandidate.StartsWith($rootWithSeparator, [System.StringComparison]::OrdinalIgnoreCase)
}

function Assert-PathWithinRoot {
    param(
        [Parameter(Mandatory = $true)][string]$RootPath,
        [Parameter(Mandatory = $true)][string]$CandidatePath,
        [Parameter(Mandatory = $true)][string]$Label
    )

    if (-not (Test-PathWithinRoot -RootPath $RootPath -CandidatePath $CandidatePath)) {
        Fail-Build "$Label path must stay inside the repository root. Path: $CandidatePath"
    }
}

function Get-VersionCore {
    param([Parameter(Mandatory = $true)][string]$VersionText)

    $match = [regex]::Match($VersionText, '(\d+)\.(\d+)(?:\.(\d+))?')
    if (-not $match.Success) {
        Fail-Build "Unable to parse version from: $VersionText"
    }

    return [version]::new(
        [int]$match.Groups[1].Value,
        [int]$match.Groups[2].Value,
        $(if ($match.Groups[3].Success) { [int]$match.Groups[3].Value } else { 0 })
    )
}

function Assert-MinimumVersion {
    param(
        [Parameter(Mandatory = $true)][string]$ToolName,
        [Parameter(Mandatory = $true)][string]$ActualText,
        [Parameter(Mandatory = $true)][version]$MinimumVersion
    )

    $actualVersion = Get-VersionCore -VersionText $ActualText
    if ($actualVersion -lt $MinimumVersion) {
        Fail-Build "$ToolName version $actualVersion is too old. Required: $MinimumVersion or newer."
    }

    Write-Success "$ToolName version $actualVersion"
}

function Invoke-CheckedCommand {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter()][string[]]$Arguments = @(),
        [string]$WorkingDirectory
    )

    if ($PSBoundParameters.ContainsKey('WorkingDirectory')) {
        Push-Location $WorkingDirectory
    }

    try {
        & $FilePath @Arguments
        if ($LASTEXITCODE -ne 0) {
            $joinedArguments = if ($Arguments.Count -gt 0) { $Arguments -join ' ' } else { '' }
            Fail-Build "Command failed with exit code ${LASTEXITCODE}: $FilePath $joinedArguments"
        }
    }
    finally {
        if ($PSBoundParameters.ContainsKey('WorkingDirectory')) {
            Pop-Location
        }
    }
}

function Invoke-FrontendStaticSync {
    param(
        [Parameter(Mandatory = $true)][string]$NodePath,
        [Parameter(Mandatory = $true)][string]$RepoRoot
    )

    Invoke-CheckedCommand -FilePath $NodePath -Arguments @((Join-Path $RepoRoot 'scripts\sync-web-static.mjs')) -WorkingDirectory $RepoRoot
}

function Ensure-Directory {
    param([Parameter(Mandatory = $true)][string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        New-Item -ItemType Directory -Path $Path | Out-Null
    }
}

function Get-RustCargoPath {
    $command = Get-Command -Name 'cargo' -ErrorAction SilentlyContinue
    if ($null -ne $command) {
        return $command.Source
    }

    $candidate = Join-Path $env:USERPROFILE '.cargo\bin\cargo.exe'
    if (Test-Path -LiteralPath $candidate) {
        return $candidate
    }

    return $null
}

function Get-RustStaticLibName {
    param([Parameter(Mandatory = $true)][string]$Target)

    if ($Target -like '*-windows-msvc') {
        return 'octopus_core.lib'
    }
    return 'liboctopus_core.a'
}

function Resolve-WindowsRustTarget {
    param(
        [Parameter(Mandatory = $true)][string]$GoArch,
        [string]$UserTarget = ''
    )

    if (-not [string]::IsNullOrWhiteSpace($UserTarget)) {
        return $UserTarget.Trim()
    }

    switch ($GoArch) {
        'amd64' { return 'x86_64-pc-windows-gnu' }
        '386' { return 'i686-pc-windows-gnu' }
        default { Fail-Build "Unsupported GOARCH for Rust FFI: $GoArch" }
    }
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Get-NormalizedPath -Path (Join-Path $scriptDir '..')
. (Join-Path $scriptDir 'use-go-env.ps1')
. (Join-Path $scriptDir 'use-node-env.ps1')
$webDir = Join-Path $repoRoot 'web'
$staticDir = Join-Path $repoRoot 'static'
$versionFile = Join-Path $repoRoot 'VERSION'
$sourceOutDir = Join-Path $webDir 'out'
$targetOutDir = Join-Path $staticDir 'out'
$buildDir = Join-Path $repoRoot 'build'
$binDir = Join-Path $buildDir 'bin'

if (-not (Test-Path -LiteralPath (Join-Path $repoRoot 'main.go'))) {
    Fail-Build "Repository root not detected: $repoRoot"
}

if (-not (Test-Path -LiteralPath (Join-Path $webDir 'package.json'))) {
    Fail-Build "Web directory not detected: $webDir"
}

Assert-PathWithinRoot -RootPath $repoRoot -CandidatePath $sourceOutDir -Label 'Frontend output'
Assert-PathWithinRoot -RootPath $repoRoot -CandidatePath $targetOutDir -Label 'Static output'
Assert-PathWithinRoot -RootPath $repoRoot -CandidatePath $binDir -Label 'Build output'

Write-Step -Message 'Checking required tools'
$goCommand = [pscustomobject]@{ Source = $env:GOEXE }
$nodeCommand = [pscustomobject]@{ Source = $env:NODEEXE }
$pnpmCommand = [pscustomobject]@{ Source = $env:PNPMEXE }
$gitCommand = Get-Command -Name 'git' -ErrorAction SilentlyContinue

Assert-MinimumVersion -ToolName 'Go' -ActualText (& $goCommand.Source version) -MinimumVersion ([version]'1.25.0')
Assert-MinimumVersion -ToolName 'Node.js' -ActualText (& $nodeCommand.Source --version) -MinimumVersion ([version]'18.0.0')
Assert-MinimumVersion -ToolName 'pnpm' -ActualText (& $pnpmCommand.Source --version) -MinimumVersion ([version]'7.0.0')

$gitVersion = 'dev'
$commitId = 'unknown'
$versionFromFile = $null
if (Test-Path -LiteralPath $versionFile) {
    $firstVersionLine = Get-Content -LiteralPath $versionFile -TotalCount 1
    if (-not [string]::IsNullOrWhiteSpace($firstVersionLine)) {
        $versionFromFile = $firstVersionLine.Trim()
    }
}
if ($null -ne $gitCommand) {
    Push-Location $repoRoot
    try {
        if (-not [string]::IsNullOrWhiteSpace($versionFromFile)) {
            $gitVersion = $versionFromFile
        }
        else {
            $describedVersion = (& $gitCommand.Source describe --tags --abbrev=0 2>$null)
            if (-not [string]::IsNullOrWhiteSpace($describedVersion)) {
                $gitVersion = $describedVersion.Trim()
            }
        }

        $resolvedCommit = (& $gitCommand.Source rev-parse --short HEAD 2>$null)
        if (-not [string]::IsNullOrWhiteSpace($resolvedCommit)) {
            $commitId = $resolvedCommit.Trim()
        }
    }
    finally {
        Pop-Location
    }
}
else {
    Write-Info 'git not found, build metadata will fall back to dev/unknown.'
    if (-not [string]::IsNullOrWhiteSpace($versionFromFile)) {
        $gitVersion = $versionFromFile
    }
}

$goArchResolved = if ([string]::IsNullOrWhiteSpace($GoArch)) {
    (& $goCommand.Source env GOARCH).Trim()
}
else {
    $GoArch.Trim().ToLowerInvariant()
}

if ([string]::IsNullOrWhiteSpace($goArchResolved)) {
    Fail-Build 'Unable to determine GOARCH.'
}

$resolvedOutputPath = if ([string]::IsNullOrWhiteSpace($OutputPath)) {
    Join-Path $binDir ("octopus-windows-{0}.exe" -f $goArchResolved)
}
else {
    Get-NormalizedPath -Path (Join-Path $repoRoot $OutputPath)
}

Assert-PathWithinRoot -RootPath $repoRoot -CandidatePath $resolvedOutputPath -Label 'Executable output'

if ($CheckOnly) {
    Write-Step -Message 'Check-only summary'
    Write-Info "Frontend build: $(-not $SkipFrontendBuild)"
    Write-Info "Static sync: $(-not $SkipFrontendSync)"
    Write-Info "Rust FFI: $EnableRust"
    Write-Info "GOOS: windows"
    Write-Info "GOARCH: $goArchResolved"
    Write-Info "Executable output: $resolvedOutputPath"
    Write-Success 'Dependency and path validation completed.'
    return
}

Ensure-Directory -Path $buildDir
Ensure-Directory -Path $binDir

if (-not $SkipFrontendBuild) {
    Write-Step -Message 'Building frontend'

    $needInstall = $ForceFrontendInstall -or -not (Test-Path -LiteralPath (Join-Path $webDir 'node_modules'))
    if ($needInstall) {
        Write-Info 'Installing frontend dependencies with pnpm install --frozen-lockfile'
        Invoke-CheckedCommand -FilePath $pnpmCommand.Source -Arguments @('install', '--frozen-lockfile') -WorkingDirectory $webDir
    }

    $previousVersion = $env:NEXT_PUBLIC_APP_VERSION
    $env:NEXT_PUBLIC_APP_VERSION = $gitVersion
    try {
        Invoke-CheckedCommand -FilePath $pnpmCommand.Source -Arguments @('run', 'build') -WorkingDirectory $webDir
    }
    finally {
        if ($null -eq $previousVersion) {
            Remove-Item Env:NEXT_PUBLIC_APP_VERSION -ErrorAction SilentlyContinue
        }
        else {
            $env:NEXT_PUBLIC_APP_VERSION = $previousVersion
        }
    }

    if (-not (Test-Path -LiteralPath $sourceOutDir)) {
        Fail-Build "Frontend build completed, but output directory was not found: $sourceOutDir"
    }

    Write-Success 'Frontend build completed.'
}

if (-not $SkipFrontendSync) {
    Write-Step -Message 'Syncing frontend output to static/out'

    if (-not (Test-Path -LiteralPath $sourceOutDir)) {
        Fail-Build "Frontend output directory not found: $sourceOutDir"
    }

    $expectedTarget = Get-NormalizedPath -Path $targetOutDir
    $canonicalTarget = Get-NormalizedPath -Path (Join-Path $staticDir 'out')
    if (-not $expectedTarget.Equals($canonicalTarget, [System.StringComparison]::OrdinalIgnoreCase)) {
        Fail-Build "Refusing to sync into unexpected target directory: $targetOutDir"
    }

    Invoke-FrontendStaticSync -NodePath $nodeCommand.Source -RepoRoot $repoRoot

    Write-Success "Frontend assets synced to $targetOutDir"
}

Write-Step -Message 'Building Windows executable'
$buildTime = [DateTimeOffset]::UtcNow.ToOffset([TimeSpan]::FromHours(8)).ToString('yyyy-MM-dd_HH:mm:ss_zzz')
$ldflags = @(
    "-X", "github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf.Version=$gitVersion",
    "-X", "github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf.BuildTime=$buildTime",
    "-X", "github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf.Author=xiaoli0412",
    "-X", "github.com/xiaoli0412/octopus-xiaoli-repo/internal/conf.Commit=$commitId",
    "-s", "-w"
)

$rustBuildTags = 'jsoniter'
$rustCgoEnabled = '0'
$rustCc = $null

if ($EnableRust) {
    $cargoExe = Get-RustCargoPath
    if ([string]::IsNullOrWhiteSpace($cargoExe)) {
        Fail-Build @"
Rust FFI was requested but cargo was not found.
Install options:
  - winget install RustLang.Rustup
  - choco install rustup
  - Or run the installer from https://rustup.rs
After installation, open a new PowerShell window and try again.
"@
    }

    $resolvedRustTarget = Resolve-WindowsRustTarget -GoArch $goArchResolved -UserTarget $RustTarget
    $rustLibName = Get-RustStaticLibName -Target $resolvedRustTarget
    $rustCoreDir = Join-Path $repoRoot 'rust\core'
    $rustArtifactDir = Join-Path $rustCoreDir (Join-Path 'target' (Join-Path $resolvedRustTarget 'release'))
    $rustArtifact = Join-Path $rustArtifactDir $rustLibName
    $rustStagingDir = Join-Path $rustCoreDir 'target\release'

    Write-Info "Building Rust static library for $resolvedRustTarget"
    Invoke-CheckedCommand -FilePath $cargoExe -Arguments @('build', '--release', '--target', $resolvedRustTarget) -WorkingDirectory $rustCoreDir

    if (-not (Test-Path -LiteralPath $rustArtifact)) {
        Fail-Build "Rust build did not produce expected artifact: $rustArtifact"
    }

    Ensure-Directory -Path $rustStagingDir
    Copy-Item -LiteralPath $rustArtifact -Destination (Join-Path $rustStagingDir $rustLibName) -Force
    Write-Success "Rust static library staged: $rustStagingDir\$rustLibName"

    $gccCommand = Get-Command -Name 'gcc' -ErrorAction SilentlyContinue
    if ($null -eq $gccCommand) {
        Fail-Build @"
Rust FFI requires a C compiler for cgo but gcc was not found in PATH.
Install MinGW-w64, for example:
  - winget install MSYS2.MSYS2
  - Or use the MinGW-w64 distribution from https://www.mingw-w64.org/downloads/
Then ensure gcc.exe is available in PATH.
"@
    }

    $rustBuildTags = 'rust jsoniter'
    $rustCgoEnabled = '1'
    $rustCc = 'gcc'
}

$previousGoos = $env:GOOS
$previousGoarch = $env:GOARCH
$previousCgo = $env:CGO_ENABLED
$previousCc = $env:CC
$env:GOOS = 'windows'
$env:GOARCH = $goArchResolved
$env:CGO_ENABLED = $rustCgoEnabled
if ($null -ne $rustCc) {
    $env:CC = $rustCc
}

try {
    Invoke-CheckedCommand -FilePath $goCommand.Source -Arguments @(
        'build',
        '-trimpath',
        '-tags', $rustBuildTags,
        '-ldflags', ($ldflags -join ' '),
        '-o', $resolvedOutputPath,
        '.'
    ) -WorkingDirectory $repoRoot
}
finally {
    if ($null -eq $previousGoos) {
        Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    }
    else {
        $env:GOOS = $previousGoos
    }

    if ($null -eq $previousGoarch) {
        Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    }
    else {
        $env:GOARCH = $previousGoarch
    }

    if ($null -eq $previousCgo) {
        Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    }
    else {
        $env:CGO_ENABLED = $previousCgo
    }

    if ($null -eq $previousCc) {
        Remove-Item Env:CC -ErrorAction SilentlyContinue
    }
    else {
        $env:CC = $previousCc
    }
}

if (-not (Test-Path -LiteralPath $resolvedOutputPath)) {
    Fail-Build "Go build reported success, but executable was not found: $resolvedOutputPath"
}

Write-Success "Windows executable created: $resolvedOutputPath"
