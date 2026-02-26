package service

import (
	"context"

	pb "github.com/sober-studio/bubble-admin-go-kratos/api/user/v1"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	pb.UnimplementedUserServer
	uc *biz.UserUseCase
}

func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{uc: uc}
}

func (s *UserService) List(ctx context.Context, req *pb.UserListRequest) (*pb.UserListReply, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	list, total, err := s.uc.List(ctx, req.Username, req.Mobile, req.Status, req.DeptId, page, pageSize)
	if err != nil {
		return nil, err
	}

	return &pb.UserListReply{
		List: s.convertToEntities(list),
		Total: total,
	}, nil
}

func (s *UserService) Get(ctx context.Context, req *pb.UserGetRequest) (*pb.UserGetReply, error) {
	user, roleIDs, err := s.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.UserGetReply{
		User:     s.convertToEntity(user),
		RoleIds: roleIDs,
	}, nil
}

func (s *UserService) Create(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserCreateReply, error) {
	tenantID := auth.GetTenantID(ctx)

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.uc.Create(ctx, req.Username, string(hashedPassword), req.Nickname, req.Mobile, req.DeptId, req.RoleIds)
	if err != nil {
		return nil, err
	}

	// 设置租户ID
	user.TenantID = tenantID

	return &pb.UserCreateReply{
		Id: user.ID,
	}, nil
}

func (s *UserService) Update(ctx context.Context, req *pb.UserUpdateRequest) (*pb.UserUpdateReply, error) {
	err := s.uc.Update(ctx, req.Id, req.Nickname, req.Mobile, req.DeptId)
	if err != nil {
		return nil, err
	}

	return &pb.UserUpdateReply{}, nil
}

func (s *UserService) Delete(ctx context.Context, req *pb.UserDeleteRequest) (*pb.UserDeleteReply, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &pb.UserDeleteReply{}, nil
}

func (s *UserService) AssignRoles(ctx context.Context, req *pb.UserAssignRolesRequest) (*pb.UserAssignRolesReply, error) {
	err := s.uc.AssignRoles(ctx, req.Id, req.RoleIds)
	if err != nil {
		return nil, err
	}

	return &pb.UserAssignRolesReply{}, nil
}

func (s *UserService) SetStatus(ctx context.Context, req *pb.UserSetStatusRequest) (*pb.UserSetStatusReply, error) {
	err := s.uc.SetStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}

	return &pb.UserSetStatusReply{}, nil
}

func (s *UserService) ResetPassword(ctx context.Context, req *pb.UserResetPasswordRequest) (*pb.UserResetPasswordReply, error) {
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	err = s.uc.ResetPassword(ctx, req.Id, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return &pb.UserResetPasswordReply{}, nil
}

func (s *UserService) convertToEntity(user *biz.SysUser) *pb.UserEntity {
	return &pb.UserEntity{
		Id:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Mobile:    user.Phone,
		DeptId:    user.DeptID,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
}

func (s *UserService) convertToEntities(users []*biz.SysUser) []*pb.UserEntity {
	result := make([]*pb.UserEntity, len(users))
	for i := range users {
		result[i] = s.convertToEntity(users[i])
	}
	return result
}
