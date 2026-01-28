package provider

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// DynamicRoute 存储预编译的正则及其对应的权限码
type DynamicRoute struct {
	Raw     string         // 原始路径，用于排序
	Pattern *regexp.Regexp // 编译后的正则
	Codes   []string       // 关联的权限码
}

type PermissionProvider struct {
	mux           sync.RWMutex
	exactRoutes   map[string][]string // 精确匹配：/api/v1/user/list
	dynamicRoutes []DynamicRoute      // 动态匹配：/api/:id/*
	repo          PermissionLoader
}

// PermissionLoader 接口，由 Data 层实现
type PermissionLoader interface {
	LoadAllApiPermissions(ctx context.Context) (map[string][]string, error)
}

func NewPermissionProvider(repo PermissionLoader) *PermissionProvider {
	p := &PermissionProvider{
		exactRoutes:   make(map[string][]string),
		dynamicRoutes: make([]DynamicRoute, 0),
		repo:          repo,
	}
	// 初始化加载
	if err := p.Load(context.Background()); err != nil {
		// 如果启动时加载失败，建议直接 panic 或 log.Fatal
		// 因为没有权限数据的鉴权系统是不安全的
		panic(fmt.Sprintf("failed to load permissions: %v", err))
	}
	return p
}

// Load 从数据库重新加载并预编译 (支持手动调用)
func (p *PermissionProvider) Load(ctx context.Context) error {
	data, err := p.repo.LoadAllApiPermissions(ctx)
	if err != nil {
		return err
	}

	exact := make(map[string][]string)
	dynamic := make([]DynamicRoute, 0)

	for apiPath, codes := range data {
		// 判断是否包含通配符或占位符
		if strings.Contains(apiPath, "*") || strings.Contains(apiPath, ":") {
			// 1. 将路径转换为正则表达式
			regStr := p.pathToRegexp(apiPath)
			compiled, err := regexp.Compile(regStr)
			if err != nil {
				continue // 如果正则非法则跳过
			}
			dynamic = append(dynamic, DynamicRoute{
				Raw:     apiPath,
				Pattern: compiled,
				Codes:   codes,
			})
		} else {
			exact[apiPath] = codes
		}
	}

	// 2. 排序动态路由：长度越长的越优先匹配（越具体越靠前）
	sort.Slice(dynamic, func(i, j int) bool {
		return len(dynamic[i].Raw) > len(dynamic[j].Raw)
	})

	p.mux.Lock()
	p.exactRoutes = exact
	p.dynamicRoutes = dynamic
	p.mux.Unlock()
	return nil
}

// GetCodes 获取路径关联的权限码
func (p *PermissionProvider) GetCodes(apiPath string) []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	// 1. O(1) 精确匹配
	if codes, ok := p.exactRoutes[apiPath]; ok {
		return codes
	}

	// 2. 正则遍历匹配
	for _, route := range p.dynamicRoutes {
		if route.Pattern.MatchString(apiPath) {
			return route.Codes
		}
	}

	return nil
}

// pathToRegexp 将混有 :id 和 * 的路径转换为正则
func (p *PermissionProvider) pathToRegexp(path string) string {
	// 1. 先对特殊正则字符进行转义（防止路径中包含 . ? + 等字符）
	res := regexp.QuoteMeta(path)

	// 2. 将转义后的 :id 占位符替换为正则
	// 因为 QuoteMeta 会把 : 变成 \:
	reParam := regexp.MustCompile(`\\:[^/\\ ]+`)
	res = reParam.ReplaceAllString(res, `[^/]+`)

	// 3. 将转义后的 * 替换为正则
	// 因为 QuoteMeta 会把 * 变成 \*
	res = strings.ReplaceAll(res, `\*`, `.*`)

	return "^" + res + "$"
}
