package model

import (
	"time"

	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/idgen"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

var globalIDGen idgen.IDGenerator

func SetIDGenerator(g idgen.IDGenerator) { globalIDGen = g }

func NextID() (int64, error) {
	if globalIDGen == nil {
		return 0, nil
	}
	return globalIDGen.NextID()
}

func (m *BaseModel) BeforeCreate(_ *gorm.DB) error {
	if m.ID != 0 {
		return nil
	}
	id, err := NextID()
	if err != nil {
		return err
	}
	if id != 0 {
		m.ID = id
	}
	return nil
}

type AuthField struct {
	TenantID  int64 `gorm:"column:tenant_id;index;comment:租户ID" json:"tenant_id"`
	CreatedBy int64 `gorm:"column:created_by;index;comment:创建者ID" json:"created_by"`
	DeptID    int64 `gorm:"column:dept_id;index;comment:所属部门ID" json:"dept_id"`
}

type BaseAuthModel struct {
	BaseModel
	AuthField
}

func (m *BaseAuthModel) BeforeCreate(tx *gorm.DB) error {
	// 1. 调用BaseModel的BeforeCreate方法处理ID生成
	if err := m.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}

	// 2. 自动填充权限字段
	// GORM v2 允许从 tx.Statement.Context 获取上下文
	ctx := tx.Statement.Context
	if ctx != nil {
		uid := auth.GetUserID(ctx)
		tid := auth.GetTenantID(ctx)
		did := auth.GetDeptID(ctx)

		// 只有当字段为0时才填充，允许手动覆盖
		if m.CreatedBy == 0 {
			m.CreatedBy = uid
		}
		if m.TenantID == 0 {
			m.TenantID = tid
		}
		if m.DeptID == 0 {
			m.DeptID = did
		}
	}

	return nil
}

/* 上述方案需要确保 context 传递
func (r *orderRepo) Create(ctx context.Context, order *biz.Order) error {
	// 这里的 WithContext(ctx) 会将包含 JWT 信息的 Context 传给 GORM
	// 从而触发 BaseAuthModel 里的 BeforeCreate 自动填充字段
	return r.data.db.WithContext(ctx).Create(order).Error
}
或者
// 快速获取带 Context 的生成的 Query 实例
func (d *Data) Query(ctx context.Context) *query.Query {
    return query.Use(d.db).WithContext(ctx)
}
*/

/*
func (m *BaseAuthModel) BeforeUpdate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	if ctx != nil {
		uid, _, _ := auth.GetAuthInfo(ctx)
		// 假设你有一个 UpdatedBy 字段
		// tx.Statement.SetColumn("updated_by", uid)
	}
	return nil
}
*/
