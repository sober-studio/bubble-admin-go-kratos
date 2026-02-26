package service

import (
	"context"

	pb "github.com/sober-studio/bubble-admin-go-kratos/api/tenant/v1"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
)

type TenantService struct {
	pb.UnimplementedTenantServer
	uc *biz.TenantUseCase
}

func NewTenantService(uc *biz.TenantUseCase) *TenantService {
	return &TenantService{uc: uc}
}

func (s *TenantService) List(ctx context.Context, req *pb.TenantListRequest) (*pb.TenantListReply, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := s.uc.List(ctx, req.Name, req.Code, req.Status, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &pb.TenantListReply{
		List: s.convertToEntities(list),
		Total: total,
	}, nil
}

func (s *TenantService) Get(ctx context.Context, req *pb.TenantGetRequest) (*pb.TenantGetReply, error) {
	tenant, packageID, packageName, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.TenantGetReply{
		Tenant:      s.convertToEntity(tenant),
		PackageId:   packageID,
		PackageName: packageName,
	}, nil
}

func (s *TenantService) Create(ctx context.Context, req *pb.TenantCreateRequest) (*pb.TenantCreateReply, error) {
	tenant, err := s.uc.Create(ctx, req.Name, req.Code, req.PackageId, req.ExpiredAt)
	if err != nil {
		return nil, err
	}

	return &pb.TenantCreateReply{
		Id: tenant.ID,
	}, nil
}

func (s *TenantService) Update(ctx context.Context, req *pb.TenantUpdateRequest) (*pb.TenantUpdateReply, error) {
	err := s.uc.Update(ctx, req.Id, req.Name, req.Code, req.PackageId, req.ExpiredAt)
	if err != nil {
		return nil, err
	}

	return &pb.TenantUpdateReply{}, nil
}

func (s *TenantService) Delete(ctx context.Context, req *pb.TenantDeleteRequest) (*pb.TenantDeleteReply, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.TenantDeleteReply{}, nil
}

func (s *TenantService) convertToEntity(tenant *biz.SysTenant) *pb.TenantEntity {
	return &pb.TenantEntity{
		Id:        tenant.ID,
		Name:      tenant.Name,
		Code:      tenant.Code,
		Status:    tenant.Status,
		ExpiredAt: tenant.ExpiredAt,
		CreatedAt: tenant.CreatedAt,
	}
}

func (s *TenantService) convertToEntities(tenants []*biz.SysTenant) []*pb.TenantEntity {
	result := make([]*pb.TenantEntity, len(tenants))
	for i := range tenants {
		result[i] = s.convertToEntity(tenants[i])
	}
	return result
}
