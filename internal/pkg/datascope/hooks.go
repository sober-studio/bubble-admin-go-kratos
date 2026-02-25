package datascope

import (
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"gorm.io/gorm"
)

// RegisterHooks 注册 GORM 回调钩子
// 在数据库初始化时调用，自动为所有查询操作注入租户隔离和数据范围
//
// 注意：此函数应在 gorm.Open 之后、AutoMigrate 之前调用
func RegisterHooks(db *gorm.DB) {
	// 注册查询前置回调 - 自动注入数据范围
	db.Callback().Query().Before("gorm:query").Register("datascope_query", queryCallback)

	// 注册创建前置回调 - 自动注入租户 ID 和创建者
	db.Callback().Create().Before("gorm:create").Register("datascope_create", createCallback)

	// 注册更新前置回调 - 自动注入租户 ID
	db.Callback().Update().Before("gorm:update").Register("datascope_update", updateCallback)
}

// queryCallback 查询前置回调
// 在每次查询执行前自动应用租户隔离和数据范围过滤
func queryCallback(db *gorm.DB) {
	// 如果没有表名，跳过
	if db.Statement == nil || db.Statement.Table == "" {
		return
	}

	// 获取上下文中的数据范围配置
	cfg, _ := GetConfig(db.Statement.Context)

	// 如果未设置配置，使用 Statement 的表名
	if cfg.TableName == "" {
		cfg.TableName = db.Statement.Table
	}

	// 应用数据范围
	db = ApplyDataScope(db, cfg)
}

// createCallback 创建前置回调
// 在每次创建记录时自动注入 tenant_id 和 created_by
func createCallback(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		return
	}

	tenantID := auth.GetTenantID(ctx)
	userID := auth.GetUserID(ctx)

	// 自动注入 tenant_id（如果模型包含此字段且租户 ID > 0）
	if tenantID > 0 {
		if stmt := db.Statement; stmt != nil && stmt.Schema != nil {
			if _, ok := stmt.Schema.FieldsByName["TenantID"]; ok {
				db.Statement.SetColumn("tenant_id", tenantID)
			}
		}
	}

	// 自动注入 created_by（如果模型包含此字段且用户 ID > 0）
	if userID > 0 {
		if stmt := db.Statement; stmt != nil && stmt.Schema != nil {
			if _, ok := stmt.Schema.FieldsByName["CreatedBy"]; ok {
				db.Statement.SetColumn("created_by", userID)
			}
		}
	}
}

// updateCallback 更新前置回调
// 在每次更新记录时自动注入租户 ID 条件，防止跨租户修改
func updateCallback(db *gorm.DB) {
	ctx := db.Statement.Context
	if ctx == nil {
		return
	}

	tenantID := auth.GetTenantID(ctx)

	// 自动注入 tenant_id 条件，防止跨租户修改
	if tenantID > 0 {
		if stmt := db.Statement; stmt != nil && stmt.Schema != nil {
			if _, ok := stmt.Schema.FieldsByName["TenantID"]; ok {
				// 直接使用 Where 条件添加租户过滤
				db = db.Where("tenant_id = ?", tenantID)
			}
		}
	}
}
