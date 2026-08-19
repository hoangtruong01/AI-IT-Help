package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"eomp/services/ai/internal/model"
)

// MockProvider implements both LLMProvider and EmbeddingProvider with domain-specific IT intelligence.
type MockProvider struct{}

// NewMockProvider creates an intelligent domain mock provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) Generate(ctx context.Context, messages []model.Message) (string, error) {
	if len(messages) == 0 {
		return "Hello! I am your EOMP AI Operations Copilot. How can I assist you with IT support, incident triage, or runbooks today?", nil
	}

	lastMsg := messages[len(messages)-1].Content
	queryLower := strings.ToLower(lastMsg)

	// Case 1: MFA / 2FA / Okta Token Reset (Test Case 6.1)
	if strings.Contains(queryLower, "mfa") || strings.Contains(queryLower, "2fa") || strings.Contains(queryLower, "okta") || (strings.Contains(queryLower, "reset") && strings.Contains(queryLower, "token")) {
		return "### 🛡️ Standard Operating Procedure: User MFA & Okta Token Reset\n\n" +
			"Based on internal documentation (**Doc: How to Reset User MFA and Okta Verify Tokens** and **Runbook: RB-SEC-02**), please follow these mandatory security steps:\n\n" +
			"1. **Mandatory Identity Verification**:\n" +
			"   - Verify employee identity via secondary corporate channel (live video call or manager approval email).\n" +
			"   - Confirm active employment status in Employee Directory.\n\n" +
			"2. **Okta Admin Reset**:\n" +
			"   - Navigate to Identity Admin Portal at `https://id.eomp.local/admin`.\n" +
			"   - Search for user by corporate email.\n" +
			"   - Go to **Authenticators / Factors** and click **Reset Multi-Factor Authentication** (or *Clear Okta Verify Enrollment*).\n\n" +
			"3. **Re-enrollment**:\n" +
			"   - Generate a **15-minute One-Time Temporary Activation Link** or enrollment QR code.\n" +
			"   - Instruct user to scan QR code using Okta Verify or Google Authenticator on their verified mobile device.\n\n" +
			"4. **Verification & Closure**:\n" +
			"   - Request user to perform test login.\n" +
			"   - Once confirmed, log incident resolution in Helpdesk.\n\n" +
			"*Reference: RB-SEC-02 (User MFA Token Reset and Identity Verification SOP)*", nil
	}

	// Case 2: VPN / Network Connection Issues
	if strings.Contains(queryLower, "vpn") || strings.Contains(queryLower, "wireguard") || strings.Contains(queryLower, "tunnel") || strings.Contains(queryLower, "staging server") {
		return "### 🌐 Diagnostic & Resolution Guide: Corporate VPN Disconnection\n\n" +
			"Based on **Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide** and **Runbook: RB-NET-01**:\n\n" +
			"1. **Root Cause Analysis**:\n" +
			"   - Potential WireGuard handshake timeout on gateway subnet (`10.8.0.1`).\n" +
			"   - MTU mismatch (default MTU should be set to `1380` for corporate tunnels).\n" +
			"   - Subnet routing conflict with local router DHCP range.\n\n" +
			"2. **Immediate Remediation**:\n" +
			"   - Verify tunnel ping: `ping 10.8.0.1`\n" +
			"   - Restart local WireGuard daemon:\n" +
			"     `Restart-Service -Name \"WireGuardTunnel$eomp\"`\n" +
			"   - Flush local DNS and route cache:\n" +
			"     `ipconfig /flushdns`\n\n" +
			"3. **Escalation**:\n" +
			"   - If gateway is unresponsive, trigger **RB-NET-01 (Emergency VPN Tunnel Failover SOP)** to route traffic via backup cluster (`10.8.1.1`).", nil
	}

	// Case 3: PostgreSQL Database Connection Pool
	if strings.Contains(queryLower, "postgres") || strings.Contains(queryLower, "database") || strings.Contains(queryLower, "pool") || strings.Contains(queryLower, "too many clients") {
		return "### 🗄️ Database Incident Guide: Connection Pool Exhaustion Recovery\n\n" +
			"Based on **PostgreSQL Connection Pool Exhaustion Policy** and **Runbook: RB-DB-03**:\n\n" +
			"1. **Immediate Inspection**:\n" +
			"   - Execute activity query to find blocking or leaking connections:\n" +
			"     `SELECT pid, usename, client_addr, state, query_start FROM pg_stat_activity WHERE state = 'idle in transaction' AND query_start < NOW() - INTERVAL '5 minutes';`\n\n" +
			"2. **Remediation**:\n" +
			"   - Terminate hanging idle queries:\n" +
			"     `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle in transaction' AND query_start < NOW() - INTERVAL '5 minutes';`\n" +
			"   - Verify that microservice connection pool settings do not exceed `SetMaxOpenConns(25)`.\n\n" +
			"*Reference: RB-DB-03 (PostgreSQL Connection Pool Exhaustion Recovery)*", nil
	}

	// Case 4: Laptop Provisioning & Setup
	if strings.Contains(queryLower, "laptop") || strings.Contains(queryLower, "provision") || strings.Contains(queryLower, "macbook") || strings.Contains(queryLower, "thinkpad") {
		return "### 💻 Hardware Baseline & Provisioning SOP\n\n" +
			"Based on **Standard Laptop Setup & Security Baseline** and **Runbook: RB-HW-04**:\n\n" +
			"1. **Encryption & Security**:\n" +
			"   - Enable **FileVault** (macOS) or **BitLocker** with TPM 2.0 (Windows).\n" +
			"   - Escrow recovery keys into CMDB (`/assets`).\n" +
			"   - Install **CrowdStrike EDR** endpoint security agent.\n\n" +
			"2. **Software Stack**:\n" +
			"   - Install Root Corporate CA Certificate.\n" +
			"   - Developer tooling: Git, Docker Desktop, Go 1.23, Node.js 20 LTS.\n\n" +
			"3. **Handover**:\n" +
			"   - Update asset status to `IN_USE` in CMDB and record recipient employee ID.", nil
	}

	// Default intelligent IT Copilot response
	return fmt.Sprintf("### 🤖 AI Operations Copilot Analysis\n\n"+
		"I have analyzed your inquiry: **\"%s\"** against the EOMP IT Knowledge Base and 11 microservice architectures.\n\n"+
		"**Recommended Actions:**\n"+
		"1. Check the relevant technical documentation in the **Knowledge Base** (`/knowledge`).\n"+
		"2. Verify service health status on the Gateway and microservice dashboards.\n"+
		"3. If this relates to an ongoing incident, use the **AI Auto-Triage** tool to classify the ticket and retrieve exact SOP Runbooks.\n\n"+
		"*Let me know if you would like me to retrieve specific commands or execute a diagnostic scan.*", lastMsg), nil
}

func (m *MockProvider) AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error) {
	combined := strings.ToLower(title + " " + description)

	// 1. VPN / Network Failure
	if strings.Contains(combined, "vpn") || strings.Contains(combined, "wireguard") || strings.Contains(combined, "network") || strings.Contains(combined, "connect") || strings.Contains(combined, "staging server") {
		return &model.TicketAnalysis{
			TicketID:            fmt.Sprintf("AI-TK-%d", time.Now().Unix()%10000),
			SuggestedCategory:   "Network & Access",
			Priority:            "HIGH",
			Summary:             "VPN tunnel connection timeout to Staging Server cluster.",
			RootCause:           "WireGuard handshake packet drop or MTU mismatch on Gateway subnet.",
			SuggestedResolution: "1. Instruct user to flush DNS cache and verify client MTU is 1380.\n2. Verify upstream VPN gateway status on 10.8.0.1.\n3. Refer to Runbook RB-NET-01 if gateway failover is required.",
			Confidence:          0.94,
			Citations: []model.Citation{
				{ArticleID: "a0000000-0000-0000-0000-000000000002", Title: "Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide", Score: 0.95, Category: "Network & Access", Type: "article"},
				{ArticleID: "r0000000-0000-0000-0000-000000000002", Title: "RB-NET-01: Emergency VPN Tunnel Failover SOP", Score: 0.94, Category: "Network & Access", Type: "runbook"},
				{ArticleID: "a0000000-0000-0000-0000-000000000006", Title: "Resolving Gateway DNS Resolution Timeout and Subnet Routing", Score: 0.88, Category: "Network & Access", Type: "article"},
			},
			RequiresHumanReview: true,
			CreatedAt:           time.Now(),
		}, nil
	}

	// 2. MFA / Access Lockout
	if strings.Contains(combined, "mfa") || strings.Contains(combined, "okta") || strings.Contains(combined, "2fa") || strings.Contains(combined, "token") || strings.Contains(combined, "login") || strings.Contains(combined, "auth") {
		return &model.TicketAnalysis{
			TicketID:            fmt.Sprintf("AI-TK-%d", time.Now().Unix()%10000),
			SuggestedCategory:   "IT Security & Access",
			Priority:            "HIGH",
			Summary:             "Employee authentication failure due to desynchronized or missing MFA token.",
			RootCause:           "Lost authenticator device or expired Okta Verify enrollment.",
			SuggestedResolution: "1. Perform secondary identity verification via video call or manager note.\n2. Access Okta Admin and issue a 15-minute temporary enrollment QR code.\n3. Verify test login before closing ticket.",
			Confidence:          0.96,
			Citations: []model.Citation{
				{ArticleID: "a0000000-0000-0000-0000-000000000001", Title: "How to Reset User MFA and Okta Verify Tokens", Score: 0.96, Category: "IT Security & Access", Type: "article"},
				{ArticleID: "r0000000-0000-0000-0000-000000000001", Title: "RB-SEC-02: User MFA Token Reset and Identity Verification SOP", Score: 0.95, Category: "IT Security & Access", Type: "runbook"},
			},
			RequiresHumanReview: true,
			CreatedAt:           time.Now(),
		}, nil
	}

	// 3. Database / Backend Outage
	if strings.Contains(combined, "postgres") || strings.Contains(combined, "database") || strings.Contains(combined, "pool") || strings.Contains(combined, "500") || strings.Contains(combined, "crash") {
		return &model.TicketAnalysis{
			TicketID:            fmt.Sprintf("AI-TK-%d", time.Now().Unix()%10000),
			SuggestedCategory:   "DevOps & Infrastructure",
			Priority:            "URGENT",
			Summary:             "Database connection pool exhaustion impacting microservice API availability.",
			RootCause:           "High volume of unclosed transactions holding connection slots in target PostgreSQL instance.",
			SuggestedResolution: "1. Inspect pg_stat_activity for idle-in-transaction queries.\n2. Terminate blocking backend PIDs using pg_terminate_backend.\n3. Restart affected microservice pods if connection leak persists.",
			Confidence:          0.95,
			Citations: []model.Citation{
				{ArticleID: "a0000000-0000-0000-0000-000000000004", Title: "PostgreSQL Database Connection Pool Exhaustion Recovery Policy", Score: 0.95, Category: "DevOps & Infrastructure", Type: "article"},
				{ArticleID: "r0000000-0000-0000-0000-000000000003", Title: "RB-DB-03: PostgreSQL Connection Pool Exhaustion Recovery", Score: 0.94, Category: "DevOps & Infrastructure", Type: "runbook"},
			},
			RequiresHumanReview: true,
			CreatedAt:           time.Now(),
		}, nil
	}

	// 4. Laptop / Hardware Replacement
	if strings.Contains(combined, "laptop") || strings.Contains(combined, "hardware") || strings.Contains(combined, "monitor") || strings.Contains(combined, "screen") || strings.Contains(combined, "macbook") {
		return &model.TicketAnalysis{
			TicketID:            fmt.Sprintf("AI-TK-%d", time.Now().Unix()%10000),
			SuggestedCategory:   "Hardware & Equipment",
			Priority:            "MEDIUM",
			Summary:             "Hardware provisioning or replacement request for employee workstation.",
			RootCause:           "New employee onboarding or hardware component failure.",
			SuggestedResolution: "1. Verify asset warranty status in CMDB (/assets).\n2. Provision replacement hardware from IN_STOCK inventory.\n3. Apply corporate security baseline (FileVault, EDR agent) and perform handover.",
			Confidence:          0.92,
			Citations: []model.Citation{
				{ArticleID: "a0000000-0000-0000-0000-000000000003", Title: "Standard Laptop Setup & Security Baseline (macOS & Windows)", Score: 0.93, Category: "Hardware & Equipment", Type: "article"},
				{ArticleID: "r0000000-0000-0000-0000-000000000004", Title: "RB-HW-04: Corporate Laptop Provisioning and Deployment Procedure", Score: 0.91, Category: "Hardware & Equipment", Type: "runbook"},
			},
			RequiresHumanReview: true,
			CreatedAt:           time.Now(),
		}, nil
	}

	// 5. Default General IT Support Triage
	return &model.TicketAnalysis{
		TicketID:            fmt.Sprintf("AI-TK-%d", time.Now().Unix()%10000),
		SuggestedCategory:   "IT Support",
		Priority:            "MEDIUM",
		Summary:             fmt.Sprintf("Operational issue: %s", title),
		RootCause:           "General user technical incident requiring standard troubleshooting.",
		SuggestedResolution: "1. Gather client log files and verify network routing.\n2. Compare issue against Knowledge Base standard operating procedures.\n3. Assign to Tier-1 Support Agent for resolution.",
		Confidence:          0.88,
		Citations: []model.Citation{
			{ArticleID: "a0000000-0000-0000-0000-000000000001", Title: "How to Reset User MFA and Okta Verify Tokens", Score: 0.85, Category: "IT Security & Access", Type: "article"},
			{ArticleID: "a0000000-0000-0000-0000-000000000002", Title: "Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide", Score: 0.84, Category: "Network & Access", Type: "article"},
		},
		RequiresHumanReview: true,
		CreatedAt:           time.Now(),
	}, nil
}

func (m *MockProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, 768)
	return vec, nil
}

func (m *MockProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, 768)
	}
	return result, nil
}

func (m *MockProvider) Dimensions() int {
	return 768
}
