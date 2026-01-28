package auth

import (
	"context"
)

// contextKey 是私有类型，防止其他包意外覆盖 Context 中的值
type contextKey string

const (
	userIDKey      contextKey = "x-user-id"      // 用户ID
	tenantIDKey    contextKey = "x-tenant-id"    // 租户ID
	deptIDKey      contextKey = "x-dept-id"      // 部门ID
	dataScopeKey   contextKey = "x-data-scope"   // 数据权限范围 (SELF, DEPT, DEPT_SUB, ALL等)
	authVersionKey contextKey = "x-auth-version" // 安全版本号
)

// ContextInfo 结构体用于一次性返回所有常用信息
type ContextInfo struct {
	UserID      int64
	TenantID    int64
	DeptID      int64
	DataScope   string
	AuthVersion int64
}

// --- Context 注入函数 (通常在 Middleware 中调用) ---

// NewContext 返回带有所有权限信息的新 Context
func NewContext(ctx context.Context, info ContextInfo) context.Context {
	ctx = context.WithValue(ctx, userIDKey, info.UserID)
	ctx = context.WithValue(ctx, tenantIDKey, info.TenantID)
	ctx = context.WithValue(ctx, deptIDKey, info.DeptID)
	ctx = context.WithValue(ctx, dataScopeKey, info.DataScope)
	ctx = context.WithValue(ctx, authVersionKey, info.AuthVersion)
	return ctx
}

// --- Context 提取函数 (通常在 Data 层或 Hooks 中调用) ---

// GetUserID 获取用户ID
func GetUserID(ctx context.Context) int64 {
	if v, ok := ctx.Value(userIDKey).(int64); ok {
		return v
	}
	return 0
}

// GetTenantID 获取租户ID
func GetTenantID(ctx context.Context) int64 {
	if v, ok := ctx.Value(tenantIDKey).(int64); ok {
		return v
	}
	return 0
}

// GetDeptID 获取部门ID
func GetDeptID(ctx context.Context) int64 {
	if v, ok := ctx.Value(deptIDKey).(int64); ok {
		return v
	}
	return 0
}

// GetDataScope 获取数据权限范围
func GetDataScope(ctx context.Context) string {
	if v, ok := ctx.Value(dataScopeKey).(string); ok {
		return v
	}
	return ""
}

// GetAuthVersion 获取安全版本号
func GetAuthVersion(ctx context.Context) int64 {
	if v, ok := ctx.Value(authVersionKey).(int64); ok {
		return v
	}
	return 0
}

// GetContextInfo 一次性获取所有权限信息
func GetContextInfo(ctx context.Context) ContextInfo {
	return ContextInfo{
		UserID:      GetUserID(ctx),
		TenantID:    GetTenantID(ctx),
		DeptID:      GetDeptID(ctx),
		DataScope:   GetDataScope(ctx),
		AuthVersion: GetAuthVersion(ctx),
	}
}

// --- 工具函数 ---

// WithTenantID 手动注入租户ID (通常用于系统初始化或异步任务)
func WithTenantID(ctx context.Context, tenantID int64) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// WithUserID 手动注入用户ID
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithDeptID 手动注入部门ID
func WithDeptID(ctx context.Context, deptID int64) context.Context {
	return context.WithValue(ctx, deptIDKey, deptID)
}

// WithDataScope 手动注入数据权限范围
func WithDataScope(ctx context.Context, dataScope string) context.Context {
	return context.WithValue(ctx, dataScopeKey, dataScope)
}

// WithAuthVersion 手动注入安全版本号
func WithAuthVersion(ctx context.Context, authVersion int64) context.Context {
	return context.WithValue(ctx, authVersionKey, authVersion)
}
