package biz

import (
	"context"
	"errors"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

var (
	ErrPackageNotFound     = kerrors.NotFound("PACKAGE_NOT_FOUND", "套餐不存在")
	ErrPackageAlreadyExist = kerrors.Conflict("PACKAGE_ALREADY_EXIST", "套餐编码已存在")
)

// SysPackage 套餐领域模型
type SysPackage struct {
	ID          int64
	Name        string
	Code        string
	Description string
	Status      int32
	TenantID    int64
	CreatedAt   int64
	UpdatedAt   int64
}

// SysPackagePermission 套餐权限关联
type SysPackagePermission struct {
	ID           int64
	PackageID    int64
	PermissionID int64
}

// SysPackageRepo 套餐仓储接口
type SysPackageRepo interface {
	Create(ctx context.Context, pkg *SysPackage) (*SysPackage, error)
	GetByID(ctx context.Context, id int64) (*SysPackage, error)
	GetByCode(ctx context.Context, code string) (*SysPackage, error)
	Update(ctx context.Context, pkg *SysPackage) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, name string, status int32, page, pageSize int64) ([]*SysPackage, int64, error)
	GetPermissions(ctx context.Context, packageID int64) ([]int64, error)
	AssignPermissions(ctx context.Context, packageID int64, permissionIDs []int64) error
}

// PackageUseCase 套餐业务逻辑
type PackageUseCase struct {
	repo SysPackageRepo
	log  *log.Helper
}

func NewPackageUseCase(repo SysPackageRepo, logger log.Logger) *PackageUseCase {
	return &PackageUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *PackageUseCase) Create(ctx context.Context, name, code, description string) (*SysPackage, error) {
	// 检查套餐编码是否已存在
	existing, err := uc.repo.GetByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPackageAlreadyExist
	}

	pkg := &SysPackage{
		Name:        name,
		Code:        code,
		Description: description,
		Status:      1, // 默认正常
	}

	return uc.repo.Create(ctx, pkg)
}

func (uc *PackageUseCase) GetByID(ctx context.Context, id int64) (*SysPackage, []int64, error) {
	pkg, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// 获取套餐权限
	permIDs, err := uc.repo.GetPermissions(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return pkg, permIDs, nil
}

func (uc *PackageUseCase) Update(ctx context.Context, id int64, name, code, description string) error {
	// 检查套餐是否存在
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPackageNotFound
		}
		return err
	}

	// 检查套餐编码是否已存在（排除自己）
	if code != existing.Code {
		other, err := uc.repo.GetByCode(ctx, code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if other != nil {
			return ErrPackageAlreadyExist
		}
	}

	existing.Name = name
	existing.Code = code
	existing.Description = description

	return uc.repo.Update(ctx, existing)
}

func (uc *PackageUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *PackageUseCase) List(ctx context.Context, name string, status int32, page, pageSize int64) ([]*SysPackage, int64, error) {
	return uc.repo.List(ctx, name, status, page, pageSize)
}

func (uc *PackageUseCase) AssignPermissions(ctx context.Context, packageID int64, permissionIDs []int64) error {
	// 检查套餐是否存在
	_, err := uc.repo.GetByID(ctx, packageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPackageNotFound
		}
		return err
	}

	return uc.repo.AssignPermissions(ctx, packageID, permissionIDs)
}
