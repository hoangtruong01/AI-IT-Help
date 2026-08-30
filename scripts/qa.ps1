<#
.SYNOPSIS
    EOMP Comprehensive QA/QC Verification Suite (Phase 11 Master QA Engine)

.DESCRIPTION
    Runs complete automated validation across Frontend, Backend Services (12 Modules),
    Cross-Service E2E Lifecycle Integration Suite, Infrastructure Health Probes, Databases, and CI/CD Pipeline.
#>

$ErrorActionPreference = "Continue"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path "$ProjectRoot\docker-compose.yml")) {
    $ProjectRoot = $PSScriptRoot
}

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "       EOMP COMPREHENSIVE QA/QC AUTOMATION SUITE            " -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

$qaResults = [ordered]@{}

# 1. FRONTEND QA
Write-Host "[1/6] Frontend QA (apps/web)..." -ForegroundColor Yellow
Push-Location "$ProjectRoot\apps\web"
try {
    pnpm.cmd typecheck 2>&1 | Out-Null
    $tcOk = ($LASTEXITCODE -eq 0)

    pnpm.cmd test 2>&1 | Out-Null
    $testOk = ($LASTEXITCODE -eq 0)

    pnpm.cmd lint 2>&1 | Out-Null
    $lintOk = ($LASTEXITCODE -eq 0)

    pnpm.cmd build 2>&1 | Out-Null
    $buildOk = ($LASTEXITCODE -eq 0)

    if ($tcOk -and $testOk -and $lintOk -and $buildOk) {
        Write-Host "  [+] Unit Test, Typecheck, Lint, Production Build: PASS" -ForegroundColor Green
        $qaResults["Frontend"] = "PASS"
    } else {
        Write-Host "  [-] Frontend verification failed" -ForegroundColor Red
        $qaResults["Frontend"] = "FAIL"
    }
} finally {
    Pop-Location
}

# 2. BACKEND GO SERVICES QA (12 Modules)
Write-Host "`n[2/6] Backend Go Services QA (12 modules + coverage)..." -ForegroundColor Yellow
$backendPass = $true
$services = @("packages/shared", "services/gateway", "services/auth", "services/employee", "services/asset", "services/helpdesk", "services/workflow", "services/notification", "services/knowledge", "services/ai", "services/audit", "services/reporting")

foreach ($svc in $services) {
    Push-Location "$ProjectRoot\$svc"
    try {
        go vet ./... 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { $backendPass = $false }

        go test ./... 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) { $backendPass = $false }

        if ($svc -ne "packages/shared") {
            go build -o tmp_qa.exe ./cmd/server 2>&1 | Out-Null
            if ($LASTEXITCODE -ne 0) { $backendPass = $false }
            Remove-Item "tmp_qa.exe" -Force -ErrorAction SilentlyContinue
        }
    } finally {
        Pop-Location
    }
}

if ($backendPass) {
    Write-Host "  [+] go vet, go test, go build (all 12 modules): PASS" -ForegroundColor Green
    $qaResults["Backend_12_Services"] = "PASS"
} else {
    Write-Host "  [-] Backend verification failed" -ForegroundColor Red
    $qaResults["Backend_12_Services"] = "FAIL"
}

# 3. CROSS-SERVICE E2E LIFECYCLE QA
Write-Host "`n[3/6] Cross-Service E2E Lifecycle QA (tests/e2e)..." -ForegroundColor Yellow
Push-Location "$ProjectRoot\tests\e2e"
try {
    go test -v ./... 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [+] 7-Step Enterprise Cross-Service Lifecycle Flow: PASS" -ForegroundColor Green
        $qaResults["Cross_Service_E2E"] = "PASS"
    } else {
        Write-Host "  [-] Cross-service E2E tests failed" -ForegroundColor Red
        $qaResults["Cross_Service_E2E"] = "FAIL"
    }
} finally {
    Pop-Location
}

# 4. INFRASTRUCTURE HEALTH
Write-Host "`n[4/6] Infrastructure Health Probes QA..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    docker info 2>$null | Out-Null
    $infraAvailable = ($LASTEXITCODE -eq 0)
    if (-not $infraAvailable) {
        Write-Host "  [!] Docker daemon unavailable; infrastructure probes were not executed" -ForegroundColor Yellow
        $qaResults["Infrastructure_Probes"] = "SKIP"
    } else {
    $checks = @(
        @{ Name = "PostgreSQL"; Test = { docker compose exec -T postgres pg_isready -U eomp 2>$null } },
        @{ Name = "Redis";      Test = { $r = docker compose exec -T redis redis-cli ping 2>$null; $r -eq "PONG" } },
        @{ Name = "RabbitMQ";   Test = { docker compose exec -T rabbitmq rabbitmq-diagnostics -q ping 2>$null } },
        @{ Name = "MinIO";      Test = { $r = Invoke-WebRequest -Uri "http://localhost:9000/minio/health/live" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Qdrant";     Test = { $r = Invoke-WebRequest -Uri "http://localhost:6333/healthz" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Prometheus"; Test = { $r = Invoke-WebRequest -Uri "http://localhost:9090/-/healthy" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Grafana";    Test = { $r = Invoke-WebRequest -Uri "http://localhost:3002/api/health" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } },
        @{ Name = "Loki";       Test = { $r = Invoke-WebRequest -Uri "http://localhost:3100/loki/api/v1/status/buildinfo" -UseBasicParsing -ErrorAction SilentlyContinue; $r.StatusCode -eq 200 } }
    )

    $infraPass = $true
    foreach ($check in $checks) {
        $padded = $check.Name.PadRight(14)
        try {
            $result = & $check.Test
            if ($LASTEXITCODE -eq 0 -or $result) {
                Write-Host "  [+] $padded OK" -ForegroundColor Green
            } else {
                Write-Host "  [-] $padded FAIL" -ForegroundColor Red
                $infraPass = $false
            }
        } catch {
            Write-Host "  [-] $padded FAIL" -ForegroundColor Red
            $infraPass = $false
        }
    }
    $qaResults["Infrastructure_Probes"] = if ($infraPass) { "PASS" } else { "FAIL" }
    }
} finally {
    Pop-Location
}

# 5. DATABASE & STORAGE
Write-Host "`n[5/6] Databases & Migrations Schema QA..." -ForegroundColor Yellow
if (-not $infraAvailable) {
    $qaResults["Database_Schemas"] = "SKIP"
    Write-Host "  [!] Database schema checks skipped because Docker is unavailable" -ForegroundColor Yellow
} else {
    $databasePass = $true
    $databaseNames = @("auth_db", "employee_db", "asset_db", "helpdesk_db", "workflow_db", "notification_db", "knowledge_db", "audit_db", "reporting_db")
    foreach ($databaseName in $databaseNames) {
        docker compose exec -T postgres psql -U eomp -d $databaseName -Atc "SELECT COUNT(*) FROM schema_migrations;" 2>$null | Out-Null
        if ($LASTEXITCODE -ne 0) { $databasePass = $false }
    }
    $qaResults["Database_Schemas"] = if ($databasePass) { "PASS" } else { "FAIL" }
    Write-Host "  Database migration trackers: $($qaResults['Database_Schemas'])" -ForegroundColor $(if ($databasePass) { "Green" } else { "Red" })
}

# 6. DOCKER & CI/CD CONFIG
Write-Host "`n[6/6] Docker & CI/CD Pipeline Configuration QA..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    docker compose -f "$ProjectRoot\docker-compose.yml" config --quiet 2>$null
    $composeOk = ($LASTEXITCODE -eq 0)
    $jenkinsOk = Test-Path "$ProjectRoot\Jenkinsfile"
    $k6Ok      = Test-Path "$ProjectRoot\infrastructure\k6\load_test.js"

    if ($composeOk -and $jenkinsOk -and $k6Ok) {
        Write-Host "  [+] docker-compose.yml, Jenkinsfile, K6 Load Testing Engine: PASS" -ForegroundColor Green
        $qaResults["Docker_CI_CD"] = "PASS"
    } else {
        Write-Host "  [-] Docker, Jenkins or K6 configuration error" -ForegroundColor Red
        $qaResults["Docker_CI_CD"] = "FAIL"
    }
} finally {
    Pop-Location
}

# FINAL SUMMARY REPORT
Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "               FINAL MASTER QA SUITE REPORT                 " -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

$overallPass = $true
$hasSkips = $false
foreach ($entry in $qaResults.GetEnumerator()) {
    $color = if ($entry.Value -eq "PASS") { "Green" } elseif ($entry.Value -eq "SKIP") { "Yellow" } else { "Red" }
    if ($entry.Value -eq "FAIL") { $overallPass = $false }
    if ($entry.Value -eq "SKIP") { $hasSkips = $true }
    Write-Host "  $($entry.Key.PadRight(24)) : " -NoNewline
    Write-Host "$($entry.Value)" -ForegroundColor $color
}

Write-Host ""
if ($overallPass -and -not $hasSkips) {
    Write-Host "  OVERALL STATUS           : PASS" -ForegroundColor Green
} elseif ($overallPass) {
    Write-Host "  OVERALL STATUS           : PASS WITH SKIPPED CHECKS (not production-certified)" -ForegroundColor Yellow
} else {
    Write-Host "  OVERALL STATUS           : FAIL (Check logs above for details)" -ForegroundColor Red
}
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""
if (-not $overallPass) { exit 1 }
