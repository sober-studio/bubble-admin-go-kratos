package service

import (
	"context"

	pb "github.com/sober-studio/bubble-admin-go-kratos/api/dept/v1"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
)

type DeptService struct {
	pb.UnimplementedDeptServer
	uc *biz.DeptUseCase
}

func NewDeptService(uc *biz.DeptUseCase) *DeptService {
	return &DeptService{uc: uc}
}

func (s *DeptService) List(ctx context.Context, req *pb.DeptListRequest) (*pb.DeptListReply, error) {
	list, err := s.uc.List(ctx)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	tree := s.uc.BuildTree(list)

	return &pb.DeptListReply{
		List: s.convertToEntities(tree),
	}, nil
}

func (s *DeptService) Get(ctx context.Context, req *pb.DeptGetRequest) (*pb.DeptGetReply, error) {
	dept, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeptGetReply{
		Dept: s.convertToEntity(dept),
	}, nil
}

func (s *DeptService) Create(ctx context.Context, req *pb.DeptCreateRequest) (*pb.DeptCreateReply, error) {
	tenantID := auth.GetTenantID(ctx)

	dept, err := s.uc.Create(ctx, req.ParentId, req.Name, req.Sort)
	if err != nil {
		return nil, err
	}

	// 设置租户ID（如果是新创建的）
	if dept.TenantID == 0 {
		dept.TenantID = tenantID
	}

	return &pb.DeptCreateReply{
		Id: dept.ID,
	}, nil
}

func (s *DeptService) Update(ctx context.Context, req *pb.DeptUpdateRequest) (*pb.DeptUpdateReply, error) {
	err := s.uc.Update(ctx, req.Id, req.ParentId, req.Name, req.Sort)
	if err != nil {
		return nil, err
	}

	return &pb.DeptUpdateReply{}, nil
}

func (s *DeptService) Delete(ctx context.Context, req *pb.DeptDeleteRequest) (*pb.DeptDeleteReply, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.DeptDeleteReply{}, nil
}

func (s *DeptService) convertToEntity(dept *biz.SysDept) *pb.DeptEntity {
	return &pb.DeptEntity{
		Id:         dept.ID,
		ParentId:   dept.ParentID,
		Name:       dept.Name,
		Ancestors:  dept.Ancestors,
		Sort:       dept.Sort,
		CreatedAt:  dept.CreatedAt,
		Children:   s.convertToEntities(dept.Children),
	}
}

func (s *DeptService) convertToEntities(depts []*biz.SysDept) []*pb.DeptEntity {
	result := make([]*pb.DeptEntity, len(depts))
	for i := range depts {
		result[i] = s.convertToEntity(depts[i])
	}
	return result
}
