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
	data *Data
	log  *log.Helper
}

func NewSysUserRepo(data *Data, logger log.Logger) biz.SysUserRepo {
	return &sysUserRepo{
		data: data,
		log:  log.NewHelper(logger),
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

func (r *sysUserRepo) toBiz(u *model.SysUser) *biz.SysUser {
	return &biz.SysUser{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Phone:        u.Mobile,
		Nickname:     u.Name,
		IsAvailable:  u.Status == 1,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
