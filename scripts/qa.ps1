<#
.SYNOPSIS
    EOMP Comprehensive QA/QC Verification Suite

.DESCRIPTION
    Runs complete automated validation across Frontend, Backend, Infrastructure, Databases, and CI pipeline.
#>

$ErrorActionPreference = "Continue"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path "$ProjectRoot\docker-compose.yml")) {
    $ProjectRoot = $PSScriptRoot
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   EOMP COMPREHENSIVE QA/QC SUITE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$qaResults = [ordered]@{}

# 1. FRONTEND
Write-Host "[1/5] Frontend QA (apps/web)..." -ForegroundColor Yellow
Push-Location "$ProjectRoot\apps\web"
try {
    pnpm typecheck 2>&1 | Out-Null
    $tcOk = ($LASTEXITCODE -eq 0)

    pnpm lint 2>&1 | Out-Null
    $lintOk = ($LASTEXITCODE -eq 0)

    pnpm build 2>&1 | Out-Null
    $buildOk = ($LASTEXITCODE -eq 0)

    if ($tcOk -and $lintOk -and $buildOk) {
        Write-Host "  [+] Typecheck, Lint, Build: PASS" -ForegroundColor Green
        $qaResults["Frontend"] = "PASS"
    } else {
        Write-Host "  [-] Frontend verification failed" -ForegroundColor Red
        $qaResults["Frontend"] = "FAIL"
    }
} finally {
    Pop-Location
}

# 2. BACKEND
Write-Host "`n[2/5] Backend Go Services QA (12 modules)..." -ForegroundColor Yellow
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
    $qaResults["Backend"] = "PASS"
} else {
    Write-Host "  [-] Backend verification failed" -ForegroundColor Red
    $qaResults["Backend"] = "FAIL"
}

# 3. INFRASTRUCTURE HEALTH
Write-Host "`n[3/5] Infrastructure Health QA..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
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

    if ($infraPass) {
        $qaResults["Infrastructure"] = "PASS"
    } else {
        $qaResults["Infrastructure"] = "FAIL"
    }
} finally {
    Pop-Location
}

# 4. DATABASE & STORAGE
Write-Host "`n[4/5] Databases & Storage QA..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    $dbs = docker compose exec -T postgres psql -U eomp -d eomp -t -c "SELECT datname FROM pg_database WHERE datname LIKE '%_db';"
    $expectedDbs = @("auth_db", "employee_db", "asset_db", "helpdesk_db", "workflow_db", "knowledge_db", "audit_db")
    $dbPass = $true
    foreach ($edb in $expectedDbs) {
        if ($dbs -match $edb) {
            Write-Host "  [+] Database $edb exists" -ForegroundColor Green
        } else {
            Write-Host "  [-] Database $edb missing" -ForegroundColor Red
            $dbPass = $false
        }
    }
    if ($dbPass) {
        $qaResults["Database"] = "PASS"
    } else {
        $qaResults["Database"] = "FAIL"
    }
} finally {
    Pop-Location
}

# 5. DOCKER & CI
Write-Host "`n[5/5] Docker & CI Configuration QA..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    docker compose config --quiet 2>&1 | Out-Null
    $composeOk = ($LASTEXITCODE -eq 0)
    $jenkinsOk = Test-Path "$ProjectRoot\Jenkinsfile"

    if ($composeOk -and $jenkinsOk) {
        Write-Host "  [+] docker-compose.yml syntax & Jenkinsfile: PASS" -ForegroundColor Green
        $qaResults["Docker_CI"] = "PASS"
    } else {
        Write-Host "  [-] Docker or Jenkins configuration error" -ForegroundColor Red
        $qaResults["Docker_CI"] = "FAIL"
    }
} finally {
    Pop-Location
}

# FINAL SUMMARY REPORT
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "           SETUP QA REPORT" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$overallPass = $true
foreach ($entry in $qaResults.GetEnumerator()) {
    $color = if ($entry.Value -eq "PASS") { "Green" } else { "Red" }
    if ($entry.Value -ne "PASS") { $overallPass = $false }
    Write-Host "  $($entry.Key.PadRight(18)) : " -NoNewline
    Write-Host "$($entry.Value)" -ForegroundColor $color
}

Write-Host ""
if ($overallPass) {
    Write-Host "  OVERALL STATUS     : PASS (Setup Acceptance Criteria 100% Met)" -ForegroundColor Green
} else {
    Write-Host "  OVERALL STATUS     : FAIL (Check logs above for details)" -ForegroundColor Red
}
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
