package casbin

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz/provider"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
)

func Middleware(enforcer *casbin.SyncedEnforcer, provider *provider.PermissionProvider) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, _ := transport.FromServerContext(ctx)

			// 1. 获取该 API 关联的所有权限码
			// 结果可能是: ["user:manage", "order:assign", "audit:view"]
			permCodes := provider.GetCodes(tr.Operation())

			userID := auth.GetUserID(ctx)
			tenantID := auth.GetTenantID(ctx)

			// 2. 遍历校验：用户只要拥有其中【任何一个】权限码，即可访问该 API
			isAllowed := false
			finalScope := ""
			for _, code := range permCodes {
				ok, policy, _ := enforcer.EnforceEx(userID, tenantID, code, "V")
				if ok {
					isAllowed = true
					currentScope := policy[4]
					finalScope = auth.GetGreaterScope(finalScope, currentScope)
					break
				}
			}

			if !isAllowed {
				return nil, errors.Forbidden("CASBIN", "forbidden")
			}
			newCtx := auth.WithDataScope(ctx, finalScope)
			return handler(newCtx, req)
		}
	}
}
