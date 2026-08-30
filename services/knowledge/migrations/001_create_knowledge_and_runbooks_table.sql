-- =============================================================================
-- Migration: 001_create_knowledge_and_runbooks_table.sql
-- Module: Knowledge Service (knowledge_db)
-- Description: Knowledge base articles, categories, SOP runbooks, and vector embeddings
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Knowledge Categories Table
CREATE TABLE IF NOT EXISTS knowledge_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    icon VARCHAR(50) NOT NULL DEFAULT 'i-lucide-folder',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Knowledge Base Articles Table
CREATE TABLE IF NOT EXISTS knowledge_articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES knowledge_categories(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    summary TEXT NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    author_id VARCHAR(100) NOT NULL DEFAULT 'system',
    author_name VARCHAR(150) NOT NULL DEFAULT 'IT System Admin',
    view_count INT NOT NULL DEFAULT 0,
    helpful_count INT NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Standard Operating Procedures (SOP) Runbooks Table
CREATE TABLE IF NOT EXISTS runbooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    prerequisites TEXT NOT NULL,
    steps_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    rollback_steps TEXT NOT NULL,
    author_name VARCHAR(150) NOT NULL DEFAULT 'IT Security Team',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Document Embeddings / Chunks Reference Table (Sync with Qdrant)
CREATE TABLE IF NOT EXISTS document_embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    article_id UUID NOT NULL REFERENCES knowledge_articles(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL DEFAULT 0,
    chunk_text TEXT NOT NULL,
    embedding_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Fulltext & Performance Indexes
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_category ON knowledge_articles(category_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_published ON knowledge_articles(is_published);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_created_at ON knowledge_articles(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_runbooks_category ON runbooks(category);
CREATE INDEX IF NOT EXISTS idx_runbooks_code ON runbooks(code);
CREATE INDEX IF NOT EXISTS idx_document_embeddings_article ON document_embeddings(article_id);

-- =============================================================================
-- SEED DATA: Knowledge Categories, SOP Articles & Runbooks
-- =============================================================================

INSERT INTO knowledge_categories (id, name, code, icon, description)
VALUES 
    ('c0000000-0000-0000-0000-000000000001', 'IT Security & Access', 'sec', 'i-lucide-shield-check', 'MFA tokens, SSO identity, zero-trust security & access management policies.'),
    ('c0000000-0000-0000-0000-000000000002', 'Network & Connectivity', 'net', 'i-lucide-network', 'WireGuard VPN, GlobalProtect, DNS routing, firewall rules & Wi-Fi setup.'),
    ('c0000000-0000-0000-0000-000000000003', 'Hardware & Equipment', 'hw', 'i-lucide-laptop', 'Laptop standard baseline, monitor setup, warranty claims & hardware replacement.'),
    ('c0000000-0000-0000-0000-000000000004', 'DevOps & Cloud Infrastructure', 'devops', 'i-lucide-server', 'PostgreSQL database access, Kubernetes clusters, MinIO storage & staging secrets.'),
    ('c0000000-0000-0000-0000-000000000005', 'Software & Productivity Apps', 'soft', 'i-lucide-layers', 'Operating systems, Slack, Microsoft 365, Docker Desktop & developer tooling.')
ON CONFLICT (code) DO NOTHING;

INSERT INTO knowledge_articles (id, category_id, title, slug, summary, content, tags, author_name, view_count, is_published)
VALUES 
    (
        'a0000000-0000-0000-0000-000000000001',
        'c0000000-0000-0000-0000-000000000001',
        'How to Reset User MFA and Okta Verify Tokens',
        'how-to-reset-user-mfa-tokens',
        'Official standard operating procedure for IT Support Agents to verify employee identity and securely reset multi-factor authentication tokens in Okta / Keycloak.',
        '# How to Reset User MFA and Okta Verify Tokens

## Overview
This document outlines the standard operating procedure (SOP) for IT Support Agents (`ROLE_AGENT`) when an employee requests an MFA or OTP token reset due to a lost device, hardware replacement, or authentication lock.

## Security Verification Checklist (Mandatory)
1. Verify the employee identity via live video call or manager approval email.
2. Confirm employee ID, department, and active employment status in Employee Directory.
3. Never send temporary backup codes via unencrypted public channels.

## Step-by-Step Resolution:
1. Log into the **Okta / Identity Admin Portal** at `https://id.eomp.local/admin`.
2. Search for the target user by corporate email (`user@eomp.local`).
3. Click on the user profile and navigate to the **Authenticators / Factors** tab.
4. Select **Reset Multi-Factor Authentication** or click **Clear Okta Verify Enrollment**.
5. Issue a **One-Time Temporary Activation Link** (valid for 15 minutes) and send directly to their verified internal inbox or via Manager escalation.
6. Instruct the user to open the activation link on their new mobile device and scan the generated QR code using Okta Verify / Google Authenticator.
7. Confirm successful test login before closing the ticket.

## Associated Runbook:
Refer to `RB-SEC-02: User MFA Token Reset and Identity Verification SOP` for emergency escalation.',
        ARRAY['MFA', 'Okta', 'Security', 'Authentication', 'SOP'],
        'Sarah Jenkins (IT Security Lead)',
        1240,
        TRUE
    ),
    (
        'a0000000-0000-0000-0000-000000000002',
        'c0000000-0000-0000-0000-000000000002',
        'Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide',
        'vpn-troubleshooting-guide',
        'Comprehensive resolution guide for remote engineers experiencing VPN disconnects, handshake timeouts, and MTU routing packet loss.',
        '# Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide

## Symptoms & Root Causes
- **Handshake Timeout**: Public IP address blocked by ISP firewall or stale WireGuard session.
- **VPN drops every 10 minutes**: Gateway MTU misconfiguration or dead peer detection timeout.
- **DNS Resolver Failure**: Subnet routing table conflict between local LAN and corporate staging gateway.

## Diagnostic Procedures
1. Test tunnel reachability:
   ```bash
   ping -c 4 10.8.0.1
   traceroute 10.8.0.1
   ```
2. Verify WireGuard client configuration:
   ```ini
   [Interface]
   PrivateKey = <employee_private_key>
   Address = 10.8.0.42/24
   DNS = 10.8.0.2
   MTU = 1380

   [Peer]
   PublicKey = <gateway_public_key>
   Endpoint = vpn.eomp.local:51820
   AllowedIPs = 10.8.0.0/16, 172.20.0.0/16
   PersistentKeepalive = 25
   ```
3. Restart local WireGuard daemon:
   ```powershell
   # Windows PowerShell
   Restart-Service -Name "WireGuardTunnel$eomp"
   ```
4. If DNS lookup fails for internal microservices, flush DNS cache:
   ```bash
   # macOS: sudo dscacheutil -flushcache; sudo killall -HUP mDNSResponder
   # Windows: ipconfig /flushdns
   ```',
        ARRAY['VPN', 'WireGuard', 'Network', 'DNS', 'GlobalProtect'],
        'Alex Rivera (Network Architect)',
        980,
        TRUE
    ),
    (
        'a0000000-0000-0000-0000-000000000003',
        'c0000000-0000-0000-0000-000000000003',
        'Standard Laptop Setup & Security Baseline (macOS & Windows)',
        'standard-laptop-setup-baseline',
        'Full checklist for provisioning MacBook Pro and ThinkPad laptops for new hires, including FileVault, BitLocker, EDR agent, and developer tooling.',
        '# Standard Laptop Setup & Security Baseline (macOS & Windows)

## Provisioning Checklist
1. **Full Disk Encryption**:
   - macOS: Enable FileVault and escrow recovery key in CMDB.
   - Windows: Enable BitLocker with TPM 2.0 PIN protection.
2. **Endpoint Detection & Response (EDR)**:
   - Install CrowdStrike / SentinelOne EDR agent.
   - Verify agent reporting status at `https://edr.eomp.local`.
3. **Identity & Certificates**:
   - Install Root Corporate CA Certificate.
   - Enroll device into MDM profile (Jamf for Mac, Intune for Windows).
4. **Developer Stack**:
   - Install Docker Desktop, Git, Go 1.23, Node.js 20 LTS.
   - Generate corporate SSH key: `ssh-keygen -t ed25519 -C "user@eomp.local"`.',
        ARRAY['Hardware', 'Laptop', 'Setup', 'Security', 'Provisioning'],
        'David Chen (IT Asset Lead)',
        850,
        TRUE
    ),
    (
        'a0000000-0000-0000-0000-000000000004',
        'c0000000-0000-0000-0000-000000000004',
        'PostgreSQL Database Connection Pool Exhaustion Recovery Policy',
        'postgres-connection-pool-recovery',
        'Emergency diagnostic runbook for handling "sorry, too many clients already" and connection spikes across the 7 microservice databases.',
        '# PostgreSQL Database Connection Pool Exhaustion Recovery

## Emergency Symptoms
- HTTP 500 error code across Gateway: `pq: remaining connection slots are reserved for non-replication superuser connections`.
- Service healthchecks reporting timeout on database ping.

## Remediation Steps:
1. Identify idle or leaking connections in target database:
   ```sql
   SELECT pid, usename, client_addr, state, query_start, query 
   FROM pg_stat_activity 
   WHERE state != ''idle'' AND query NOT ILIKE ''%pg_stat_activity%''
   ORDER BY query_start ASC;
   ```
2. Terminate rogue blocking queries:
   ```sql
   SELECT pg_terminate_backend(pid) 
   FROM pg_stat_activity 
   WHERE state = ''idle in transaction'' 
     AND query_start < NOW() - INTERVAL ''5 minutes'';
   ```
3. Verify connection pool parameters in service config (`SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`).',
        ARRAY['Database', 'PostgreSQL', 'Performance', 'DevOps', 'Troubleshooting'],
        'Sarah Jenkins (DevOps Lead)',
        640,
        TRUE
    ),
    (
        'a0000000-0000-0000-0000-000000000005',
        'c0000000-0000-0000-0000-000000000003',
        'Hardware Replacement & Warranty Claim Workflow',
        'hardware-replacement-warranty-workflow',
        'Step-by-step procedure for replacing damaged monitors, swollen laptop batteries, and processing manufacturer RMA claims.',
        '# Hardware Replacement & Warranty Claim Workflow

## Eligibility & Lifecycle
1. Inspect asset in CMDB (`/assets`) to check `warranty_expiry` date and serial number.
2. If asset is damaged under active warranty, initiate manufacturer RMA ticket (Dell ProSupport / AppleCare Enterprise).
3. Temporary Loaner Device:
   - Assign temporary replacement asset from `IN_STOCK` pool with status `ASSIGNED`.
   - Update asset assignment log in `/assets` with handover condition checklist.',
        ARRAY['Asset', 'Hardware', 'Warranty', 'RMA', 'CMDB'],
        'David Chen (IT Asset Lead)',
        420,
        TRUE
    ),
    (
        'a0000000-0000-0000-0000-000000000006',
        'c0000000-0000-0000-0000-000000000002',
        'Resolving Gateway DNS Resolution Timeout and Subnet Routing',
        'dns-timeout-subnet-routing',
        'Troubleshooting guide for DNS resolver timeout on gateway subnet and updating cached routing tables.',
        '# Resolving Gateway DNS Resolution Timeout and Subnet Routing

## Root Cause Analysis
DNS resolver timeouts on the Gateway subnet occur when upstream recursive resolvers fail to respond within the 2000ms threshold or when local iptables masquerading rules encounter packet drops.

## Fix:
1. Reload local CoreDNS / systemd-resolved configuration.
2. Flush cached routing table:
   ```bash
   sudo ip route flush cache
   ```
3. Switch primary fallback resolver to `1.1.1.1` and `8.8.8.8`.',
        ARRAY['Gateway', 'DNS', 'Network', 'Subnet', 'Routing'],
        'Alex Rivera (Network Architect)',
        310,
        TRUE
    )
ON CONFLICT (slug) DO NOTHING;

INSERT INTO runbooks (id, code, title, category, description, prerequisites, steps_json, rollback_steps, author_name, is_active)
VALUES 
    (
        'b0000000-0000-0000-0000-000000000001',
        'RB-SEC-02',
        'User MFA Token Reset and Identity Verification SOP',
        'IT Security',
        'Standardized operational procedure for identity authentication and Okta/Keycloak multi-factor token reissue.',
        '1. Active IT Support Agent credentials.\n2. Manager written approval or video verification log.\n3. User employee ID and corporate email.',
        '[
            {"step": 1, "action": "Verify Employee Identity via secondary communication channel", "command": "Check employee record in /employees directory", "expected": "Active status confirmed"},
            {"step": 2, "action": "Open Okta / Auth Admin console", "command": "Navigate to /admin/users/{email}/factors", "expected": "List of active factors displayed"},
            {"step": 3, "action": "Revoke existing MFA token registration", "command": "POST /api/v1/auth/mfa/revoke", "expected": "Token status changed to REVOKED"},
            {"step": 4, "action": "Generate temporary one-time enrollment QR code", "command": "POST /api/v1/auth/mfa/re-enroll", "expected": "15-minute activation token generated"},
            {"step": 5, "action": "Send activation link securely and verify test authentication", "command": "Test login flow on staging portal", "expected": "MFA prompt succeeds with new device"}
        ]'::jsonb,
        'If user fails verification, immediately lock account for 30 minutes and notify Security Operations Center (SOC).',
        'Sarah Jenkins (IT Security Lead)',
        TRUE
    ),
    (
        'b0000000-0000-0000-0000-000000000002',
        'RB-NET-01',
        'Emergency VPN Tunnel Failover SOP',
        'Network',
        'Failover procedure when primary WireGuard VPN gateway server exhibits packet loss or hardware crash.',
        '1. Access to Secondary VPN Gateway (10.8.1.1).\n2. Cloudflare DNS admin permissions for vpn.eomp.local.',
        '[
            {"step": 1, "action": "Check primary VPN tunnel status", "command": "wg show", "expected": "Identify disconnected peers and transfer stats"},
            {"step": 2, "action": "Switch DNS endpoint record to secondary cluster", "command": "cf-cli dns update vpn.eomp.local --ip 198.51.100.2", "expected": "DNS propagated within 60s"},
            {"step": 3, "action": "Restart WireGuard client daemon on connected hosts", "command": "systemctl restart wg-quick@wg0", "expected": "Tunnel handshake establishes with backup gateway"}
        ]'::jsonb,
        'Revert DNS A record back to primary gateway IP once upstream network link recovers.',
        'Alex Rivera (Network Architect)',
        TRUE
    ),
    (
        'b0000000-0000-0000-0000-000000000003',
        'RB-DB-03',
        'PostgreSQL Connection Pool Exhaustion Recovery',
        'DevOps',
        'Rapid mitigation procedure for resolving high database connection spikes and unblocking microservices.',
        '1. Superuser access to postgres connection.\n2. Access to Grafana database dashboard.',
        '[
            {"step": 1, "action": "Inspect active connections", "command": "SELECT count(*), state FROM pg_stat_activity GROUP BY state;", "expected": "Identify idle in transaction queries"},
            {"step": 2, "action": "Terminate stale connections older than 5 minutes", "command": "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = ''idle in transaction'' AND query_start < NOW() - INTERVAL ''5 min'';", "expected": "Connections freed"},
            {"step": 3, "action": "Restart service connection pool if connection leak persists", "command": "docker compose restart helpdesk asset", "expected": "Connection count drops below 25"}
        ]'::jsonb,
        'If termination causes data inconsistency, rollback transaction via WAL point-in-time recovery.',
        'Sarah Jenkins (DevOps Lead)',
        TRUE
    ),
    (
        'b0000000-0000-0000-0000-000000000004',
        'RB-HW-04',
        'Corporate Laptop Provisioning and Deployment Procedure',
        'Hardware',
        'Standard procedure for imaging, encrypting, and handing over corporate MacBook Pro / ThinkPad hardware.',
        '1. Hardware asset registered in CMDB with status IN_STOCK.\n2. Approved Hardware Request Workflow instance.',
        '[
            {"step": 1, "action": "Boot into Enterprise MDM NetInstall", "command": "Option + Command + R (Mac) / F12 PXE (ThinkPad)", "expected": "Corporate OS image installs"},
            {"step": 2, "action": "Enable Full Disk Encryption and upload recovery key", "command": "fdesetup enable (Mac) / manage-bde -on C: (Windows)", "expected": "Recovery key stored in CMDB"},
            {"step": 3, "action": "Install EDR Agent & Security Baseline", "command": "install-agent.sh --token $EDR_TOKEN", "expected": "Host enrolled in EDR console"},
            {"step": 4, "action": "Perform Handover to Employee in CMDB", "command": "POST /api/v1/assets/{id}/assign", "expected": "Asset status changes to IN_USE"}
        ]'::jsonb,
        'Wipe drive with DoD 5220.22-M standard if provisioning is aborted.',
        'David Chen (IT Asset Lead)',
        TRUE
    )
ON CONFLICT (code) DO NOTHING;
