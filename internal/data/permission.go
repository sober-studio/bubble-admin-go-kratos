package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz/provider"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/data/model"
)

type permissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewPermissionRepo(data *Data, logger log.Logger) provider.PermissionLoader {
	return &permissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *permissionRepo) LoadAllApiPermissions(ctx context.Context) (map[string][]string, error) {
	var list []model.SysPermission
	err := r.data.DB(ctx).Where("type = ?", "API").Find(&list).Error
	if err != nil {
		return nil, err
	}
	results := make(map[string][]string)
	for _, p := range list {
		results[p.APIPath] = append(results[p.APIPath], p.Code)
	}
	return results, err
}
