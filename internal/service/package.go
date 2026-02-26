package service

import (
	"context"

	pb "github.com/sober-studio/bubble-admin-go-kratos/api/package/v1"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/datascope"
)

type PackageService struct {
	pb.UnimplementedPackageServer
	uc *biz.PackageUseCase
}

func NewPackageService(uc *biz.PackageUseCase) *PackageService {
	return &PackageService{uc: uc}
}

func (s *PackageService) List(ctx context.Context, req *pb.PackageListRequest) (*pb.PackageListReply, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// 套餐不属手租户，需要禁用租户隔离
	ctx = datascope.WithConfig(ctx, datascope.Config{DisableTenant: true})
	list, total, err := s.uc.List(ctx, req.Name, req.Status, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &pb.PackageListReply{
		List: s.convertToEntities(list),
		Total: total,
	}, nil
}

func (s *PackageService) Get(ctx context.Context, req *pb.PackageGetRequest) (*pb.PackageGetReply, error) {
	// 套餐不属手租户，需要禁用租户隔离
	ctx = datascope.WithConfig(ctx, datascope.Config{DisableTenant: true})
	pkg, permIDs, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.PackageGetReply{
		Pkg:           s.convertToEntity(pkg),
		PermissionIds: permIDs,
	}, nil
}

func (s *PackageService) Create(ctx context.Context, req *pb.PackageCreateRequest) (*pb.PackageCreateReply, error) {
	tenantID := auth.GetTenantID(ctx)

	pkg, err := s.uc.Create(ctx, req.Name, req.Code, req.Remark)
	if err != nil {
		return nil, err
	}

	// 设置租户ID
	pkg.TenantID = tenantID

	return &pb.PackageCreateReply{
		Id: pkg.ID,
	}, nil
}

func (s *PackageService) Update(ctx context.Context, req *pb.PackageUpdateRequest) (*pb.PackageUpdateReply, error) {
	err := s.uc.Update(ctx, req.Id, req.Name, req.Code, req.Remark)
	if err != nil {
		return nil, err
	}

	return &pb.PackageUpdateReply{}, nil
}

func (s *PackageService) Delete(ctx context.Context, req *pb.PackageDeleteRequest) (*pb.PackageDeleteReply, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.PackageDeleteReply{}, nil
}

func (s *PackageService) AssignPermissions(ctx context.Context, req *pb.PackageAssignPermissionsRequest) (*pb.PackageAssignPermissionsReply, error) {
	err := s.uc.AssignPermissions(ctx, req.Id, req.PermissionIds)
	if err != nil {
		return nil, err
	}

	return &pb.PackageAssignPermissionsReply{}, nil
}

func (s *PackageService) convertToEntity(pkg *biz.SysPackage) *pb.PackageEntity {
	return &pb.PackageEntity{
		Id:        pkg.ID,
		Name:      pkg.Name,
		Code:      pkg.Code,
		Remark:    pkg.Remark,
		Status:    pkg.Status,
		CreatedAt: pkg.CreatedAt,
	}
}

func (s *PackageService) convertToEntities(pkgs []*biz.SysPackage) []*pb.PackageEntity {
	result := make([]*pb.PackageEntity, len(pkgs))
	for i := range pkgs {
		result[i] = s.convertToEntity(pkgs[i])
	}
	return result
}
