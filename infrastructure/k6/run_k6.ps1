<#
.SYNOPSIS
    Runs K6 Load Testing Suite against EOMP API Gateway
#>

param (
    [string]$GatewayUrl = "http://localhost:8080",
    [string]$TestScript = "load_test.js"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   EOMP K6 LOAD & PERFORMANCE TEST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Target URL : $GatewayUrl" -ForegroundColor Yellow
Write-Host "Script     : $TestScript" -ForegroundColor Yellow
Write-Host ""

$k6Path = Get-Command k6 -ErrorAction SilentlyContinue

if ($k6Path) {
    & k6 run -e GATEWAY_URL=$GatewayUrl "$PSScriptRoot\$TestScript"
} else {
    Write-Host "k6 binary not found in PATH. Using Docker container to run k6..." -ForegroundColor Yellow
    docker run --rm -i --network host -e GATEWAY_URL=$GatewayUrl -v "${PSScriptRoot}:/scripts" grafana/k6 run "/scripts/$TestScript"
}
