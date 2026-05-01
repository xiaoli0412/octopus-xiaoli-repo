[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

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

function Test-InvalidGoProxyValue {
    param([string]$Value)

    if ([string]::IsNullOrWhiteSpace($Value)) {
        return $true
    }

    $trimmed = $Value.Trim()
    if ($trimmed -match '^[,|]+$') {
        return $true
    }

    foreach ($entry in ($trimmed -split ',')) {
        if ([string]::IsNullOrWhiteSpace($entry)) {
            return $true
        }
    }

    return $false
}

function Invoke-GoProbe {
    param(
        [Parameter(Mandatory = $true)][string]$CommandPath,
        [Parameter(Mandatory = $true)][string[]]$Arguments
    )

    try {
        $output = & $CommandPath @Arguments 2>&1
        $exitCode = if ($null -ne $LASTEXITCODE) { $LASTEXITCODE } else { 0 }
        $outputText = ($output | Out-String).Trim()
        return [pscustomobject]@{
            Success = ($exitCode -eq 0)
            Output = $outputText
            Error = if ($exitCode -eq 0) { $null } else { $outputText }
            ExitCode = $exitCode
        }
    }
    catch {
        return [pscustomobject]@{
            Success = $false
            Output = $null
            Error = $_.Exception.Message
            ExitCode = -1
        }
    }
}

function Add-UniqueCandidate {
    param(
        [Parameter(Mandatory = $true)]
        [AllowEmptyCollection()]
        [System.Collections.IList]$Candidates,
        [string]$GoBin,
        [string]$GoCommand,
        [string]$GoFmtCommand,
        [string]$Label
    )

    if ([string]::IsNullOrWhiteSpace($GoCommand) -or [string]::IsNullOrWhiteSpace($GoFmtCommand)) {
        return
    }

    $key = ($GoCommand + '|' + $GoFmtCommand).ToLowerInvariant()
    foreach ($candidate in $Candidates) {
        if ($candidate.Key -eq $key) {
            return
        }
    }

    $Candidates.Add([pscustomobject]@{
        Key = $key
        GoBin = $GoBin
        GoCommand = $GoCommand
        GoFmtCommand = $GoFmtCommand
        Label = $Label
    })
}

function Get-GoToolchainProbe {
    param(
        [string]$GoBin,
        [Parameter(Mandatory = $true)][string]$GoCommand,
        [Parameter(Mandatory = $true)][string]$GoFmtCommand,
        [string]$Label
    )

    $goExists = Test-Path -LiteralPath $GoCommand
    $gofmtExists = Test-Path -LiteralPath $GoFmtCommand
    if (-not $goExists) {
        return [pscustomobject]@{
            GoBin = $GoBin
            GoExe = $GoCommand
            GoFmtExe = $GoFmtCommand
            Status = 'missing go command'
            Detail = $Label
            GoVersion = $null
        }
    }
    if (-not $gofmtExists) {
        return [pscustomobject]@{
            GoBin = $GoBin
            GoExe = $GoCommand
            GoFmtExe = $GoFmtCommand
            Status = 'missing gofmt command'
            Detail = $Label
            GoVersion = $null
        }
    }

    $goProbe = Invoke-GoProbe -CommandPath $GoCommand -Arguments @('version')
    if (-not $goProbe.Success) {
        return [pscustomobject]@{
            GoBin = $GoBin
            GoExe = $GoCommand
            GoFmtExe = $GoFmtCommand
            Status = 'go command not runnable'
            Detail = $goProbe.Error
            GoVersion = $null
        }
    }

    $gofmtProbe = Invoke-GoProbe -CommandPath $GoFmtCommand -Arguments @('-h')
    $gofmtText = if (-not [string]::IsNullOrWhiteSpace($gofmtProbe.Error)) { $gofmtProbe.Error } else { $gofmtProbe.Output }
    $gofmtUsable = $gofmtProbe.Success -or ($gofmtText -match 'usage:\s*gofmt')
    if (-not $gofmtUsable) {
        return [pscustomobject]@{
            GoBin = $GoBin
            GoExe = $GoCommand
            GoFmtExe = $GoFmtCommand
            Status = 'gofmt command not runnable'
            Detail = $gofmtProbe.Error
            GoVersion = $null
        }
    }

    return [pscustomobject]@{
        GoBin = $GoBin
        GoExe = $GoCommand
        GoFmtExe = $GoFmtCommand
        Status = 'ok'
        Detail = $Label
        GoVersion = $goProbe.Output
    }
}

function Resolve-GoToolchain {
    $candidates = New-Object 'System.Collections.Generic.List[object]'

    if (-not [string]::IsNullOrWhiteSpace($env:OCTOPUS_GO_BIN)) {
        $goBin = $env:OCTOPUS_GO_BIN.Trim()
        Add-UniqueCandidate -Candidates $candidates -GoBin $goBin -GoCommand (Join-Path $goBin 'go.exe') -GoFmtCommand (Join-Path $goBin 'gofmt.exe') -Label 'OCTOPUS_GO_BIN'
    }

    if (-not [string]::IsNullOrWhiteSpace($env:GO_BIN)) {
        $goBin = $env:GO_BIN.Trim()
        Add-UniqueCandidate -Candidates $candidates -GoBin $goBin -GoCommand (Join-Path $goBin 'go.exe') -GoFmtCommand (Join-Path $goBin 'gofmt.exe') -Label 'GO_BIN'
    }

    if (-not [string]::IsNullOrWhiteSpace($env:GOROOT)) {
        $goBin = Join-Path $env:GOROOT 'bin'
        Add-UniqueCandidate -Candidates $candidates -GoBin $goBin -GoCommand (Join-Path $goBin 'go.exe') -GoFmtCommand (Join-Path $goBin 'gofmt.exe') -Label 'GOROOT'
    }

    if (-not [string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
        $localAppGoBin = Join-Path $env:LOCALAPPDATA 'Programs\go\bin'
        Add-UniqueCandidate -Candidates $candidates -GoBin $localAppGoBin -GoCommand (Join-Path $localAppGoBin 'go.exe') -GoFmtCommand (Join-Path $localAppGoBin 'gofmt.exe') -Label 'LOCALAPPDATA go bin'
    }

    if (-not [string]::IsNullOrWhiteSpace($env:APPDATA)) {
        $npmBin = Join-Path $env:APPDATA 'npm'
        Add-UniqueCandidate -Candidates $candidates -GoBin $npmBin -GoCommand (Join-Path $npmBin 'go.cmd') -GoFmtCommand (Join-Path $npmBin 'gofmt.cmd') -Label 'npm wrappers'
    }

    $goCommand = Get-Command go -ErrorAction SilentlyContinue
    if ($null -ne $goCommand) {
        $goSource = $goCommand.Source
        $goBin = Split-Path -Parent $goSource
        $gofmtCmd = Get-Command gofmt -ErrorAction SilentlyContinue
        $gofmtSource = if ($null -ne $gofmtCmd) { $gofmtCmd.Source } else { Join-Path $goBin 'gofmt.exe' }
        Add-UniqueCandidate -Candidates $candidates -GoBin $goBin -GoCommand $goSource -GoFmtCommand $gofmtSource -Label 'PATH go/gofmt'
    }

    $probes = foreach ($candidate in $candidates) {
        Get-GoToolchainProbe -GoBin $candidate.GoBin -GoCommand $candidate.GoCommand -GoFmtCommand $candidate.GoFmtCommand -Label $candidate.Label
    }

    $resolved = $probes | Where-Object { $_.Status -eq 'ok' } | Select-Object -First 1
    if ($null -ne $resolved) {
        return $resolved
    }

    $details = if ($probes.Count -gt 0) {
        ($probes | ForEach-Object {
            $suffix = if ([string]::IsNullOrWhiteSpace($_.Detail)) { '' } else { ' - ' + $_.Detail }
            '  - ' + $_.GoExe + ' [' + $_.Status + ']' + $suffix
        }) -join "`n"
    }
    else {
        '  - no candidate paths were discovered'
    }

    throw ("Unable to find a runnable Go toolchain.`nChecked:`n" + $details + "`nSet OCTOPUS_GO_BIN to a working Go bin directory if this machine uses a non-default installation.")
}

$goToolchain = Resolve-GoToolchain
$goBin = $goToolchain.GoBin
$goExe = $goToolchain.GoExe
$gofmtExe = $goToolchain.GoFmtExe

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))
$toolsRoot = Join-Path $repoRoot '.tools'
$tempRoot = Join-Path $toolsRoot 'tmp'
$goCacheRoot = Join-Path $toolsRoot 'gocache'
$goModCacheRoot = Join-Path $toolsRoot 'gomodcache'
$goTmpRoot = Join-Path $toolsRoot 'gotmp'

foreach ($path in @($toolsRoot, $tempRoot, $goCacheRoot, $goModCacheRoot, $goTmpRoot)) {
    if (-not (Test-Path -LiteralPath $path)) {
        New-Item -ItemType Directory -Path $path -Force | Out-Null
    }
}

if (-not [string]::IsNullOrWhiteSpace($goBin) -and (($env:PATH -split ';') -notcontains $goBin)) {
    $env:PATH = $goBin + ';' + $env:PATH
}

if (-not [string]::IsNullOrWhiteSpace($goBin)) {
    $env:GOROOT = Split-Path -Parent $goBin
}

if ([string]::IsNullOrWhiteSpace($env:LOCALAPPDATA) -and -not [string]::IsNullOrWhiteSpace($env:USERPROFILE)) {
    $env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'
}

$env:TEMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TEMP -FallbackPath $tempRoot
$env:TMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TMP -FallbackPath $env:TEMP
$env:GOCACHE = Resolve-WritableWorkspaceDir -CurrentValue $env:GOCACHE -FallbackPath $goCacheRoot
$env:GOMODCACHE = Resolve-WritableWorkspaceDir -CurrentValue $env:GOMODCACHE -FallbackPath $goModCacheRoot
$env:GOTMPDIR = Resolve-WritableWorkspaceDir -CurrentValue $env:GOTMPDIR -FallbackPath $goTmpRoot

if (Test-InvalidGoProxyValue -Value $env:GOPROXY) {
    $env:GOPROXY = 'https://proxy.golang.org,direct'
}
else {
    $env:GOPROXY = $env:GOPROXY.Trim()
}

if ([string]::IsNullOrWhiteSpace($env:GOSUMDB)) {
    $env:GOSUMDB = 'sum.golang.org'
}
else {
    $env:GOSUMDB = $env:GOSUMDB.Trim()
}

$env:GOEXE = $goExe
$env:GOFMTEXE = $gofmtExe

Set-Alias -Name go.exe -Value $goExe -Scope Script
Set-Alias -Name gofmt.exe -Value $gofmtExe -Scope Script
Set-Alias -Name go -Value $goExe -Scope Script
Set-Alias -Name gofmt -Value $gofmtExe -Scope Script

Write-Host 'Enabled Go toolchain for this PowerShell session:'
Write-Host ('  GO_BIN=' + $goBin)
Write-Host ('  GOROOT=' + $env:GOROOT)
Write-Host ('  GOEXE=' + $goExe)
Write-Host ('  GOFMTEXE=' + $gofmtExe)
Write-Host ('  LOCALAPPDATA=' + $env:LOCALAPPDATA)
Write-Host ('  TEMP=' + $env:TEMP)
Write-Host ('  TMP=' + $env:TMP)
Write-Host ('  GOCACHE=' + $env:GOCACHE)
Write-Host ('  GOMODCACHE=' + $env:GOMODCACHE)
Write-Host ('  GOTMPDIR=' + $env:GOTMPDIR)
Write-Host ('  GOPROXY=' + $env:GOPROXY)
Write-Host ('  GOSUMDB=' + $env:GOSUMDB)
Write-Host ('  GO_VERSION=' + $goToolchain.GoVersion)
