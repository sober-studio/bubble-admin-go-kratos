package data

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"gorm.io/gorm"
)

// BaseRepo 定义基础仓库结构
type BaseRepo struct {
	data *Data       // 包含 gorm.DB 和 Casbin Enforcer 的 Data 对象
	log  *log.Helper // 日志辅助工具
}

// NewBaseRepo 构造函数
func NewBaseRepo(data *Data, logger log.Logger) BaseRepo {
	return BaseRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// DataScope 是 BaseRepo 的一个方法，所有继承它的 Repo 都能直接使用
func (r *BaseRepo) DataScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 这里是从 Context 中提取中间件注入的权限标识（前面讨论的逻辑）
		scope, _ := ctx.Value("DataScope").(string)
		deptID, _ := ctx.Value("UserDeptID").(int64)
		userID, _ := ctx.Value("UserID").(int64)
		tenantID, _ := ctx.Value("TenantID").(int64)

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

/* 使用示例
package data

import (
    "context"
    "your_project/internal/biz"
)

type orderRepo struct {
    BaseRepo // 匿名嵌入，这样 orderRepo 就拥有了 DataScope 方法
}

func NewOrderRepo(data *Data, logger log.Logger) biz.OrderRepo {
    return &orderRepo{
        BaseRepo: NewBaseRepo(data, logger),
    }
}

func (r *orderRepo) ListOrders(ctx context.Context) ([]*biz.Order, error) {
    var orders []*Order
    // 直接调用 r.DataScope(ctx)，代码非常简洁
    err := r.data.db.WithContext(ctx).
        Scopes(r.DataScope(ctx)).
        Find(&orders).Error

    return orders, err
}
*/

func (r *BaseRepo) TenantScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 从 Context 获取租户 ID
		tid := auth.GetTenantID(ctx)
		return db.Where("tenant_id = ?", tid)
	}
}

// 使用：
// r.data.db.WithContext(ctx).Scopes(r.TenantScope(ctx)).Find(&configs)

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

// 使用：
// r.data.db.Scopes(Paginate(req.Page, req.PageSize)).Find(&users)

// OnlyTrashed 只看已逻辑删除的数据
func (r *BaseRepo) OnlyTrashed(db *gorm.DB) *gorm.DB {
	return db.Unscoped().Where("deleted_at IS NOT NULL")
}

// 使用：
// r.data.db.Scopes(OnlyTrashed).Find(&deletedUsers)

// SortBy 排序
func (r *BaseRepo) SortBy(field string, ascending bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		order := "DESC"
		if ascending {
			order = "ASC"
		}
		return db.Order(fmt.Sprintf("%s %s", field, order))
	}
}

// 使用：
// r.data.db.Scopes(SortBy("created_at", false)).Find(&orders)

// IsOnSale 在售商品
func (r *BaseRepo) IsOnSale(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", 1).Where("stock > ?", 0)
}

// 使用：
// r.data.db.Scopes(IsOnSale).Find(&products)
