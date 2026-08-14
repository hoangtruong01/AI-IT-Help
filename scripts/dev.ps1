<#
.SYNOPSIS
    EOMP Development Helper - PowerShell equivalent of Makefile for Windows.

.DESCRIPTION
    Provides common development commands for the EOMP platform.

.EXAMPLE
    .\scripts\dev.ps1 help
    .\scripts\dev.ps1 docker-up
    .\scripts\dev.ps1 health
#>

param(
    [Parameter(Position = 0)]
    [ValidateSet(
        "help", "dev", "build", "test", "lint", "format",
        "docker-up", "docker-down", "docker-reset",
        "logs", "health", "clean"
    )]
    [string]$Command = "help",

    [Parameter(Position = 1)]
    [string]$Service = ""
)

$ErrorActionPreference = "Stop"

# Resolve project root (scripts/ is one level below root)
$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path "$ProjectRoot\docker-compose.yml")) {
    $ProjectRoot = $PSScriptRoot
}

function Show-Help {
    Write-Host ""
    Write-Host "EOMP - Enterprise Operations Management Platform" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\scripts\dev.ps1 [command] [service]" -ForegroundColor White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Yellow
    Write-Host "  help          Show this help message"
    Write-Host "  dev           Start frontend development server"
    Write-Host "  build         Build all Go services and frontend"
    Write-Host "  test          Run all tests"
    Write-Host "  lint          Run linters"
    Write-Host "  format        Format Go code"
    Write-Host "  docker-up     Start infrastructure services"
    Write-Host "  docker-down   Stop infrastructure services"
    Write-Host "  docker-reset  Stop and remove all volumes"
    Write-Host "  logs          Show container logs (optionally filter by service)"
    Write-Host "  health        Check health of all infrastructure"
    Write-Host "  clean         Clean build artifacts"
    Write-Host ""
}

function Start-Dev {
    Write-Host "Starting frontend development server..." -ForegroundColor Cyan
    Push-Location "$ProjectRoot\apps\web"
    try { pnpm dev } finally { Pop-Location }
}

function Start-Build {
    Write-Host "Building Go services..." -ForegroundColor Cyan
    Get-ChildItem "$ProjectRoot\services" -Directory | ForEach-Object {
        $goMod = Join-Path $_.FullName "go.mod"
        if (Test-Path $goMod) {
            Write-Host "  Building $($_.Name)..." -ForegroundColor Gray
            Push-Location $_.FullName
            try { go build ./cmd/... } finally { Pop-Location }
        }
    }
    Write-Host "Building frontend..." -ForegroundColor Cyan
    Push-Location "$ProjectRoot\apps\web"
    try { pnpm build } finally { Pop-Location }
    Write-Host "Build complete." -ForegroundColor Green
}

function Start-Test {
    Write-Host "Running Go tests..." -ForegroundColor Cyan
    Get-ChildItem "$ProjectRoot\services" -Directory | ForEach-Object {
        $goMod = Join-Path $_.FullName "go.mod"
        if (Test-Path $goMod) {
            Write-Host "  Testing $($_.Name)..." -ForegroundColor Gray
            Push-Location $_.FullName
            try { go test ./... } finally { Pop-Location }
        }
    }
    Write-Host "All tests passed." -ForegroundColor Green
}

function Start-Lint {
    Write-Host "Linting Go services..." -ForegroundColor Cyan
    Get-ChildItem "$ProjectRoot\services" -Directory | ForEach-Object {
        $goMod = Join-Path $_.FullName "go.mod"
        if (Test-Path $goMod) {
            Write-Host "  Linting $($_.Name)..." -ForegroundColor Gray
            Push-Location $_.FullName
            try { go vet ./... } finally { Pop-Location }
        }
    }
    Write-Host "Linting frontend..." -ForegroundColor Cyan
    Push-Location "$ProjectRoot\apps\web"
    try { pnpm lint } finally { Pop-Location }
}

function Start-Format {
    Write-Host "Formatting Go code..." -ForegroundColor Cyan
    Get-ChildItem "$ProjectRoot\services" -Directory | ForEach-Object {
        $goMod = Join-Path $_.FullName "go.mod"
        if (Test-Path $goMod) {
            $goFiles = Get-ChildItem $_.FullName -Recurse -Filter "*.go"
            if ($goFiles) {
                gofmt -w $_.FullName
            }
        }
    }
    Write-Host "Format complete." -ForegroundColor Green
}

function Start-DockerUp {
    Write-Host "Starting infrastructure services..." -ForegroundColor Cyan
    Push-Location $ProjectRoot
    try { docker compose up -d } finally { Pop-Location }
    Write-Host "Infrastructure started." -ForegroundColor Green
}

function Start-DockerDown {
    Write-Host "Stopping infrastructure services..." -ForegroundColor Cyan
    Push-Location $ProjectRoot
    try { docker compose down } finally { Pop-Location }
    Write-Host "Infrastructure stopped." -ForegroundColor Green
}

function Start-DockerReset {
    Write-Host "Stopping and removing all volumes..." -ForegroundColor Yellow
    Push-Location $ProjectRoot
    try { docker compose down -v } finally { Pop-Location }
    Write-Host "Reset complete." -ForegroundColor Green
}

function Show-Logs {
    Push-Location $ProjectRoot
    try {
        if ($Service) {
            docker compose logs -f $Service
        } else {
            docker compose logs -f
        }
    } finally { Pop-Location }
}

function Test-ServiceHealth {
    Write-Host ""
    Write-Host "=== EOMP Infrastructure Health ===" -ForegroundColor Cyan
    Write-Host ""

    $checks = @(
        @{ Name = "PostgreSQL"; Test = { docker compose exec -T postgres pg_isready -U eomp 2>$null } },
        @{ Name = "Redis";      Test = { $r = docker compose exec -T redis redis-cli ping 2>$null; $r -eq "PONG" } },
        @{ Name = "RabbitMQ";   Test = { docker compose exec -T rabbitmq rabbitmq-diagnostics -q ping 2>$null } },
        @{ Name = "MinIO";      Test = { $r = Invoke-WebRequest -Uri "http://localhost:9000/minio/health/live" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Qdrant";     Test = { $r = Invoke-WebRequest -Uri "http://localhost:6333/healthz" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Prometheus"; Test = { $r = Invoke-WebRequest -Uri "http://localhost:9090/-/healthy" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Grafana";    Test = { $r = Invoke-WebRequest -Uri "http://localhost:3001/api/health" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Loki";       Test = { $r = Invoke-WebRequest -Uri "http://localhost:3100/ready" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } }
    )

    Push-Location $ProjectRoot
    try {
        foreach ($check in $checks) {
            $padded = $check.Name.PadRight(14)
            try {
                $result = & $check.Test
                if ($LASTEXITCODE -eq 0 -or $result) {
                    Write-Host "  $padded" -NoNewline
                    Write-Host "OK" -ForegroundColor Green
                } else {
                    Write-Host "  $padded" -NoNewline
                    Write-Host "FAIL" -ForegroundColor Red
                }
            } catch {
                Write-Host "  $padded" -NoNewline
                Write-Host "FAIL" -ForegroundColor Red
            }
        }
    } finally { Pop-Location }
    Write-Host ""
}

function Start-Clean {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan

    Get-ChildItem "$ProjectRoot\services" -Recurse -Include "*.exe", "*.test" -ErrorAction SilentlyContinue | Remove-Item -Force

    $webDirs = @(".nuxt", ".output", "dist", "node_modules")
    foreach ($dir in $webDirs) {
        $path = Join-Path "$ProjectRoot\apps\web" $dir
        if (Test-Path $path) {
            Remove-Item -Recurse -Force $path
        }
    }

    Write-Host "Clean complete." -ForegroundColor Green
}

# ==============================
# Command Router
# ==============================
switch ($Command) {
    "help"         { Show-Help }
    "dev"          { Start-Dev }
    "build"        { Start-Build }
    "test"         { Start-Test }
    "lint"         { Start-Lint }
    "format"       { Start-Format }
    "docker-up"    { Start-DockerUp }
    "docker-down"  { Start-DockerDown }
    "docker-reset" { Start-DockerReset }
    "logs"         { Show-Logs }
    "health"       { Test-ServiceHealth }
    "clean"        { Start-Clean }
    default        { Show-Help }
}
