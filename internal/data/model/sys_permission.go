package model

// SysPermission 权限/菜单表
type SysPermission struct {
	BaseModel
	ParentID  int64  `gorm:"column:parent_id;type:bigint;default:0;comment:父权限 ID" json:"parent_id"`
	Name      string `gorm:"column:name;type:varchar(64);not null;comment:权限名称" json:"name"`
	Code      string `gorm:"column:code;type:varchar(64);not null;comment:权限编码" json:"code"`
	Type      string `gorm:"column:type;type:varchar(20);not null;comment:类型: MENU/BUTTON/API" json:"type"`
	APIPath   string `gorm:"column:api_path;type:varchar(255);comment:Kratos内部路径/API路径" json:"api_path"`
	APIMethod string `gorm:"column:api_method;type:varchar(20);default:V;comment:API方法" json:"api_method"`
	Sort      int32  `gorm:"column:sort;type:int;default:0;comment:排序" json:"sort"`
}

func (*SysPermission) TableName() string {
	return "sys_permission"
}
