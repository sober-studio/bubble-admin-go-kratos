package service

import (
	"context"

	pb "github.com/sober-studio/bubble-admin-go-kratos/api/permission/v1"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
)

type PermissionService struct {
	pb.UnimplementedPermissionServer
	uc *biz.PermissionUseCase
}

func NewPermissionService(uc *biz.PermissionUseCase) *PermissionService {
	return &PermissionService{uc: uc}
}

func (s *PermissionService) GetUserMenu(ctx context.Context, req *pb.GetUserMenuRequest) (*pb.GetUserMenuReply, error) {
	// 从 JWT token 获取用户ID
	userID := auth.GetUserID(ctx)
	tenantID := auth.GetTenantID(ctx)

	// 获取用户菜单
	menus, err := s.uc.GetUserMenuByUserID(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserMenuReply{
		List: s.convertToMenuItems(menus),
	}, nil
}

func (s *PermissionService) Tree(ctx context.Context, req *pb.PermissionTreeRequest) (*pb.PermissionTreeReply, error) {
	list, err := s.uc.List(ctx)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	tree := s.uc.BuildTree(list)

	return &pb.PermissionTreeReply{
		List: s.convertToEntities(tree),
	}, nil
}

func (s *PermissionService) Get(ctx context.Context, req *pb.PermissionGetRequest) (*pb.PermissionGetReply, error) {
	perm, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.PermissionGetReply{
		Permission: s.convertToEntity(perm),
	}, nil
}

func (s *PermissionService) Create(ctx context.Context, req *pb.PermissionCreateRequest) (*pb.PermissionCreateReply, error) {
	tenantID := auth.GetTenantID(ctx)

	perm, err := s.uc.Create(ctx, req.ParentId, req.Name, req.Code, req.Type, req.ApiPath, req.ApiMethod, req.Sort)
	if err != nil {
		return nil, err
	}

	// 设置租户ID
	perm.TenantID = tenantID

	return &pb.PermissionCreateReply{
		Id: perm.ID,
	}, nil
}

func (s *PermissionService) Update(ctx context.Context, req *pb.PermissionUpdateRequest) (*pb.PermissionUpdateReply, error) {
	err := s.uc.Update(ctx, req.Id, req.ParentId, req.Name, req.Code, req.Type, req.ApiPath, req.ApiMethod, req.Sort)
	if err != nil {
		return nil, err
	}

	return &pb.PermissionUpdateReply{}, nil
}

func (s *PermissionService) Delete(ctx context.Context, req *pb.PermissionDeleteRequest) (*pb.PermissionDeleteReply, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.PermissionDeleteReply{}, nil
}

func (s *PermissionService) convertToEntity(perm *biz.SysPermission) *pb.PermissionEntity {
	return &pb.PermissionEntity{
		Id:         perm.ID,
		ParentId:   perm.ParentID,
		Name:       perm.Name,
		Code:       perm.Code,
		Type:       perm.Type,
		ApiPath:    perm.APIPath,
		ApiMethod:  perm.APIMethod,
		Sort:       perm.Sort,
		CreatedAt:  perm.CreatedAt,
		Children:   s.convertToEntities(perm.Children),
		Path:       perm.Path,
		Component:  perm.Component,
		Redirect:   perm.Redirect,
		Icon:       perm.Icon,
		OrderNo:    perm.OrderNo,
		Hidden:     perm.Hidden,
		KeepAlive:  perm.KeepAlive,
		FrameSrc:   perm.FrameSrc,
		FrameBlank: perm.FrameBlank,
	}
}

func (s *PermissionService) convertToEntities(perms []*biz.SysPermission) []*pb.PermissionEntity {
	result := make([]*pb.PermissionEntity, len(perms))
	for i := range perms {
		result[i] = s.convertToEntity(perms[i])
	}
	return result
}

func (s *PermissionService) convertToMenuItems(perms []*biz.SysPermission) []*pb.MenuItem {
	result := make([]*pb.MenuItem, len(perms))
	for i := range perms {
		result[i] = s.convertToMenuItem(perms[i])
	}
	return result
}

func (s *PermissionService) convertToMenuItem(perm *biz.SysPermission) *pb.MenuItem {
	return &pb.MenuItem{
		Path:      perm.Path,
		Name:      perm.Name,
		Component: perm.Component,
		Redirect:  perm.Redirect,
		Meta: &pb.MenuMeta{
			Title: map[string]string{
				"zh_CN": perm.Name,
				"en_US": "",
			},
			Icon:        perm.Icon,
			OrderNo:     perm.OrderNo,
			Hidden:      perm.Hidden,
			KeepAlive:   perm.KeepAlive,
			FrameSrc:    perm.FrameSrc,
			FrameBlank:  perm.FrameBlank,
		},
		Children: s.convertToMenuItems(perm.Children),
	}
}
