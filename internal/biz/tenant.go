package biz

import (
	"context"
	"errors"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

var (
	ErrTenantNotFound     = kerrors.NotFound("TENANT_NOT_FOUND", "租户不存在")
	ErrTenantAlreadyExist = kerrors.Conflict("TENANT_ALREADY_EXIST", "租户编码已存在")
)

// SysTenant 租户领域模型
type SysTenant struct {
	ID        int64
	Name      string
	Code      string
	Status    int32
	PackageID int64
	ExpiredAt int64
	CreatedAt int64
	UpdatedAt int64
}

// SysTenantRepo 租户仓储接口
type SysTenantRepo interface {
	Create(ctx context.Context, tenant *SysTenant) (*SysTenant, error)
	GetByID(ctx context.Context, id int64) (*SysTenant, error)
	GetByCode(ctx context.Context, code string) (*SysTenant, error)
	Update(ctx context.Context, tenant *SysTenant) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, name, code string, status int32, page, pageSize int64) ([]*SysTenant, int64, error)
}

// TenantUseCase 租户业务逻辑
type TenantUseCase struct {
	repo        SysTenantRepo
	packageRepo SysPackageRepo
	log         *log.Helper
}

func NewTenantUseCase(repo SysTenantRepo, packageRepo SysPackageRepo, logger log.Logger) *TenantUseCase {
	return &TenantUseCase{
		repo:        repo,
		packageRepo: packageRepo,
		log:         log.NewHelper(logger),
	}
}

func (uc *TenantUseCase) Create(ctx context.Context, name, code string, packageID int64, expiredAt int64) (*SysTenant, error) {
	// 检查租户编码是否已存在
	existing, err := uc.repo.GetByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTenantAlreadyExist
	}

	// 检查套餐是否存在
	if packageID > 0 {
		_, err := uc.packageRepo.GetByID(ctx, packageID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrPackageNotFound
			}
			return nil, err
		}
	}

	tenant := &SysTenant{
		Name:      name,
		Code:      code,
		Status:    1, // 默认正常
		PackageID: packageID,
		ExpiredAt: expiredAt,
	}

	return uc.repo.Create(ctx, tenant)
}

func (uc *TenantUseCase) GetByID(ctx context.Context, id int64) (*SysTenant, int64, string, error) {
	tenant, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, 0, "", err
	}

	// 获取套餐信息
	var packageName string
	if tenant.PackageID > 0 {
		pkg, err := uc.packageRepo.GetByID(ctx, tenant.PackageID)
		if err == nil {
			packageName = pkg.Name
		}
	}

	return tenant, tenant.PackageID, packageName, nil
}

func (uc *TenantUseCase) Update(ctx context.Context, id int64, name, code string, packageID int64, expiredAt int64) error {
	// 检查租户是否存在
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTenantNotFound
		}
		return err
	}

	// 检查租户编码是否已存在（排除自己）
	if code != existing.Code {
		other, err := uc.repo.GetByCode(ctx, code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if other != nil {
			return ErrTenantAlreadyExist
		}
	}

	// 检查套餐是否存在
	if packageID > 0 {
		_, err := uc.packageRepo.GetByID(ctx, packageID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPackageNotFound
			}
			return err
		}
	}

	existing.Name = name
	existing.Code = code
	existing.PackageID = packageID
	existing.ExpiredAt = expiredAt

	return uc.repo.Update(ctx, existing)
}

func (uc *TenantUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *TenantUseCase) List(ctx context.Context, name, code string, status int32, page, pageSize int64) ([]*SysTenant, int64, error) {
	return uc.repo.List(ctx, name, code, status, page, pageSize)
}
