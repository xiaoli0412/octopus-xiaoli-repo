[CmdletBinding()]
param(
    [switch]$BackendOnly,
    [switch]$FrontendOnly,
    [switch]$NoNewWindow,
    [switch]$InstallFrontendDependencies,
    [switch]$CheckOnly,
    [string]$ApiBaseUrl = 'http://127.0.0.1:8080'
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

function Fail-Dev {
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
        Fail-Dev "Missing required command '$Name'. $InstallHint"
    }

    return $command
}

function Get-VersionCore {
    param([Parameter(Mandatory = $true)][string]$VersionText)

    $match = [regex]::Match($VersionText, '(\d+)\.(\d+)(?:\.(\d+))?')
    if (-not $match.Success) {
        Fail-Dev "Unable to parse version from: $VersionText"
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
        Fail-Dev "$ToolName version $actualVersion is too old. Required: $MinimumVersion or newer."
    }

    Write-Success "$ToolName version $actualVersion"
}

function Invoke-CheckedCommand {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter()][string[]]$Arguments = @(),
        [Parameter(Mandatory = $true)][string]$WorkingDirectory
    )

    Push-Location $WorkingDirectory
    try {
        & $FilePath @Arguments
        if ($LASTEXITCODE -ne 0) {
            $joinedArguments = if ($Arguments.Count -gt 0) { $Arguments -join ' ' } else { '' }
            Fail-Dev "Command failed with exit code ${LASTEXITCODE}: $FilePath $joinedArguments"
        }
    }
    finally {
        Pop-Location
    }
}

function Start-WindowProcess {
    param(
        [Parameter(Mandatory = $true)][string]$Title,
        [Parameter(Mandatory = $true)][string]$WorkingDirectory,
        [string]$Command,
        [string]$FilePath,
        [string[]]$Arguments = @(),
        [hashtable]$Environment
    )

    $startProcessSupportsEnvironment = (Get-Command -Name 'Start-Process' -ErrorAction Stop).Parameters.ContainsKey('Environment')

    if (-not [string]::IsNullOrWhiteSpace($FilePath)) {
        if ($null -ne $Environment -and -not $startProcessSupportsEnvironment) {
            $powershellPath = (Get-RequiredCommand -Name 'powershell' -InstallHint 'Windows PowerShell is required to launch dev windows.').Source
            $commandParts = @()

            foreach ($entry in $Environment.GetEnumerator()) {
                $literalName = Convert-ToSingleQuotedLiteral -Value ([string]$entry.Key)
                $literalValue = Convert-ToSingleQuotedLiteral -Value ([string]$entry.Value)
                $commandParts += "Set-Item -Path 'Env:$literalName' -Value '$literalValue'"
            }

            $literalFilePath = Convert-ToSingleQuotedLiteral -Value $FilePath
            $quotedArguments = foreach ($argument in $Arguments) {
                "'$(Convert-ToSingleQuotedLiteral -Value $argument)'"
            }
            $invocation = "& '$literalFilePath'"
            if ($quotedArguments.Count -gt 0) {
                $invocation += " " + ($quotedArguments -join ' ')
            }
            $commandParts += $invocation

            $process = Start-Process -FilePath $powershellPath -WorkingDirectory $WorkingDirectory -ArgumentList @(
                '-NoExit',
                '-ExecutionPolicy', 'Bypass',
                '-Command', ($commandParts -join '; ')
            ) -PassThru
        }
        else {
            $startProcessParams = @{
                FilePath = $FilePath
                WorkingDirectory = $WorkingDirectory
                ArgumentList = $Arguments
                PassThru = $true
            }
            if ($null -ne $Environment) {
                $startProcessParams.Environment = $Environment
            }
            $process = Start-Process @startProcessParams
        }
    }
    else {
        if ([string]::IsNullOrWhiteSpace($Command)) {
            Fail-Dev 'Start-WindowProcess requires either Command or FilePath.'
        }
        $powershellPath = (Get-RequiredCommand -Name 'powershell' -InstallHint 'Windows PowerShell is required to launch dev windows.').Source
        $process = Start-Process -FilePath $powershellPath -WorkingDirectory $WorkingDirectory -ArgumentList @(
            '-NoExit',
            '-ExecutionPolicy', 'Bypass',
            '-Command', $Command
        ) -PassThru
    }

    Write-Success "$Title started in a new PowerShell window (PID $($process.Id))."
}

function Convert-ToSingleQuotedLiteral {
    param([Parameter(Mandatory = $true)][string]$Value)

    return $Value.Replace("'", "''")
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))
$webDir = Join-Path $repoRoot 'web'

if (-not (Test-Path -LiteralPath (Join-Path $repoRoot 'main.go'))) {
    Fail-Dev "Repository root not detected: $repoRoot"
}

if (-not (Test-Path -LiteralPath (Join-Path $webDir 'package.json'))) {
    Fail-Dev "Web directory not detected: $webDir"
}

$startBackend = -not $FrontendOnly
$startFrontend = -not $BackendOnly

if (-not $startBackend -and -not $startFrontend) {
    Fail-Dev 'No service selected to start.'
}

if ($NoNewWindow -and $startBackend -and $startFrontend) {
    Fail-Dev 'The -NoNewWindow mode can only be used with -BackendOnly or -FrontendOnly.'
}

Write-Step -Message 'Checking required tools'
$goCommand = $null
$goExe = $null
$nodeCommand = $null
$pnpmCommand = $null

if ($startBackend) {
    . (Join-Path $scriptDir 'use-go-env.ps1')
    $goExe = $env:GOEXE
    $goCommand = [pscustomobject]@{ Source = $goExe }
    Assert-MinimumVersion -ToolName 'Go' -ActualText (& $goCommand.Source version) -MinimumVersion ([version]'1.24.4')
}

if ($startFrontend) {
    $nodeCommand = Get-RequiredCommand -Name 'node' -InstallHint 'Install Node.js 18+ and make sure it is on PATH.'
    $pnpmCommand = Get-RequiredCommand -Name 'pnpm' -InstallHint 'Install pnpm and make sure it is on PATH.'
    Assert-MinimumVersion -ToolName 'Node.js' -ActualText (& $nodeCommand.Source --version) -MinimumVersion ([version]'18.0.0')
    Assert-MinimumVersion -ToolName 'pnpm' -ActualText (& $pnpmCommand.Source --version) -MinimumVersion ([version]'7.0.0')
}

if ($startFrontend -and ($InstallFrontendDependencies -or -not (Test-Path -LiteralPath (Join-Path $webDir 'node_modules')))) {
    Write-Step -Message 'Installing frontend dependencies'
    Invoke-CheckedCommand -FilePath $pnpmCommand.Source -Arguments @('install', '--frozen-lockfile') -WorkingDirectory $webDir
}

$backendCommand = if ($startBackend) {
    "$goExe run main.go start"
}
else {
    $null
}

$frontendCommand = if ($startFrontend) {
    "NEXT_PUBLIC_API_BASE_URL=$ApiBaseUrl $($pnpmCommand.Source) run dev"
}
else {
    $null
}
if ($CheckOnly) {
    Write-Step -Message 'Check-only summary'
    if ($startBackend) {
        Write-Info "Backend command: $backendCommand"
    }

    if ($startFrontend) {
        Write-Info "Frontend command: $frontendCommand"
    }

    Write-Success 'Dependency and launch validation completed.'
    return
}

if ($NoNewWindow) {
    if ($startBackend) {
        Write-Step -Message 'Starting backend in current window'
        $previousDebug = $env:OCTOPUS_DEBUG
        $env:OCTOPUS_DEBUG = 'true'
        try {
            Invoke-CheckedCommand -FilePath $goCommand.Source -Arguments @('run', 'main.go', 'start') -WorkingDirectory $repoRoot
        }
        finally {
            if ($null -eq $previousDebug) {
                Remove-Item Env:OCTOPUS_DEBUG -ErrorAction SilentlyContinue
            }
            else {
                $env:OCTOPUS_DEBUG = $previousDebug
            }
        }
    }
    elseif ($startFrontend) {
        Write-Step -Message 'Starting frontend in current window'
        $previousApiBaseUrl = $env:NEXT_PUBLIC_API_BASE_URL
        $env:NEXT_PUBLIC_API_BASE_URL = $ApiBaseUrl
        try {
            Invoke-CheckedCommand -FilePath $pnpmCommand.Source -Arguments @('run', 'dev') -WorkingDirectory $webDir
        }
        finally {
            if ($null -eq $previousApiBaseUrl) {
                Remove-Item Env:NEXT_PUBLIC_API_BASE_URL -ErrorAction SilentlyContinue
            }
            else {
                $env:NEXT_PUBLIC_API_BASE_URL = $previousApiBaseUrl
            }
        }
    }

    return
}

Write-Step -Message 'Starting development services'
if ($startBackend) {
    Start-WindowProcess -Title 'Backend dev server' -WorkingDirectory $repoRoot -FilePath $goExe -Arguments @('run', 'main.go', 'start') -Environment @{ OCTOPUS_DEBUG = 'true' }
}

if ($startFrontend) {
    Start-WindowProcess -Title 'Frontend dev server' -WorkingDirectory $webDir -FilePath $pnpmCommand.Source -Arguments @('run', 'dev') -Environment @{ NEXT_PUBLIC_API_BASE_URL = $ApiBaseUrl }
}
