-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Contful - 数据库种子数据
-- 用途：开发和环境测试的初始数据
-- 执行方式：psql -h <host> -U <user> -d <db> -f seed_data.sql
-- 幂等设计：使用 INSERT ... WHERE NOT EXISTS 避免重复插入
-- =============================================================================

-- =============================================================================
-- 1. 默认站点
-- =============================================================================

INSERT INTO sites (id, name, slug, description, locale, timezone, is_active, settings)
SELECT
    '00000000-0000-0000-0000-000000000001'::uuid,
    '默认站点',
    'default',
    '系统默认站点',
    'zh-CN',
    'Asia/Shanghai',
    TRUE,
    '{}'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM sites WHERE id = '00000000-0000-0000-0000-000000000001'::uuid);

-- =============================================================================
-- 2. 系统角色
-- =============================================================================

-- 超级管理员角色（拥有全部系统级权限）
INSERT INTO system_roles (id, name, description, is_system, permissions, created_time, updated_time)
SELECT
    '00000000-0000-0000-0000-000000000101'::uuid,
    '超级管理员',
    '系统超级管理员，拥有全部权限',
    TRUE,
    '["dashboard:read",
      "users:read","users:write","users:delete",
      "sites:read","sites:write","sites:delete",
      "tokens:read","tokens:write","tokens:delete",
      "settings:read","settings:write",
      "audit:read","audit:export",
      "roles:read","roles:write","roles:delete"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000101'::uuid);

-- 审计员角色（仅可查看和导出审计日志）
INSERT INTO system_roles (id, name, description, is_system, permissions, created_time, updated_time)
SELECT
    '00000000-0000-0000-0000-000000000103'::uuid,
    '审计人员',
    '审计人员，仅可查看和导出审计日志',
    TRUE,
    '["users:read","sites:read","tokens:read",
      "settings:read","audit:read","audit:export"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000103'::uuid);

-- =============================================================================
-- 2.5 权限元数据（分组 + 权限项，用于角色权限配置）
-- =============================================================================

-- 权限分组
INSERT INTO system_permission_groups (id, group_key, label, label_en, sort_order)
SELECT * FROM (VALUES
    ('00000000-0000-0000-0000-000000000301'::uuid, 'dashboard',      '仪表盘',     'Dashboard',      0),
    ('00000000-0000-0000-0000-000000000302'::uuid, 'users',          '用户管理',   'User Management', 1),
    ('00000000-0000-0000-0000-000000000303'::uuid, 'sites',          '站点管理',   'Site Management',  2),
    ('00000000-0000-0000-0000-000000000304'::uuid, 'tokens',         'API Token',  'API Tokens',     3),
    ('00000000-0000-0000-0000-000000000305'::uuid, 'settings',       '系统设置',   'System Settings',  4),
    ('00000000-0000-0000-0000-000000000306'::uuid, 'audit',          '审计日志',   'Audit Logs',     5),
    ('00000000-0000-0000-0000-000000000307'::uuid, 'roles',          '角色管理',   'Role Management', 6),
    ('00000000-0000-0000-0000-000000000308'::uuid, 'schema', '内容模型',   'Content Schemas', 7),
    ('00000000-0000-0000-0000-000000000309'::uuid, 'entry',          '内容条目',   'Entries',        8),
    ('00000000-0000-0000-0000-000000000310'::uuid, 'asset',          '媒体文件',   'Assets',         9)
) AS t(id, group_key, label, label_en, sort_order)
WHERE NOT EXISTS (SELECT 1 FROM system_permission_groups WHERE id = t.id::uuid);

-- 权限项
INSERT INTO system_permissions (group_id, action, label, label_en, sort_order)
SELECT g.id, t.action, t.label, t.label_en, t.sort_order
FROM (VALUES
    ('00000000-0000-0000-0000-000000000301'::uuid, 'read',   '查看',      'View',             0),
    ('00000000-0000-0000-0000-000000000302'::uuid, 'read',   '查看用户',  'View Users',        0),
    ('00000000-0000-0000-0000-000000000302'::uuid, 'write',  '管理用户',  'Manage Users',      1),
    ('00000000-0000-0000-0000-000000000302'::uuid, 'delete', '删除用户',  'Delete Users',      2),
    ('00000000-0000-0000-0000-000000000303'::uuid, 'read',   '查看站点',  'View Sites',        0),
    ('00000000-0000-0000-0000-000000000303'::uuid, 'write',  '管理站点',  'Manage Sites',      1),
    ('00000000-0000-0000-0000-000000000303'::uuid, 'delete', '删除站点',  'Delete Sites',      2),
    ('00000000-0000-0000-0000-000000000304'::uuid, 'read',   '查看 Token','View Tokens',       0),
    ('00000000-0000-0000-0000-000000000304'::uuid, 'write',  '管理 Token','Manage Tokens',     1),
    ('00000000-0000-0000-0000-000000000304'::uuid, 'delete', '删除 Token','Delete Tokens',     2),
    ('00000000-0000-0000-0000-000000000305'::uuid, 'read',   '查看设置',  'View Settings',     0),
    ('00000000-0000-0000-0000-000000000305'::uuid, 'write',  '修改设置',  'Edit Settings',     1),
    ('00000000-0000-0000-0000-000000000306'::uuid, 'read',   '查看日志',  'View Logs',         0),
    ('00000000-0000-0000-0000-000000000306'::uuid, 'export', '导出日志',  'Export Logs',       1),
    ('00000000-0000-0000-0000-000000000307'::uuid, 'read',   '查看角色',  'View Roles',        0),
    ('00000000-0000-0000-0000-000000000307'::uuid, 'write',  '管理角色',  'Manage Roles',      1),
    ('00000000-0000-0000-0000-000000000307'::uuid, 'delete', '删除角色',  'Delete Roles',      2),
    ('00000000-0000-0000-0000-000000000308'::uuid, 'read',   '查看模型',  'View Schemas',      0),
    ('00000000-0000-0000-0000-000000000308'::uuid, 'write',  '管理模型',  'Manage Schemas',    1),
    ('00000000-0000-0000-0000-000000000308'::uuid, 'delete', '删除模型',  'Delete Schemas',    2),
    ('00000000-0000-0000-0000-000000000309'::uuid, 'read',   '查看条目',  'View Entries',      0),
    ('00000000-0000-0000-0000-000000000309'::uuid, 'write',  '编辑条目',  'Edit Entries',      1),
    ('00000000-0000-0000-0000-000000000309'::uuid, 'publish','发布条目',  'Publish Entries',   2),
    ('00000000-0000-0000-0000-000000000309'::uuid, 'delete', '删除条目',  'Delete Entries',    3),
    ('00000000-0000-0000-0000-000000000310'::uuid, 'read',   '查看文件',  'View Assets',       0),
    ('00000000-0000-0000-0000-000000000310'::uuid, 'write',  '管理文件',  'Manage Assets',     1),
    ('00000000-0000-0000-0000-000000000310'::uuid, 'delete', '删除文件',  'Delete Assets',     2)
) AS t(group_id, action, label, label_en, sort_order)
JOIN system_permission_groups g ON g.id = t.group_id::uuid
WHERE NOT EXISTS (
    SELECT 1 FROM system_permissions p WHERE p.group_id = g.id AND p.action = t.action
);

-- =============================================================================
-- 3. 系统用户
-- =============================================================================

-- 默认管理员用户（密码：contful@com）
INSERT INTO system_users (id, email, password_hash, nickname, status, is_super_admin, created_time, updated_time)
SELECT
    '00000000-0000-0000-0000-000000000001'::uuid,
    'admin@contful.com',
    '$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu',  -- 密码：contful@com
    'Administrator',
    'active',
    TRUE,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_users WHERE id = '00000000-0000-0000-0000-000000000001'::uuid);

-- =============================================================================
-- 4. 用户-角色关联
-- =============================================================================

-- 关联 admin 用户与 Super Admin 角色
INSERT INTO system_user_roles (user_id, role_id, created_time)
SELECT
    '00000000-0000-0000-0000-000000000001'::uuid,
    '00000000-0000-0000-0000-000000000101'::uuid,
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM system_user_roles 
    WHERE user_id = '00000000-0000-0000-0000-000000000001'::uuid 
    AND role_id = '00000000-0000-0000-0000-000000000101'::uuid
);

-- =============================================================================
-- 5. 系统配置
-- =============================================================================

INSERT INTO system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
VALUES
    ('password_expire_days', '90', 'number', '密码有效期（天），0 表示永不过期', FALSE, TRUE, NOW(), NOW()),
    ('site_name', 'Contful', 'string', '系统名称', TRUE, TRUE, NOW(), NOW()),
    ('site_description', '开源 Headless CMS', 'string', '系统描述', TRUE, TRUE, NOW(), NOW()),
    ('logo_url', '', 'string', '系统 Logo 图片地址', TRUE, TRUE, NOW(), NOW()),
    ('login_background_url', '', 'string', '登录页背景图片地址', TRUE, TRUE, NOW(), NOW()),
    ('password_min_length', '8', 'number', '密码最小长度', FALSE, TRUE, NOW(), NOW()),
    ('password_require_uppercase', 'true', 'boolean', '密码必须包含大写字母', FALSE, TRUE, NOW(), NOW()),
    ('password_require_lowercase', 'true', 'boolean', '密码必须包含小写字母', FALSE, TRUE, NOW(), NOW()),
    ('password_require_number', 'true', 'boolean', '密码必须包含数字', FALSE, TRUE, NOW(), NOW()),
    ('password_require_special', 'false', 'boolean', '密码必须包含特殊字符', FALSE, TRUE, NOW(), NOW()),
    ('mfa_enforced', 'false', 'boolean', '是否强制所有用户启用 MFA 双因子认证', TRUE, TRUE, NOW(), NOW()),
    ('login_max_attempts', '5', 'number', '登录失败次数上限（连续失败达到此值后锁定）', TRUE, TRUE, NOW(), NOW()),
    ('login_lock_duration', '30', 'number', '账号锁定时长（分钟）', TRUE, TRUE, NOW(), NOW())
ON CONFLICT (config_key) DO NOTHING;

-- =============================================================================
-- 完成提示
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE '种子数据插入完成！';
    RAISE NOTICE '默认管理员账号：admin@contful.com';
    RAISE NOTICE '默认管理员密码：Admin@123';
    RAISE NOTICE '请登录后立即修改默认密码！';
END $$;
