package service_test

import (
	"context"
	"testing"
	"time"

	"eomp/packages/shared/pkg/auth"
	"eomp/services/auth/internal/model"
	"eomp/services/auth/internal/service"
)

type mockUserRepo struct {
	users     map[string]*model.User
	tokens    map[string]*tokenRecord
	auditLogs []model.LoginAuditLog
}

type tokenRecord struct {
	userID    string
	tokenHash string
	expiresAt time.Time
	revoked   bool
}

func newMockUserRepo() *mockUserRepo {
	hashedPwd, _ := auth.HashPassword("Admin@123456")
	return &mockUserRepo{
		users: map[string]*model.User{
			"admin@eomp.local": {
				ID:           "u-admin-01",
				Email:        "admin@eomp.local",
				PasswordHash: hashedPwd,
				FullName:     "System Administrator",
				Role:         "ROLE_ADMIN",
				IsActive:     true,
			},
			"disabled@eomp.local": {
				ID:           "u-disabled-01",
				Email:        "disabled@eomp.local",
				PasswordHash: hashedPwd,
				FullName:     "Disabled User",
				Role:         "ROLE_EMPLOYEE",
				IsActive:     false,
			},
		},
		tokens:    make(map[string]*tokenRecord),
		auditLogs: make([]model.LoginAuditLog, 0),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = "u-new-01"
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	m.tokens[tokenHash] = &tokenRecord{
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		revoked:   false,
	}
	return nil
}

func (m *mockUserRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if r, ok := m.tokens[tokenHash]; ok {
		r.revoked = true
	}
	return nil
}

func (m *mockUserRepo) ValidateRefreshToken(ctx context.Context, tokenHash string) (string, error) {
	if r, ok := m.tokens[tokenHash]; ok {
		if !r.revoked && r.expiresAt.After(time.Now()) {
			return r.userID, nil
		}
	}
	return "", nil
}

func (m *mockUserRepo) RecordLoginAudit(ctx context.Context, log *model.LoginAuditLog) error {
	log.ID = "audit-01"
	log.CreatedAt = time.Now()
	m.auditLogs = append(m.auditLogs, *log)
	return nil
}

func (m *mockUserRepo) GetLoginHistory(ctx context.Context, email string, limit int) ([]model.LoginAuditLog, error) {
	var res []model.LoginAuditLog
	for i := len(m.auditLogs) - 1; i >= 0; i-- {
		l := m.auditLogs[i]
		if email == "" || l.Email == email {
			res = append(res, l)
			if len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}

func TestAuthService_LoginAuditAndLogoutRevocation(t *testing.T) {
	ctx := context.Background()
	jwtManager := auth.NewJWTManager("eomp-enterprise-super-secret-jwt-key-2026", 1*time.Hour, 7*24*time.Hour)
	repo := newMockUserRepo()
	authSvc := service.NewAuthService(repo, jwtManager)

	// 1. Failed Login (Wrong Password)
	_, err := authSvc.LoginWithAudit(ctx, &model.LoginRequest{
		Email:    "admin@eomp.local",
		Password: "WrongPassword!",
	}, "192.168.1.50", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected login to fail with invalid password")
	}

	// 2. Failed Login (Account Deactivated)
	_, err = authSvc.LoginWithAudit(ctx, &model.LoginRequest{
		Email:    "disabled@eomp.local",
		Password: "Admin@123456",
	}, "192.168.1.51", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected login to fail for deactivated account")
	}

	// 3. Successful Login
	authResp, err := authSvc.LoginWithAudit(ctx, &model.LoginRequest{
		Email:    "admin@eomp.local",
		Password: "Admin@123456",
	}, "10.0.0.100", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("expected login to succeed, got: %v", err)
	}

	// 4. Verify Audit logs
	history, err := authSvc.GetLoginHistory(ctx, "admin@eomp.local", 10)
	if err != nil || len(history) < 2 {
		t.Fatalf("expected at least 2 audit logs, got %d", len(history))
	}

	// 5. Refresh token works
	refreshResp, err := authSvc.RefreshToken(ctx, &model.RefreshTokenRequest{
		RefreshToken: authResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh token failed: %v", err)
	}

	// 6. Logout and verify token is revoked
	err = authSvc.Logout(ctx, &model.LogoutRequest{
		RefreshToken: refreshResp.RefreshToken,
	})
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// 7. Revoked token must be rejected
	_, err = authSvc.RefreshToken(ctx, &model.RefreshTokenRequest{
		RefreshToken: refreshResp.RefreshToken,
	})
	if err == nil {
		t.Fatal("expected revoked token to be rejected")
	}
}
