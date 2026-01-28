package provider

import (
	"context"
	"fmt"
	"sync"
)

type PackageProvider struct {
	mux sync.RWMutex
	// Key: TenantID, Value: 该租户当前套餐拥有的所有 PermCodes
	tenantPerms map[int64][]string
	repo        PackageLoader
}

// PackageLoader 接口，由 Data 层实现
type PackageLoader interface {
	// LoadAllTenantPackagePerms 联表查询所有租户及其对应套餐的权限码
	// SQL 逻辑大概是:
	// SELECT t.id as tenant_id, p.perm_codes
	// FROM sys_tenant t
	// JOIN sys_package p ON t.package_id = p.id
	LoadAllTenantPackagePerms(ctx context.Context) (map[int64][]string, error)
}

func NewPackageProvider(repo PackageLoader) *PackageProvider {
	p := &PackageProvider{
		tenantPerms: make(map[int64][]string),
		repo:        repo,
	}
	if err := p.Load(context.Background()); err != nil {
		// 如果启动时加载失败，建议直接 panic 或 log.Fatal
		// 因为没有权限数据的鉴权系统是不安全的
		panic(fmt.Sprintf("failed to load permissions: %v", err))
	}
	return p
}

// Load 全量刷新内存映射
func (p *PackageProvider) Load(ctx context.Context) error {
	data, err := p.repo.LoadAllTenantPackagePerms(ctx)
	if err != nil {
		return err
	}
	p.mux.Lock()
	p.tenantPerms = data
	p.mux.Unlock()
	return nil
}

// IsTenantPermAllowed 校验特定租户是否有权使用该功能码
func (p *PackageProvider) IsTenantPermAllowed(tenantID int64, permCode string) bool {
	p.mux.RLock()
	defer p.mux.RUnlock()

	codes, ok := p.tenantPerms[tenantID]
	if !ok {
		return false
	}

	for _, c := range codes {
		if c == permCode {
			return true
		}
	}
	return false
}
