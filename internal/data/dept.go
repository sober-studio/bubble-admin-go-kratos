package data

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/data/model"
	"gorm.io/gorm"
)

var _ biz.SysDeptRepo = (*sysDeptRepo)(nil)

type sysDeptRepo struct {
	BaseRepo
	data *Data
	log  *log.Helper
}

func NewSysDeptRepo(data *Data, logger log.Logger) biz.SysDeptRepo {
	return &sysDeptRepo{
		BaseRepo: NewBaseRepo(data, logger),
		data:     data,
		log:      log.NewHelper(logger),
	}
}

func (r *sysDeptRepo) Create(ctx context.Context, dept *biz.SysDept) (*biz.SysDept, error) {
	m := &model.SysDept{
		ParentID:  dept.ParentID,
		Name:      dept.Name,
		Ancestors: dept.Ancestors,
		Sort:      dept.Sort,
		BaseAuthModel: model.BaseAuthModel{
			AuthField: model.AuthField{
				TenantID: dept.TenantID,
			},
		},
	}

	if err := r.data.DB(ctx).Create(m).Error; err != nil {
		return nil, err
	}

	return r.toBiz(m), nil
}

func (r *sysDeptRepo) GetByID(ctx context.Context, id int64) (*biz.SysDept, error) {
	var dept model.SysDept
	if err := r.data.DB(ctx).Where("id = ?", id).First(&dept).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrDeptNotFound
		}
		return nil, err
	}
	return r.toBiz(&dept), nil
}

func (r *sysDeptRepo) Update(ctx context.Context, dept *biz.SysDept) error {
	return r.data.DB(ctx).
		Model(&model.SysDept{}).
		Where("id = ?", dept.ID).
		Updates(map[string]interface{}{
			"parent_id":  dept.ParentID,
			"name":       dept.Name,
			"ancestors":  dept.Ancestors,
			"sort":       dept.Sort,
		}).Error
}

func (r *sysDeptRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).Where("id = ?", id).Delete(&model.SysDept{}).Error
}

func (r *sysDeptRepo) List(ctx context.Context) ([]*biz.SysDept, error) {
	var depts []model.SysDept
	if err := r.data.DB(ctx).Order("sort ASC, id ASC").Find(&depts).Error; err != nil {
		return nil, err
	}

	result := make([]*biz.SysDept, len(depts))
	for i := range depts {
		result[i] = r.toBiz(&depts[i])
	}
	return result, nil
}

func (r *sysDeptRepo) HasChildren(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.data.DB(ctx).Model(&model.SysDept{}).Where("parent_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *sysDeptRepo) HasUsers(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.data.DB(ctx).Model(&model.SysUser{}).Where("dept_id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *sysDeptRepo) toBiz(m *model.SysDept) *biz.SysDept {
	return &biz.SysDept{
		ID:         m.ID,
		ParentID:   m.ParentID,
		Name:       m.Name,
		Ancestors:  m.Ancestors,
		Sort:       m.Sort,
		TenantID:   m.TenantID,
		CreatedAt:  m.CreatedAt.Unix(),
		UpdatedAt:  m.UpdatedAt.Unix(),
	}
}
