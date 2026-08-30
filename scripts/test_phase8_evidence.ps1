# ==============================================================================
# EOMP Phase 8 — Master Enterprise Validation & Evidence Collection Runner
# ==============================================================================

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "   EOMP PHASE 8: MASTER ENTERPRISE VALIDATION & FINAL SIGN-OFF    " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host ""

$workspaceRoot = (Get-Item -Path "$PSScriptRoot\..").FullName
$failedChecks = 0
$skippedChecks = 0
$startTime = Get-Date

# -----------------------------------------------------------------------------
# Check 1: Security & Compliance Verification
# -----------------------------------------------------------------------------
Write-Host "[1/6] Running Security & Compliance Verification..." -ForegroundColor Yellow
Push-Location $workspaceRoot
$secTestOutput = & go test -v ./tests/e2e -run TestPhase8_SecurityAndComplianceValidation 2>&1
$secExitCode = $LASTEXITCODE
Pop-Location

if ($secExitCode -eq 0) {
    Write-Host "  [+] JWT RBAC claims, Dynamic CORS whitelisting & Anti-Spoofing verified." -ForegroundColor Green
    Write-Host "  [+] Distributed Rate Limiter (10r/m & 100r/m) throttles correctly (429)." -ForegroundColor Green
    Write-Host "  [+] Tamper-Evident SHA-256 Checksum & Data Masking operational." -ForegroundColor Green
} else {
    Write-Host "  [-] Security & Compliance verification failed!" -ForegroundColor Red
    Write-Host $secTestOutput -ForegroundColor Red
    $failedChecks++
}

# -----------------------------------------------------------------------------
# Check 2: Business & AI Operations Golden Flow Full Lifecycle
# -----------------------------------------------------------------------------
Write-Host ""
Write-Host "[2/6] Running Business & AI Golden Flow Lifecycle Verification..." -ForegroundColor Yellow
Push-Location $workspaceRoot
$goldenTestOutput = & go test -v ./tests/e2e -run TestPhase8_BusinessAndAIGoldenFlowValidation 2>&1
$goldenExitCode = $LASTEXITCODE
Pop-Location

if ($goldenExitCode -eq 0) {
    Write-Host "  [+] Multi-Role Matrix: Employee, Manager, IT Agent & Admin verified." -ForegroundColor Green
    Write-Host "  [+] Ticket creation & dynamic SLA deadline calculated accurately." -ForegroundColor Green
    Write-Host "  [+] AI Auto-Triage & Confidence Score (>= 88%) verified." -ForegroundColor Green
    Write-Host "  [+] Qdrant Vector Semantic RAG Retrieval with Citation Runbooks verified." -ForegroundColor Green
    Write-Host "  [+] Optimistic Locking CAS (50 Goroutines) prevents lost updates (409 Conflict)." -ForegroundColor Green
} else {
    Write-Host "  [-] Business & AI Golden Flow verification failed!" -ForegroundColor Red
    Write-Host $goldenTestOutput -ForegroundColor Red
    $failedChecks++
}

# -----------------------------------------------------------------------------
# Check 3: SRE Resilience, Fallback & Disaster Recovery Drill
# -----------------------------------------------------------------------------
Write-Host ""
Write-Host "[3/6] Running SRE Resilience & Disaster Recovery Verification..." -ForegroundColor Yellow
Push-Location $workspaceRoot
$sreTestOutput = & go test -v ./tests/e2e -run TestPhase8_SREResilienceAndDisasterRecoveryValidation 2>&1
$sreExitCode = $LASTEXITCODE
Pop-Location

if ($sreExitCode -eq 0 -and $sreTestOutput -match "no real restore drill was executed") {
    Write-Host "  [!] SRE unit checks passed, but DR restore evidence is missing (SKIP)." -ForegroundColor Yellow
    $skippedChecks++
} elseif ($sreExitCode -eq 0) {
    Write-Host "  [+] PostgreSQL Connection Pool parameters optimized across all 11 services." -ForegroundColor Green
    Write-Host "  [+] Graceful In-Memory Fallback ensures zero downtime during broker failover." -ForegroundColor Green
    Write-Host "  [+] Disaster Recovery evidence met the configured RPO/RTO thresholds." -ForegroundColor Green
} else {
    Write-Host "  [-] SRE Resilience & DR verification failed!" -ForegroundColor Red
    Write-Host $sreTestOutput -ForegroundColor Red
    $failedChecks++
}

# -----------------------------------------------------------------------------
# Check 4: Full 13-Module Go Test Suite Execution
# -----------------------------------------------------------------------------
Write-Host ""
Write-Host "[4/6] Executing Comprehensive Go Test Suite across all 13 modules..." -ForegroundColor Yellow
Push-Location $workspaceRoot
$allTestOutput = & go test ./packages/shared/... ./services/ai/... ./services/asset/... ./services/audit/... ./services/auth/... ./services/employee/... ./services/gateway/... ./services/helpdesk/... ./services/knowledge/... ./services/notification/... ./services/reporting/... ./services/workflow/... ./tests/e2e/... 2>&1
$allExitCode = $LASTEXITCODE
Pop-Location

if ($allExitCode -eq 0) {
    Write-Host "  [+] 100% Pass Rate across all 13 Go Modules & E2E Suites (0 failures)." -ForegroundColor Green
} else {
    Write-Host "  [-] Master Go Test Suite failed!" -ForegroundColor Red
    Write-Host $allTestOutput -ForegroundColor Red
    $failedChecks++
}

# -----------------------------------------------------------------------------
# Check 5: DevSecOps, Docker Image Pinning & Kubernetes Hardening
# -----------------------------------------------------------------------------
Write-Host ""
Write-Host "[5/6] Verifying DevSecOps Gates, Docker Pinning & Kubernetes Manifests..." -ForegroundColor Yellow
$devsecopsScript = "$workspaceRoot\scripts\test_devsecops.ps1"
if (Test-Path $devsecopsScript) {
    $devsecOutput = & powershell -ExecutionPolicy Bypass -File $devsecopsScript 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [+] All 5 DevSecOps checks (Formatting, Tests, Non-root Docker, K8s CIS, Alerts) PASSED." -ForegroundColor Green
    } else {
        Write-Host "  [-] DevSecOps Gate check failed!" -ForegroundColor Red
        $failedChecks++
    }
} else {
    Write-Host "  [-] DevSecOps test script missing!" -ForegroundColor Red
    $failedChecks++
}

# -----------------------------------------------------------------------------
# Check 6: Frontend Nuxt 4 Build & Bundle Inspection
# -----------------------------------------------------------------------------
Write-Host ""
Write-Host "[6/6] Verifying Frontend Nuxt 4 SSR Production Bundle..." -ForegroundColor Yellow
$webDir = "$workspaceRoot\apps\web"
Push-Location $webDir
& pnpm.cmd build 2>&1 | Out-Null
$webBuildExitCode = $LASTEXITCODE
Pop-Location
if ($webBuildExitCode -eq 0) {
    Write-Host "  [+] Nuxt 4 production build completed successfully." -ForegroundColor Green
} else {
    Write-Host "  [-] Nuxt 4 .output directory not found! Run 'pnpm run build' in apps/web." -ForegroundColor Red
    $failedChecks++
}

# -----------------------------------------------------------------------------
# Summary & Sign-off Decision
# -----------------------------------------------------------------------------
$elapsed = (Get-Date) - $startTime
Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
if ($failedChecks -eq 0 -and $skippedChecks -eq 0) {
    Write-Host "   SUCCESS: all executed Phase 8 checks passed.                 " -ForegroundColor Green
    Write-Host "   Execution Time: $($elapsed.TotalSeconds.ToString('F2')) seconds                           " -ForegroundColor Green
    Write-Host "   This script does not by itself grant production acceptance.  " -ForegroundColor Green
    Write-Host "=================================================================" -ForegroundColor Cyan
    exit 0
} elseif ($failedChecks -eq 0) {
    Write-Host "   PASS WITH $skippedChecks SKIPPED CHECK(S); NOT PRODUCTION-CERTIFIED. " -ForegroundColor Yellow
    Write-Host "=================================================================" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "   FAILED: $failedChecks check(s) failed during Phase 8 validation." -ForegroundColor Red
    Write-Host "=================================================================" -ForegroundColor Cyan
    exit 1
}
