# EOMP — Database

> Database documentation will be expanded as modules are implemented.

## Strategy

- Each microservice owns its database
- No shared database access between services
- Database schema managed per service via migrations
- Databases will be created when each module begins development

## Planned Databases

| Service | Database | Status |
|---|---|---|
| auth | `auth_db` | Planned |
| employee | `employee_db` | Planned |
| asset | `asset_db` | Planned |
| helpdesk | `helpdesk_db` | Planned |
| workflow | `workflow_db` | Planned |
| knowledge | `knowledge_db` | Planned |
| audit | `audit_db` | Planned |

## Migration Tool

To be determined when first module begins development.
