[CmdletBinding(PositionalBinding = $false)]
param(
    [ValidateSet('self-start', 'external', 'check-only')]
    [string]$Mode = 'self-start',

    [ValidateSet('cdp', 'cli')]
    [string]$Driver = 'cdp',

    [int]$FrontendPort = 3101,
    [int]$BackendPort = 18081,
    [int]$CdpPort = 9222,

    [string]$FrontendUrl,
    [string]$BackendUrl,
    [string]$CdpUrl,
    [string]$NodePath,
    [string]$GoPath,
    [string]$BackendBin,
    [Alias('BrowserPath')]
    [string]$Browser,
    [int]$NodeSmokeTimeoutSeconds = 120,

    [ValidateRange(1000, 300000)]
    [int]$CdpCommandTimeoutMs = 15000,

    [ValidateSet('default', 'relaxed', 'headed-relaxed')]
    [string]$EdgeLaunchPreset = 'default',

    [ValidateSet('temp-random', 'workspace-fixed')]
    [string]$EdgeProfileStrategy = 'temp-random',

    [ValidateSet('auto', 'json-new', 'attached-session')]
    [string]$CdpPageBootstrapStrategy = 'attached-session',

    [ValidateSet('page-lifecycle-runtime', 'runtime-page-lifecycle')]
    [string]$CdpBootstrapCommandOrder = 'page-lifecycle-runtime',

    [switch]$BootstrapExternalCdpSession,
    [switch]$RequireExternalCdpPreflight,
    [switch]$SelfStartServices,
    [switch]$KeepArtifacts,
    [switch]$KeepProcessesOnFailure
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$scriptDir = Split-Path -Parent $PSCommandPath

if ($Driver -eq 'cdp') {
    $cdpWrapper = Join-Path $scriptDir 'verify-group-create-browser-smoke-cdp.ps1'

    if (-not (Test-Path -LiteralPath $cdpWrapper)) {
        throw "CDP wrapper not found: $cdpWrapper"
    }

    $forwardParams = @{
        Mode = $Mode
        Driver = 'cdp'
        FrontendPort = $FrontendPort
        BackendPort = $BackendPort
        CdpPort = $CdpPort
        NodeSmokeTimeoutSeconds = $NodeSmokeTimeoutSeconds
        CdpCommandTimeoutMs = $CdpCommandTimeoutMs
        EdgeLaunchPreset = $EdgeLaunchPreset
        EdgeProfileStrategy = $EdgeProfileStrategy
        CdpPageBootstrapStrategy = $CdpPageBootstrapStrategy
        CdpBootstrapCommandOrder = $CdpBootstrapCommandOrder
    }

    if ($PSBoundParameters.ContainsKey('FrontendUrl')) { $forwardParams.FrontendUrl = $FrontendUrl }
    if ($PSBoundParameters.ContainsKey('BackendUrl')) { $forwardParams.BackendUrl = $BackendUrl }
    if ($PSBoundParameters.ContainsKey('CdpUrl')) { $forwardParams.CdpUrl = $CdpUrl }
    if ($PSBoundParameters.ContainsKey('NodePath')) { $forwardParams.NodePath = $NodePath }
    if ($PSBoundParameters.ContainsKey('GoPath')) { $forwardParams.GoPath = $GoPath }
    if ($PSBoundParameters.ContainsKey('BackendBin')) { $forwardParams.BackendBin = $BackendBin }
    if ($PSBoundParameters.ContainsKey('Browser')) { $forwardParams.BrowserPath = $Browser }
    if ($BootstrapExternalCdpSession) { $forwardParams.BootstrapExternalCdpSession = $true }
    if ($RequireExternalCdpPreflight) { $forwardParams.RequireExternalCdpPreflight = $true }
    if ($SelfStartServices) { $forwardParams.SelfStartServices = $true }
    if ($KeepArtifacts) { $forwardParams.KeepArtifacts = $true }
    if ($KeepProcessesOnFailure) { $forwardParams.KeepProcessesOnFailure = $true }

    & $cdpWrapper @forwardParams
    if (Test-Path Variable:\LASTEXITCODE) {
        exit $LASTEXITCODE
    }
    exit 0
}

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

$env:OCTOPUS_UI_SMOKE_SCRIPT = 'scripts/verify-group-create-browser-smoke.mjs'
$env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER = 'group-create-browser-smoke passed'
$env:OCTOPUS_UI_SMOKE_LABEL = 'group create'

& $sharedWrapper @forwardParams
if (Test-Path Variable:\LASTEXITCODE) {
    exit $LASTEXITCODE
}
exit 0
