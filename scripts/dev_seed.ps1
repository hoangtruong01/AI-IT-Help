[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $repoRoot '.env'

if (-not (Test-Path -LiteralPath $envFile)) {
    throw '.env is required so the seed command can verify APP_ENV.'
}

$settings = @{}
Get-Content -LiteralPath $envFile | ForEach-Object {
    if ($_ -match '^\s*([^#][^=]*)=(.*)$') {
        $settings[$matches[1].Trim()] = $matches[2].Trim().Trim('"')
    }
}

$appEnv = $settings['APP_ENV']
if ($appEnv -notin @('development', 'test')) {
    throw "Refusing to seed APP_ENV='$appEnv'. Only development or test is allowed."
}

$dbUser = $settings['POSTGRES_USER']
if (-not $dbUser) { $dbUser = 'eomp' }

function Get-SettingOrDefault {
    param([string]$Name, [string]$Default)
    $value = $settings[$Name]
    if ($value) { return $value }
    return $Default
}

$targets = @(
    @{ Database = (Get-SettingOrDefault 'ASSET_DB_NAME' 'asset_db'); File = 'asset.sql' },
    @{ Database = (Get-SettingOrDefault 'HELPDESK_DB_NAME' 'helpdesk_db'); File = 'helpdesk.sql' },
    @{ Database = (Get-SettingOrDefault 'WORKFLOW_DB_NAME' 'workflow_db'); File = 'workflow.sql' },
    @{ Database = (Get-SettingOrDefault 'KNOWLEDGE_DB_NAME' 'knowledge_db'); File = 'knowledge.sql' },
    @{ Database = (Get-SettingOrDefault 'NOTIFICATION_DB_NAME' 'notification_db'); File = 'notification.sql' }
)

foreach ($target in $targets) {
    $seedPath = Join-Path $PSScriptRoot (Join-Path 'dev-seeds' $target.File)
    Write-Host "Seeding $($target.Database) from $($target.File)..."
    Get-Content -Raw -LiteralPath $seedPath |
        docker compose --project-directory $repoRoot exec -T postgres `
            psql -v ON_ERROR_STOP=1 -U $dbUser -d $target.Database
    if ($LASTEXITCODE -ne 0) {
        throw "Seed failed for $($target.Database)."
    }
}

Write-Host "Development seed completed for APP_ENV=$appEnv."
