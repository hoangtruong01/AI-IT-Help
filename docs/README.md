# EOMP documentation hub

The authoritative project status is [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md). The current code version is `0.1.0`; production acceptance is pending. Historical phase reports and portfolio documents are retained for context but are not release evidence.

## Start here

- [Implementation status](IMPLEMENTATION_STATUS.md): verified inventory, completed remediation, test evidence and release blockers.
- [Quickstart URLs](QUICKSTART_URLS.md): local startup and endpoint reference without published credentials.
- [Setup guide](setup.md): development setup.
- [Environment example](../.env.example): supported environment-variable names. Replace all example credentials outside an isolated local machine.
- [Handover acceptance](sre/project_handover_acceptance.md): unsigned release checklist.

## Architecture and API

- [Architecture overview](architecture.md)
- [C4 diagrams](architecture/c4_model_diagrams.md)
- [Database overview](database.md)
- [Data dictionary](architecture/database_erd_and_data_dictionary.md)
- [API overview](api.md)
- [OpenAPI specification](openapi/eomp-openapi-spec.yaml) — complete 101/101 operation inventory; richer domain schemas remain tracked work.

## Engineering and operations

- [Development guide](development.md)
- [Testing strategy](testing.md) — targets are not evidence unless a current run is attached.
- [Deployment guide](deployment.md)
- [Operations manual](sre/operations_manual.md)
- [Incident response playbook](sre/incident_response_playbook.md)
- [Disaster-recovery plan](sre/disaster_recovery_plan.md) — RPO/RTO are targets until a real restore drill produces evidence.

## Historical documents

These documents may contain outdated versions, simulated results or superseded completion claims:

- [Project structure and changelog](PROJECT_STRUCTURE_AND_CHANGELOG.md)
- [Database audit report](DATABASE_AUDIT_REPORT.md)
- [Master upgrade plan](MASTER_UPGRADE_PLAN_P0_TO_P8.md)
- [Phase 6–14 roadmap](PHASE_6_TO_14_ROADMAP_SPECIFICATION.md)
- [Phase 8 portfolio case study](PHASE_8_ENTERPRISE_VALIDATION_AND_PORTFOLIO_CASE_STUDY.md)
- [Phase 8 evidence report](sre/PHASE_8_ENTERPRISE_VALIDATION_EVIDENCE_REPORT.md)

Do not use a historical document to approve a deployment. A release requires objective results for every P0 blocker and signatures in the handover checklist.
