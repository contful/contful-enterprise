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
INSERT INTO system_roles (id, name, description, is_system, permissions, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000101'::uuid,
    'Super Admin',
    '系统超级管理员，拥有全部权限',
    TRUE,
    '["system:users:read","system:users:write","system:users:delete",
      "system:sites:read","system:sites:write","system:sites:delete",
      "system:tokens:read","system:tokens:write","system:tokens:delete",
      "system:settings:read","system:settings:write",
      "system:audit:read","system:audit:export",
      "system:roles:read","system:roles:write","system:roles:delete"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000101'::uuid);

-- 插件管理员角色
INSERT INTO system_roles (id, name, description, is_system, permissions, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000102'::uuid,
    'Plugin Manager',
    '插件管理员',
    TRUE,
    '["plugins:read", "plugins:write", "plugins:install", "plugins:uninstall"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000102'::uuid);

-- 审计员角色（仅可查看和导出审计日志）
INSERT INTO system_roles (id, name, description, is_system, permissions, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000103'::uuid,
    'Auditor',
    '审计人员，仅可查看和导出审计日志',
    TRUE,
    '["system:users:read","system:sites:read","system:tokens:read",
      "system:settings:read","system:audit:read","system:audit:export"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000103'::uuid);

-- =============================================================================
-- 3. 系统用户
-- =============================================================================

-- 默认管理员用户（密码：Admin@123）
INSERT INTO system_users (id, email, password_hash, nickname, status, is_super_admin, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000001'::uuid,
    'admin@contful.com',
    '$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu',  -- 密码：Admin@123
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
INSERT INTO system_user_roles (user_id, role_id, created_at)
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

INSERT INTO system_config (config_key, config_value, value_type, description, is_public, created_at, updated_at)
VALUES
    ('password_expire_days', '90', 'number', '密码有效期（天），0 表示永不过期', FALSE, NOW(), NOW()),
    ('site_name', 'Contful', 'string', '系统名称', TRUE, NOW(), NOW()),
    ('logo_url', '', 'string', '系统 Logo 图片地址', TRUE, NOW(), NOW()),
    ('login_background_url', '', 'string', '登录页背景图片地址', TRUE, NOW(), NOW()),
    ('password_min_length', '8', 'number', '密码最小长度', FALSE, NOW(), NOW()),
    ('password_require_uppercase', 'true', 'boolean', '密码必须包含大写字母', FALSE, NOW(), NOW()),
    ('password_require_lowercase', 'true', 'boolean', '密码必须包含小写字母', FALSE, NOW(), NOW()),
    ('password_require_number', 'true', 'boolean', '密码必须包含数字', FALSE, NOW(), NOW())
ON CONFLICT (config_key) DO NOTHING;

-- =============================================================================
-- 6. 示例内容模型（可选）
-- =============================================================================

-- 示例：文章模型
INSERT INTO schemas (id, site_id, name, slug, description, kind, display_config, api_config, is_active, sort_order, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000201'::uuid,
    '00000000-0000-0000-0000-000000000001'::uuid,
    '文章',
    'posts',
    '文章内容模型',
    'collection',
    '{"icon":"file-text","color":"blue"}'::jsonb,
    '{"publicRead":false,"publicWrite":false}'::jsonb,
    TRUE,
    0,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM schemas WHERE id = '00000000-0000-0000-0000-000000000201'::uuid);

-- 示例字段：标题
INSERT INTO fields (id, schema_id, name, label, description, field_type, config, validation, display, sort_order, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000301'::uuid,
    '00000000-0000-0000-0000-000000000201'::uuid,
    'title',
    '标题',
    '文章标题',
    'text',
    '{"maxLength":200}'::jsonb,
    '{"required":true,"minLength":1,"maxLength":200}'::jsonb,
    '{"placeholder":"请输入标题"}'::jsonb,
    0,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM fields WHERE id = '00000000-0000-0000-0000-000000000301'::uuid);

-- 示例字段：内容
INSERT INTO fields (id, schema_id, name, label, description, field_type, config, validation, display, sort_order, created_at, updated_at)
SELECT
    '00000000-0000-0000-0000-000000000302'::uuid,
    '00000000-0000-0000-0000-000000000201'::uuid,
    'content',
    '内容',
    '文章正文',
    'rich_text',
    '{}'::jsonb,
    '{"required":true}'::jsonb,
    '{}'::jsonb,
    1,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM fields WHERE id = '00000000-0000-0000-0000-000000000302'::uuid);

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
