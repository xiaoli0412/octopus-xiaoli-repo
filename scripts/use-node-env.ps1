[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

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

	if ([string]::IsNullOrWhiteSpace($env:ProgramData)) {
		$env:ProgramData = 'C:\ProgramData'
	}

	if ([string]::IsNullOrWhiteSpace($env:LOCALAPPDATA) -and -not [string]::IsNullOrWhiteSpace($env:USERPROFILE)) {
		$env:LOCALAPPDATA = Join-Path $env:USERPROFILE 'AppData\Local'
	}

	if ([string]::IsNullOrWhiteSpace($env:APPDATA) -and -not [string]::IsNullOrWhiteSpace($env:USERPROFILE)) {
		$env:APPDATA = Join-Path $env:USERPROFILE 'AppData\Roaming'
	}

	if ([string]::IsNullOrWhiteSpace($env:TEMP) -and -not [string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
		$env:TEMP = Join-Path $env:LOCALAPPDATA 'Temp'
	}

	if ([string]::IsNullOrWhiteSpace($env:TMP) -and -not [string]::IsNullOrWhiteSpace($env:TEMP)) {
		$env:TMP = $env:TEMP
	}
}

function Test-IsCodexBundledNode {
	param(
		[Parameter(Mandatory = $true)]
		[string]$Candidate
	)

	$normalized = $Candidate.Replace('/', '\').ToLowerInvariant()
	return $normalized.Contains('\appdata\local\openai\codex\bin\node.exe') -or $normalized.Contains('\windowsapps\openai.codex_')
}

function Test-UsableNodeExe {
	param(
		[Parameter(Mandatory = $true)]
		[string]$Candidate
	)

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
	} catch {
		return $false
	}
}

function Resolve-NodeExe {
	if (-not [string]::IsNullOrWhiteSpace($env:OCTOPUS_NODE_EXE)) {
		$candidate = $env:OCTOPUS_NODE_EXE.Trim()
		if (Test-UsableNodeExe -Candidate $candidate) {
			return $candidate
		}
	}

	if (-not [string]::IsNullOrWhiteSpace($env:OCTOPUS_NODE_BIN)) {
		$candidate = Join-Path $env:OCTOPUS_NODE_BIN.Trim() 'node.exe'
		if (Test-UsableNodeExe -Candidate $candidate) {
			return $candidate
		}
	}

	if (-not [string]::IsNullOrWhiteSpace($env:LOCALAPPDATA)) {
		$candidate = Join-Path $env:LOCALAPPDATA 'Programs\nodejs\node.exe'
		if (Test-UsableNodeExe -Candidate $candidate) {
			return $candidate
		}
	}

	foreach ($candidate in @(
		'D:\gptm\live-2d\node\node.exe',
		'D:\gol1\node.exe'
	)) {
		if (Test-UsableNodeExe -Candidate $candidate) {
			return $candidate
		}
	}

	if (-not [string]::IsNullOrWhiteSpace($env:NODE_BIN)) {
		$candidate = Join-Path $env:NODE_BIN.Trim() 'node.exe'
		if (Test-UsableNodeExe -Candidate $candidate) {
			return $candidate
		}
	}

	$nodeCommand = Get-Command node -ErrorAction SilentlyContinue
	if ($null -ne $nodeCommand -and -not [string]::IsNullOrWhiteSpace($nodeCommand.Source) -and (Test-UsableNodeExe -Candidate $nodeCommand.Source)) {
		return $nodeCommand.Source
	}

	throw 'Unable to find a runnable node.exe. Set OCTOPUS_NODE_BIN to a valid Node bin directory.'
}

function Resolve-PnpmVersion {
	param(
		[Parameter(Mandatory = $true)]
		[string]$RepoRoot
	)

	$candidates = @(
		(Join-Path $RepoRoot 'web\package.json')
		(Join-Path $RepoRoot 'package.json')
	)

	foreach ($candidate in $candidates) {
		if (-not (Test-Path -LiteralPath $candidate)) {
			continue
		}

		try {
			$packageJson = Get-Content -LiteralPath $candidate -Raw | ConvertFrom-Json -ErrorAction Stop
		}
		catch {
			throw ('Unable to parse package.json for pnpm version: ' + $candidate + ' (' + $_.Exception.Message + ')')
		}

		$packageManager = [string]$packageJson.packageManager
		if ([string]::IsNullOrWhiteSpace($packageManager)) {
			continue
		}

		$match = [regex]::Match($packageManager, '^pnpm@(?<version>[^+]+)')
		if ($match.Success) {
			return $match.Groups['version'].Value
		}
	}

	if (-not [string]::IsNullOrWhiteSpace($env:OCTOPUS_PNPM_VERSION)) {
		return $env:OCTOPUS_PNPM_VERSION.Trim()
	}

	throw 'Unable to resolve the required pnpm version from packageManager. Set OCTOPUS_PNPM_VERSION explicitly if needed.'
}

function Set-AsciiFile {
	param(
		[Parameter(Mandatory = $true)]
		[string]$Path,
		[Parameter(Mandatory = $true)]
		[string]$Content
	)

	[System.IO.File]::WriteAllText($Path, $Content, [System.Text.Encoding]::ASCII)
}

Ensure-CoreWindowsEnvironment

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$toolCacheRoot = Join-Path $repoRoot '.tmp-tooling'
$corepackHome = Join-Path $toolCacheRoot 'corepack'
$pnpmHome = Join-Path $toolCacheRoot 'pnpm-home'
$xdgConfigHome = Join-Path $toolCacheRoot 'xdg-config'
$userConfigPath = Join-Path $toolCacheRoot 'npmrc'

foreach ($pathToCreate in @($toolCacheRoot, $corepackHome, $pnpmHome, $xdgConfigHome)) {
	if (-not (Test-Path -LiteralPath $pathToCreate)) {
		New-Item -ItemType Directory -Force -Path $pathToCreate | Out-Null
	}
}

if (-not (Test-Path -LiteralPath $userConfigPath)) {
	New-Item -ItemType File -Force -Path $userConfigPath | Out-Null
}

$nodeExe = Resolve-NodeExe
$nodeBin = Split-Path -Parent $nodeExe
$corepackCli = Join-Path $nodeBin 'node_modules\corepack\dist\corepack.js'
$pnpmVersion = Resolve-PnpmVersion -RepoRoot $repoRoot
$pnpmSpec = 'pnpm@' + $pnpmVersion
$pnpmWrapperPath = Join-Path $toolCacheRoot 'pnpm.cmd'

if (($env:PATH -split ';') -notcontains $nodeBin) {
	$env:PATH = $nodeBin + ';' + $env:PATH
}

$pnpmVersionText = (& $nodeExe $corepackCli $pnpmSpec '--version').Trim()
if ([string]::IsNullOrWhiteSpace($pnpmVersionText)) {
	throw ('Failed to resolve ' + $pnpmSpec + ' through corepack.')
}

$pnpmWrapper = @"
@echo off
setlocal
"%NODEEXE%" "%COREPACK_CLI%" $pnpmSpec %*
"@
$shouldWriteWrapper = $true
if (Test-Path -LiteralPath $pnpmWrapperPath) {
	try {
		$existingWrapper = Get-Content -LiteralPath $pnpmWrapperPath -Raw -ErrorAction Stop
		if ($existingWrapper -eq $pnpmWrapper) {
			$shouldWriteWrapper = $false
		}
	}
	catch {
		$shouldWriteWrapper = $true
	}
}

if ($shouldWriteWrapper) {
	try {
		Set-AsciiFile -Path $pnpmWrapperPath -Content $pnpmWrapper
	}
	catch {
		if (-not (Test-Path -LiteralPath $pnpmWrapperPath)) {
			throw
		}

		$existingWrapper = Get-Content -LiteralPath $pnpmWrapperPath -Raw -ErrorAction Stop
		if ($existingWrapper -ne $pnpmWrapper) {
			throw
		}
	}
}

$env:NODEEXE = $nodeExe
$env:NODE_BIN = $nodeBin
$env:COREPACK_HOME = $corepackHome
$env:PNPM_HOME = $pnpmHome
$env:COREPACK_CLI = $corepackCli
$env:PNPMEXE = $pnpmWrapperPath
$env:OCTOPUS_PNPM_VERSION = $pnpmVersion
$env:OCTOPUS_PNPM_SPEC = $pnpmSpec
$env:XDG_CONFIG_HOME = $xdgConfigHome
$env:npm_config_userconfig = $userConfigPath

Set-Alias -Name node -Value $nodeExe -Scope Script
Set-Alias -Name node.exe -Value $nodeExe -Scope Script

function global:pnpm {
	param(
		[Parameter(ValueFromRemainingArguments = $true)]
		[object[]]$Arguments
	)

	if (-not (Test-Path -LiteralPath $env:PNPMEXE)) {
		throw ('pnpm wrapper not found: ' + $env:PNPMEXE)
	}

	& $env:PNPMEXE @Arguments
}

$version = & $nodeExe -v

Write-Host 'Enabled Node toolchain for this PowerShell session:'
Write-Host ('  NODE_BIN=' + $nodeBin)
Write-Host ('  LOCALAPPDATA=' + $env:LOCALAPPDATA)
Write-Host ('  APPDATA=' + $env:APPDATA)
Write-Host ('  NODEEXE=' + $nodeExe)
Write-Host ('  NODE_VERSION=' + $version)
Write-Host ('  COREPACK_HOME=' + $env:COREPACK_HOME)
Write-Host ('  COREPACK_CLI=' + $env:COREPACK_CLI)
Write-Host ('  PNPM_HOME=' + $env:PNPM_HOME)
Write-Host ('  PNPMEXE=' + $env:PNPMEXE)
Write-Host ('  PNPM_VERSION=' + $pnpmVersionText)
