package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/errors"
	"eomp/services/auth/internal/model"
	"eomp/services/auth/internal/repository"
)

// AuthService defines business operations for auth
type AuthService interface {
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

func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.AuthResponse, error) {
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return nil, errors.BadRequest("email, password, and full_name are required")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if email already registered
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to check existing user: %v", err))
	}
	if existing != nil {
		return nil, errors.Conflict("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, errors.InternalServerError("failed to hash password")
	}

	role := req.Role
	if role == "" {
		role = model.RoleEmployee
	}

	user := &model.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         role,
		DepartmentID: req.DepartmentID,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to create user: %v", err))
	}

	// Generate tokens
	var deptID string
	if user.DepartmentID != nil {
		deptID = *user.DepartmentID
	}
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role, deptID, user.FullName)
	if err != nil {
		return nil, errors.InternalServerError("failed to generate token pair")
	}

	// Store refresh token
	refreshHash := hashToken(refreshToken)
	_ = s.repo.SaveRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(7*24*time.Hour))

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
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
		return nil, errors.InternalServerError(fmt.Sprintf("database error: %v", err))
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
		return nil, errors.InternalServerError("failed to generate token pair")
	}

	refreshHash := hashToken(refreshToken)
	_ = s.repo.SaveRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(7*24*time.Hour))

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User:         user.ToResponse(),
	}, nil
}

func (s *authService) Logout(ctx context.Context, req *model.LogoutRequest) error {
	if req.RefreshToken == "" {
		return errors.BadRequest("refresh_token is required for logout")
	}

	refreshHash := hashToken(req.RefreshToken)
	if err := s.repo.RevokeRefreshToken(ctx, refreshHash); err != nil {
		return errors.InternalServerError(fmt.Sprintf("failed to revoke refresh token: %v", err))
	}
	return nil
}

func (s *authService) RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.BadRequest("refresh_token is required")
	}

	refreshHash := hashToken(req.RefreshToken)
	userID, err := s.repo.ValidateRefreshToken(ctx, refreshHash)
	if err != nil || userID == "" {
		return nil, errors.Unauthorized("invalid or expired refresh token")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil || !user.IsActive {
		return nil, errors.Unauthorized("user not found or inactive")
	}

	// Revoke old refresh token (token rotation)
	_ = s.repo.RevokeRefreshToken(ctx, refreshHash)

	var deptID string
	if user.DepartmentID != nil {
		deptID = *user.DepartmentID
	}
	newAccessToken, newRefreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Role, deptID, user.FullName)
	if err != nil {
		return nil, errors.InternalServerError("failed to generate new token pair")
	}

	newRefreshHash := hashToken(newRefreshToken)
	_ = s.repo.SaveRefreshToken(ctx, user.ID, newRefreshHash, time.Now().Add(7*24*time.Hour))

	return &model.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User:         user.ToResponse(),
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID string) (*model.UserResponse, error) {
	if userID == "" {
		return nil, errors.Unauthorized("unauthorized: missing user id")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to get user: %v", err))
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
