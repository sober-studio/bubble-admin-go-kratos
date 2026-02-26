package data

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/data/model"
	"gorm.io/gorm"
)

var _ biz.SysRoleRepo = (*roleRepo)(nil)

type roleRepo struct {
	BaseRepo
	data *Data
	log  *log.Helper
}

func NewRoleRepo(data *Data, logger log.Logger) biz.SysRoleRepo {
	return &roleRepo{
		BaseRepo: NewBaseRepo(data, logger),
		data:     data,
		log:      log.NewHelper(logger),
	}
}

func (r *roleRepo) Create(ctx context.Context, role *biz.SysRole) (*biz.SysRole, error) {
	m := &model.SysRole{
		Name: role.Name,
		Code: role.Code,
		BaseAuthModel: model.BaseAuthModel{
			AuthField: model.AuthField{
				TenantID: role.TenantID,
			},
		},
	}

	if err := r.data.DB(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	return r.toBiz(m), nil
}

func (r *roleRepo) GetByID(ctx context.Context, id int64) (*biz.SysRole, error) {
	var role model.SysRole
	if err := r.data.DB(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrRoleNotFound
		}
		return nil, err
	}
	return r.toBiz(&role), nil
}

func (r *roleRepo) GetByCode(ctx context.Context, code string) (*biz.SysRole, error) {
	var role model.SysRole
	if err := r.data.DB(ctx).Where("code = ?", code).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrRoleNotFound
		}
		return nil, err
	}
	return r.toBiz(&role), nil
}

func (r *roleRepo) Update(ctx context.Context, role *biz.SysRole) error {
	return r.data.DB(ctx).
		Model(&model.SysRole{}).
		Where("id = ?", role.ID).
		Updates(map[string]interface{}{
			"name": role.Name,
			"code": role.Code,
		}).Error
}

func (r *roleRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Where("id = ?", id).Delete(&model.SysRole{}).Error
}

func (r *roleRepo) List(ctx context.Context, name, code string, page, pageSize int64) ([]*biz.SysRole, int64, error) {
	var roles []model.SysRole
	var total int64

	db := r.data.DB(ctx)

	// 条件查询
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		db = db.Where("code LIKE ?", "%"+code+"%")
	}

	// 统计总数
	if err := db.Model(&model.SysRole{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*biz.SysRole, len(roles))
	for i := range roles {
		result[i] = r.toBiz(&roles[i])
	}
	return result, total, nil
}

func (r *roleRepo) HasUsers(ctx context.Context, roleID int64) (bool, error) {
	var count int64
	err := r.data.DB(ctx).Model(&model.SysUserRole{}).Where("role_id = ?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roleRepo) GetPermissions(ctx context.Context, roleID int64) ([]*biz.SysRolePermission, error) {
	var perms []model.SysRolePermission
	if err := r.data.DB(ctx).Where("role_id = ?", roleID).Find(&perms).Error; err != nil {
		return nil, err
	}

	result := make([]*biz.SysRolePermission, len(perms))
	for i := range perms {
		result[i] = &biz.SysRolePermission{
			ID:           perms[i].ID,
			RoleID:       perms[i].RoleID,
			PermissionID: perms[i].PermissionID,
		}
	}
	return result, nil
}

func (r *roleRepo) AssignPermissions(ctx context.Context, roleID int64, permissionIDs []int64, dataScope int32, deptIDs []int64) error {
	return r.data.InTx(ctx, func(ctx context.Context) error {
		// 删除旧权限
		if err := r.data.DB(ctx).Where("role_id = ?", roleID).Delete(&model.SysRolePermission{}).Error; err != nil {
			return err
		}

		// 添加新权限
		if len(permissionIDs) > 0 {
			deptIDStrs := make([]string, len(deptIDs))
			for i, d := range deptIDs {
				deptIDStrs[i] = strconv.FormatInt(d, 10)
			}

			rolePerms := make([]*model.SysRolePermission, len(permissionIDs))
			for i, permID := range permissionIDs {
				rolePerms[i] = &model.SysRolePermission{
					RoleID:       roleID,
					PermissionID: permID,
					DataScope:    fmt.Sprintf("%d", dataScope),
					DeptIDs:      strings.Join(deptIDStrs, ","),
				}
			}

			if err := r.data.DB(ctx).Create(rolePerms).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *roleRepo) toBiz(m *model.SysRole) *biz.SysRole {
	return &biz.SysRole{
		ID:        m.ID,
		Name:      m.Name,
		Code:      m.Code,
		TenantID:  m.TenantID,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
	}
}
