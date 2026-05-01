# Octopus 项目文件变更监测脚本
# 使用方法: .\scripts\watch-changes.ps1
# 功能: 监控项目文件变更并自动触发审查检查

param(
    [string]$ProjectRoot = (Split-Path $PSScriptRoot -Parent),
    [string[]]$WatchPatterns = @(
        "*.go", "*.ts", "*.tsx", "*.js", "*.jsx",
        "*.json", "*.yaml", "*.yml", "*.md",
        "*.css", "*.sh", "*.ps1", "*.py"
    ),
    [int]$DebounceMs = 2000
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Octopus 项目变更监测器" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "监控目录: $ProjectRoot" -ForegroundColor Yellow
Write-Host "监控模式: $($WatchPatterns -join ', ')" -ForegroundColor Yellow
Write-Host "防抖时间: ${DebounceMs}ms" -ForegroundColor Yellow
Write-Host ""
Write-Host "按 Ctrl+C 停止监测" -ForegroundColor Yellow
Write-Host ""

# 排除不需要监控的目录
$ExcludeDirs = @(".git", ".next", ".tmp-gocache", "node_modules", "build", "out", "data")

# 创建文件系统监听器
$watcher = New-Object System.IO.FileSystemWatcher
$watcher.Path = $ProjectRoot
$watcher.Filter = "*.*"
$watcher.IncludeSubdirectories = $true
$watcher.NotifyFilter = [System.IO.NotifyFilters]::LastWrite -bor 
                        [System.IO.NotifyFilters]::FileName -bor 
                        [System.IO.NotifyFilters]::DirectoryName

# 注册事件处理
$onChange = Register-ObjectEvent -InputObject $watcher -EventName Changed -Action {
    $path = $Event.SourceEventArgs.FullPath
    $changeType = $Event.SourceEventArgs.ChangeType
    
    # 检查是否在排除目录中
    $shouldExclude = $false
    foreach ($exclude in $using:ExcludeDirs) {
        if ($path -like "*\$exclude\*" -or $path -like "*\$exclude/*") {
            $shouldExclude = $true
            break
        }
    }
    
    # 检查文件扩展名是否在监控范围内
    $ext = [System.IO.Path]::GetExtension($path)
    $shouldWatch = $false
    foreach ($pattern in $using:WatchPatterns) {
        $patternExt = $pattern.Substring($pattern.IndexOf('.'))
        if ($ext -eq $patternExt) {
            $shouldWatch = $true
            break
        }
    }
    
    if (-not $shouldExclude -and $shouldWatch) {
        $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
        $relativePath = $path.Replace($using:ProjectRoot, "").TrimStart('\', '/')
        
        Write-Host "[$timestamp] 变更检测: $changeType - $relativePath" -ForegroundColor Green
        
        # 根据文件类型触发不同的检查
        if ($ext -eq ".go") {
            Write-Host "  → Go 文件变更，触发检查..." -ForegroundColor Cyan
            Push-Location $using:ProjectRoot
            try {
                # 运行 go vet (不阻塞监控)
                Start-Process -FilePath "go" -ArgumentList "vet", $relativePath -NoNewWindow -Wait -RedirectStandardError "null" 2>$null
            } catch {
                Write-Host "  → go vet 检查完成" -ForegroundColor Gray
            }
            Pop-Location
        }
        elseif ($ext -match "\.(ts|tsx)$") {
            Write-Host "  → TypeScript 文件变更" -ForegroundColor Cyan
        }
        elseif ($ext -eq ".md") {
            Write-Host "  → Markdown 文档变更" -ForegroundColor Cyan
        }
    }
}

$onCreated = Register-ObjectEvent -InputObject $watcher -EventName Created -Action {
    $path = $Event.SourceEventArgs.FullPath
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $relativePath = $path.Replace($using:ProjectRoot, "").TrimStart('\', '/')
    Write-Host "[$timestamp] 新建文件: $relativePath" -ForegroundColor Magenta
}

$onDeleted = Register-ObjectEvent -InputObject $watcher -EventName Deleted -Action {
    $path = $Event.SourceEventArgs.FullPath
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $relativePath = $path.Replace($using:ProjectRoot, "").TrimStart('\', '/')
    Write-Host "[$timestamp] 删除文件: $relativePath" -ForegroundColor Red
}

$onRenamed = Register-ObjectEvent -InputObject $watcher -EventName Renamed -Action {
    $oldPath = $Event.SourceEventArgs.OldFullPath.Replace($using:ProjectRoot, "").TrimStart('\', '/')
    $newPath = $Event.SourceEventArgs.FullPath.Replace($using:ProjectRoot, "").TrimStart('\', '/')
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] 重命名: $oldPath → $newPath" -ForegroundColor Yellow
}

# 保持脚本运行
try {
    while ($true) {
        Start-Sleep -Milliseconds 100
    }
}
finally {
    # 清理资源
    Unregister-Event -SourceIdentifier $onChange.Name
    Unregister-Event -SourceIdentifier $onCreated.Name
    Unregister-Event -SourceIdentifier $onDeleted.Name
    Unregister-Event -SourceIdentifier $onRenamed.Name
    $watcher.Dispose()
    Write-Host ""
    Write-Host "变更监测已停止" -ForegroundColor Yellow
}
