package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth/model"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth/store"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.Unauthorized("INVALID_TOKEN", "无效的 Token")
	ErrTokenExpired     = errors.Unauthorized("TOKEN_EXPIRED", "Token 已过期")
	ErrJWTGenerateError = errors.Unauthorized("JWT_GENERATE_ERROR", "JWT 生成错误")
)

// TokenService 令牌服务接口，用于生成和解析 JWT 令牌
type TokenService interface {
	// GenerateToken 生成令牌
	GenerateToken(ctx context.Context, userID string, deptID int64, tenantID int64) (string, error)
	// ParseTokenFromTokenString 解析令牌，返回 Claims
	ParseTokenFromTokenString(ctx context.Context, tokenStr string) (*model.CustomClaims, error)
	// ParseTokenFromContext 解析令牌，返回 Claims
	ParseTokenFromContext(ctx context.Context) (*model.CustomClaims, error)
	// GetUserIDFromTokenString 获取用户ID
	GetUserIDFromTokenString(ctx context.Context, tokenStr string) (int64, error)
	// GetUserIDFromContext 获取用户ID
	GetUserIDFromContext(ctx context.Context) (int64, error)
	// GetDeptIDFromTokenString 获取部门ID
	GetDeptIDFromTokenString(ctx context.Context, tokenStr string) (int64, error)
	// GetDeptIDFromContext 获取部门ID
	GetDeptIDFromContext(ctx context.Context) (int64, error)
	// GetTenantIDFromTokenString 获取租户ID
	GetTenantIDFromTokenString(ctx context.Context, tokenStr string) (int64, error)
	// GetTenantIDFromContext 获取租户ID
	GetTenantIDFromContext(ctx context.Context) (int64, error)
	// GetUserTokens 获取用户令牌
	GetUserTokens(ctx context.Context, userID string) (*[]model.UserToken, error)
	// RevokeToken 撤销令牌，如果 jti 为空，则从 context 中获取当前 token 的 jti
	RevokeToken(ctx context.Context, jti string) error
	// RevokeAllTokens 撤销用户所有令牌
	RevokeAllTokens(ctx context.Context) error
	// RevokeAllTokensByUserID 根据用户ID撤销所有令牌
	RevokeAllTokensByUserID(ctx context.Context, userID int64) error
	// GetSecretKey 获取密钥
	GetSecretKey() []byte
}

var _ TokenService = (*JWTTokenService)(nil)

// JWTTokenService JWT 令牌服务接口
type JWTTokenService struct {
	secretKey []byte
	ttl       time.Duration
	store     store.TokenStore
}

func NewJWTTokenService(secretKey string, ttl time.Duration, store store.TokenStore) TokenService {
	return &JWTTokenService{
		secretKey: []byte(secretKey),
		ttl:       ttl,
		store:     store,
	}
}

func (s *JWTTokenService) GenerateToken(ctx context.Context, userID string, deptID int64, tenantID int64) (string, error) {
	jti := uuid.New().String()
	now := time.Now()
	claims := model.CustomClaims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwtv5.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ID:        jti,
		},
		DeptID:   deptID,
		TenantID: tenantID,
	}
	tokenStr, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims).SignedString(s.secretKey)
	if err != nil {
		log.Errorf("Failed to generate token: %v", err)
		return "", ErrJWTGenerateError
	}
	log.Infof("Generated token: %s", tokenStr)
	token := &model.UserToken{
		JTI:       jti,
		UserID:    userID,
		DeptID:    deptID,
		TenantID:  tenantID,
		IssuedAt:  now,
		ExpiresAt: now.Add(s.ttl),
		TokenStr:  tokenStr,
	}

	if err := s.store.SaveToken(ctx, token); err != nil {
		log.Error("Failed to save token: %v", err)
		return "", ErrJWTGenerateError
	}

	return tokenStr, nil
}

func (s *JWTTokenService) ParseTokenFromTokenString(ctx context.Context, tokenStr string) (*model.CustomClaims, error) {
	t, err := jwtv5.ParseWithClaims(tokenStr, &model.CustomClaims{}, func(token *jwtv5.Token) (interface{}, error) {
		return s.secretKey, nil
	})
	if err != nil || !t.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := t.Claims.(*model.CustomClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	stored, err := s.store.GetToken(ctx, claims.ID)
	if err != nil || stored.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}
	return claims, nil
}

func (s *JWTTokenService) ParseTokenFromContext(ctx context.Context) (*model.CustomClaims, error) {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		log.Errorf("invalid token")
		return nil, ErrInvalidToken
	}
	customClaims, ok := claims.(*model.CustomClaims)
	if !ok {
		log.Errorf("invalid token")
		return nil, ErrInvalidToken
	}

	stored, err := s.store.GetToken(ctx, customClaims.ID)
	if err != nil || stored.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}
	return customClaims, nil
}

func (s *JWTTokenService) GetUserIDFromTokenString(ctx context.Context, tokenStr string) (int64, error) {
	claims, err := s.ParseTokenFromTokenString(ctx, tokenStr)
	if err != nil {
		return 0, err
	}
	return parseUserID(claims.Subject)
}

func (s *JWTTokenService) GetUserIDFromContext(ctx context.Context) (int64, error) {
	claims, err := s.ParseTokenFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return parseUserID(claims.Subject)
}

func (s *JWTTokenService) GetDeptIDFromTokenString(ctx context.Context, tokenStr string) (int64, error) {
	claims, err := s.ParseTokenFromTokenString(ctx, tokenStr)
	if err != nil {
		return 0, err
	}
	return claims.DeptID, nil
}

func (s *JWTTokenService) GetDeptIDFromContext(ctx context.Context) (int64, error) {
	claims, err := s.ParseTokenFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return claims.DeptID, nil
}

func (s *JWTTokenService) GetTenantIDFromTokenString(ctx context.Context, tokenStr string) (int64, error) {
	claims, err := s.ParseTokenFromTokenString(ctx, tokenStr)
	if err != nil {
		return 0, err
	}
	return claims.TenantID, nil
}

func (s *JWTTokenService) GetTenantIDFromContext(ctx context.Context) (int64, error) {
	claims, err := s.ParseTokenFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return claims.TenantID, nil
}

func (s *JWTTokenService) GetUserTokens(ctx context.Context, userID string) (*[]model.UserToken, error) {
	return s.store.GetUserTokens(ctx, userID)
}

// 将 storedUserID 字符串转换为 int64 类型的 userID
func parseUserID(storedUserID string) (int64, error) {
	userID, err := strconv.ParseInt(storedUserID, 10, 64)
	if err != nil {
		return 0, ErrInvalidToken.WithCause(err)
	}
	return userID, nil
}

func (s *JWTTokenService) RevokeToken(ctx context.Context, jti string) error {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return ErrInvalidToken
	}
	customClaims, ok := claims.(*model.CustomClaims)
	if !ok {
		return ErrInvalidToken
	}
	userID := customClaims.Subject

	// 如果 jti 为空，则撤销当前 token
	if jti == "" {
		jti = customClaims.ID
	}

	return s.store.DeleteUserToken(ctx, userID, jti)
}

func (s *JWTTokenService) RevokeAllTokens(ctx context.Context) error {
	claims, err := s.ParseTokenFromContext(ctx)
	if err != nil {
		return err
	}
	return s.store.DeleteUserTokens(ctx, claims.Subject)
}

func (s *JWTTokenService) RevokeAllTokensByUserID(ctx context.Context, userID int64) error {
	userIDStr := strconv.FormatInt(userID, 10)
	return s.store.DeleteUserTokens(ctx, userIDStr)
}

func (s *JWTTokenService) GetSecretKey() []byte {
	return s.secretKey
}
