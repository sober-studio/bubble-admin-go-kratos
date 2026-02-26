package data

import (
	"fmt"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type SysPermissionAdapter struct {
	db *gorm.DB
}

func NewSysPermissionAdapter(db *gorm.DB) persist.Adapter {
	return &SysPermissionAdapter{db: db}
}

// LoadPolicy 将业务数据库的关联关系"翻译"给 Casbin 内存
func (a *SysPermissionAdapter) LoadPolicy(m model.Model) error {
	helper := log.NewHelper(log.With(log.GetLogger(), "module", "casbin-adapter"))

	// 1. 加载用户-角色继承 (g)
	type UserRole struct {
		UserID   int64  `gorm:"column:user_id"`
		RoleCode string `gorm:"column:role_code"`
		TenantID int64  `gorm:"column:tenant_id"`
	}
	var urList []UserRole
	err := a.db.Table("sys_user_role ur").
		Select("ur.user_id, r.code as role_code, ur.tenant_id").
		Joins("left join sys_role r on ur.role_id = r.id").
		Scan(&urList).Error
	if err != nil {
		helper.Errorf("Failed to load user-role mappings: %v", err)
		return err
	}

	helper.Infof("Loaded %d user-role mappings", len(urList))
	for _, ur := range urList {
		// 对应 g(sub, role, dom)
		// 注意：使用字符串格式以确保 Casbin 正确匹配
		policyLine := fmt.Sprintf("g, %d, %s, %d", ur.UserID, ur.RoleCode, ur.TenantID)
		helper.Infof("Loading g policy: %s", policyLine)
		persist.LoadPolicyLine(policyLine, m)
	}

	// 2. 加载角色-权限-范围策略 (p)
	type RolePerm struct {
		RoleCode  string `gorm:"column:role_code"`
		TenantID  string `gorm:"column:tenant_id"`
		PermCode  string `gorm:"column:perm_code"`
		DataScope string `gorm:"column:data_scope"`
	}
	var rpList []RolePerm
	err = a.db.Table("sys_role_permission rp").
		Select("r.code as role_code, rp.tenant_id, p.code as perm_code, rp.data_scope").
		Joins("left join sys_role r on rp.role_id = r.id").
		Joins("left join sys_permission p on rp.permission_id = p.id").
		Scan(&rpList).Error
	if err != nil {
		helper.Errorf("Failed to load role-permission mappings: %v", err)
		return err
	}

	helper.Infof("Loaded %d role-permission mappings", len(rpList))
	for _, rp := range rpList {
		// 对应 p(sub, dom, obj, act, scope)
		policyLine := fmt.Sprintf("p, %s, %s, %s, V, %s", rp.RoleCode, rp.TenantID, rp.PermCode, rp.DataScope)
		helper.Infof("Loading p policy: %s", policyLine)
		persist.LoadPolicyLine(policyLine, m)
	}
	return nil
}

// 以下方法留空，因为权限由业务 Service 管理

func (a *SysPermissionAdapter) SavePolicy(m model.Model) error                          { return nil }
func (a *SysPermissionAdapter) AddPolicy(sec string, ptype string, rule []string) error { return nil }
func (a *SysPermissionAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}
func (a *SysPermissionAdapter) RemoveFilteredPolicy(sec string, ptype string, f int, v ...string) error {
	return nil
}
