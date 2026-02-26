package model

// SysPackage 租户套餐表
type SysPackage struct {
	BaseAuthModel
	Name        string `gorm:"column:name;type:varchar(128);not null;comment:套餐名称" json:"name"`
	Code        string `gorm:"column:code;type:varchar(64);not null;comment:套餐编码" json:"code"`
	Description string `gorm:"column:description;type:varchar(512);comment:描述" json:"description"`
	Status      int16  `gorm:"column:status;type:smallint;default:1;comment:状态" json:"status"`
}

func (*SysPackage) TableName() string {
	return "sys_package"
}
