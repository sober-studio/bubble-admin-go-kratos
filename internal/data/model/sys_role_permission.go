package model

// SysRolePermission 角色权限关联表
type SysRolePermission struct {
	BaseModel
	RoleID       int64  `gorm:"column:role_id;type:bigint;not null;comment:角色 ID" json:"role_id"`
	PermissionID int64  `gorm:"column:permission_id;type:bigint;not null;comment:权限 ID" json:"permission_id"`
	DataScope    string `gorm:"column:data_scope;type:varchar(20);default:SELF;comment:数据范围: SELF(个人), DEPT(本部门), DEPT_SUB(本部门及下级), ALL(全租户)" json:"data_scope"`
	TenantID     int64  `gorm:"column:tenant_id;type:bigint;not null;comment:租户 ID" json:"tenant_id"`
}

func (*SysRolePermission) TableName() string {
	return "sys_role_permission"
}
