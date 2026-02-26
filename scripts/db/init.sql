-- =========================================================
-- [GLOBAL]
-- 数据库类型: PostgreSQL 15
-- 字符集: UTF8
-- 排序规则: 使用数据库实例默认(例如: zh_CN.UTF-8 或 en_US.UTF-8)
--
-- 表命名规范:
--   - 表名前缀：sys_
--   - 使用下划线风格, 如: sys_user
--
-- 主键规范:
--   - 字段名: id
--   - 类型: BIGINT
--
-- 时间字段规范:
--   - created_at: 创建时间(由 GORM 自动维护，无需触发器)
--   - updated_at: 修改时间(由 GORM 自动维护，无需触发器)
--   - deleted_at: 删除时间(由 GORM 自动维护，逻辑删除)
-- =========================================================

-- =========================================================
-- [COMMON_COLUMNS]
-- 公共字段
-- id BIGINT PRIMARY KEY, -- 主键 ID（雪花算法）
-- created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 创建时间
-- updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, -- 更新时间
-- deleted_at TIMESTAMP WITH TIME ZONE -- 删除时间
-- =========================================================

-- =========================================================
-- 1. 租户套餐表 (sys_package)
-- =========================================================
CREATE TABLE sys_package (
    id BIGINT PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    status SMALLINT DEFAULT 1,      -- 状态 (1:正常, 2:禁用)
    remark VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
COMMENT ON TABLE sys_package IS '租户套餐表';

-- =========================================================
-- 2. 租户表 (sys_tenant)
-- =========================================================
CREATE TABLE sys_tenant (
    id BIGINT PRIMARY KEY,
    code VARCHAR(64) NOT NULL,      -- 租户编码
    name VARCHAR(128) NOT NULL,     -- 租户名称
    package_id BIGINT,              -- 关联套餐 ID
    expire_time TIMESTAMP WITH TIME ZONE, -- 套餐过期时间
    status SMALLINT DEFAULT 1,      -- 状态 (1:正常, 2:禁用)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_tenant_code ON sys_tenant(code) WHERE deleted_at IS NULL;
COMMENT ON TABLE sys_tenant IS '租户表';

-- =========================================================
-- 3. 权限/菜单表 (sys_permission)
-- =========================================================
CREATE TABLE sys_permission (
    id BIGINT PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,      -- 权限码 (如: user:list)
    type VARCHAR(20) NOT NULL,      -- MENU, BUTTON, API
    api_path VARCHAR(255),          -- Kratos Operation (如: /api.user.v1.User/ListUsers)
    api_method VARCHAR(20) DEFAULT 'V', -- 默认为 V (用于 Casbin Action)
    sort INT DEFAULT 0,
    -- 前端菜单字段
    path VARCHAR(255),              -- 路由路径
    component VARCHAR(255),         -- 组件路径
    redirect VARCHAR(255),          -- 重定向路径
    icon VARCHAR(64),               -- 图标
    order_no INT DEFAULT 0,         -- 排序号
    hidden BOOLEAN DEFAULT FALSE,   -- 是否隐藏
    keep_alive BOOLEAN DEFAULT FALSE, -- 是否缓存
    frame_src VARCHAR(255),        -- IFrame 地址
    frame_blank BOOLEAN DEFAULT FALSE, -- 是否新窗口
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    tenant_id BIGINT DEFAULT 1
);
CREATE UNIQUE INDEX uk_perm_code ON sys_permission(code) WHERE deleted_at IS NULL;
COMMENT ON COLUMN sys_permission.api_path IS 'Kratos内部路径/API路径';
COMMENT ON COLUMN sys_permission.path IS '路由路径';
COMMENT ON COLUMN sys_permission.component IS '组件路径';
COMMENT ON COLUMN sys_permission.redirect IS '重定向路径';
COMMENT ON COLUMN sys_permission.icon IS '图标';
COMMENT ON COLUMN sys_permission.order_no IS '排序号';
COMMENT ON COLUMN sys_permission.hidden IS '是否隐藏';
COMMENT ON COLUMN sys_permission.keep_alive IS '是否缓存';
COMMENT ON COLUMN sys_permission.frame_src IS 'IFrame地址';
COMMENT ON COLUMN sys_permission.frame_blank IS '是否新窗口';

-- =========================================================
-- 4. 套餐权限关联表 (sys_package_permission)
-- =========================================================
CREATE TABLE sys_package_permission (
    id BIGINT PRIMARY KEY,
    package_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE sys_package_permission IS '套餐与权限码关联表(定义套餐功能边界)';

-- =========================================================
-- 5. 部门表 (sys_dept)
-- =========================================================
CREATE TABLE sys_dept (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    parent_id BIGINT DEFAULT 0,
    name VARCHAR(128) NOT NULL,
    ancestors VARCHAR(512),         -- 祖先路径 (0,1,2)
    sort INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_dept_tenant ON sys_dept(tenant_id);

-- =========================================================
-- 6. 用户表 (sys_user)
-- =========================================================
CREATE TABLE sys_user (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    dept_id BIGINT,
    username VARCHAR(64) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(64),
    mobile VARCHAR(20),
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
-- 唯一索引：同一个租户下用户名唯一
CREATE UNIQUE INDEX uk_user_tenant_name ON sys_user(tenant_id, username) WHERE deleted_at IS NULL;

-- =========================================================
-- 7. 角色表 (sys_role)
-- =========================================================
CREATE TABLE sys_role (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    name VARCHAR(64) NOT NULL,
    code VARCHAR(64) NOT NULL,      -- 角色标识 (如: admin)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
-- 唯一索引：同一个租户下角色编码唯一
CREATE UNIQUE INDEX uk_role_tenant_code ON sys_role(tenant_id, code) WHERE deleted_at IS NULL;

-- =========================================================
-- 8. 用户角色关联表 (sys_user_role)
-- =========================================================
CREATE TABLE sys_user_role (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =========================================================
-- 9. 角色权限关联表 (sys_role_permission)
-- =========================================================
CREATE TABLE sys_role_permission (
    id BIGINT PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    data_scope VARCHAR(20) DEFAULT 'SELF', -- SELF, DEPT, DEPT_SUB, ALL
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON COLUMN sys_role_permission.data_scope IS '数据范围: SELF(个人), DEPT(本部门), DEPT_SUB(本部门及下级), ALL(全租户)';

-- =========================================================
-- 初始化数据 (Seed Data)
-- =========================================================

-- 1. 初始化套餐 (ID: 1)
INSERT INTO sys_package (id, name, created_at, updated_at) VALUES (1, '全功能版套餐', NOW(), NOW());

-- 2. 初始化系统租户 (ID: 1)
INSERT INTO sys_tenant (id, code, name, package_id, created_at, updated_at) 
VALUES (1, 'system', '系统管理总部', 1, NOW(), NOW());

-- 3. 初始化根部门 (ID: 1)
INSERT INTO sys_dept (id, tenant_id, parent_id, name, ancestors, sort, created_at, updated_at) 
VALUES (1, 1, 0, '总经办', '0', 0, NOW(), NOW());

-- 4. 初始化超级管理员 (ID: 1)
INSERT INTO sys_user (id, tenant_id, dept_id, username, password_hash, name, status, created_at, updated_at) 
VALUES (1, 1, 1, 'root', '$2a$10$EIxZaYVK1fsbw1ZfbX3OXePaWxn96p36WQoeG6Lruj3vjPGga31lW', '超级管理员', 1, NOW(), NOW());

-- 5. 初始化角色 (ID: 1, 编码: admin)
INSERT INTO sys_role (id, tenant_id, name, code, created_at, updated_at)
VALUES (1, 1, '系统超级管理员', 'admin', NOW(), NOW());

-- 6. 绑定用户角色
INSERT INTO sys_user_role (id, tenant_id, user_id, role_id, created_at)
VALUES (1, 1, 1, 1, NOW());

-- 7. 初始化权限数据（sys_permission）
-- 7.1 菜单权限
INSERT INTO sys_permission (id, tenant_id, parent_id, name, code, type, path, component, redirect, icon, sort, order_no, created_at, updated_at) VALUES
(1, 1, 0, '系统管理', 'system:common', 'MENU', '/system', 'LAYOUT', '/system/user', 'setting', 1, 1, NOW(), NOW()),
(2, 1, 1, '用户管理', 'system:user:common', 'MENU', '/system/user', '/system/user/index', '', 'user', 1, 1, NOW(), NOW()),
(3, 1, 1, '角色管理', 'system:role:common', 'MENU', '/system/role', '/system/role/index', '', 'role', 2, 2, NOW(), NOW()),
(4, 1, 1, '权限管理', 'system:permission:common', 'MENU', '/system/permission', '/system/permission/index', '', 'lock', 3, 3, NOW(), NOW()),
(5, 1, 1, '部门管理', 'system:dept:common', 'MENU', '/system/dept', '/system/dept/index', '', 'department', 4, 4, NOW(), NOW()),
(6, 1, 1, '租户管理', 'system:tenant:common', 'MENU', '/system/tenant', '/system/tenant/index', '', 'tenant', 5, 5, NOW(), NOW()),
(7, 1, 1, '套餐管理', 'system:package:common', 'MENU', '/system/package', '/system/package/index', '', 'package', 6, 6, NOW(), NOW());

-- 7.2 API 权限
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
(606, 1, 7, '分配套餐权限', 'system:package:assignPermissions', 'API', '/api/v1/package/:id/permissions', 'PUT', 6, NOW(), NOW());

-- 8. 绑定角色权限（admin 角色拥有所有权限）
INSERT INTO sys_role_permission (id, tenant_id, role_id, permission_id, data_scope, dept_ids, created_at)
SELECT id, tenant_id, role_id, permission_id, 1, '', NOW()
FROM (
    SELECT 1 as id, 1 as tenant_id, 1 as role_id, permission_id
    FROM sys_permission
    WHERE tenant_id = 1
) t;