[CmdletBinding(PositionalBinding = $false)]
param(
    [ValidateSet('self-start', 'external', 'check-only')]
    [string]$Mode = 'self-start',

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
$sharedWrapper = Join-Path $scriptDir 'verify-channel-create-browser-smoke-cdp.ps1'

if (-not (Test-Path -LiteralPath $sharedWrapper)) {
    throw "Shared CDP wrapper not found: $sharedWrapper"
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
    NodeSmokeScript = 'scripts/verify-channel-create-browser-smoke-cdp.mjs'
    NodeSmokeSuccessMarker = 'home-layout-browser-smoke-cdp passed'
    SmokeLabel = 'home layout'
}

$env:OCTOPUS_UI_SMOKE_SCENARIO = 'home-layout'

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

& $sharedWrapper @forwardParams
if (Test-Path Variable:\LASTEXITCODE) {
    exit $LASTEXITCODE
}
exit 0
