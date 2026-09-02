# EOMP — Gate D evidence audit

**Audit date:** 2026-09-02  
**Source revision:** `f72aa39a35655feca29a7238b569ec7b0751e87a` plus the current working tree  
**Decision:** **GATE D NOT YET ACCEPTED — CONTROLLED PILOT IS NOT APPROVED BY THIS REPORT**

Current estimate:

- **Local technical readiness: 90%** — the complete Docker application stack, core PostgreSQL integration, deployed API lifecycle, local controlled load and database restore have passed.
- **Formal Gate D acceptance evidence: 70%** — browser E2E, CI evidence, CVE scan, real staging/TLS/network checks, WAL/full-service recovery objectives and owner sign-off remain open.

These percentages are engineering estimates, not release approval. A test file, local result or skipped tool is not proof that an environment-level acceptance criterion passed.

## Observed results

| Area | What was observed | Decision |
|---|---|---|
| Go regression | All workspace Go modules passed `go test ./...` and `go vet ./...`; targeted Gateway and Helpdesk regression also passed after the runtime fixes. | Pass as local regression evidence. |
| PostgreSQL integration | Strict execution on PostgreSQL 17.11 used four isolated temporary databases and allowed no skips. Auth, Helpdesk, Audit, migration and parameterized-query scenarios passed; temporary databases were removed. | Accepted for implemented core scenarios; D-01 remains partial because dedicated Notification, Reporting/date-filter and CI provisioning evidence are missing. |
| Deployed API E2E | Real Auth, Gateway and domain services performed login, create → assign → comment → resolve, optimistic-lock `409`, refresh rotation, logout/revocation, spoof rejection and cross-department hiding. PostgreSQL assertions confirmed Helpdesk, Audit, Reporting and Notification effects. | Pass as local deployed-stack E2E; evidence: `docs/evidence/gate-d/deployed_stack_e2e.json`. It is not browser or external staging evidence. |
| Frontend route policy | Vitest passed `81/81`; Nuxt typecheck, ESLint and production build passed. No browser was launched. | Unit/contract/build evidence only; Playwright remains open. |
| OpenAPI parity | `go run scripts/check_openapi_coverage.go` passed `107/107`. | Accepted for operation inventory only. |
| Docker application stack | One Docker Compose project named `eomp` contains 20 healthy running containers. All 12 application images were built and run as `10001:10001`; immutable local digests were recorded. | Pass as local image/runtime evidence; evidence: `docs/evidence/gate-d/image_manifest.json`. CVE scanning is separate and still open. |
| Controlled local load | 100 virtual users ran for 30 seconds against authenticated health, ticket and reporting paths: 2,928 requests, zero failures, p95 `30.3439 ms`, p99 `65.0058 ms`. | Thresholds passed locally; evidence: `docs/evidence/gate-d/local_load_summary.json`. This is not the required k6/TLS staging run. |
| Database DR | All nine PostgreSQL databases were dumped and restored into temporary databases. PostgreSQL 17.11; database-only restore `6.602s`; every restored database contained public tables. | Accepted as database restore evidence only; not WAL RPO or full-service RTO. |

## D-01 — PostgreSQL integration

Test code and real PostgreSQL execution cover Auth refresh rotation/session revocation/rollback, Helpdesk row scope/sequence/CAS, Audit integrity/tamper detection, migration concurrency and parameterized-query payloads.

Remaining acceptance work:

1. Provision the same isolated PostgreSQL databases in CI and retain the run URL, commit SHA, database version and full log.
2. Keep `INTEGRATION_REQUIRED=1` so missing or unreachable databases fail instead of skip.
3. Add dedicated Notification, Reporting and reporting date-filter integration scenarios required by the Gate D plan.

Fail-closed commands:

```powershell
$env:AUTH_INTEGRATION_DSN = '<auth test DSN>'
$env:HELPDESK_INTEGRATION_DSN = '<helpdesk test DSN>'
$env:AUDIT_INTEGRATION_DSN = '<audit test DSN>'
$env:INTEGRATION_POSTGRES_DSN = '<migration test DSN>'
.\scripts\staging_verify.ps1 -RequirePostgres
```

```bash
REQUIRE_POSTGRES=1 ./scripts/staging_verify.sh
```

## D-02 — deployed API and browser E2E

The deployed-stack portion now passes locally through the real Gateway/Auth/services and PostgreSQL. The retained evidence includes HTTP status per step and downstream database assertions. `tests/integration/http_contract_test.go` remains a fast contract harness and is not counted as that E2E proof.

Still required for formal acceptance:

- run the same deployed-stack scenario in the release/staging environment and retain its versioned report;
- add Playwright journeys for login, route guard, create ticket, `409` handling, refresh and logout;
- retain browser report/video/trace according to CI policy.

## D-03 — staging, security, load and DR

| Evidence artifact | Current result and minimum acceptance |
|---|---|
| `docs/evidence/gate-d/image_manifest.json` | **Present:** 12 application images with local immutable digests, non-root runtime user and 20 healthy containers. |
| Trivy/Scout JSON | **Missing:** Docker Scout required Docker ID login on this machine. Every release image must still be scanned with zero High/Critical findings or formally approved exceptions. |
| Helm/render/deploy log | **Missing:** chart lint/render, staging rollout, migrations and smoke tests must pass with a real TLS secret. |
| Network/TLS result | **Missing:** TLS must meet organization policy and monitoring must be unreachable from an untrusted network. |
| `docs/evidence/gate-d/local_load_summary.json` | **Present, local only:** 2,928 requests, zero failures, p95 `30.3439 ms`, p99 `65.0058 ms`. A versioned k6 summary against the real TLS staging URL is still required. |
| `docs/evidence/gate-d/dr_evidence_20260902.json` | **Present:** all nine databases restored and validated in `6.602s`. WAL-based RPO and full-service RTO remain open. |
| Owner acceptance | **Missing:** named approver, timestamp, release scope and accepted exceptions. |

The DR plan defines RPO `< 5 minutes` and full-service RTO `< 15 minutes`. Snapshot age alone does not prove WAL-based RPO, and database restore time alone does not prove full-service RTO.

## Corrections made during this audit

- Database integration helpers now fail on missing/unreachable DSNs when integration is required; real PostgreSQL execution exposed and corrected invalid UUID fixtures and incorrect revoked-token/anti-enumeration assertions.
- Gateway admin routes now use the trusted forwarded-role middleware used by the real Gateway authentication path.
- Helpdesk ticket creation derives department identity from the authenticated claim, and ticket events include requester/reporter recipients so Notification receives the event correctly.
- Service modules can build independently in Docker; 12 application images were built and verified as non-root.
- The application Compose profile uses the existing `eomp` project/network/volumes, so Docker Desktop shows one `eomp` project rather than a second `eomp-gate-d` project.
- Local load login is performed during setup rather than once per iteration, preventing authentication rate limiting from invalidating the workload.
- HTTP/Vitest suites remain labeled as contracts rather than real API/browser E2E.
- Verification scripts no longer announce “Gate D 100%” or pilot approval.

Gate D can be closed only after the remaining environment evidence is produced and reviewed; this repository must not self-approve staging or owner acceptance.
