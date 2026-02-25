package datascope

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"gorm.io/gorm"
)

// Config 数据范围配置
// 用于在 Context 中传递数据范围配置，控制数据范围过滤行为
type Config struct {
	TableName        string // 主表名，用于联表查询时区分，如 "orders", "products"
	TableAlias       string // 表别名，用于已设置别名的情况
	DisableTenant    bool   // 是否禁用租户隔离（某些系统表不需要）
	DisableDataScope bool   // 是否禁用数据范围过滤
	SubQuery         bool   // 是否为子查询模式
}

// DataScopeContextKey 用于在 Context 中传递数据范围配置
const DataScopeContextKey = "x-data-scope-config"

// WithConfig 设置数据范围配置
// ctx: 原始上下文
// cfg: 数据范围配置
// 返回: 携带配置的新上下文
func WithConfig(ctx context.Context, cfg Config) context.Context {
	return context.WithValue(ctx, DataScopeContextKey, cfg)
}

// GetConfig 获取数据范围配置
// ctx: 上下文
// 返回: 配置对象和是否设置过配置
func GetConfig(ctx context.Context) (Config, bool) {
	v := ctx.Value(DataScopeContextKey)
	if cfg, ok := v.(Config); ok {
		return cfg, true
	}
	return Config{}, false
}

// ApplyDataScope 应用数据范围
// 核心方法：根据配置和上下文中的权限信息，自动注入租户隔离和数据范围过滤
//
// 参数:
//   - db: GORM 数据库对象
//   - cfg: 数据范围配置
//
// 返回: 应用过滤后的数据库对象
func ApplyDataScope(db *gorm.DB, cfg Config) *gorm.DB {
	ctx := db.Statement.Context
	if ctx == nil {
		return db
	}

	tenantID := auth.GetTenantID(ctx)
	userID := auth.GetUserID(ctx)
	deptID := auth.GetDeptID(ctx)
	scope := auth.GetDataScope(ctx)

	// 1. 租户隔离
	if !cfg.DisableTenant && tenantID > 0 {
		tenantField := getTenantField(cfg)
		db = db.Where(fmt.Sprintf("%s = ?", tenantField), tenantID)
	}

	// 2. 数据范围过滤
	if cfg.DisableDataScope || scope == "" || scope == "ALL" {
		return db
	}

	switch scope {
	case "SELF":
		createdField := getCreatedField(cfg)
		db = db.Where(fmt.Sprintf("%s = ?", createdField), userID)

	case "DEPT":
		deptField := getDeptField(cfg)
		db = db.Where(fmt.Sprintf("%s = ?", deptField), deptID)

	case "DEPT_SUB":
		deptField := getDeptField(cfg)
		db = db.Where(fmt.Sprintf("%s IN (SELECT id FROM sys_dept WHERE id = ? OR ancestors LIKE ?)",
			deptField), deptID, fmt.Sprintf("%%,%d,%%", deptID))

	default:
		// 未知范围，默认无权
		db = db.Where("1 = 0")
	}

	return db
}

// getTenantField 获取租户字段名
// 根据配置返回租户字段的完整表示
func getTenantField(cfg Config) string {
	if cfg.TableName != "" {
		return cfg.TableName + ".tenant_id"
	}
	if cfg.TableAlias != "" {
		return cfg.TableAlias + ".tenant_id"
	}
	return "tenant_id"
}

// getDeptField 获取部门字段名
// 根据配置返回部门字段的完整表示
func getDeptField(cfg Config) string {
	if cfg.TableName != "" {
		return cfg.TableName + ".dept_id"
	}
	if cfg.TableAlias != "" {
		return cfg.TableAlias + ".dept_id"
	}
	return "dept_id"
}

// getCreatedField 获取创建者字段名
// 根据配置返回创建者字段的完整表示
func getCreatedField(cfg Config) string {
	if cfg.TableName != "" {
		return cfg.TableName + ".created_by"
	}
	if cfg.TableAlias != "" {
		return cfg.TableAlias + ".created_by"
	}
	return "created_by"
}

// formatInt64 将 int64 转换为字符串
func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}
