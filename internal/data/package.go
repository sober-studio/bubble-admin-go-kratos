package data

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/data/model"
	"gorm.io/gorm"
)

var _ biz.SysPackageRepo = (*packageRepo)(nil)

type packageRepo struct {
	BaseRepo
	data *Data
	log  *log.Helper
}

func NewPackageRepo(data *Data, logger log.Logger) biz.SysPackageRepo {
	return &packageRepo{
		BaseRepo: NewBaseRepo(data, logger),
		data:     data,
		log:      log.NewHelper(logger),
	}
}

func (r *packageRepo) Create(ctx context.Context, pkg *biz.SysPackage) (*biz.SysPackage, error) {
	m := &model.SysPackage{
		Name:   pkg.Name,
		Code:   pkg.Code,
		Remark: pkg.Remark,
		Status: int16(pkg.Status),
		BaseAuthModel: model.BaseAuthModel{
			AuthField: model.AuthField{
				TenantID: pkg.TenantID,
			},
		},
	}

	if err := r.data.DB(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	return r.toBiz(m), nil
}

func (r *packageRepo) GetByID(ctx context.Context, id int64) (*biz.SysPackage, error) {
	var pkg model.SysPackage
	if err := r.data.DB(ctx).Where("id = ?", id).First(&pkg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPackageNotFound
		}
		return nil, err
	}
	return r.toBiz(&pkg), nil
}

func (r *packageRepo) GetByCode(ctx context.Context, code string) (*biz.SysPackage, error) {
	var pkg model.SysPackage
	if err := r.data.DB(ctx).Where("code = ?", code).First(&pkg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrPackageNotFound
		}
		return nil, err
	}
	return r.toBiz(&pkg), nil
}

func (r *packageRepo) Update(ctx context.Context, pkg *biz.SysPackage) error {
	return r.data.DB(ctx).
		Model(&model.SysPackage{}).
		Where("id = ?", pkg.ID).
		Updates(map[string]interface{}{
			"name":   pkg.Name,
			"code":   pkg.Code,
			"remark": pkg.Remark,
		}).Error
}

func (r *packageRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Where("id = ?", id).Delete(&model.SysPackage{}).Error
}

func (r *packageRepo) List(ctx context.Context, name string, status int32, page, pageSize int64) ([]*biz.SysPackage, int64, error) {
	var pkgs []model.SysPackage
	var total int64

	db := r.data.DB(ctx)

	// 条件查询
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	// 统计总数
	if err := db.Model(&model.SysPackage{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&pkgs).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*biz.SysPackage, len(pkgs))
	for i := range pkgs {
		result[i] = r.toBiz(&pkgs[i])
	}
	return result, total, nil
}

func (r *packageRepo) GetPermissions(ctx context.Context, packageID int64) ([]int64, error) {
	var pkgPerms []model.SysPackagePermission
	if err := r.data.DB(ctx).Where("package_id = ?", packageID).Find(&pkgPerms).Error; err != nil {
		return nil, err
	}

	permIDs := make([]int64, len(pkgPerms))
	for i, p := range pkgPerms {
		permIDs[i] = p.PermissionID
	}
	return permIDs, nil
}

func (r *packageRepo) AssignPermissions(ctx context.Context, packageID int64, permissionIDs []int64) error {
	return r.data.InTx(ctx, func(ctx context.Context) error {
		// 删除旧权限
		if err := r.data.DB(ctx).Where("package_id = ?", packageID).Delete(&model.SysPackagePermission{}).Error; err != nil {
			return err
		}

		// 添加新权限
		if len(permissionIDs) > 0 {
			pkgPerms := make([]*model.SysPackagePermission, len(permissionIDs))
			for i, permID := range permissionIDs {
				pkgPerms[i] = &model.SysPackagePermission{
					PackageID:    packageID,
					PermissionID: permID,
				}
			}
			if err := r.data.DB(ctx).Create(pkgPerms).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *packageRepo) toBiz(m *model.SysPackage) *biz.SysPackage {
	return &biz.SysPackage{
		ID:       m.ID,
		Name:     m.Name,
		Code:     m.Code,
		Remark:   m.Remark,
		Status:   int32(m.Status),
		TenantID: m.TenantID,
		CreatedAt:   m.CreatedAt.Unix(),
		UpdatedAt:   m.UpdatedAt.Unix(),
	}
}
