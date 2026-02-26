package casbin

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz/provider"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
)

func Middleware(enforcer *casbin.SyncedEnforcer, provider *provider.PermissionProvider) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, _ := transport.FromServerContext(ctx)
			helper := log.NewHelper(log.With(log.GetLogger(), "module", "casbin-middleware"))

			// 1. 获取该 API 关联的所有权限码
			// 结果可能是: ["user:manage", "order:assign", "audit:view"]
			permCodes := provider.GetCodes(tr.Operation())

			userID := auth.GetUserID(ctx)
			tenantID := auth.GetTenantID(ctx)

			// 添加调试日志
			helper.Infof("userID=%d, tenantID=%d, permCodes=%v", userID, tenantID, permCodes)

			// 查看所有角色
			roles, _ := enforcer.GetRolesForUser(fmt.Sprintf("%d", userID), fmt.Sprintf("%d", tenantID))
			helper.Infof("All groups for user %d: %v", userID, roles)

			// 2. 遍历校验：用户只要拥有其中【任何一个】权限码，即可访问该 API
			isAllowed := false
			finalScope := ""
			for _, code := range permCodes {
				ok, policy, _ := enforcer.EnforceEx(fmt.Sprintf("%d", userID), fmt.Sprintf("%d", tenantID), code, "V")
				helper.Infof("EnforceEx(userID=%d, tenantID=%d, code=%s, act=V) = ok=%v, policy=%v", userID, tenantID, code, ok, policy)
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
