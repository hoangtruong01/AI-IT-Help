<#
.SYNOPSIS
    EOMP Gate D-01 — Automated PostgreSQL Integration Runner (PowerShell/Windows/CI)

.DESCRIPTION
    Executes fail-closed repository integration suites across six isolated databases:
    1. auth_ci
    2. helpdesk_ci
    3. audit_ci
    4. migration_ci
    5. notification_ci
    6. reporting_ci

    Enforces INTEGRATION_REQUIRED=1 so missing databases or connections trigger a hard failure
    rather than a skip. Records execution evidence to docs/evidence/gate-d/.
#>

param(
    [string]$PostgresHost = "127.0.0.1",
    [int]$PostgresPort = 5432,
    [string]$PostgresUser = "eomp_ci",
    [string]$PostgresPassword = "eomp_ci_password",
    [switch]$SkipDockerLaunch = $false
)

$ErrorActionPreference = "Continue"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not (Test-Path "$ProjectRoot\docker-compose.yml")) {
    $ProjectRoot = $PSScriptRoot
}

$EvidenceDir = "$ProjectRoot\docs\evidence\gate-d"
if (-not (Test-Path $EvidenceDir)) {
    New-Item -ItemType Directory -Path $EvidenceDir -Force | Out-Null
}
$LogFile = "$EvidenceDir\ci_postgres_integration.log"
$ResultFile = "$EvidenceDir\ci_postgres_integration.json"

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "    EOMP GATE D-01: AUTOMATED POSTGRESQL INTEGRATION SUITE       " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "  Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss UTC')" -ForegroundColor Gray
Write-Host "  Project:   $ProjectRoot" -ForegroundColor Gray
Write-Host ""

$previousIntegrationRequired = $env:INTEGRATION_REQUIRED
$env:INTEGRATION_REQUIRED = "1"

try {
    # Check if DSNs are preconfigured in environment
    $hasExistingDsns = (-not [string]::IsNullOrWhiteSpace($env:AUTH_INTEGRATION_DSN) -and
                        -not [string]::IsNullOrWhiteSpace($env:HELPDESK_INTEGRATION_DSN) -and
                        -not [string]::IsNullOrWhiteSpace($env:AUDIT_INTEGRATION_DSN) -and
                        -not [string]::IsNullOrWhiteSpace($env:INTEGRATION_POSTGRES_DSN) -and
                        -not [string]::IsNullOrWhiteSpace($env:NOTIFICATION_INTEGRATION_DSN) -and
                        -not [string]::IsNullOrWhiteSpace($env:REPORTING_INTEGRATION_DSN))

    if (-not $hasExistingDsns) {
        $baseDsn = "postgres://${PostgresUser}:${PostgresPassword}@${PostgresHost}:${PostgresPort}"
        $env:AUTH_INTEGRATION_DSN = "${baseDsn}/auth_ci?sslmode=disable"
        $env:HELPDESK_INTEGRATION_DSN = "${baseDsn}/helpdesk_ci?sslmode=disable"
        $env:AUDIT_INTEGRATION_DSN = "${baseDsn}/audit_ci?sslmode=disable"
        $env:INTEGRATION_POSTGRES_DSN = "${baseDsn}/migration_ci?sslmode=disable"
        $env:NOTIFICATION_INTEGRATION_DSN = "${baseDsn}/notification_ci?sslmode=disable"
        $env:REPORTING_INTEGRATION_DSN = "${baseDsn}/reporting_ci?sslmode=disable"
    }

    Write-Host "[1/2] Verifying Target Modules & Compiling Integration Tests..." -ForegroundColor Yellow
    $integrationTargets = @(
        @{ Dir = "$ProjectRoot\services\auth"; Pkg = "./internal/repository" },
        @{ Dir = "$ProjectRoot\services\helpdesk"; Pkg = "./internal/repository" },
        @{ Dir = "$ProjectRoot\services\audit"; Pkg = "./internal/repository" },
        @{ Dir = "$ProjectRoot\services\notification"; Pkg = "./internal/repository" },
        @{ Dir = "$ProjectRoot\services\reporting"; Pkg = "./internal/repository" },
        @{ Dir = "$ProjectRoot\tests\integration"; Pkg = "./..." }
    )

    $compilationFailed = $false
    foreach ($target in $integrationTargets) {
        Push-Location $target.Dir
        try {
            $null = & go test -c -o test_bin.tmp $($target.Pkg) 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Host "  [-] Failed to compile integration test in $($target.Dir)" -ForegroundColor Red
                $compilationFailed = $true
            } else {
                Remove-Item "test_bin.tmp" -Force -ErrorAction SilentlyContinue
            }
        } finally {
            Pop-Location
        }
    }

    if ($compilationFailed) {
        Write-Host "[-] Integration test compilation failed." -ForegroundColor Red
        exit 1
    }
    Write-Host "  [+] All 6 integration test packages compiled cleanly." -ForegroundColor Green

    Write-Host "`n[2/2] Running Fail-Closed Repository Integration Suites..." -ForegroundColor Yellow
    $suiteFailed = $false
    $testOutputLines = @()

    foreach ($target in $integrationTargets) {
        Push-Location $target.Dir
        try {
            $output = & go test -count=1 -v $($target.Pkg) 2>&1
            $exitCode = $LASTEXITCODE
            $testOutputLines += $output
            if ($exitCode -ne 0) {
                Write-Host "  [-] Test suite failed in $($target.Dir)" -ForegroundColor Red
                $suiteFailed = $true
            } else {
                Write-Host "  [+] PASS: $($target.Dir) $($target.Pkg)" -ForegroundColor Green
            }
        } finally {
            Pop-Location
        }
    }

    $testOutputLines | Set-Content -LiteralPath $LogFile -Encoding utf8

    $sourceRevision = (& git -C $ProjectRoot rev-parse HEAD 2>$null)
    if (-not $sourceRevision) { $sourceRevision = "local" } else { $sourceRevision = $sourceRevision.Trim() }

    $resultStatus = if (-not $suiteFailed) { "PASS" } else { "FAIL" }

    $resultData = [ordered]@{
        schema_version = 1
        task_id = "TASK-REL-002"
        gate = "Gate D-01"
        result = $resultStatus
        source_revision = $sourceRevision
        completed_at_utc = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
        database_count = 6
        integration_required = $true
        ci_build_url = if ($env:BUILD_URL) { $env:BUILD_URL } else { "local-powershell-runner" }
    }
    $resultData | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $ResultFile -Encoding utf8

    Write-Host ""
    Write-Host "=================================================================" -ForegroundColor Cyan
    Write-Host "         INTEGRATION RESULT: $resultStatus (Log: $LogFile)      " -ForegroundColor Cyan
    Write-Host "=================================================================" -ForegroundColor Cyan
    Write-Host ""

    if ($suiteFailed) { exit 1 } else { exit 0 }
} finally {
    $env:INTEGRATION_REQUIRED = $previousIntegrationRequired
}
