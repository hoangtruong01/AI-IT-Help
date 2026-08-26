module eomp/tests/e2e

go 1.25.0

require (
	eomp/packages/shared v0.0.0
	eomp/services/ai v0.0.0
	eomp/services/asset v0.0.0
	eomp/services/audit v0.0.0
	eomp/services/auth v0.0.0
	eomp/services/employee v0.0.0
	eomp/services/gateway v0.0.0
	eomp/services/helpdesk v0.0.0
	eomp/services/knowledge v0.0.0
	eomp/services/notification v0.0.0
	eomp/services/reporting v0.0.0
	eomp/services/workflow v0.0.0
)

replace (
	eomp/packages/shared => ../../packages/shared
	eomp/services/ai => ../../services/ai
	eomp/services/asset => ../../services/asset
	eomp/services/audit => ../../services/audit
	eomp/services/auth => ../../services/auth
	eomp/services/employee => ../../services/employee
	eomp/services/gateway => ../../services/gateway
	eomp/services/helpdesk => ../../services/helpdesk
	eomp/services/knowledge => ../../services/knowledge
	eomp/services/notification => ../../services/notification
	eomp/services/reporting => ../../services/reporting
	eomp/services/workflow => ../../services/workflow
)
