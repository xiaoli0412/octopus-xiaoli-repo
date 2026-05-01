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
    [string]$CdpBootstrapCommandOrder = 'runtime-page-lifecycle',

    [ValidateRange(1, 720)]
    [int]$StableDiagnosticFreshnessThresholdHours = 24,

    [switch]$UseHostFriendlyExternalDefaults,
    [switch]$BootstrapExternalCdpSession,
    [switch]$RequireExternalCdpPreflight,
    [switch]$SelfStartServices,
    [switch]$SelfStartLocalServices,
    [switch]$KeepArtifacts,
    [switch]$KeepProcessesOnFailure
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Get-LearningSmokeRepoRoot {
    $scriptDir = Split-Path -Parent $PSCommandPath
    return [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))
}

function Get-StableExternalPreflightDiagnosticCopyDirectory {
    $repoRoot = Get-LearningSmokeRepoRoot
    return Join-Path $repoRoot 'build\verify-ai-automation-learning'
}

function Normalize-ExternalPreflightPageBootstrapStrategy {
    param(
        [string]$PageBootstrapStrategy,
        [string]$PageMode
    )

    foreach ($candidate in @($PageBootstrapStrategy, $PageMode)) {
        $normalized = [string]$candidate
        if ([string]::IsNullOrWhiteSpace($normalized)) {
            continue
        }

        switch ($normalized.Trim()) {
            'attached-session' { return 'attached-session' }
            'json-new' { return 'json-new' }
            'auto' { return 'auto' }
        }
    }

    return $null
}

function Get-StableExternalPreflightDiagnosticCopyVariantName {
    param(
        [bool]$RequireCdp,
        [string]$PageBootstrapStrategy,
        [switch]$Generic
    )

    $variantLabel = if ($RequireCdp) { 'require-cdp' } else { 'optional-cdp' }
    $normalizedPageBootstrapStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy $PageBootstrapStrategy

    if ($Generic -or [string]::IsNullOrWhiteSpace($normalizedPageBootstrapStrategy)) {
        return ('latest-external-preflight-diagnostic-{0}.json' -f $variantLabel)
    }

    return ('latest-external-preflight-diagnostic-{0}-{1}.json' -f $variantLabel, $normalizedPageBootstrapStrategy)
}

function Get-StableExternalPreflightDiagnosticCopyPath {
    param(
        [object]$RequireCdp,
        [string]$PageBootstrapStrategy,
        [switch]$Generic,
        [switch]$Legacy
    )

    $stableDir = Get-StableExternalPreflightDiagnosticCopyDirectory
    if ($Legacy -or $null -eq $RequireCdp) {
        return Join-Path $stableDir 'latest-external-preflight-diagnostic.json'
    }

    return Join-Path $stableDir (Get-StableExternalPreflightDiagnosticCopyVariantName -RequireCdp ([bool]$RequireCdp) -PageBootstrapStrategy $PageBootstrapStrategy -Generic:$Generic)
}

function Publish-ExternalPreflightDiagnosticCopy {
    param([string]$SourcePath)

    if ([string]::IsNullOrWhiteSpace($SourcePath) -or -not (Test-Path -LiteralPath $SourcePath)) {
        return $null
    }

    try {
        $parent = Get-StableExternalPreflightDiagnosticCopyDirectory
        if (-not [string]::IsNullOrWhiteSpace($parent)) {
            New-Item -ItemType Directory -Path $parent -Force | Out-Null
        }

        $legacyPath = Get-StableExternalPreflightDiagnosticCopyPath -Legacy
        Copy-Item -LiteralPath $SourcePath -Destination $legacyPath -Force

        $diagnostic = Get-ExternalPreflightDiagnostic -Path $SourcePath
        $matchedPath = $legacyPath
        $genericVariantPath = $null
        if ($null -ne $diagnostic -and $null -ne $diagnostic.requireCdp) {
            $pageBootstrapStrategy = $null
            if ($null -ne $diagnostic.cdpDiagnostic) {
                $pageBootstrapStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy ([string]$diagnostic.cdpDiagnostic.pageStrategy) -PageMode ([string]$diagnostic.cdpDiagnostic.pageMode)
            }

            $matchedPath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp ([bool]$diagnostic.requireCdp) -PageBootstrapStrategy $pageBootstrapStrategy
            if ($matchedPath -ne $legacyPath) {
                Copy-Item -LiteralPath $SourcePath -Destination $matchedPath -Force
            }

            $genericVariantPath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp ([bool]$diagnostic.requireCdp) -Generic
            if (-not [string]::IsNullOrWhiteSpace($genericVariantPath) -and $genericVariantPath -ne $legacyPath -and $genericVariantPath -ne $matchedPath) {
                Copy-Item -LiteralPath $SourcePath -Destination $genericVariantPath -Force
            }
        }

        return [pscustomobject]@{
            StableCopyPath = $matchedPath
            GenericVariantPath = $genericVariantPath
            LegacyPath = $legacyPath
        }
    }
    catch {
        return $null
    }
}

function Sync-StableExternalPreflightDiagnosticVariantCopyFromLegacy {
    $legacyPath = Get-StableExternalPreflightDiagnosticCopyPath -Legacy
    if (-not (Test-Path -LiteralPath $legacyPath)) {
        return $null
    }

    $legacyDiagnostic = Get-ExternalPreflightDiagnostic -Path $legacyPath
    if ($null -eq $legacyDiagnostic -or $null -eq $legacyDiagnostic.requireCdp) {
        return $null
    }

    $pageBootstrapStrategy = $null
    if ($null -ne $legacyDiagnostic.cdpDiagnostic) {
        $pageBootstrapStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy ([string]$legacyDiagnostic.cdpDiagnostic.pageStrategy) -PageMode ([string]$legacyDiagnostic.cdpDiagnostic.pageMode)
    }

    $matchedPath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp ([bool]$legacyDiagnostic.requireCdp) -PageBootstrapStrategy $pageBootstrapStrategy
    if (-not (Test-Path -LiteralPath $matchedPath)) {
        try {
            Copy-Item -LiteralPath $legacyPath -Destination $matchedPath -Force
        }
        catch {
            return $null
        }
    }

    $genericVariantPath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp ([bool]$legacyDiagnostic.requireCdp) -Generic
    if (-not [string]::IsNullOrWhiteSpace($genericVariantPath) -and -not (Test-Path -LiteralPath $genericVariantPath)) {
        try {
            Copy-Item -LiteralPath $legacyPath -Destination $genericVariantPath -Force
        }
        catch {
            return $null
        }
    }

    return [pscustomobject]@{
        StableCopyPath = $matchedPath
        GenericVariantPath = $genericVariantPath
        LegacyPath = $legacyPath
    }
}

function Get-StableExternalPreflightCopySelectionPaths {
    param(
        [object]$RequestedRequireCdp,
        [string]$CurrentPageBootstrapStrategy
    )

    Sync-StableExternalPreflightDiagnosticVariantCopyFromLegacy | Out-Null

    $legacyPath = Get-StableExternalPreflightDiagnosticCopyPath -Legacy
    $matchingPath = $null
    $genericMatchingPath = $null
    $alternatePath = $null
    $genericAlternatePath = $null
    $normalizedPageBootstrapStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy $CurrentPageBootstrapStrategy
    if ($null -ne $RequestedRequireCdp) {
        $matchingPath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp ([bool]$RequestedRequireCdp) -PageBootstrapStrategy $normalizedPageBootstrapStrategy
        $genericMatchingPath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp ([bool]$RequestedRequireCdp) -Generic
        $alternatePath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp (-not [bool]$RequestedRequireCdp) -PageBootstrapStrategy $normalizedPageBootstrapStrategy
        $genericAlternatePath = Get-StableExternalPreflightDiagnosticCopyPath -RequireCdp (-not [bool]$RequestedRequireCdp) -Generic
    }

    $orderedPaths = [System.Collections.Generic.List[string]]::new()
    foreach ($candidatePath in @($matchingPath, $genericMatchingPath, $legacyPath, $alternatePath, $genericAlternatePath)) {
        if (-not [string]::IsNullOrWhiteSpace($candidatePath) -and -not $orderedPaths.Contains($candidatePath)) {
            $orderedPaths.Add($candidatePath)
        }
    }

    return [pscustomobject]@{
        MatchingPath = $matchingPath
        GenericMatchingPath = $genericMatchingPath
        LegacyPath = $legacyPath
        AlternatePath = $alternatePath
        GenericAlternatePath = $genericAlternatePath
        RequestedRequireCdp = $RequestedRequireCdp
        RequestedPageBootstrapStrategy = $normalizedPageBootstrapStrategy
        OrderedPaths = $orderedPaths.ToArray()
    }
}

function Get-StableExternalPreflightCopySelectionNote {
    param(
        [string]$SelectedPath,
        [string]$MatchingPath,
        [bool]$SelectedMatchingFallback,
        [bool]$MatchingPathStrategyMismatch,
        [bool]$MatchingPathExists,
        [bool]$MatchingPathParseFailed,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy
    )

    if ([string]::IsNullOrWhiteSpace($MatchingPath) -or $null -eq $RequestedRequireCdp) {
        return $null
    }

    $strategyDescriptor = if ([string]::IsNullOrWhiteSpace($RequestedPageBootstrapStrategy)) {
        'current page-bootstrap strategy'
    }
    else {
        ('page-bootstrap strategy ''{0}''' -f $RequestedPageBootstrapStrategy)
    }

    $selectedFullPath = [System.IO.Path]::GetFullPath($SelectedPath)
    $matchingFullPath = [System.IO.Path]::GetFullPath($MatchingPath)
    if ($selectedFullPath -eq $matchingFullPath -and -not $MatchingPathStrategyMismatch) {
        return ('matched this invocation''s external CDP expectation and {0} with a requirement-specific repo-local diagnostic copy.' -f $strategyDescriptor)
    }

    if ($SelectedMatchingFallback) {
        return ('no requirement-specific repo-local diagnostic copy exists yet for this invocation''s {0}; previewing the same-expectation fallback copy instead.' -f $strategyDescriptor)
    }

    if ($MatchingPathStrategyMismatch) {
        return ('no requirement-specific repo-local diagnostic copy exists yet for this invocation''s {0}; previewing the closest available saved diagnostic instead.' -f $strategyDescriptor)
    }

    if ($MatchingPathParseFailed) {
        return ('the requirement-specific repo-local diagnostic copy for this invocation and {0} exists but could not be parsed; previewing the closest available saved diagnostic instead.' -f $strategyDescriptor)
    }

    if (-not $MatchingPathExists) {
        return ('no requirement-specific repo-local diagnostic copy exists yet for this invocation and {0}; previewing the closest available saved diagnostic instead.' -f $strategyDescriptor)
    }

    return $null
}

function Get-StableExternalPreflightCopySelectionMismatchReason {
    param(
        $State,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy
    )

    if ($null -eq $State -or -not $State.Exists -or -not $State.Parsed) {
        return $null
    }

    if ($null -ne $RequestedRequireCdp -and $null -ne $State.RequireCdp -and [bool]$State.RequireCdp -ne [bool]$RequestedRequireCdp) {
        return 'require-cdp'
    }

    $normalizedRequestedStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy $RequestedPageBootstrapStrategy
    if ([string]::IsNullOrWhiteSpace($normalizedRequestedStrategy)) {
        return $null
    }

    $normalizedDiagnosticStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy $State.PageStrategy -PageMode $State.PageMode
    if ([string]::IsNullOrWhiteSpace($normalizedDiagnosticStrategy)) {
        return 'page-strategy-missing'
    }

    if ($normalizedDiagnosticStrategy -ne $normalizedRequestedStrategy) {
        return 'page-strategy'
    }

    return $null
}

function Get-StableExternalPreflightCopyState {
    param([string]$Path)

    if ([string]::IsNullOrWhiteSpace($Path)) {
        return $null
    }

    $exists = Test-Path -LiteralPath $Path
    $parsed = $false
    $diagnosticRequireCdp = $null
    $diagnosticPageStrategy = $null
    $diagnosticPageMode = $null
    $checkedAtInfo = $null
    if ($exists) {
        $diagnostic = Get-ExternalPreflightDiagnostic -Path $Path
        if ($null -ne $diagnostic) {
            $parsed = $true
            if ($null -ne $diagnostic.requireCdp) {
                $diagnosticRequireCdp = [bool]$diagnostic.requireCdp
            }
            if ($null -ne $diagnostic.cdpDiagnostic) {
                $diagnosticPageStrategy = [string]$diagnostic.cdpDiagnostic.pageStrategy
                $diagnosticPageMode = [string]$diagnostic.cdpDiagnostic.pageMode
            }

            $checkedAtInfo = Get-ExternalPreflightCheckedAtInfo -Diagnostic $diagnostic
        }
    }

    return [pscustomobject]@{
        Path = $Path
        Exists = $exists
        Parsed = $parsed
        RequireCdp = $diagnosticRequireCdp
        PageStrategy = $diagnosticPageStrategy
        PageMode = $diagnosticPageMode
        CheckedAtInfo = $checkedAtInfo
    }
}

function Get-StableExternalPreflightCopyStateLabel {
    param($State)

    if ($null -eq $State) {
        return $null
    }

    if (-not $State.Exists) {
        return 'missing'
    }

    if (-not $State.Parsed) {
        return 'present but could not be parsed'
    }

    if ($null -eq $State.RequireCdp) {
        return 'present and parsed (requireCdp unavailable in saved diagnostic)'
    }

    if ([bool]$State.RequireCdp) {
        return 'present and parsed (recorded with requireCdp=true)'
    }

    return 'present and parsed (recorded with requireCdp=false)'
}

function Get-StableExternalPreflightFreshnessThreshold {
    param([int]$ThresholdHours)

    return [TimeSpan]::FromHours([double]$ThresholdHours)
}

function Get-StableExternalPreflightCopyFreshnessLabel {
    param(
        $State,
        [int]$ThresholdHours
    )

    if ($null -eq $State -or -not $State.Parsed) {
        return $null
    }

    $checkedAtInfo = $State.CheckedAtInfo
    if ($null -eq $checkedAtInfo -or $null -eq $checkedAtInfo.ParsedValue) {
        return $null
    }

    $freshnessThreshold = Get-StableExternalPreflightFreshnessThreshold -ThresholdHours $ThresholdHours
    $age = [DateTimeOffset]::Now - $checkedAtInfo.ParsedValue
    if ($age.TotalSeconds -lt 0) {
        $age = [TimeSpan]::Zero
    }

    if ($age -le $freshnessThreshold) {
        return ('fresh against {0}h threshold' -f $ThresholdHours)
    }

    return ('stale against {0}h threshold' -f $ThresholdHours)
}

function Get-StableExternalPreflightCopyStateFreshnessSuffix {
    param(
        $State,
        [int]$ThresholdHours
    )

    if ($null -eq $State -or -not $State.Parsed) {
        return $null
    }

    $checkedAtInfo = $State.CheckedAtInfo
    if ($null -eq $checkedAtInfo -or [string]::IsNullOrWhiteSpace([string]$checkedAtInfo.RawValue)) {
        return $null
    }

    $suffixParts = [System.Collections.Generic.List[string]]::new()
    $suffixParts.Add(('checked at {0}' -f $checkedAtInfo.RawValue))
    if ($null -ne $checkedAtInfo.ParsedValue) {
        $suffixParts.Add(('age {0}' -f (Format-ExternalPreflightDiagnosticAge -CheckedAt $checkedAtInfo.ParsedValue)))
    }

    $freshnessLabel = Get-StableExternalPreflightCopyFreshnessLabel -State $State -ThresholdHours $ThresholdHours
    if (-not [string]::IsNullOrWhiteSpace($freshnessLabel)) {
        $suffixParts.Add($freshnessLabel)
    }

    return '; ' + ($suffixParts -join '; ')
}

function Get-StableExternalPreflightCopyAge {
    param($State)

    if ($null -eq $State -or -not $State.Parsed) {
        return $null
    }

    $checkedAtInfo = $State.CheckedAtInfo
    if ($null -eq $checkedAtInfo -or $null -eq $checkedAtInfo.ParsedValue) {
        return $null
    }

    $age = [DateTimeOffset]::Now - $checkedAtInfo.ParsedValue
    if ($age.TotalSeconds -lt 0) {
        return [TimeSpan]::Zero
    }

    return $age
}

function Get-StableExternalPreflightCopyStatusEntries {
    param(
        [string]$SelectedPath,
        $Selection,
        [int]$FreshnessThresholdHours = 24
    )

    if ($null -eq $Selection) {
        return @()
    }

    $selectedFullPath = $null
    if (-not [string]::IsNullOrWhiteSpace($SelectedPath)) {
        $selectedFullPath = [System.IO.Path]::GetFullPath($SelectedPath)
    }

    $seenPaths = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::OrdinalIgnoreCase)
    $statusEntries = [System.Collections.Generic.List[object]]::new()
    $copyDescriptors = @(
        @{ Label = 'matching'; Path = $Selection.MatchingPath },
        @{ Label = 'matching-generic'; Path = $Selection.GenericMatchingPath },
        @{ Label = 'alternate'; Path = $Selection.AlternatePath },
        @{ Label = 'alternate-generic'; Path = $Selection.GenericAlternatePath },
        @{ Label = 'legacy'; Path = $Selection.LegacyPath }
    )

    foreach ($descriptor in $copyDescriptors) {
        $path = [string]$descriptor.Path
        if ([string]::IsNullOrWhiteSpace($path)) {
            continue
        }

        $fullPath = [System.IO.Path]::GetFullPath($path)
        if (-not $seenPaths.Add($fullPath)) {
            continue
        }

        $state = Get-StableExternalPreflightCopyState -Path $path
        $stateLabel = Get-StableExternalPreflightCopyStateLabel -State $state
        if ([string]::IsNullOrWhiteSpace($stateLabel)) {
            continue
        }

        $freshnessLabel = Get-StableExternalPreflightCopyFreshnessLabel -State $state -ThresholdHours $FreshnessThresholdHours
        $freshnessSuffix = Get-StableExternalPreflightCopyStateFreshnessSuffix -State $state -ThresholdHours $FreshnessThresholdHours

        $statusEntries.Add([pscustomobject]@{
            Label = [string]$descriptor.Label
            Path = $path
            FullPath = $fullPath
            State = $state
            MismatchReason = $(if ([string]$descriptor.Label -eq 'matching') { Get-StableExternalPreflightCopySelectionMismatchReason -State $state -RequestedRequireCdp $Selection.RequestedRequireCdp -RequestedPageBootstrapStrategy $Selection.RequestedPageBootstrapStrategy } else { $null })
            StateLabel = $stateLabel
            FreshnessLabel = $freshnessLabel
            FreshnessSuffix = $freshnessSuffix
            Age = Get-StableExternalPreflightCopyAge -State $state
            Selected = (-not [string]::IsNullOrWhiteSpace($selectedFullPath) -and $selectedFullPath -eq $fullPath)
        })
    }

    return $statusEntries.ToArray()
}

function Get-StableExternalPreflightAlternateRequirementSpecificCopySummary {
    param([object[]]$StatusEntries)

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    $selectedEntry = $StatusEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedEntry -or [string]$selectedEntry.Label -ne 'legacy') {
        return $null
    }

    $alternateEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate' -and $_.State.Parsed } | Select-Object -First 1
    if ($null -eq $alternateEntry) {
        return $null
    }

    $relation = 'available'
    $freshnessNoteSuffix = ' The alternate requirement-specific copy is also available.'
    $decisionSummarySentence = 'Repo-local diagnostics already include an alternate requirement-specific copy for the opposite CDP expectation.'

    $selectedCheckedAt = $null
    if ($null -ne $selectedEntry.State -and $null -ne $selectedEntry.State.CheckedAtInfo) {
        $selectedCheckedAt = $selectedEntry.State.CheckedAtInfo.ParsedValue
    }

    $alternateCheckedAt = $null
    if ($null -ne $alternateEntry.State -and $null -ne $alternateEntry.State.CheckedAtInfo) {
        $alternateCheckedAt = $alternateEntry.State.CheckedAtInfo.ParsedValue
    }

    if ($null -ne $selectedCheckedAt -and $null -ne $alternateCheckedAt) {
        if ($alternateCheckedAt -eq $selectedCheckedAt) {
            $relation = 'same-freshness'
            $freshnessNoteSuffix = ' The alternate requirement-specific copy is available at the same freshness.'
            $decisionSummarySentence = 'Repo-local diagnostics already include an alternate requirement-specific copy at the same freshness for the opposite CDP expectation.'
        }
        elseif ($alternateCheckedAt -gt $selectedCheckedAt) {
            $relation = 'fresher'
            $freshnessNoteSuffix = ' The alternate requirement-specific copy is fresher than the selected legacy fallback.'
            $decisionSummarySentence = 'Repo-local diagnostics already include a fresher alternate requirement-specific copy for the opposite CDP expectation.'
        }
        else {
            $relation = 'older'
            $freshnessNoteSuffix = ' The alternate requirement-specific copy is older than the selected legacy fallback.'
            $decisionSummarySentence = 'The alternate requirement-specific copy for the opposite CDP expectation is older than the selected legacy fallback.'
        }
    }

    return [pscustomobject]@{
        Relation = $relation
        FreshnessNoteSuffix = $freshnessNoteSuffix
        DecisionSummarySentence = $decisionSummarySentence
    }
}

function Get-StableExternalPreflightCopyFreshnessNote {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [int]$FreshnessThresholdHours = 24
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    $selectedEntry = $StatusEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedEntry) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $selectedFreshnessLabel = [string]$selectedEntry.FreshnessLabel

    if ($selectedEntry.Label -ne 'matching') {
        $selectedDescriptor = switch ($selectedEntry.Label) {
            'alternate' { 'alternate fallback preview' }
            'legacy' { 'legacy fallback preview' }
            default { 'selected fallback preview' }
        }

        $alternateSuffix = ''
        if ($selectedEntry.Label -eq 'legacy') {
            $alternateSummary = Get-StableExternalPreflightAlternateRequirementSpecificCopySummary -StatusEntries $StatusEntries
            if ($null -ne $alternateSummary -and -not [string]::IsNullOrWhiteSpace([string]$alternateSummary.FreshnessNoteSuffix)) {
                $alternateSuffix = [string]$alternateSummary.FreshnessNoteSuffix
            }
        }

        if ($null -ne $matchingEntry -and (($matchingEntry.MismatchReason -eq 'page-strategy') -or ($matchingEntry.MismatchReason -eq 'page-strategy-missing'))) {
            $preferenceState = Get-StableExternalPreflightSelectedPreferenceState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
            $preferenceCaution = Get-StableExternalPreflightSelectedPreferenceFallbackCaution -PreferenceState $preferenceState
            if ([string]::IsNullOrWhiteSpace($selectedFreshnessLabel)) {
                return ('no matching requirement-specific repo-local copy exists yet for this invocation''s page-bootstrap strategy; previewing the {0}.{1}{2}' -f $selectedDescriptor, $alternateSuffix, $preferenceCaution)
            }

            if ($selectedFreshnessLabel -like 'stale*') {
                return ('no matching requirement-specific repo-local copy exists yet for this invocation''s page-bootstrap strategy; the {0} is {1}, so rerun the preferred refresh command on a healthy host.{2}{3}' -f $selectedDescriptor, $selectedFreshnessLabel, $alternateSuffix, $preferenceCaution)
            }

            return ('no matching requirement-specific repo-local copy exists yet for this invocation''s page-bootstrap strategy; the {0} is {1}.{2}{3}' -f $selectedDescriptor, $selectedFreshnessLabel, $alternateSuffix, $preferenceCaution)
        }

        if ($null -ne $matchingEntry -and $matchingEntry.State.Exists -and -not $matchingEntry.State.Parsed) {
            if ([string]::IsNullOrWhiteSpace($selectedFreshnessLabel)) {
                return ('the matching requirement-specific repo-local copy exists but is not parseable yet; previewing the {0}.{1}' -f $selectedDescriptor, $alternateSuffix)
            }

            return ('the matching requirement-specific repo-local copy exists but is not parseable yet; the {0} is {1}.{2}' -f $selectedDescriptor, $selectedFreshnessLabel, $alternateSuffix)
        }

        if ([string]::IsNullOrWhiteSpace($selectedFreshnessLabel)) {
            return ('no matching requirement-specific repo-local copy exists yet for this invocation; previewing the {0}.{1}' -f $selectedDescriptor, $alternateSuffix)
        }

        if ($selectedFreshnessLabel -like 'stale*') {
            return ('no matching requirement-specific repo-local copy exists yet for this invocation; the {0} is {1}, so rerun the preferred refresh command on a healthy host.{2}' -f $selectedDescriptor, $selectedFreshnessLabel, $alternateSuffix)
        }

        return ('no matching requirement-specific repo-local copy exists yet for this invocation; the {0} is {1}.{2}' -f $selectedDescriptor, $selectedFreshnessLabel, $alternateSuffix)
    }

    if ([string]::IsNullOrWhiteSpace($selectedFreshnessLabel)) {
        return 'selected preview matches this invocation, but its freshness could not be classified from the saved diagnostic.'
    }

    if ($selectedFreshnessLabel -like 'stale*') {
        return ('selected preview matches this invocation but is {0}; rerun the preferred refresh command if you need newer external evidence.' -f $selectedFreshnessLabel)
    }

    return ('selected preview matches this invocation and remains {0}.' -f $selectedFreshnessLabel)
}

function Get-StableExternalPreflightCopyDescriptor {
    param($Entry)

    if ($null -eq $Entry) {
        return $null
    }

    switch ([string]$Entry.Label) {
        'matching' { return 'matching requirement-specific copy' }
        'matching-generic' { return 'same-expectation fallback copy' }
        'alternate' { return 'alternate requirement-specific copy' }
        'alternate-generic' { return 'opposite-expectation fallback copy' }
        'legacy' { return 'legacy fallback copy' }
        default { return 'saved repo-local copy' }
    }
}

function Get-StableExternalPreflightEntryRecordedPageBootstrapStrategy {
    param($Entry)

    if ($null -eq $Entry -or $null -eq $Entry.State -or -not $Entry.State.Parsed) {
        return $null
    }

    return Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy ([string]$Entry.State.PageStrategy) -PageMode ([string]$Entry.State.PageMode)
}

function Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy {
    param(
        $Entry,
        [string]$RequestedPageBootstrapStrategy
    )

    $requestedStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy $RequestedPageBootstrapStrategy
    if ([string]::IsNullOrWhiteSpace($requestedStrategy)) {
        return $null
    }

    $recordedStrategy = Get-StableExternalPreflightEntryRecordedPageBootstrapStrategy -Entry $Entry
    if ([string]::IsNullOrWhiteSpace($recordedStrategy) -or $recordedStrategy -eq $requestedStrategy) {
        return $null
    }

    return $recordedStrategy
}

function Format-StableExternalPreflightCopyDescriptorList {
    param([object[]]$Entries)

    if ($null -eq $Entries -or $Entries.Count -eq 0) {
        return $null
    }

    $descriptors = @(
        $Entries |
            ForEach-Object { Get-StableExternalPreflightCopyDescriptor -Entry $_ } |
            Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) }
    )

    if ($descriptors.Count -eq 0) {
        return $null
    }

    if ($descriptors.Count -eq 1) {
        return $descriptors[0]
    }

    if ($descriptors.Count -eq 2) {
        return ($descriptors -join ' and ')
    }

    return ((($descriptors | Select-Object -First ($descriptors.Count - 1)) -join ', ') + ', and ' + $descriptors[-1])
}

function Get-StableExternalPreflightSelectedPreferenceState {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    $selectedEntry = $StatusEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedEntry -or $null -eq $selectedEntry.State -or -not $selectedEntry.State.Parsed) {
        return $null
    }

    $selectedCheckedAtInfo = $selectedEntry.State.CheckedAtInfo
    if ($null -eq $selectedCheckedAtInfo -or $null -eq $selectedCheckedAtInfo.ParsedValue) {
        return $null
    }

    $newerEntries = @(
        $StatusEntries |
            Where-Object {
                -not $_.Selected -and
                $_.State.Parsed -and
                $null -ne $_.State.CheckedAtInfo -and
                $null -ne $_.State.CheckedAtInfo.ParsedValue -and
                $_.State.CheckedAtInfo.ParsedValue -gt $selectedCheckedAtInfo.ParsedValue
            }
    )
    if ($newerEntries.Count -eq 0) {
        return $null
    }

    $sameExpectationEntries = [System.Collections.Generic.List[object]]::new()
    $oppositeExpectationEntries = [System.Collections.Generic.List[object]]::new()
    $alternateStrategyEntries = [System.Collections.Generic.List[object]]::new()
    $unknownFitEntries = [System.Collections.Generic.List[object]]::new()

    foreach ($entry in $newerEntries) {
        $mismatchReason = Get-StableExternalPreflightCopySelectionMismatchReason -State $entry.State -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
        switch ([string]$mismatchReason) {
            '' {
                $sameExpectationEntries.Add($entry)
            }
            'require-cdp' {
                $oppositeExpectationEntries.Add($entry)
            }
            'page-strategy' {
                $alternateStrategyEntries.Add($entry)
            }
            'page-strategy-missing' {
                $alternateStrategyEntries.Add($entry)
            }
            default {
                $unknownFitEntries.Add($entry)
            }
        }
    }

    return [pscustomobject]@{
        SelectedEntry = $selectedEntry
        NewerEntries = $newerEntries
        SameExpectationEntries = $sameExpectationEntries.ToArray()
        OppositeExpectationEntries = $oppositeExpectationEntries.ToArray()
        AlternateStrategyEntries = $alternateStrategyEntries.ToArray()
        UnknownFitEntries = $unknownFitEntries.ToArray()
    }
}

function Get-StableExternalPreflightSelectedPreferenceRole {
    param($SelectedEntry)

    if ($null -eq $SelectedEntry) {
        return 'best available replay'
    }

    switch ([string]$SelectedEntry.Label) {
        'matching' { return 'best requirement-specific replay' }
        'matching-generic' { return 'best same-expectation replay' }
        default { return 'best available replay' }
    }
}

function Get-StableExternalPreflightSelectedPreferenceFreshestSuffix {
    param($PreferenceState)

    if ($null -eq $PreferenceState) {
        return $null
    }

    if ($PreferenceState.SameExpectationEntries.Count -gt 0) {
        return $null
    }

    $selectedRole = Get-StableExternalPreflightSelectedPreferenceRole -SelectedEntry $PreferenceState.SelectedEntry
    if ($PreferenceState.OppositeExpectationEntries.Count -gt 0 -and $PreferenceState.AlternateStrategyEntries.Count -gt 0) {
        return (' These fresher copies are comparison-only artifacts because they either flip the requested CDP expectation or come from a different page-bootstrap strategy, so the selected preview remains the {0} for this invocation.' -f $selectedRole)
    }

    if ($PreferenceState.OppositeExpectationEntries.Count -gt 0) {
        return (' These fresher copies do not preserve the requested CDP expectation, so the selected preview remains the {0} for this invocation.' -f $selectedRole)
    }

    if ($PreferenceState.AlternateStrategyEntries.Count -gt 0) {
        return (' These fresher copies were captured for a different page-bootstrap strategy, so the selected preview remains the {0} for this invocation.' -f $selectedRole)
    }

    if ($PreferenceState.UnknownFitEntries.Count -gt 0) {
        return (' These fresher copies are not a better fit for this invocation than the selected preview, so the selected preview remains the {0}.' -f $selectedRole)
    }

    return $null
}

function Get-StableExternalPreflightSelectedPreferenceFallbackCaution {
    param($PreferenceState)

    if ($null -eq $PreferenceState) {
        return $null
    }

    if ($PreferenceState.SameExpectationEntries.Count -gt 0) {
        return $null
    }

    if ($null -eq $PreferenceState.SelectedEntry -or [string]$PreferenceState.SelectedEntry.Label -ne 'matching-generic') {
        return $null
    }

    if ($PreferenceState.OppositeExpectationEntries.Count -gt 0 -and $PreferenceState.AlternateStrategyEntries.Count -gt 0) {
        return ' A fresher comparison copy may also exist repo-locally, but it either flips the requested CDP expectation or comes from a different page-bootstrap strategy; keep it only for comparison instead of replacing the selected same-expectation replay.'
    }

    if ($PreferenceState.OppositeExpectationEntries.Count -gt 0) {
        return ' A fresher opposite-expectation copy may also exist repo-locally, but keep it only for comparison instead of replacing the selected same-expectation replay.'
    }

    if ($PreferenceState.AlternateStrategyEntries.Count -gt 0) {
        return ' A fresher copy from a different page-bootstrap strategy may also exist repo-locally, but keep it only for comparison instead of replacing the selected same-expectation replay.'
    }

    if ($PreferenceState.UnknownFitEntries.Count -gt 0) {
        return ' A fresher repo-local copy may also exist, but it is not a better fit for this invocation than the selected same-expectation replay.'
    }

    return $null
}

function Get-StableExternalPreflightFreshestCopyNote {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    $selectedEntry = $StatusEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedEntry) {
        return $null
    }

    $comparableEntries = @(
        $StatusEntries |
            Where-Object {
                $_.State.Parsed -and
                $null -ne $_.State.CheckedAtInfo -and
                $null -ne $_.State.CheckedAtInfo.ParsedValue
            }
    )
    if ($comparableEntries.Count -eq 0) {
        return $null
    }

    $selectedComparableEntry = $comparableEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedComparableEntry) {
        return $null
    }

    $freshestComparableEntry = $comparableEntries |
        Sort-Object -Property @{ Expression = { $_.State.CheckedAtInfo.ParsedValue.UtcDateTime } } -Descending |
        Select-Object -First 1
    if ($null -eq $freshestComparableEntry) {
        return $null
    }

    $freshestCheckedAt = $freshestComparableEntry.State.CheckedAtInfo.ParsedValue
    $freshestEntries = @(
        $comparableEntries |
            Where-Object { $_.State.CheckedAtInfo.ParsedValue -eq $freshestCheckedAt }
    )
    if ($freshestEntries.Count -eq 0) {
        return $null
    }

    $selectedIsFreshest = @($freshestEntries | Where-Object { $_.Selected }).Count -gt 0
    if ($selectedIsFreshest) {
        if ($comparableEntries.Count -eq 1) {
            return 'selected preview is currently the only parseable repo-local diagnostic copy with comparable timestamp data.'
        }

        $otherFreshestEntries = @($freshestEntries | Where-Object { -not $_.Selected })
        if ($otherFreshestEntries.Count -eq 0) {
            return 'selected preview is currently the freshest parseable repo-local diagnostic copy saved in this workspace.'
        }

        $descriptorList = Format-StableExternalPreflightCopyDescriptorList -Entries $otherFreshestEntries
        if (-not [string]::IsNullOrWhiteSpace($descriptorList)) {
            return ('selected preview is tied for the freshest parseable repo-local diagnostic copy with the {0}.' -f $descriptorList)
        }

        return 'selected preview is tied for the freshest parseable repo-local diagnostic copy with another saved repo-local copy.'
    }

    $newerEntries = @($freshestEntries | Where-Object { -not $_.Selected })
    if ($newerEntries.Count -eq 0) {
        return $null
    }

    $descriptorList = Format-StableExternalPreflightCopyDescriptorList -Entries $newerEntries
    $preferenceState = Get-StableExternalPreflightSelectedPreferenceState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $preferenceSuffix = Get-StableExternalPreflightSelectedPreferenceFreshestSuffix -PreferenceState $preferenceState
    if ([string]::IsNullOrWhiteSpace($descriptorList)) {
        return ('selected preview is older than the freshest parseable repo-local diagnostic copy saved in this workspace.{0}' -f $preferenceSuffix)
    }

    $verb = if ($newerEntries.Count -eq 1) { 'was' } else { 'were' }
    return ('selected preview is older than the freshest parseable repo-local diagnostic copy saved in this workspace; the {0} {1} checked more recently.{2}' -f $descriptorList, $verb, $preferenceSuffix)
}

function Get-StableExternalPreflightCoverageNote {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0 -or $null -eq $RequestedRequireCdp) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $matchingGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching-generic' } | Select-Object -First 1
    $alternateEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate' } | Select-Object -First 1
    $alternateGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate-generic' } | Select-Object -First 1

    $requestedLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=true' } else { 'requireCdp=false' }
    $alternateLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=false' } else { 'requireCdp=true' }
    $strategySuffix = if ([string]::IsNullOrWhiteSpace($RequestedPageBootstrapStrategy)) { '' } else { (' for page-bootstrap strategy ''{0}''' -f $RequestedPageBootstrapStrategy) }

    $matchingParseable = ($null -ne $matchingEntry) -and $matchingEntry.State.Parsed
    $matchingGenericParseable = ($null -ne $matchingGenericEntry) -and $matchingGenericEntry.State.Parsed
    $alternateParseable = ($null -ne $alternateEntry) -and $alternateEntry.State.Parsed
    $alternateGenericParseable = ($null -ne $alternateGenericEntry) -and $alternateGenericEntry.State.Parsed
    $matchingExistsButUnparseable = ($null -ne $matchingEntry) -and $matchingEntry.State.Exists -and (-not $matchingEntry.State.Parsed)
    $alternateExistsButUnparseable = ($null -ne $alternateEntry) -and $alternateEntry.State.Exists -and (-not $alternateEntry.State.Parsed)
    $matchingGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $matchingGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $alternateGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $alternateGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy

    if ($matchingParseable -and $alternateParseable) {
        return ('repo-local stable coverage includes parseable {0} and {1} variants{2}.' -f $requestedLabel, $alternateLabel, $strategySuffix)
    }

    if ($matchingParseable) {
        if ($alternateExistsButUnparseable) {
            return ('repo-local stable coverage currently only has a parseable {0} variant; the saved {1} variant exists but could not be parsed yet.' -f $requestedLabel, $alternateLabel)
        }

        return ('repo-local stable coverage currently only includes a parseable {0} variant{2}; no parseable {1} variant has been captured yet.' -f $requestedLabel, $alternateLabel, $strategySuffix)
    }

    if ($matchingGenericParseable) {
        if (-not [string]::IsNullOrWhiteSpace($matchingGenericAlternateStrategy)) {
            return ('repo-local stable coverage includes a parseable {0} fallback variant from page-bootstrap strategy ''{2}'', but no parseable {0} variant has been captured yet{1}.' -f $requestedLabel, $strategySuffix, $matchingGenericAlternateStrategy)
        }

        return ('repo-local stable coverage includes a parseable {0} fallback variant from a different page-bootstrap strategy, but no parseable {0} variant has been captured yet{1}.' -f $requestedLabel, $strategySuffix)
    }

    if ($alternateParseable) {
        if ($matchingExistsButUnparseable) {
            return ('repo-local stable coverage currently only has a parseable {0} variant; the saved {1} variant exists but could not be parsed yet.' -f $alternateLabel, $requestedLabel)
        }

        return ('repo-local stable coverage currently only includes a parseable {0} variant{2}; no parseable {1} variant has been captured yet.' -f $alternateLabel, $requestedLabel, $strategySuffix)
    }

    if ($alternateGenericParseable) {
        if (-not [string]::IsNullOrWhiteSpace($alternateGenericAlternateStrategy)) {
            return ('repo-local stable coverage includes only an opposite-expectation fallback variant from page-bootstrap strategy ''{1}''; no parseable {0} variant has been captured yet{2}.' -f $requestedLabel, $alternateGenericAlternateStrategy, $strategySuffix)
        }

        return ('repo-local stable coverage includes only an opposite-expectation fallback variant from a different page-bootstrap strategy; no parseable {0} variant has been captured yet{1}.' -f $requestedLabel, $strategySuffix)
    }

    if ($matchingExistsButUnparseable -and $alternateExistsButUnparseable) {
        return ('repo-local stable coverage does not yet include any parseable {0} or {1} variants; both saved copies need inspection.' -f $requestedLabel, $alternateLabel)
    }

    if ($matchingExistsButUnparseable) {
        return ('repo-local stable coverage does not yet include a parseable {0} variant, and no parseable {1} variant has been captured yet.' -f $requestedLabel, $alternateLabel)
    }

    if ($alternateExistsButUnparseable) {
        return ('repo-local stable coverage does not yet include a parseable {1} variant, and no parseable {0} variant has been captured yet.' -f $requestedLabel, $alternateLabel)
    }

    return ('repo-local stable coverage does not yet include any parseable {0} or {1} variants{2}.' -f $requestedLabel, $alternateLabel, $strategySuffix)
}

function Get-StableExternalPreflightCoverageDecisionSentence {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0 -or $null -eq $RequestedRequireCdp) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $matchingGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching-generic' } | Select-Object -First 1
    $alternateEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate' } | Select-Object -First 1
    $alternateGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate-generic' } | Select-Object -First 1

    $requestedLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=true' } else { 'requireCdp=false' }
    $alternateLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=false' } else { 'requireCdp=true' }
    $strategySuffix = if ([string]::IsNullOrWhiteSpace($RequestedPageBootstrapStrategy)) { '' } else { (' for page-bootstrap strategy ''{0}''' -f $RequestedPageBootstrapStrategy) }

    $matchingParseable = ($null -ne $matchingEntry) -and $matchingEntry.State.Parsed
    $matchingGenericParseable = ($null -ne $matchingGenericEntry) -and $matchingGenericEntry.State.Parsed
    $alternateParseable = ($null -ne $alternateEntry) -and $alternateEntry.State.Parsed
    $alternateGenericParseable = ($null -ne $alternateGenericEntry) -and $alternateGenericEntry.State.Parsed
    $matchingExistsButUnparseable = ($null -ne $matchingEntry) -and $matchingEntry.State.Exists -and (-not $matchingEntry.State.Parsed)
    $alternateExistsButUnparseable = ($null -ne $alternateEntry) -and $alternateEntry.State.Exists -and (-not $alternateEntry.State.Parsed)
    $matchingGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $matchingGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $alternateGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $alternateGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy

    if ($matchingParseable -and $alternateParseable) {
        return ('Repo-local stable coverage is complete for both {0} and {1} variants{2}, so only freshness and live reachability remain relevant now.' -f $requestedLabel, $alternateLabel, $strategySuffix)
    }

    if ($matchingGenericParseable) {
        if (-not [string]::IsNullOrWhiteSpace($matchingGenericAlternateStrategy)) {
            return ('Repo-local stable coverage is still incomplete for this invocation because only a {0} fallback copy from page-bootstrap strategy ''{2}'' is available{1}.' -f $requestedLabel, $strategySuffix, $matchingGenericAlternateStrategy)
        }

        return ('Repo-local stable coverage is still incomplete for this invocation because only a {0} fallback copy from a different page-bootstrap strategy is available{1}.' -f $requestedLabel, $strategySuffix)
    }

    if ($alternateGenericParseable) {
        if (-not [string]::IsNullOrWhiteSpace($alternateGenericAlternateStrategy)) {
            return ('Repo-local stable coverage is still incomplete for this invocation because only an opposite-expectation fallback copy from page-bootstrap strategy ''{0}'' is available{1}.' -f $alternateGenericAlternateStrategy, $strategySuffix)
        }

        return ('Repo-local stable coverage is still incomplete for this invocation because only an opposite-expectation fallback copy from a different page-bootstrap strategy is available{0}.' -f $strategySuffix)
    }

    if ($matchingParseable) {
        if ($alternateExistsButUnparseable) {
            return ('Repo-local stable coverage is still incomplete because the saved {0} variant exists but is not parseable yet.' -f $alternateLabel)
        }

        return ('Repo-local stable coverage is still incomplete because no parseable {0} variant has been captured yet.' -f $alternateLabel)
    }

    if ($alternateParseable) {
        if ($matchingExistsButUnparseable) {
            return ('Repo-local stable coverage is still incomplete for this invocation because the saved {0} variant exists but is not parseable yet.' -f $requestedLabel)
        }

        return ('Repo-local stable coverage is still incomplete for this invocation because no parseable {0} variant has been captured yet.' -f $requestedLabel)
    }

    if ($matchingExistsButUnparseable -and $alternateExistsButUnparseable) {
        return ('Repo-local stable coverage is still incomplete because neither saved {0} nor {1} variant is parseable yet.' -f $requestedLabel, $alternateLabel)
    }

    if ($matchingExistsButUnparseable) {
        return ('Repo-local stable coverage is still incomplete because the saved {0} variant is not parseable and no parseable {1} variant has been captured yet.' -f $requestedLabel, $alternateLabel)
    }

    if ($alternateExistsButUnparseable) {
        return ('Repo-local stable coverage is still incomplete because the saved {0} variant is not parseable and no parseable {1} variant has been captured yet.' -f $alternateLabel, $requestedLabel)
    }

    return ('Repo-local stable coverage is still incomplete because no parseable {0} or {1} variant has been captured yet.' -f $requestedLabel, $alternateLabel)
}

function Get-ExternalPreflightDiagnosticSummaryFragment {
    param($Diagnostic)

    if ($null -eq $Diagnostic) {
        return 'diagnostic unavailable'
    }

    $classification = [string]$Diagnostic.overallClassification
    if ([string]::IsNullOrWhiteSpace($classification)) {
        $classification = 'classification unavailable'
    }

    $failedChecks = @($Diagnostic.failedChecks | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    if ($failedChecks.Count -gt 0) {
        return ('{0}; failed checks {1}' -f $classification, ($failedChecks -join ', '))
    }

    $skippedChecks = @($Diagnostic.skippedChecks | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    if ($skippedChecks.Count -gt 0) {
        return ('{0}; skipped checks {1}' -f $classification, ($skippedChecks -join ', '))
    }

    if (-not [string]::IsNullOrWhiteSpace([string]$Diagnostic.primaryBlockingCheck)) {
        return ('{0}; primary blocking check {1}' -f $classification, [string]$Diagnostic.primaryBlockingCheck)
    }

    return $classification
}

function Get-ExternalPreflightServiceReachabilityState {
    param($Diagnostic)

    if ($null -eq $Diagnostic) {
        return $null
    }

    $backendReachable = ($null -ne $Diagnostic.backend) -and ($Diagnostic.backend.reachable -eq $true)
    $frontendReachable = ($null -ne $Diagnostic.frontend) -and ($Diagnostic.frontend.reachable -eq $true)
    if ($backendReachable -and $frontendReachable) {
        return 'services_reachable'
    }

    return 'services_unreachable'
}

function Get-StableExternalPreflightAlignedCdpBootstrapFocusState {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0 -or $null -eq $RequestedRequireCdp) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $alternateEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate' } | Select-Object -First 1
    if ($null -eq $matchingEntry -or $null -eq $alternateEntry) {
        return $null
    }

    if (-not $matchingEntry.State.Parsed -or -not $alternateEntry.State.Parsed) {
        return $null
    }

    $matchingDiagnostic = Get-ExternalPreflightDiagnostic -Path $matchingEntry.Path
    $alternateDiagnostic = Get-ExternalPreflightDiagnostic -Path $alternateEntry.Path
    if ($null -eq $matchingDiagnostic -or $null -eq $alternateDiagnostic) {
        return $null
    }

    if ((Get-ExternalPreflightServiceReachabilityState -Diagnostic $matchingDiagnostic) -ne 'services_reachable') {
        return $null
    }

    if ((Get-ExternalPreflightServiceReachabilityState -Diagnostic $alternateDiagnostic) -ne 'services_reachable') {
        return $null
    }

    $matchingFailedChecks = @($matchingDiagnostic.failedChecks | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    $alternateFailedChecks = @($alternateDiagnostic.failedChecks | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    $matchingCdpOnly = ([string]$matchingDiagnostic.overallClassification -eq 'cdp_smoke_failed') -and $matchingFailedChecks.Count -eq 1 -and $matchingFailedChecks[0] -eq 'cdp'
    $alternateCdpOnly = ([string]$alternateDiagnostic.overallClassification -eq 'cdp_smoke_failed') -and $alternateFailedChecks.Count -eq 1 -and $alternateFailedChecks[0] -eq 'cdp'
    if (-not $matchingCdpOnly -or -not $alternateCdpOnly) {
        return $null
    }

    $requestedLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=true' } else { 'requireCdp=false' }
    $alternateLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=false' } else { 'requireCdp=true' }

    $matchingCdpBootstrapState = Get-ExternalPreflightAttachedSessionBootstrapTimeoutState -Diagnostic $matchingDiagnostic
    $alternateCdpBootstrapState = Get-ExternalPreflightAttachedSessionBootstrapTimeoutState -Diagnostic $alternateDiagnostic

    return [pscustomobject]@{
        RequestedLabel = $requestedLabel
        AlternateLabel = $alternateLabel
        MatchingSummary = Get-ExternalPreflightDiagnosticSummaryFragment -Diagnostic $matchingDiagnostic
        AlternateSummary = Get-ExternalPreflightDiagnosticSummaryFragment -Diagnostic $alternateDiagnostic
        MatchingCdpBootstrapState = $matchingCdpBootstrapState
        AlternateCdpBootstrapState = $alternateCdpBootstrapState
    }
}

function Get-ExternalPreflightAttachedSessionBootstrapTimeoutState {
    param($Diagnostic)

    if ($null -eq $Diagnostic -or $null -eq $Diagnostic.cdpDiagnostic) {
        return $null
    }

    $cdpDiagnostic = $Diagnostic.cdpDiagnostic
    $classification = [string]$cdpDiagnostic.classification
    $pageMode = [string]$cdpDiagnostic.pageMode
    if ($classification -ne 'page_bootstrap_timeout_attached_session' -or $pageMode -ne 'attached-session') {
        return $null
    }

    $commandTimeoutMs = 0
    try {
        $commandTimeoutMs = [int]$cdpDiagnostic.commandTimeoutMs
    }
    catch {
        $commandTimeoutMs = 0
    }

    return [pscustomobject]@{
        Classification = $classification
        PageMode = $pageMode
        PageStrategy = [string]$cdpDiagnostic.pageStrategy
        BootstrapCommandOrder = [string]$cdpDiagnostic.bootstrapCommandOrder
        CommandTimeoutMs = $commandTimeoutMs
        Hint = [string]$cdpDiagnostic.hint
    }
}

function Get-ExternalPreflightPageBootstrapTimeoutState {
    param($Diagnostic)

    if ($null -eq $Diagnostic -or $null -eq $Diagnostic.cdpDiagnostic) {
        return $null
    }

    $cdpDiagnostic = $Diagnostic.cdpDiagnostic
    $classification = [string]$cdpDiagnostic.classification
    if ($classification -notlike 'page_bootstrap_timeout_*') {
        return $null
    }

    $pageStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy ([string]$cdpDiagnostic.pageStrategy) -PageMode ([string]$cdpDiagnostic.pageMode)
    if ([string]::IsNullOrWhiteSpace($pageStrategy)) {
        return $null
    }

    $commandTimeoutMs = 0
    try {
        $commandTimeoutMs = [int]$cdpDiagnostic.commandTimeoutMs
    }
    catch {
        $commandTimeoutMs = 0
    }

    return [pscustomobject]@{
        Classification = $classification
        PageMode = [string]$cdpDiagnostic.pageMode
        PageStrategy = $pageStrategy
        BootstrapCommandOrder = [string]$cdpDiagnostic.bootstrapCommandOrder
        CommandTimeoutMs = $commandTimeoutMs
        Hint = [string]$cdpDiagnostic.hint
    }
}

function Get-StableExternalPreflightPageBootstrapBlockerConfirmedStateForEntryPair {
    param(
        [object[]]$StatusEntries,
        [string]$MatchingLabel,
        [string]$AlternateLabel,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [int]$MinimumCommandTimeoutMs = 60000
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    if ($null -eq $RequestedRequireCdp) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq $MatchingLabel } | Select-Object -First 1
    $alternateEntry = $StatusEntries | Where-Object { $_.Label -eq $AlternateLabel } | Select-Object -First 1
    if ($null -eq $matchingEntry -or $null -eq $alternateEntry -or -not $matchingEntry.State.Parsed -or -not $alternateEntry.State.Parsed) {
        return $null
    }

    $matchingDiagnostic = Get-ExternalPreflightDiagnostic -Path $matchingEntry.Path
    $alternateDiagnostic = Get-ExternalPreflightDiagnostic -Path $alternateEntry.Path
    if ($null -eq $matchingDiagnostic -or $null -eq $alternateDiagnostic) {
        return $null
    }

    if ((Get-ExternalPreflightServiceReachabilityState -Diagnostic $matchingDiagnostic) -ne 'services_reachable') {
        return $null
    }

    if ((Get-ExternalPreflightServiceReachabilityState -Diagnostic $alternateDiagnostic) -ne 'services_reachable') {
        return $null
    }

    $matchingFailedChecks = @($matchingDiagnostic.failedChecks | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    $alternateFailedChecks = @($alternateDiagnostic.failedChecks | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) })
    $matchingCdpOnly = ([string]$matchingDiagnostic.overallClassification -eq 'cdp_smoke_failed') -and $matchingFailedChecks.Count -eq 1 -and $matchingFailedChecks[0] -eq 'cdp'
    $alternateCdpOnly = ([string]$alternateDiagnostic.overallClassification -eq 'cdp_smoke_failed') -and $alternateFailedChecks.Count -eq 1 -and $alternateFailedChecks[0] -eq 'cdp'
    if (-not $matchingCdpOnly -or -not $alternateCdpOnly) {
        return $null
    }

    $matchingBootstrapState = Get-ExternalPreflightPageBootstrapTimeoutState -Diagnostic $matchingDiagnostic
    $alternateBootstrapState = Get-ExternalPreflightPageBootstrapTimeoutState -Diagnostic $alternateDiagnostic
    if ($null -eq $matchingBootstrapState -or $null -eq $alternateBootstrapState) {
        return $null
    }

    $normalizedRequestedStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy $RequestedPageBootstrapStrategy
    if ([string]::IsNullOrWhiteSpace($normalizedRequestedStrategy)) {
        return $null
    }

    if ($matchingBootstrapState.PageStrategy -ne $normalizedRequestedStrategy -or $alternateBootstrapState.PageStrategy -ne $normalizedRequestedStrategy) {
        return $null
    }

    if ($matchingBootstrapState.CommandTimeoutMs -lt $MinimumCommandTimeoutMs -or $alternateBootstrapState.CommandTimeoutMs -lt $MinimumCommandTimeoutMs) {
        return $null
    }

    $requestedLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=true' } else { 'requireCdp=false' }
    $alternateLabelValue = if ([bool]$RequestedRequireCdp) { 'requireCdp=false' } else { 'requireCdp=true' }

    return [pscustomobject]@{
        RequestedLabel = $requestedLabel
        AlternateLabel = $alternateLabelValue
        RequestedPageBootstrapStrategy = $normalizedRequestedStrategy
        MatchingSummary = Get-ExternalPreflightDiagnosticSummaryFragment -Diagnostic $matchingDiagnostic
        AlternateSummary = Get-ExternalPreflightDiagnosticSummaryFragment -Diagnostic $alternateDiagnostic
        MatchingBootstrapState = $matchingBootstrapState
        AlternateBootstrapState = $alternateBootstrapState
        MinimumCommandTimeoutMs = $MinimumCommandTimeoutMs
    }
}

function Get-StableExternalPreflightPageBootstrapBlockerConfirmedState {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [int]$MinimumCommandTimeoutMs = 60000
    )

    return Get-StableExternalPreflightPageBootstrapBlockerConfirmedStateForEntryPair -StatusEntries $StatusEntries -MatchingLabel 'matching' -AlternateLabel 'alternate' -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy -MinimumCommandTimeoutMs $MinimumCommandTimeoutMs
}

function Get-StableExternalPreflightAttachedSessionBlockerCandidateState {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [int]$MinimumCommandTimeoutMs = 45000
    )

    $alignedCdpFocusState = Get-StableExternalPreflightAlignedCdpBootstrapFocusState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    if ($null -eq $alignedCdpFocusState) {
        return $null
    }

    $matchingCdpBootstrapState = $alignedCdpFocusState.MatchingCdpBootstrapState
    $alternateCdpBootstrapState = $alignedCdpFocusState.AlternateCdpBootstrapState
    if ($null -eq $matchingCdpBootstrapState -or $null -eq $alternateCdpBootstrapState) {
        return $null
    }

    if ($matchingCdpBootstrapState.CommandTimeoutMs -lt $MinimumCommandTimeoutMs -or $alternateCdpBootstrapState.CommandTimeoutMs -lt $MinimumCommandTimeoutMs) {
        return $null
    }

    return [pscustomobject]@{
        RequestedLabel = $alignedCdpFocusState.RequestedLabel
        AlternateLabel = $alignedCdpFocusState.AlternateLabel
        MatchingSummary = $alignedCdpFocusState.MatchingSummary
        AlternateSummary = $alignedCdpFocusState.AlternateSummary
        MatchingCdpBootstrapState = $matchingCdpBootstrapState
        AlternateCdpBootstrapState = $alternateCdpBootstrapState
        MinimumCommandTimeoutMs = $MinimumCommandTimeoutMs
    }
}

function Get-StableExternalPreflightAttachedSessionBlockerConfirmedState {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [int]$MinimumCommandTimeoutMs = 60000
    )

    return Get-StableExternalPreflightAttachedSessionBlockerCandidateState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -MinimumCommandTimeoutMs $MinimumCommandTimeoutMs
}

function Get-StableExternalPreflightInvocationProfileAlignmentNote {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0 -or $null -eq $RequestedRequireCdp) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $alternateEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate' } | Select-Object -First 1
    if ($null -eq $matchingEntry -or $null -eq $alternateEntry) {
        return $null
    }

    if (-not $matchingEntry.State.Parsed -or -not $alternateEntry.State.Parsed) {
        return $null
    }

    $matchingDiagnostic = Get-ExternalPreflightDiagnostic -Path $matchingEntry.Path
    $alternateDiagnostic = Get-ExternalPreflightDiagnostic -Path $alternateEntry.Path
    if ($null -eq $matchingDiagnostic -or $null -eq $alternateDiagnostic) {
        return $null
    }

    $matchingReachabilityState = Get-ExternalPreflightServiceReachabilityState -Diagnostic $matchingDiagnostic
    $alternateReachabilityState = Get-ExternalPreflightServiceReachabilityState -Diagnostic $alternateDiagnostic
    if ([string]::IsNullOrWhiteSpace($matchingReachabilityState) -or [string]::IsNullOrWhiteSpace($alternateReachabilityState)) {
        return $null
    }

    $requestedLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=true' } else { 'requireCdp=false' }
    $alternateLabel = if ([bool]$RequestedRequireCdp) { 'requireCdp=false' } else { 'requireCdp=true' }
    $matchingSummary = Get-ExternalPreflightDiagnosticSummaryFragment -Diagnostic $matchingDiagnostic
    $alternateSummary = Get-ExternalPreflightDiagnosticSummaryFragment -Diagnostic $alternateDiagnostic
    $alignedCdpFocusState = Get-StableExternalPreflightAlignedCdpBootstrapFocusState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp

    if ($matchingReachabilityState -eq 'services_reachable' -and $alternateReachabilityState -eq 'services_reachable') {
        if ($null -ne $alignedCdpFocusState) {
            return ('Invocation profiles are aligned at the service-reachability layer: both the matched {0} replay and the opposite {1} replay already reach backend/frontend. Matched profile: {2}. Opposite profile: {3}. Keep the next live reruns focused on attached-session CDP bootstrap behavior instead of reopening service reachability work.' -f $requestedLabel, $alternateLabel, $matchingSummary, $alternateSummary)
        }

        return ('Invocation profiles are aligned at the service-reachability layer: both the matched {0} replay and the opposite {1} replay already reach backend/frontend. Matched profile: {2}. Opposite profile: {3}.' -f $requestedLabel, $alternateLabel, $matchingSummary, $alternateSummary)
    }

    if ($matchingReachabilityState -eq 'services_unreachable' -and $alternateReachabilityState -eq 'services_reachable') {
        return ('Invocation profiles are not aligned yet: the matched {0} replay is still blocked before backend/frontend reachability ({1}), while the opposite {2} replay already reaches backend/frontend ({3}). Refresh the {0} external profile before reopening CDP/bootstrap tuning.' -f $requestedLabel, $matchingSummary, $alternateLabel, $alternateSummary)
    }

    if ($matchingReachabilityState -eq 'services_reachable' -and $alternateReachabilityState -eq 'services_unreachable') {
        return ('Invocation profiles are not aligned yet: the matched {0} replay already reaches backend/frontend ({1}), while the opposite {2} replay is still blocked earlier ({3}). Refresh the {2} external profile if you want both invocation paths compared at the same service-reachability layer.' -f $requestedLabel, $matchingSummary, $alternateLabel, $alternateSummary)
    }

    return ('Invocation profiles are aligned at the earlier service-preflight layer: both the matched {0} replay and the opposite {1} replay are still blocked before backend/frontend reachability. Matched profile: {2}. Opposite profile: {3}.' -f $requestedLabel, $alternateLabel, $matchingSummary, $alternateSummary)
}

function Get-StableExternalPreflightHostHandoffNote {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [string]$RecommendedActionClass
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    if ([string]::IsNullOrWhiteSpace($RecommendedActionClass)) {
        return $null
    }

    $pageBootstrapBlockerConfirmedState = Get-StableExternalPreflightPageBootstrapBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    if ($null -ne $pageBootstrapBlockerConfirmedState) {
        return ('Current host handoff: repo-local diagnostics already confirm a page-bootstrap blocker for strategy ''{0}'' at {1} after {2}ms for both invocation profiles. Do not spend another round rerunning the same live command on this host; switch to a different host or a genuinely different execution path.' -f $pageBootstrapBlockerConfirmedState.RequestedPageBootstrapStrategy, $pageBootstrapBlockerConfirmedState.MatchingBootstrapState.Classification, $pageBootstrapBlockerConfirmedState.MinimumCommandTimeoutMs)
    }

    $attachedSessionBlockerConfirmedState = Get-StableExternalPreflightAttachedSessionBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    if ($null -ne $attachedSessionBlockerConfirmedState) {
        return ('Current host handoff: repo-local diagnostics already confirm an attached-session bootstrap blocker at page_bootstrap_timeout_attached_session after {0}ms for both invocation profiles. Do not spend another round rerunning the same attached-session path on this host; switch hosts or move to a genuinely different execution path.' -f $attachedSessionBlockerConfirmedState.MinimumCommandTimeoutMs)
    }

    if ($RecommendedActionClass -ne 'fallback_replay_ready') {
        return $null
    }

    $matchingGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching-generic' } | Select-Object -First 1
    $matchingGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $matchingGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    if ([string]::IsNullOrWhiteSpace($matchingGenericAlternateStrategy)) {
        return $null
    }

    $alternateStrategyBlockerState = Get-StableExternalPreflightPageBootstrapBlockerConfirmedStateForEntryPair -StatusEntries $StatusEntries -MatchingLabel 'matching-generic' -AlternateLabel 'alternate-generic' -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $matchingGenericAlternateStrategy
    if ($null -ne $alternateStrategyBlockerState) {
        return ('Current host handoff: the selected same-expectation fallback replay comes from alternate page-bootstrap strategy ''{0}'', and repo-local diagnostics already confirm that this host also blocks that strategy at {1} after {2}ms. Use the alternate execution path command only on a different host; on this host, further live reruns are unlikely to add new evidence.' -f $matchingGenericAlternateStrategy, $alternateStrategyBlockerState.MatchingBootstrapState.Classification, $alternateStrategyBlockerState.MinimumCommandTimeoutMs)
    }

    return ('Current host handoff: the selected same-expectation fallback replay comes from alternate page-bootstrap strategy ''{0}''. If you need a live rerun, use the alternate execution path command first; if that alternate strategy is already known to fail on this host, move straight to a different host instead of refreshing the same blocked path again.' -f $matchingGenericAlternateStrategy)
}

function Get-StableExternalPreflightDecisionSummary {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [bool]$InvocationStartsServices
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    $selectedEntry = $StatusEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedEntry) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $selectedDescriptor = Get-StableExternalPreflightCopyDescriptor -Entry $selectedEntry
    if ([string]::IsNullOrWhiteSpace($selectedDescriptor)) {
        $selectedDescriptor = 'selected replay copy'
    }

    $alternateDecisionSuffix = ''
    $alternateSummary = Get-StableExternalPreflightAlternateRequirementSpecificCopySummary -StatusEntries $StatusEntries
    if ($null -ne $alternateSummary -and -not [string]::IsNullOrWhiteSpace([string]$alternateSummary.DecisionSummarySentence)) {
        $alternateDecisionSuffix = ' ' + [string]$alternateSummary.DecisionSummarySentence
    }

    $coverageDecisionSuffix = ''
    $coverageDecisionSentence = Get-StableExternalPreflightCoverageDecisionSentence -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    if (-not [string]::IsNullOrWhiteSpace($coverageDecisionSentence)) {
        $coverageDecisionSuffix = ' ' + $coverageDecisionSentence
    }

    $decisionContextSuffix = $coverageDecisionSuffix + $alternateDecisionSuffix

    $selectedFreshnessLabel = [string]$selectedEntry.FreshnessLabel
    $matchingMissing = ($null -eq $matchingEntry) -or (-not $matchingEntry.State.Exists)
    $matchingParseFailed = ($null -ne $matchingEntry) -and $matchingEntry.State.Exists -and (-not $matchingEntry.State.Parsed)
    $matchingGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching-generic' } | Select-Object -First 1
    $matchingGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $matchingGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $alignedCdpFocusState = Get-StableExternalPreflightAlignedCdpBootstrapFocusState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $pageBootstrapBlockerConfirmedState = Get-StableExternalPreflightPageBootstrapBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $attachedSessionBlockerConfirmedState = Get-StableExternalPreflightAttachedSessionBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $attachedSessionBlockerCandidateState = Get-StableExternalPreflightAttachedSessionBlockerCandidateState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $refreshScope = if ($InvocationStartsServices) {
        'rerun the preferred refresh command on this host once local services are reachable'
    }
    else {
        'rerun the preferred refresh command on a healthy host or after exposing the target services'
    }

    if ([string]::IsNullOrWhiteSpace($selectedFreshnessLabel)) {
        if ($selectedEntry.Label -eq 'matching') {
            return 'refresh or inspect the saved diagnostic before relying on this invocation; the selected preview matches this invocation but its freshness could not be classified from the saved diagnostic.' + $decisionContextSuffix
        }

        if ($matchingParseFailed) {
            return ('refresh or inspect the saved diagnostic before relying on this replay; the matching requirement-specific copy is not parseable yet and the selected {0} could not be freshness-classified.' -f $selectedDescriptor) + $decisionContextSuffix
        }

        if ($matchingMissing) {
            return ('refresh or inspect the saved diagnostic before relying on this replay; this invocation still has no matching requirement-specific copy and the selected {0} could not be freshness-classified.' -f $selectedDescriptor) + $decisionContextSuffix
        }

        return ('refresh or inspect the saved diagnostic before relying on this replay; the selected {0} could not be freshness-classified.' -f $selectedDescriptor) + $decisionContextSuffix
    }

    if ($selectedEntry.Label -eq 'matching') {
        if ($selectedFreshnessLabel -like 'stale*') {
            return ('refresh external evidence before relying on this invocation; the selected preview matches this invocation but is {0}, so {1}.' -f $selectedFreshnessLabel, $refreshScope) + $decisionContextSuffix
        }

        if ($null -ne $pageBootstrapBlockerConfirmedState) {
            return ('keep using the repo-local replay as confirmed blocker evidence; the selected preview already matches this invocation and remains {0}, and both invocation profiles still stop at {1} after {2}ms using page-bootstrap strategy ''{3}'' ({4} vs {5}). This host is now confirmed as a page-bootstrap blocker for strategy ''{3}''; stop tuning this wrapper on this host and switch to a different host or a genuinely different execution path.' -f $selectedFreshnessLabel, $pageBootstrapBlockerConfirmedState.MatchingBootstrapState.Classification, $pageBootstrapBlockerConfirmedState.MinimumCommandTimeoutMs, $pageBootstrapBlockerConfirmedState.RequestedPageBootstrapStrategy, $pageBootstrapBlockerConfirmedState.MatchingSummary, $pageBootstrapBlockerConfirmedState.AlternateSummary) + $decisionContextSuffix
        }

        if ($null -ne $attachedSessionBlockerConfirmedState) {
            return ('keep using the repo-local replay as confirmed blocker evidence; the selected preview already matches this invocation and remains {0}, and both invocation profiles still stop at page_bootstrap_timeout_attached_session after {1}ms attached-session bootstrap timeouts ({2} vs {3}). This host is now confirmed as an attached-session bootstrap blocker; stop tuning this wrapper and record the host-level blocker unless a different execution path becomes available.' -f $selectedFreshnessLabel, $attachedSessionBlockerConfirmedState.MinimumCommandTimeoutMs, $attachedSessionBlockerConfirmedState.MatchingSummary, $attachedSessionBlockerConfirmedState.AlternateSummary) + $decisionContextSuffix
        }

        if ($null -ne $attachedSessionBlockerCandidateState) {
            return ('keep using the repo-local replay as blocker evidence; the selected preview already matches this invocation and remains {0}, and both invocation profiles still stop at page_bootstrap_timeout_attached_session after {1}ms attached-session bootstrap timeouts ({2} vs {3}). Only run one final bounded timeout rerun before classifying this host as an attached-session bootstrap blocker.' -f $selectedFreshnessLabel, $attachedSessionBlockerCandidateState.MinimumCommandTimeoutMs, $attachedSessionBlockerCandidateState.MatchingSummary, $attachedSessionBlockerCandidateState.AlternateSummary) + $decisionContextSuffix
        }

        if ($null -ne $alignedCdpFocusState) {
            return ('keep using the repo-local replay to compare attached-session CDP bootstrap behavior; the selected preview already matches this invocation and remains {0}, and both invocation profiles are now aligned at backend/frontend reachability with CDP-only failure ({1} vs {2}).' -f $selectedFreshnessLabel, $alignedCdpFocusState.MatchingSummary, $alignedCdpFocusState.AlternateSummary) + $decisionContextSuffix
        }

        return ('keep using the repo-local replay for blocked-host triage; the selected preview already matches this invocation and remains {0}.' -f $selectedFreshnessLabel) + $decisionContextSuffix
    }

    if ($matchingParseFailed) {
        if ($selectedFreshnessLabel -like 'stale*') {
            return ('refresh before relying on this replay; the matching requirement-specific copy exists but is not parseable, and the selected {0} is {1}, so {2}.' -f $selectedDescriptor, $selectedFreshnessLabel, $refreshScope) + $decisionContextSuffix
        }

        return ('keep using the fallback replay for blocked-host triage, but refresh before treating it as requirement-specific evidence; the matching requirement-specific copy is not parseable yet, while the selected {0} remains {1}.' -f $selectedDescriptor, $selectedFreshnessLabel) + $decisionContextSuffix
    }

    if ($matchingMissing) {
        if (-not [string]::IsNullOrWhiteSpace($matchingGenericAlternateStrategy)) {
            $preferenceState = Get-StableExternalPreflightSelectedPreferenceState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
            $preferenceCaution = Get-StableExternalPreflightSelectedPreferenceFallbackCaution -PreferenceState $preferenceState
            if ($selectedFreshnessLabel -like 'stale*') {
                return ('refresh before relying on this replay; this invocation still has no matching requirement-specific copy for page-bootstrap strategy ''{0}'', and the selected same-expectation copy from alternate page-bootstrap strategy ''{1}'' is {2}, so {3}.{4}' -f $RequestedPageBootstrapStrategy, $matchingGenericAlternateStrategy, $selectedFreshnessLabel, $refreshScope, $preferenceCaution) + $decisionContextSuffix
            }

            return ('keep using the same-expectation alternate-strategy replay for blocked-host triage; this invocation still has no matching requirement-specific copy for page-bootstrap strategy ''{0}'', but repo-local diagnostics already include a {2} same-expectation copy for alternate page-bootstrap strategy ''{1}''. Run the alternate execution path command before refreshing if you want a live rerun for that saved strategy.{3}' -f $RequestedPageBootstrapStrategy, $matchingGenericAlternateStrategy, $selectedFreshnessLabel, $preferenceCaution) + $decisionContextSuffix
        }

        if ($selectedFreshnessLabel -like 'stale*') {
            return ('refresh before relying on this replay; this invocation still has no matching requirement-specific copy, and the selected {0} is {1}, so {2}.' -f $selectedDescriptor, $selectedFreshnessLabel, $refreshScope) + $decisionContextSuffix
        }

        return ('keep using the fallback replay for blocked-host triage, but refresh before treating it as requirement-specific evidence; this invocation still has no matching requirement-specific copy, even though the selected {0} remains {1}.' -f $selectedDescriptor, $selectedFreshnessLabel) + $decisionContextSuffix
    }

    if ($selectedFreshnessLabel -like 'stale*') {
        return ('refresh before relying on this replay; the selected {0} is {1}, so {2}.' -f $selectedDescriptor, $selectedFreshnessLabel, $refreshScope) + $decisionContextSuffix
    }

    return ('keep using the repo-local replay for blocked-host triage; the selected {0} remains {1}.' -f $selectedDescriptor, $selectedFreshnessLabel) + $decisionContextSuffix
}

function Get-StableExternalPreflightRecommendedActionClass {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [bool]$InvocationStartsServices
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    $selectedEntry = $StatusEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedEntry) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $selectedFreshnessLabel = [string]$selectedEntry.FreshnessLabel
    $matchingMissing = ($null -eq $matchingEntry) -or (-not $matchingEntry.State.Exists)
    $matchingParseFailed = ($null -ne $matchingEntry) -and $matchingEntry.State.Exists -and (-not $matchingEntry.State.Parsed)
    $matchingGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching-generic' } | Select-Object -First 1
    $matchingGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $matchingGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $alignedCdpFocusState = Get-StableExternalPreflightAlignedCdpBootstrapFocusState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $pageBootstrapBlockerConfirmedState = Get-StableExternalPreflightPageBootstrapBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $attachedSessionBlockerConfirmedState = Get-StableExternalPreflightAttachedSessionBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $attachedSessionBlockerCandidateState = Get-StableExternalPreflightAttachedSessionBlockerCandidateState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp

    if ([string]::IsNullOrWhiteSpace($selectedFreshnessLabel)) {
        if ($selectedEntry.Label -eq 'matching') {
            return 'saved_copy_inspection_required'
        }

        if ($matchingParseFailed -or $matchingMissing) {
            return 'fallback_only_refresh_required'
        }

        return 'saved_copy_inspection_required'
    }

    if ($selectedEntry.Label -eq 'matching') {
        if ($selectedFreshnessLabel -like 'stale*') {
            return 'matched_replay_refresh_recommended'
        }

        if ($null -ne $pageBootstrapBlockerConfirmedState) {
            return 'page_bootstrap_strategy_blocker_confirmed'
        }

        if ($null -ne $attachedSessionBlockerConfirmedState) {
            return 'attached_session_bootstrap_blocker_confirmed'
        }

        if ($null -ne $attachedSessionBlockerCandidateState) {
            return 'attached_session_bootstrap_blocker_candidate'
        }

        if ($null -ne $alignedCdpFocusState) {
            return 'aligned_cdp_bootstrap_focus'
        }

        return 'matched_replay_ready'
    }

    if ($matchingParseFailed -or $matchingMissing) {
        if ($matchingMissing -and -not [string]::IsNullOrWhiteSpace($matchingGenericAlternateStrategy) -and -not ($selectedFreshnessLabel -like 'stale*')) {
            return 'fallback_replay_ready'
        }

        return 'fallback_only_refresh_required'
    }

    if ($selectedFreshnessLabel -like 'stale*') {
        return 'saved_copy_refresh_recommended'
    }

    return 'fallback_replay_ready'
}

function Get-StableExternalPreflightRecommendedAction {
    param(
        [object[]]$StatusEntries,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [bool]$InvocationStartsServices
    )

    if ($null -eq $StatusEntries -or $StatusEntries.Count -eq 0) {
        return $null
    }

    $selectedEntry = $StatusEntries | Where-Object { $_.Selected } | Select-Object -First 1
    if ($null -eq $selectedEntry) {
        return $null
    }

    $matchingEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching' } | Select-Object -First 1
    $alternateEntry = $StatusEntries | Where-Object { $_.Label -eq 'alternate' } | Select-Object -First 1

    $selectedDescriptor = Get-StableExternalPreflightCopyDescriptor -Entry $selectedEntry
    if ([string]::IsNullOrWhiteSpace($selectedDescriptor)) {
        $selectedDescriptor = 'selected replay copy'
    }

    $selectedFreshnessLabel = [string]$selectedEntry.FreshnessLabel
    $matchingMissing = ($null -eq $matchingEntry) -or (-not $matchingEntry.State.Exists)
    $matchingParseFailed = ($null -ne $matchingEntry) -and $matchingEntry.State.Exists -and (-not $matchingEntry.State.Parsed)
    $matchingGenericEntry = $StatusEntries | Where-Object { $_.Label -eq 'matching-generic' } | Select-Object -First 1
    $matchingGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $matchingGenericEntry -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $alternateParseable = ($null -ne $alternateEntry) -and $alternateEntry.State.Parsed
    $alignedCdpFocusState = Get-StableExternalPreflightAlignedCdpBootstrapFocusState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $pageBootstrapBlockerConfirmedState = Get-StableExternalPreflightPageBootstrapBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $attachedSessionBlockerConfirmedState = Get-StableExternalPreflightAttachedSessionBlockerConfirmedState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $attachedSessionBlockerCandidateState = Get-StableExternalPreflightAttachedSessionBlockerCandidateState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp
    $refreshScope = if ($InvocationStartsServices) {
        'run the preferred refresh command on this host once local services are reachable'
    }
    else {
        'run the preferred refresh command on a healthy host or after exposing the target services'
    }

    if ([string]::IsNullOrWhiteSpace($selectedFreshnessLabel)) {
        if ($selectedEntry.Label -eq 'matching') {
            return 'Inspect or refresh the matched saved diagnostic before relying on this invocation, because its freshness could not be classified yet.'
        }

        if ($matchingParseFailed) {
            return ('Inspect or refresh the matching requirement-specific copy before relying on this replay; until then, treat the selected {0} as fallback-only context.' -f $selectedDescriptor)
        }

        if ($matchingMissing) {
            return ('Treat the selected {0} as fallback-only context and {1} to capture a requirement-specific saved diagnostic for this invocation.' -f $selectedDescriptor, $refreshScope)
        }

        return ('Inspect or refresh the selected {0} before relying on this replay.' -f $selectedDescriptor)
    }

    if ($selectedEntry.Label -eq 'matching') {
        if ($selectedFreshnessLabel -like 'stale*') {
            return ('Refresh the matched saved diagnostic before relying on this invocation; {0} because the current copy is {1}.' -f $refreshScope, $selectedFreshnessLabel)
        }

        if ($null -ne $pageBootstrapBlockerConfirmedState) {
            return ('Treat this host as a confirmed page-bootstrap blocker for strategy ''{0}'': both invocation profiles already reach backend/frontend, and both saved replays still fail at {1} after {2}ms. Stop tuning this wrapper on this host and switch to a different host or a genuinely different execution path.' -f $pageBootstrapBlockerConfirmedState.RequestedPageBootstrapStrategy, $pageBootstrapBlockerConfirmedState.MatchingBootstrapState.Classification, $pageBootstrapBlockerConfirmedState.MinimumCommandTimeoutMs)
        }

        if ($null -ne $attachedSessionBlockerConfirmedState) {
            return ('Treat this host as a confirmed attached-session bootstrap blocker: both invocation profiles already reach backend/frontend, and both saved replays still fail at page_bootstrap_timeout_attached_session after {0}ms. Stop tuning this wrapper and record the host-level blocker unless you can switch to a different execution path.' -f $attachedSessionBlockerConfirmedState.MinimumCommandTimeoutMs)
        }

        if ($null -ne $attachedSessionBlockerCandidateState) {
            return ('Treat this host as an attached-session bootstrap blocker candidate: both invocation profiles already reach backend/frontend, and both saved replays still fail at page_bootstrap_timeout_attached_session after {0}ms. Only run one final bounded timeout comparison if you still need stronger host-level proof; otherwise stop tuning this wrapper and record the host-level blocker.' -f $attachedSessionBlockerCandidateState.MinimumCommandTimeoutMs)
        }

        if ($null -ne $alignedCdpFocusState) {
            return 'Keep using the matched repo-local replay as the baseline for CDP/bootstrap comparison; both invocation profiles already reach backend/frontend and now fail only at the attached-session CDP bootstrap layer. Rerun live only when you need fresher CDP evidence or want to compare command order/timeout changes on this host.'
        }

        if ($null -ne $RequestedRequireCdp -and $alternateParseable) {
            return ('Keep using the matched repo-local replay for blocked-host triage for now; refresh only when you need newer live evidence.' )
        }

        return 'Keep using the matched repo-local replay for blocked-host triage for now; refresh only when you need newer live evidence.'
    }

    if ($matchingParseFailed) {
        if ($selectedFreshnessLabel -like 'stale*') {
            return ('Refresh the matching requirement-specific copy before relying on this replay; the selected {0} is only {1} fallback context.' -f $selectedDescriptor, $selectedFreshnessLabel)
        }

        return ('Keep using the selected {0} only as fallback context, then refresh the matching requirement-specific copy before treating this replay as invocation-specific evidence.' -f $selectedDescriptor)
    }

    if ($matchingMissing) {
        if (-not [string]::IsNullOrWhiteSpace($matchingGenericAlternateStrategy)) {
            $preferenceState = Get-StableExternalPreflightSelectedPreferenceState -StatusEntries $StatusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
            $preferenceCaution = Get-StableExternalPreflightSelectedPreferenceFallbackCaution -PreferenceState $preferenceState
            if ($selectedFreshnessLabel -like 'stale*') {
                return ('Refresh before relying on this replay; no matching requirement-specific copy exists yet for page-bootstrap strategy ''{0}'', and the selected same-expectation copy from alternate page-bootstrap strategy ''{1}'' is {2}.{3}' -f $RequestedPageBootstrapStrategy, $matchingGenericAlternateStrategy, $selectedFreshnessLabel, $preferenceCaution)
            }

            return ('Keep using the selected {0} as same-expectation evidence from alternate page-bootstrap strategy ''{1}'', then run the alternate execution path command if you want a live rerun for that saved strategy. Use the preferred refresh command only if you specifically need a saved diagnostic for page-bootstrap strategy ''{2}''.{3}' -f $selectedDescriptor, $matchingGenericAlternateStrategy, $RequestedPageBootstrapStrategy, $preferenceCaution)
        }

        if ($selectedFreshnessLabel -like 'stale*') {
            return ('Refresh requirement-specific evidence before relying on this replay; the selected {0} is only {1} fallback context and no matching saved copy exists yet.' -f $selectedDescriptor, $selectedFreshnessLabel)
        }

        return ('Keep using the selected {0} only as fallback context, then {1} to capture a matching requirement-specific saved diagnostic for this invocation.' -f $selectedDescriptor, $refreshScope)
    }

    if ($selectedFreshnessLabel -like 'stale*') {
        return ('Refresh the selected {0} before relying on this replay; {1} because it is {2}.' -f $selectedDescriptor, $refreshScope, $selectedFreshnessLabel)
    }

    return ('Keep using the selected {0} for blocked-host triage for now; refresh only when you need newer live evidence.' -f $selectedDescriptor)
}

function Get-StableExternalPreflightCopyStatusLines {
    param(
        [object]$RequestedRequireCdp,
        [string]$SelectedPath,
        $Selection,
        [int]$FreshnessThresholdHours = 24
    )

    if ($null -eq $Selection) {
        $Selection = Get-StableExternalPreflightCopySelectionPaths -RequestedRequireCdp $RequestedRequireCdp
    }

    $statusLines = [System.Collections.Generic.List[string]]::new()
    foreach ($entry in @(Get-StableExternalPreflightCopyStatusEntries -SelectedPath $SelectedPath -Selection $Selection -FreshnessThresholdHours $FreshnessThresholdHours)) {
        $selectedSuffix = ''
        if ($entry.Selected) {
            $selectedSuffix = ' (selected for preview)'
        }

        $stateDescriptor = [string]$entry.StateLabel
        if ([string]$entry.Label -eq 'matching') {
            switch ([string]$entry.MismatchReason) {
                'page-strategy' {
                    $stateDescriptor += '; strategy mismatch for this invocation'
                }
                'page-strategy-missing' {
                    $stateDescriptor += '; strategy missing in saved diagnostic for this invocation'
                }
            }
        }
        elseif ([string]$entry.Label -eq 'matching-generic') {
            $alternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $entry -RequestedPageBootstrapStrategy $Selection.RequestedPageBootstrapStrategy
            if (-not [string]::IsNullOrWhiteSpace($alternateStrategy)) {
                $stateDescriptor += ('; saved with alternate page-bootstrap strategy ''{0}'' for this invocation' -f $alternateStrategy)
            }
        }
        elseif ([string]$entry.Label -eq 'alternate-generic') {
            $alternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $entry -RequestedPageBootstrapStrategy $Selection.RequestedPageBootstrapStrategy
            if (-not [string]::IsNullOrWhiteSpace($alternateStrategy)) {
                $stateDescriptor += ('; saved with opposite CDP expectation for alternate page-bootstrap strategy ''{0}'' for this invocation' -f $alternateStrategy)
            }
        }

        $statusLines.Add(('External preflight {0} stable diagnostic copy status: {1}{2}{3} [{4}]' -f $entry.Label, $stateDescriptor, $entry.FreshnessSuffix, $selectedSuffix, $entry.Path))
    }

    return $statusLines.ToArray()
}

function Get-ExternalPreflightDiagnosticPathFromMessage {
    param([string]$Message)

    if ([string]::IsNullOrWhiteSpace($Message)) {
        return $null
    }

    $match = [regex]::Match($Message, '(?m)^Diagnostic:\s*(.+external-preflight-diagnostic\.json)\s*$')
    if (-not $match.Success) {
        return $null
    }

    return $match.Groups[1].Value.Trim()
}

function Get-ExternalPreflightCdpDiagnosticPathFromMessage {
    param([string]$Message)

    if ([string]::IsNullOrWhiteSpace($Message)) {
        return $null
    }

    $match = [regex]::Match($Message, '(?m)^CDP diagnostic file:\s*(.+cdp\.diagnostic\.json)\s*$')
    if (-not $match.Success) {
        return $null
    }

    return $match.Groups[1].Value.Trim()
}

function Get-ExternalPreflightCdpDiagnosticFromMessage {
    param([string]$Message)

    if ([string]::IsNullOrWhiteSpace($Message)) {
        return $null
    }

    $classificationMatch = [regex]::Match($Message, '(?m)^CDP diagnostic classification:\s*(.+)\s*$')
    $errorMatch = [regex]::Match($Message, '(?m)^CDP diagnostic error:\s*(.+)\s*$')
    if ((-not $classificationMatch.Success) -and (-not $errorMatch.Success)) {
        return $null
    }

    $pageModeMatch = [regex]::Match($Message, '(?m)^CDP diagnostic page mode:\s*(.+)\s*$')
    $pageStrategyMatch = [regex]::Match($Message, '(?m)^CDP diagnostic page strategy:\s*(.+)\s*$')
    $bootstrapOrderMatch = [regex]::Match($Message, '(?m)^CDP diagnostic bootstrap order:\s*(.+)\s*$')
    $commandTimeoutMatch = [regex]::Match($Message, '(?m)^CDP diagnostic command timeout \(ms\):\s*(\d+)\s*$')
    $hintMatch = [regex]::Match($Message, '(?m)^CDP diagnostic hint:\s*(.+)\s*$')

    $commandTimeoutMs = $null
    if ($commandTimeoutMatch.Success) {
        try {
            $commandTimeoutMs = [int]$commandTimeoutMatch.Groups[1].Value.Trim()
        }
        catch {
            $commandTimeoutMs = $null
        }
    }

    return [pscustomobject]@{
        classification = $(if ($classificationMatch.Success) { $classificationMatch.Groups[1].Value.Trim() } else { $null })
        errorName = $(if ($errorMatch.Success) { $errorMatch.Groups[1].Value.Trim() } else { $null })
        pageMode = $(if ($pageModeMatch.Success) { $pageModeMatch.Groups[1].Value.Trim() } else { $null })
        pageStrategy = $(if ($pageStrategyMatch.Success) { $pageStrategyMatch.Groups[1].Value.Trim() } else { $null })
        bootstrapCommandOrder = $(if ($bootstrapOrderMatch.Success) { $bootstrapOrderMatch.Groups[1].Value.Trim() } else { $null })
        commandTimeoutMs = $commandTimeoutMs
        hint = $(if ($hintMatch.Success) { $hintMatch.Groups[1].Value.Trim() } else { $null })
    }
}

function Get-AdjacentExternalPreflightDiagnosticPathFromCdpDiagnosticPath {
    param([string]$CdpDiagnosticPath)

    if ([string]::IsNullOrWhiteSpace($CdpDiagnosticPath)) {
        return $null
    }

    try {
        $parent = Split-Path -Parent $CdpDiagnosticPath
        if ([string]::IsNullOrWhiteSpace($parent)) {
            return $null
        }

        $candidatePath = Join-Path $parent 'external-preflight-diagnostic.json'
        if (Test-Path -LiteralPath $candidatePath) {
            return $candidatePath
        }
    }
    catch {
        return $null
    }

    return $null
}

function Get-ExternalPreflightDiagnosticBridgePathFromCdpDiagnosticPath {
    param([string]$CdpDiagnosticPath)

    if ([string]::IsNullOrWhiteSpace($CdpDiagnosticPath)) {
        return $null
    }

    try {
        $parent = Split-Path -Parent $CdpDiagnosticPath
        if ([string]::IsNullOrWhiteSpace($parent)) {
            return $null
        }

        return Join-Path $parent 'external-preflight-diagnostic.cdp-bridge.json'
    }
    catch {
        return $null
    }
}

function Write-ExternalPreflightDiagnosticJsonArtifact {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)]$Value
    )

    $parent = Split-Path -Parent $Path
    if (-not [string]::IsNullOrWhiteSpace($parent)) {
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
    }

    $json = $Value | ConvertTo-Json -Depth 10
    [System.IO.File]::WriteAllText($Path, $json, [System.Text.UTF8Encoding]::new($false))
}

function Get-ExternalPreflightSummaryLineFromEntry {
    param($Entry)

    if ($null -eq $Entry) {
        return $null
    }

    $label = [string]$Entry.label
    if ([string]::IsNullOrWhiteSpace($label)) {
        $label = 'external check'
    }

    $url = [string]$Entry.url
    if ([bool]$Entry.reachable) {
        if ([string]::IsNullOrWhiteSpace($url)) {
            return ('- {0}: reachable during this diagnostic run' -f $label)
        }

        return ('- {0}: reachable at {1} during this diagnostic run' -f $label, $url)
    }

    $classification = [string]$Entry.classification
    $lastError = [string]$Entry.lastError
    if (-not [string]::IsNullOrWhiteSpace($url)) {
        if (-not [string]::IsNullOrWhiteSpace($classification) -and -not [string]::IsNullOrWhiteSpace($lastError)) {
            return ('- {0}: unreachable at {1} ({2}, {3})' -f $label, $url, $classification, $lastError)
        }

        if (-not [string]::IsNullOrWhiteSpace($classification)) {
            return ('- {0}: unreachable at {1} ({2})' -f $label, $url, $classification)
        }

        if (-not [string]::IsNullOrWhiteSpace($lastError)) {
            return ('- {0}: unreachable at {1} ({2})' -f $label, $url, $lastError)
        }

        return ('- {0}: unreachable at {1}' -f $label, $url)
    }

    if (-not [string]::IsNullOrWhiteSpace($classification) -and -not [string]::IsNullOrWhiteSpace($lastError)) {
        return ('- {0}: unreachable ({1}, {2})' -f $label, $classification, $lastError)
    }

    if (-not [string]::IsNullOrWhiteSpace($classification)) {
        return ('- {0}: unreachable ({1})' -f $label, $classification)
    }

    if (-not [string]::IsNullOrWhiteSpace($lastError)) {
        return ('- {0}: unreachable ({1})' -f $label, $lastError)
    }

    return ('- {0}: unreachable' -f $label)
}

function Get-ExternalPreflightCdpFailureSummaryLine {
    param(
        $CdpDiagnostic,
        [string]$CdpUrl
    )

    if ($null -eq $CdpDiagnostic) {
        return $null
    }

    $detailParts = @()
    foreach ($detail in @(
        [string]$CdpDiagnostic.classification,
        [string]$CdpDiagnostic.errorName,
        $(if (-not [string]::IsNullOrWhiteSpace([string]$CdpDiagnostic.pageMode)) { 'page mode ' + [string]$CdpDiagnostic.pageMode } else { $null }),
        $(if (-not [string]::IsNullOrWhiteSpace([string]$CdpDiagnostic.pageStrategy)) { 'page strategy ' + [string]$CdpDiagnostic.pageStrategy } else { $null }),
        $(if (-not [string]::IsNullOrWhiteSpace([string]$CdpDiagnostic.bootstrapCommandOrder)) { 'bootstrap order ' + [string]$CdpDiagnostic.bootstrapCommandOrder } else { $null }),
        $(if ($null -ne $CdpDiagnostic.commandTimeoutMs -and [int]$CdpDiagnostic.commandTimeoutMs -gt 0) { 'command timeout ' + [string]$CdpDiagnostic.commandTimeoutMs + 'ms' } else { $null })
    )) {
        if (-not [string]::IsNullOrWhiteSpace([string]$detail)) {
            $detailParts += [string]$detail
        }
    }

    $detailSuffix = if ($detailParts.Count -gt 0) { ' (' + ($detailParts -join '; ') + ')' } else { '' }
    $normalizedCdpUrl = [string]$CdpUrl
    if ([string]::IsNullOrWhiteSpace($normalizedCdpUrl)) {
        return ('- Edge CDP page bootstrap: failed during this diagnostic run{0}' -f $detailSuffix)
    }

    return ('- Edge CDP page bootstrap: failed at {0}{1}' -f $normalizedCdpUrl.TrimEnd('/'), $detailSuffix)
}

function New-ExternalPreflightDiagnosticFromCdpDiagnostic {
    param(
        [string]$CdpDiagnosticPath,
        [string]$Message,
        [object]$RequestedRequireCdp,
        [string]$FrontendUrl,
        [string]$BackendUrl,
        [string]$CdpUrl
    )

    if ([string]::IsNullOrWhiteSpace($CdpDiagnosticPath) -and [string]::IsNullOrWhiteSpace($Message)) {
        return $null
    }

    $cdpDiagnostic = $null
    if (-not [string]::IsNullOrWhiteSpace($CdpDiagnosticPath) -and (Test-Path -LiteralPath $CdpDiagnosticPath)) {
        $cdpDiagnostic = Get-ExternalPreflightDiagnostic -Path $CdpDiagnosticPath
    }
    if ($null -eq $cdpDiagnostic) {
        $cdpDiagnostic = Get-ExternalPreflightCdpDiagnosticFromMessage -Message $Message
    }
    if ($null -eq $cdpDiagnostic) {
        return $null
    }

    $bridgePath = Get-ExternalPreflightDiagnosticBridgePathFromCdpDiagnosticPath -CdpDiagnosticPath $CdpDiagnosticPath
    if ([string]::IsNullOrWhiteSpace($bridgePath)) {
        return $null
    }

    $adjacentDiagnosticPath = Get-AdjacentExternalPreflightDiagnosticPathFromCdpDiagnosticPath -CdpDiagnosticPath $CdpDiagnosticPath
    $baseDiagnostic = Get-ExternalPreflightDiagnostic -Path $adjacentDiagnosticPath

    $checkedAtRaw = $null
    try {
        if (-not [string]::IsNullOrWhiteSpace($CdpDiagnosticPath) -and (Test-Path -LiteralPath $CdpDiagnosticPath)) {
            $checkedAtRaw = (Get-Item -LiteralPath $CdpDiagnosticPath).LastWriteTime.ToString('o')
        }
    }
    catch {
        $checkedAtRaw = $null
    }
    if ([string]::IsNullOrWhiteSpace($checkedAtRaw) -and $null -ne $baseDiagnostic -and -not [string]::IsNullOrWhiteSpace([string]$baseDiagnostic.checkedAt)) {
        $checkedAtRaw = [string]$baseDiagnostic.checkedAt
    }
    if ([string]::IsNullOrWhiteSpace($checkedAtRaw)) {
        $checkedAtRaw = (Get-Date).ToString('o')
    }

    $resolvedFrontendUrl = if ([string]::IsNullOrWhiteSpace($FrontendUrl)) { $null } else { $FrontendUrl.TrimEnd('/') }
    $resolvedBackendUrl = if ([string]::IsNullOrWhiteSpace($BackendUrl)) { $null } else { $BackendUrl.TrimEnd('/') }
    $resolvedCdpUrl = if ([string]::IsNullOrWhiteSpace($CdpUrl)) { $null } else { $CdpUrl.TrimEnd('/') }
    $requireCdp = $false
    if ($null -ne $RequestedRequireCdp) {
        $requireCdp = [bool]$RequestedRequireCdp
    }

    $backendEntry = if ($null -ne $baseDiagnostic -and $null -ne $baseDiagnostic.backend) {
        $baseDiagnostic.backend
    }
    else {
        [ordered]@{
            label = 'backend healthcheck'
            url = $(if ([string]::IsNullOrWhiteSpace($resolvedBackendUrl)) { $null } else { $resolvedBackendUrl + '/healthz' })
            reachable = $true
            classification = 'reachable'
            statusCode = 200
            statusDescription = 'reachable during this diagnostic run'
            attempts = 1
            elapsedMs = 0
            lastError = $null
            checkedAt = $checkedAtRaw
        }
    }

    $frontendEntry = if ($null -ne $baseDiagnostic -and $null -ne $baseDiagnostic.frontend) {
        $baseDiagnostic.frontend
    }
    else {
        [ordered]@{
            label = 'frontend shell'
            url = $resolvedFrontendUrl
            reachable = $true
            classification = 'reachable'
            statusCode = 200
            statusDescription = 'reachable during this diagnostic run'
            attempts = 1
            elapsedMs = 0
            lastError = $null
            checkedAt = $checkedAtRaw
        }
    }

    $cdpLastErrorParts = @()
    foreach ($value in @([string]$cdpDiagnostic.errorName, [string]$cdpDiagnostic.hint)) {
        if (-not [string]::IsNullOrWhiteSpace([string]$value)) {
            $cdpLastErrorParts += [string]$value
        }
    }
    $cdpLastError = if ($cdpLastErrorParts.Count -gt 0) {
        $cdpLastErrorParts -join ' | '
    }
    else {
        'Edge CDP page bootstrap failed after the external preflight had already passed.'
    }

    $cdpEntry = [ordered]@{
        label = 'Edge CDP page bootstrap'
        url = $resolvedCdpUrl
        reachable = $false
        classification = [string]$cdpDiagnostic.classification
        statusCode = $null
        statusDescription = $null
        attempts = 1
        elapsedMs = 0
        lastError = $cdpLastError
        checkedAt = $checkedAtRaw
    }

    $summaryLines = [System.Collections.Generic.List[string]]::new()
    $baseSummaryLines = @()
    if ($null -ne $baseDiagnostic) {
        $baseSummaryLines = @($baseDiagnostic.summaryLines)
    }
    if ($baseSummaryLines.Count -gt 0) {
        foreach ($line in $baseSummaryLines) {
            if (-not [string]::IsNullOrWhiteSpace([string]$line)) {
                $summaryLines.Add([string]$line)
            }
        }
    }
    else {
        foreach ($entry in @($backendEntry, $frontendEntry)) {
            $line = Get-ExternalPreflightSummaryLineFromEntry -Entry $entry
            if (-not [string]::IsNullOrWhiteSpace($line)) {
                $summaryLines.Add($line)
            }
        }

        if ($null -ne $baseDiagnostic -and $null -ne $baseDiagnostic.cdp) {
            $cdpVersionLine = Get-ExternalPreflightSummaryLineFromEntry -Entry $baseDiagnostic.cdp
            if (-not [string]::IsNullOrWhiteSpace($cdpVersionLine)) {
                $summaryLines.Add($cdpVersionLine)
            }
        }
    }

    $cdpFailureLine = Get-ExternalPreflightCdpFailureSummaryLine -CdpDiagnostic $cdpDiagnostic -CdpUrl $resolvedCdpUrl
    if (-not [string]::IsNullOrWhiteSpace($cdpFailureLine)) {
        $summaryLines.Add($cdpFailureLine)
    }

    $hints = [System.Collections.Generic.List[string]]::new()
    $hints.Add('Backend and frontend were already reachable during this run; keep the same external command and focus on the Edge CDP page bootstrap path.')
    if ($null -ne $baseDiagnostic -and $null -ne $baseDiagnostic.cdp -and [bool]$baseDiagnostic.cdp.reachable) {
        $hints.Add('The external CDP version endpoint already responded during preflight; this failure is downstream of basic CDP reachability.')
    }
    if (-not [string]::IsNullOrWhiteSpace([string]$cdpDiagnostic.hint)) {
        $hints.Add([string]$cdpDiagnostic.hint)
    }

    $bridgeDiagnostic = [ordered]@{
        schemaVersion = 2
        checkedAt = $checkedAtRaw
        requireCdp = $requireCdp
        failedChecks = @('cdp')
        skippedChecks = @()
        overallClassification = 'cdp_smoke_failed'
        primaryBlockingCheck = 'cdp'
        summaryLines = $summaryLines.ToArray()
        hints = $hints.ToArray()
        checkDetails = @(
            [ordered]@{
                name = 'backend'
                label = [string]$backendEntry.label
                url = [string]$backendEntry.url
                reachable = [bool]$backendEntry.reachable
                classification = [string]$backendEntry.classification
                lastError = [string]$backendEntry.lastError
            },
            [ordered]@{
                name = 'frontend'
                label = [string]$frontendEntry.label
                url = [string]$frontendEntry.url
                reachable = [bool]$frontendEntry.reachable
                classification = [string]$frontendEntry.classification
                lastError = [string]$frontendEntry.lastError
            },
            [ordered]@{
                name = 'cdp'
                label = [string]$cdpEntry.label
                url = [string]$cdpEntry.url
                reachable = [bool]$cdpEntry.reachable
                classification = [string]$cdpEntry.classification
                lastError = [string]$cdpEntry.lastError
            }
        )
        frontend = $frontendEntry
        backend = $backendEntry
        cdp = $cdpEntry
        cdpDiagnostic = [ordered]@{
            path = $CdpDiagnosticPath
            classification = [string]$cdpDiagnostic.classification
            errorName = [string]$cdpDiagnostic.errorName
            pageMode = [string]$cdpDiagnostic.pageMode
            pageStrategy = [string]$cdpDiagnostic.pageStrategy
            bootstrapCommandOrder = [string]$cdpDiagnostic.bootstrapCommandOrder
            commandTimeoutMs = $cdpDiagnostic.commandTimeoutMs
            hint = [string]$cdpDiagnostic.hint
        }
    }

    Write-ExternalPreflightDiagnosticJsonArtifact -Path $bridgePath -Value $bridgeDiagnostic

    return [pscustomobject]@{
        DiagnosticPath = $bridgePath
        Diagnostic = $bridgeDiagnostic
    }
}

function Resolve-ExternalPreflightInvocationUrl {
    param(
        [Parameter(Mandatory = $true)][System.Collections.IDictionary]$ForwardParams,
        [Parameter(Mandatory = $true)][string]$Key,
        [Parameter(Mandatory = $true)][int]$FallbackPort
    )

    if ($ForwardParams.Contains($Key) -and -not [string]::IsNullOrWhiteSpace([string]$ForwardParams[$Key])) {
        return ([string]$ForwardParams[$Key]).TrimEnd('/')
    }

    return ('http://127.0.0.1:{0}' -f $FallbackPort)
}

function Get-ExternalPreflightDiagnostic {
    param([string]$Path)

    if ([string]::IsNullOrWhiteSpace($Path) -or -not (Test-Path -LiteralPath $Path)) {
        return $null
    }

    try {
        return Get-Content -LiteralPath $Path -Raw | ConvertFrom-Json
    }
    catch {
        return $null
    }
}

function Get-ExternalPreflightDiagnosticSourceLabel {
    param(
        [string]$Path,
        [string]$StableCopyPath
    )

    $hasPath = -not [string]::IsNullOrWhiteSpace($Path)
    $hasStableCopyPath = -not [string]::IsNullOrWhiteSpace($StableCopyPath)

    if ($hasPath -and $hasStableCopyPath) {
        return 'fresh external artifact + repo-local stable copy'
    }

    if ($hasStableCopyPath) {
        return 'repo-local stable copy preview'
    }

    if ($hasPath) {
        return 'fresh external artifact'
    }

    return $null
}

function Get-ExternalPreflightCheckedAtInfo {
    param($Diagnostic)

    if ($null -eq $Diagnostic) {
        return $null
    }

    $checkedAtRaw = [string]$Diagnostic.checkedAt
    if ([string]::IsNullOrWhiteSpace($checkedAtRaw)) {
        return $null
    }

    $parsedValue = $null
    try {
        $parsedValue = [DateTimeOffset]::Parse($checkedAtRaw, [System.Globalization.CultureInfo]::InvariantCulture)
    }
    catch {
        $parsedValue = $null
    }

    return [pscustomobject]@{
        RawValue = $checkedAtRaw
        ParsedValue = $parsedValue
    }
}

function Format-ExternalPreflightDiagnosticAge {
    param([DateTimeOffset]$CheckedAt)

    $age = [DateTimeOffset]::Now - $CheckedAt
    if ($age.TotalSeconds -lt 0) {
        $age = [TimeSpan]::Zero
    }

    if ($age.TotalDays -ge 1) {
        return ('{0}d {1}h ago' -f [math]::Floor($age.TotalDays), $age.Hours)
    }

    if ($age.TotalHours -ge 1) {
        return ('{0}h {1}m ago' -f [math]::Floor($age.TotalHours), $age.Minutes)
    }

    if ($age.TotalMinutes -ge 1) {
        return ('{0}m {1}s ago' -f [math]::Floor($age.TotalMinutes), $age.Seconds)
    }

    return ('{0}s ago' -f [math]::Max([int][math]::Floor($age.TotalSeconds), 0))
}

function Get-ExternalPreflightPreviewNote {
    param(
        [string]$Path,
        [string]$StableCopyPath
    )

    if ([string]::IsNullOrWhiteSpace($Path) -and -not [string]::IsNullOrWhiteSpace($StableCopyPath)) {
        return 'check-only is replaying the best matching saved external diagnostic; rerun the external smoke to probe live services again.'
    }

    return $null
}

function Get-ExternalPreflightRequireCdpLabel {
    param($Diagnostic)

    if ($null -eq $Diagnostic) {
        return $null
    }

    if ($null -eq $Diagnostic.requireCdp) {
        return $null
    }

    if ([bool]$Diagnostic.requireCdp) {
        return 'required during this diagnostic run'
    }

    return 'not required during this diagnostic run'
}

function Get-RequestedExternalPreflightRequireCdp {
    param([hashtable]$ForwardParams)

    $bootstrapExternalCdpSession = $false
    if ($null -ne $ForwardParams -and $ForwardParams.ContainsKey('BootstrapExternalCdpSession')) {
        $bootstrapExternalCdpSession = [bool]$ForwardParams.BootstrapExternalCdpSession
    }

    $explicitRequireCdp = $false
    if ($null -ne $ForwardParams -and $ForwardParams.ContainsKey('RequireExternalCdpPreflight')) {
        $explicitRequireCdp = [bool]$ForwardParams.RequireExternalCdpPreflight
    }

    return ($explicitRequireCdp -or -not $bootstrapExternalCdpSession)
}

function Get-RequestedExternalPreflightRequireCdpLabel {
    param([object]$RequestedRequireCdp)

    if ($null -eq $RequestedRequireCdp) {
        return $null
    }

    if ([bool]$RequestedRequireCdp) {
        return 'required for this invocation'
    }

    return 'not required for this invocation'
}

function ConvertTo-PowerShellSingleQuotedLiteral {
    param([object]$Value)

    $stringValue = [string]$Value
    return "'" + $stringValue.Replace("'", "''") + "'"
}

function Get-ExternalPreflightAlternateBootstrapCommandOrder {
    param([string]$CurrentCommandOrder)

    if ([string]$CurrentCommandOrder -eq 'page-lifecycle-runtime') {
        return 'runtime-page-lifecycle'
    }

    return 'page-lifecycle-runtime'
}

function Get-ExternalPreflightRecommendedRefreshCommand {
    param(
        [hashtable]$BoundParameters,
        [hashtable]$Overrides
    )

    $commandParts = [System.Collections.Generic.List[string]]::new()
    $commandParts.Add('powershell')
    $commandParts.Add('-NoProfile')
    $commandParts.Add('-ExecutionPolicy')
    $commandParts.Add('Bypass')
    $commandParts.Add('-File')
    $commandParts.Add('.\\scripts\\verify-ai-automation-learning-browser-smoke.ps1')
    $commandParts.Add('-Mode')
    $commandParts.Add('external')

    $effectiveParameters = @{}
    if ($null -ne $BoundParameters) {
        foreach ($entry in $BoundParameters.GetEnumerator()) {
            $effectiveParameters[[string]$entry.Key] = $entry.Value
        }
    }

    if ($null -ne $Overrides) {
        foreach ($entry in $Overrides.GetEnumerator()) {
            $effectiveParameters[[string]$entry.Key] = $entry.Value
        }
    }

    foreach ($switchName in @('UseHostFriendlyExternalDefaults', 'BootstrapExternalCdpSession', 'RequireExternalCdpPreflight', 'SelfStartServices', 'SelfStartLocalServices')) {
        if ($effectiveParameters.ContainsKey($switchName) -and [bool]$effectiveParameters[$switchName]) {
            $commandParts.Add("-$switchName")
        }
    }

    foreach ($valueName in @('FrontendUrl', 'BackendUrl', 'CdpUrl', 'NodePath', 'GoPath', 'BackendBin', 'Browser', 'NodeSmokeTimeoutSeconds', 'CdpCommandTimeoutMs', 'EdgeLaunchPreset', 'EdgeProfileStrategy', 'CdpPageBootstrapStrategy', 'CdpBootstrapCommandOrder')) {
        if (-not $effectiveParameters.ContainsKey($valueName)) {
            continue
        }

        $value = $effectiveParameters[$valueName]
        if ($null -eq $value -or [string]::IsNullOrWhiteSpace([string]$value)) {
            continue
        }

        $commandParts.Add("-$valueName")
        $commandParts.Add((ConvertTo-PowerShellSingleQuotedLiteral -Value $value))
    }

    return $commandParts -join ' '
}

function Get-ExternalPreflightBootstrapComparisonCommand {
    param(
        [hashtable]$BoundParameters,
        [string]$CurrentCommandOrder
    )

    $alternateCommandOrder = Get-ExternalPreflightAlternateBootstrapCommandOrder -CurrentCommandOrder $CurrentCommandOrder
    if ([string]::IsNullOrWhiteSpace($alternateCommandOrder)) {
        return $null
    }

    return Get-ExternalPreflightRecommendedRefreshCommand -BoundParameters $BoundParameters -Overrides @{
        CdpBootstrapCommandOrder = $alternateCommandOrder
    }
}

function Get-ExternalPreflightAlternatePageBootstrapStrategy {
    param([string]$CurrentPageBootstrapStrategy)

    if ([string]$CurrentPageBootstrapStrategy -eq 'json-new') {
        return 'attached-session'
    }

    return 'json-new'
}

function Get-ExternalPreflightAlternateExecutionPathCommand {
    param(
        [hashtable]$BoundParameters,
        [string]$CurrentPageBootstrapStrategy,
        [object]$CurrentCommandTimeoutMs,
        [object]$CurrentNodeSmokeTimeoutSeconds
    )

    $alternatePageBootstrapStrategy = Get-ExternalPreflightAlternatePageBootstrapStrategy -CurrentPageBootstrapStrategy $CurrentPageBootstrapStrategy
    if ([string]::IsNullOrWhiteSpace($alternatePageBootstrapStrategy)) {
        return $null
    }

    $effectiveCommandTimeoutMs = $CurrentCommandTimeoutMs
    try {
        $effectiveCommandTimeoutMs = [int]$CurrentCommandTimeoutMs
    }
    catch {
        $effectiveCommandTimeoutMs = 0
    }

    if ($effectiveCommandTimeoutMs -le 0) {
        $effectiveCommandTimeoutMs = 60000
    }

    $effectiveNodeSmokeTimeoutSeconds = Get-ExternalPreflightSuggestedNodeSmokeTimeoutSeconds -CurrentNodeSmokeTimeoutSeconds $CurrentNodeSmokeTimeoutSeconds -CommandTimeoutMs $effectiveCommandTimeoutMs
    return Get-ExternalPreflightRecommendedRefreshCommand -BoundParameters $BoundParameters -Overrides @{
        CdpPageBootstrapStrategy = $alternatePageBootstrapStrategy
        CdpCommandTimeoutMs = $effectiveCommandTimeoutMs
        NodeSmokeTimeoutSeconds = $effectiveNodeSmokeTimeoutSeconds
    }
}

function Get-ExternalPreflightEffectiveAlternateExecutionPathCommand {
    param(
        [hashtable]$BoundParameters,
        $Diagnostic,
        [string]$CurrentPageBootstrapStrategy,
        [object]$CurrentCommandTimeoutMs,
        [object]$CurrentNodeSmokeTimeoutSeconds
    )

    $effectiveCommandTimeoutMs = $CurrentCommandTimeoutMs
    if ($null -ne $Diagnostic) {
        $pageBootstrapTimeoutState = Get-ExternalPreflightPageBootstrapTimeoutState -Diagnostic $Diagnostic
        if ($null -ne $pageBootstrapTimeoutState -and $pageBootstrapTimeoutState.CommandTimeoutMs -gt 0) {
            $effectiveCommandTimeoutMs = $pageBootstrapTimeoutState.CommandTimeoutMs
        }
    }

    return Get-ExternalPreflightAlternateExecutionPathCommand -BoundParameters $BoundParameters -CurrentPageBootstrapStrategy $CurrentPageBootstrapStrategy -CurrentCommandTimeoutMs $effectiveCommandTimeoutMs -CurrentNodeSmokeTimeoutSeconds $CurrentNodeSmokeTimeoutSeconds
}

function Get-ExternalPreflightSuggestedComparisonCommandTimeoutMs {
    param([object]$CurrentCommandTimeoutMs)

    $currentTimeoutMs = 0
    try {
        $currentTimeoutMs = [int]$CurrentCommandTimeoutMs
    }
    catch {
        $currentTimeoutMs = 0
    }

    if ($currentTimeoutMs -lt 45000) {
        return 45000
    }

    if ($currentTimeoutMs -lt 60000) {
        return 60000
    }

    if ($currentTimeoutMs -lt 90000) {
        return 90000
    }

    $nextTimeoutMs = $currentTimeoutMs + 30000
    if ($nextTimeoutMs -gt 300000) {
        $nextTimeoutMs = 300000
    }

    if ($nextTimeoutMs -le $currentTimeoutMs) {
        return $null
    }

    return $nextTimeoutMs
}

function Get-ExternalPreflightSuggestedNodeSmokeTimeoutSeconds {
    param(
        [object]$CurrentNodeSmokeTimeoutSeconds,
        [int]$CommandTimeoutMs
    )

    $currentNodeSmokeTimeoutSeconds = 0
    try {
        $currentNodeSmokeTimeoutSeconds = [int]$CurrentNodeSmokeTimeoutSeconds
    }
    catch {
        $currentNodeSmokeTimeoutSeconds = 0
    }

    $minimumNodeSmokeTimeoutSeconds = [int][Math]::Ceiling((([double]$CommandTimeoutMs * 3) + 20000) / 1000)
    if ($minimumNodeSmokeTimeoutSeconds -lt 70) {
        $minimumNodeSmokeTimeoutSeconds = 70
    }

    if ($currentNodeSmokeTimeoutSeconds -gt $minimumNodeSmokeTimeoutSeconds) {
        return $currentNodeSmokeTimeoutSeconds
    }

    return $minimumNodeSmokeTimeoutSeconds
}

function Get-ExternalPreflightTimeoutComparisonCommand {
    param(
        [hashtable]$BoundParameters,
        [object]$CurrentCommandTimeoutMs,
        [object]$CurrentNodeSmokeTimeoutSeconds
    )

    $alternateCommandTimeoutMs = Get-ExternalPreflightSuggestedComparisonCommandTimeoutMs -CurrentCommandTimeoutMs $CurrentCommandTimeoutMs
    if ($null -eq $alternateCommandTimeoutMs) {
        return $null
    }

    $alternateNodeSmokeTimeoutSeconds = Get-ExternalPreflightSuggestedNodeSmokeTimeoutSeconds -CurrentNodeSmokeTimeoutSeconds $CurrentNodeSmokeTimeoutSeconds -CommandTimeoutMs $alternateCommandTimeoutMs
    return Get-ExternalPreflightRecommendedRefreshCommand -BoundParameters $BoundParameters -Overrides @{
        CdpCommandTimeoutMs = $alternateCommandTimeoutMs
        NodeSmokeTimeoutSeconds = $alternateNodeSmokeTimeoutSeconds
    }
}

function Get-ExternalPreflightEffectiveTimeoutComparisonCommand {
    param(
        [hashtable]$BoundParameters,
        $Diagnostic,
        [object]$CurrentCommandTimeoutMs,
        [object]$CurrentNodeSmokeTimeoutSeconds
    )

    $effectiveCommandTimeoutMs = $CurrentCommandTimeoutMs
    if ($null -ne $Diagnostic) {
        $attachedSessionBootstrapTimeoutState = Get-ExternalPreflightAttachedSessionBootstrapTimeoutState -Diagnostic $Diagnostic
        if ($null -ne $attachedSessionBootstrapTimeoutState -and $attachedSessionBootstrapTimeoutState.CommandTimeoutMs -gt 0) {
            $effectiveCommandTimeoutMs = $attachedSessionBootstrapTimeoutState.CommandTimeoutMs
        }
    }

    return Get-ExternalPreflightTimeoutComparisonCommand -BoundParameters $BoundParameters -CurrentCommandTimeoutMs $effectiveCommandTimeoutMs -CurrentNodeSmokeTimeoutSeconds $CurrentNodeSmokeTimeoutSeconds
}

function Get-ExternalPreflightSelfStartCommand {
    param([string]$Command)

    if ([string]::IsNullOrWhiteSpace($Command)) {
        return $null
    }

    foreach ($switchPattern in @('(^|\s)-SelfStartServices(\s|$)', '(^|\s)-SelfStartLocalServices(\s|$)')) {
        if ($Command -match $switchPattern) {
            return $null
        }
    }

    foreach ($urlPattern in @('(^|\s)-FrontendUrl(\s|$)', '(^|\s)-BackendUrl(\s|$)')) {
        if ($Command -match $urlPattern) {
            return $null
        }
    }

    return ($Command + ' -SelfStartServices')
}

function Get-ExternalPreflightCdpRequirementMismatchNote {
    param(
        $Diagnostic,
        [object]$RequestedRequireCdp
    )

    if ($null -eq $Diagnostic -or $null -eq $RequestedRequireCdp -or $null -eq $Diagnostic.requireCdp) {
        return $null
    }

    $diagnosticRequireCdp = [bool]$Diagnostic.requireCdp
    $requestedRequireCdpBool = [bool]$RequestedRequireCdp

    if ($diagnosticRequireCdp -eq $requestedRequireCdpBool) {
        return $null
    }

    if ($requestedRequireCdpBool) {
        return 'this invocation expects an external CDP preflight, but the saved diagnostic was recorded without one; rerun the external smoke with -RequireExternalCdpPreflight to refresh the stable copy.'
    }

    return 'this invocation would keep the external CDP preflight optional, but the saved diagnostic was recorded with CDP required; rerun the external smoke without -RequireExternalCdpPreflight only if you need a service-only comparison.'
}

function Write-StableExternalPreflightDiagnosticPreview {
    param(
        [object]$RequestedRequireCdp,
        [string]$RecommendedRefreshCommand,
        [string]$RecommendedBootstrapComparisonCommand,
        [string]$RecommendedAlternateExecutionPathCommand,
        [hashtable]$BoundParameters,
        [string]$CurrentPageBootstrapStrategy,
        [object]$CurrentCommandTimeoutMs,
        [object]$CurrentNodeSmokeTimeoutSeconds,
        [int]$FreshnessThresholdHours,
        [bool]$InvocationStartsServices
    )

    Write-Host ('Stable external preflight freshness threshold (hours): {0}' -f $FreshnessThresholdHours)

    $selection = Get-StableExternalPreflightCopySelectionPaths -RequestedRequireCdp $RequestedRequireCdp -CurrentPageBootstrapStrategy $CurrentPageBootstrapStrategy
    $matchingPathExists = $false
    $matchingPathParseFailed = $false
    $matchingPathStrategyMismatch = $false
    $matchingGenericPathExists = $false

    foreach ($candidatePath in $selection.OrderedPaths) {
        if (-not (Test-Path -LiteralPath $candidatePath)) {
            continue
        }

        if (-not [string]::IsNullOrWhiteSpace($selection.MatchingPath)) {
            $candidateFullPath = [System.IO.Path]::GetFullPath($candidatePath)
            $matchingFullPath = [System.IO.Path]::GetFullPath($selection.MatchingPath)
            if ($candidateFullPath -eq $matchingFullPath) {
                $matchingPathExists = $true
            }
        }
        if (-not [string]::IsNullOrWhiteSpace($selection.GenericMatchingPath)) {
            $candidateFullPath = [System.IO.Path]::GetFullPath($candidatePath)
            $genericMatchingFullPath = [System.IO.Path]::GetFullPath($selection.GenericMatchingPath)
            if ($candidateFullPath -eq $genericMatchingFullPath) {
                $matchingGenericPathExists = $true
            }
        }

        $candidateDiagnostic = Get-ExternalPreflightDiagnostic -Path $candidatePath
        if ($null -ne $candidateDiagnostic) {
            $candidateState = Get-StableExternalPreflightCopyState -Path $candidatePath
            if (-not [string]::IsNullOrWhiteSpace($selection.MatchingPath)) {
                $candidateFullPath = [System.IO.Path]::GetFullPath($candidatePath)
                $matchingFullPath = [System.IO.Path]::GetFullPath($selection.MatchingPath)
                if ($candidateFullPath -eq $matchingFullPath) {
                    $mismatchReason = Get-StableExternalPreflightCopySelectionMismatchReason -State $candidateState -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $selection.RequestedPageBootstrapStrategy
                    $matchingPathStrategyMismatch = ($mismatchReason -eq 'page-strategy' -or $mismatchReason -eq 'page-strategy-missing')
                }
            }

            $effectiveRecommendedTimeoutComparisonCommand = Get-ExternalPreflightEffectiveTimeoutComparisonCommand -BoundParameters $BoundParameters -Diagnostic $candidateDiagnostic -CurrentCommandTimeoutMs $CurrentCommandTimeoutMs -CurrentNodeSmokeTimeoutSeconds $CurrentNodeSmokeTimeoutSeconds
            $effectiveRecommendedAlternateExecutionPathCommand = Get-ExternalPreflightEffectiveAlternateExecutionPathCommand -BoundParameters $BoundParameters -Diagnostic $candidateDiagnostic -CurrentPageBootstrapStrategy $CurrentPageBootstrapStrategy -CurrentCommandTimeoutMs $CurrentCommandTimeoutMs -CurrentNodeSmokeTimeoutSeconds $CurrentNodeSmokeTimeoutSeconds
            $selectedMatchingFallback = $false
            if (-not [string]::IsNullOrWhiteSpace($selection.GenericMatchingPath)) {
                $candidateFullPath = [System.IO.Path]::GetFullPath($candidatePath)
                $genericMatchingFullPath = [System.IO.Path]::GetFullPath($selection.GenericMatchingPath)
                $selectedMatchingFallback = ($candidateFullPath -eq $genericMatchingFullPath) -and (-not $matchingPathExists)
            }
            $selectionNote = Get-StableExternalPreflightCopySelectionNote -SelectedPath $candidatePath -MatchingPath $selection.MatchingPath -SelectedMatchingFallback:$selectedMatchingFallback -MatchingPathStrategyMismatch:$matchingPathStrategyMismatch -MatchingPathExists:$matchingPathExists -MatchingPathParseFailed:$matchingPathParseFailed -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $selection.RequestedPageBootstrapStrategy
            $legacyPathToShow = $null
            if (-not [string]::IsNullOrWhiteSpace($selection.LegacyPath)) {
                $selectedFullPath = [System.IO.Path]::GetFullPath($candidatePath)
                $legacyFullPath = [System.IO.Path]::GetFullPath($selection.LegacyPath)
                if ($selectedFullPath -ne $legacyFullPath) {
                    $legacyPathToShow = $selection.LegacyPath
                }
            }

            Write-ExternalPreflightDiagnosticSummary -Diagnostic $candidateDiagnostic -StableCopyPath $candidatePath -LegacyStableCopyPath $legacyPathToShow -StableCopySelectionNote $selectionNote -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $CurrentPageBootstrapStrategy -RecommendedRefreshCommand $RecommendedRefreshCommand -RecommendedBootstrapComparisonCommand $RecommendedBootstrapComparisonCommand -RecommendedTimeoutComparisonCommand $effectiveRecommendedTimeoutComparisonCommand -RecommendedAlternateExecutionPathCommand $effectiveRecommendedAlternateExecutionPathCommand -FreshnessThresholdHours $FreshnessThresholdHours -InvocationStartsServices:$InvocationStartsServices
            return
        }

        if (-not [string]::IsNullOrWhiteSpace($selection.MatchingPath)) {
            $candidateFullPath = [System.IO.Path]::GetFullPath($candidatePath)
            $matchingFullPath = [System.IO.Path]::GetFullPath($selection.MatchingPath)
            if ($candidateFullPath -eq $matchingFullPath) {
                $matchingPathParseFailed = $true
            }
        }
    }

    if ($selection.OrderedPaths.Count -gt 0) {
        Write-Host ('External preflight stable diagnostic copy: {0}' -f $selection.OrderedPaths[0])
    }
    if (-not [string]::IsNullOrWhiteSpace($selection.LegacyPath) -and $selection.OrderedPaths.Count -gt 0) {
        $firstFullPath = [System.IO.Path]::GetFullPath($selection.OrderedPaths[0])
        $legacyFullPath = [System.IO.Path]::GetFullPath($selection.LegacyPath)
        if ($firstFullPath -ne $legacyFullPath) {
            Write-Host ('External preflight legacy stable diagnostic copy: {0}' -f $selection.LegacyPath)
        }
    }

    $existingSelectionPaths = @($selection.OrderedPaths | Where-Object { Test-Path -LiteralPath $_ })
    if ($existingSelectionPaths.Count -gt 0) {
        foreach ($line in @(Get-StableExternalPreflightCopyStatusLines -RequestedRequireCdp $RequestedRequireCdp -Selection $selection -FreshnessThresholdHours $FreshnessThresholdHours)) {
            Write-Host $line
        }
    }

    if ($existingSelectionPaths.Count -gt 0) {
        if (-not [string]::IsNullOrWhiteSpace($RecommendedRefreshCommand)) {
            Write-Host ('External preflight preferred refresh command: {0}' -f $RecommendedRefreshCommand)
        }
        Write-Host 'Stable external preflight diagnostic copy exists but none of the available repo-local copies could be parsed; inspect the JSON files directly.'
        return
    }

    if (-not [string]::IsNullOrWhiteSpace($RecommendedRefreshCommand)) {
        Write-Host ('External preflight preferred refresh command: {0}' -f $RecommendedRefreshCommand)
    }
    Write-Host 'No stable external preflight diagnostic copy found yet; run the external smoke once to seed it.'
}

function Write-ExternalPreflightDiagnosticSummary {
    param(
        $Diagnostic,
        [string]$Path,
        [string]$StableCopyPath,
        [string]$LegacyStableCopyPath,
        [string]$StableCopySelectionNote,
        [object]$RequestedRequireCdp,
        [string]$RequestedPageBootstrapStrategy,
        [string]$RecommendedRefreshCommand,
        [string]$RecommendedBootstrapComparisonCommand,
        [string]$RecommendedTimeoutComparisonCommand,
        [string]$RecommendedAlternateExecutionPathCommand,
        [int]$FreshnessThresholdHours,
        [bool]$InvocationStartsServices
    )

    if ($null -eq $Diagnostic) {
        if ((-not [string]::IsNullOrWhiteSpace($Path)) -or (-not [string]::IsNullOrWhiteSpace($StableCopyPath))) {
            Write-Host ''
            Write-Host '== Latest external preflight diagnostic =='
            if (-not [string]::IsNullOrWhiteSpace($Path)) {
                Write-Host ('External preflight diagnostic path: {0}' -f $Path)
            }
            if (-not [string]::IsNullOrWhiteSpace($StableCopyPath)) {
                Write-Host ('External preflight stable diagnostic copy: {0}' -f $StableCopyPath)
            }
            if (-not [string]::IsNullOrWhiteSpace($LegacyStableCopyPath)) {
                Write-Host ('External preflight legacy stable diagnostic copy: {0}' -f $LegacyStableCopyPath)
            }
            if (-not [string]::IsNullOrWhiteSpace($StableCopySelectionNote)) {
                Write-Host ('External preflight stable copy note: {0}' -f $StableCopySelectionNote)
            }
        }
        return
    }

    $selfStartRecommendedRefreshCommand = Get-ExternalPreflightSelfStartCommand -Command $RecommendedRefreshCommand
    $selfStartRecommendedBootstrapComparisonCommand = Get-ExternalPreflightSelfStartCommand -Command $RecommendedBootstrapComparisonCommand
    $selfStartRecommendedTimeoutComparisonCommand = Get-ExternalPreflightSelfStartCommand -Command $RecommendedTimeoutComparisonCommand
    $selfStartRecommendedAlternateExecutionPathCommand = Get-ExternalPreflightSelfStartCommand -Command $RecommendedAlternateExecutionPathCommand

    Write-Host ''
    Write-Host '== Latest external preflight diagnostic =='
    if (-not [string]::IsNullOrWhiteSpace($Path)) {
        Write-Host ('External preflight diagnostic path: {0}' -f $Path)
    }
    if (-not [string]::IsNullOrWhiteSpace($StableCopyPath)) {
        Write-Host ('External preflight stable diagnostic copy: {0}' -f $StableCopyPath)
    }
    if (-not [string]::IsNullOrWhiteSpace($LegacyStableCopyPath)) {
        Write-Host ('External preflight legacy stable diagnostic copy: {0}' -f $LegacyStableCopyPath)
    }
    if (-not [string]::IsNullOrWhiteSpace($StableCopySelectionNote)) {
        Write-Host ('External preflight stable copy note: {0}' -f $StableCopySelectionNote)
    }

    $selectedStableCopyPath = $StableCopyPath
    if ([string]::IsNullOrWhiteSpace($selectedStableCopyPath) -and -not [string]::IsNullOrWhiteSpace($LegacyStableCopyPath)) {
        $selectedStableCopyPath = $LegacyStableCopyPath
    }
    $selection = Get-StableExternalPreflightCopySelectionPaths -RequestedRequireCdp $RequestedRequireCdp -CurrentPageBootstrapStrategy $RequestedPageBootstrapStrategy
    $statusEntries = @(Get-StableExternalPreflightCopyStatusEntries -SelectedPath $selectedStableCopyPath -Selection $selection -FreshnessThresholdHours $FreshnessThresholdHours)
    foreach ($line in @(Get-StableExternalPreflightCopyStatusLines -RequestedRequireCdp $RequestedRequireCdp -SelectedPath $selectedStableCopyPath -Selection $selection -FreshnessThresholdHours $FreshnessThresholdHours)) {
        Write-Host $line
    }
    $stableFreshnessNote = Get-StableExternalPreflightCopyFreshnessNote -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy -FreshnessThresholdHours $FreshnessThresholdHours
    if (-not [string]::IsNullOrWhiteSpace($stableFreshnessNote)) {
        Write-Host ('External preflight stable freshness note: {0}' -f $stableFreshnessNote)
    }
    $stableFreshestCopyNote = Get-StableExternalPreflightFreshestCopyNote -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy
    if (-not [string]::IsNullOrWhiteSpace($stableFreshestCopyNote)) {
        Write-Host ('External preflight stable freshest copy note: {0}' -f $stableFreshestCopyNote)
    }
    $stableCoverageNote = Get-StableExternalPreflightCoverageNote -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $RequestedPageBootstrapStrategy
    if (-not [string]::IsNullOrWhiteSpace($stableCoverageNote)) {
        Write-Host ('External preflight stable coverage note: {0}' -f $stableCoverageNote)
    }
    $stableInvocationProfileAlignmentNote = Get-StableExternalPreflightInvocationProfileAlignmentNote -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp
    if (-not [string]::IsNullOrWhiteSpace($stableInvocationProfileAlignmentNote)) {
        Write-Host ('External preflight invocation profile alignment note: {0}' -f $stableInvocationProfileAlignmentNote)
    }
    $requestedPageBootstrapStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy $RequestedPageBootstrapStrategy
    if ([string]::IsNullOrWhiteSpace($requestedPageBootstrapStrategy)) {
        $requestedPageBootstrapStrategy = Normalize-ExternalPreflightPageBootstrapStrategy -PageBootstrapStrategy ([string]$Diagnostic.cdpDiagnostic.pageStrategy) -PageMode ([string]$Diagnostic.cdpDiagnostic.pageMode)
    }
    $stableDecisionSummary = Get-StableExternalPreflightDecisionSummary -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy -InvocationStartsServices:$InvocationStartsServices
    if (-not [string]::IsNullOrWhiteSpace($stableDecisionSummary)) {
        Write-Host ('External preflight decision summary: {0}' -f $stableDecisionSummary)
    }
    $stableRecommendedActionClass = Get-StableExternalPreflightRecommendedActionClass -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy -InvocationStartsServices:$InvocationStartsServices
    if (-not [string]::IsNullOrWhiteSpace($stableRecommendedActionClass)) {
        Write-Host ('External preflight recommended action class: {0}' -f $stableRecommendedActionClass)
    }
    $stableRecommendedAction = Get-StableExternalPreflightRecommendedAction -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy -InvocationStartsServices:$InvocationStartsServices
    if (-not [string]::IsNullOrWhiteSpace($stableRecommendedAction)) {
        Write-Host ('External preflight recommended action: {0}' -f $stableRecommendedAction)
    }
    $stableHostHandoffNote = Get-StableExternalPreflightHostHandoffNote -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy -RecommendedActionClass $stableRecommendedActionClass
    if (-not [string]::IsNullOrWhiteSpace($stableHostHandoffNote)) {
        Write-Host ('External preflight host handoff note: {0}' -f $stableHostHandoffNote)
    }
    $alignedCdpFocusState = Get-StableExternalPreflightAlignedCdpBootstrapFocusState -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp
    $pageBootstrapBlockerConfirmedState = Get-StableExternalPreflightPageBootstrapBlockerConfirmedState -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy
    $attachedSessionBlockerConfirmedState = Get-StableExternalPreflightAttachedSessionBlockerConfirmedState -StatusEntries $statusEntries -RequestedRequireCdp $RequestedRequireCdp
    $matchingGenericEntry = $statusEntries | Where-Object { $_.Label -eq 'matching-generic' } | Select-Object -First 1
    $matchingGenericAlternateStrategy = Get-StableExternalPreflightEntryAlternatePageBootstrapStrategy -Entry $matchingGenericEntry -RequestedPageBootstrapStrategy $requestedPageBootstrapStrategy
    if ($null -eq $pageBootstrapBlockerConfirmedState -and $null -eq $attachedSessionBlockerConfirmedState -and $null -ne $alignedCdpFocusState -and -not [string]::IsNullOrWhiteSpace($RecommendedBootstrapComparisonCommand)) {
        Write-Host ('External preflight CDP bootstrap comparison command: {0}' -f $RecommendedBootstrapComparisonCommand)
        if (-not [string]::IsNullOrWhiteSpace($selfStartRecommendedBootstrapComparisonCommand)) {
            Write-Host ('External preflight self-start CDP bootstrap comparison command: {0}' -f $selfStartRecommendedBootstrapComparisonCommand)
        }
    }
    if ($null -eq $pageBootstrapBlockerConfirmedState -and $null -eq $attachedSessionBlockerConfirmedState -and $null -ne $alignedCdpFocusState -and -not [string]::IsNullOrWhiteSpace($RecommendedTimeoutComparisonCommand)) {
        Write-Host ('External preflight CDP timeout comparison command: {0}' -f $RecommendedTimeoutComparisonCommand)
        if (-not [string]::IsNullOrWhiteSpace($selfStartRecommendedTimeoutComparisonCommand)) {
            Write-Host ('External preflight self-start CDP timeout comparison command: {0}' -f $selfStartRecommendedTimeoutComparisonCommand)
        }
    }
    if (($null -ne $pageBootstrapBlockerConfirmedState -or $null -ne $attachedSessionBlockerConfirmedState) -and -not [string]::IsNullOrWhiteSpace($RecommendedAlternateExecutionPathCommand)) {
        Write-Host ('External preflight alternate execution path command: {0}' -f $RecommendedAlternateExecutionPathCommand)
        if (-not [string]::IsNullOrWhiteSpace($selfStartRecommendedAlternateExecutionPathCommand)) {
            Write-Host ('External preflight self-start alternate execution path command: {0}' -f $selfStartRecommendedAlternateExecutionPathCommand)
        }
    }
    elseif (-not [string]::IsNullOrWhiteSpace($matchingGenericAlternateStrategy) -and -not [string]::IsNullOrWhiteSpace($RecommendedAlternateExecutionPathCommand)) {
        Write-Host ('External preflight alternate execution path command: {0}' -f $RecommendedAlternateExecutionPathCommand)
        if (-not [string]::IsNullOrWhiteSpace($selfStartRecommendedAlternateExecutionPathCommand)) {
            Write-Host ('External preflight self-start alternate execution path command: {0}' -f $selfStartRecommendedAlternateExecutionPathCommand)
        }
    }

    $sourceLabel = Get-ExternalPreflightDiagnosticSourceLabel -Path $Path -StableCopyPath $StableCopyPath
    if (-not [string]::IsNullOrWhiteSpace($sourceLabel)) {
        Write-Host ('External preflight diagnostic source: {0}' -f $sourceLabel)
    }

    $checkedAtInfo = Get-ExternalPreflightCheckedAtInfo -Diagnostic $Diagnostic
    if ($null -ne $checkedAtInfo) {
        Write-Host ('External preflight checked at: {0}' -f $checkedAtInfo.RawValue)
        if ($null -ne $checkedAtInfo.ParsedValue) {
            Write-Host ('External preflight diagnostic age: {0}' -f (Format-ExternalPreflightDiagnosticAge -CheckedAt $checkedAtInfo.ParsedValue))
        }
    }

    $previewNote = Get-ExternalPreflightPreviewNote -Path $Path -StableCopyPath $StableCopyPath
    if (-not [string]::IsNullOrWhiteSpace($previewNote)) {
        Write-Host ('External preflight preview note: {0}' -f $previewNote)
    }

    $requireCdpLabel = Get-ExternalPreflightRequireCdpLabel -Diagnostic $Diagnostic
    if (-not [string]::IsNullOrWhiteSpace($requireCdpLabel)) {
        Write-Host ('External preflight CDP requirement: {0}' -f $requireCdpLabel)
    }

    $requestedRequireCdpLabel = Get-RequestedExternalPreflightRequireCdpLabel -RequestedRequireCdp $RequestedRequireCdp
    if (-not [string]::IsNullOrWhiteSpace($requestedRequireCdpLabel)) {
        Write-Host ('External preflight CDP expectation for this invocation: {0}' -f $requestedRequireCdpLabel)
    }

    $cdpRequirementMismatchNote = Get-ExternalPreflightCdpRequirementMismatchNote -Diagnostic $Diagnostic -RequestedRequireCdp $RequestedRequireCdp
    if (-not [string]::IsNullOrWhiteSpace($cdpRequirementMismatchNote)) {
        Write-Host ('External preflight CDP mismatch note: {0}' -f $cdpRequirementMismatchNote)
    }

    if (-not [string]::IsNullOrWhiteSpace($RecommendedRefreshCommand)) {
        Write-Host ('External preflight preferred refresh command: {0}' -f $RecommendedRefreshCommand)
        if (-not [string]::IsNullOrWhiteSpace($selfStartRecommendedRefreshCommand)) {
            Write-Host ('External preflight self-start refresh command: {0}' -f $selfStartRecommendedRefreshCommand)
        }
    }

    Write-Host ('External preflight overall classification: {0}' -f $Diagnostic.overallClassification)

    $failedChecks = @($Diagnostic.failedChecks)
    Write-Host ('External preflight failed checks: {0}' -f $(if ($failedChecks.Count -gt 0) { $failedChecks -join ', ' } else { 'none' }))

    $skippedChecks = @($Diagnostic.skippedChecks)
    Write-Host ('External preflight skipped checks: {0}' -f $(if ($skippedChecks.Count -gt 0) { $skippedChecks -join ', ' } else { 'none' }))

    if (-not [string]::IsNullOrWhiteSpace([string]$Diagnostic.primaryBlockingCheck)) {
        Write-Host ('External preflight primary blocking check: {0}' -f $Diagnostic.primaryBlockingCheck)
    }

    $summaryLines = @($Diagnostic.summaryLines)
    if ($summaryLines.Count -gt 0) {
        Write-Host 'External preflight summary lines:'
        foreach ($line in $summaryLines) {
            Write-Host $line
        }
    }

    $hints = @($Diagnostic.hints)
    if ($hints.Count -gt 0) {
        Write-Host 'External preflight hints:'
        foreach ($hint in $hints) {
            Write-Host ('- ' + $hint)
        }
    }

    $nextStepLines = @(Get-ExternalPreflightNextStepLines -Diagnostic $Diagnostic -RequestedRequireCdp $RequestedRequireCdp -RecommendedRefreshCommand $RecommendedRefreshCommand -InvocationStartsServices:$InvocationStartsServices)
    if ($nextStepLines.Count -gt 0) {
        Write-Host 'External preflight next steps:'
        foreach ($line in $nextStepLines) {
            Write-Host ('- ' + $line)
        }
    }
}

function Get-ExternalPreflightNextStepLines {
    param(
        $Diagnostic,
        [object]$RequestedRequireCdp,
        [string]$RecommendedRefreshCommand,
        [bool]$InvocationStartsServices
    )

    if ($null -eq $Diagnostic) {
        return @()
    }

    $failedChecks = @($Diagnostic.failedChecks)
    $nextSteps = [System.Collections.Generic.List[string]]::new()
    $hasRecommendedRefreshCommand = -not [string]::IsNullOrWhiteSpace($RecommendedRefreshCommand)

    if ($failedChecks.Count -eq 0) {
        if ($hasRecommendedRefreshCommand) {
            $nextSteps.Add('If you want to refresh the saved external diagnostic for this same invocation profile, rerun: ' + $RecommendedRefreshCommand)
        }
        $nextSteps.Add('No failed checks were recorded; inspect the full diagnostic artifact if the wrapper still rethrows.')
        return $nextSteps.ToArray()
    }

    if (($failedChecks -contains 'backend') -or ($failedChecks -contains 'frontend')) {
        if ($hasRecommendedRefreshCommand) {
            $nextSteps.Add('After verifying the target URLs, rerun the same external profile for this invocation: ' + $RecommendedRefreshCommand)
        }
        else {
            $nextSteps.Add('If backend/frontend should already be running, verify their URLs first and then rerun: powershell -NoProfile -ExecutionPolicy Bypass -File .\\scripts\\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults')
        }

        if (-not $InvocationStartsServices) {
            $nextSteps.Add('If this host should start local services for comparison, rerun: powershell -NoProfile -ExecutionPolicy Bypass -File .\\scripts\\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -SelfStartServices')
        }

        if ($null -eq $RequestedRequireCdp -or -not [bool]$RequestedRequireCdp) {
            $nextSteps.Add('If you also want the first external preflight to verify Edge CDP reachability instead of leaving it skipped, rerun: powershell -NoProfile -ExecutionPolicy Bypass -File .\\scripts\\verify-ai-automation-learning-browser-smoke.ps1 -Mode external -UseHostFriendlyExternalDefaults -RequireExternalCdpPreflight')
        }
    }

    if ($failedChecks -contains 'cdp') {
        $nextSteps.Add('If backend/frontend are already reachable, keep the same external command and focus on the CDP endpoint or Edge bootstrap instead of reopening service reachability work.')
    }

    if ($nextSteps.Count -eq 0) {
        if ($hasRecommendedRefreshCommand) {
            $nextSteps.Add('After addressing the failed checks above, rerun the same external profile for this invocation: ' + $RecommendedRefreshCommand)
        }
        else {
        $nextSteps.Add('Rerun the same external command after addressing the failed checks above, then inspect the diagnostic artifact again if it still fails.')
        }
    }

    return $nextSteps.ToArray()
}

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
    NodeSmokeSuccessMarker = 'ai-automation-learning-browser-smoke-cdp passed'
    SmokeLabel = 'ai automation learning'
}

$env:OCTOPUS_UI_SMOKE_SCENARIO = 'ai-learning'

if ($UseHostFriendlyExternalDefaults) {
    if ($Mode -eq 'self-start') {
        throw 'UseHostFriendlyExternalDefaults only supports external or check-only mode for ai learning smoke.'
    }

    if (-not $PSBoundParameters.ContainsKey('CdpUrl')) { $forwardParams.CdpUrl = 'http://127.0.0.1:9233' }
    if (-not $PSBoundParameters.ContainsKey('NodeSmokeTimeoutSeconds')) { $forwardParams.NodeSmokeTimeoutSeconds = 70 }
    if (-not $PSBoundParameters.ContainsKey('CdpCommandTimeoutMs')) { $forwardParams.CdpCommandTimeoutMs = 30000 }
    if (-not $PSBoundParameters.ContainsKey('EdgeLaunchPreset')) { $forwardParams.EdgeLaunchPreset = 'relaxed' }
    if (-not $PSBoundParameters.ContainsKey('EdgeProfileStrategy')) { $forwardParams.EdgeProfileStrategy = 'workspace-fixed' }
    $forwardParams.BootstrapExternalCdpSession = $true
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
if ($SelfStartServices -or $SelfStartLocalServices) { $forwardParams.SelfStartServices = $true }
if ($KeepArtifacts) { $forwardParams.KeepArtifacts = $true }
if ($KeepProcessesOnFailure) { $forwardParams.KeepProcessesOnFailure = $true }

$requestedRequireCdp = Get-RequestedExternalPreflightRequireCdp -ForwardParams $forwardParams
$recommendedRefreshCommand = Get-ExternalPreflightRecommendedRefreshCommand -BoundParameters $PSBoundParameters
$recommendedBootstrapComparisonCommand = Get-ExternalPreflightBootstrapComparisonCommand -BoundParameters $PSBoundParameters -CurrentCommandOrder ([string]$forwardParams.CdpBootstrapCommandOrder)
$recommendedTimeoutComparisonCommand = Get-ExternalPreflightTimeoutComparisonCommand -BoundParameters $PSBoundParameters -CurrentCommandTimeoutMs $forwardParams.CdpCommandTimeoutMs -CurrentNodeSmokeTimeoutSeconds $forwardParams.NodeSmokeTimeoutSeconds
$recommendedAlternateExecutionPathCommand = Get-ExternalPreflightAlternateExecutionPathCommand -BoundParameters $PSBoundParameters -CurrentPageBootstrapStrategy ([string]$forwardParams.CdpPageBootstrapStrategy) -CurrentCommandTimeoutMs $forwardParams.CdpCommandTimeoutMs -CurrentNodeSmokeTimeoutSeconds $forwardParams.NodeSmokeTimeoutSeconds
$invocationStartsServices = $forwardParams.ContainsKey('SelfStartServices') -and [bool]$forwardParams.SelfStartServices

if ($Mode -eq 'check-only') {
    Write-StableExternalPreflightDiagnosticPreview -RequestedRequireCdp $requestedRequireCdp -RecommendedRefreshCommand $recommendedRefreshCommand -RecommendedBootstrapComparisonCommand $recommendedBootstrapComparisonCommand -RecommendedAlternateExecutionPathCommand $recommendedAlternateExecutionPathCommand -BoundParameters $PSBoundParameters -CurrentPageBootstrapStrategy ([string]$forwardParams.CdpPageBootstrapStrategy) -CurrentCommandTimeoutMs $forwardParams.CdpCommandTimeoutMs -CurrentNodeSmokeTimeoutSeconds $forwardParams.NodeSmokeTimeoutSeconds -FreshnessThresholdHours $StableDiagnosticFreshnessThresholdHours -InvocationStartsServices:$invocationStartsServices
    exit 0
}

try {
    & $sharedWrapper @forwardParams
    if (Test-Path Variable:\LASTEXITCODE) {
        exit $LASTEXITCODE
    }
    exit 0
}
catch {
    $diagnosticPath = Get-ExternalPreflightDiagnosticPathFromMessage -Message $_.Exception.Message
    if ([string]::IsNullOrWhiteSpace($diagnosticPath)) {
        $cdpDiagnosticPath = Get-ExternalPreflightCdpDiagnosticPathFromMessage -Message $_.Exception.Message
        if (-not [string]::IsNullOrWhiteSpace($cdpDiagnosticPath)) {
            $bridgedDiagnostic = New-ExternalPreflightDiagnosticFromCdpDiagnostic `
                -CdpDiagnosticPath $cdpDiagnosticPath `
                -Message $_.Exception.Message `
                -RequestedRequireCdp $requestedRequireCdp `
                -FrontendUrl (Resolve-ExternalPreflightInvocationUrl -ForwardParams $forwardParams -Key 'FrontendUrl' -FallbackPort $FrontendPort) `
                -BackendUrl (Resolve-ExternalPreflightInvocationUrl -ForwardParams $forwardParams -Key 'BackendUrl' -FallbackPort $BackendPort) `
                -CdpUrl (Resolve-ExternalPreflightInvocationUrl -ForwardParams $forwardParams -Key 'CdpUrl' -FallbackPort $CdpPort)
            if ($null -ne $bridgedDiagnostic -and -not [string]::IsNullOrWhiteSpace([string]$bridgedDiagnostic.DiagnosticPath)) {
                $diagnosticPath = [string]$bridgedDiagnostic.DiagnosticPath
            }
        }
    }
    if (-not [string]::IsNullOrWhiteSpace($diagnosticPath)) {
        $publishedDiagnosticCopies = Publish-ExternalPreflightDiagnosticCopy -SourcePath $diagnosticPath
        $diagnostic = Get-ExternalPreflightDiagnostic -Path $diagnosticPath
        $stableDiagnosticPath = $null
        $legacyStableDiagnosticPath = $null
        if ($null -ne $publishedDiagnosticCopies) {
            $stableDiagnosticPath = $publishedDiagnosticCopies.StableCopyPath
            $legacyStableDiagnosticPath = $publishedDiagnosticCopies.LegacyPath
        }
        Write-ExternalPreflightDiagnosticSummary -Diagnostic $diagnostic -Path $diagnosticPath -StableCopyPath $stableDiagnosticPath -LegacyStableCopyPath $legacyStableDiagnosticPath -RequestedRequireCdp $requestedRequireCdp -RequestedPageBootstrapStrategy ([string]$forwardParams.CdpPageBootstrapStrategy) -RecommendedRefreshCommand $recommendedRefreshCommand -RecommendedBootstrapComparisonCommand $recommendedBootstrapComparisonCommand -RecommendedTimeoutComparisonCommand $recommendedTimeoutComparisonCommand -RecommendedAlternateExecutionPathCommand $recommendedAlternateExecutionPathCommand -FreshnessThresholdHours $StableDiagnosticFreshnessThresholdHours -InvocationStartsServices:$invocationStartsServices
    }

    throw
}
