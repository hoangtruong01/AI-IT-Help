<#
.SYNOPSIS
    EOMP Multi-Database Backup & Disaster Recovery CLI.

.DESCRIPTION
    Performs automated compressed SQL backups and integrity restore testing
    for all 9 EOMP PostgreSQL databases.

.EXAMPLE
    .\scripts\backup_restore.ps1 backup
    .\scripts\backup_restore.ps1 list
    .\scripts\backup_restore.ps1 test-restore
#>

param(
    [Parameter(Position = 0)]
    [ValidateSet("help", "backup", "list", "test-restore")]
    [string]$Command = "help"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BackupDir = "$ProjectRoot\backups"

if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null
}

$Databases = @(
    "auth_db", "employee_db", "asset_db", "helpdesk_db",
    "workflow_db", "notification_db", "knowledge_db", "reporting_db", "audit_db"
)

function Show-Help {
    Write-Host ""
    Write-Host "================================================================" -ForegroundColor Green
    Write-Host "  EOMP - Phase 14 Database Backup & Disaster Recovery CLI       " -ForegroundColor Green
    Write-Host "================================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage: .\scripts\backup_restore.ps1 [command]" -ForegroundColor White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Yellow
    Write-Host "  backup        Perform full backup of all 9 PostgreSQL databases" -ForegroundColor White
    Write-Host "  list          List all archived backup snapshots" -ForegroundColor White
    Write-Host "  test-restore  Validate integrity of latest backup snapshot" -ForegroundColor White
    Write-Host ""
}

function Invoke-Backup {
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    Write-Host "=== Starting Full Backup of 9 EOMP Databases [$timestamp] ===" -ForegroundColor Cyan

    $containerName = "eomp-postgres"
    docker inspect $containerName 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) {
        $containerName = "eomp-prod-postgres"
    }
    docker inspect $containerName 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "PostgreSQL container is not running; backup was not created."
    }

    foreach ($db in $Databases) {
        $outFile = "$BackupDir\${db}_${timestamp}.sql"
        Write-Host "  [BACKUP] Dumping database $db..." -ForegroundColor White
        
        docker exec $containerName pg_dump -U eomp -d $db -F p | Out-File -FilePath $outFile -Encoding utf8
        if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $outFile) -or (Get-Item -LiteralPath $outFile).Length -lt 100) {
            Remove-Item -LiteralPath $outFile -Force -ErrorAction SilentlyContinue
            throw "pg_dump failed or produced an invalid artifact for $db"
        }
        Write-Host "    -> Saved $outFile ($((Get-Item -LiteralPath $outFile).Length) bytes)" -ForegroundColor DarkGreen
    }

    Write-Host "`nAll 9 databases backed up successfully to $BackupDir" -ForegroundColor Green
}

function Invoke-List {
    Write-Host "=== Existing Database Backup Archives in $BackupDir ===" -ForegroundColor Cyan
    $files = Get-ChildItem -Path $BackupDir -Filter "*.sql*"
    if ($files.Count -eq 0) {
        Write-Host "No backup files found." -ForegroundColor Yellow
    } else {
        foreach ($f in $files) {
            Write-Host "  $($f.Name) ($($f.Length) bytes, $($f.LastWriteTime))" -ForegroundColor White
        }
    }
}

function Invoke-TestRestore {
    Write-Host "=== Running Integrity & Disaster Recovery Verification Test ===" -ForegroundColor Cyan
    $containerName = "eomp-postgres"
    docker inspect $containerName 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) { $containerName = "eomp-prod-postgres" }
    docker inspect $containerName 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "PostgreSQL container is not running; restore drill cannot run." }

    $latestFiles = @()
    foreach ($db in $Databases) {
        $file = Get-ChildItem -LiteralPath $BackupDir -Filter "${db}_*.sql" |
            Where-Object { $_.Length -ge 100 } |
            Sort-Object LastWriteTimeUtc -Descending |
            Select-Object -First 1
        if ($null -eq $file) { throw "No non-empty backup found for $db" }
        $latestFiles += $file
    }

    $startedAt = Get-Date
    $index = 0
    foreach ($file in $latestFiles) {
        $index++
        $restoreDb = "eomp_restore_verify_$([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())_$index"
        try {
            docker exec $containerName createdb -U eomp $restoreDb
            if ($LASTEXITCODE -ne 0) { throw "createdb failed for $restoreDb" }
            Get-Content -LiteralPath $file.FullName -Raw | docker exec -i $containerName psql -v ON_ERROR_STOP=1 -U eomp -d $restoreDb | Out-Null
            if ($LASTEXITCODE -ne 0) { throw "restore failed for $($file.Name)" }
            $tableCount = docker exec $containerName psql -U eomp -d $restoreDb -Atc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';"
            if ($LASTEXITCODE -ne 0 -or [int]$tableCount -le 0) { throw "restored database has no public tables: $($file.Name)" }
        } finally {
            docker exec $containerName dropdb -U eomp --if-exists $restoreDb 2>$null | Out-Null
        }
    }

    $duration = (Get-Date) - $startedAt
    $oldestBackup = $latestFiles | Sort-Object LastWriteTimeUtc | Select-Object -First 1
    $rpo = (Get-Date).ToUniversalTime() - $oldestBackup.LastWriteTimeUtc
    $evidence = [ordered]@{
        backup_created_at = $oldestBackup.LastWriteTimeUtc.ToString("o")
        restore_duration_seconds = [Math]::Round($duration.TotalSeconds, 3)
        database_count = $latestFiles.Count
        verified_at = (Get-Date).ToUniversalTime().ToString("o")
    }
    $evidence | ConvertTo-Json | Set-Content -LiteralPath "$BackupDir\dr_evidence.json" -Encoding utf8
    Write-Host "[VERIFIED] Real restore drill completed for $($latestFiles.Count) databases. RPO=$([Math]::Round($rpo.TotalSeconds, 1))s, RTO=$([Math]::Round($duration.TotalSeconds, 1))s" -ForegroundColor Green
}

switch ($Command) {
    "help"         { Show-Help }
    "backup"       { Invoke-Backup }
    "list"         { Invoke-List }
    "test-restore" { Invoke-TestRestore }
}
