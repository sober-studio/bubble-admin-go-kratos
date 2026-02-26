package data

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz/provider"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/data/model"
	"gorm.io/gorm"
)

var _ biz.SysPermissionRepo = (*permissionRepo)(nil)
var _ provider.PermissionLoader = (*permissionRepo)(nil)

type permissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewPermissionRepo(data *Data, logger log.Logger) biz.SysPermissionRepo {
	return &permissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func NewPermissionLoader(data *Data, logger log.Logger) provider.PermissionLoader {
	return &permissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// PermissionLoader 接口实现 - 加载所有 API 权限
func (r *permissionRepo) LoadAllApiPermissions(ctx context.Context) (map[string][]string, error) {
	var list []model.SysPermission
	err := r.data.DB(ctx).Where("type = ?", "API").Find(&list).Error
	if err != nil {
		return nil, err
	}
	results := make(map[string][]string)
	for _, p := range list {
		results[p.APIPath] = append(results[p.APIPath], p.Code)
	}
	return results, err
}

// Create 创建权限
func (r *permissionRepo) Create(ctx context.Context, perm *biz.SysPermission) (*biz.SysPermission, error) {
	m := &model.SysPermission{
		ParentID:   perm.ParentID,
		Name:       perm.Name,
		Code:       perm.Code,
		Type:       perm.Type,
		APIPath:    perm.APIPath,
		APIMethod:  perm.APIMethod,
		Sort:       perm.Sort,
		Path:       perm.Path,
		Component:  perm.Component,
		Redirect:   perm.Redirect,
		Icon:       perm.Icon,
		OrderNo:    perm.OrderNo,
		Hidden:     perm.Hidden,
		KeepAlive:  perm.KeepAlive,
		FrameSrc:   perm.FrameSrc,
		FrameBlank: perm.FrameBlank,
		BaseAuthModel: model.BaseAuthModel{
			AuthField: model.AuthField{
				TenantID: perm.TenantID,
			},
		},
	}

	if err := r.data.DB(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	return r.toBiz(m), nil
}

// GetByID 根据ID获取权限
func (r *permissionRepo) GetByID(ctx context.Context, id int64) (*biz.SysPermission, error) {
	var perm model.SysPermission
	if err := r.data.DB(ctx).Where("id = ?", id).First(&perm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPermissionNotFound
		}
		return nil, err
	}
	return r.toBiz(&perm), nil
}

// GetByCode 根据编码获取权限
func (r *permissionRepo) GetByCode(ctx context.Context, code string) (*biz.SysPermission, error) {
	var perm model.SysPermission
	if err := r.data.DB(ctx).Where("code = ?", code).First(&perm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPermissionNotFound
		}
		return nil, err
	}
	return r.toBiz(&perm), nil
}

// Update 更新权限
func (r *permissionRepo) Update(ctx context.Context, perm *biz.SysPermission) error {
	return r.data.DB(ctx).
		Model(&model.SysPermission{}).
		Where("id = ?", perm.ID).
		Updates(map[string]interface{}{
			"parent_id":   perm.ParentID,
			"name":        perm.Name,
			"code":        perm.Code,
			"type":        perm.Type,
			"api_path":    perm.APIPath,
			"api_method":  perm.APIMethod,
			"sort":        perm.Sort,
			"path":        perm.Path,
			"component":   perm.Component,
			"redirect":    perm.Redirect,
			"icon":        perm.Icon,
			"order_no":    perm.OrderNo,
			"hidden":      perm.Hidden,
			"keep_alive":  perm.KeepAlive,
			"frame_src":   perm.FrameSrc,
			"frame_blank": perm.FrameBlank,
		}).Error
}

// Delete 删除权限
func (r *permissionRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Where("id = ?", id).Delete(&model.SysPermission{}).Error
}

// List 获取所有权限
func (r *permissionRepo) List(ctx context.Context) ([]*biz.SysPermission, error) {
	var perms []model.SysPermission
	if err := r.data.DB(ctx).Order("sort ASC, id ASC").Find(&perms).Error; err != nil {
		return nil, err
	}

	result := make([]*biz.SysPermission, len(perms))
	for i := range perms {
		result[i] = r.toBiz(&perms[i])
	}
	return result, nil
}

// HasChildren 检查是否有子权限
func (r *permissionRepo) HasChildren(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.data.DB(ctx).Model(&model.SysPermission{}).Where("parent_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *permissionRepo) toBiz(m *model.SysPermission) *biz.SysPermission {
	return &biz.SysPermission{
		ID:         m.ID,
		ParentID:   m.ParentID,
		Name:       m.Name,
		Code:       m.Code,
		Type:       m.Type,
		APIPath:    m.APIPath,
		APIMethod:  m.APIMethod,
		Sort:       m.Sort,
		TenantID:   m.TenantID,
		CreatedAt:  m.CreatedAt.Unix(),
		UpdatedAt:  m.UpdatedAt.Unix(),
		Path:       m.Path,
		Component:  m.Component,
		Redirect:   m.Redirect,
		Icon:       m.Icon,
		OrderNo:    m.OrderNo,
		Hidden:     m.Hidden,
		KeepAlive: m.KeepAlive,
		FrameSrc:   m.FrameSrc,
		FrameBlank: m.FrameBlank,
	}
}

// GetUserMenuByUserID 根据用户ID获取菜单权限
func (r *permissionRepo) GetUserMenuByUserID(ctx context.Context, userID int64, tenantID int64) ([]*biz.SysPermission, error) {
	// 1. 查询用户关联的角色
	var userRoles []model.SysUserRole
	if err := r.data.DB(ctx).Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return nil, nil
	}

	// 获取角色ID列表
	roleIDs := make([]int64, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}

	// 2. 查询角色关联的权限（仅 MENU 类型）
	var rolePerms []model.SysRolePermission
	if err := r.data.DB(ctx).Where("role_id IN ?", roleIDs).Find(&rolePerms).Error; err != nil {
		return nil, err
	}

	if len(rolePerms) == 0 {
		return nil, nil
	}

	// 获取权限ID列表
	permIDs := make([]int64, len(rolePerms))
	for i, rp := range rolePerms {
		permIDs[i] = rp.PermissionID
	}

	// 3. 查询权限详情（仅 MENU 类型）
	var perms []model.SysPermission
	db := r.data.DB(ctx).Where("id IN ? AND type = ?", permIDs, biz.PermissionTypeMenu)
	if tenantID > 0 {
		db = db.Where("tenant_id = ?", tenantID)
	}
	if err := db.Order("sort ASC, id ASC").Find(&perms).Error; err != nil {
		return nil, err
	}

	// 4. 转换为 biz 模型
	result := make([]*biz.SysPermission, len(perms))
	for i := range perms {
		result[i] = r.toBiz(&perms[i])
	}

	return result, nil
}
