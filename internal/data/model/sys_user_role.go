package model

// SysUserRole 用户角色关联表
type SysUserRole struct {
	BaseAuthModel
	UserID int64 `gorm:"column:user_id;type:bigint;not null;comment:用户 ID" json:"user_id"`
	RoleID int64 `gorm:"column:role_id;type:bigint;not null;comment:角色 ID" json:"role_id"`
}

func (*SysUserRole) TableName() string {
	return "sys_user_role"
}
