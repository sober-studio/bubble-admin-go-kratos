package data

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/datascope"
	"gorm.io/gorm"
)

// BaseRepo 定义基础仓库结构
// 所有业务 Repository 都应嵌入此结构，以获得数据范围和租户隔离能力
type BaseRepo struct {
	data      *Data       // 包含 gorm.DB 和 Casbin Enforcer 的 Data 对象
	log       *log.Helper // 日志辅助工具
	tableName string      // 主表名，用于联表查询时区分
}

// NewBaseRepo 创建基础仓库（简单模式）
// 不指定表名，适用于单表查询场景
func NewBaseRepo(data *Data, logger log.Logger) BaseRepo {
	return BaseRepo{
		data:      data,
		log:       log.NewHelper(logger),
		tableName: "",
	}
}

// NewBaseRepoWithTable 创建基础仓库（指定表名）
// tableName: 主表名，如 "orders", "products"，联表查询时需要指定
func NewBaseRepoWithTable(data *Data, logger log.Logger, tableName string) BaseRepo {
	return BaseRepo{
		data:      data,
		log:       log.NewHelper(logger),
		tableName: tableName,
	}
}

// DataScope 是 BaseRepo 的一个方法，所有继承它的 Repo 都能直接使用
// 自动从 Context 中提取租户 ID 和数据范围，生成 GORM Scope 函数
//
// 使用示例：
//
//	err := r.data.db.WithContext(ctx).
//	    Scopes(r.DataScope(ctx)).
//	    Find(&orders).Error
func (r *BaseRepo) DataScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 从 Context 中提取权限信息（使用 auth 包的函数，与中间件注入保持一致）
		scope := auth.GetDataScope(ctx)
		deptID := auth.GetDeptID(ctx)
		userID := auth.GetUserID(ctx)
		tenantID := auth.GetTenantID(ctx)

		// 1. 租户物理隔离（每个 SQL 必加）
		db = db.Where("tenant_id = ?", tenantID)

		// 2. 根据数据范围动态拼接 SQL
		switch scope {
		case "ALL":
			return db
		case "DEPT_SUB":
			return db.Where("dept_id IN (SELECT id FROM sys_dept WHERE id = ? OR ancestors LIKE ?)",
				deptID, fmt.Sprintf("%%,%d,%%", deptID))
		case "DEPT":
			return db.Where("dept_id = ?", deptID)
		case "SELF":
			return db.Where("created_by = ?", userID)
		default:
			return db.Where("1 = 0") // 默认无权，防止逻辑漏洞
		}
	}
}

// WithDataScope 数据范围 Scope（推荐用于联表/复杂查询）
// 自动识别表名并应用租户隔离 + 数据范围过滤
//
// 使用示例：
//
//	err := r.data.db.WithContext(ctx).
//	    Table("orders").
//	    Joins("LEFT JOIN users ON users.id = orders.user_id").
//	    Scopes(r.WithDataScope(ctx)).
//	    Find(&orders).Error
func (r *BaseRepo) WithDataScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		cfg := datascope.Config{
			TableName: r.tableName,
		}
		return datascope.ApplyDataScope(db, cfg)
	}
}

// WithTenantScope 租户隔离 Scope（仅租户隔离，不需要数据范围）
// 适用于不需要数据范围控制的场景，如系统配置、字典表等
//
// 使用示例：
//
//	err := r.data.db.WithContext(ctx).
//	    Scopes(r.WithTenantScope(ctx)).
//	    Find(&configs).Error
func (r *BaseRepo) WithTenantScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tenantID := auth.GetTenantID(ctx)
		if tenantID > 0 {
			if r.tableName != "" {
				return db.Where(r.tableName+".tenant_id = ?", tenantID)
			}
			return db.Where("tenant_id = ?", tenantID)
		}
		return db
	}
}

// WithSubQuery 子查询 Scope（用于子查询场景）
// 子查询需要单独应用数据范围过滤
//
// 使用示例：
//
//	userSubQuery := r.data.db.WithContext(ctx).
//	    Table("users").
//	    Select("id").
//	    Scopes(r.WithSubQuery(ctx, "users"))
//
//	orders := r.data.db.WithContext(ctx).
//	    Table("orders").
//	    Where("user_id IN (?)", userSubQuery).
//	    Scopes(r.WithDataScope(ctx))
func (r *BaseRepo) WithSubQuery(ctx context.Context, subTableName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		cfg := datascope.Config{
			TableName: subTableName,
			SubQuery:  true,
		}
		return datascope.ApplyDataScope(db, cfg)
	}
}

// WithConfig 高级配置 Scope（用于特殊场景）
// 允许显式配置数据范围行为
//
// 使用示例：
//
//	// 禁用数据范围（超级管理员）
//	ctx = datascope.WithConfig(ctx, datascope.Config{
//	    DisableDataScope: true,
//	})
//
//	// 指定表别名
//	ctx = datascope.WithConfig(ctx, datascope.Config{
//	    TableAlias: "o",
//	    FieldMap: map[string]string{
//	        "tenant_id": "o.tenant_id",
//	        "dept_id": "o.dept_id",
//	    },
//	})
func (r *BaseRepo) WithConfig(ctx context.Context, cfg datascope.Config) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 合并配置：优先使用传入的配置，表名优先使用 BaseRepo 的 tableName
		if cfg.TableName == "" && r.tableName != "" {
			cfg.TableName = r.tableName
		}
		return datascope.ApplyDataScope(db, cfg)
	}
}

// Paginate 分页 Scope
// 自动处理分页参数，防止页码异常
//
// 参数:
//   - page: 页码，从 1 开始
//   - pageSize: 每页数量，最大 100
//
// 使用示例：
//
//	err := r.data.db.WithContext(ctx).
//	    Scopes(r.Paginate(req.Page, req.PageSize)).
//	    Find(&users).Error
func (r *BaseRepo) Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100 // 强制限制最大分页大小，保护数据库
		case pageSize <= 0:
			pageSize = 10
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// OnlyTrashed 只看已逻辑删除的数据
// 用于查看已删除记录的回收站功能
//
// 使用示例：
//
//	err := r.data.db.WithContext(ctx).
//	    Scopes(r.OnlyTrashed).
//	    Find(&deletedUsers).Error
func (r *BaseRepo) OnlyTrashed(db *gorm.DB) *gorm.DB {
	return db.Unscoped().Where("deleted_at IS NOT NULL")
}

// SortBy 排序 Scope
// 通用排序功能
//
// 参数:
//   - field: 排序字段名
//   - ascending: true 升序，false 降序
//
// 使用示例：
//
//	err := r.data.db.WithContext(ctx).
//	    Scopes(r.SortBy("created_at", false)).
//	    Find(&orders).Error
func (r *BaseRepo) SortBy(field string, ascending bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		order := "DESC"
		if ascending {
			order = "ASC"
		}
		return db.Order(fmt.Sprintf("%s %s", field, order))
	}
}
