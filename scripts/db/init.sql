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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX uk_perm_code ON sys_permission(code) WHERE deleted_at IS NULL;
COMMENT ON COLUMN sys_permission.api_path IS 'Kratos内部路径/API路径';

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