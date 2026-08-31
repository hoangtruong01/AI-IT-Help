# Project handover and production acceptance

Status: **pending**
Last reviewed: 2026-08-31

This document is an acceptance checklist, not a certificate. No default user credentials are distributed with the platform.

## Evidence available

- All Go module tests pass locally.
- Frontend unit tests, typecheck, lint and production build pass locally.
- Workflow authorization, requester isolation, per-recipient notification reads and CAS paths pass their local Go test/vet suites.
- Source-level security remediation is recorded in `docs/IMPLEMENTATION_STATUS.md`.
- Deployment configuration no longer stores production secret values in Git.
- Runtime/OpenAPI operation parity gate passes at 102/102 routes.

## Required before acceptance

- Rotate every credential that previously appeared in repository history.
- Provision the initial administrator through approved secret management, then remove the bootstrap password.
- Build and scan every container image; record immutable digests and SBOMs.
- Render the Helm chart and validate/apply Kubernetes resources in a non-production cluster.
- Execute database migrations and rollback procedures against disposable databases.
- Run API integration, browser E2E and load tests against the deployed stack.
- Run `scripts/backup_restore.ps1 backup` and `test-restore`; retain the generated DR evidence and verify the agreed RPO/RTO.
- Replace generic OpenAPI request/response schemas with domain contracts and run conformance tests.
- Obtain approval from Product/BA, Engineering, QA, Security and SRE owners.

## Acceptance record

| Role | Owner | Decision | Date | Evidence link |
|---|---|---|---|---|
| Product / BA | Unassigned | Pending | — | — |
| Engineering | Unassigned | Pending | — | — |
| QA | Unassigned | Pending | — | — |
| Security | Unassigned | Pending | — | — |
| SRE / Operations | Unassigned | Pending | — | — |

Production deployment is not approved while any required item or owner decision remains pending.
