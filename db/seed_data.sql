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

-- 插件管理员角色
INSERT INTO system_roles (id, name, description, is_system, permissions, created_time, updated_time)
SELECT
    '00000000-0000-0000-0000-000000000102'::uuid,
    '插件管理员',
    '插件管理员',
    TRUE,
    '["plugins:read", "plugins:write", "plugins:install", "plugins:uninstall"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM system_roles WHERE id = '00000000-0000-0000-0000-000000000102'::uuid);

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
