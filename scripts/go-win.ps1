[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [ValidateSet('go', 'gofmt')]
    [string]$Tool,
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Arguments
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
. (Join-Path $scriptDir 'use-go-env.ps1')

$commandPath = if ($Tool -eq 'go') {
    $env:GOEXE
}
else {
    $env:GOFMTEXE
}

if (-not (Test-Path -LiteralPath $commandPath)) {
    throw "Unable to find $Tool executable: $commandPath"
}

& $commandPath @Arguments
$exitCode = if ($null -ne $LASTEXITCODE) { $LASTEXITCODE } else { 0 }
exit $exitCode
