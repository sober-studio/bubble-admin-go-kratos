package biz

import (
	"context"
	"errors"
	"strconv"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

var (
	ErrDeptNotFound     = kerrors.NotFound("DEPT_NOT_FOUND", "部门不存在")
	ErrDeptAlreadyExist = kerrors.Conflict("DEPT_ALREADY_EXIST", "部门已存在")
	ErrDeptHasChildren  = kerrors.Conflict("DEPT_HAS_CHILDREN", "部门存在子部门，无法删除")
	ErrDeptHasUsers     = kerrors.Conflict("DEPT_HAS_USERS", "部门存在用户，无法删除")
)

// SysDept 部门领域模型
type SysDept struct {
	ID             int64
	ParentID       int64
	Name           string
	Ancestors      string
	Sort           int32
	LeaderUserID   int64
	LeaderUserName string
	Phone          string
	Email          string
	Status         int32
	TenantID       int64
	CreatedAt      int64
	UpdatedAt      int64
	Children       []*SysDept
}

// SysDeptRepo 部门仓储接口
type SysDeptRepo interface {
	Create(ctx context.Context, dept *SysDept) (*SysDept, error)
	GetByID(ctx context.Context, id int64) (*SysDept, error)
	Update(ctx context.Context, dept *SysDept) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*SysDept, error)
	HasChildren(ctx context.Context, id int64) (bool, error)
	HasUsers(ctx context.Context, id int64) (bool, error)
}

// DeptUseCase 部门业务逻辑
type DeptUseCase struct {
	repo SysDeptRepo
	log  *log.Helper
}

func NewDeptUseCase(repo SysDeptRepo, logger log.Logger) *DeptUseCase {
	return &DeptUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *DeptUseCase) Create(ctx context.Context, parentID int64, name string, sort int32, leaderUserID int64, leaderUserName string, phone string, email string, status int32) (*SysDept, error) {
	// 如果有父部门，检查父部门是否存在
	if parentID > 0 {
		parent, err := uc.repo.GetByID(ctx, parentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrDeptNotFound
			}
			return nil, err
		}
		// 构建 ancestors：父级的 ancestors + "," + 父级ID
		dept := &SysDept{
			ParentID:       parentID,
			Name:           name,
			Sort:           sort,
			Ancestors:      parent.Ancestors + "," + formatInt64(parentID),
			LeaderUserID:   leaderUserID,
			LeaderUserName: leaderUserName,
			Phone:          phone,
			Email:          email,
			Status:         status,
		}
		return uc.repo.Create(ctx, dept)
	}

	// 顶级部门
	dept := &SysDept{
		ParentID:       0,
		Name:           name,
		Sort:           sort,
		Ancestors:      "0",
		LeaderUserID:   leaderUserID,
		LeaderUserName: leaderUserName,
		Phone:          phone,
		Email:          email,
		Status:         status,
	}
	return uc.repo.Create(ctx, dept)
}

func (uc *DeptUseCase) GetByID(ctx context.Context, id int64) (*SysDept, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *DeptUseCase) Update(ctx context.Context, id int64, parentID int64, name string, sort int32, leaderUserID int64, leaderUserName string, phone string, email string, status int32) error {
	dept, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDeptNotFound
		}
		return err
	}

	// 如果父部门变了，需要更新 ancestors
	if dept.ParentID != parentID {
		var ancestors string
		if parentID == 0 {
			ancestors = "0"
		} else {
			parent, err := uc.repo.GetByID(ctx, parentID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrDeptNotFound
				}
				return err
			}
			ancestors = parent.Ancestors + "," + formatInt64(parentID)
		}

		dept.Ancestors = ancestors
		dept.ParentID = parentID
	}

	dept.Name = name
	dept.Sort = sort
	dept.LeaderUserID = leaderUserID
	dept.LeaderUserName = leaderUserName
	dept.Phone = phone
	dept.Email = email
	dept.Status = status

	return uc.repo.Update(ctx, dept)
}

func (uc *DeptUseCase) Delete(ctx context.Context, id int64) error {
	// 检查是否有子部门
	hasChildren, err := uc.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return ErrDeptHasChildren
	}

	// 检查是否有用户
	hasUsers, err := uc.repo.HasUsers(ctx, id)
	if err != nil {
		return err
	}
	if hasUsers {
		return ErrDeptHasUsers
	}

	return uc.repo.Delete(ctx, id)
}

func (uc *DeptUseCase) List(ctx context.Context) ([]*SysDept, error) {
	return uc.repo.List(ctx)
}

// BuildTree 构建部门树
func (uc *DeptUseCase) BuildTree(depts []*SysDept) []*SysDept {
	// 构建 ID 到部门的映射
	deptMap := make(map[int64]*SysDept)
	for i := range depts {
		deptMap[depts[i].ID] = depts[i]
	}

	var roots []*SysDept
	// 构建树结构
	for _, dept := range depts {
		if dept.ParentID == 0 {
			// 根部门
			roots = append(roots, dept)
		} else {
			// 找到父部门，添加到子部门列表
			if parent, ok := deptMap[dept.ParentID]; ok {
				parent.Children = append(parent.Children, dept)
			}
		}
	}

	return roots
}

func formatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}
