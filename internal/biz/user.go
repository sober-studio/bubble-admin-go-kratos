package biz

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// UserUseCase 用户业务逻辑
type UserUseCase struct {
	repo SysUserRepo
	log  *log.Helper
}

func NewUserUseCase(repo SysUserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *UserUseCase) Create(ctx context.Context, username, passwordHash, nickname, phone string, deptID int64, roleIDs []int64) (*SysUser, error) {
	// 检查用户是否已存在
	existing, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	user := &SysUser{
		Username:     username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
		Phone:        phone,
		DeptID:       deptID,
		Status:       1, // 默认正常
		IsAvailable:  true,
	}

	createdUser, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// 分配角色
	if len(roleIDs) > 0 {
		if err := uc.repo.AssignRoles(ctx, createdUser.ID, roleIDs); err != nil {
			return nil, err
		}
	}

	return createdUser, nil
}

func (uc *UserUseCase) GetByID(ctx context.Context, id int64) (*SysUser, []int64, error) {
	user, err := uc.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// 获取用户角色
	roleIDs, err := uc.repo.GetUserRoles(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return user, roleIDs, nil
}

func (uc *UserUseCase) Update(ctx context.Context, id int64, nickname, phone string, deptID int64) error {
	user, err := uc.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	user.Nickname = nickname
	user.Phone = phone
	user.DeptID = deptID

	return uc.repo.UpdateUser(ctx, user)
}

func (uc *UserUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.DeleteUser(ctx, id)
}

func (uc *UserUseCase) List(ctx context.Context, username, mobile string, status int32, deptID int64, page, pageSize int64) ([]*SysUser, int64, error) {
	return uc.repo.List(ctx, username, mobile, status, deptID, page, pageSize)
}

func (uc *UserUseCase) AssignRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	// 检查用户是否存在
	_, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return uc.repo.AssignRoles(ctx, userID, roleIDs)
}

func (uc *UserUseCase) SetStatus(ctx context.Context, id int64, status int32) error {
	// 检查用户是否存在
	_, err := uc.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return uc.repo.SetStatus(ctx, id, status)
}

func (uc *UserUseCase) ResetPassword(ctx context.Context, id int64, passwordHash string) error {
	// 检查用户是否存在
	_, err := uc.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return uc.repo.UpdatePassword(ctx, id, passwordHash)
}
