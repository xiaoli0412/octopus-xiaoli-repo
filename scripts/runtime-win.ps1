[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('status', 'stop', 'healthcheck', 'check-only')]
    [string]$Action = 'status',
    [int[]]$Ports = @(3000, 3001, 8080),
    [switch]$IncludeNodeWorkers
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

function Test-IsServiceProviderFailureMessage {
    param([string]$Message)

    if ([string]::IsNullOrWhiteSpace($Message)) {
        return $false
    }

    return $Message -match '无法加载或初始化请求的服务提供程序|requested service provider|service provider could not be loaded or initialized'
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

function Resolve-HealthcheckFailure {
    param(
        [Parameter(Mandatory = $true)][int]$ExitCode,
        [Parameter(Mandatory = $true)][AllowEmptyString()][string]$StdoutText,
        [Parameter(Mandatory = $true)][AllowEmptyString()][string]$StderrText
    )

    $combined = @($StdoutText, $StderrText) -join [Environment]::NewLine
    if (Test-IsServiceProviderFailureMessage -Message $combined) {
        return "Local healthcheck hit a Windows service-provider initialization failure. Treat this as a host networking blocker instead of an Octopus regression.`nExit code: $ExitCode`nOutput:`n$combined"
    }

    return "healthcheck exited with code $ExitCode`nOutput:`n$combined"
}

function Test-IsGoBootstrapEnvironmentFailure {
    param(
        [Parameter(Mandatory = $true)][AllowEmptyString()][string]$StdoutText,
        [Parameter(Mandatory = $true)][AllowEmptyString()][string]$StderrText
    )

    $combined = @($StdoutText, $StderrText) -join [Environment]::NewLine
    if ([string]::IsNullOrWhiteSpace($combined)) {
        return $false
    }

    return ($combined -match '@v\/v[^"\s]+\.mod') -or
        ($combined -match 'module lookup disabled by GOPROXY') -or
        ($combined -match '\bGOMODCACHE\b|\bGOCACHE\b|\bGOTMPDIR\b') -or
        ($combined -match 'build cache is required')
}

function Invoke-GoHealthcheckCommand {
    param(
        [Parameter(Mandatory = $true)][string]$GoExecutable,
        [Parameter(Mandatory = $true)][string]$RepoRoot
    )

    $toolsRoot = Ensure-Directory -Path (Join-Path $RepoRoot '.tools')
    $healthcheckTemp = Ensure-Directory -Path (Join-Path $toolsRoot 'tmp')

    $stdoutPath = Join-Path $healthcheckTemp ('runtime-healthcheck-' + [guid]::NewGuid().ToString('N') + '.stdout.log')
    $stderrPath = Join-Path $healthcheckTemp ('runtime-healthcheck-' + [guid]::NewGuid().ToString('N') + '.stderr.log')

    try {
        $proc = Start-Process -FilePath $GoExecutable -ArgumentList @('run', 'main.go', 'healthcheck') -WorkingDirectory $RepoRoot -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru -Wait -WindowStyle Hidden
        $stdoutText = if (Test-Path -LiteralPath $stdoutPath) { [System.IO.File]::ReadAllText($stdoutPath) } else { '' }
        $stderrText = if (Test-Path -LiteralPath $stderrPath) { [System.IO.File]::ReadAllText($stderrPath) } else { '' }
        return [pscustomobject]@{
            StdoutText = $stdoutText.TrimEnd()
            StderrText = $stderrText.TrimEnd()
            ExitCode = $proc.ExitCode
        }
    }
    finally {
        Remove-Item -LiteralPath $stdoutPath, $stderrPath -Force -ErrorAction SilentlyContinue
    }
}

function Invoke-GoHealthcheckWithFallback {
    param(
        [Parameter(Mandatory = $true)][string]$GoExecutable,
        [Parameter(Mandatory = $true)][string]$RepoRoot
    )

    $result = Invoke-GoHealthcheckCommand -GoExecutable $GoExecutable -RepoRoot $RepoRoot
    if ($result.ExitCode -eq 0 -or -not (Test-IsGoBootstrapEnvironmentFailure -StdoutText $result.StdoutText -StderrText $result.StderrText)) {
        return $result
    }

    Write-Info 'Healthcheck hit a Go bootstrap environment failure; retrying with sanitized GOENV/GOPROXY for this process.'

    $savedGoEnv = $env:GOENV
    $savedGoProxy = $env:GOPROXY
    try {
        $env:GOENV = 'off'
        $env:GOPROXY = 'https://proxy.golang.org,direct'
        return (Invoke-GoHealthcheckCommand -GoExecutable $GoExecutable -RepoRoot $RepoRoot)
    }
    finally {
        $env:GOENV = $savedGoEnv
        $env:GOPROXY = $savedGoProxy
    }
}

function Test-LoopbackBindCapability {
    param([string]$BindHost = '127.0.0.1')

    $listener = $null
    try {
        $listener = New-Object System.Net.Sockets.TcpListener ([System.Net.IPAddress]::Parse($BindHost)), 0
        $listener.Start()
        return [pscustomobject]@{
            Ready = $true
            Detail = 'bind ok'
            FailureMessage = ''
        }
    }
    catch {
        $message = Get-ErrorMessageChain -ErrorRecord $_
        return [pscustomobject]@{
            Ready = $false
            Detail = $message
            FailureMessage = $message
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
    $serviceProviderBlocked = ((-not $bindProbe.Ready) -and (Test-IsServiceProviderFailureMessage -Message $bindProbe.FailureMessage)) -or ((-not $clientProbe.Ready) -and (Test-IsServiceProviderFailureMessage -Message $clientProbe.FailureMessage))

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
        Write-Info ('Loopback localhost readiness: blocked on {0} by Windows service-provider initialization; local self-start/external smoke will fail until the host networking stack is repaired.' -f $Capability.Host)
    }
    else {
        Write-Info ('Loopback localhost readiness: usable on {0} for local smoke preflight.' -f $Capability.Host)
    }

    $bindSummary = if ($Capability.BindProbe.Ready) { 'ready (bind ok)' } else { 'blocked ({0})' -f $Capability.BindProbe.Detail }
    $clientSummary = if ($Capability.ClientProbe.Ready) { $Capability.ClientProbe.Detail } else { 'blocked ({0})' -f $Capability.ClientProbe.Detail }

    Write-Info ('Loopback bind probe: {0}' -f $bindSummary)
    Write-Info ('Loopback client probe: {0}' -f $clientSummary)
}

function Assert-LoopbackHealthcheckReady {
    param([Parameter(Mandatory = $true)]$Capability)

    if (-not $Capability.ServiceProviderBlocked) {
        return
    }

    $bindDetail = if ([string]::IsNullOrWhiteSpace($Capability.BindProbe.Detail)) { 'bind probe unavailable' } else { $Capability.BindProbe.Detail }
    $clientDetail = if ([string]::IsNullOrWhiteSpace($Capability.ClientProbe.Detail)) { 'client probe unavailable' } else { $Capability.ClientProbe.Detail }

    throw ("Local healthcheck preflight is blocked before go run main.go healthcheck because loopback localhost on {0} cannot initialize the Windows service provider. Treat this as a host networking blocker instead of an Octopus regression.{1}Bind probe: {2}{1}Client probe: {3}" -f $Capability.Host, [Environment]::NewLine, $bindDetail, $clientDetail)
}

function Test-UsableCommandPath {
    param(
        [string]$Candidate,
        [string[]]$Arguments = @('--version')
    )

    if ([string]::IsNullOrWhiteSpace($Candidate)) {
        return $false
    }

    if (-not (Test-Path -LiteralPath $Candidate)) {
        return $false
    }

    try {
        & $Candidate @Arguments *> $null
        return $true
    }
    catch {
        return $false
    }
}

function Resolve-GoCommandPath {
    param([Parameter(Mandatory = $true)][string]$RepoRoot)

    $candidatePaths = @(
        $env:GOEXE,
        $env:OCTOPUS_UI_SMOKE_GO,
        (Join-Path $RepoRoot '.tools\go\go\bin\go.exe')
    )

    foreach ($candidate in $candidatePaths) {
        if (Test-UsableCommandPath -Candidate $candidate -Arguments @('version')) {
            return (Resolve-Path -LiteralPath $candidate).ProviderPath
        }
    }

    $goCommand = Get-Command -Name 'go' -ErrorAction SilentlyContinue
    if ($null -ne $goCommand -and (Test-UsableCommandPath -Candidate $goCommand.Source -Arguments @('version'))) {
        return $goCommand.Source
    }

    return $null
}

function Ensure-GoToolchainEnvironment {
    param([Parameter(Mandatory = $true)][string]$GoExecutable)

    $goBin = Split-Path -Parent $GoExecutable
    $goRoot = Split-Path -Parent $goBin

    $env:GOEXE = $GoExecutable
    $env:GOROOT = $goRoot

    $pathEntries = @($env:Path -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    if ($pathEntries -notcontains $goBin) {
        $env:Path = ($goBin + ';' + (($pathEntries -join ';').Trim(';'))).Trim(';')
    }
}

function Ensure-Directory {
    param([Parameter(Mandatory = $true)][string]$Path)

    if (-not (Test-Path -LiteralPath $Path)) {
        New-Item -ItemType Directory -Path $Path -Force | Out-Null
    }

    return $Path
}

function Test-WritableDirectory {
    param([Parameter(Mandatory = $true)][string]$Path)

    try {
        if ([string]::IsNullOrWhiteSpace($Path)) {
            return $false
        }
        if (-not (Test-Path -LiteralPath $Path)) {
            New-Item -ItemType Directory -Path $Path -Force | Out-Null
        }
        $probeFile = Join-Path $Path ('.octopus-write-test-' + [guid]::NewGuid().ToString('N'))
        [System.IO.File]::WriteAllText($probeFile, 'ok', [System.Text.Encoding]::UTF8)
        Remove-Item -LiteralPath $probeFile -Force -ErrorAction SilentlyContinue
        return $true
    }
    catch {
        return $false
    }
}

function Resolve-WritableWorkspaceDir {
    param(
        [string]$CurrentValue,
        [Parameter(Mandatory = $true)][string]$FallbackPath
    )

    if (-not [string]::IsNullOrWhiteSpace($CurrentValue) -and (Test-WritableDirectory -Path $CurrentValue)) {
        return $CurrentValue
    }

    if (-not (Test-WritableDirectory -Path $FallbackPath)) {
        throw ('Unable to provision a writable workspace directory at ' + $FallbackPath)
    }

    return $FallbackPath
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

function Ensure-GoWorkspaceEnvironment {
    param([Parameter(Mandatory = $true)][string]$RepoRoot)

    Ensure-CoreWindowsEnvironment

    $toolsRoot = Ensure-Directory -Path (Join-Path $RepoRoot '.tools')
    $defaultTemp = Ensure-Directory -Path (Join-Path $toolsRoot 'tmp')
    $defaultGoCache = Ensure-Directory -Path (Join-Path $toolsRoot 'gocache')
    $defaultGoModCache = Ensure-Directory -Path (Join-Path $toolsRoot 'gomodcache')
    $defaultGoTmp = Ensure-Directory -Path (Join-Path $toolsRoot 'gotmp')

    $env:TEMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TEMP -FallbackPath $defaultTemp
    $env:TMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TMP -FallbackPath $env:TEMP
    $env:GOCACHE = Resolve-WritableWorkspaceDir -CurrentValue $env:GOCACHE -FallbackPath $defaultGoCache
    $env:GOMODCACHE = Resolve-WritableWorkspaceDir -CurrentValue $env:GOMODCACHE -FallbackPath $defaultGoModCache
    $env:GOTMPDIR = Resolve-WritableWorkspaceDir -CurrentValue $env:GOTMPDIR -FallbackPath $defaultGoTmp
}

function Get-ListeningPortSnapshot {
    param([int[]]$Ports)

    $normalizedPorts = @($Ports | ForEach-Object { [int]$_ } | Sort-Object -Unique)
    $requestedPortMap = @{}
    foreach ($port in $normalizedPorts) {
        $requestedPortMap[$port] = $true
    }

    $listeners = [System.Collections.Generic.List[object]]::new()

    function Add-ListenerRecord {
        param(
            [System.Collections.Generic.List[object]]$Bucket,
            [int]$Port,
            [Nullable[int]]$OwningProcess
        )

        $Bucket.Add([pscustomobject]@{
            Port = $Port
            OwningProcess = if ($null -ne $OwningProcess) { [int]$OwningProcess } else { $null }
        })
    }

    $netTcpCommand = Get-Command -Name 'Get-NetTCPConnection' -ErrorAction SilentlyContinue
    if ($null -ne $netTcpCommand) {
        try {
            foreach ($listener in @(Get-NetTCPConnection -State Listen -ErrorAction Stop)) {
                $localPort = [int]$listener.LocalPort
                if ($requestedPortMap.ContainsKey($localPort)) {
                    Add-ListenerRecord -Bucket $listeners -Port $localPort -OwningProcess ([int]$listener.OwningProcess)
                }
            }

            return [pscustomobject]@{
                Listeners = @($listeners | Sort-Object Port -Unique)
                ScanMode = 'nettcp'
                ScanDetails = 'Get-NetTCPConnection'
                OwnershipAvailable = $true
            }
        }
        catch {
        }
    }

    try {
        $netstatOutput = & cmd /c netstat -ano -p tcp 2>$null
        if ($LASTEXITCODE -eq 0 -and $null -ne $netstatOutput) {
            foreach ($line in @($netstatOutput)) {
                if ([string]::IsNullOrWhiteSpace($line)) {
                    continue
                }

                $trimmedLine = $line.Trim()
                if ($trimmedLine -notmatch '^(TCP)\s+') {
                    continue
                }

                $columns = $trimmedLine -split '\s+'
                if ($columns.Length -lt 5) {
                    continue
                }

                $state = $columns[3]
                if ($state -ne 'LISTENING') {
                    continue
                }

                $localAddress = $columns[1]
                $lastColonIndex = $localAddress.LastIndexOf(':')
                if ($lastColonIndex -lt 0 -or $lastColonIndex -ge ($localAddress.Length - 1)) {
                    continue
                }

                $portText = $localAddress.Substring($lastColonIndex + 1)
                $localPort = 0
                if (-not [int]::TryParse($portText, [ref]$localPort)) {
                    continue
                }

                if (-not $requestedPortMap.ContainsKey($localPort)) {
                    continue
                }

                $owningProcess = $null
                $pidText = $columns[4]
                $parsedPid = 0
                if ([int]::TryParse($pidText, [ref]$parsedPid)) {
                    $owningProcess = $parsedPid
                }

                Add-ListenerRecord -Bucket $listeners -Port $localPort -OwningProcess $owningProcess
            }

            return [pscustomobject]@{
                Listeners = @($listeners | Sort-Object Port -Unique)
                ScanMode = 'netstat'
                ScanDetails = 'netstat -ano -p tcp'
                OwnershipAvailable = $true
            }
        }
    }
    catch {
    }

    try {
        foreach ($listener in @([System.Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners())) {
            $localPort = [int]$listener.Port
            if ($requestedPortMap.ContainsKey($localPort)) {
                Add-ListenerRecord -Bucket $listeners -Port $localPort -OwningProcess $null
            }
        }

        return [pscustomobject]@{
            Listeners = @($listeners | Sort-Object Port -Unique)
            ScanMode = 'dotnet'
            ScanDetails = '.NET IPGlobalProperties.GetActiveTcpListeners() (owning process unavailable)'
            OwnershipAvailable = $false
        }
    }
    catch {
        return [pscustomobject]@{
            Listeners = @()
            ScanMode = 'unavailable'
            ScanDetails = ('Port probing unavailable: {0}' -f $_.Exception.Message)
            OwnershipAvailable = $false
        }
    }
}

function Get-OctopusRepoProcess {
    param(
        [switch]$IncludeNodeWorkers,
        [int[]]$Ports = @(3000, 3001, 8080)
    )

    $patterns = @(
        '*D:\GPT-codex\octopus_repo*',
        '*octopus_repo\\main.go*',
        '*octopus_repo\\web\\*',
        '*octopus_repo\\static\\out*',
        '*octopus_repo\\build\\*',
        '*go-build*'
    )

    function New-ProcessRecord {
        param(
            [int]$ProcessId,
            [string]$Name,
            [string]$ExecutablePath,
            [string]$CommandLine
        )

        [pscustomobject]@{
            ProcessId = $ProcessId
            Name = $Name
            ExecutablePath = $ExecutablePath
            CommandLine = $CommandLine
        }
    }

    function Add-ProcessIfRelevant {
        param(
            [System.Collections.Generic.List[object]]$Bucket,
            [int]$ProcessId,
            [string]$Name,
            [string]$ExecutablePath,
            [string]$CommandLine
        )

        $nameLower = [string]$Name
        $isRelevantName = (
            $nameLower -in @('go.exe', 'node.exe', 'main.exe', 'main', 'go', 'node') -or
            $ExecutablePath -like '*go.exe' -or
            $ExecutablePath -like '*node.exe'
        )
        if (-not $isRelevantName) {
            return
        }

        $matched = $false
        foreach ($pattern in $patterns) {
            if ($ExecutablePath -like $pattern -or $CommandLine -like $pattern) {
                $matched = $true
                break
            }
        }

        if (-not $matched) {
            return
        }

        if (-not $IncludeNodeWorkers -and ($nameLower -eq 'node.exe' -or $nameLower -eq 'node')) {
            if ($CommandLine -notlike '*next*' -and $CommandLine -notlike '*octopus_repo\\web\\*' -and $ExecutablePath -notlike '*octopus_repo\\web\\*') {
                return
            }
        }

        $Bucket.Add((New-ProcessRecord -ProcessId $ProcessId -Name $Name -ExecutablePath $ExecutablePath -CommandLine $CommandLine))
    }

    $processes = [System.Collections.Generic.List[object]]::new()
    $scanMode = 'cim'
    $scanDetails = 'Get-CimInstance Win32_Process'

    try {
        Get-CimInstance Win32_Process | ForEach-Object {
            Add-ProcessIfRelevant -Bucket $processes -ProcessId $_.ProcessId -Name $_.Name -ExecutablePath ([string]$_.ExecutablePath) -CommandLine ([string]$_.CommandLine)
        }
    }
    catch {
        $portSnapshot = Get-ListeningPortSnapshot -Ports $Ports
        $watchedPids = @()
        $scanMode = 'fallback'
        if ($portSnapshot.OwnershipAvailable) {
            $watchedPids = @($portSnapshot.Listeners | ForEach-Object { $_.OwningProcess } | Where-Object { $null -ne $_ })
            $scanDetails = ('Get-Process + {0}' -f $portSnapshot.ScanDetails)
        }
        else {
            $scanDetails = ('Get-Process only ({0})' -f $portSnapshot.ScanDetails)
        }

        $watchedPids = @($watchedPids | Sort-Object -Unique)

        foreach ($proc in Get-Process -Name go,node,main -ErrorAction SilentlyContinue) {
            $path = [string]$proc.Path
            $commandLine = [string]::Empty
            $matchesRepoPath = $path -like '*octopus_repo*' -or $path -like '*go-build*'
            $isToolchainGo = $proc.ProcessName -eq 'go' -and $path -like '*\.tools\go\go\bin\go.exe'
            if ($watchedPids -contains $proc.Id -or ($matchesRepoPath -and -not $isToolchainGo)) {
                Add-ProcessIfRelevant -Bucket $processes -ProcessId $proc.Id -Name ($proc.ProcessName + '.exe') -ExecutablePath $path -CommandLine $commandLine
            }
        }

        foreach ($watchedProcessId in $watchedPids | Sort-Object -Unique) {
            try {
                $proc = Get-Process -Id $watchedProcessId -ErrorAction Stop
                Add-ProcessIfRelevant -Bucket $processes -ProcessId $proc.Id -Name ($proc.ProcessName + '.exe') -ExecutablePath ([string]$proc.Path) -CommandLine ''
            }
            catch {
                continue
            }
        }
    }

    return [pscustomobject]@{
        Processes = @($processes | Sort-Object ProcessId -Unique)
        ScanMode = $scanMode
        ScanDetails = $scanDetails
    }
}

function Get-PortStatus {
    param([int[]]$Ports)

    $snapshot = Get-ListeningPortSnapshot -Ports $Ports
    $ownerNameMap = @{}

    if ($snapshot.OwnershipAvailable) {
        $ownedPids = @($snapshot.Listeners | ForEach-Object { $_.OwningProcess } | Where-Object { $null -ne $_ } | Sort-Object -Unique)
        foreach ($ownedProcessId in $ownedPids) {
            try {
                $proc = Get-Process -Id ([int]$ownedProcessId) -ErrorAction Stop
                $ownerNameMap[[int]$ownedProcessId] = $proc.ProcessName
            }
            catch {
                $ownerNameMap[[int]$ownedProcessId] = 'unresolved'
            }
        }
    }

    $listenerMap = @{}
    foreach ($listener in $snapshot.Listeners) {
        if (-not $listenerMap.ContainsKey($listener.Port)) {
            $listenerMap[$listener.Port] = $listener
        }
    }

    $rows = foreach ($port in @($Ports | ForEach-Object { [int]$_ } | Sort-Object -Unique)) {
        $listener = if ($listenerMap.ContainsKey($port)) { $listenerMap[$port] } else { $null }
        [pscustomobject]@{
            Port = $port
            Listening = ($null -ne $listener)
            OwningProcess = if ($null -ne $listener -and $null -ne $listener.OwningProcess) { [int]$listener.OwningProcess } else { $null }
            OwningProcessName = if ($null -ne $listener -and $null -ne $listener.OwningProcess -and $ownerNameMap.ContainsKey([int]$listener.OwningProcess)) { [string]$ownerNameMap[[int]$listener.OwningProcess] } else { '' }
        }
    }

    return [pscustomobject]@{
        Rows = @($rows)
        ScanMode = $snapshot.ScanMode
        ScanDetails = $snapshot.ScanDetails
    }
}

function Show-Status {
    param(
        [System.Collections.IEnumerable]$Processes,
        [System.Collections.IEnumerable]$PortRows,
        [string]$PortScanDetails,
        [string]$ScanMode,
        [string]$ScanDetails,
        $LoopbackCapability
    )

    Write-Step -Message 'Octopus Repo Runtime Status'
    Write-Info 'Workspace root: D:\GPT-codex\octopus_repo'
    Write-Info 'Default policy: keep the project stopped locally; let automation use check-only/self-start/external flows when needed.'

    if ($ScanMode -eq 'fallback') {
        Write-Info 'Process scan mode: low-privilege fallback (CIM denied; using Get-Process + port probing)'
    }
    else {
        Write-Info ('Process scan mode: {0}' -f $ScanDetails)
    }

    if (-not [string]::IsNullOrWhiteSpace($PortScanDetails)) {
        Write-Info ('Port scan mode: {0}' -f $PortScanDetails)
    }

    $portOwnerHints = @($PortRows | Where-Object { $_.Listening -and (-not [string]::IsNullOrWhiteSpace($_.OwningProcessName)) })
    if ((@($Processes).Count -eq 0) -and ($portOwnerHints.Count -gt 0)) {
        Write-Info 'Port owner hints are informational only in low-privilege mode; stop still targets workspace-attributed octopus_repo processes only.'
    }

    if ($null -ne $LoopbackCapability) {
        Write-LoopbackCapabilityReport -Capability $LoopbackCapability
    }

    Write-Host ''
    Write-Host 'Processes:'
    if (-not $Processes -or @($Processes).Count -eq 0) {
        Write-Host '  none'
    }
    else {
        $Processes | Select-Object ProcessId, Name, ExecutablePath, CommandLine | Format-Table -AutoSize
    }

    Write-Host ''
    Write-Host 'Ports:'
    $PortRows | Select-Object Port, Listening, OwningProcess, OwningProcessName | Format-Table -AutoSize

    Write-Host ''
    Write-Host 'Automation entrypoints:'
    Write-Host '  root: D:\GPT-codex\octopus_repo'
    Write-Host '  frontend: D:\GPT-codex\octopus_repo\web'
    Write-Host '  scripts: D:\GPT-codex\octopus_repo\scripts'
    Write-Host '  docs: D:\GPT-codex\octopus_repo\docs'
}

function Stop-OctopusRepoProcesses {
    param([System.Collections.IEnumerable]$Processes)

    if (-not $Processes -or @($Processes).Count -eq 0) {
        Write-Success 'No octopus_repo runtime process is currently running.'
        return
    }

    Write-Step -Message 'Stopping octopus_repo runtime processes'
    foreach ($procItem in $Processes) {
        try {
            Stop-Process -Id $procItem.ProcessId -Force -ErrorAction Stop
            Write-Success ("Stopped PID {0} ({1})" -f $procItem.ProcessId, $procItem.Name)
        }
        catch {
            Write-Info ("Skipped PID {0}: {1}" -f $procItem.ProcessId, $_.Exception.Message)
        }
    }
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))

if (-not (Test-Path -LiteralPath (Join-Path $repoRoot 'main.go'))) {
    throw "Repository root not detected: $repoRoot"
}

if ($Action -eq 'check-only') {
    $resolvedGoPath = Resolve-GoCommandPath -RepoRoot $repoRoot
    $loopbackCapability = Get-LoopbackCapabilitySummary
    Write-Step -Message 'Runtime management check-only'
    Write-Info ('Command: powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action status')
    Write-Info ('Command: powershell -ExecutionPolicy Bypass -File .\scripts\runtime-win.ps1 -Action stop')
    if ($resolvedGoPath) {
        Write-Info ('Healthcheck Go executable: {0}' -f $resolvedGoPath)
    }
    else {
        Write-Info 'Healthcheck Go executable: unresolved (run .\scripts\verify-go-env.ps1 if this machine does not expose Go on PATH)'
    }
    Write-LoopbackCapabilityReport -Capability $loopbackCapability
    Write-Info ('Command: go run main.go healthcheck')
    Write-Success 'Runtime management entrypoints are available.'
    return
}

$runtimeScan = Get-OctopusRepoProcess -IncludeNodeWorkers:$IncludeNodeWorkers -Ports $Ports
$processes = @($runtimeScan.Processes)

switch ($Action) {
    'status' {
        $portsState = Get-PortStatus -Ports $Ports
        $loopbackCapability = Get-LoopbackCapabilitySummary
        Show-Status -Processes $processes -PortRows $portsState.Rows -PortScanDetails $portsState.ScanDetails -ScanMode $runtimeScan.ScanMode -ScanDetails $runtimeScan.ScanDetails -LoopbackCapability $loopbackCapability
        break
    }
    'stop' {
        Stop-OctopusRepoProcesses -Processes $processes
        $remainingScan = Get-OctopusRepoProcess -IncludeNodeWorkers:$IncludeNodeWorkers -Ports $Ports
        $remaining = @($remainingScan.Processes)
        $portsState = Get-PortStatus -Ports $Ports
        $loopbackCapability = Get-LoopbackCapabilitySummary
        Show-Status -Processes $remaining -PortRows $portsState.Rows -PortScanDetails $portsState.ScanDetails -ScanMode $remainingScan.ScanMode -ScanDetails $remainingScan.ScanDetails -LoopbackCapability $loopbackCapability
        break
    }
    'healthcheck' {
        $resolvedGoPath = Resolve-GoCommandPath -RepoRoot $repoRoot
        if (-not $resolvedGoPath) {
            throw 'Unable to resolve a runnable Go executable for healthcheck. Run .\scripts\verify-go-env.ps1 or .\scripts\use-go-env.ps1 first.'
        }

        Ensure-GoToolchainEnvironment -GoExecutable $resolvedGoPath

        $loopbackCapability = Get-LoopbackCapabilitySummary
        Assert-LoopbackHealthcheckReady -Capability $loopbackCapability

        Ensure-GoWorkspaceEnvironment -RepoRoot $repoRoot

        Write-Step -Message 'Running healthcheck against the current local service'
        Write-Info ('Go executable: {0}' -f $resolvedGoPath)
        $healthcheckResult = Invoke-GoHealthcheckWithFallback -GoExecutable $resolvedGoPath -RepoRoot $repoRoot

        if (-not [string]::IsNullOrWhiteSpace($healthcheckResult.StdoutText)) {
            Write-Host $healthcheckResult.StdoutText
        }
        if (-not [string]::IsNullOrWhiteSpace($healthcheckResult.StderrText)) {
            Write-Host $healthcheckResult.StderrText
        }

        if ($healthcheckResult.ExitCode -ne 0) {
            throw (Resolve-HealthcheckFailure -ExitCode $healthcheckResult.ExitCode -StdoutText $healthcheckResult.StdoutText.Trim() -StderrText $healthcheckResult.StderrText.Trim())
        }
        break
    }
}
