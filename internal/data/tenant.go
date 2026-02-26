package data

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz/provider"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/data/model"
	"gorm.io/gorm"
)

var _ biz.SysTenantRepo = (*tenantRepo)(nil)
var _ provider.PackageLoader = (*tenantRepo)(nil)

type tenantRepo struct {
	data *Data
	log  *log.Helper
}

func NewTenantRepo(data *Data, logger log.Logger) biz.SysTenantRepo {
	return &tenantRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func NewTenantPackageLoader(data *Data, logger log.Logger) provider.PackageLoader {
	return &tenantRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// PackageLoader 接口实现 - 加载所有租户套餐权限
func (r *tenantRepo) LoadAllTenantPackagePerms(ctx context.Context) (map[int64][]string, error) {
	// 定义内部临时结构体，用于接收联表查询结果
	type Row struct {
		TenantID int64  `gorm:"column:tenant_id"`
		PermCode string `gorm:"column:perm_code"`
	}
	var rows []Row

	// 三表联查逻辑：
	// 1. sys_tenant (t): 获取租户ID和套餐ID
	// 2. sys_package_permission (pp): 根据套餐ID获取权限ID
	// 3. sys_permission (p): 根据权限ID获取权限编码 (Code)
	err := r.data.db.WithContext(ctx).Table("sys_tenant t").
		Select("t.id as tenant_id, p.code as perm_code").
		Joins("JOIN sys_package_permission pp ON t.package_id = pp.package_id").
		Joins("JOIN sys_permission p ON pp.permission_id = p.id").
		// 过滤已软删除的记录（假设 sys_tenant 和 sys_permission 使用了 BaseModel）
		Where("t.deleted_at IS NULL AND p.deleted_at IS NULL").
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	// 使用两层 Map 进行去重和分组
	// 第一层 key: TenantID
	// 第二层 key: PermCode (利用 map key 唯一性去重)
	tempMap := make(map[int64]map[string]struct{})

	for _, row := range rows {
		if row.PermCode == "" {
			continue
		}

		// 如果该租户还没在 map 中，初始化它的权限集合
		if _, ok := tempMap[row.TenantID]; !ok {
			tempMap[row.TenantID] = make(map[string]struct{})
		}

		// 自动去重：即使 SQL 查出重复项，这里也会被覆盖
		tempMap[row.TenantID][row.PermCode] = struct{}{}
	}

	// 转换为最终要求的 map[int64][]string 格式
	result := make(map[int64][]string)
	for tid, codesMap := range tempMap {
		codes := make([]string, 0, len(codesMap))
		for code := range codesMap {
			codes = append(codes, code)
		}
		result[tid] = codes
	}

	return result, nil
}

func (r *tenantRepo) Create(ctx context.Context, tenant *biz.SysTenant) (*biz.SysTenant, error) {
	m := &model.SysTenant{
		Name:      tenant.Name,
		Code:      tenant.Code,
		Status:    int16(tenant.Status),
		PackageID: tenant.PackageID,
		ExpiredAt: tenant.ExpiredAt,
	}

	if err := r.data.DB(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	return r.toBiz(m), nil
}

func (r *tenantRepo) GetByID(ctx context.Context, id int64) (*biz.SysTenant, error) {
	var tenant model.SysTenant
	if err := r.data.DB(ctx).Where("id = ?", id).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrTenantNotFound
		}
		return nil, err
	}
	return r.toBiz(&tenant), nil
}

func (r *tenantRepo) GetByCode(ctx context.Context, code string) (*biz.SysTenant, error) {
	var tenant model.SysTenant
	if err := r.data.DB(ctx).Where("code = ?", code).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrTenantNotFound
		}
		return nil, err
	}
	return r.toBiz(&tenant), nil
}

func (r *tenantRepo) Update(ctx context.Context, tenant *biz.SysTenant) error {
	return r.data.DB(ctx).
		Model(&model.SysTenant{}).
		Where("id = ?", tenant.ID).
		Updates(map[string]interface{}{
			"name":       tenant.Name,
			"code":       tenant.Code,
			"package_id": tenant.PackageID,
			"expired_at": tenant.ExpiredAt,
		}).Error
}

func (r *tenantRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Where("id = ?", id).Delete(&model.SysTenant{}).Error
}

func (r *tenantRepo) List(ctx context.Context, name, code string, status int32, page, pageSize int64) ([]*biz.SysTenant, int64, error) {
	var tenants []model.SysTenant
	var total int64

	db := r.data.DB(ctx)

	// 条件查询
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		db = db.Where("code LIKE ?", "%"+code+"%")
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	// 统计总数
	if err := db.Model(&model.SysTenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(int(offset)).Limit(int(pageSize)).Order("id DESC").Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*biz.SysTenant, len(tenants))
	for i := range tenants {
		result[i] = r.toBiz(&tenants[i])
	}
	return result, total, nil
}

func (r *tenantRepo) toBiz(m *model.SysTenant) *biz.SysTenant {
	return &biz.SysTenant{
		ID:        m.ID,
		Name:      m.Name,
		Code:      m.Code,
		Status:    int32(m.Status),
		PackageID: m.PackageID,
		ExpiredAt: m.ExpiredAt,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
	}
}
