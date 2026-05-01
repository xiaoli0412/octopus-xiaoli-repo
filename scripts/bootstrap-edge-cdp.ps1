[CmdletBinding(PositionalBinding = $false)]
param(
    [string]$BrowserPath,

    [int]$Port = 9222,

    [int]$PortSearchWindow = 20,

    [ValidateSet('default', 'relaxed', 'headed-relaxed')]
    [string]$EdgeLaunchPreset = 'relaxed',

    [ValidateSet('temp-random', 'workspace-fixed')]
    [string]$EdgeProfileStrategy = 'temp-random',

    [int]$ReadyTimeoutSeconds = 30,

    [int]$StableReadySeconds = 3,

    [string]$OutputJsonPath
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ''
    Write-Host ('== ' + $Message + ' ==')
}

function Get-RepoRoot {
    $scriptDir = Split-Path -Parent $PSCommandPath
    return [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))
}

function Normalize-Url {
    param([Parameter(Mandatory = $true)][string]$Value)

    return $Value.TrimEnd('/')
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
            return (Resolve-Path -LiteralPath $candidate).ProviderPath
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

function Resolve-OptionalOutputPath {
    param([string]$PathValue)

    if ([string]::IsNullOrWhiteSpace($PathValue)) {
        return $null
    }

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }

    return [System.IO.Path]::GetFullPath((Join-Path (Get-Location).Path $PathValue))
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

function Wait-Http {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [int]$TimeoutSeconds = 30
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        if (Test-HttpEndpoint -Url $Url -TimeoutSeconds 2) {
            return
        }

        Start-Sleep -Milliseconds 500
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
        return $false
    }
    finally {
        if ($null -ne $listener) {
            $listener.Stop()
        }
    }
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

function Start-LoggedProcess {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter()][string[]]$ArgumentList = @(),
        [Parameter(Mandatory = $true)][string]$WorkingDirectory,
        [Parameter(Mandatory = $true)][string]$StdoutPath,
        [Parameter(Mandatory = $true)][string]$StderrPath,
        [ValidateSet('Hidden', 'Normal', 'Minimized', 'Maximized')][string]$WindowStyle = 'Hidden'
    )

    $startParams = @{
        FilePath = $FilePath
        ArgumentList = $ArgumentList
        WorkingDirectory = $WorkingDirectory
        RedirectStandardOutput = $StdoutPath
        RedirectStandardError = $StderrPath
        PassThru = $true
        WindowStyle = $WindowStyle
    }

    return Start-Process @startParams
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

function Resolve-EdgeProfile {
    param(
        [Parameter(Mandatory = $true)][string]$RepoRoot,
        [Parameter(Mandatory = $true)][string]$TempRoot,
        [ValidateSet('temp-random', 'workspace-fixed')][string]$Strategy,
        [ValidateSet('default', 'relaxed', 'headed-relaxed')][string]$Preset
    )

    if ($Strategy -eq 'workspace-fixed') {
        $profileRoot = Join-Path $RepoRoot '.tools\verify-setting-help-browser-smoke\edge-profile'
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

function Get-LogTail {
    param(
        [string]$Path,
        [int]$LineCount = 60
    )

    if ([string]::IsNullOrWhiteSpace($Path) -or -not (Test-Path -LiteralPath $Path)) {
        return ''
    }

    return ((Get-Content -LiteralPath $Path -Tail $LineCount) -join [Environment]::NewLine).Trim()
}

function Write-Summary {
    param(
        [Parameter(Mandatory = $true)]$Summary,
        [string]$OutputJsonPath
    )

    $json = $Summary | ConvertTo-Json -Depth 5
    Write-Output $json

    if (-not [string]::IsNullOrWhiteSpace($OutputJsonPath)) {
        [System.IO.File]::WriteAllText($OutputJsonPath, ($json + [Environment]::NewLine), (New-Object System.Text.UTF8Encoding($false)))
    }
}

function Assert-StableCdpEndpoint {
    param(
        [Parameter(Mandatory = $true)][string]$BaseUrl,
        [Parameter(Mandatory = $true)][System.Diagnostics.Process]$Process,
        [int]$StableReadySeconds = 3
    )

    if ($StableReadySeconds -le 0) {
        return
    }

    $checks = [Math]::Max($StableReadySeconds * 2, 1)
    for ($index = 0; $index -lt $checks; $index++) {
        Start-Sleep -Milliseconds 500
        $Process.Refresh()

        if ($Process.HasExited) {
            throw "Edge process exited during CDP stability window."
        }

        if (-not (Test-HttpEndpoint -Url "$BaseUrl/json/version" -TimeoutSeconds 2)) {
            throw "CDP endpoint became unreachable during stability window at $BaseUrl/json/version"
        }
    }
}

$repoRoot = Get-RepoRoot
$tempRoot = Join-Path $env:TEMP ('octopus-edge-cdp-bootstrap-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempRoot | Out-Null
$resolvedOutputJsonPath = Resolve-OptionalOutputPath -PathValue $OutputJsonPath

$resolvedBrowserPath = Resolve-CommandPath -Candidates @(
    $BrowserPath,
    $env:OCTOPUS_UI_SMOKE_BROWSER_PATH,
    'C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe',
    'C:\Program Files\Microsoft\Edge\Application\msedge.exe'
) -CommandName $null

if (-not $resolvedBrowserPath) {
    throw 'Unable to resolve Edge executable.'
}

$cdpResolution = Resolve-CdpPort -PreferredPort $Port -SearchWindow $PortSearchWindow
$effectivePort = $cdpResolution.Port
$baseUrl = Normalize-Url -Value "http://127.0.0.1:$effectivePort"

if ($cdpResolution.ReuseExisting) {
    Write-Step 'Reusing existing Edge CDP endpoint'
    $summary = [pscustomobject]@{
        result = 'reused-existing-endpoint'
        browserPath = $resolvedBrowserPath
        baseUrl = $baseUrl
        jsonVersionUrl = "$baseUrl/json/version"
        port = $effectivePort
        reusedExisting = $true
        pid = $null
        profilePath = $null
        tempRoot = $null
    }

    Write-Summary -Summary $summary -OutputJsonPath $OutputJsonPath
    return
}

Write-Step 'Starting Edge remote debugging session'
$browserStdout = Join-Path $tempRoot 'edge.stdout.log'
$browserStderr = Join-Path $tempRoot 'edge.stderr.log'
$browserProfile = Resolve-EdgeProfile -RepoRoot $repoRoot -TempRoot $tempRoot -Strategy $EdgeProfileStrategy -Preset $EdgeLaunchPreset
$browserArgs = Get-EdgeLaunchArguments -RemoteDebuggingPort $effectivePort -UserDataDir $browserProfile.Path -Preset $EdgeLaunchPreset
$browserWindowStyle = Get-EdgeWindowStyle -Preset $EdgeLaunchPreset

Write-Host ("Browser: {0}" -f $resolvedBrowserPath)
Write-Host ("Profile: {0}" -f $browserProfile.Path)
Write-Host ("Port: {0}" -f $effectivePort)

$browserProc = Start-LoggedProcess -FilePath $resolvedBrowserPath -ArgumentList $browserArgs -WorkingDirectory $repoRoot -StdoutPath $browserStdout -StderrPath $browserStderr -WindowStyle $browserWindowStyle

try {
    Wait-Http -Url "$baseUrl/json/version" -TimeoutSeconds $ReadyTimeoutSeconds
    Assert-StableCdpEndpoint -BaseUrl $baseUrl -Process $browserProc -StableReadySeconds $StableReadySeconds

    $summary = [pscustomobject]@{
        result = 'started-new-endpoint'
        browserPath = $resolvedBrowserPath
        baseUrl = $baseUrl
        jsonVersionUrl = "$baseUrl/json/version"
        port = $effectivePort
        reusedExisting = $false
        pid = $browserProc.Id
        profilePath = $browserProfile.Path
        tempRoot = $tempRoot
        stdoutPath = $browserStdout
        stderrPath = $browserStderr
        launchPreset = $EdgeLaunchPreset
        profileStrategy = $EdgeProfileStrategy
        stableReadySeconds = $StableReadySeconds
    }

    Write-Summary -Summary $summary -OutputJsonPath $resolvedOutputJsonPath
}
catch {
    $stdoutTail = Get-LogTail -Path $browserStdout
    $stderrTail = Get-LogTail -Path $browserStderr

    Stop-ProcessTree -Process $browserProc

    throw ("Failed to bootstrap Edge CDP endpoint at {0}.`nArtifacts: {1}`nSTDOUT tail:`n{2}`nSTDERR tail:`n{3}" -f $baseUrl, $tempRoot, $stdoutTail, $stderrTail)
}
