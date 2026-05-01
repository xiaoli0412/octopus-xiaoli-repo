[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)

    Write-Host ""
    Write-Host ("== " + $Message + " ==")
}

function Format-MaskedSecret {
    param([AllowEmptyString()][string]$Value)

    if ([string]::IsNullOrEmpty($Value)) {
        return '<empty>'
    }

    $visible = [Math]::Min(4, $Value.Length)
    return ('***' + $Value.Substring($Value.Length - $visible, $visible))
}

function Repair-ProcessPathKey {
    $processEnv = [Environment]::GetEnvironmentVariables('Process')
    if ($processEnv.Contains('PATH')) {
        $pathValue = [Environment]::GetEnvironmentVariable('Path', 'Process')
        Remove-Item Env:PATH -ErrorAction SilentlyContinue
        if (-not [string]::IsNullOrWhiteSpace($pathValue)) {
            $env:Path = $pathValue
        }
    }
}

function Get-RunnablePythonCommand {
    $candidates = @(
        @('C:\Program Files\Python312\python.exe'),
        @('python.exe'),
        @('py.exe', '-3')
    )

    foreach ($candidate in $candidates) {
        $command = $candidate[0]
        $args = if ($candidate.Length -gt 1) { $candidate[1..($candidate.Length - 1)] } else { @() }

        if ($command -like '*\\*' -and -not (Test-Path -LiteralPath $command)) {
            continue
        }

        try {
            & $command @args --version *> $null
            if ($LASTEXITCODE -eq 0) {
                return ,$candidate
            }
        }
        catch {
        }
    }

    throw 'Python 3 is required to run scripts\smoke-win-backend.ps1'
}

function Resolve-SmokeStaticDir {
    if (-not [string]::IsNullOrWhiteSpace($env:OCTOPUS_SMOKE_STATIC_DIR)) {
        return $env:OCTOPUS_SMOKE_STATIC_DIR
    }

    $syncedDir = Join-Path $repoRoot 'static\out'
    $sourceDir = Join-Path $repoRoot 'web\out'
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
$repoToolsRoot = Join-Path $repoRoot '.tools'
$env:GOCACHE = Resolve-WritableWorkspaceDir -CurrentValue $env:GOCACHE -FallbackPath (Join-Path $repoToolsRoot 'gocache')
$env:GOTMPDIR = Resolve-WritableWorkspaceDir -CurrentValue $env:GOTMPDIR -FallbackPath (Join-Path $repoToolsRoot 'gotmp')
$env:TEMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TEMP -FallbackPath (Join-Path $repoToolsRoot 'tmp')
$env:TMP = Resolve-WritableWorkspaceDir -CurrentValue $env:TMP -FallbackPath $env:TEMP
Repair-ProcessPathKey
$exePath = Join-Path $repoRoot 'build\octopus-smoke.exe'
$goEnvReady = $false
try {
    . (Join-Path $scriptDir 'use-go-env.ps1')
    $goEnvReady = $true
}
catch {
    if (-not (Test-Path -LiteralPath $exePath)) {
        throw
    }

    Write-Host ('Go toolchain unavailable for this smoke run; reusing existing smoke binary: ' + $_.Exception.Message)
}

$smokePort = 18084 + (Get-Random -Minimum 0 -Maximum 200)
$mockPort = 19091 + (Get-Random -Minimum 0 -Maximum 200)
$smokeStaticDir = Resolve-SmokeStaticDir
$smokeAdminUsername = if ($env:OCTOPUS_SMOKE_ADMIN_USERNAME) { $env:OCTOPUS_SMOKE_ADMIN_USERNAME } else { 'admin' }
$smokeAdminPassword = if ($env:OCTOPUS_SMOKE_ADMIN_PASSWORD) { $env:OCTOPUS_SMOKE_ADMIN_PASSWORD } else { 'admin' }

$pythonCommand = Get-RunnablePythonCommand

Push-Location $repoRoot
try {
    Write-Step -Message 'Building temporary smoke binary'
    if ($goEnvReady) {
        & $env:GOEXE build -o $exePath .
        if ($LASTEXITCODE -ne 0) {
            throw ('Local Go build failed with exit code {0}; refusing to reuse existing smoke binary at {1}.' -f $LASTEXITCODE, $exePath)
        }
    }
    else {
        Write-Host 'Skipping local Go build because this session has no runnable Go toolchain.'
    }

    if (-not $goEnvReady -and -not (Test-Path -LiteralPath $exePath)) {
        throw ('Smoke binary build failed and no reusable binary was found: ' + $exePath)
    }

    $tempDir = Join-Path $env:TEMP ('octopus-e2e-' + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tempDir | Out-Null

    $configPath = Join-Path $tempDir 'config.json'
    $dbPath = Join-Path $tempDir 'octopus.db'
    $serverOut = Join-Path $tempDir 'server.stdout.log'
    $serverErr = Join-Path $tempDir 'server.stderr.log'
    $mockScript = Join-Path $tempDir 'mock_upstream.py'
    $mockOut = Join-Path $tempDir 'mock.stdout.log'
    $mockErr = Join-Path $tempDir 'mock.stderr.log'

    $configJson = @"
{
  "server": {
    "host": "127.0.0.1",
    "port": $smokePort,
    "static_dir": "$($smokeStaticDir.Replace('\', '/'))"
  },
  "database": {
    "type": "sqlite",
    "path": "$($dbPath.Replace('\', '/'))"
  },
  "log": {
    "level": "info"
  }
}
"@
    [System.IO.File]::WriteAllText($configPath, $configJson, (New-Object System.Text.UTF8Encoding($true)))

    $mockPy = @"
import json
from http.server import BaseHTTPRequestHandler, HTTPServer

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length) if length else b"{}"
        try:
            payload = json.loads(body.decode("utf-8"))
        except Exception:
            payload = {}
        response = {
            "id": "chatcmpl-mock-1",
            "object": "chat.completion",
            "created": 1713436800,
            "model": payload.get("model") or "gpt-4o-mini",
            "choices": [{"index": 0, "message": {"role": "assistant", "content": "mock-ok"}, "finish_reason": "stop"}],
            "usage": {"prompt_tokens": 5, "completion_tokens": 2, "total_tokens": 7}
        }
        encoded = json.dumps(response).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(encoded)))
        self.end_headers()
        self.wfile.write(encoded)

    def log_message(self, format, *args):
        return

HTTPServer(("127.0.0.1", $mockPort), Handler).serve_forever()
"@
    [System.IO.File]::WriteAllText($mockScript, $mockPy, (New-Object System.Text.UTF8Encoding($false)))

    $pythonArgs = @()
    if ($pythonCommand.Length -gt 1) {
        $pythonArgs += $pythonCommand[1..($pythonCommand.Length - 1)]
    }
    $pythonArgs += $mockScript

    $mockProc = Start-Process -FilePath $pythonCommand[0] -ArgumentList $pythonArgs -RedirectStandardOutput $mockOut -RedirectStandardError $mockErr -PassThru
    $serverProc = $null
    $serverStdoutTask = $null
    $serverStderrTask = $null

    try {
        Write-Step -Message 'Starting mock upstream and Octopus server'
        $mockReady = $false
        for ($i = 0; $i -lt 40; $i++) {
            Start-Sleep -Milliseconds 250
            try {
                & curl.exe --silent --show-error --max-time 2 "http://127.0.0.1:$mockPort/" *> $null
            }
            catch {
            }
            if ($LASTEXITCODE -eq 0) {
                $mockReady = $true
                break
            }
            if ($mockProc.HasExited) {
                throw 'mock upstream exited before becoming ready'
            }
        }
        if (-not $mockReady) {
            throw 'mock upstream did not become ready'
        }
        $serverStartInfo = [System.Diagnostics.ProcessStartInfo]::new()
        $serverStartInfo.FileName = $exePath
        $serverStartInfo.WorkingDirectory = $repoRoot
        $serverStartInfo.RedirectStandardOutput = $true
        $serverStartInfo.RedirectStandardError = $true
        $serverStartInfo.UseShellExecute = $false
        [void]$serverStartInfo.Arguments
        $serverStartInfo.Arguments = ('start --config "{0}"' -f $configPath)
        $serverStartInfo.Environment['OCTOPUS_ADMIN_USERNAME'] = $smokeAdminUsername
        $serverStartInfo.Environment['OCTOPUS_ADMIN_PASSWORD'] = $smokeAdminPassword
        $serverProc = [System.Diagnostics.Process]::Start($serverStartInfo)
        $serverStdoutTask = [System.Threading.Tasks.Task]::Run([System.Action]{
            $text = $serverProc.StandardOutput.ReadToEnd()
            [System.IO.File]::WriteAllText($serverOut, $text, [System.Text.Encoding]::UTF8)
        })
        $serverStderrTask = [System.Threading.Tasks.Task]::Run([System.Action]{
            $text = $serverProc.StandardError.ReadToEnd()
            [System.IO.File]::WriteAllText($serverErr, $text, [System.Text.Encoding]::UTF8)
        })

        $base = "http://127.0.0.1:$smokePort"
        $health = $null
        for ($i = 0; $i -lt 80; $i++) {
            Start-Sleep -Milliseconds 500
            $health = & curl.exe --silent --show-error --max-time 5 "$base/healthz"
            if ($LASTEXITCODE -eq 0 -and $health) {
                break
            }
        }
        if (-not $health) {
            throw 'healthz did not respond successfully'
        }

        Write-Step -Message 'Verifying frontend shell and static assets'
        $rootHtml = & curl.exe --silent --show-error --max-time 10 "$base/"
        $rootHtmlText = ($rootHtml | Out-String)
        if ($LASTEXITCODE -ne 0 -or -not $rootHtmlText -or -not $rootHtmlText.Contains('<title>Octopus</title>')) {
            throw 'frontend shell did not render the expected Octopus title'
        }
        $manifest = Invoke-RestMethod -Uri "$base/manifest.json" -Method Get -TimeoutSec 10
        if (-not $manifest -or $manifest.name -ne 'Octopus') {
            throw "manifest name = $($manifest.name), want Octopus"
        }

        Write-Step -Message 'Driving minimal management and gateway smoke flow'
        $loginBody = @{ username = $smokeAdminUsername; password = $smokeAdminPassword; expire = 86400 } | ConvertTo-Json -Compress
        $loginResp = Invoke-RestMethod -Uri "$base/api/v1/user/login" -Method Post -ContentType 'application/json' -Body $loginBody -TimeoutSec 5
        $jwt = $loginResp.data.token
        $authHeaders = @{ Authorization = "Bearer $jwt" }

        $channelBody = @{
            name = 'mock-openai-demo'
            type = 0
            enabled = $true
            base_urls = @(@{ url = "http://127.0.0.1:$mockPort"; delay = 0 })
            keys = @(@{ enabled = $true; channel_key = 'mock-upstream-key' })
            model = 'gpt-4o-mini'
        } | ConvertTo-Json -Depth 8 -Compress
        $channelResp = Invoke-RestMethod -Uri "$base/api/v1/channel/create" -Method Post -Headers $authHeaders -ContentType 'application/json' -Body $channelBody -TimeoutSec 10

        $groupBody = @{
            name = 'gpt-4o-mini'
            mode = 1
            items = @(@{ channel_id = $channelResp.data.id; model_name = 'gpt-4o-mini'; priority = 1; weight = 1 })
        } | ConvertTo-Json -Depth 8 -Compress
        $groupResp = Invoke-RestMethod -Uri "$base/api/v1/group/create" -Method Post -Headers $authHeaders -ContentType 'application/json' -Body $groupBody -TimeoutSec 10

        $apiKeyBody = @{ name = 'smoke-key'; enabled = $true } | ConvertTo-Json -Compress
        $apiKeyResp = Invoke-RestMethod -Uri "$base/api/v1/apikey/create" -Method Post -Headers $authHeaders -ContentType 'application/json' -Body $apiKeyBody -TimeoutSec 10

        $chatHeaders = @{ Authorization = "Bearer $($apiKeyResp.data.api_key)" }
        $chatBody = @{ model = 'gpt-4o-mini'; messages = @(@{ role = 'user'; content = 'hello' }) } | ConvertTo-Json -Depth 8 -Compress
        $chatResp = Invoke-RestMethod -Uri "$base/v1/chat/completions" -Method Post -Headers $chatHeaders -ContentType 'application/json' -Body $chatBody -TimeoutSec 10

        Write-Host "HEALTH=$health"
        Write-Host ("SMOKE_PORT=" + $smokePort)
        Write-Host ("MOCK_PORT=" + $mockPort)
        Write-Host 'FRONTEND_TITLE=Octopus'
        Write-Host ('MANIFEST_NAME=' + $manifest.name)
        Write-Host ('STATIC_DIR=' + $smokeStaticDir)
        Write-Host ('GOCACHE=' + $env:GOCACHE)
        Write-Host ('GOTMPDIR=' + $env:GOTMPDIR)
        Write-Host ('TEMP=' + $env:TEMP)
        Write-Host ('TMP=' + $env:TMP)
        Write-Host "CHANNEL_ID=$($channelResp.data.id)"
        Write-Host "GROUP_ID=$($groupResp.data.id)"
        Write-Host ('GATEWAY_KEY_MASKED=' + (Format-MaskedSecret -Value $apiKeyResp.data.api_key))
        Write-Host ('CHAT_RESPONSE=' + ($chatResp | ConvertTo-Json -Depth 8 -Compress))
        Write-Host ('TEMP_DIR=' + $tempDir)
    }
    finally {
        if ($serverProc -and -not $serverProc.HasExited) {
            Stop-Process -Id $serverProc.Id -Force
        }
        if ($serverProc) {
            $serverProc.WaitForExit()
        }
        if ($mockProc -and -not $mockProc.HasExited) {
            Stop-Process -Id $mockProc.Id -Force
        }
        if ($serverStdoutTask) {
            try {
                $serverStdoutTask.Wait()
            }
            catch {
            }
        }
        if ($serverStderrTask) {
            try {
                $serverStderrTask.Wait()
            }
            catch {
            }
        }
    }
}
finally {
    Pop-Location
}
