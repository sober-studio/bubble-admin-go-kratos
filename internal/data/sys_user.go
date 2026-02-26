package data

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/data/model"
	"gorm.io/gorm"
)

var _ biz.SysUserRepo = (*sysUserRepo)(nil)

type sysUserRepo struct {
	BaseRepo
	data *Data
	log  *log.Helper
}

func NewSysUserRepo(data *Data, logger log.Logger) biz.SysUserRepo {
	return &sysUserRepo{
		BaseRepo: NewBaseRepo(data, logger),
		data:     data,
		log:      log.NewHelper(logger),
	}
}

func (r *sysUserRepo) CreateUser(ctx context.Context, u *biz.SysUser) (*biz.SysUser, error) {
	status := int16(2)
	if u.IsAvailable {
		status = 1
	}
	user := &model.SysUser{
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Mobile:       u.Phone,
		Name:         u.Nickname,
		Status:       status,
		BaseAuthModel: model.BaseAuthModel{
			AuthField: model.AuthField{
				DeptID:   u.DeptID,
				TenantID: u.TenantID,
			},
		},
	}

	if err := r.data.DB(ctx).Create(user).Error; err != nil {
		return nil, err
	}

	return r.toBiz(user), nil
}

func (r *sysUserRepo) GetUserByUsername(ctx context.Context, username string) (*biz.SysUser, error) {
	var user model.SysUser
	if err := r.data.DB(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return r.toBiz(&user), nil
}

func (r *sysUserRepo) GetUserByPhone(ctx context.Context, phone string) (*biz.SysUser, error) {
	var user model.SysUser
	if err := r.data.DB(ctx).Where("mobile = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return r.toBiz(&user), nil
}

func (r *sysUserRepo) GetUserByID(ctx context.Context, id int64) (*biz.SysUser, error) {
	var user model.SysUser
	if err := r.data.DB(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return r.toBiz(&user), nil
}

func (r *sysUserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	return r.data.DB(ctx).
		Model(&model.SysUser{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *sysUserRepo) UpdatePhone(ctx context.Context, id int64, phone string) error {
	return r.data.DB(ctx).
		Model(&model.SysUser{}).
		Where("id = ?", id).
		Update("mobile", phone).Error
}

func (r *sysUserRepo) UpdateUser(ctx context.Context, u *biz.SysUser) error {
	return r.data.DB(ctx).
		Model(&model.SysUser{}).
		Where("id = ?", u.ID).
		Updates(map[string]interface{}{
			"name":     u.Nickname,
			"mobile":   u.Phone,
			"dept_id":  u.DeptID,
		}).Error
}

func (r *sysUserRepo) List(ctx context.Context, username, mobile string, status int32, deptID int64, page, pageSize int64) ([]*biz.SysUser, int64, error) {
	var users []model.SysUser
	var total int64

	db := r.data.DB(ctx)

	// 条件查询
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if mobile != "" {
		db = db.Where("mobile LIKE ?", "%"+mobile+"%")
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}
	if deptID > 0 {
		db = db.Where("dept_id = ?", deptID)
	}

	// 统计总数
	if err := db.Model(&model.SysUser{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*biz.SysUser, len(users))
	for i := range users {
		result[i] = r.toBiz(&users[i])
	}
	return result, total, nil
}

func (r *sysUserRepo) DeleteUser(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Where("id = ?", id).Delete(&model.SysUser{}).Error
}

func (r *sysUserRepo) SetStatus(ctx context.Context, id int64, status int32) error {
	return r.data.DB(ctx).
		Model(&model.SysUser{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *sysUserRepo) GetUserRoles(ctx context.Context, userID int64) ([]int64, error) {
	var userRoles []model.SysUserRole
	if err := r.data.DB(ctx).Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	roleIDs := make([]int64, len(userRoles))
	for i, ur := range userRoles {
		roleIDs[i] = ur.RoleID
	}
	return roleIDs, nil
}

func (r *sysUserRepo) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return r.data.InTx(ctx, func(ctx context.Context) error {
		// 删除旧角色
		if err := r.data.DB(ctx).Where("user_id = ?", userID).Delete(&model.SysUserRole{}).Error; err != nil {
			return err
		}

		// 添加新角色
		if len(roleIDs) > 0 {
			userRoles := make([]*model.SysUserRole, len(roleIDs))
			for i, roleID := range roleIDs {
				userRoles[i] = &model.SysUserRole{
					UserID: userID,
					RoleID: roleID,
				}
			}
			if err := r.data.DB(ctx).Create(userRoles).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *sysUserRepo) toBiz(u *model.SysUser) *biz.SysUser {
	return &biz.SysUser{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Phone:        u.Mobile,
		Nickname:     u.Name,
		DeptID:       u.DeptID,
		TenantID:     u.TenantID,
		Status:       int32(u.Status),
		IsAvailable:  u.Status == 1,
		CreatedAt:    u.CreatedAt.Unix(),
		UpdatedAt:    u.UpdatedAt.Unix(),
	}
}
