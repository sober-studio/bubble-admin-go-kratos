package model

// SysDept 部门表
type SysDept struct {
	BaseModel
	TenantID  int64  `gorm:"column:tenant_id;type:bigint;not null;comment:租户 ID" json:"tenant_id"`
	ParentID  int64  `gorm:"column:parent_id;type:bigint;default:0;comment:父部门 ID" json:"parent_id"`
	Name      string `gorm:"column:name;type:varchar(128);not null;comment:部门名称" json:"name"`
	Ancestors string `gorm:"column:ancestors;type:varchar(512);comment:祖先路径" json:"ancestors"`
	Sort      int32  `gorm:"column:sort;type:int;default:0;comment:排序序号" json:"sort"`
}

func (*SysDept) TableName() string {
	return "sys_dept"
}
