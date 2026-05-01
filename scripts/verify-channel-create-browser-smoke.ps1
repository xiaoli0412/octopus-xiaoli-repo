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

$nodeSmokeScript = if ([string]::IsNullOrWhiteSpace($env:OCTOPUS_UI_SMOKE_SCRIPT)) { 'scripts/verify-channel-create-browser-smoke.mjs' } else { $env:OCTOPUS_UI_SMOKE_SCRIPT }
$nodeSmokeSuccessMarker = if ([string]::IsNullOrWhiteSpace($env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER)) { 'channel-create-browser-smoke passed' } else { $env:OCTOPUS_UI_SMOKE_SUCCESS_MARKER }
$smokeLabel = if ([string]::IsNullOrWhiteSpace($env:OCTOPUS_UI_SMOKE_LABEL)) { 'channel create' } else { $env:OCTOPUS_UI_SMOKE_LABEL }

if ($Driver -eq 'cdp') {
    $scriptDir = Split-Path -Parent $PSCommandPath
    $cdpWrapper = Join-Path $scriptDir 'verify-channel-create-browser-smoke-cdp.ps1'

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

    if ($PSBoundParameters.ContainsKey('FrontendUrl')) {
        $forwardParams.FrontendUrl = $FrontendUrl
    }
    if ($PSBoundParameters.ContainsKey('BackendUrl')) {
        $forwardParams.BackendUrl = $BackendUrl
    }
    if ($PSBoundParameters.ContainsKey('CdpUrl')) {
        $forwardParams.CdpUrl = $CdpUrl
    }
    if ($PSBoundParameters.ContainsKey('NodePath')) {
        $forwardParams.NodePath = $NodePath
    }
    if ($PSBoundParameters.ContainsKey('GoPath')) {
        $forwardParams.GoPath = $GoPath
    }
    if ($PSBoundParameters.ContainsKey('BackendBin')) {
        $forwardParams.BackendBin = $BackendBin
    }
    if ($PSBoundParameters.ContainsKey('Browser')) {
        $forwardParams.BrowserPath = $Browser
    }
    if ($BootstrapExternalCdpSession) {
        $forwardParams.BootstrapExternalCdpSession = $true
    }
    if ($RequireExternalCdpPreflight) {
        $forwardParams.RequireExternalCdpPreflight = $true
    }
    if ($SelfStartServices) {
        $forwardParams.SelfStartServices = $true
    }
    if ($KeepArtifacts) {
        $forwardParams.KeepArtifacts = $true
    }
    if ($KeepProcessesOnFailure) {
        $forwardParams.KeepProcessesOnFailure = $true
    }

    & $cdpWrapper @forwardParams
    if (Test-Path Variable:\LASTEXITCODE) {
        exit $LASTEXITCODE
    }
    exit 0
}

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ''
    Write-Host ('== ' + $Message + ' ==')
}

function Ensure-CoreWindowsEnvironment {
    if ([string]::IsNullOrWhiteSpace($env:SystemRoot)) {
        $env:SystemRoot = 'C:\Windows'
    }

    if ([string]::IsNullOrWhiteSpace($env:windir)) {
        $env:windir = $env:SystemRoot
    }

    if ([string]::IsNullOrWhiteSpace($env:COMSPEC)) {
        $env:COMSPEC = Join-Path $env:SystemRoot 'System32\cmd.exe'
    }

    if ([string]::IsNullOrWhiteSpace($env:APPDATA) -and -not [string]::IsNullOrWhiteSpace($env:USERPROFILE)) {
        $env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'
    }

    if ([string]::IsNullOrWhiteSpace($env:LOCALAPPDATA) -and -not [string]::IsNullOrWhiteSpace($env:USERPROFILE)) {
        $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'
    }
}

function Get-RepoRoot {
    $scriptDir = Split-Path -Parent $PSCommandPath
    return [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))
}

function Normalize-Url {
    param([Parameter(Mandatory = $true)][string]$Value)

    return $Value.TrimEnd('/')
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

function Test-IsLocalUrl {
    param([Parameter(Mandatory = $true)][string]$Url)

    try {
        $uri = [System.Uri]$Url
    }
    catch {
        return $false
    }

    return @('127.0.0.1', 'localhost', '::1').Contains($uri.Host)
}

function Test-IsServiceProviderFailureMessage {
    param([string]$Message)

    if ([string]::IsNullOrWhiteSpace($Message)) {
        return $false
    }

    return $Message -match '无法加载或初始化请求的服务提供程序|requested service provider|service provider could not be loaded or initialized'
}

function Test-LoopbackBindCapability {
    param([string]$BindHost = '127.0.0.1')

    $listener = $null
    try {
        $listener = New-Object System.Net.Sockets.TcpListener ([System.Net.IPAddress]::Parse($BindHost)), 0
        $listener.Start()
        return [pscustomobject]@{
            CanBind = $true
            FailureMessage = ''
        }
    }
    catch {
        return [pscustomobject]@{
            CanBind = $false
            FailureMessage = Get-ErrorMessageChain -ErrorRecord $_
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

function Test-LoopbackClientCapability {
    param(
        [string]$ConnectHost = '127.0.0.1',
        [int]$Port = 1
    )

    $client = $null
    try {
        $client = New-Object System.Net.Sockets.TcpClient
        $client.Connect($ConnectHost, $Port)
        return [pscustomobject]@{
            Ready = $true
            Detail = ('connected to {0}:{1}' -f $ConnectHost, $Port)
            FailureMessage = ''
        }
    }
    catch {
        $message = Get-ErrorMessageChain -ErrorRecord $_
        if (Test-IsServiceProviderFailureMessage -Message $message) {
            return [pscustomobject]@{
                Ready = $false
                Detail = $message
                FailureMessage = $message
            }
        }

        return [pscustomobject]@{
            Ready = $true
            Detail = ('socket stack responded while probing {0}:{1}: {2}' -f $ConnectHost, $Port, $message)
            FailureMessage = $message
        }
    }
    finally {
        if ($null -ne $client) {
            try {
                $client.Close()
            }
            catch {
            }
        }
    }
}

function Get-LoopbackCapabilitySummary {
    param([string]$LoopbackHost = '127.0.0.1')

    $bindProbe = Test-LoopbackBindCapability -BindHost $LoopbackHost
    $clientProbe = Test-LoopbackClientCapability -ConnectHost $LoopbackHost -Port 1
    $serviceProviderBlocked = ((-not $bindProbe.CanBind) -and (Test-IsServiceProviderFailureMessage -Message $bindProbe.FailureMessage)) -or ((-not $clientProbe.Ready) -and (Test-IsServiceProviderFailureMessage -Message $clientProbe.FailureMessage))

    return [pscustomobject]@{
        Host = $LoopbackHost
        BindProbe = $bindProbe
        ClientProbe = $clientProbe
        ServiceProviderBlocked = $serviceProviderBlocked
    }
}

function Write-LoopbackCapabilityReport {
    param([Parameter(Mandatory = $true)]$Capability)

    if ($Capability.ServiceProviderBlocked) {
        Write-Host ('Loopback localhost readiness: blocked on {0} by Windows service-provider initialization; local self-start/external smoke will fail until the host networking stack is repaired.' -f $Capability.Host)
    }
    else {
        Write-Host ('Loopback localhost readiness: usable on {0} for local smoke preflight.' -f $Capability.Host)
    }

    $bindSummary = if ($Capability.BindProbe.CanBind) { 'ready (bind ok)' } else { 'blocked ({0})' -f $Capability.BindProbe.FailureMessage }
    $clientSummary = if ($Capability.ClientProbe.Ready) { $Capability.ClientProbe.Detail } else { 'blocked ({0})' -f $Capability.ClientProbe.Detail }

    Write-Host ('Loopback bind probe: {0}' -f $bindSummary)
    Write-Host ('Loopback client probe: {0}' -f $clientSummary)
}

function Assert-LoopbackSelfStartReady {
    param([string]$BindHost = '127.0.0.1')

    $probe = Test-LoopbackBindCapability -BindHost $BindHost
    if ($probe.CanBind) {
        return
    }

    if (Test-IsServiceProviderFailureMessage -Message $probe.FailureMessage) {
        throw ("Local self-start cannot bind a loopback TCP listener on {0}. This host currently cannot initialize the Windows service provider for localhost sockets, so treat this as a host networking blocker instead of an Octopus regression.`nLast error: {1}" -f $BindHost, $probe.FailureMessage)
    }
}

function Assert-LoopbackLocalHttpReady {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$ScenarioLabel
    )

    if (-not (Test-IsLocalUrl -Url $Url)) {
        return
    }

    $uri = [System.Uri]$Url
    $probeHost = switch ($uri.Host.ToLowerInvariant()) {
        'localhost' { '127.0.0.1' }
        default { $uri.Host }
    }
    $capability = Get-LoopbackCapabilitySummary -LoopbackHost $probeHost
    if (-not $capability.ServiceProviderBlocked) {
        return
    }

    $bindDetail = if ($capability.BindProbe.CanBind) { 'ready (bind ok)' } else { $capability.BindProbe.FailureMessage }
    $clientDetail = if ($capability.ClientProbe.Ready) { $capability.ClientProbe.Detail } else { $capability.ClientProbe.Detail }

    throw ("Local {0} preflight is blocked before HTTP probing because loopback localhost on {1} cannot initialize the Windows service provider. Treat this as a host networking blocker instead of a page regression.`nBind probe: {2}`nClient probe: {3}`nTarget URL: {4}" -f $ScenarioLabel, $probeHost, $bindDetail, $clientDetail, $Url)
}

function Resolve-BackendSelfStartFailure {
    param(
        [Parameter(Mandatory = $true)][string]$StdoutText,
        [Parameter(Mandatory = $true)][string]$StderrText,
        [Parameter(Mandatory = $true)][string]$BackendUrl,
        [Parameter(Mandatory = $true)][string]$StdoutPath,
        [Parameter(Mandatory = $true)][string]$StderrPath
    )

    $combined = (@($StdoutText, $StderrText) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }) -join [Environment]::NewLine
    if (Test-IsServiceProviderFailureMessage -Message $combined) {
        return "Local backend self-start could not bind $BackendUrl because Windows service-provider initialization failed. Treat this as a host networking blocker instead of a page regression.`nLog files: stdout=$StdoutPath`nstderr=$StderrPath`nRelevant output:`n$combined"
    }

    return "Backend self-start exited early.`nLog files: stdout=$StdoutPath`nstderr=$StderrPath`nRelevant output:`n$combined"
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

function Wait-Http {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [int]$TimeoutSeconds = 60
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    $lastFailure = ''
    $isLocalUrl = Test-IsLocalUrl -Url $Url
    while ((Get-Date) -lt $deadline) {
        try {
            $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 500) {
                return
            }

            $lastFailure = ('HTTP {0}' -f $response.StatusCode)
        }
        catch {
            $lastFailure = Get-ErrorMessageChain -ErrorRecord $_

            if ($isLocalUrl -and (Test-IsServiceProviderFailureMessage -Message $lastFailure)) {
                throw ("Local HTTP preflight for {0} hit a Windows service-provider initialization failure. This host currently cannot complete localhost HTTP checks reliably, so treat this as a host networking blocker instead of a page regression.`nLast error: {1}" -f $Url, $lastFailure)
            }
        }

        Start-Sleep -Milliseconds 500
    }

    if (-not [string]::IsNullOrWhiteSpace($lastFailure)) {
        throw ("Timed out waiting for {0}. Last failure: {1}" -f $Url, $lastFailure)
    }

    throw "Timed out waiting for $Url"
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

function Test-TcpPortAvailable {
    param(
        [Parameter(Mandatory = $true)][int]$Port,
        [string]$BindHost = '127.0.0.1'
    )

    if ((Get-ListeningTcpPorts) -contains $Port) {
        return $false
    }

    $listener = $null
    try {
        $listener = New-Object System.Net.Sockets.TcpListener ([System.Net.IPAddress]::Parse($BindHost)), $Port
        $listener.Start()
        return $true
    }
    catch [System.Net.Sockets.SocketException] {
        if ($_.Exception.Message -match '无法加载或初始化请求的服务提供程序|requested service provider') {
            return $true
        }
        return $false
    }
    catch {
        if ($_.Exception.Message -match '无法加载或初始化请求的服务提供程序|requested service provider') {
            return $true
        }
        return $false
    }
    finally {
        if ($null -ne $listener) {
            $listener.Stop()
        }
    }
}

function Resolve-FreeTcpPort {
    param(
        [Parameter(Mandatory = $true)][int]$PreferredPort,
        [string]$BindHost = '127.0.0.1',
        [int]$SearchWindow = 200
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

function Write-ConfigJson {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$StaticDir,
        [Parameter(Mandatory = $true)][string]$DbPath,
        [Parameter(Mandatory = $true)][int]$Port
    )

    $configObject = [ordered]@{
        server = [ordered]@{
            host = '127.0.0.1'
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
            Process = $process
            StdoutPath = $StdoutPath
            StderrPath = $StderrPath
            StdoutTail = $stdoutTail
            StderrTail = $stderrTail
        }
    }
    finally {
        if ($process -and -not $process.HasExited) {
            Stop-ProcessTree -Process $process
        }
    }
}

function Assert-NodeSmokeSucceeded {
    param(
        [Parameter(Mandatory = $true)][pscustomobject]$Result,
        [Parameter(Mandatory = $true)][string]$ExpectedResult,
        [string]$SmokeLabel = 'browser smoke'
    )

    $stdout = Read-LogContent -Path $Result.StdoutPath
    $stderr = Read-LogContent -Path $Result.StderrPath

    $labelText = if ([string]::IsNullOrWhiteSpace($SmokeLabel)) { 'browser smoke' } else { $SmokeLabel }
    $combined = (@($stdout, $stderr) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }) -join [Environment]::NewLine

    if ($combined -match '(?m)^Error:\s+spawn EPERM\b|spawnSync .* EPERM' -and $combined -match 'playwright-cli start') {
        throw ("Node {0} is blocked by a host-level child-process 'spawn EPERM' failure while launching Playwright CLI. Treat this as a host environment blocker instead of a page regression.`nLog files: stdout={1}`nstderr={2}`nSTDOUT tail:`n{3}`nSTDERR tail:`n{4}" -f $labelText, $Result.StdoutPath, $Result.StderrPath, $Result.StdoutTail, $Result.StderrTail)
    }

    if ($stdout -notmatch [regex]::Escape($ExpectedResult)) {
        throw ("Node {0} did not emit expected success marker '{1}'.`nLog files: stdout={2}`nstderr={3}`nSTDOUT tail:`n{4}`nSTDERR tail:`n{5}" -f $labelText, $ExpectedResult, $Result.StdoutPath, $Result.StderrPath, $Result.StdoutTail, $Result.StderrTail)
    }

    if ($stderr -match '(?m)^([A-Za-z][A-Za-z0-9]*(Error|Exception)|AssertionError|Error|SyntaxError|TypeError|ReferenceError|RangeError):|\bat async main\b|\bprocess\.exit\b') {
        throw ("Node {0} wrote an error-like stderr despite the success marker.`nLog files: stdout={1}`nstderr={2}`nSTDOUT tail:`n{3}`nSTDERR tail:`n{4}" -f $labelText, $Result.StdoutPath, $Result.StderrPath, $Result.StdoutTail, $Result.StderrTail)
    }
}

$repoRoot = Get-RepoRoot
Ensure-CoreWindowsEnvironment
$smokeLabelSlug = (($smokeLabel.ToLowerInvariant() -replace '[^a-z0-9]+', '-').Trim('-'))
if ([string]::IsNullOrWhiteSpace($smokeLabelSlug)) {
    $smokeLabelSlug = 'browser-smoke'
}
$tempRoot = Join-Path $env:TEMP ('octopus-' + $smokeLabelSlug + '-smoke-' + [guid]::NewGuid().ToString('N'))
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
    throw ("Unable to resolve Node.js executable for {0} smoke." -f $smokeLabel)
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

$npxCliScript = Join-Path (Split-Path -Parent $resolvedNodePath) 'node_modules\npm\bin\npx-cli.js'
$effectiveBackendPort = $BackendPort
$effectiveFrontendPort = $FrontendPort
$bootstrapLocalServices = $Mode -eq 'self-start'

if ($bootstrapLocalServices) {
    $effectiveBackendPort = Resolve-FreeTcpPort -PreferredPort $BackendPort
    $effectiveFrontendPort = Resolve-FreeTcpPort -PreferredPort $FrontendPort
}

$hasExplicitFrontendUrl = $PSBoundParameters.ContainsKey('FrontendUrl') -and -not [string]::IsNullOrWhiteSpace($FrontendUrl)
$requestedFrontendBaseUrl = Normalize-Url -Value $(if ($hasExplicitFrontendUrl) { $FrontendUrl } else { "http://127.0.0.1:$FrontendPort" })
$backendBaseUrl = Normalize-Url -Value $(if ($BackendUrl) { $BackendUrl } else { "http://127.0.0.1:$effectiveBackendPort" })
$frontendBaseUrl = if ($bootstrapLocalServices) {
    $backendBaseUrl
}
elseif ($hasExplicitFrontendUrl) {
    $requestedFrontendBaseUrl
}
else {
    $backendBaseUrl
}

$env:OCTOPUS_UI_SMOKE_FRONTEND_URL = $frontendBaseUrl
$env:OCTOPUS_UI_SMOKE_BACKEND_URL = $backendBaseUrl
$env:OCTOPUS_UI_SMOKE_NODE = $resolvedNodePath
if ($resolvedBackendBin) {
    $env:OCTOPUS_UI_SMOKE_BACKEND_BIN = $resolvedBackendBin
}
if (Test-Path -LiteralPath $npxCliScript) {
    $env:OCTOPUS_UI_SMOKE_NPX_SCRIPT = $npxCliScript
}
if (-not [string]::IsNullOrWhiteSpace($Browser)) {
    $env:OCTOPUS_UI_SMOKE_BROWSER = $Browser
}

if ($Mode -eq 'check-only') {
    $loopbackCapability = Get-LoopbackCapabilitySummary
    Write-Step 'Check-only summary'
    Write-Host ("Node: $resolvedNodePath")
    Write-Host ("Go: $resolvedGoPath")
    Write-Host ("Backend binary: $resolvedBackendBin")
    Write-Host ("Frontend URL: $frontendBaseUrl")
    Write-Host ("Backend URL: $backendBaseUrl")
    if (-not $hasExplicitFrontendUrl -and $frontendBaseUrl -eq $backendBaseUrl -and $requestedFrontendBaseUrl -ne $backendBaseUrl) {
        Write-Host 'Frontend URL source: defaulted to backend URL for backend-served static smoke (pass -FrontendUrl to target a separate frontend dev server).'
    }
    Write-LoopbackCapabilityReport -Capability $loopbackCapability
    Write-Host ("Node smoke timeout (seconds): $NodeSmokeTimeoutSeconds")
    & $resolvedNodePath $nodeSmokeScript '--check-only'
    exit $LASTEXITCODE
}

$processes = @()
$verificationSucceeded = $false

try {
    if ($bootstrapLocalServices) {
        Write-Step ("Starting backend and frontend for {0} smoke" -f $smokeLabel)

        Assert-LoopbackSelfStartReady -BindHost '127.0.0.1'

        if ($effectiveBackendPort -ne $BackendPort) {
            Write-Host ("Backend port {0} is busy; falling back to {1}." -f $BackendPort, $effectiveBackendPort)
        }

        if (-not $resolvedBackendBin -and -not $resolvedGoPath) {
            throw 'Neither backend smoke binary nor Go toolchain is available.'
        }

        $configPath = Join-Path $tempRoot 'config.json'
        $dbPath = Join-Path $tempRoot 'octopus.db'
        Write-ConfigJson -Path $configPath -StaticDir $staticDir -DbPath $dbPath -Port $effectiveBackendPort

        $backendStdout = Join-Path $tempRoot 'backend.stdout.log'
        $backendStderr = Join-Path $tempRoot 'backend.stderr.log'
        $backendEnv = @{
            OCTOPUS_ADMIN_USERNAME = 'admin'
            OCTOPUS_ADMIN_PASSWORD = 'admin'
        }
        if ($resolvedBackendBin) {
            $backendProc = Start-LoggedProcess -FilePath $resolvedBackendBin -ArgumentList @('start', '--config', $configPath) -WorkingDirectory $repoRoot -StdoutPath $backendStdout -StderrPath $backendStderr -ProcessEnvironment $backendEnv
        }
        else {
            $backendProc = Start-LoggedProcess -FilePath $resolvedGoPath -ArgumentList @('run', 'main.go', 'start', '--config', $configPath) -WorkingDirectory $repoRoot -StdoutPath $backendStdout -StderrPath $backendStderr -ProcessEnvironment $backendEnv
        }
        $processes += $backendProc

        Start-Sleep -Milliseconds 1200
        if ($backendProc.HasExited) {
            $backendStdoutTail = Get-LogTail -Path $backendStdout -LineCount 80
            $backendStderrTail = Get-LogTail -Path $backendStderr -LineCount 80
            throw (Resolve-BackendSelfStartFailure -StdoutText $backendStdoutTail -StderrText $backendStderrTail -BackendUrl $backendBaseUrl -StdoutPath $backendStdout -StderrPath $backendStderr)
        }

        Wait-Http -Url "$backendBaseUrl/healthz" -TimeoutSeconds 60
        Wait-Http -Url $frontendBaseUrl -TimeoutSeconds 90
    }
    else {
        Write-Step 'Verifying external backend and frontend'
        Assert-LoopbackLocalHttpReady -Url "$backendBaseUrl/healthz" -ScenarioLabel 'external backend healthcheck'
        Assert-LoopbackLocalHttpReady -Url $frontendBaseUrl -ScenarioLabel 'external frontend healthcheck'
        Wait-Http -Url "$backendBaseUrl/healthz" -TimeoutSeconds 30
        Wait-Http -Url $frontendBaseUrl -TimeoutSeconds 30
    }

    Write-Step ("Running {0} browser smoke" -f $smokeLabel)

    $nodeSmokeStdout = Join-Path $tempRoot 'node-smoke.stdout.log'
    $nodeSmokeStderr = Join-Path $tempRoot 'node-smoke.stderr.log'
    $nodeSmokeResult = Invoke-LoggedProcessWait -FilePath $resolvedNodePath -ArgumentList @($nodeSmokeScript, '--external') -WorkingDirectory $repoRoot -StdoutPath $nodeSmokeStdout -StderrPath $nodeSmokeStderr -TimeoutSeconds $NodeSmokeTimeoutSeconds -Description ("Node {0} smoke (cli)" -f $smokeLabel)
    Assert-NodeSmokeSucceeded -Result $nodeSmokeResult -ExpectedResult $nodeSmokeSuccessMarker -SmokeLabel $smokeLabel

    $verificationSucceeded = $true
    Write-Step (("{0} browser smoke passed" -f $smokeLabel.Substring(0,1).ToUpper() + $smokeLabel.Substring(1)))
    Write-Host ("Node smoke stdout: {0}" -f $nodeSmokeResult.StdoutPath)
    Write-Host ("Node smoke stderr: {0}" -f $nodeSmokeResult.StderrPath)
    Write-Host ("Frontend URL: $frontendBaseUrl")
    Write-Host ("Backend URL: $backendBaseUrl")
    Write-Host ("Artifacts: $tempRoot")
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
