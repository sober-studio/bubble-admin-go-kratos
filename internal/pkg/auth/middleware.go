package auth

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth/model"
)

// PathAccessConfig 路径访问配置
type PathAccessConfig struct {
	// 无需认证的路径
	PublicPaths map[string]struct{}
	// 认证后可访问的路径
	AuthPaths map[string]struct{}
}

// NewDefaultPathAccessConfig 创建默认路径访问配置
func NewDefaultPathAccessConfig() *PathAccessConfig {
	return &PathAccessConfig{
		PublicPaths: map[string]struct{}{
			"": {},
		},
		AuthPaths: map[string]struct{}{
			// 目前暂不判断，除公开接口列表中的路径外，均需要认证
		},
	}
}

// PathAccessConfigWithPublicList 创建路径访问配置
func PathAccessConfigWithPublicList(publicPaths []string) *PathAccessConfig {
	pathAccessConfig := &PathAccessConfig{
		PublicPaths: make(map[string]struct{}),
	}
	for _, path := range publicPaths {
		pathAccessConfig.PublicPaths[path] = struct{}{}
	}
	return pathAccessConfig
}

// IsPublicPath 判断是否为公开路径
func IsPublicPath(ctx context.Context, operation string, config *PathAccessConfig) bool {
	return Match(operation, config.PublicPaths)
}

// Match 判断路径是否匹配
func Match(operation string, paths map[string]struct{}) bool {
	_, ok := paths[operation]
	// 路径匹配
	if ok {
		return true
	}
	// 前缀匹配
	for path := range paths {
		if len(path) > 0 && path[len(path)-1] == '/' && len(operation) >= len(path) {
			if operation[:len(path)] == path {
				return true
			}
		}
	}
	return false
}

// JWTRecheck JWT 再次验证，从 TokenStore 中查询信息并验证
func JWTRecheck(tokenService TokenService) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 1. 验证 token 解析，这一步会访问缓存，确保 token 没有被吊销（注销登录/后台踢下线）
			_, err := tokenService.ParseTokenFromContext(ctx)
			if err != nil {
				return nil, err
			}
			// 2. 验证当前用户是否需要因修改密码、权限变更而需要重新登录
			// TODO：引入黑名单机制，确保 token 的签发时间在执行需要重新登录的操作时间之后
			// 验证通过，继续处理
			return handler(ctx, req)
		}
	}
}

func IdentityMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 1. 调用 Kratos JWT 标准库的方法获取 Claims (断言一次)
			claims, ok := jwt.FromContext(ctx)
			if !ok {
				return nil, errors.Unauthorized("UNAUTHORIZED", "无效的令牌")
			}

			// 2. 这里的 claims 是你定义的 MyClaims 结构体
			customClaims := claims.(*model.CustomClaims)
			userID, err := strconv.ParseInt(customClaims.Subject, 10, 64)
			if err != nil {
				return nil, errors.Unauthorized("UNAUTHORIZED", "无效的令牌")
			}

			// 3. 将这些分散的字段，通过我们 pkg/auth/context.go 的工具函数注入
			// 后面所有的中间件直接调用 auth.GetUserID(ctx) 即可，不再需要解析 JWT
			newCtx := NewContext(ctx, ContextInfo{
				UserID:   userID,
				TenantID: customClaims.TenantID,
				DeptID:   customClaims.DeptID,
			})

			return handler(newCtx, req)
		}
	}
}
