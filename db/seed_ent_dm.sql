-- =============================================================================
-- Contful Enterprise DM8 种子数据
-- =============================================================================

-- 1. 系统角色
INSERT INTO CONTFUL_ENT.contful_system_roles (id, name, description, is_system, permissions, created_time, updated_time)
SELECT '00000000-0000-0000-0000-000000000101', 'Super Admin', '超级管理员，拥有所有权限', '1', '["users:read","users:write","users:delete","sites:read","sites:write","sites:delete","tokens:read","tokens:write","tokens:delete","settings:read","settings:write","audit:read","audit:export","roles:read","roles:write","roles:delete","schema:read","schema:write","entry:read","entry:write","asset:read","asset:write"]', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_roles WHERE id = '00000000-0000-0000-0000-000000000101');

INSERT INTO CONTFUL_ENT.contful_system_roles (id, name, description, is_system, permissions, created_time, updated_time)
SELECT '00000000-0000-0000-0000-000000000102', 'Content Editor', '内容编辑，可管理内容和媒体文件', '1', '["users:read","sites:read","tokens:read","settings:read","audit:read","schema:read","schema:write","entry:read","entry:write","asset:read","asset:write"]', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_roles WHERE id = '00000000-0000-0000-0000-000000000102');

INSERT INTO CONTFUL_ENT.contful_system_roles (id, name, description, is_system, permissions, created_time, updated_time)
SELECT '00000000-0000-0000-0000-000000000103', 'Auditor', '审计人员，仅可查看和导出审计日志', '1', '["users:read","sites:read","tokens:read","settings:read","audit:read","audit:export"]', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_roles WHERE id = '00000000-0000-0000-0000-000000000103');

-- 2. 权限元数据
MERGE INTO CONTFUL_ENT.contful_system_permission_groups t
USING (SELECT '00000000-0000-0000-0000-000000000301' AS id, 'dashboard' AS gkey, 'Dashboard' AS label, 'Dashboard' AS label_en, 0 AS srt FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000302', 'users', 'User Management', 'User Management', 1 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000303', 'sites', 'Site Management', 'Site Management', 2 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000304', 'tokens', 'API Tokens', 'API Tokens', 3 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000305', 'settings', 'System Settings', 'System Settings', 4 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000306', 'audit', 'Audit Logs', 'Audit Logs', 5 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000307', 'roles', 'Role Management', 'Role Management', 6 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000308', 'schema', 'Content Schemas', 'Content Schemas', 7 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000309', 'entry', 'Entries', 'Entries', 8 FROM DUAL UNION ALL
       SELECT '00000000-0000-0000-0000-000000000310', 'asset', 'Assets', 'Assets', 9 FROM DUAL) s
ON (t.group_key = s.gkey)
WHEN NOT MATCHED THEN INSERT (id, group_key, label, label_en, sort_order) VALUES (s.id, s.gkey, s.label, s.label_en, s.srt);

-- 3. 管理员用户（密码：contful@com）
INSERT INTO CONTFUL_ENT.contful_system_users (id, email, password_hash, nickname, status, is_super_admin, created_time, updated_time)
SELECT '00000000-0000-0000-0000-000000000001', 'admin@contful.com', '$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu', 'Administrator', 'active', '1', SYSTIMESTAMP, SYSTIMESTAMP
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_users WHERE id = '00000000-0000-0000-0000-000000000001');

-- 4. 用户-角色关联
INSERT INTO CONTFUL_ENT.contful_system_user_roles (user_id, role_id, created_time)
SELECT '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000101', SYSTIMESTAMP
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_user_roles WHERE user_id = '00000000-0000-0000-0000-000000000001' AND role_id = '00000000-0000-0000-0000-000000000101');

-- 5. 系统配置
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'password_expire_days', '90', 'number', 'Password expiry (days), 0 = never', '0', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'password_expire_days');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'site_name', 'Contful', 'string', 'Site name', '1', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'site_name');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'site_description', 'Open Source Headless CMS', 'string', 'Site description', '1', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'site_description');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'logo_url', '', 'string', 'Logo URL', '1', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'logo_url');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'login_background_url', '', 'string', 'Login background', '1', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'login_background_url');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'password_min_length', '8', 'number', 'Min password length', '0', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'password_min_length');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'password_require_uppercase', 'true', 'boolean', 'Require uppercase', '0', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'password_require_uppercase');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'password_require_lowercase', 'true', 'boolean', 'Require lowercase', '0', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'password_require_lowercase');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'password_require_number', 'true', 'boolean', 'Require number', '0', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'password_require_number');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'password_require_special', 'false', 'boolean', 'Require special char', '0', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'password_require_special');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'mfa_enforced', 'false', 'boolean', 'Enforce MFA', '1', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'mfa_enforced');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'login_max_attempts', '5', 'number', 'Max login attempts', '1', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'login_max_attempts');
INSERT INTO CONTFUL_ENT.contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
SELECT 'login_lock_duration', '30', 'number', 'Lock duration (min)', '1', '1', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_config WHERE config_key = 'login_lock_duration');
