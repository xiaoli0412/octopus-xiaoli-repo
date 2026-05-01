$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Title,
        [Parameter(Mandatory = $true)]
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host ("== " + $Title + " ==")
    & $Action
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir
$webDir = Join-Path $repoRoot 'web'

. (Join-Path $scriptDir 'use-go-env.ps1')

if (-not (Test-Path (Join-Path $repoRoot 'main.go'))) {
    throw "Repository root not detected: $repoRoot"
}

if (-not (Test-Path (Join-Path $webDir 'package.json'))) {
    throw "Web directory not detected: $webDir"
}

Push-Location $repoRoot
try {
    Invoke-Step -Title 'go test' -Action {
        & $env:GOEXE test ./...
    }

    Invoke-Step -Title 'go build' -Action {
        & $env:GOEXE build ./...
    }

    Invoke-Step -Title 'go help' -Action {
        & $env:GOEXE run main.go --help
    }

    Push-Location $webDir
    try {
        Invoke-Step -Title 'web build' -Action {
            pnpm build
        }
    }
    finally {
        Pop-Location
    }
}
finally {
    Pop-Location
}
