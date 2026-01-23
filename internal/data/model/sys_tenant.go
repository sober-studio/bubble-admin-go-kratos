package model

import "time"

// SysTenant 租户表
type SysTenant struct {
	BaseModel
	CreatedBy  int64     `gorm:"column:created_by;index;comment:创建者ID" json:"created_by"`
	Code       string    `gorm:"column:code;type:varchar(64);not null;comment:租户编码" json:"code"`
	Name       string    `gorm:"column:name;type:varchar(128);not null;comment:租户名称" json:"name"`
	PackageID  int64     `gorm:"column:package_id;type:bigint;comment:租户套餐 ID" json:"package_id"`
	ExpireTime time.Time `gorm:"column:expire_time;type:timestamp with time zone;comment:过期时间" json:"expire_time"`
	Status     int16     `gorm:"column:status;type:smallint;default:1;comment:状态" json:"status"`
}

func (*SysTenant) TableName() string {
	return "sys_tenant"
}
