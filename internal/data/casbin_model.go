package data

import (
	"github.com/casbin/casbin/v3/model"
)

// NewCasbinModel 使用 v3 编程式构建模型
func NewCasbinModel() (model.Model, error) {
	m := model.NewModel()

	// 1. 请求定义: sub(用户), dom(租户), obj(权限码), act(动作)
	m.AddDef("r", "r", "sub, dom, obj, act")

	// 2. 策略定义: sub(角色), dom(租户), obj(权限码), act(动作), scope(数据权限)
	m.AddDef("p", "p", "sub, dom, obj, act, scope")

	// 3. 角色定义: sub, role, dom (域 RBAC)
	m.AddDef("g", "g", "_, _, _")

	// 4. 策略效果: 只要有一条匹配成功就通过
	m.AddDef("e", "e", "some(where (p.eft == allow))")

	// 5. 匹配器: 角色匹配、租户匹配、权限码匹配、动作匹配
	//  [matchers]
	//	# 逻辑：
	//	# 1. 如果用户在 '1' 租户下拥有 'admin' 角色，直接放行 (忽略 r.dom 和 p.dom)
	//	#    '1' 租户在单租户模式下代表 'default' 租户，在多租户模式下代表 'system' 租户
	//	# 2. 否则，走正常的租户 RBAC 匹配
	m.AddDef("m", "m", "g(r.sub, 'admin', '1') || (g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act)")

	return m, nil
}
