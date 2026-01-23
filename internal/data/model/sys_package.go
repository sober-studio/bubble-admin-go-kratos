package model

// SysPackage 租户套餐表
type SysPackage struct {
	BaseAuthModel
	Name   string `gorm:"column:name;type:varchar(128);not null;comment:套餐名称" json:"name"`
	Status int16  `gorm:"column:status;type:smallint;default:1;comment:状态" json:"status"`
	Remark string `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`
}

func (*SysPackage) TableName() string {
	return "sys_package"
}
