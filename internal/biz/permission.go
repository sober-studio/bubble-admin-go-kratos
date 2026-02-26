package biz

import (
	"context"
	"errors"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

var (
	ErrPermissionNotFound     = kerrors.NotFound("PERMISSION_NOT_FOUND", "权限不存在")
	ErrPermissionAlreadyExist = kerrors.Conflict("PERMISSION_ALREADY_EXIST", "权限编码已存在")
	ErrPermissionHasChildren  = kerrors.Conflict("PERMISSION_HAS_CHILDREN", "权限存在子权限，无法删除")
)

// PermissionType 权限类型
const (
	PermissionTypeMenu   = "MENU"
	PermissionTypeButton = "BUTTON"
	PermissionTypeAPI    = "API"
)

// SysPermission 权限领域模型
type SysPermission struct {
	ID        int64
	ParentID  int64
	Name      string
	Code      string
	Type      string
	APIPath   string
	APIMethod string
	Sort      int32
	TenantID  int64
	CreatedAt int64
	UpdatedAt int64
	Children  []*SysPermission
}

// SysPermissionRepo 权限仓储接口
type SysPermissionRepo interface {
	Create(ctx context.Context, perm *SysPermission) (*SysPermission, error)
	GetByID(ctx context.Context, id int64) (*SysPermission, error)
	GetByCode(ctx context.Context, code string) (*SysPermission, error)
	Update(ctx context.Context, perm *SysPermission) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*SysPermission, error)
	HasChildren(ctx context.Context, id int64) (bool, error)
}

// PermissionUseCase 权限业务逻辑
type PermissionUseCase struct {
	repo SysPermissionRepo
	log  *log.Helper
}

func NewPermissionUseCase(repo SysPermissionRepo, logger log.Logger) *PermissionUseCase {
	return &PermissionUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *PermissionUseCase) Create(ctx context.Context, parentID int64, name, code, permType, apiPath, apiMethod string, sort int32) (*SysPermission, error) {
	// 检查权限编码是否已存在
	existing, err := uc.repo.GetByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPermissionAlreadyExist
	}

	perm := &SysPermission{
		ParentID:  parentID,
		Name:      name,
		Code:      code,
		Type:      permType,
		APIPath:   apiPath,
		APIMethod: apiMethod,
		Sort:      sort,
	}

	return uc.repo.Create(ctx, perm)
}

func (uc *PermissionUseCase) GetByID(ctx context.Context, id int64) (*SysPermission, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *PermissionUseCase) Update(ctx context.Context, id int64, parentID int64, name, code, permType, apiPath, apiMethod string, sort int32) error {
	// 检查权限是否存在
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPermissionNotFound
		}
		return err
	}

	// 检查权限编码是否已存在（排除自己）
	if code != existing.Code {
		other, err := uc.repo.GetByCode(ctx, code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if other != nil {
			return ErrPermissionAlreadyExist
		}
	}

	existing.ParentID = parentID
	existing.Name = name
	existing.Code = code
	existing.Type = permType
	existing.APIPath = apiPath
	existing.APIMethod = apiMethod
	existing.Sort = sort

	return uc.repo.Update(ctx, existing)
}

func (uc *PermissionUseCase) Delete(ctx context.Context, id int64) error {
	// 检查是否有子权限
	hasChildren, err := uc.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return ErrPermissionHasChildren
	}

	return uc.repo.Delete(ctx, id)
}

func (uc *PermissionUseCase) List(ctx context.Context) ([]*SysPermission, error) {
	return uc.repo.List(ctx)
}

// BuildTree 构建权限树
func (uc *PermissionUseCase) BuildTree(perms []*SysPermission) []*SysPermission {
	// 构建 ID 到权限的映射
	permMap := make(map[int64]*SysPermission)
	for i := range perms {
		permMap[perms[i].ID] = perms[i]
	}

	var roots []*SysPermission
	// 构建树结构
	for _, perm := range perms {
		if perm.ParentID == 0 {
			// 根权限
			roots = append(roots, perm)
		} else {
			// 找到父权限，添加到子权限列表
			if parent, ok := permMap[perm.ParentID]; ok {
				parent.Children = append(parent.Children, perm)
			}
		}
	}

	return roots
}
