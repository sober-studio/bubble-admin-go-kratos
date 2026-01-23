package model

import "time"

// SysUser 用户表
type SysUser struct {
	BaseAuthModel
	Username          string    `gorm:"column:username;type:varchar(64);not null;comment:登录用户名" json:"username"`
	PasswordHash      string    `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希" json:"password_hash"`
	Name              string    `gorm:"column:name;type:varchar(64);comment:用户名称" json:"name"`
	Mobile            string    `gorm:"column:mobile;type:varchar(20);comment:手机号" json:"mobile"`
	Avatar            string    `gorm:"column:avatar;type:varchar(255);comment:头像" json:"avatar"`
	Status            int16     `gorm:"column:status;type:smallint;default:1;comment:可用状态" json:"status"`
	LoginFailedCount  int       `gorm:"column:login_failed_count;type:int;default:0;comment:登录失败次数" json:"login_failed_count"`
	LastLoginFailedAt time.Time `gorm:"column:last_login_failed_at;type:timestamp with time zone;comment:上次登录失败时间" json:"last_login_failed_at"`
}

func (*SysUser) TableName() string {
	return "sys_user"
}
