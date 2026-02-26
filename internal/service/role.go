package service

import (
	"context"

	pb "github.com/sober-studio/bubble-admin-go-kratos/api/role/v1"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
)

type RoleService struct {
	pb.UnimplementedRoleServer
	uc *biz.RoleUseCase
}

func NewRoleService(uc *biz.RoleUseCase) *RoleService {
	return &RoleService{uc: uc}
}

func (s *RoleService) List(ctx context.Context, req *pb.RoleListRequest) (*pb.RoleListReply, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := s.uc.List(ctx, req.Name, req.Code, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &pb.RoleListReply{
		List: s.convertToEntities(list),
		Total: total,
	}, nil
}

func (s *RoleService) Get(ctx context.Context, req *pb.RoleGetRequest) (*pb.RoleGetReply, error) {
	role, permIDs, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.RoleGetReply{
		Role:          s.convertToEntity(role),
		PermissionIds: permIDs,
	}, nil
}

func (s *RoleService) Create(ctx context.Context, req *pb.RoleCreateRequest) (*pb.RoleCreateReply, error) {
	tenantID := auth.GetTenantID(ctx)

	role, err := s.uc.Create(ctx, req.Name, req.Code)
	if err != nil {
		return nil, err
	}

	// 设置租户ID
	role.TenantID = tenantID

	return &pb.RoleCreateReply{
		Id: role.ID,
	}, nil
}

func (s *RoleService) Update(ctx context.Context, req *pb.RoleUpdateRequest) (*pb.RoleUpdateReply, error) {
	err := s.uc.Update(ctx, req.Id, req.Name, req.Code)
	if err != nil {
		return nil, err
	}

	return &pb.RoleUpdateReply{}, nil
}

func (s *RoleService) Delete(ctx context.Context, req *pb.RoleDeleteRequest) (*pb.RoleDeleteReply, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.RoleDeleteReply{}, nil
}

func (s *RoleService) AssignPermissions(ctx context.Context, req *pb.RoleAssignPermissionsRequest) (*pb.RoleAssignPermissionsReply, error) {
	err := s.uc.AssignPermissions(ctx, req.Id, req.PermissionIds, req.DataScope, req.DeptIds)
	if err != nil {
		return nil, err
	}

	return &pb.RoleAssignPermissionsReply{}, nil
}

func (s *RoleService) convertToEntity(role *biz.SysRole) *pb.RoleEntity {
	return &pb.RoleEntity{
		Id:        role.ID,
		Name:      role.Name,
		Code:      role.Code,
		CreatedAt: role.CreatedAt,
	}
}

func (s *RoleService) convertToEntities(roles []*biz.SysRole) []*pb.RoleEntity {
	result := make([]*pb.RoleEntity, len(roles))
	for i := range roles {
		result[i] = s.convertToEntity(roles[i])
	}
	return result
}
