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
| auth | `auth_db` | Provisioned (Active) |
| employee | `employee_db` | Provisioned (Active) |
| asset | `asset_db` | Provisioned (Active) |
| helpdesk | `helpdesk_db` | Provisioned (Active) |
| workflow | `workflow_db` | Provisioned (Active) |
| knowledge | `knowledge_db` | Provisioned (Active) |
| audit | `audit_db` | Provisioned (Active) |

## Migration Tool

To be determined when first module begins development.
