package biz

import (
	"context"
	"errors"
	"strconv"
	"time"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/conf"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = kerrors.NotFound("USER_NOT_FOUND", "用户不存在")
	ErrUserAlreadyExists  = kerrors.Conflict("USER_ALREADY_EXISTS", "用户已存在")
	ErrPasswordInvalid    = kerrors.BadRequest("PASSWORD_INVALID", "密码错误")
	ErrMobileAlreadyBound = kerrors.Conflict("MOBILE_ALREADY_BOUND", "手机号已被绑定")
	ErrUserDisabled       = kerrors.Forbidden("USER_DISABLED", "账号已被禁用")
)

type SysUser struct {
	ID           int64
	Username     string
	PasswordHash string
	Phone        string
	Nickname     string
	DeptID       int64
	TenantID     int64
	IsAvailable  bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SysUserRepo interface {
	CreateUser(ctx context.Context, user *SysUser) (*SysUser, error)
	GetUserByUsername(ctx context.Context, username string) (*SysUser, error)
	GetUserByPhone(ctx context.Context, phone string) (*SysUser, error)
	GetUserByID(ctx context.Context, id int64) (*SysUser, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdatePhone(ctx context.Context, id int64, phone string) error
}

type PassportUseCase struct {
	auth    auth.TokenService
	sysUser SysUserRepo
	conf    *conf.App_Auth_Passport
	log     *log.Helper
}

func NewPassportUseCase(
	auth auth.TokenService,
	sysUser SysUserRepo,
	conf *conf.App,
	logger log.Logger,
) *PassportUseCase {
	return &PassportUseCase{
		auth:    auth,
		sysUser: sysUser,
		conf:    conf.Auth.Passport,
		log:     log.NewHelper(logger),
	}
}

func (uc *PassportUseCase) LoginByPassword(ctx context.Context, username, password string) (string, error) {
	// 查询用户
	user, err := uc.sysUser.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			// 如果按用户名未找到，尝试按手机号查找
			u, errPhone := uc.sysUser.GetUserByPhone(ctx, username)
			if errPhone != nil {
				if errors.Is(errPhone, ErrUserNotFound) {
					return "", ErrUserNotFound
				}
				return "", errPhone
			}
			user = u
		} else {
			return "", err
		}
	}

	// 校验密码
	if !uc.checkPassword(password, user.PasswordHash) {
		return "", ErrPasswordInvalid
	}

	if !user.IsAvailable {
		return "", ErrUserDisabled
	}

	return uc.auth.GenerateToken(ctx, uc.formatUserID(user.ID), user.DeptID, user.TenantID)
}

func (uc *PassportUseCase) LoginByOtp(ctx context.Context, phone string) (string, error) {
	// 查询用户
	user, err := uc.sysUser.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	if !user.IsAvailable {
		return "", ErrUserDisabled
	}

	return uc.auth.GenerateToken(ctx, uc.formatUserID(user.ID), user.DeptID, user.TenantID)
}

func (uc *PassportUseCase) Logout(ctx context.Context) error {
	// 撤销当前 Token
	return uc.auth.RevokeToken(ctx, "")
}

func (uc *PassportUseCase) UserInfo(ctx context.Context) (*SysUser, error) {
	userId, err := uc.auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return uc.sysUser.GetUserByID(ctx, userId)
}

func (uc *PassportUseCase) UpdatePassword(ctx context.Context, oldPassword, newPassword string) error {
	userId, err := uc.auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	user, err := uc.sysUser.GetUserByID(ctx, userId)
	if err != nil {
		return err
	}

	if !uc.checkPassword(oldPassword, user.PasswordHash) {
		return ErrPasswordInvalid
	}

	hash, err := uc.hashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := uc.sysUser.UpdatePassword(ctx, userId, hash); err != nil {
		return err
	}

	// 密码修改完成后，撤销用户所有的令牌
	return uc.auth.RevokeAllTokens(ctx)
}

func (uc *PassportUseCase) BindMobile(ctx context.Context, mobile string) error {
	userId, err := uc.auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 检查手机号是否已被使用
	if u, _ := uc.sysUser.GetUserByPhone(ctx, mobile); u != nil {
		return ErrMobileAlreadyBound
	}

	return uc.sysUser.UpdatePhone(ctx, userId, mobile)
}

func (uc *PassportUseCase) UpdateMobile(ctx context.Context, mobile string) error {
	userId, err := uc.auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// 检查手机号是否已被使用
	if u, _ := uc.sysUser.GetUserByPhone(ctx, mobile); u != nil {
		return ErrMobileAlreadyBound
	}

	return uc.sysUser.UpdatePhone(ctx, userId, mobile)
}

// CheckPhoneRegistered 检查手机号是否已注册
func (uc *PassportUseCase) CheckPhoneRegistered(ctx context.Context, phone string) error {
	_, err := uc.sysUser.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (uc *PassportUseCase) ResetPassword(ctx context.Context, mobile, newPassword string) error {
	user, err := uc.sysUser.GetUserByPhone(ctx, mobile)
	if err != nil {
		return ErrUserNotFound
	}

	hash, err := uc.hashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := uc.sysUser.UpdatePassword(ctx, user.ID, hash); err != nil {
		return err
	}

	// 密码重置完成后，撤销用户所有的令牌
	if err := uc.auth.RevokeAllTokensByUserID(ctx, user.ID); err != nil {
		return err
	}
	return nil
}

func (uc *PassportUseCase) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (uc *PassportUseCase) checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (uc *PassportUseCase) formatUserID(id int64) string {
	return strconv.FormatInt(id, 10)
}
