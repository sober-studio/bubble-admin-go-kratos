package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http/binding"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	passportV1 "github.com/sober-studio/bubble-admin-go-kratos/api/passport/v1"
	publicV1 "github.com/sober-studio/bubble-admin-go-kratos/api/public/v1"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/biz/provider"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/conf"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/auth/model"
	pkgCasbin "github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/casbin"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/debug"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/pkg/render"
	"github.com/sober-studio/bubble-admin-go-kratos/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	app *conf.App,
	public *service.PublicService,
	passport *service.PassportService,
	tokenService auth.TokenService,
	wsSvc *service.WebsocketService,
	enforcer *casbin.SyncedEnforcer,
	permissionProvider *provider.PermissionProvider,
	packageProvider *provider.PackageProvider,
	logger log.Logger,
) *http.Server {

	pathConfig := auth.PathAccessConfigWithPublicList(app.Auth.PublicPaths)

	var opts = []http.ServerOption{
		// 中间件配置
		http.Middleware(
			recovery.Recovery(),
			selector.Server(
				// 1. JWT 认证中间件
				jwt.Server(
					func(token *jwtv5.Token) (interface{}, error) {
						return tokenService.GetSecretKey(), nil
					},
					jwt.WithSigningMethod(jwtv5.SigningMethodHS256),
					jwt.WithClaims(func() jwtv5.Claims {
						return &model.CustomClaims{}
					}),
				),
				// 2. JWT 再次验证，从 TokenStore 中查询信息并验证，
				//    确保 token 没有被吊销（注销登录/后台踢下线），
				//    且没有因修改密码、权限变更而需要重新登录（开发中）
				auth.JWTRecheck(tokenService),
				// 3. 身份桥接中间件
				auth.IdentityMiddleware(),
				// 4. 租户上下文
				func(handler middleware.Handler) middleware.Handler {
					return func(ctx context.Context, req interface{}) (interface{}, error) {
						// 默认租户：单租户模式为 default 租户，多租户模式为 system 租户
						defaultTenantID := int64(1)
						// systemTenantID := int64(1)
						// 单租户模式：强制注入硬编码的租户 ID
						if !app.EnableMultiTenant {
							ctx = auth.WithTenantID(ctx, defaultTenantID)
							return handler(ctx, req)
						}
						return handler(ctx, req)
					}
				},
				// 5. 租户套餐权限验证（只在多租户模式启用）
				selector.Server(func(handler middleware.Handler) middleware.Handler {
					return func(ctx context.Context, req interface{}) (interface{}, error) {
						tr, ok := transport.FromServerContext(ctx)
						if !ok {
							return handler(ctx, req)
						}
						apiPath := tr.Operation()

						// 1. 获取当前 API 关联的功能码 (支持 :id 和 *)
						permCodes := permissionProvider.GetCodes(apiPath)
						if len(permCodes) == 0 {
							return nil, errors.Forbidden("FORBIDDEN", "接口权限未定义")
						}

						// 2. 获取租户 ID (从之前的 TenantContext 中间件或 JWT 中间件中获取)
						// 注意：顺序必须在身份识别中间件之后
						tenantID := auth.GetTenantID(ctx)

						// 3. 校验租户套餐边界
						allowed := false
						for _, code := range permCodes {
							if packageProvider.IsTenantPermAllowed(tenantID, code) {
								allowed = true
								break
							}
						}

						if !allowed {
							return nil, errors.Forbidden("PACKAGE_LIMIT", "您的租户套餐暂不支持此功能")
						}

						return handler(ctx, req)
					}
				}).Match(func(ctx context.Context, operation string) bool {
					return app.EnableMultiTenant
				}).Build(),
				// 6. 权限校验
				pkgCasbin.Middleware(enforcer, permissionProvider),
			).Match(func(ctx context.Context, operation string) bool {
				return !auth.IsPublicPath(ctx, operation, pathConfig)
			}).Build(),
		),
		http.Filter(debug.Filter),
		http.RequestDecoder(MultipartRequestDecoder),
		http.ResponseEncoder(render.ResponseEncoder),
		http.ErrorEncoder(render.ErrorEncoder),
	}

	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)

	// 同端口集成点：手动绑定路由
	// 注意：这里用 Handlers.HandleFunc 是绕过 Kratos 的 Proto 解析，直接处理原始 HTTP 请求
	srv.HandleFunc("/ws", wsSvc.WSHandler)

	passportV1.RegisterPassportHTTPServer(srv, passport)
	publicV1.RegisterPublicHTTPServer(srv, public)

	return srv
}

// MultipartRequestDecoder 识别 multipart/form-data 并解析非文件字段
func MultipartRequestDecoder(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")

	// 如果是文件上传
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// 1. 解析 multipart 表单
		// 这一步执行后，非文件字段会被自动填充到 r.Form 中
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return err
		}

		// debug
		fmt.Printf("Form Values: %+v\n", r.Form)

		// 2. 直接传入 r
		// Kratos 会自动从 r.Form 中提取数据并匹配到结构体 v
		if err := binding.BindForm(r, v); err != nil {
			return err
		}
		return nil
	}

	// 如果是普通的 JSON 请求，走默认的解码逻辑
	return http.DefaultRequestDecoder(r, v)
}
