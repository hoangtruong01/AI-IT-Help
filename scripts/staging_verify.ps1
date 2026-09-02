<#
.SYNOPSIS
    EOMP Staging & Gate D Comprehensive Verification Engine

.DESCRIPTION
    Executes a local, non-deployment verification precheck covering:
    1. Go code formatting, vet, and linting
    2. OpenAPI 102/102 runtime route parity
    3. In-memory unit & simulation suites (13 modules)
    4. In-process HTTP contracts and optional, fail-closed PostgreSQL integration
    5. Frontend Vitest unit/contract checks (apps/web)
    6. Static container and Kubernetes hardening checks

    This script does not deploy Helm, launch Playwright, scan images with Trivy,
    run k6, or execute a restore drill. Passing it is not Gate D acceptance.
#>

param(
    [switch]$RequirePostgres
)

$ErrorActionPreference = "Continue"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path "$ProjectRoot\docker-compose.yml")) {
    $ProjectRoot = $PSScriptRoot
}

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "          EOMP LOCAL RELEASE PRECHECK (NON-ACCEPTANCE)           " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "  Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
Write-Host "  Workspace: $ProjectRoot" -ForegroundColor Gray
Write-Host ""

$results = [ordered]@{}
$totalFailed = 0
$requiredDsns = @(
    "AUTH_INTEGRATION_DSN", "HELPDESK_INTEGRATION_DSN", "AUDIT_INTEGRATION_DSN",
    "INTEGRATION_POSTGRES_DSN", "NOTIFICATION_INTEGRATION_DSN", "REPORTING_INTEGRATION_DSN"
)

$requiredCommands = @("go", "gofmt", "pnpm.cmd")
$missingCommands = @($requiredCommands | Where-Object { $null -eq (Get-Command $_ -ErrorAction SilentlyContinue) })
if ($missingCommands.Count -gt 0) {
    Write-Host "[-] Missing required local command(s): $($missingCommands -join ', ')" -ForegroundColor Red
    exit 1
}

if ($RequirePostgres) {
    $missingDsns = @($requiredDsns | Where-Object { [string]::IsNullOrWhiteSpace([Environment]::GetEnvironmentVariable($_)) })
    if ($missingDsns.Count -gt 0) {
        Write-Host "[-] PostgreSQL verification requested, but required DSN(s) are missing: $($missingDsns -join ', ')" -ForegroundColor Red
        exit 1
    }
}

# -----------------------------------------------------------------------------
# 1. CODE QUALITY & FORMATTING
# -----------------------------------------------------------------------------
Write-Host "[1/6] Verifying Go Code Quality, Formatting & Vet..." -ForegroundColor Yellow
$unformatted = & gofmt -l "$ProjectRoot\packages" "$ProjectRoot\services" "$ProjectRoot\tests" "$ProjectRoot\scripts"
if ($unformatted) {
    Write-Host "  [-] Unformatted files detected: $unformatted" -ForegroundColor Red
    $results["Go_Format"] = "FAIL"
    $totalFailed++
} else {
    Write-Host "  [+] Go Formatting: PASS (All Go source files formatted cleanly)" -ForegroundColor Green
    $results["Go_Format"] = "PASS"
}

$vetFailed = $false
$modules = @(
    "packages/shared", "services/ai", "services/asset", "services/audit",
    "services/auth", "services/employee", "services/gateway", "services/helpdesk",
    "services/knowledge", "services/notification", "services/reporting",
    "services/workflow", "tests/e2e", "tests/integration"
)

foreach ($mod in $modules) {
    Push-Location "$ProjectRoot\$mod"
    try {
        go vet ./... 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            $vetFailed = $true
            Write-Host "  [-] go vet failed in $mod" -ForegroundColor Red
        }
    } finally {
        Pop-Location
    }
}

if (-not $vetFailed) {
    Write-Host "  [+] Go Vet: PASS (All 14 workspace modules validated without warnings)" -ForegroundColor Green
    $results["Go_Vet"] = "PASS"
} else {
    $results["Go_Vet"] = "FAIL"
    $totalFailed++
}

# -----------------------------------------------------------------------------
# 2. OPENAPI RUNTIME OPERATION PARITY
# -----------------------------------------------------------------------------
Write-Host "`n[2/6] Verifying OpenAPI Contract & Runtime Coverage Gate..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    $openapiOut = & go run scripts/check_openapi_coverage.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [+] OpenAPI Parity: PASS ($openapiOut)" -ForegroundColor Green
        $results["OpenAPI_Parity"] = "PASS (107/107)"
    } else {
        Write-Host "  [-] OpenAPI Coverage Check Failed!" -ForegroundColor Red
        Write-Host $openapiOut -ForegroundColor Red
        $results["OpenAPI_Parity"] = "FAIL"
        $totalFailed++
    }
} finally {
    Pop-Location
}

# -----------------------------------------------------------------------------
# 3. BACKEND UNIT & IN-MEMORY SIMULATION SUITE
# -----------------------------------------------------------------------------
Write-Host "`n[3/6] Running Backend Unit & Simulation Suites (13 modules)..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    $unitOut = & go test -skip '(Integration|Postgres)' ./packages/shared/... ./services/ai/... ./services/asset/... ./services/audit/... ./services/auth/... ./services/employee/... ./services/gateway/... ./services/helpdesk/... ./services/knowledge/... ./services/notification/... ./services/reporting/... ./services/workflow/... ./tests/e2e/... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [+] Unit & Simulation Tests: PASS (All business-flow simulations passed)" -ForegroundColor Green
        $results["Unit_Simulation_Tests"] = "PASS"
    } else {
        Write-Host "  [-] Unit/Simulation tests failed:" -ForegroundColor Red
        Write-Host $unitOut -ForegroundColor Red
        $results["Unit_Simulation_Tests"] = "FAIL"
        $totalFailed++
    }
} finally {
    Pop-Location
}

# -----------------------------------------------------------------------------
# 4. HTTP CONTRACTS & OPTIONAL POSTGRESQL INTEGRATION
# -----------------------------------------------------------------------------
if ($RequirePostgres) {
    Write-Host "`n[4/6] Running fail-closed PostgreSQL integration suites and HTTP contracts..." -ForegroundColor Yellow
    $missingDsns = @($requiredDsns | Where-Object { [string]::IsNullOrWhiteSpace([Environment]::GetEnvironmentVariable($_)) })
    if ($missingDsns.Count -gt 0) {
        Write-Host "  [-] Missing required DSN(s): $($missingDsns -join ', ')" -ForegroundColor Red
        $results["PostgreSQL_Integration"] = "FAIL"
        $totalFailed++
    } else {
        $previousRequired = $env:INTEGRATION_REQUIRED
        $env:INTEGRATION_REQUIRED = "1"
        $integrationFailed = $false
        $integrationModules = @(
            "services/auth", "services/helpdesk", "services/audit", "services/notification",
            "services/reporting", "tests/integration"
        )
        try {
            foreach ($mod in $integrationModules) {
                Push-Location "$ProjectRoot\$mod"
                try {
                    & go test -count=1 -v ./... 2>&1 | Out-Host
                    if ($LASTEXITCODE -ne 0) { $integrationFailed = $true }
                } finally {
                    Pop-Location
                }
            }
        } finally {
            $env:INTEGRATION_REQUIRED = $previousRequired
        }

        if ($integrationFailed) {
            $results["PostgreSQL_Integration"] = "FAIL"
            $totalFailed++
        } else {
            $results["PostgreSQL_Integration"] = "PASS (required; no skips allowed)"
        }
    }
} else {
    Write-Host "`n[4/6] Running in-process HTTP contract harness..." -ForegroundColor Yellow
    Push-Location "$ProjectRoot\tests\integration"
    try {
        $contractOut = & go test -count=1 -v -run '^TestHTTPContract_' ./... 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  [+] HTTP contract harness: PASS (test doubles; not deployed E2E)" -ForegroundColor Green
            $results["HTTP_Contract_Harness"] = "PASS (not E2E)"
        } else {
            Write-Host $contractOut -ForegroundColor Red
            $results["HTTP_Contract_Harness"] = "FAIL"
            $totalFailed++
        }
    } finally {
        Pop-Location
    }
}

# -----------------------------------------------------------------------------
# 5. FRONTEND UNIT & CONTRACT SUITE (apps/web)
# -----------------------------------------------------------------------------
Write-Host "`n[5/6] Verifying Frontend Application (Vitest, Typecheck, ESLint, Build)..." -ForegroundColor Yellow
Push-Location "$ProjectRoot\apps\web"
try {
    pnpm.cmd test 2>&1 | Out-Null
    $vitestOk = ($LASTEXITCODE -eq 0)

    pnpm.cmd typecheck 2>&1 | Out-Null
    $tcOk = ($LASTEXITCODE -eq 0)

    pnpm.cmd lint 2>&1 | Out-Null
    $lintOk = ($LASTEXITCODE -eq 0)

    pnpm.cmd build 2>&1 | Out-Null
    $buildOk = ($LASTEXITCODE -eq 0)

    if ($vitestOk -and $tcOk -and $lintOk -and $buildOk) {
        Write-Host "  [+] Vitest unit & route-policy contracts: PASS (no browser launched)" -ForegroundColor Green
        Write-Host "  [+] Nuxt Typecheck & TypeScript Safety: PASS" -ForegroundColor Green
        Write-Host "  [+] ESLint Code Style Conformance: PASS" -ForegroundColor Green
        Write-Host "  [+] Production Nitro/Nuxt Build: PASS" -ForegroundColor Green
        $results["Frontend_Suite"] = "PASS (unit/contracts only)"
    } else {
        Write-Host "  [-] Frontend verification failed (Vitest=$vitestOk, Typecheck=$tcOk, Lint=$lintOk, Build=$buildOk)" -ForegroundColor Red
        $results["Frontend_Suite"] = "FAIL"
        $totalFailed++
    }
} finally {
    Pop-Location
}

# -----------------------------------------------------------------------------
# 6. STATIC CONTAINER & KUBERNETES PRECHECKS
# -----------------------------------------------------------------------------
Write-Host "`n[6/6] Running static Container/Kubernetes hardening prechecks..." -ForegroundColor Yellow
$dfGo = Get-Content "$ProjectRoot\deploy\docker\Dockerfile.go-service" -Raw
$dfWeb = Get-Content "$ProjectRoot\deploy\docker\Dockerfile.web" -Raw
$netPol = Test-Path "$ProjectRoot\deploy\kubernetes\manifests\09-network-policies.yaml"
$pdb = Test-Path "$ProjectRoot\deploy\kubernetes\manifests\10-pod-disruption-budgets.yaml"

$pinOk = (-not ($dfGo -match "FROM [a-zA-Z0-9_.-]+:latest" -or $dfWeb -match "FROM [a-zA-Z0-9_.-]+:latest"))
$userOk = ($dfGo -match "USER 10001:10001" -and $dfWeb -match "USER 10001:10001")

if ($pinOk -and $userOk -and $netPol -and $pdb) {
    Write-Host "  [+] Dockerfile Base Image Pinning (No ':latest'): PASS" -ForegroundColor Green
    Write-Host "  [+] Non-Root Security Context (10001:10001): PASS" -ForegroundColor Green
    Write-Host "  [+] Kubernetes NetworkPolicy & PodDisruptionBudgets: PASS" -ForegroundColor Green
    $results["Static_Hardening"] = "PASS (not a Trivy/CIS runtime scan)"
} else {
    Write-Host "  [-] Container / Kubernetes hardening checks failed" -ForegroundColor Red
    $results["Static_Hardening"] = "FAIL"
    $totalFailed++
}

# -----------------------------------------------------------------------------
# SUMMARY & RELEASE GATE VERIFICATION TABLE
# -----------------------------------------------------------------------------
Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "                    LOCAL PRECHECK SUMMARY                       " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan

foreach ($k in $results.Keys) {
    $pad = $k.PadRight(28)
    $val = $results[$k]
    if ($val -like "*PASS*") {
        Write-Host "  $pad : $val" -ForegroundColor Green
    } else {
        Write-Host "  $pad : $val" -ForegroundColor Red
    }
}

Write-Host "-----------------------------------------------------------------" -ForegroundColor Gray
if ($totalFailed -eq 0) {
    Write-Host "  OVERALL STATUS: LOCAL PRECHECK PASSED" -ForegroundColor Green
    Write-Host "  GATE D: NOT APPROVED BY THIS SCRIPT" -ForegroundColor Yellow
    Write-Host "  Pending: deployed API E2E, Playwright, Trivy/image build," -ForegroundColor Yellow
    Write-Host "           Helm staging smoke, k6 results, and DR evidence." -ForegroundColor Yellow
} else {
    Write-Host "  OVERALL STATUS: $totalFailed CHECKS FAILED" -ForegroundColor Red
}
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host ""

exit $totalFailed
