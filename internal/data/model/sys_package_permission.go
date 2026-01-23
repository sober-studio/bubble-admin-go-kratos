package model

import "time"

// SysPackagePermission 套餐权限关联表
type SysPackagePermission struct {
	ID           int64     `gorm:"column:id;type:bigint;primaryKey" json:"id"`
	PackageID    int64     `gorm:"column:package_id;type:bigint;not null;comment:套餐 ID" json:"package_id"`
	PermissionID int64     `gorm:"column:permission_id;type:bigint;not null;comment:权限 ID" json:"permission_id"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (*SysPackagePermission) TableName() string {
	return "sys_package_permission"
}
