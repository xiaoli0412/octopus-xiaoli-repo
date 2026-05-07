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

    [string]$NodeSmokeScript = 'scripts/verify-channel-create-browser-smoke-cdp.mjs',
    [string]$NodeSmokeSuccessMarker = 'channel-create-browser-smoke-cdp passed',
    [string]$SmokeLabel = 'channel create',

    [switch]$BootstrapExternalCdpSession,
    [switch]$RequireExternalCdpPreflight,
    [switch]$SelfStartServices,
    [switch]$KeepArtifacts,
    [switch]$KeepProcessesOnFailure
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Get-SmokeSlug {
    param([Parameter(Mandatory = $true)][string]$Label)

    $slug = $Label.ToLowerInvariant()
    $slug = [regex]::Replace($slug, '[^a-z0-9]+', '-')
    $slug = $slug.Trim('-')
    if ([string]::IsNullOrWhiteSpace($slug)) {
        return 'browser-smoke'
    }
    return $slug
}

$smokeSlug = Get-SmokeSlug -Label $SmokeLabel

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ''
    Write-Host ("== " + $Message + " ==")
}

function Get-RepoRoot {
    $scriptDir = Split-Path -Parent $PSCommandPath
    return [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))
}

function Normalize-Url {
    param([Parameter(Mandatory = $true)][string]$Value)

    return $Value.TrimEnd('/')
}

function Resolve-StaticDir {
    param([Parameter(Mandatory = $true)][string]$RepoRoot)

    $syncedDir = Join-Path $RepoRoot 'static\out'
    $sourceDir = Join-Path $RepoRoot 'web\out'
    $syncedIndex = Join-Path $syncedDir 'index.html'
    $sourceIndex = Join-Path $sourceDir 'index.html'

    if ((Test-Path -LiteralPath $sourceIndex) -and ((-not (Test-Path -LiteralPath $syncedIndex)) -or ((Get-Item -LiteralPath $sourceIndex).LastWriteTimeUtc -gt (Get-Item -LiteralPath $syncedIndex).LastWriteTimeUtc))) {
        return $sourceDir
    }

    if (Test-Path -LiteralPath $syncedDir) {
        return $syncedDir
    }

    if (Test-Path -LiteralPath $sourceDir) {
        return $sourceDir
    }

    return $syncedDir
}

function Resolve-CommandPath {
    param(
        [string[]]$Candidates,
        [string]$CommandName
    )

    foreach ($candidate in $Candidates) {
        if ([string]::IsNullOrWhiteSpace($candidate)) {
            continue
        }

        if (Test-Path -LiteralPath $candidate) {
            $resolved = Resolve-Path -LiteralPath $candidate
            return $resolved.ProviderPath
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($CommandName)) {
        $command = Get-Command -Name $CommandName -ErrorAction SilentlyContinue
        if ($null -ne $command) {
            return $command.Source
        }
    }

    return $null
}

function Get-LogTail {
    param(
        [string]$Path,
        [int]$LineCount = 60
    )

    if ([string]::IsNullOrWhiteSpace($Path) -or -not (Test-Path -LiteralPath $Path)) {
        return ''
    }

    try {
        return ((Get-Content -LiteralPath $Path -Tail $LineCount) -join [Environment]::NewLine).Trim()
    }
    catch {
        return ''
    }
}

function Get-JsonFileContent {
    param([string]$Path)

    if ([string]::IsNullOrWhiteSpace($Path) -or -not (Test-Path -LiteralPath $Path)) {
        return $null
    }

    $candidatePaths = New-Object System.Collections.Generic.List[string]

    try {
        $resolved = Resolve-Path -LiteralPath $Path -ErrorAction Stop
        if ($null -ne $resolved -and -not [string]::IsNullOrWhiteSpace($resolved.ProviderPath)) {
            $candidatePaths.Add($resolved.ProviderPath)
        }
    }
    catch {
    }

    $candidatePaths.Add([string]$Path)

    foreach ($candidate in ($candidatePaths | Select-Object -Unique)) {
        foreach ($reader in @(
            { param($value) [System.IO.File]::ReadAllText($value) },
            { param($value) Get-Content -LiteralPath $value -Raw -Encoding utf8 -ErrorAction Stop },
            { param($value) Get-Content -LiteralPath $value -Raw -ErrorAction Stop }
        )) {
            try {
                $raw = & $reader $candidate
                if ([string]::IsNullOrWhiteSpace($raw)) {
                    continue
                }

                $normalized = ([string]$raw).TrimStart([char]0xFEFF).Trim([char]0)
                if ([string]::IsNullOrWhiteSpace($normalized)) {
                    continue
                }

                return $normalized | ConvertFrom-Json -ErrorAction Stop
            }
            catch {
                continue
            }
        }
    }

    return $null
}

function Get-OptionalObjectPropertyValue {
    param(
        $Object,
        [string]$Name
    )

    if ($null -eq $Object -or [string]::IsNullOrWhiteSpace($Name)) {
        return $null
    }

    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $null
    }

    return $property.Value
}

function Read-LogContent {
    param([string]$Path)

    if ([string]::IsNullOrWhiteSpace($Path) -or -not (Test-Path -LiteralPath $Path)) {
        return ''
    }

    try {
        $content = [System.IO.File]::ReadAllText($Path)
        if ($null -eq $content) {
            return ''
        }
        return [string]$content
    }
    catch {
        try {
            $fallback = Get-Content -LiteralPath $Path -Raw -ErrorAction Stop
            if ($null -eq $fallback) {
                return ''
            }
            return [string]$fallback
        }
        catch {
            return ''
        }
    }
}

function Write-JsonArtifact {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)]$Value
    )

    $parent = Split-Path -Parent $Path
    if (-not [string]::IsNullOrWhiteSpace($parent)) {
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
    }

    $Value | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $Path -Encoding utf8
}

function Get-StableSmokeDiagnosticDirectory {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$SmokeLabel
    )

    switch ($SmokeLabel) {
        'backup' { return Join-Path $RepoRoot 'build\verify-backup-browser' }
        default { return $null }
    }
}

function Publish-StableSmokeDiagnosticCopies {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$SmokeLabel,
        [string]$CdpDiagnosticPath,
        [string]$ExternalPreflightDiagnosticPath
    )

    $stableDir = Get-StableSmokeDiagnosticDirectory -RepoRoot $RepoRoot -SmokeLabel $SmokeLabel
    if ([string]::IsNullOrWhiteSpace($stableDir)) {
        return
    }

    New-Item -ItemType Directory -Path $stableDir -Force | Out-Null

    foreach ($entry in @(@{ Source = $CdpDiagnosticPath; Dest = 'latest-cdp-diagnostic.json'; Label = 'Stable smoke CDP diagnostic copy' }, @{ Source = $ExternalPreflightDiagnosticPath; Dest = 'latest-external-preflight-diagnostic.json'; Label = 'Stable smoke external preflight diagnostic copy' })) {
        if ([string]::IsNullOrWhiteSpace($entry.Source) -or -not (Test-Path -LiteralPath $entry.Source)) {
            continue
        }

        $destPath = Join-Path $stableDir $entry.Dest
        Copy-Item -LiteralPath $entry.Source -Destination $destPath -Force
        Write-Host ('{0}: {1}' -f $entry.Label, $destPath)
    }
}

function Write-StableSmokeDiagnosticPreview {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$SmokeLabel
    )

    $stableDir = Get-StableSmokeDiagnosticDirectory -RepoRoot $RepoRoot -SmokeLabel $SmokeLabel
    if ([string]::IsNullOrWhiteSpace($stableDir) -or -not (Test-Path -LiteralPath $stableDir)) {
        return
    }

    foreach ($entry in @(@{ Path = Join-Path $stableDir 'latest-cdp-diagnostic.json'; Label = 'Stable smoke CDP diagnostic'; Keys = @('generatedAt','checkedAt','classification','errorName','pageStrategy','bootstrapCommandOrder','hint') }, @{ Path = Join-Path $stableDir 'latest-external-preflight-diagnostic.json'; Label = 'Stable smoke external preflight diagnostic'; Keys = @('generatedAt','checkedAt','overallClassification','primaryBlockingCheck') })) {
        if (-not (Test-Path -LiteralPath $entry.Path)) {
            continue
        }

        Write-Host ('{0} copy: {1}' -f $entry.Label, $entry.Path)
        $payload = Get-JsonFileContent -Path $entry.Path
        if ($null -eq $payload) {
            Write-Host ('{0} preview unavailable: could not parse JSON.' -f $entry.Label)
            continue
        }

        foreach ($key in $entry.Keys) {
            $value = Get-OptionalObjectPropertyValue -Object $payload -Name $key
            if ($null -eq $value -or [string]::IsNullOrWhiteSpace([string]$value)) {
                continue
            }
            Write-Host ('{0} {1}: {2}' -f $entry.Label, $key, $value)
        }

        $commandTimeoutMs = Get-OptionalObjectPropertyValue -Object $payload -Name 'commandTimeoutMs'
        if ($entry.Path -like '*latest-cdp-diagnostic.json' -and $null -ne $commandTimeoutMs) {
            Write-Host ('{0} commandTimeoutMs: {1}' -f $entry.Label, $commandTimeoutMs)
        }

        $failedChecks = @(Get-OptionalObjectPropertyValue -Object $payload -Name 'failedChecks')
        if ($entry.Path -like '*external-preflight*' -and $failedChecks.Count -gt 0) {
            Write-Host ('{0} failedChecks: {1}' -f $entry.Label, ($failedChecks -join ', '))
        }
    }
}

function Write-StableSmokeFailureSummary {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$SmokeLabel,
        [Parameter(Mandatory = $true)][string]$SummaryText,
        [string]$Classification = 'host_blocker'
    )

    $stableDir = Get-StableSmokeDiagnosticDirectory -RepoRoot $RepoRoot -SmokeLabel $SmokeLabel
    if ([string]::IsNullOrWhiteSpace($stableDir)) {
        return
    }

    New-Item -ItemType Directory -Path $stableDir -Force | Out-Null
    $path = Join-Path $stableDir 'latest-wrapper-failure-summary.txt'
    $payload = @(
        ('classification: {0}' -f $Classification),
        ('checkedAt: {0}' -f (Get-Date).ToString('o')),
        '',
        $SummaryText
    ) -join [Environment]::NewLine
    Set-Content -LiteralPath $path -Value $payload -Encoding utf8
    Write-Host ('Stable smoke wrapper failure summary: {0}' -f $path)
}

function Get-HttpReachabilityResult {
    param(
        [Parameter(Mandatory = $true)][string]$Label,
        [Parameter(Mandatory = $true)][string]$Url,
        [int]$TimeoutSeconds = 30
    )

    $startedAt = Get-Date
    $deadline = $startedAt.AddSeconds($TimeoutSeconds)
    $attempts = 0
    $lastError = $null
    $lastStatusCode = $null
    $lastStatusDescription = $null
    $classification = 'timeout'

    while ((Get-Date) -lt $deadline) {
        $attempts++
        try {
            $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec ([Math]::Min([Math]::Max($TimeoutSeconds, 1), 5))
            $lastStatusCode = [int]$response.StatusCode
            $lastStatusDescription = [string]$response.StatusDescription

            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 500) {
                return [pscustomobject]@{
                    label = $Label
                    url = $Url
                    reachable = $true
                    classification = 'ok'
                    statusCode = $lastStatusCode
                    statusDescription = $lastStatusDescription
                    attempts = $attempts
                    elapsedMs = [int]((Get-Date) - $startedAt).TotalMilliseconds
                    lastError = $null
                    checkedAt = (Get-Date).ToString('o')
                }
            }

            $classification = 'http_status_unreachable'
            $lastError = if ([string]::IsNullOrWhiteSpace($lastStatusDescription)) {
                'HTTP {0}' -f $lastStatusCode
            }
            else {
                'HTTP {0} {1}' -f $lastStatusCode, $lastStatusDescription
            }
        }
        catch {
            $lastError = Get-ErrorMessageChain -ErrorRecord $_
            $classification = if (Test-IsServiceProviderFailureMessage -Message $lastError) {
                'host_networking_blocker'
            }
            else {
                'request_failed'
            }
        }

        Start-Sleep -Milliseconds 500
    }

    return [pscustomobject]@{
        label = $Label
        url = $Url
        reachable = $false
        classification = $classification
        statusCode = $lastStatusCode
        statusDescription = $lastStatusDescription
        attempts = $attempts
        elapsedMs = [int]((Get-Date) - $startedAt).TotalMilliseconds
        lastError = $lastError
        checkedAt = (Get-Date).ToString('o')
    }
}

function Format-HttpReachabilitySummaryLine {
    param([Parameter(Mandatory = $true)]$Result)

    if ($Result.classification -eq 'skipped') {
        $detail = if (-not [string]::IsNullOrWhiteSpace([string]$Result.lastError)) {
            $Result.lastError
        }
        else {
            'skipped without additional detail'
        }

        return ('- {0}: skipped at {1} ({2})' -f $Result.label, $Result.url, $detail)
    }

    if ($Result.reachable) {
        $statusText = if ($null -ne $Result.statusCode) {
            if ([string]::IsNullOrWhiteSpace([string]$Result.statusDescription)) {
                'HTTP {0}' -f $Result.statusCode
            }
            else {
                'HTTP {0} {1}' -f $Result.statusCode, $Result.statusDescription
            }
        }
        else {
            'reachable'
        }

        return ('- {0}: reachable at {1} after {2} attempts ({3}, {4}ms)' -f $Result.label, $Result.url, $Result.attempts, $statusText, $Result.elapsedMs)
    }

    $detail = if (-not [string]::IsNullOrWhiteSpace([string]$Result.lastError)) {
        $Result.lastError
    }
    elseif ($null -ne $Result.statusCode) {
        if ([string]::IsNullOrWhiteSpace([string]$Result.statusDescription)) {
            'HTTP {0}' -f $Result.statusCode
        }
        else {
            'HTTP {0} {1}' -f $Result.statusCode, $Result.statusDescription
        }
    }
    else {
        'no response details captured'
    }

    return ('- {0}: unreachable at {1} after {2} attempts ({3}, {4})' -f $Result.label, $Result.url, $Result.attempts, $Result.classification, $detail)
}

function Set-ExternalPreflightEntry {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Summary,
        [Parameter(Mandatory = $true)][string]$EntryName,
        [Parameter(Mandatory = $true)]$Result
    )

    $Summary[$EntryName] = [ordered]@{
        label = $Result.label
        url = $Result.url
        reachable = [bool]$Result.reachable
        classification = $Result.classification
        statusCode = $Result.statusCode
        statusDescription = $Result.statusDescription
        attempts = $Result.attempts
        elapsedMs = $Result.elapsedMs
        lastError = $Result.lastError
        checkedAt = $Result.checkedAt
    }
}

function New-ExternalPreflightSkippedEntry {
    param(
        [Parameter(Mandatory = $true)][string]$Label,
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$Reason
    )

    return [pscustomobject]@{
        label = $Label
        url = $Url
        reachable = $false
        classification = 'skipped'
        statusCode = $null
        statusDescription = $null
        attempts = 0
        elapsedMs = 0
        lastError = $Reason
        checkedAt = (Get-Date).ToString('o')
    }
}

function Get-ExternalPreflightDiagnosticPath {
    param([Parameter(Mandatory = $true)][string]$TempRoot)

    return Join-Path $TempRoot 'external-preflight-diagnostic.json'
}

function Get-ExternalPreflightSummaryLines {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Summary,
        [string[]]$EntryNames = @('backend', 'frontend', 'cdp')
    )

    $lines = @()
    foreach ($entryName in $EntryNames) {
        $result = $Summary[$entryName]
        if ($null -eq $result) {
            continue
        }

        $lines += Format-HttpReachabilitySummaryLine -Result $result
    }

    return $lines
}

function Get-ExternalPreflightFailedCheckNames {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Summary,
        [string[]]$EntryNames = @('backend', 'frontend', 'cdp')
    )

    $failedChecks = @()
    foreach ($entryName in $EntryNames) {
        $result = $Summary[$entryName]
        if ($null -eq $result) {
            continue
        }

        if ((-not [bool]$result.reachable) -and $result.classification -ne 'skipped') {
            $failedChecks += $entryName
        }
    }

    return $failedChecks
}

function Get-ExternalPreflightCheckDetails {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Summary,
        [string[]]$EntryNames = @('backend', 'frontend', 'cdp')
    )

    $details = @()
    foreach ($entryName in $EntryNames) {
        $result = $Summary[$entryName]
        if ($null -eq $result) {
            continue
        }

        $details += [ordered]@{
            name = $entryName
            label = $result.label
            url = $result.url
            reachable = [bool]$result.reachable
            classification = $result.classification
            lastError = $result.lastError
        }
    }

    return $details
}

function Get-ExternalPreflightHints {
    param(
        [string[]]$FailedChecks,
        [bool]$RequireCdp,
        [int]$CdpPort
    )

    $hints = [System.Collections.Generic.List[string]]::new()

    if ($FailedChecks -contains 'backend') {
        $hints.Add('Start or expose the backend before rerunning the external smoke.')
    }

    if ($FailedChecks -contains 'frontend') {
        $hints.Add('Ensure the external frontend URL is reachable, or add -SelfStartServices if this host should start local services for comparison.')
    }

    if ($RequireCdp -and ($FailedChecks -contains 'cdp')) {
        $hints.Add(('Start or reuse Edge with --remote-debugging-port={0}, or rerun with -BootstrapExternalCdpSession on a localhost CDP URL.' -f $CdpPort))
    }

    if ($hints.Count -eq 0) {
        $hints.Add('Inspect external-preflight-diagnostic.json for the structured reachability results.')
    }

    return $hints
}

function New-ExternalPreflightFailureMessage {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$Summary,
        [Parameter(Mandatory = $true)][string]$DiagnosticPath,
        [bool]$RequireCdp,
        [int]$CdpPort
    )

    $failedChecks = @($Summary.failedChecks)
    if ($failedChecks.Count -eq 0) {
        $failedChecks = @(Get-ExternalPreflightFailedCheckNames -Summary $Summary)
    }

    if ($failedChecks.Count -eq 0) {
        return $null
    }

    $messageLines = @()
    $messageLines += ('External smoke preflight failed for: {0}.' -f ($failedChecks -join ', '))
    if (-not [string]::IsNullOrWhiteSpace([string]$Summary.overallClassification)) {
        $messageLines += ('Overall classification: {0}' -f $Summary.overallClassification)
    }
    if (-not [string]::IsNullOrWhiteSpace([string]$Summary.primaryBlockingCheck)) {
        $messageLines += ('Primary blocking check: {0}' -f $Summary.primaryBlockingCheck)
    }

    $summaryLines = @($Summary.summaryLines)
    if ($summaryLines.Count -eq 0) {
        $summaryLines = @(Get-ExternalPreflightSummaryLines -Summary $Summary)
    }
    $messageLines += $summaryLines
    $messageLines += ('Diagnostic: {0}' -f $DiagnosticPath)

    $hints = @($Summary.hints)
    if ($hints.Count -eq 0) {
        $hints = @(Get-ExternalPreflightHints -FailedChecks $failedChecks -RequireCdp $RequireCdp -CdpPort $CdpPort)
    }
    if ($hints.Count -gt 0) {
        $messageLines += 'Hints:'
        foreach ($hint in $hints) {
            $messageLines += ('- ' + $hint)
        }
    }

    return ($messageLines -join [Environment]::NewLine)
}

function Format-CdpDiagnosticSummary {
    param($Diagnostic)

    if ($null -eq $Diagnostic) {
        return ''
    }

    $lines = @()
    $lines += ('CDP diagnostic classification: {0}' -f $Diagnostic.classification)
    $lines += ('CDP diagnostic error: {0}' -f $Diagnostic.errorName)

    if (-not [string]::IsNullOrWhiteSpace([string]$Diagnostic.pageMode)) {
        $lines += ('CDP diagnostic page mode: {0}' -f $Diagnostic.pageMode)
    }

    if (-not [string]::IsNullOrWhiteSpace([string]$Diagnostic.pageStrategy)) {
        $lines += ('CDP diagnostic page strategy: {0}' -f $Diagnostic.pageStrategy)
    }

    if (-not [string]::IsNullOrWhiteSpace([string]$Diagnostic.bootstrapCommandOrder)) {
        $lines += ('CDP diagnostic bootstrap order: {0}' -f $Diagnostic.bootstrapCommandOrder)
    }

    if ($null -ne $Diagnostic.commandTimeoutMs -and [int]$Diagnostic.commandTimeoutMs -gt 0) {
        $lines += ('CDP diagnostic command timeout (ms): {0}' -f $Diagnostic.commandTimeoutMs)
    }

    if (-not [string]::IsNullOrWhiteSpace([string]$Diagnostic.fallbackFrom)) {
        $lines += ('CDP diagnostic fallback from: {0}' -f $Diagnostic.fallbackFrom)
    }

    if (-not [string]::IsNullOrWhiteSpace([string]$Diagnostic.hint)) {
        $lines += ('CDP diagnostic hint: {0}' -f $Diagnostic.hint)
    }

    return ($lines -join [Environment]::NewLine)
}

function Test-TraceTailPattern {
    param(
        [Parameter(Mandatory = $true)][string]$TraceTail,
        [Parameter(Mandatory = $true)][string]$Pattern
    )

    return [regex]::IsMatch($TraceTail, $Pattern, [System.Text.RegularExpressions.RegexOptions]::Multiline)
}

function Get-CdpBootstrapTimeoutMethodsFromTraceTail {
    param([string]$TraceTail)

    $methods = @()
    foreach ($method in @('Page.enable', 'Page.setLifecycleEventsEnabled', 'Runtime.enable')) {
        if (Test-TraceTailPattern -TraceTail $TraceTail -Pattern ('^.*cdp-command:timeout \{[^\r\n]*"method":"' + [regex]::Escape($method) + '"')) {
            $methods += $method
        }
    }

    return $methods
}

function Get-CdpDiagnosticFromTraceTail {
    param(
        [string]$TraceTail,
        [string]$RequestedPageStrategy = 'auto',
        [string]$RequestedBootstrapOrder = 'page-lifecycle-runtime',
        [int]$RequestedCommandTimeoutMs = 15000
    )

    if ([string]::IsNullOrWhiteSpace($TraceTail)) {
        return $null
    }

    $timedOutMethods = Get-CdpBootstrapTimeoutMethodsFromTraceTail -TraceTail $TraceTail
    if ($timedOutMethods.Count -eq 0) {
        return $null
    }

    $timedOutMethodSummary = $timedOutMethods -join ', '

    if ((Test-TraceTailPattern -TraceTail $TraceTail -Pattern 'cdp-target:open-strategy \{.*"strategy":"attached-session"') -or (Test-TraceTailPattern -TraceTail $TraceTail -Pattern 'smoke-page:opened \{.*"pageMode":"attached-session"')) {
        return [pscustomobject]@{
            classification = 'page_bootstrap_timeout_attached_session'
            errorName = 'CdpPageBootstrapPendingTimeout'
            pageMode = 'attached-session'
            pageStrategy = $RequestedPageStrategy
            bootstrapCommandOrder = $RequestedBootstrapOrder
            commandTimeoutMs = $RequestedCommandTimeoutMs
            fallbackFrom = $null
            hint = ('Explicit attached-session page bootstrap still stalls on {0} for this host. Compare command order or timeout only if we still need more host-level evidence.' -f $timedOutMethodSummary)
        }
    }

    if (Test-TraceTailPattern -TraceTail $TraceTail -Pattern 'cdp-target:json-new') {
        $classification = if ($RequestedPageStrategy -eq 'auto') { 'page_bootstrap_timeout_preempted' } else { 'page_bootstrap_timeout_json_new' }
        $hint = if ($RequestedPageStrategy -eq 'auto') {
            'Wrapper hit the total timeout before Node finished the fallback attached-session path. The trace already shows json-new bootstrap stalling, so the next run should only vary command timeout or bootstrap order if we still need stronger host evidence.'
        }
        else {
            ('Explicit json-new bootstrap still stalls on {0}; this remains upstream of the real page assertions.' -f $timedOutMethodSummary)
        }

        return [pscustomobject]@{
            classification = $classification
            errorName = 'CdpPageBootstrapPendingTimeout'
            pageMode = 'json-new'
            pageStrategy = $RequestedPageStrategy
            bootstrapCommandOrder = $RequestedBootstrapOrder
            commandTimeoutMs = $RequestedCommandTimeoutMs
            fallbackFrom = $null
            hint = $hint
        }
    }

    return $null
}

function Get-ErrorMessageChain {
    param([Parameter(Mandatory = $true)]$ErrorRecord)

    $messages = [System.Collections.Generic.List[string]]::new()
    $exception = $ErrorRecord.Exception
    while ($null -ne $exception) {
        if (-not [string]::IsNullOrWhiteSpace($exception.Message)) {
            $messages.Add($exception.Message)
        }
        $exception = $exception.InnerException
    }

    if ($messages.Count -eq 0 -and -not [string]::IsNullOrWhiteSpace([string]$ErrorRecord.ToString())) {
        $messages.Add([string]$ErrorRecord.ToString())
    }

    return ($messages | Select-Object -Unique) -join ' | '
}

function Test-IsServiceProviderFailureMessage {
    param([string]$Message)

    if ([string]::IsNullOrWhiteSpace($Message)) {
        return $false
    }

    return $Message -match '无法加载或初始化请求的服务提供程序|requested service provider|service provider could not be loaded or initialized'
}

function Test-IsCodexBundledNode {
    param([Parameter(Mandatory = $true)][string]$Candidate)

    $normalized = $Candidate.Replace('/', '\\').ToLowerInvariant()
    return $normalized.Contains('\\appdata\\local\\openai\\codex\\bin\\node.exe') -or $normalized.Contains('\\windowsapps\\openai.codex_')
}

function Test-UsableNodeExe {
    param([Parameter(Mandatory = $true)][string]$Candidate)

    if ([string]::IsNullOrWhiteSpace($Candidate)) {
        return $false
    }

    if (-not (Test-Path -LiteralPath $Candidate)) {
        return $false
    }

    if (Test-IsCodexBundledNode -Candidate $Candidate) {
        return $false
    }

    try {
        & $Candidate -e "console.log('node-ok')" *> $null
        return $true
    }
    catch {
        return $false
    }
}

function Resolve-NodeCommandPath {
    param([string[]]$Candidates)

    foreach ($candidate in $Candidates) {
        if ([string]::IsNullOrWhiteSpace([string]$candidate)) {
            continue
        }

        if (Test-UsableNodeExe -Candidate $candidate) {
            return (Resolve-Path -LiteralPath $candidate).ProviderPath
        }
    }

    $pathCommand = Get-Command -Name 'node' -ErrorAction SilentlyContinue
    if ($null -ne $pathCommand -and (Test-UsableNodeExe -Candidate $pathCommand.Source)) {
        return $pathCommand.Source
    }

    return $null
}

function Wait-Http {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [int]$TimeoutSeconds = 60
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        try {
            $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 500) {
                return
            }
        }
        catch {
        }

        Start-Sleep -Milliseconds 500
    }

    throw "Timed out waiting for $Url"
}

function Test-HttpEndpoint {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [int]$TimeoutSeconds = 3
    )

    try {
        $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec $TimeoutSeconds
        return ($response.StatusCode -ge 200 -and $response.StatusCode -lt 500)
    }
    catch {
        return $false
    }
}

function Get-ListeningTcpPorts {
    try {
        return [System.Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners() |
            ForEach-Object { $_.Port } |
            Select-Object -Unique
    }
    catch {
        return @()
    }
}

function Test-LoopbackBindCapability {
    param([string]$BindHost = '127.0.0.1')

    $listener = $null
    try {
        $listener = New-Object System.Net.Sockets.TcpListener ([System.Net.IPAddress]::Parse($BindHost)), 0
        $listener.Start()
        return $true
    }
    catch {
        return $false
    }
    finally {
        if ($null -ne $listener) {
            try {
                $listener.Stop()
            }
            catch {
            }
        }
    }
}

function Test-TcpPortAvailable {
    param(
        [Parameter(Mandatory = $true)][int]$Port,
        [string]$BindHost = '127.0.0.1',
        [int]$ServiceProviderRetryCount = 3,
        [int]$ServiceProviderRetryDelayMs = 250
    )

    if ((Get-ListeningTcpPorts) -contains $Port) {
        return $false
    }

    $lastServiceProviderMessage = ''

    for ($attempt = 1; $attempt -le [Math]::Max(1, $ServiceProviderRetryCount); $attempt++) {
        $listener = $null
        try {
            $listener = New-Object System.Net.Sockets.TcpListener ([System.Net.IPAddress]::Parse($BindHost)), $Port
            $listener.Start()
            return $true
        }
        catch [System.Net.Sockets.SocketException] {
            if (Test-IsServiceProviderFailureMessage -Message $_.Exception.Message) {
                $lastServiceProviderMessage = $_.Exception.Message
                if ($attempt -lt $ServiceProviderRetryCount) {
                    Start-Sleep -Milliseconds $ServiceProviderRetryDelayMs
                    continue
                }
            }
            else {
                return $false
            }
        }
        catch {
            $message = Get-ErrorMessageChain -ErrorRecord $_
            if (Test-IsServiceProviderFailureMessage -Message $message) {
                $lastServiceProviderMessage = $message
                if ($attempt -lt $ServiceProviderRetryCount) {
                    Start-Sleep -Milliseconds $ServiceProviderRetryDelayMs
                    continue
                }
            }
            else {
                return $false
            }
        }
        finally {
            if ($null -ne $listener) {
                try {
                    $listener.Stop()
                }
                catch {
                }
            }
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($lastServiceProviderMessage)) {
        $currentListeners = Get-ListeningTcpPorts
        if (-not ($currentListeners -contains $Port)) {
            $loopbackBindReady = Test-LoopbackBindCapability -BindHost $BindHost
            if ($loopbackBindReady) {
                Write-Host ("[INFO] Loopback bind probe for {0}:{1} hit a transient Windows service-provider initialization error; continuing based on active listener snapshot." -f $BindHost, $Port)
            }
            else {
                Write-Host ("[INFO] Loopback bind probe for {0}:{1} is currently unstable in this PowerShell session; continuing with active listener snapshot because the port is not in use. A real startup failure will still surface during process launch." -f $BindHost, $Port)
            }
            return $true
        }

        throw ("Loopback TCP probing on {0}:{1} is blocked by Windows service-provider initialization after {2} attempts, and the port already appears in the active listener snapshot. Treat this as a host networking blocker instead of a smoke regression. Last error: {3}" -f $BindHost, $Port, $ServiceProviderRetryCount, $lastServiceProviderMessage)
    }

    return $false
}

function Resolve-FreeTcpPort {
    param(
        [Parameter(Mandatory = $true)][int]$PreferredPort,
        [string]$BindHost = '127.0.0.1',
        [int]$SearchWindow = 20
    )

    if (Test-TcpPortAvailable -Port $PreferredPort -BindHost $BindHost) {
        return $PreferredPort
    }

    for ($candidate = $PreferredPort + 1; $candidate -lt ($PreferredPort + $SearchWindow); $candidate++) {
        if (Test-TcpPortAvailable -Port $candidate -BindHost $BindHost) {
            return $candidate
        }
    }

    throw "Unable to find an available TCP port starting at $PreferredPort on $BindHost"
}

function Resolve-CdpPort {
    param(
        [Parameter(Mandatory = $true)][int]$PreferredPort,
        [string]$BindHost = '127.0.0.1',
        [int]$SearchWindow = 20
    )

    $preferredBaseUrl = "http://$BindHost`:$PreferredPort"
    if (Test-HttpEndpoint -Url "$preferredBaseUrl/json/version") {
        return [pscustomobject]@{
            Port = $PreferredPort
            ReuseExisting = $true
        }
    }

    if (Test-TcpPortAvailable -Port $PreferredPort -BindHost $BindHost) {
        return [pscustomobject]@{
            Port = $PreferredPort
            ReuseExisting = $false
        }
    }

    for ($candidate = $PreferredPort + 1; $candidate -lt ($PreferredPort + $SearchWindow); $candidate++) {
        $candidateBaseUrl = "http://$BindHost`:$candidate"
        if (Test-HttpEndpoint -Url "$candidateBaseUrl/json/version") {
            return [pscustomobject]@{
                Port = $candidate
                ReuseExisting = $true
            }
        }

        if (Test-TcpPortAvailable -Port $candidate -BindHost $BindHost) {
            return [pscustomobject]@{
                Port = $candidate
                ReuseExisting = $false
            }
        }
    }

    throw "Unable to resolve a usable CDP port starting at $PreferredPort on $BindHost"
}

function Write-ConfigJson {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$StaticDir,
        [Parameter(Mandatory = $true)][string]$DbPath,
        [Parameter(Mandatory = $true)][int]$Port
    )

    $configObject = [ordered]@{
        server = [ordered]@{
            # Use the project default bind host for self-start smoke runs.
            # Health checks still go through 127.0.0.1, but binding to 0.0.0.0
            # avoids transient Windows loopback-provider failures seen in some sessions.
            host = '0.0.0.0'
            port = $Port
            static_dir = $StaticDir
        }
        database = [ordered]@{
            type = 'sqlite'
            path = $DbPath
        }
        log = [ordered]@{
            level = 'info'
        }
    }

    $config = $configObject | ConvertTo-Json -Depth 4
    [System.IO.File]::WriteAllText($Path, $config, (New-Object System.Text.UTF8Encoding($false)))
}

function Start-BackendSelfStartProcess {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$StaticDir,
        [Parameter(Mandatory = $true)][string]$TempRoot,
        [Parameter(Mandatory = $true)][int]$PreferredPort,
        [string]$ResolvedBackendBin,
        [string]$ResolvedGoPath,
        [int]$MaxAttempts = 3
    )

    $backendEnv = @{
        OCTOPUS_ADMIN_USERNAME = 'admin'
        OCTOPUS_ADMIN_PASSWORD = 'admin'
    }

    $attemptErrors = [System.Collections.Generic.List[string]]::new()

    for ($attempt = 1; $attempt -le [Math]::Max(1, $MaxAttempts); $attempt++) {
        $candidatePort = if ($attempt -eq 1) { $PreferredPort } else { Resolve-FreeTcpPort -PreferredPort ($PreferredPort + $attempt - 1) }
        $configPath = Join-Path $TempRoot ("config.$attempt.json")
        $dbPath = Join-Path $TempRoot ("octopus.$attempt.db")
        $backendStdout = Join-Path $TempRoot ("backend.$attempt.stdout.log")
        $backendStderr = Join-Path $TempRoot ("backend.$attempt.stderr.log")

        Write-ConfigJson -Path $configPath -StaticDir $StaticDir -DbPath $dbPath -Port $candidatePort

        if ($ResolvedBackendBin) {
            $backendProc = Start-LoggedProcess -FilePath $ResolvedBackendBin -ArgumentList @('start', '--config', $configPath) -WorkingDirectory $RepoRoot -StdoutPath $backendStdout -StderrPath $backendStderr -ProcessEnvironment $backendEnv
        }
        else {
            $backendProc = Start-LoggedProcess -FilePath $ResolvedGoPath -ArgumentList @('run', 'main.go', 'start', '--config', $configPath) -WorkingDirectory $RepoRoot -StdoutPath $backendStdout -StderrPath $backendStderr -ProcessEnvironment $backendEnv
        }

        Start-Sleep -Milliseconds 1200
        if ($backendProc.HasExited) {
            $backendError = ''
            if (Test-Path -LiteralPath $backendStderr) {
                $backendError = (Get-Content -LiteralPath $backendStderr -Raw)
            }
            if ([string]::IsNullOrWhiteSpace($backendError) -and (Test-Path -LiteralPath $backendStdout)) {
                $backendError = (Get-Content -LiteralPath $backendStdout -Raw)
            }

            $trimmedError = $backendError.Trim()
            if (Test-IsServiceProviderFailureMessage -Message $trimmedError -and $attempt -lt $MaxAttempts) {
                $attemptErrors.Add(("attempt {0}: transient backend self-start service-provider failure on port {1}" -f $attempt, $candidatePort))
                Write-Host ("[INFO] Backend self-start hit a transient Windows service-provider initialization error on port {0}; retrying with a fresh temp config." -f $candidatePort)
                Start-Sleep -Milliseconds 500
                continue
            }

            throw ("Backend self-start exited early. {0}" -f $trimmedError)
        }

        return [pscustomobject]@{
            Process = $backendProc
            Port = $candidatePort
            ConfigPath = $configPath
            DbPath = $dbPath
            StdoutPath = $backendStdout
            StderrPath = $backendStderr
            Attempt = $attempt
            AttemptErrors = @($attemptErrors)
        }
    }

    $summary = if ($attemptErrors.Count -gt 0) { ($attemptErrors -join '; ') } else { 'unknown backend self-start failure' }
    throw ("Backend self-start did not produce a running process after {0} attempts. {1}" -f $MaxAttempts, $summary)
}

function Start-LoggedProcess {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter()][string[]]$ArgumentList = @(),
        [Parameter(Mandatory = $true)][string]$WorkingDirectory,
        [Parameter(Mandatory = $true)][string]$StdoutPath,
        [Parameter(Mandatory = $true)][string]$StderrPath,
        [hashtable]$ProcessEnvironment = @{},
        [ValidateSet('Hidden', 'Normal', 'Minimized', 'Maximized')][string]$WindowStyle = 'Hidden'
    )

    $previousEnvironment = @{}
    foreach ($pair in $ProcessEnvironment.GetEnumerator()) {
        $previousEnvironment[$pair.Key] = [Environment]::GetEnvironmentVariable($pair.Key, 'Process')
        [Environment]::SetEnvironmentVariable($pair.Key, [string]$pair.Value, 'Process')
    }

    $startParams = @{
        FilePath = $FilePath
        ArgumentList = $ArgumentList
        WorkingDirectory = $WorkingDirectory
        RedirectStandardOutput = $StdoutPath
        RedirectStandardError = $StderrPath
        PassThru = $true
        WindowStyle = $WindowStyle
    }

    try {
        return Start-Process @startParams
    }
    finally {
        foreach ($pair in $previousEnvironment.GetEnumerator()) {
            [Environment]::SetEnvironmentVariable($pair.Key, $pair.Value, 'Process')
        }
    }
}

function Test-CdpEndpointReady {
    param(
        [Parameter(Mandatory = $true)][string]$BaseUrl,
        [int]$TimeoutSeconds = 3
    )

    return (Test-HttpEndpoint -Url ((Normalize-Url -Value $BaseUrl) + '/json/version') -TimeoutSeconds $TimeoutSeconds)
}

function Resolve-PortHintFromUrl {
    param(
        [Parameter(Mandatory = $true)][string]$BaseUrl,
        [int]$FallbackPort
    )

    try {
        $uri = [System.Uri](Normalize-Url -Value $BaseUrl)
        if ($uri.Port -gt 0) {
            return $uri.Port
        }
    }
    catch {
    }

    return $FallbackPort
}

function Test-LocalhostBaseUrl {
    param([Parameter(Mandatory = $true)][string]$BaseUrl)

    try {
        $uri = [System.Uri](Normalize-Url -Value $BaseUrl)
        return @('127.0.0.1', 'localhost', '::1').Contains($uri.Host)
    }
    catch {
        return $false
    }
}

function Invoke-ExternalCdpBootstrapHelper {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$TempRoot,
        [Parameter(Mandatory = $true)][string]$BaseUrl,
        [Parameter(Mandatory = $true)][int]$FallbackPort,
        [string]$Browser,
        [ValidateSet('default', 'relaxed', 'headed-relaxed')][string]$Preset,
        [ValidateSet('temp-random', 'workspace-fixed')][string]$ProfileStrategy
    )

    if (-not (Test-LocalhostBaseUrl -BaseUrl $BaseUrl)) {
        throw ("BootstrapExternalCdpSession only supports localhost CDP URLs. Received {0}." -f (Normalize-Url -Value $BaseUrl))
    }

    $bootstrapScript = Join-Path $RepoRoot 'scripts\bootstrap-edge-cdp.ps1'
    if (-not (Test-Path -LiteralPath $bootstrapScript)) {
        throw ("External CDP bootstrap helper script was not found at {0}." -f $bootstrapScript)
    }

    $bootstrapJsonPath = Join-Path $TempRoot 'external-edge-cdp-bootstrap.json'
    $bootstrapPort = Resolve-PortHintFromUrl -BaseUrl $BaseUrl -FallbackPort $FallbackPort
    $bootstrapArgs = @{
        Port = $bootstrapPort
        EdgeLaunchPreset = $Preset
        EdgeProfileStrategy = $ProfileStrategy
        ReadyTimeoutSeconds = 30
        StableReadySeconds = 3
        OutputJsonPath = $bootstrapJsonPath
    }

    if (-not [string]::IsNullOrWhiteSpace($Browser)) {
        $bootstrapArgs.BrowserPath = $Browser
    }

    $bootstrapOutput = & $bootstrapScript @bootstrapArgs
    if ($bootstrapOutput) {
        $bootstrapOutput | Out-Host
    }

    $summary = Get-JsonFileContent -Path $bootstrapJsonPath
    if ($null -eq $summary) {
        throw ("External CDP bootstrap helper did not produce a readable summary at {0}." -f $bootstrapJsonPath)
    }

    return $summary
}

function Assert-ExternalCdpEndpointReady {
    param(
        [Parameter(Mandatory = $true)][string]$BaseUrl,
        [Parameter(Mandatory = $true)][int]$Port
    )

    if (Test-CdpEndpointReady -BaseUrl $BaseUrl) {
        return
    }

    $hintPort = Resolve-PortHintFromUrl -BaseUrl $BaseUrl -FallbackPort $Port

    throw (
        "External CDP mode could not reach an existing Edge remote debugging endpoint at {0}. Start Edge separately with --remote-debugging-port={1} or rerun with -Mode self-start." -f (Normalize-Url -Value $BaseUrl), $hintPort
    )
}

function Invoke-ExternalDependencyPreflight {
    param(
        [Parameter(Mandatory = $true)][string]$TempRoot,
        [Parameter(Mandatory = $true)][string]$FrontendUrl,
        [Parameter(Mandatory = $true)][string]$BackendUrl,
        [Parameter(Mandatory = $true)][string]$CdpUrl,
        [Parameter(Mandatory = $true)][int]$BackendTimeoutSeconds,
        [Parameter(Mandatory = $true)][int]$FrontendTimeoutSeconds,
        [Parameter(Mandatory = $true)][int]$CdpTimeoutSeconds,
        [Parameter(Mandatory = $true)][int]$CdpPort,
        [switch]$RequireCdp
    )

    $summary = [ordered]@{
        schemaVersion = 2
        checkedAt = (Get-Date).ToString('o')
        requireCdp = [bool]$RequireCdp
        failedChecks = @()
        skippedChecks = @()
        overallClassification = 'ok'
        primaryBlockingCheck = $null
        summaryLines = @()
        hints = @()
        checkDetails = @()
        frontend = $null
        backend = $null
        cdp = $null
    }

    $diagnosticPath = Get-ExternalPreflightDiagnosticPath -TempRoot $TempRoot

    $backendResult = Get-HttpReachabilityResult -Label 'backend healthcheck' -Url "$BackendUrl/healthz" -TimeoutSeconds $BackendTimeoutSeconds
    Set-ExternalPreflightEntry -Summary $summary -EntryName 'backend' -Result $backendResult

    $frontendResult = Get-HttpReachabilityResult -Label 'frontend shell' -Url $FrontendUrl -TimeoutSeconds $FrontendTimeoutSeconds
    Set-ExternalPreflightEntry -Summary $summary -EntryName 'frontend' -Result $frontendResult

    if ($RequireCdp) {
        $cdpResult = Get-HttpReachabilityResult -Label 'Edge CDP version endpoint' -Url "$CdpUrl/json/version" -TimeoutSeconds $CdpTimeoutSeconds
        Set-ExternalPreflightEntry -Summary $summary -EntryName 'cdp' -Result $cdpResult
    }
    else {
        $summary.cdp = New-ExternalPreflightSkippedEntry -Label 'Edge CDP version endpoint' -Url "$CdpUrl/json/version" -Reason 'skipped until the external service preflight passes or CDP is explicitly required'
    }

    $summary.failedChecks = @(Get-ExternalPreflightFailedCheckNames -Summary $summary)
    $summary.checkDetails = @(Get-ExternalPreflightCheckDetails -Summary $summary)
    $summary.summaryLines = @(Get-ExternalPreflightSummaryLines -Summary $summary)
    $summary.skippedChecks = @($summary.checkDetails | Where-Object { $_.classification -eq 'skipped' } | ForEach-Object { $_.name })
    if ($summary.failedChecks.Count -gt 0) {
        $summary.overallClassification = 'preflight_failed'
        $summary.primaryBlockingCheck = $summary.failedChecks[0]
        $summary.hints = @(Get-ExternalPreflightHints -FailedChecks $summary.failedChecks -RequireCdp:$RequireCdp -CdpPort $CdpPort)
    }

    Write-JsonArtifact -Path $diagnosticPath -Value $summary

    $failureMessage = New-ExternalPreflightFailureMessage -Summary $summary -DiagnosticPath $diagnosticPath -RequireCdp:$RequireCdp -CdpPort $CdpPort
    if (-not [string]::IsNullOrWhiteSpace($failureMessage)) {
        throw $failureMessage
    }

    return $summary
}

function Stop-ProcessTree {
    param([System.Diagnostics.Process]$Process)

    if ($null -eq $Process) {
        return
    }

    try {
        Stop-Process -Id $Process.Id -Force -ErrorAction Stop
    }
    catch {
    }

    try {
        Get-CimInstance Win32_Process -Filter "ParentProcessId = $($Process.Id)" -ErrorAction Stop | ForEach-Object {
            try {
                Stop-Process -Id $_.ProcessId -Force -ErrorAction Stop
            }
            catch {
            }
        }
    }
    catch {
    }
}

function Invoke-LoggedProcessWait {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter()][string[]]$ArgumentList = @(),
        [Parameter(Mandatory = $true)][string]$WorkingDirectory,
        [Parameter(Mandatory = $true)][string]$StdoutPath,
        [Parameter(Mandatory = $true)][string]$StderrPath,
        [int]$TimeoutSeconds = 120,
        [string]$Description = 'Process',
        [hashtable]$ProcessEnvironment = @{}
    )

    $process = Start-LoggedProcess -FilePath $FilePath -ArgumentList $ArgumentList -WorkingDirectory $WorkingDirectory -StdoutPath $StdoutPath -StderrPath $StderrPath -ProcessEnvironment $ProcessEnvironment

    try {
        try {
            Wait-Process -Id $process.Id -Timeout ([Math]::Max($TimeoutSeconds, 1)) -ErrorAction Stop
            $exited = $true
        }
        catch [System.TimeoutException] {
            $exited = $false
        }

        if (-not $exited) {
            Stop-ProcessTree -Process $process
            Start-Sleep -Milliseconds 300
            $process.Refresh()
            $stdoutTail = Get-LogTail -Path $StdoutPath
            $stderrTail = Get-LogTail -Path $StderrPath
            throw ("{0} timed out after {1}s.`nLog files: stdout={2}`nstderr={3}`nSTDOUT tail:`n{4}`nSTDERR tail:`n{5}" -f $Description, $TimeoutSeconds, $StdoutPath, $StderrPath, $stdoutTail, $stderrTail)
        }

        $process.WaitForExit()
        $stdoutTail = Get-LogTail -Path $StdoutPath
        $stderrTail = Get-LogTail -Path $StderrPath
        $process.Refresh()
        $exitCode = [int]$process.ExitCode
        if ($exitCode -ne 0) {
            throw ("{0} exited with code {1}.`nLog files: stdout={2}`nstderr={3}`nSTDOUT tail:`n{4}`nSTDERR tail:`n{5}" -f $Description, $exitCode, $StdoutPath, $StderrPath, $stdoutTail, $stderrTail)
        }

        return [pscustomobject]@{
            StdoutPath = $StdoutPath
            StderrPath = $StderrPath
            StdoutTail = $stdoutTail
            StderrTail = $stderrTail
        }
    }
    finally {
        if (-not $process.HasExited) {
            Stop-ProcessTree -Process $process
        }
    }
}

function Assert-NodeSmokeSucceeded {
    param(
        [Parameter(Mandatory = $true)][pscustomobject]$Result,
        [Parameter(Mandatory = $true)][string]$ExpectedResult
    )

    $stdout = Read-LogContent -Path $Result.StdoutPath
    $stderr = Read-LogContent -Path $Result.StderrPath

    if ($stdout -notmatch [regex]::Escape($ExpectedResult)) {
        throw ("Node {0} smoke did not emit expected success marker '{1}'.`nLog files: stdout={2}`nstderr={3}`nSTDOUT tail:`n{4}`nSTDERR tail:`n{5}" -f $SmokeLabel, $ExpectedResult, $Result.StdoutPath, $Result.StderrPath, $Result.StdoutTail, $Result.StderrTail)
    }

    if ($stderr -match '(?m)^([A-Za-z][A-Za-z0-9]*(Error|Exception)|AssertionError|Error|SyntaxError|TypeError|ReferenceError|RangeError):|\bat async main\b|\bprocess\.exit\b') {
        throw ("Node {0} smoke wrote an error-like stderr despite the success marker.`nLog files: stdout={1}`nstderr={2}`nSTDOUT tail:`n{3}`nSTDERR tail:`n{4}" -f $SmokeLabel, $Result.StdoutPath, $Result.StderrPath, $Result.StdoutTail, $Result.StderrTail)
    }
}

function Resolve-EdgeProfile {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$TempRoot,
        [ValidateSet('temp-random', 'workspace-fixed')][string]$Strategy,
        [ValidateSet('default', 'relaxed', 'headed-relaxed')][string]$Preset
    )

    if ($Strategy -eq 'workspace-fixed') {
        $profileRoot = Join-Path $RepoRoot ('.tools\verify-' + $smokeSlug + '\edge-profile')
        $profilePath = Join-Path $profileRoot $Preset
    }
    else {
        $profileRoot = Join-Path $TempRoot 'edge-profile-root'
        $profilePath = Join-Path $profileRoot ([guid]::NewGuid().ToString('N'))
    }

    New-Item -ItemType Directory -Path $profileRoot -Force | Out-Null
    New-Item -ItemType Directory -Path $profilePath -Force | Out-Null

    return [pscustomobject]@{
        Root = $profileRoot
        Path = $profilePath
    }
}

function Get-EdgeLaunchArguments {
    param(
        [Parameter(Mandatory = $true)][int]$RemoteDebuggingPort,
        [Parameter(Mandatory = $true)][string]$UserDataDir,
        [ValidateSet('default', 'relaxed', 'headed-relaxed')][string]$Preset
    )

    $useHeadless = $Preset -ne 'headed-relaxed'
    $arguments = @(
        "--remote-debugging-port=$RemoteDebuggingPort",
        "--user-data-dir=$UserDataDir",
        '--disable-extensions',
        '--disable-component-extensions-with-background-pages',
        '--disable-background-networking',
        '--disable-breakpad',
        '--disable-component-update',
        '--disable-gpu',
        '--no-sandbox',
        '--no-service-autorun',
        '--no-first-run',
        '--no-default-browser-check'
    )

    if ($useHeadless) {
        $arguments = @('--headless=new') + $arguments
    }

    if ($Preset -eq 'relaxed' -or $Preset -eq 'headed-relaxed') {
        $arguments += @(
            '--disable-crash-reporter',
            '--disable-crashpad-for-testing',
            '--disable-default-apps',
            '--disable-sync',
            '--disable-features=msEdgeEntraConnectedBrowser,EdgeWalletCheckoutUi,RendererCodeIntegrity,OptimizationGuideModelDownloading,CalculateNativeWinOcclusion,NetworkServiceSandbox',
            '--disk-cache-size=1',
            '--media-cache-size=1',
            '--disable-gpu-shader-disk-cache',
            '--disable-http-cache',
            '--metrics-recording-only',
            '--noerrdialogs',
            '--enable-logging=stderr',
            '--v=1'
        )
    }
    else {
        $arguments += @(
            '--disable-features=msEdgeEntraConnectedBrowser,EdgeWalletCheckoutUi,RendererCodeIntegrity,NetworkServiceSandbox'
        )
    }

    $arguments += 'about:blank'
    return $arguments
}

function Get-EdgeWindowStyle {
    param(
        [ValidateSet('default', 'relaxed', 'headed-relaxed')][string]$Preset
    )

    if ($Preset -eq 'headed-relaxed') {
        return 'Normal'
    }

    return 'Hidden'
}

$repoRoot = Get-RepoRoot
$tempRoot = Join-Path $env:TEMP ('octopus-' + $smokeSlug + '-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempRoot | Out-Null
$staticDir = Resolve-StaticDir -RepoRoot $repoRoot

$resolvedNodePath = Resolve-NodeCommandPath -Candidates @(
    $NodePath,
    $env:OCTOPUS_UI_SMOKE_NODE,
    $env:NODEEXE,
    $(if (-not [string]::IsNullOrWhiteSpace($env:NODE_BIN)) { Join-Path $env:NODE_BIN 'node.exe' } else { $null }),
    $env:OCTOPUS_NODE_EXE,
    $(if (-not [string]::IsNullOrWhiteSpace($env:OCTOPUS_NODE_BIN)) { Join-Path $env:OCTOPUS_NODE_BIN 'node.exe' } else { $null }),
    $(if (-not [string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) { Join-Path $env:LOCALAPPDATA 'Programs\nodejs\node.exe' } else { $null }),
    'D:\gptm\live-2d\node\node.exe',
    'D:\gol1\node.exe'
)
if (-not $resolvedNodePath) {
    throw ("Unable to resolve Node.js executable for {0} smoke." -f $SmokeLabel)
}

$resolvedGoPath = Resolve-CommandPath -Candidates @(
    $GoPath,
    $env:OCTOPUS_UI_SMOKE_GO,
    (Join-Path $repoRoot '.tools\go\go\bin\go.exe')
) -CommandName 'go'

$resolvedBackendBin = Resolve-CommandPath -Candidates @(
    $BackendBin,
    $env:OCTOPUS_UI_SMOKE_BACKEND_BIN,
    (Join-Path $repoRoot 'build\octopus-smoke.exe')
) -CommandName $null

$resolvedBrowserPath = Resolve-CommandPath -Candidates @(
    $Browser,
    $env:OCTOPUS_UI_SMOKE_BROWSER_PATH,
    'C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe',
    'C:\Program Files\Microsoft\Edge\Application\msedge.exe'
) -CommandName $null

$npxCliScript = Join-Path (Split-Path -Parent $resolvedNodePath) 'node_modules\npm\bin\npx-cli.js'
$effectiveBackendPort = $BackendPort
$effectiveFrontendPort = $FrontendPort
$effectiveCdpPort = $CdpPort
$reuseExistingCdpSession = $false
$bootstrapLocalServices = ($Mode -eq 'self-start') -or ($Mode -eq 'external' -and $SelfStartServices)
$requireExternalCdpPreflight = ($Driver -eq 'cdp') -and ($RequireExternalCdpPreflight -or -not $BootstrapExternalCdpSession)

if ($bootstrapLocalServices) {
    $effectiveBackendPort = Resolve-FreeTcpPort -PreferredPort $BackendPort
    $effectiveFrontendPort = Resolve-FreeTcpPort -PreferredPort $FrontendPort
    if ($Mode -eq 'self-start' -and $Driver -eq 'cdp' -and -not $CdpUrl) {
        $cdpResolution = Resolve-CdpPort -PreferredPort $CdpPort
        $effectiveCdpPort = $cdpResolution.Port
        $reuseExistingCdpSession = [bool]$cdpResolution.ReuseExisting
    }
}

$requestedFrontendBaseUrl = Normalize-Url -Value $(if ($FrontendUrl) { $FrontendUrl } else { "http://127.0.0.1:$FrontendPort" })
$backendBaseUrl = Normalize-Url -Value $(if ($BackendUrl) { $BackendUrl } else { "http://127.0.0.1:$effectiveBackendPort" })
$cdpBaseUrl = Normalize-Url -Value $(if ($CdpUrl) { $CdpUrl } else { "http://127.0.0.1:$effectiveCdpPort" })
$frontendBaseUrl = if ($bootstrapLocalServices) { $backendBaseUrl } else { $requestedFrontendBaseUrl }

$checkOnlyArgs = @($NodeSmokeScript, '--check-only')

if ($Driver -eq 'cdp') {
    $env:OCTOPUS_UI_SMOKE_CDP_URL = $cdpBaseUrl
    $env:OCTOPUS_UI_SMOKE_CDP_COMMAND_TIMEOUT_MS = [string]$CdpCommandTimeoutMs
    $env:OCTOPUS_UI_SMOKE_EDGE_LAUNCH_PRESET = $EdgeLaunchPreset
    $env:OCTOPUS_UI_SMOKE_EDGE_PROFILE_STRATEGY = $EdgeProfileStrategy
    $env:OCTOPUS_UI_SMOKE_CDP_PAGE_BOOTSTRAP_STRATEGY = $CdpPageBootstrapStrategy
    $env:OCTOPUS_UI_SMOKE_CDP_BOOTSTRAP_COMMAND_ORDER = $CdpBootstrapCommandOrder
}

if ($Mode -eq 'check-only') {
    Write-Step 'Check-only summary'
    Write-Host ("Driver: $Driver")
    Write-Host ("Node: $resolvedNodePath")
    Write-Host ("Go: $resolvedGoPath")
    Write-Host ("Backend binary: $resolvedBackendBin")
    if ($Driver -eq 'cdp') {
        Write-Host ("Browser: $resolvedBrowserPath")
        Write-Host ("CDP URL: $cdpBaseUrl")
        Write-Host ("Edge launch preset: $EdgeLaunchPreset")
        Write-Host ("Edge profile strategy: $EdgeProfileStrategy")
        Write-Host ("CDP command timeout (ms): $CdpCommandTimeoutMs")
        Write-Host ("CDP page bootstrap strategy: $CdpPageBootstrapStrategy")
        Write-Host ("CDP bootstrap command order: $CdpBootstrapCommandOrder")
        Write-Host ("Explicit external CDP bootstrap helper: {0}" -f $(if ($BootstrapExternalCdpSession) { 'enabled' } else { 'disabled' }))
        Write-Host ("Explicit external CDP preflight requirement: {0}" -f $(if ($RequireExternalCdpPreflight) { 'enabled' } else { 'disabled' }))
        Write-Host ("External mode initial CDP preflight: {0}" -f $(if ($requireExternalCdpPreflight) { 'required' } else { 'skipped until the service preflight passes or the flag is enabled' }))
        Write-Host ("External mode local service bootstrap option: {0}" -f $(if ($SelfStartServices) { 'enabled' } else { 'disabled' }))
        Write-Host ("External CDP mode requires an already-running Edge remote debugging endpoint at {0}." -f $cdpBaseUrl)
        if ($BootstrapExternalCdpSession) {
            Write-Host ("When enabled in external mode, the wrapper will call scripts/bootstrap-edge-cdp.ps1 to start or reuse a local Edge endpoint before preflight.")
        }
        Write-Host ("Self-start CDP mode may launch Edge automatically when no endpoint is reachable.")
    }
    if ($Mode -eq 'external') {
        Write-Host ("External mode local service bootstrap: {0}" -f $(if ($SelfStartServices) { 'enabled' } else { 'disabled' }))
        Write-Host ('External preflight diagnostic artifact: {0}' -f (Join-Path $tempRoot 'external-preflight-diagnostic.json'))
    }
    Write-Host ("Frontend URL: $frontendBaseUrl")
    Write-Host ("Backend URL: $backendBaseUrl")
    Write-Host ("Node smoke timeout (seconds): $NodeSmokeTimeoutSeconds")
    if ($bootstrapLocalServices) {
        Write-Host ("Local service port resolution: backend=$effectiveBackendPort frontend=$effectiveFrontendPort cdp=$effectiveCdpPort")
    }
    & $resolvedNodePath @checkOnlyArgs
    Write-StableSmokeDiagnosticPreview -RepoRoot $repoRoot -SmokeLabel $SmokeLabel
    exit $LASTEXITCODE
}

$processes = @()
$verificationSucceeded = $false
$externalCdpBootstrapSummary = $null
$externalPreflightSummary = $null

try {
    if ($bootstrapLocalServices) {
        if ($Mode -eq 'self-start') {
            Write-Step ("Starting backend and frontend for {0} smoke" -f $SmokeLabel)
        }
        else {
            Write-Step 'Starting local backend and frontend for external CDP smoke'
        }

        if ($effectiveBackendPort -ne $BackendPort) {
            Write-Host ("Backend port {0} is busy; falling back to {1}." -f $BackendPort, $effectiveBackendPort)
        }

        if ($Driver -eq 'cdp' -and $effectiveCdpPort -ne $CdpPort) {
            Write-Host ("CDP port {0} is unavailable; using {1}." -f $CdpPort, $effectiveCdpPort)
        }

        if (-not $resolvedBackendBin -and -not $resolvedGoPath) {
            throw 'Neither backend smoke binary nor Go toolchain is available.'
        }

        $backendStart = Start-BackendSelfStartProcess -RepoRoot $repoRoot -StaticDir $staticDir -TempRoot $tempRoot -PreferredPort $effectiveBackendPort -ResolvedBackendBin $resolvedBackendBin -ResolvedGoPath $resolvedGoPath
        $backendProc = $backendStart.Process
        $effectiveBackendPort = [int]$backendStart.Port
        $backendBaseUrl = Normalize-Url -Value "http://127.0.0.1:$effectiveBackendPort"
        $frontendBaseUrl = $backendBaseUrl
        $processes += $backendProc

        if ($backendStart.Attempt -gt 1) {
            Write-Host ("[INFO] Backend self-start stabilized on attempt {0} using port {1}." -f $backendStart.Attempt, $effectiveBackendPort)
        }

        Wait-Http -Url "$backendBaseUrl/healthz" -TimeoutSeconds 60
        Wait-Http -Url $frontendBaseUrl -TimeoutSeconds 90
    }
    else {
        Write-Step 'Verifying external backend and frontend'
        $externalPreflightSummary = Invoke-ExternalDependencyPreflight -TempRoot $tempRoot -FrontendUrl $frontendBaseUrl -BackendUrl $backendBaseUrl -CdpUrl $cdpBaseUrl -BackendTimeoutSeconds 30 -FrontendTimeoutSeconds 30 -CdpTimeoutSeconds 10 -CdpPort $effectiveCdpPort -RequireCdp:$requireExternalCdpPreflight
        Write-Host ('External backend preflight passed: {0}' -f $externalPreflightSummary.backend.url)
        Write-Host ('External frontend preflight passed: {0}' -f $externalPreflightSummary.frontend.url)
        if ($requireExternalCdpPreflight) {
            Write-Host ('External CDP preflight passed: {0}' -f $externalPreflightSummary.cdp.url)
        }
    }

    if ($Driver -eq 'cdp' -and $Mode -eq 'external' -and $BootstrapExternalCdpSession) {
        Write-Step 'Bootstrapping external Edge CDP endpoint'
        $externalCdpBootstrapSummary = Invoke-ExternalCdpBootstrapHelper -RepoRoot $repoRoot -TempRoot $tempRoot -BaseUrl $cdpBaseUrl -FallbackPort $effectiveCdpPort -Browser $resolvedBrowserPath -Preset $EdgeLaunchPreset -ProfileStrategy $EdgeProfileStrategy
        Write-Host ("External CDP helper result: {0} at {1}" -f $externalCdpBootstrapSummary.result, $externalCdpBootstrapSummary.jsonVersionUrl)

        if (-not [bool]$externalCdpBootstrapSummary.reusedExisting -and $null -ne $externalCdpBootstrapSummary.pid) {
            try {
                $processes += Get-Process -Id ([int]$externalCdpBootstrapSummary.pid) -ErrorAction Stop
            }
            catch {
            }
        }
    }

    if ($Driver -eq 'cdp' -and $Mode -eq 'external') {
        Write-Step 'Preflighting external Edge CDP endpoint'
        $externalPreflightSummary = Invoke-ExternalDependencyPreflight -TempRoot $tempRoot -FrontendUrl $frontendBaseUrl -BackendUrl $backendBaseUrl -CdpUrl $cdpBaseUrl -BackendTimeoutSeconds 5 -FrontendTimeoutSeconds 5 -CdpTimeoutSeconds 10 -CdpPort $effectiveCdpPort -RequireCdp
        Write-Host ("External Edge CDP endpoint is reachable at {0}." -f $cdpBaseUrl)
    }

    if ($Driver -eq 'cdp') {
        Write-Step 'Preparing Edge CDP session'
        $cdpReady = Test-CdpEndpointReady -BaseUrl $cdpBaseUrl

        if (-not $cdpReady) {
            if ($Mode -ne 'self-start') {
                Assert-ExternalCdpEndpointReady -BaseUrl $cdpBaseUrl -Port $effectiveCdpPort
            }

            if (-not $resolvedBrowserPath) {
                throw 'Unable to resolve Edge executable for CDP smoke.'
            }

            $browserStdout = Join-Path $tempRoot 'edge.stdout.log'
            $browserStderr = Join-Path $tempRoot 'edge.stderr.log'
            $browserProfile = Resolve-EdgeProfile -RepoRoot $repoRoot -TempRoot $tempRoot -Strategy $EdgeProfileStrategy -Preset $EdgeLaunchPreset
            $browserArgs = Get-EdgeLaunchArguments -RemoteDebuggingPort $effectiveCdpPort -UserDataDir $browserProfile.Path -Preset $EdgeLaunchPreset
            $browserWindowStyle = Get-EdgeWindowStyle -Preset $EdgeLaunchPreset
            Write-Host ("Launching Edge preset '{0}' with profile strategy '{1}' at {2}" -f $EdgeLaunchPreset, $EdgeProfileStrategy, $browserProfile.Path)
            $browserProc = Start-LoggedProcess -FilePath $resolvedBrowserPath -ArgumentList $browserArgs -WorkingDirectory $repoRoot -StdoutPath $browserStdout -StderrPath $browserStderr -WindowStyle $browserWindowStyle
            $processes += $browserProc
            try {
                Wait-Http -Url "$cdpBaseUrl/json/version" -TimeoutSeconds 30
            }
            catch {
                Write-StableSmokeFailureSummary -RepoRoot $repoRoot -SmokeLabel $SmokeLabel -Classification 'cdp_endpoint_timeout' -SummaryText ("Timed out waiting for {0}/json/version during Edge bootstrap.`nEdge launch preset: {1}`nEdge profile strategy: {2}`nBrowser stdout: {3}`nBrowser stderr: {4}`nError: {5}" -f $cdpBaseUrl, $EdgeLaunchPreset, $EdgeProfileStrategy, $browserStdout, $browserStderr, $_.Exception.Message)
                throw
            }
        }
        elseif ($reuseExistingCdpSession -or $Mode -ne 'self-start') {
            Write-Host ("Reusing existing Edge CDP endpoint at {0}." -f $cdpBaseUrl)
        }
    }

    Write-Step ("Running {0} browser smoke" -f $SmokeLabel)
    $env:OCTOPUS_UI_SMOKE_FRONTEND_URL = $frontendBaseUrl
    $env:OCTOPUS_UI_SMOKE_BACKEND_URL = $backendBaseUrl
    $env:OCTOPUS_UI_SMOKE_NODE = $resolvedNodePath
    if ($resolvedBackendBin) {
        $env:OCTOPUS_UI_SMOKE_BACKEND_BIN = $resolvedBackendBin
    }
    if (Test-Path -LiteralPath $npxCliScript) {
        $env:OCTOPUS_UI_SMOKE_NPX_SCRIPT = $npxCliScript
    }

    $nodeSmokeStdout = Join-Path $tempRoot ("node-smoke.{0}.stdout.log" -f $Driver)
    $nodeSmokeStderr = Join-Path $tempRoot ("node-smoke.{0}.stderr.log" -f $Driver)
    $nodeSmokeArgs = @($NodeSmokeScript)
    $nodeSmokeDescription = if ($Driver -eq 'cdp') { ("Node {0} smoke (cdp)" -f $SmokeLabel) } else { ("Node {0} smoke (cli)" -f $SmokeLabel) }
    $nodeSmokeTracePath = $null
    $nodeSmokeDiagnosticPath = $null

    if ($Driver -eq 'cdp') {
        $env:OCTOPUS_UI_SMOKE_CDP_URL = $cdpBaseUrl
        $env:OCTOPUS_UI_SMOKE_CDP_COMMAND_TIMEOUT_MS = [string]$CdpCommandTimeoutMs
        $env:OCTOPUS_UI_SMOKE_CDP_TRACE_FILE = (Join-Path $tempRoot 'cdp.trace.log')
        $env:OCTOPUS_UI_SMOKE_CDP_DIAGNOSTIC_FILE = (Join-Path $tempRoot 'cdp.diagnostic.json')
        $env:OCTOPUS_UI_SMOKE_EDGE_LAUNCH_PRESET = $EdgeLaunchPreset
        $env:OCTOPUS_UI_SMOKE_EDGE_PROFILE_STRATEGY = $EdgeProfileStrategy
        $env:OCTOPUS_UI_SMOKE_CDP_PAGE_BOOTSTRAP_STRATEGY = $CdpPageBootstrapStrategy
        $env:OCTOPUS_UI_SMOKE_CDP_BOOTSTRAP_COMMAND_ORDER = $CdpBootstrapCommandOrder
        $nodeSmokeTracePath = $env:OCTOPUS_UI_SMOKE_CDP_TRACE_FILE
        $nodeSmokeDiagnosticPath = $env:OCTOPUS_UI_SMOKE_CDP_DIAGNOSTIC_FILE
        Write-Host ("CDP trace file: {0}" -f $env:OCTOPUS_UI_SMOKE_CDP_TRACE_FILE)
        Write-Host ("CDP diagnostic file: {0}" -f $env:OCTOPUS_UI_SMOKE_CDP_DIAGNOSTIC_FILE)
    }

    try {
        $nodeSmokeResult = Invoke-LoggedProcessWait -FilePath $resolvedNodePath -ArgumentList $nodeSmokeArgs -WorkingDirectory $repoRoot -StdoutPath $nodeSmokeStdout -StderrPath $nodeSmokeStderr -TimeoutSeconds $NodeSmokeTimeoutSeconds -Description $nodeSmokeDescription
        Assert-NodeSmokeSucceeded -Result $nodeSmokeResult -ExpectedResult $NodeSmokeSuccessMarker
    }
    catch {
        Publish-StableSmokeDiagnosticCopies -RepoRoot $repoRoot -SmokeLabel $SmokeLabel -CdpDiagnosticPath $nodeSmokeDiagnosticPath -ExternalPreflightDiagnosticPath (Get-ExternalPreflightDiagnosticPath -TempRoot $tempRoot)
        $diagnosticSummary = ''
        if ($nodeSmokeDiagnosticPath) {
            $diagnostic = Get-JsonFileContent -Path $nodeSmokeDiagnosticPath
            $diagnosticSummary = Format-CdpDiagnosticSummary -Diagnostic $diagnostic
        }

        if ($nodeSmokeTracePath -and (Test-Path -LiteralPath $nodeSmokeTracePath)) {
            $traceTail = Get-LogTail -Path $nodeSmokeTracePath -LineCount 80
            if ([string]::IsNullOrWhiteSpace($diagnosticSummary)) {
                $diagnostic = Get-CdpDiagnosticFromTraceTail -TraceTail $traceTail -RequestedPageStrategy $CdpPageBootstrapStrategy -RequestedBootstrapOrder $CdpBootstrapCommandOrder -RequestedCommandTimeoutMs $CdpCommandTimeoutMs
                $diagnosticSummary = Format-CdpDiagnosticSummary -Diagnostic $diagnostic
            }

            if (-not [string]::IsNullOrWhiteSpace($diagnosticSummary)) {
                throw ("{0}`n{1}`nCDP diagnostic file: {2}`nCDP trace file: {3}`nCDP trace tail:`n{4}" -f $_.Exception.Message, $diagnosticSummary, $nodeSmokeDiagnosticPath, $nodeSmokeTracePath, $traceTail)
            }

            throw ("{0}`nCDP trace file: {1}`nCDP trace tail:`n{2}" -f $_.Exception.Message, $nodeSmokeTracePath, $traceTail)
        }

        if (-not [string]::IsNullOrWhiteSpace($diagnosticSummary)) {
            throw ("{0}`n{1}`nCDP diagnostic file: {2}" -f $_.Exception.Message, $diagnosticSummary, $nodeSmokeDiagnosticPath)
        }

        throw
    }

    $verificationSucceeded = $true
    Write-Step ((Get-Culture).TextInfo.ToTitleCase($SmokeLabel) + ' browser smoke passed')
    Write-Host ("Node smoke stdout: {0}" -f $nodeSmokeResult.StdoutPath)
    Write-Host ("Node smoke stderr: {0}" -f $nodeSmokeResult.StderrPath)
    Write-Host ("Frontend URL: $frontendBaseUrl")
    Write-Host ("Backend URL: $backendBaseUrl")
    if ($Driver -eq 'cdp') {
        Write-Host ("CDP URL: $cdpBaseUrl")
    }
    Write-Host ("Artifacts: $tempRoot")
    Publish-StableSmokeDiagnosticCopies -RepoRoot $repoRoot -SmokeLabel $SmokeLabel -CdpDiagnosticPath $nodeSmokeDiagnosticPath -ExternalPreflightDiagnosticPath (Get-ExternalPreflightDiagnosticPath -TempRoot $tempRoot)
}
finally {
    if (-not $verificationSucceeded -and $KeepProcessesOnFailure) {
        Write-Host ("Keeping processes for inspection. Artifacts: $tempRoot")
    }
    else {
        for ($index = $processes.Count - 1; $index -ge 0; $index--) {
            $proc = $processes[$index]
            if ($proc -and -not $proc.HasExited) {
                Stop-ProcessTree -Process $proc
            }
        }
    }

    if (-not $KeepArtifacts -and $verificationSucceeded) {
        Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}

