package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/errors"
	"eomp/services/auth/internal/model"
	"eomp/services/auth/internal/repository"
)

// AuthService defines business operations for auth
type AuthService interface {
	BootstrapAdmin(ctx context.Context, email, password, fullName string) error
	Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	LoginWithAudit(ctx context.Context, req *model.LoginRequest, ipAddress, userAgent string) (*model.AuthResponse, error)
	Logout(ctx context.Context, req *model.LogoutRequest) error
	RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.AuthResponse, error)
	GetMe(ctx context.Context, userID string) (*model.UserResponse, error)
	GetLoginHistory(ctx context.Context, email string, limit int) ([]model.LoginAuditLog, error)
}

type authService struct {
	repo       repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService constructs a new AuthService instance
func NewAuthService(repo repository.UserRepository, jwtManager *auth.JWTManager) AuthService {
	return &authService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func validateRegistration(req *model.RegisterRequest) error {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return errors.BadRequest("email, password, and full_name are required")
	}
	parsed, err := mail.ParseAddress(req.Email)
	if err != nil || parsed.Address != req.Email {
		return errors.BadRequest("email must be a valid address")
	}
	if len(req.Password) < 12 || len(req.Password) > 72 {
		return errors.BadRequest("password must contain between 12 and 72 characters")
	}
	var upper, lower, digit, special bool
	for _, r := range req.Password {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			special = true
		}
	}
	if !upper || !lower || !digit || !special {
		return errors.BadRequest("password must include uppercase, lowercase, number, and special characters")
	}
	return nil
}

// BootstrapAdmin creates the first administrator only when no account exists at
// the configured address. It never promotes an existing non-admin account.
func (s *authService) BootstrapAdmin(ctx context.Context, email, password, fullName string) error {
	if email == "" && password == "" {
		return nil
	}
	req := &model.RegisterRequest{Email: email, Password: password, FullName: fullName}
	if err := validateRegistration(req); err != nil {
		return err
	}
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("failed to check bootstrap administrator: %w", err)
	}
	if existing != nil {
		if existing.Role != model.RoleAdmin {
			return fmt.Errorf("bootstrap email already belongs to a non-admin account")
		}
		return nil
	}
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, &model.User{
		Email: req.Email, PasswordHash: hashedPassword, FullName: req.FullName,
		Role: model.RoleAdmin, IsActive: true,
	})
}

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	if err := validateRegistration(req); err != nil {
		return nil, err
	}

	// Check if email already registered
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Internal(ctx, "auth check existing user", err)
	}
	if existing != nil {
		return nil, errors.Conflict("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, errors.Internal(ctx, "auth hash password", err)
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         model.RoleEmployee,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, errors.Internal(ctx, "auth create user", err)
	}

	// Generate tokens
	var deptID string
	if user.DepartmentID != nil {
		deptID = *user.DepartmentID
	}
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role, deptID, user.FullName)
	if err != nil {
		return nil, errors.Internal(ctx, "auth generate registration token pair", err)
	}

	// Store refresh token
	refreshHash := hashToken(refreshToken)
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(s.jwtManager.RefreshTTL())); err != nil {
		return nil, errors.Internal(ctx, "auth persist registration refresh token", err)
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.AccessTTL().Seconds()),
		User:         user.ToResponse(),
	}, nil
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	return s.LoginWithAudit(ctx, req, "127.0.0.1", "Unknown")
}

func (s *authService) LoginWithAudit(ctx context.Context, req *model.LoginRequest, ipAddress, userAgent string) (*model.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.BadRequest("email and password are required")
	}

	if ipAddress == "" {
		ipAddress = "127.0.0.1"
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Internal(ctx, "auth find user for login", err)
	}

	if user == nil {
		reason := "user not found"
		_ = s.repo.RecordLoginAudit(ctx, &model.LoginAuditLog{
			Email:         req.Email,
			IPAddress:     ipAddress,
			UserAgent:     &userAgent,
			Status:        model.LoginStatusFailed,
			FailureReason: &reason,
		})
		return nil, errors.Unauthorized("invalid email or password")
	}

	if !user.IsActive {
		reason := "account deactivated"
		_ = s.repo.RecordLoginAudit(ctx, &model.LoginAuditLog{
			UserID:        &user.ID,
			Email:         req.Email,
			IPAddress:     ipAddress,
			UserAgent:     &userAgent,
			Status:        model.LoginStatusLocked,
			FailureReason: &reason,
		})
		return nil, errors.Forbidden("account is deactivated, please contact IT support")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		reason := "invalid password"
		_ = s.repo.RecordLoginAudit(ctx, &model.LoginAuditLog{
			UserID:        &user.ID,
			Email:         req.Email,
			IPAddress:     ipAddress,
			UserAgent:     &userAgent,
			Status:        model.LoginStatusFailed,
			FailureReason: &reason,
		})
		return nil, errors.Unauthorized("invalid email or password")
	}

	// Login successful
	_ = s.repo.RecordLoginAudit(ctx, &model.LoginAuditLog{
		UserID:    &user.ID,
		Email:     req.Email,
		IPAddress: ipAddress,
		UserAgent: &userAgent,
		Status:    model.LoginStatusSuccess,
	})

	var deptID string
	if user.DepartmentID != nil {
		deptID = *user.DepartmentID
	}
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role, deptID, user.FullName)
	if err != nil {
		return nil, errors.Internal(ctx, "auth generate login token pair", err)
	}

	refreshHash := hashToken(refreshToken)
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(s.jwtManager.RefreshTTL())); err != nil {
		return nil, errors.Internal(ctx, "auth persist login refresh token", err)
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.AccessTTL().Seconds()),
		User:         user.ToResponse(),
	}, nil
}

func (s *authService) Logout(ctx context.Context, req *model.LogoutRequest) error {
	if req.RefreshToken == "" {
		return errors.BadRequest("refresh_token is required for logout")
	}

	refreshHash := hashToken(req.RefreshToken)
	if err := s.repo.RevokeRefreshToken(ctx, refreshHash); err != nil {
		return errors.Internal(ctx, "auth revoke refresh token", err)
	}
	return nil
}

func (s *authService) RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.BadRequest("refresh_token is required")
	}

	claims, err := s.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.Unauthorized("invalid or expired refresh token")
	}

	refreshHash := hashToken(req.RefreshToken)
	userID, err := s.repo.ValidateRefreshToken(ctx, refreshHash)
	if err != nil || userID == "" {
		return nil, errors.Unauthorized("invalid or expired refresh token")
	}
	if claims.Subject != userID {
		return nil, errors.Unauthorized("invalid refresh token subject")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil || !user.IsActive {
		return nil, errors.Unauthorized("user not found or inactive")
	}

	var deptID string
	if user.DepartmentID != nil {
		deptID = *user.DepartmentID
	}
	newAccessToken, newRefreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role, deptID, user.FullName)
	if err != nil {
		return nil, errors.Internal(ctx, "auth generate rotated token pair", err)
	}

	newRefreshHash := hashToken(newRefreshToken)
	if err := s.repo.SaveRefreshToken(ctx, user.ID, newRefreshHash, time.Now().Add(s.jwtManager.RefreshTTL())); err != nil {
		return nil, errors.Internal(ctx, "auth persist rotated refresh token", err)
	}
	if err := s.repo.RevokeRefreshToken(ctx, refreshHash); err != nil {
		return nil, errors.Internal(ctx, "auth revoke previous refresh token", err)
	}

	return &model.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.AccessTTL().Seconds()),
		User:         user.ToResponse(),
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID string) (*model.UserResponse, error) {
	if userID == "" {
		return nil, errors.Unauthorized("unauthorized: missing user id")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.Internal(ctx, "auth get current user", err)
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *authService) GetLoginHistory(ctx context.Context, email string, limit int) ([]model.LoginAuditLog, error) {
	return s.repo.GetLoginHistory(ctx, email, limit)
}
