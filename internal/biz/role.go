package biz

import (
	"context"
	"errors"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound     = kerrors.NotFound("ROLE_NOT_FOUND", "角色不存在")
	ErrRoleAlreadyExist = kerrors.Conflict("ROLE_ALREADY_EXIST", "角色编码已存在")
	ErrRoleHasUsers     = kerrors.Conflict("ROLE_HAS_USERS", "角色存在用户，无法删除")
)

// SysRole 角色领域模型
type SysRole struct {
	ID        int64
	Name      string
	Code      string
	TenantID  int64
	CreatedAt int64
	UpdatedAt int64
}

// SysRolePermission 角色权限关联
type SysRolePermission struct {
	ID           int64
	RoleID       int64
	PermissionID int64
	DataScope    int32
	DeptIDs      string
}

// SysRoleRepo 角色仓储接口
type SysRoleRepo interface {
	Create(ctx context.Context, role *SysRole) (*SysRole, error)
	GetByID(ctx context.Context, id int64) (*SysRole, error)
	GetByCode(ctx context.Context, code string) (*SysRole, error)
	Update(ctx context.Context, role *SysRole) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, name, code string, page, pageSize int64) ([]*SysRole, int64, error)
	HasUsers(ctx context.Context, roleID int64) (bool, error)

	// 角色权限关联
	GetPermissions(ctx context.Context, roleID int64) ([]*SysRolePermission, error)
	AssignPermissions(ctx context.Context, roleID int64, permissionIDs []int64, dataScope int32, deptIDs []int64) error
}

// RoleUseCase 角色业务逻辑
type RoleUseCase struct {
	repo          SysRoleRepo
	permissionRepo SysPermissionRepo
	log           *log.Helper
}

func NewRoleUseCase(repo SysRoleRepo, permissionRepo SysPermissionRepo, logger log.Logger) *RoleUseCase {
	return &RoleUseCase{
		repo:          repo,
		permissionRepo: permissionRepo,
		log:           log.NewHelper(logger),
	}
}

func (uc *RoleUseCase) Create(ctx context.Context, name, code string) (*SysRole, error) {
	// 检查角色编码是否已存在
	existing, err := uc.repo.GetByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRoleAlreadyExist
	}

	role := &SysRole{
		Name: name,
		Code: code,
	}

	return uc.repo.Create(ctx, role)
}

func (uc *RoleUseCase) GetByID(ctx context.Context, id int64) (*SysRole, []int64, error) {
	role, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// 获取角色权限
	perms, err := uc.repo.GetPermissions(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	permIDs := make([]int64, len(perms))
	for i, p := range perms {
		permIDs[i] = p.PermissionID
	}

	return role, permIDs, nil
}

func (uc *RoleUseCase) Update(ctx context.Context, id int64, name, code string) error {
	// 检查角色是否存在
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	// 检查角色编码是否已存在（排除自己）
	if code != existing.Code {
		other, err := uc.repo.GetByCode(ctx, code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if other != nil {
			return ErrRoleAlreadyExist
		}
	}

	existing.Name = name
	existing.Code = code

	return uc.repo.Update(ctx, existing)
}

func (uc *RoleUseCase) Delete(ctx context.Context, id int64) error {
	// 检查是否有用户
	hasUsers, err := uc.repo.HasUsers(ctx, id)
	if err != nil {
		return err
	}
	if hasUsers {
		return ErrRoleHasUsers
	}

	return uc.repo.Delete(ctx, id)
}

func (uc *RoleUseCase) List(ctx context.Context, name, code string, page, pageSize int64) ([]*SysRole, int64, error) {
	return uc.repo.List(ctx, name, code, page, pageSize)
}

func (uc *RoleUseCase) AssignPermissions(ctx context.Context, roleID int64, permissionIDs []int64, dataScope int32, deptIDs []int64) error {
	// 检查角色是否存在
	_, err := uc.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	// 分配权限
	return uc.repo.AssignPermissions(ctx, roleID, permissionIDs, dataScope, deptIDs)
}
