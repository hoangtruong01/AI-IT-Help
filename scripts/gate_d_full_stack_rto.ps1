[CmdletBinding()]
param(
    [string]$ProjectName = 'eomp',
    [string]$EvidencePath = 'docs/evidence/gate-d/dr_full_service.json',
    [int]$ExpectedHealthyContainers = 20,
    [int]$ThresholdSeconds = 900
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw 'docker is required'
}

function Get-EompContainerState {
    $rows = @(& docker ps --filter "name=$ProjectName-" --format '{{.Names}}|{{.Status}}')
    if ($LASTEXITCODE -ne 0) { throw 'docker ps failed' }
    return @($rows | Where-Object { $_ } | ForEach-Object {
        $parts = $_ -split '\|', 2
        [ordered]@{
            name = $parts[0]
            status = $parts[1]
            healthy = $parts[1] -match '\(healthy\)'
        }
    })
}

$containerNames = @(& docker ps -a --filter "name=$ProjectName-" --format '{{.Names}}')
if ($LASTEXITCODE -ne 0) { throw 'failed to enumerate EOMP containers' }
$containerNames = @($containerNames | Where-Object { $_ -and $_ -ne "$ProjectName-minio-init" } | Sort-Object -Unique)
if ($containerNames.Count -ne $ExpectedHealthyContainers) {
    throw "expected $ExpectedHealthyContainers existing EOMP containers, found $($containerNames.Count)"
}

$infraNames = @(
    "$ProjectName-postgres", "$ProjectName-redis", "$ProjectName-rabbitmq", "$ProjectName-minio",
    "$ProjectName-qdrant", "$ProjectName-prometheus", "$ProjectName-grafana", "$ProjectName-loki"
) | Where-Object { $containerNames -contains $_ }
$appNames = @($containerNames | Where-Object { $infraNames -notcontains $_ })
if ($infraNames.Count -ne 8 -or $appNames.Count -ne 12) {
    throw "expected 8 infrastructure and 12 application containers, found $($infraNames.Count) and $($appNames.Count)"
}

Write-Host 'Stopping the 12 application and 8 infrastructure containers without deleting containers or volumes...'
& docker stop --time 30 @appNames | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'failed to stop application containers' }
& docker stop --time 30 @infraNames | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'failed to stop infrastructure containers' }

$startedAt = [DateTimeOffset]::UtcNow
$stopwatch = [Diagnostics.Stopwatch]::StartNew()

Write-Host 'Starting infrastructure and application containers from a cold stopped state...'
& docker start @infraNames | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'failed to start infrastructure containers' }
& docker start @appNames | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'failed to start application containers' }

$deadline = [DateTimeOffset]::UtcNow.AddSeconds($ThresholdSeconds)
$lastState = @()
$gatewayReady = $false
$webReady = $false
while ([DateTimeOffset]::UtcNow -lt $deadline) {
    $lastState = @(Get-EompContainerState)
    $containersReady = $lastState.Count -eq $ExpectedHealthyContainers -and @($lastState | Where-Object { -not $_.healthy }).Count -eq 0
    if ($containersReady) {
        try {
            $gatewayReady = (Invoke-WebRequest -Uri 'http://127.0.0.1:8080/health' -UseBasicParsing -TimeoutSec 5).StatusCode -eq 200
        } catch { $gatewayReady = $false }
        try {
            $webReady = (Invoke-WebRequest -Uri 'http://127.0.0.1:3000/' -UseBasicParsing -TimeoutSec 5).StatusCode -eq 200
        } catch { $webReady = $false }
        if ($gatewayReady -and $webReady) { break }
    }
    Start-Sleep -Seconds 2
}
$stopwatch.Stop()

$lastState = @(Get-EompContainerState)
$allHealthy = $lastState.Count -eq $ExpectedHealthyContainers -and @($lastState | Where-Object { -not $_.healthy }).Count -eq 0
$passed = $allHealthy -and $gatewayReady -and $webReady -and $stopwatch.Elapsed.TotalSeconds -lt $ThresholdSeconds
$revision = (& git -C $projectRoot rev-parse HEAD).Trim()

$evidence = [ordered]@{
    schema_version = 1
    result = if ($passed) { 'PASS' } else { 'FAIL' }
    scope = 'local Docker cold stopped-state to full-service readiness; WAL/PITR RPO is separate'
    source_revision = $revision
    started_at_utc = $startedAt.ToString('o')
    completed_at_utc = [DateTimeOffset]::UtcNow.ToString('o')
    rto_seconds = [math]::Round($stopwatch.Elapsed.TotalSeconds, 3)
    threshold_seconds = $ThresholdSeconds
    expected_container_count = $ExpectedHealthyContainers
    healthy_container_count = @($lastState | Where-Object { $_.healthy }).Count
    gateway_health_http_200 = $gatewayReady
    web_http_200 = $webReady
    containers = $lastState
}

$fullEvidencePath = Join-Path $projectRoot $EvidencePath
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $fullEvidencePath) | Out-Null
$evidence | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $fullEvidencePath -Encoding UTF8

if (-not $passed) {
    throw "Full-service RTO verification failed after $($stopwatch.Elapsed.TotalSeconds)s. Evidence: $fullEvidencePath"
}

Write-Host "Full-service local RTO PASS: $($evidence.rto_seconds)s, $($evidence.healthy_container_count)/$ExpectedHealthyContainers healthy"
Write-Host "Evidence: $fullEvidencePath"
