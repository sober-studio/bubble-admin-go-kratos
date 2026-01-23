package model

import (
	"time"

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

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
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
