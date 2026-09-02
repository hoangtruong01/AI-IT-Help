[CmdletBinding()]
param(
    [string]$BaseUrl = 'http://127.0.0.1:8080',
    [string]$AdminEmail = 'gate-admin@local.test',
    [string]$EvidencePath = 'docs/evidence/gate-d/deployed_stack_e2e.json',
    [string]$PostgresContainer = 'eomp-postgres',
    [string]$PostgresUser = 'eomp'
)

$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($env:GATE_D_ADMIN_PASSWORD)) {
    throw 'GATE_D_ADMIN_PASSWORD is required'
}

function Invoke-EompRequest {
    param(
        [Parameter(Mandatory)][string]$Method,
        [Parameter(Mandatory)][string]$Path,
        [Parameter(Mandatory)][int[]]$ExpectedStatus,
        [object]$Body,
        [string]$Token,
        [hashtable]$AdditionalHeaders = @{}
    )

    $headers = @{}
    if ($Token) {
        $headers.Authorization = "Bearer $Token"
    }
    foreach ($key in $AdditionalHeaders.Keys) {
        $headers[$key] = $AdditionalHeaders[$key]
    }

    $params = @{
        Uri = $BaseUrl.TrimEnd('/') + $Path
        Method = $Method
        Headers = $headers
        UseBasicParsing = $true
        ErrorAction = 'Stop'
    }
    if ($null -ne $Body) {
        $params.ContentType = 'application/json'
        $params.Body = $Body | ConvertTo-Json -Depth 10 -Compress
    }

    $status = 0
    $content = ''
    try {
        $response = Invoke-WebRequest @params
        $status = [int]$response.StatusCode
        $content = [string]$response.Content
    } catch {
        $webResponse = $_.Exception.Response
        if ($null -eq $webResponse) {
            throw
        }
        $status = [int]$webResponse.StatusCode
        $stream = $webResponse.GetResponseStream()
        if ($null -ne $stream) {
            $reader = New-Object System.IO.StreamReader($stream)
            try { $content = $reader.ReadToEnd() } finally { $reader.Dispose() }
        }
    }

    if ($ExpectedStatus -notcontains $status) {
        throw "$Method $Path returned HTTP $status; expected $($ExpectedStatus -join ','). Body: $content"
    }

    $data = $null
    if (-not [string]::IsNullOrWhiteSpace($content)) {
        try { $data = $content | ConvertFrom-Json } catch { $data = $content }
    }
    return [pscustomobject]@{ Status = $status; Data = $data }
}

function Invoke-DatabaseScalar {
    param([Parameter(Mandatory)][string]$Database, [Parameter(Mandatory)][string]$Sql)
    $output = & docker exec $PostgresContainer psql -v ON_ERROR_STOP=1 -U $PostgresUser -d $Database -Atc $Sql 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw "PostgreSQL assertion failed for $Database`: $($output -join [Environment]::NewLine)"
    }
    return (($output | ForEach-Object { [string]$_ }) -join '').Trim()
}

function Wait-DatabaseValue {
    param(
        [Parameter(Mandatory)][string]$Database,
        [Parameter(Mandatory)][string]$Sql,
        [Parameter(Mandatory)][scriptblock]$Accept,
        [int]$Attempts = 30
    )
    for ($attempt = 1; $attempt -le $Attempts; $attempt++) {
        $value = Invoke-DatabaseScalar -Database $Database -Sql $Sql
        if (& $Accept $value) { return $value }
        Start-Sleep -Milliseconds 500
    }
    throw "Timed out waiting for database assertion in $Database. Last value: $value"
}

$startedAt = [DateTimeOffset]::UtcNow
$runId = $startedAt.ToString('yyyyMMddHHmmss')
$departmentA = [guid]::NewGuid().ToString()
$departmentB = [guid]::NewGuid().ToString()
$operatorPassword = $env:GATE_D_ADMIN_PASSWORD + 'Op1!'
$employeePassword = $env:GATE_D_ADMIN_PASSWORD + 'Emp1!'
$otherPassword = $env:GATE_D_ADMIN_PASSWORD + 'Other1!'
$agentEmail = "gate-agent-$runId@local.test"
$employeeEmail = "gate-employee-$runId@local.test"
$otherEmail = "gate-other-$runId@local.test"
$steps = New-Object System.Collections.Generic.List[object]

function Add-Step([string]$Name, [int]$Status) {
    $steps.Add([ordered]@{ name = $Name; http_status = $Status })
}

$adminLogin = Invoke-EompRequest -Method POST -Path '/api/v1/auth/login' -ExpectedStatus 200 -Body @{
    email = $AdminEmail; password = $env:GATE_D_ADMIN_PASSWORD
}
$adminToken = $adminLogin.Data.access_token
if (-not $adminToken) { throw 'Admin login returned no access_token' }
Add-Step 'admin_login' $adminLogin.Status

$agent = Invoke-EompRequest -Method POST -Path '/api/v1/users' -ExpectedStatus 201 -Token $adminToken -Body @{
    email = $agentEmail; password = $operatorPassword; full_name = 'Gate D Agent'
    role = 'ROLE_AGENT'; department_id = $departmentA
}
Add-Step 'create_agent' $agent.Status

$employee = Invoke-EompRequest -Method POST -Path '/api/v1/users' -ExpectedStatus 201 -Token $adminToken -Body @{
    email = $employeeEmail; password = $employeePassword; full_name = 'Gate D Employee'
    role = 'ROLE_EMPLOYEE'; department_id = $departmentA
}
Add-Step 'create_employee' $employee.Status

$otherEmployee = Invoke-EompRequest -Method POST -Path '/api/v1/users' -ExpectedStatus 201 -Token $adminToken -Body @{
    email = $otherEmail; password = $otherPassword; full_name = 'Gate D Other Department'
    role = 'ROLE_EMPLOYEE'; department_id = $departmentB
}
Add-Step 'create_cross_scope_employee' $otherEmployee.Status

$agentLogin = Invoke-EompRequest -Method POST -Path '/api/v1/auth/login' -ExpectedStatus 200 -Body @{
    email = $agentEmail; password = $operatorPassword
}
$employeeLogin = Invoke-EompRequest -Method POST -Path '/api/v1/auth/login' -ExpectedStatus 200 -Body @{
    email = $employeeEmail; password = $employeePassword
}
$otherLogin = Invoke-EompRequest -Method POST -Path '/api/v1/auth/login' -ExpectedStatus 200 -Body @{
    email = $otherEmail; password = $otherPassword
}
Add-Step 'role_logins' 200

$ticket = Invoke-EompRequest -Method POST -Path '/api/v1/tickets' -ExpectedStatus 201 -Token $employeeLogin.Data.access_token -Body @{
    title = "Gate D deployed lifecycle $runId"
    description = 'Created through the real Gateway and Helpdesk containers.'
    category = 'Cloud & DevOps'; priority = 'HIGH'
}
$ticketId = [string]$ticket.Data.id
$ticketNumber = [string]$ticket.Data.ticket_number
$parsedTicketId = [guid]::Empty
if (-not [guid]::TryParse($ticketId, [ref]$parsedTicketId)) { throw "Invalid ticket id returned: $ticketId" }
if ($ticket.Data.department_id -ne $departmentA) { throw 'Ticket department was not derived from the employee JWT' }
Add-Step 'employee_create_ticket' $ticket.Status

$spoof = Invoke-EompRequest -Method GET -Path "/api/v1/tickets/$ticketId" -ExpectedStatus 401 -AdditionalHeaders @{
    'X-User-ID' = $agent.Data.id; 'X-User-Role' = 'ROLE_ADMIN'
}
Add-Step 'identity_header_spoof_rejected' $spoof.Status

$crossScope = Invoke-EompRequest -Method GET -Path "/api/v1/tickets/$ticketId" -ExpectedStatus 404 -Token $otherLogin.Data.access_token
Add-Step 'cross_department_probe_hidden' $crossScope.Status

$assigned = Invoke-EompRequest -Method PATCH -Path "/api/v1/tickets/$ticketId/assign" -ExpectedStatus 200 -Token $agentLogin.Data.access_token -Body @{
    assignee_id = $agent.Data.id; assignee_name = 'Gate D Agent'; version = [int]$ticket.Data.version
}
Add-Step 'agent_self_assign' $assigned.Status

$comment = Invoke-EompRequest -Method POST -Path "/api/v1/tickets/$ticketId/comments" -ExpectedStatus 201 -Token $agentLogin.Data.access_token -Body @{
    content = 'Gate D real-stack comment'; is_internal = $true
}
Add-Step 'agent_comment' $comment.Status

$inProgress = Invoke-EompRequest -Method PATCH -Path "/api/v1/tickets/$ticketId/status" -ExpectedStatus 200 -Token $agentLogin.Data.access_token -Body @{
    status = 'IN_PROGRESS'; notes = 'Work started'; version = [int]$assigned.Data.version
}
Add-Step 'status_in_progress' $inProgress.Status

$conflict = Invoke-EompRequest -Method PATCH -Path "/api/v1/tickets/$ticketId/status" -ExpectedStatus 409 -Token $agentLogin.Data.access_token -Body @{
    status = 'RESOLVED'; notes = 'Stale write must fail'; version = [int]$assigned.Data.version
}
Add-Step 'optimistic_lock_conflict' $conflict.Status

$resolved = Invoke-EompRequest -Method PATCH -Path "/api/v1/tickets/$ticketId/status" -ExpectedStatus 200 -Token $agentLogin.Data.access_token -Body @{
    status = 'RESOLVED'; notes = 'Gate D lifecycle completed'; version = [int]$inProgress.Data.version
}
if ($resolved.Data.status -ne 'RESOLVED') { throw 'Ticket did not reach RESOLVED' }
Add-Step 'status_resolved' $resolved.Status

$refresh = Invoke-EompRequest -Method POST -Path '/api/v1/auth/refresh' -ExpectedStatus 200 -Body @{
    refresh_token = $employeeLogin.Data.refresh_token
}
Add-Step 'refresh_rotation' $refresh.Status
$logout = Invoke-EompRequest -Method POST -Path '/api/v1/auth/logout' -ExpectedStatus 200 -Body @{
    refresh_token = $refresh.Data.refresh_token
}
Add-Step 'logout_revoke' $logout.Status
$revokedRefresh = Invoke-EompRequest -Method POST -Path '/api/v1/auth/refresh' -ExpectedStatus 401 -Body @{
    refresh_token = $refresh.Data.refresh_token
}
Add-Step 'revoked_refresh_rejected' $revokedRefresh.Status

$ticketSql = "SELECT status || '|' || version::text || '|' || COALESCE(assignee_id,'') || '|' || department_id FROM tickets WHERE id = '$ticketId';"
$ticketState = Invoke-DatabaseScalar -Database 'helpdesk_db' -Sql $ticketSql
$expectedState = "RESOLVED|$($resolved.Data.version)|$($agent.Data.id)|$departmentA"
if ($ticketState -ne $expectedState) { throw "Unexpected persisted ticket state: $ticketState" }
$commentCount = Invoke-DatabaseScalar -Database 'helpdesk_db' -Sql "SELECT COUNT(*) FROM ticket_comments WHERE ticket_id = '$ticketId';"
$timelineCount = Invoke-DatabaseScalar -Database 'helpdesk_db' -Sql "SELECT COUNT(*) FROM ticket_timeline WHERE ticket_id = '$ticketId' AND action IN ('TICKET_CREATED','ASSIGNED','COMMENT_ADDED','STATUS_CHANGED');"
if ([int]$commentCount -lt 1 -or [int]$timelineCount -lt 5) { throw 'Helpdesk persistence assertions failed' }

$auditCount = Wait-DatabaseValue -Database 'audit_db' -Sql "SELECT COUNT(*) FROM audit_logs WHERE resource_id = '$ticketId';" -Accept { param($value) [int]$value -ge 4 }
$reportingState = Wait-DatabaseValue -Database 'reporting_db' -Sql "SELECT status FROM raw_incident_records WHERE ticket_number = '$ticketNumber';" -Accept { param($value) $value -eq 'RESOLVED' }
$notificationCount = Wait-DatabaseValue -Database 'notification_db' -Sql "SELECT COUNT(*) FROM notifications WHERE recipient_id = '$($employee.Data.id)';" -Accept { param($value) [int]$value -ge 1 }

$revision = (& git rev-parse HEAD).Trim()
$evidence = [ordered]@{
    schema_version = 1
    result = 'PASS'
    scope = 'local Docker deployed-stack E2E; not TLS/staging acceptance'
    started_at_utc = $startedAt.ToString('o')
    completed_at_utc = [DateTimeOffset]::UtcNow.ToString('o')
    source_revision = $revision
    base_url = $BaseUrl
    ticket_id = $ticketId
    ticket_number = $ticketNumber
    final_status = [string]$resolved.Data.status
    final_version = [int]$resolved.Data.version
    assertions = [ordered]@{
        helpdesk_comment_count = [int]$commentCount
        helpdesk_timeline_count = [int]$timelineCount
        audit_event_count = [int]$auditCount
        reporting_projection_status = $reportingState
        notification_count = [int]$notificationCount
    }
    steps = $steps
}

$fullEvidencePath = Join-Path (Get-Location) $EvidencePath
$parent = Split-Path -Parent $fullEvidencePath
New-Item -ItemType Directory -Force -Path $parent | Out-Null
$evidence | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $fullEvidencePath -Encoding UTF8
Write-Host "Gate D deployed-stack E2E PASS: $ticketNumber"
Write-Host "Evidence: $fullEvidencePath"
