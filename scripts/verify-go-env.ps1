[CmdletBinding()]
param(
    [switch]$GoFmt,
    [switch]$GoTest,
    [switch]$GoBuild,
    [switch]$PhaseA
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ''
    Write-Host ('== ' + $Message + ' ==')
}

function Invoke-NativeCommand {
    param(
        [Parameter(Mandatory = $true)][string]$Command,
        [Parameter(Mandatory = $true)][scriptblock]$ScriptBlock
    )

    & $ScriptBlock
    if ($LASTEXITCODE -ne 0) {
        throw ($Command + ' failed with exit code ' + $LASTEXITCODE)
    }
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

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $scriptDir '..'))

if (-not (Test-Path -LiteralPath (Join-Path $repoRoot 'go.mod'))) {
    throw ('Repository root not detected: ' + $repoRoot)
}

$repoToolsRoot = Join-Path $repoRoot '.tools'
$env:GOCACHE = Resolve-WritableWorkspaceDir -CurrentValue $env:GOCACHE -FallbackPath (Join-Path $repoToolsRoot 'gocache')
$env:GOTMPDIR = Resolve-WritableWorkspaceDir -CurrentValue $env:GOTMPDIR -FallbackPath (Join-Path $repoToolsRoot 'gotmp')
$env:TEMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TEMP -FallbackPath (Join-Path $repoToolsRoot 'tmp')
$env:TMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TMP -FallbackPath $env:TEMP

. (Join-Path $scriptDir 'use-go-env.ps1')

Push-Location $repoRoot
try {
    Write-Step -Message 'Go environment'
    Invoke-NativeCommand -Command 'go version' -ScriptBlock { & $env:GOEXE version }
    Invoke-NativeCommand -Command 'go env GOROOT GOPATH GOMOD GOEXE GOTOOLDIR' -ScriptBlock { & $env:GOEXE env GOROOT GOPATH GOMOD GOEXE GOTOOLDIR }
    Write-Host ('GOCACHE=' + $env:GOCACHE)
    Write-Host ('GOTMPDIR=' + $env:GOTMPDIR)
    Write-Host ('TEMP=' + $env:TEMP)
    Write-Host ('TMP=' + $env:TMP)

    if (-not ($GoFmt -or $GoTest -or $GoBuild -or $PhaseA)) {
        Write-Host ''
        Write-Host 'No extra validation switches requested. Use -GoFmt, -GoTest, -GoBuild, or -PhaseA.'
    }

    if ($GoFmt) {
        Write-Step -Message 'gofmt -l'
        $excludedRoots = @(
            [System.IO.Path]::GetFullPath((Join-Path $repoRoot '.git')),
            [System.IO.Path]::GetFullPath((Join-Path $repoRoot '.tools')),
            [System.IO.Path]::GetFullPath((Join-Path $repoRoot '.gocache')),
            [System.IO.Path]::GetFullPath((Join-Path $repoRoot '.gomodcache')),
            [System.IO.Path]::GetFullPath((Join-Path $repoRoot 'node_modules'))
        )
        $goFiles = Get-ChildItem -LiteralPath $repoRoot -Recurse -Filter '*.go' -File -ErrorAction SilentlyContinue |
            Where-Object {
                $fullName = [System.IO.Path]::GetFullPath($_.FullName)
                foreach ($excludedRoot in $excludedRoots) {
                    if ($fullName.StartsWith($excludedRoot + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
                        return $false
                    }
                }
                return $true
            } |
            Sort-Object FullName
        $needsFormat = New-Object System.Collections.Generic.List[string]

        foreach ($goFile in $goFiles) {
            $result = & $env:GOFMTEXE -l $goFile.FullName
            if ($LASTEXITCODE -ne 0) {
                throw ('gofmt failed for ' + $goFile.FullName)
            }

            if ($result) {
                $needsFormat.Add($goFile.FullName)
            }
        }

        Write-Host ('GO_FILE_COUNT=' + $goFiles.Count)
        Write-Host ('UNFORMATTED_COUNT=' + $needsFormat.Count)
        foreach ($path in $needsFormat) {
            Write-Host $path
        }
    }

    if ($GoTest) {
        Write-Step -Message 'go test ./...'
        Invoke-NativeCommand -Command 'go test ./...' -ScriptBlock { & $env:GOEXE test ./... }
    }

    if ($GoBuild) {
        Write-Step -Message 'go build ./...'
        Invoke-NativeCommand -Command 'go build ./...' -ScriptBlock { & $env:GOEXE build ./... }
    }

    if ($PhaseA) {
        Write-Step -Message 'scripts\\phase-a-check.ps1'
        Invoke-NativeCommand -Command 'scripts\\phase-a-check.ps1' -ScriptBlock { & (Join-Path $scriptDir 'phase-a-check.ps1') }
    }
}
finally {
    Pop-Location
}
