-- =========================================================
-- 权限数据初始化脚本
-- 执行前请确保数据库已创建表结构
-- =========================================================

-- 1. 添加新字段（如果表结构没有这些字段）
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS path VARCHAR(255);
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS component VARCHAR(255);
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS redirect VARCHAR(255);
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS icon VARCHAR(64);
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS order_no INT DEFAULT 0;
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS hidden BOOLEAN DEFAULT FALSE;
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS keep_alive BOOLEAN DEFAULT FALSE;
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS frame_src VARCHAR(255);
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS frame_blank BOOLEAN DEFAULT FALSE;
ALTER TABLE sys_permission ADD COLUMN IF NOT EXISTS tenant_id BIGINT DEFAULT 1;

-- 2. 插入菜单权限
INSERT INTO sys_permission (id, tenant_id, parent_id, name, code, type, path, component, redirect, icon, sort, order_no, created_at, updated_at) VALUES
(1, 1, 0, '系统管理', 'system:common', 'MENU', '/system', 'LAYOUT', '/system/user', 'setting', 1, 1, NOW(), NOW()),
(2, 1, 1, '用户管理', 'system:user:common', 'MENU', '/system/user', '/system/user/index', '', 'user', 1, 1, NOW(), NOW()),
(3, 1, 1, '角色管理', 'system:role:common', 'MENU', '/system/role', '/system/role/index', '', 'role', 2, 2, NOW(), NOW()),
(4, 1, 1, '权限管理', 'system:permission:common', 'MENU', '/system/permission', '/system/permission/index', '', 'lock', 3, 3, NOW(), NOW()),
(5, 1, 1, '部门管理', 'system:dept:common', 'MENU', '/system/dept', '/system/dept/index', '', 'department', 4, 4, NOW(), NOW()),
(6, 1, 1, '租户管理', 'system:tenant:common', 'MENU', '/system/tenant', '/system/tenant/index', '', 'tenant', 5, 5, NOW(), NOW()),
(7, 1, 1, '套餐管理', 'system:package:common', 'MENU', '/system/package', '/system/package/index', '', 'package', 6, 6, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 3. 插入 API 权限
INSERT INTO sys_permission (id, tenant_id, parent_id, name, code, type, api_path, api_method, sort, created_at, updated_at) VALUES
-- 菜单列表 API
(101, 1, 1, '获取用户菜单', 'system:menu:list', 'API', '/api/v1/menu-list', 'GET', 1, NOW(), NOW()),
-- 用户管理 API
(102, 1, 2, '用户列表', 'system:user:list', 'API', '/api/v1/user/list', 'GET', 1, NOW(), NOW()),
(103, 1, 2, '用户详情', 'system:user:get', 'API', '/api/v1/user/:id', 'GET', 2, NOW(), NOW()),
(104, 1, 2, '创建用户', 'system:user:create', 'API', '/api/v1/user', 'POST', 3, NOW(), NOW()),
(105, 1, 2, '更新用户', 'system:user:update', 'API', '/api/v1/user/:id', 'PUT', 4, NOW(), NOW()),
(106, 1, 2, '删除用户', 'system:user:delete', 'API', '/api/v1/user/:id', 'DELETE', 5, NOW(), NOW()),
(107, 1, 2, '分配用户角色', 'system:user:assignRoles', 'API', '/api/v1/user/:id/roles', 'PUT', 6, NOW(), NOW()),
(108, 1, 2, '设置用户状态', 'system:user:setStatus', 'API', '/api/v1/user/:id/status', 'PUT', 7, NOW(), NOW()),
(109, 1, 2, '重置密码', 'system:user:resetPassword', 'API', '/api/v1/user/:id/reset-password', 'PUT', 8, NOW(), NOW()),
-- 角色管理 API
(201, 1, 3, '角色列表', 'system:role:list', 'API', '/api/v1/role/list', 'GET', 1, NOW(), NOW()),
(202, 1, 3, '角色详情', 'system:role:get', 'API', '/api/v1/role/:id', 'GET', 2, NOW(), NOW()),
(203, 1, 3, '创建角色', 'system:role:create', 'API', '/api/v1/role', 'POST', 3, NOW(), NOW()),
(204, 1, 3, '更新角色', 'system:role:update', 'API', '/api/v1/role/:id', 'PUT', 4, NOW(), NOW()),
(205, 1, 3, '删除角色', 'system:role:delete', 'API', '/api/v1/role/:id', 'DELETE', 5, NOW(), NOW()),
(206, 1, 3, '分配角色权限', 'system:role:assignPermissions', 'API', '/api/v1/role/:id/permissions', 'PUT', 6, NOW(), NOW()),
-- 权限管理 API
(301, 1, 4, '权限树', 'system:permission:tree', 'API', '/api/v1/permission/tree', 'GET', 1, NOW(), NOW()),
(302, 1, 4, '权限详情', 'system:permission:get', 'API', '/api/v1/permission/:id', 'GET', 2, NOW(), NOW()),
(303, 1, 4, '创建权限', 'system:permission:create', 'API', '/api/v1/permission', 'POST', 3, NOW(), NOW()),
(304, 1, 4, '更新权限', 'system:permission:update', 'API', '/api/v1/permission/:id', 'PUT', 4, NOW(), NOW()),
(305, 1, 4, '删除权限', 'system:permission:delete', 'API', '/api/v1/permission/:id', 'DELETE', 5, NOW(), NOW()),
-- 部门管理 API
(401, 1, 5, '部门列表', 'system:dept:list', 'API', '/api/v1/dept/list', 'GET', 1, NOW(), NOW()),
(402, 1, 5, '部门详情', 'system:dept:get', 'API', '/api/v1/dept/:id', 'GET', 2, NOW(), NOW()),
(403, 1, 5, '创建部门', 'system:dept:create', 'API', '/api/v1/dept', 'POST', 3, NOW(), NOW()),
(404, 1, 5, '更新部门', 'system:dept:update', 'API', '/api/v1/dept/:id', 'PUT', 4, NOW(), NOW()),
(405, 1, 5, '删除部门', 'system:dept:delete', 'API', '/api/v1/dept/:id', 'DELETE', 5, NOW(), NOW()),
-- 租户管理 API
(501, 1, 6, '租户列表', 'system:tenant:list', 'API', '/api/v1/tenant/list', 'GET', 1, NOW(), NOW()),
(502, 1, 6, '租户详情', 'system:tenant:get', 'API', '/api/v1/tenant/:id', 'GET', 2, NOW(), NOW()),
(503, 1, 6, '创建租户', 'system:tenant:create', 'API', '/api/v1/tenant', 'POST', 3, NOW(), NOW()),
(504, 1, 6, '更新租户', 'system:tenant:update', 'API', '/api/v1/tenant/:id', 'PUT', 4, NOW(), NOW()),
(505, 1, 6, '删除租户', 'system:tenant:delete', 'API', '/api/v1/tenant/:id', 'DELETE', 5, NOW(), NOW()),
-- 套餐管理 API
(601, 1, 7, '套餐列表', 'system:package:list', 'API', '/api/v1/package/list', 'GET', 1, NOW(), NOW()),
(602, 1, 7, '套餐详情', 'system:package:get', 'API', '/api/v1/package/:id', 'GET', 2, NOW(), NOW()),
(603, 1, 7, '创建套餐', 'system:package:create', 'API', '/api/v1/package', 'POST', 3, NOW(), NOW()),
(604, 1, 7, '更新套餐', 'system:package:update', 'API', '/api/v1/package/:id', 'PUT', 4, NOW(), NOW()),
(605, 1, 7, '删除套餐', 'system:package:delete', 'API', '/api/v1/package/:id', 'DELETE', 5, NOW(), NOW()),
(606, 1, 7, '分配套餐权限', 'system:package:assignPermissions', 'API', '/api/v1/package/:id/permissions', 'PUT', 6, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 4. 绑定角色权限（admin 角色拥有所有权限）
INSERT INTO sys_role_permission (id, tenant_id, role_id, permission_id, data_scope, dept_ids, created_at)
SELECT id, tenant_id, role_id, permission_id, 1, '', NOW()
FROM (
    SELECT 1 as id, 1 as tenant_id, 1 as role_id, permission_id
    FROM sys_permission
    WHERE tenant_id = 1
) t
ON CONFLICT (id) DO NOTHING;

-- 5. 验证数据
SELECT '菜单权限' as type, COUNT(*) as count FROM sys_permission WHERE type = 'MENU';
SELECT 'API权限' as type, COUNT(*) as count FROM sys_permission WHERE type = 'API';
SELECT '角色权限绑定' as type, COUNT(*) as count FROM sys_role_permission;
