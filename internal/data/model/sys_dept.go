package model

// SysDept 部门表
type SysDept struct {
	BaseAuthModel
	ParentID       int64  `gorm:"column:parent_id;type:bigint;default:0;comment:父部门 ID" json:"parent_id"`
	Name           string `gorm:"column:name;type:varchar(128);not null;comment:部门名称" json:"name"`
	Ancestors      string `gorm:"column:ancestors;type:varchar(512);comment:祖先路径" json:"ancestors"`
	Sort           int32  `gorm:"column:sort;type:int;default:0;comment:排序序号" json:"sort"`
	LeaderUserID   int64  `gorm:"column:leader_user_id;type:bigint;default:0;comment:负责人用户ID" json:"leader_user_id"`
	LeaderUserName string `gorm:"column:leader_user_name;type:varchar(64);comment:负责人姓名" json:"leader_user_name"`
	Phone          string `gorm:"column:phone;type:varchar(32);comment:联系电话" json:"phone"`
	Email          string `gorm:"column:email;type:varchar(128);comment:邮箱" json:"email"`
	Status         int32  `gorm:"column:status;type:int;default:1;comment:状态：0=禁用，1=正常" json:"status"`
}

func (*SysDept) TableName() string {
	return "sys_dept"
}
