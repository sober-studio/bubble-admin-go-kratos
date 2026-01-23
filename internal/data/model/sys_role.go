package model

// SysRole 角色表
type SysRole struct {
	BaseModel
	TenantID int64  `gorm:"column:tenant_id;type:bigint;not null;comment:租户 ID" json:"tenant_id"`
	Name     string `gorm:"column:name;type:varchar(64);not null;comment:角色名称" json:"name"`
	Code     string `gorm:"column:code;type:varchar(64);not null;comment:角色编码" json:"code"`
}

func (*SysRole) TableName() string {
	return "sys_role"
}
