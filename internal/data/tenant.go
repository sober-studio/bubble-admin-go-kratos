package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz/provider"
)

type tenantRepo struct {
	data *Data
	log  *log.Helper
}

func NewTenantRepo(data *Data, logger log.Logger) provider.PackageLoader {
	return &tenantRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

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
