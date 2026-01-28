package model

import (
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义 JWT Claims
type CustomClaims struct {
	jwtv5.RegisteredClaims
	DeptID   int64 `json:"dept_id"`
	TenantID int64 `json:"tenant_id"`
}

// UserToken 用于持久化
type UserToken struct {
	JTI          string    // JWT ID
	UserID       string    // 用户 ID
	DeptID       int64     // 部门 ID
	TenantID     int64     // 租户 ID
	IssuedAt     time.Time // 签发时间
	ExpiresAt    time.Time // 过期时间
	TokenStr     string    // JWT 原文
	Revoked      bool      // 是否被强制注销
	RevokeReason string    // 注销原因
}
