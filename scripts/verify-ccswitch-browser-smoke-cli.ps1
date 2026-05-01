[CmdletBinding(PositionalBinding = $false)]
param(
    [ValidateSet('self-start', 'external', 'check-only')]
    [string]$Mode = 'self-start',

    [int]$FrontendPort = 3101,
    [int]$BackendPort = 18081,
    [string]$FrontendUrl,
    [string]$BackendUrl,
    [string]$NodePath,
    [string]$GoPath,
    [string]$BackendBin,
    [Alias('BrowserPath')]
    [string]$Browser,
    [int]$NodeSmokeTimeoutSeconds = 120,
    [switch]$KeepArtifacts,
    [switch]$KeepProcessesOnFailure
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$scriptDir = Split-Path -Parent $PSCommandPath
$sharedWrapper = Join-Path $scriptDir 'verify-channel-create-browser-smoke.ps1'

if (-not (Test-Path -LiteralPath $sharedWrapper)) {
    throw "Shared CLI wrapper not found: $sharedWrapper"
}

$forwardParams = @{
    Mode = $Mode
    Driver = 'cli'
    FrontendPort = $FrontendPort
    BackendPort = $BackendPort
    NodeSmokeTimeoutSeconds = $NodeSmokeTimeoutSeconds
}

if ($PSBoundParameters.ContainsKey('FrontendUrl')) { $forwardParams.FrontendUrl = $FrontendUrl }
if ($PSBoundParameters.ContainsKey('BackendUrl')) { $forwardParams.BackendUrl = $BackendUrl }
if ($PSBoundParameters.ContainsKey('NodePath')) { $forwardParams.NodePath = $NodePath }
if ($PSBoundParameters.ContainsKey('GoPath')) { $forwardParams.GoPath = $GoPath }
if ($PSBoundParameters.ContainsKey('BackendBin')) { $forwardParams.BackendBin = $BackendBin }
if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.Browser = $Browser }
if ($KeepArtifacts) { $forwardParams.KeepArtifacts = $true }
if ($KeepProcessesOnFailure) { $forwardParams.KeepProcessesOnFailure = $true }

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))
$env:OCTOPUS_UI_SMOKE_SCRIPT = 'scripts/verify-ccswitch-browser-smoke.mjs'
$env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER = 'ccswitch-browser-smoke passed'
$env:OCTOPUS_UI_SMOKE_LABEL = 'ccswitch'

& $sharedWrapper @forwardParams
if (Test-Path Variable:\LASTEXITCODE) {
    exit $LASTEXITCODE
}
exit 0
