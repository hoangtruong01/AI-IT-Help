[CmdletBinding()]
param([string]$OutputPath = 'docs/evidence/gate-d/image_manifest.json')

$ErrorActionPreference = 'Stop'
$references = @(docker image ls 'eomp/*:gate-d' --format '{{.Repository}}:{{.Tag}}')
if ($LASTEXITCODE -ne 0) { throw 'Unable to list EOMP Gate D images' }
if ($references.Count -ne 12) { throw "Expected 12 Gate D images, found $($references.Count)" }

$images = foreach ($reference in ($references | Sort-Object)) {
    $raw = docker image inspect $reference | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0 -or $raw.Count -ne 1) { throw "Unable to inspect $reference" }
    $image = $raw[0]
    $user = [string]$image.Config.User
    if ([string]::IsNullOrWhiteSpace($user) -or $user -eq '0' -or $user -eq 'root' -or $user.StartsWith('0:')) {
        throw "$reference does not declare a non-root runtime user"
    }
    [ordered]@{
        reference = $reference
        digest = [string]$image.Id
        created_at = [string]$image.Created
        size_bytes = [long]$image.Size
        runtime_user = $user
    }
}

$containers = @(docker ps --filter 'name=eomp-' --format '{{.Names}}|{{.Status}}')
if ($LASTEXITCODE -ne 0) { throw 'Unable to list EOMP containers' }
$unhealthy = @($containers | Where-Object { $_ -notmatch '\(healthy\)' })
if ($containers.Count -ne 20 -or $unhealthy.Count -ne 0) {
    throw "Expected 20 healthy EOMP containers; found $($containers.Count), unhealthy=$($unhealthy.Count)"
}

$revision = (git rev-parse HEAD).Trim()
$evidence = [ordered]@{
    schema_version = 1
    result = 'PASS'
    scope = 'local Docker image build/runtime metadata; vulnerability scan is separate'
    generated_at_utc = [DateTimeOffset]::UtcNow.ToString('o')
    source_revision = $revision
    image_count = $images.Count
    healthy_container_count = $containers.Count
    images = $images
}

$fullPath = Join-Path (Get-Location) $OutputPath
New-Item -ItemType Directory -Force -Path (Split-Path -Parent $fullPath) | Out-Null
$evidence | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $fullPath -Encoding UTF8
Write-Host "Gate D image manifest PASS: $($images.Count) non-root images, $($containers.Count) healthy containers"
Write-Host "Evidence: $fullPath"
