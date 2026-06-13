-- =============================================================================
-- Contful Enterprise DM8 数据库初始化脚本
-- 达梦数据库 (Oracle 兼容模式) — UTF-8, 不区分大小写
-- 用法: 以 SYSDBA 身份执行，需先手动创建 CONTFUL_ENT 用户并授权 DBA
-- =============================================================================

-- =============================================================================
-- UUID 辅助函数：SYS_GUID() → 标准 UUID 格式 (8-4-4-4-12)
-- =============================================================================
CREATE OR REPLACE FUNCTION CONTFUL_ENT.GEN_UUID RETURN VARCHAR2 IS
  v_raw VARCHAR2(32);
  v_hex VARCHAR2(36);
BEGIN
  SELECT RAWTOHEX(SYS_GUID()) INTO v_raw FROM DUAL;
  v_hex := LOWER(SUBSTR(v_raw,1,8) || '-' || SUBSTR(v_raw,9,4) || '-' ||
           SUBSTR(v_raw,13,4) || '-' || SUBSTR(v_raw,17,4) || '-' ||
           SUBSTR(v_raw,21,12));
  RETURN v_hex;
END;
/

-- =============================================================================
-- 1. 系统用户表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_system_users (
    id VARCHAR2(36) ,
    email VARCHAR2(255) NOT NULL,
    password_hash VARCHAR2(255) NOT NULL,
    nickname VARCHAR2(100),
    avatar_url VARCHAR2(500),
    mfa_secret VARCHAR2(255),
    mfa_enabled CHAR(1) DEFAULT '0',
    is_super_admin CHAR(1) DEFAULT '0',
    status VARCHAR2(20) DEFAULT 'active',
    last_login_time TIMESTAMP,
    last_login_ip VARCHAR2(45),
    login_failures NUMBER DEFAULT 0,
    locked_until TIMESTAMP,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_system_users PRIMARY KEY (id),
    CONSTRAINT uq_system_users_email UNIQUE (email)
);
CREATE INDEX idx_system_users_status ON CONTFUL_ENT.contful_system_users(status);
CREATE INDEX idx_system_users_deleted ON CONTFUL_ENT.contful_system_users(deleted_time);

-- 触发器：自动更新 updated_time
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_system_users_ut
BEFORE UPDATE ON CONTFUL_ENT.contful_system_users FOR EACH ROW
BEGIN :NEW.updated_time := SYSTIMESTAMP; END;
/

-- =============================================================================
-- 2. 站点表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_sites (
    id VARCHAR2(36) ,
    name VARCHAR2(255) NOT NULL,
    slug VARCHAR2(255) NOT NULL,
    description CLOB,
    locale VARCHAR2(10) DEFAULT 'zh-CN',
    timezone VARCHAR2(50) DEFAULT 'Asia/Shanghai',
    is_active CHAR(1) DEFAULT '1',
    settings CLOB DEFAULT '{}',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_sites PRIMARY KEY (id),
    CONSTRAINT uq_sites_slug UNIQUE (slug)
);
CREATE INDEX idx_sites_active ON CONTFUL_ENT.contful_sites(is_active);

-- =============================================================================
-- 3. 内容模型表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_schemas (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    slug VARCHAR2(255) NOT NULL,
    description CLOB,
    type VARCHAR2(20) NOT NULL CHECK (type IN ('collection', 'single')),
    is_active CHAR(1) DEFAULT '1',
    settings CLOB DEFAULT '{}',
    sort_order NUMBER DEFAULT 0,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_schemas PRIMARY KEY (id),
    CONSTRAINT fk_schemas_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.contful_sites(id),
    CONSTRAINT uq_schemas_slug UNIQUE (site_id, slug)
);
CREATE INDEX idx_schemas_site ON CONTFUL_ENT.contful_schemas(site_id);

-- =============================================================================
-- 4. 字段定义表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_fields (
    id VARCHAR2(36) ,
    schema_id VARCHAR2(36) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    label VARCHAR2(255),
    field_type VARCHAR2(50) NOT NULL,
    required CHAR(1) DEFAULT '0',
    default_value CLOB,
    validation_rules CLOB DEFAULT '{}',
    options CLOB DEFAULT '{}',
    sort_order NUMBER DEFAULT 0,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_fields PRIMARY KEY (id),
    CONSTRAINT fk_fields_schema FOREIGN KEY (schema_id) REFERENCES CONTFUL_ENT.contful_schemas(id)
);
CREATE INDEX idx_fields_schema ON CONTFUL_ENT.contful_fields(schema_id);

-- =============================================================================
-- 5. 内容条目表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_entries (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36) NOT NULL,
    schema_id VARCHAR2(36) NOT NULL,
    slug VARCHAR2(500),
    locale VARCHAR2(10) DEFAULT 'zh-CN',
    status VARCHAR2(20) DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    title VARCHAR2(500),
    seo_title VARCHAR2(500),
    seo_description CLOB,
    seo_keywords VARCHAR2(500),
    version NUMBER DEFAULT 1,
    published_time TIMESTAMP,
    scheduled_publish_time TIMESTAMP,
    scheduled_unpublish_time TIMESTAMP,
    created_by VARCHAR2(36),
    updated_by VARCHAR2(36),
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_entries PRIMARY KEY (id),
    CONSTRAINT fk_entries_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.contful_sites(id),
    CONSTRAINT fk_entries_schema FOREIGN KEY (schema_id) REFERENCES CONTFUL_ENT.contful_schemas(id)
);
CREATE INDEX idx_entries_site ON CONTFUL_ENT.contful_entries(site_id);
CREATE INDEX idx_entries_schema ON CONTFUL_ENT.contful_entries(schema_id);
CREATE INDEX idx_entries_status ON CONTFUL_ENT.contful_entries(status);
CREATE INDEX idx_entries_slug ON CONTFUL_ENT.contful_entries(site_id, schema_id, slug);
CREATE INDEX idx_entries_scheduled_publish ON CONTFUL_ENT.contful_entries(scheduled_publish_time);
CREATE INDEX idx_entries_scheduled_unpublish ON CONTFUL_ENT.contful_entries(scheduled_unpublish_time);

-- =============================================================================
-- 6. 条目值表（EAV 模式）
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_entry_values (
    id VARCHAR2(36) ,
    entry_id VARCHAR2(36) NOT NULL,
    field_id VARCHAR2(36) NOT NULL,
    field_name VARCHAR2(255),
    value_text CLOB,
    value_number NUMBER,
    value_boolean CHAR(1),
    value_json CLOB,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_entry_values PRIMARY KEY (id),
    CONSTRAINT fk_entry_values_entry FOREIGN KEY (entry_id) REFERENCES CONTFUL_ENT.contful_entries(id)
);
CREATE INDEX idx_entry_values_entry ON CONTFUL_ENT.contful_entry_values(entry_id);
CREATE INDEX idx_entry_values_field ON CONTFUL_ENT.contful_entry_values(entry_id, field_id);

-- =============================================================================
-- 7. 资产文件夹
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_asset_folders (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    slug VARCHAR2(255) NOT NULL,
    path VARCHAR2(500) NOT NULL,
    parent_id VARCHAR2(36),
    sort_order NUMBER DEFAULT 0,
    created_by VARCHAR2(36),
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_asset_folders PRIMARY KEY (id),
    CONSTRAINT fk_asset_folders_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.contful_sites(id)
);
CREATE INDEX idx_asset_folders_site ON CONTFUL_ENT.contful_asset_folders(site_id);
CREATE INDEX idx_asset_folders_parent ON CONTFUL_ENT.contful_asset_folders(parent_id);

-- =============================================================================
-- 8. 资产表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_assets (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36) NOT NULL,
    folder_id VARCHAR2(36),
    filename VARCHAR2(500) NOT NULL,
    original_name VARCHAR2(500),
    mime_type VARCHAR2(100),
    file_size NUMBER,
    width NUMBER,
    height NUMBER,
    alt_text VARCHAR2(500),
    caption CLOB,
    tags VARCHAR2(500),
    storage_driver VARCHAR2(50),
    url VARCHAR2(2000),
    metadata CLOB DEFAULT '{}',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_assets PRIMARY KEY (id),
    CONSTRAINT fk_assets_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.contful_sites(id),
    CONSTRAINT fk_assets_folder FOREIGN KEY (folder_id) REFERENCES CONTFUL_ENT.contful_asset_folders(id)
);
CREATE INDEX idx_assets_site ON CONTFUL_ENT.contful_assets(site_id);
CREATE INDEX idx_assets_folder ON CONTFUL_ENT.contful_assets(folder_id);

-- =============================================================================
-- 9. API Token 表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_tokens (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    description CLOB,
    token_prefix VARCHAR2(10),
    token_hash VARCHAR2(255),
    encrypted_token VARCHAR2(2000),
    expires_time TIMESTAMP,
    status VARCHAR2(20) DEFAULT 'active',
    last_used_time TIMESTAMP,
    last_used_ip VARCHAR2(45),
    created_by VARCHAR2(36),
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_tokens PRIMARY KEY (id),
    CONSTRAINT fk_tokens_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.contful_sites(id)
);
CREATE INDEX idx_tokens_site ON CONTFUL_ENT.contful_tokens(site_id);

-- =============================================================================
-- 10. 审计日志表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_audit_logs (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36),
    user_id VARCHAR2(36),
    action VARCHAR2(100) NOT NULL,
    resource_type VARCHAR2(50),
    resource_id VARCHAR2(36),
    category VARCHAR2(50),
    level VARCHAR2(20) DEFAULT 'info',
    details CLOB,
    ip_address VARCHAR2(45),
    user_agent CLOB,
    data_signature VARCHAR2(255),
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_audit_logs PRIMARY KEY (id)
);
CREATE INDEX idx_audit_logs_category ON CONTFUL_ENT.contful_audit_logs(category, created_time);
CREATE INDEX idx_audit_logs_level ON CONTFUL_ENT.contful_audit_logs(level, created_time);
CREATE INDEX idx_audit_logs_user ON CONTFUL_ENT.contful_audit_logs(user_id);

-- =============================================================================
-- 11. 系统角色
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_system_roles (
    id VARCHAR2(36) ,
    name VARCHAR2(255) NOT NULL,
    description CLOB,
    is_system CHAR(1) DEFAULT '0',
    permissions CLOB DEFAULT '[]',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_system_roles PRIMARY KEY (id)
);

-- =============================================================================
-- 12. 用户-角色关联
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_system_user_roles (
    user_id VARCHAR2(36) NOT NULL,
    role_id VARCHAR2(36) NOT NULL,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_system_user_roles PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_sur_user FOREIGN KEY (user_id) REFERENCES CONTFUL_ENT.contful_system_users(id),
    CONSTRAINT fk_sur_role FOREIGN KEY (role_id) REFERENCES CONTFUL_ENT.contful_system_roles(id)
);

-- =============================================================================
-- 13. 系统配置
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_system_config (
    config_key VARCHAR2(100) NOT NULL,
    config_value CLOB,
    value_type VARCHAR2(20) DEFAULT 'string',
    description CLOB,
    is_public CHAR(1) DEFAULT '0',
    is_system CHAR(1) DEFAULT '0',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_system_config PRIMARY KEY (config_key)
);

-- =============================================================================
-- 14. 权限分组
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_system_permission_groups (
    id VARCHAR2(36) ,
    group_key VARCHAR2(100) NOT NULL,
    label VARCHAR2(255),
    label_en VARCHAR2(255),
    sort_order NUMBER DEFAULT 0,
    CONSTRAINT pk_permission_groups PRIMARY KEY (id),
    CONSTRAINT uq_permission_groups_key UNIQUE (group_key)
);

-- =============================================================================
-- 15. 权限项
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_system_permissions (
    group_id VARCHAR2(36) NOT NULL,
    action VARCHAR2(100) NOT NULL,
    label VARCHAR2(255),
    label_en VARCHAR2(255),
    sort_order NUMBER DEFAULT 0,
    CONSTRAINT pk_permissions PRIMARY KEY (group_id, action),
    CONSTRAINT fk_perms_group FOREIGN KEY (group_id) REFERENCES CONTFUL_ENT.contful_system_permission_groups(id)
);

-- =============================================================================
-- 16. Webhook 配置表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_webhooks (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    url VARCHAR2(2000) NOT NULL,
    events CLOB DEFAULT '{}',
    secret VARCHAR2(255),
    is_active CHAR(1) DEFAULT '1',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_webhooks PRIMARY KEY (id),
    CONSTRAINT fk_webhooks_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.contful_sites(id)
);
CREATE INDEX idx_webhooks_site ON CONTFUL_ENT.contful_webhooks(site_id);
CREATE INDEX idx_webhooks_active ON CONTFUL_ENT.contful_webhooks(is_active);

-- =============================================================================
-- 17. Webhook 投递记录
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_webhook_deliveries (
    id VARCHAR2(36) ,
    webhook_id VARCHAR2(36) NOT NULL,
    event VARCHAR2(50),
    payload CLOB,
    response_status NUMBER,
    response_body CLOB,
    status VARCHAR2(20) DEFAULT 'pending',
    attempt NUMBER DEFAULT 1,
    error_message CLOB,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_webhook_deliveries PRIMARY KEY (id),
    CONSTRAINT fk_wd_webhook FOREIGN KEY (webhook_id) REFERENCES CONTFUL_ENT.contful_webhooks(id)
);
CREATE INDEX idx_wd_webhook ON CONTFUL_ENT.contful_webhook_deliveries(webhook_id);
CREATE INDEX idx_wd_status ON CONTFUL_ENT.contful_webhook_deliveries(status);
CREATE INDEX idx_wd_created ON CONTFUL_ENT.contful_webhook_deliveries(created_time);

-- =============================================================================

-- =============================================================================
-- UUID 主键 BEFORE INSERT 触发器
-- =============================================================================
-- UUID 主键 BEFORE INSERT 触发器
-- DM8 DEFAULT 不支持函数调用，改用触发器
-- =============================================================================
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_system_users_bi
BEFORE INSERT ON CONTFUL_ENT.contful_system_users FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_sites_bi
BEFORE INSERT ON CONTFUL_ENT.contful_sites FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_schemas_bi
BEFORE INSERT ON CONTFUL_ENT.contful_schemas FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_fields_bi
BEFORE INSERT ON CONTFUL_ENT.contful_fields FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_entries_bi
BEFORE INSERT ON CONTFUL_ENT.contful_entries FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_entry_values_bi
BEFORE INSERT ON CONTFUL_ENT.contful_entry_values FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_asset_folders_bi
BEFORE INSERT ON CONTFUL_ENT.contful_asset_folders FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_assets_bi
BEFORE INSERT ON CONTFUL_ENT.contful_assets FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_tokens_bi
BEFORE INSERT ON CONTFUL_ENT.contful_tokens FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_audit_logs_bi
BEFORE INSERT ON CONTFUL_ENT.contful_audit_logs FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_system_roles_bi
BEFORE INSERT ON CONTFUL_ENT.contful_system_roles FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_perm_groups_bi
BEFORE INSERT ON CONTFUL_ENT.contful_system_permission_groups FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_webhooks_bi
BEFORE INSERT ON CONTFUL_ENT.contful_webhooks FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_wd_bi
BEFORE INSERT ON CONTFUL_ENT.contful_webhook_deliveries FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;

-- =============================================================================
-- 初始化完成
-- =============================================================================
COMMIT;
-- =============================================================================
-- Contful Enterprise DM8 种子数据
-- =============================================================================

-- 1. 默认站点
INSERT INTO CONTFUL_ENT.contful_sites (id, name, slug, description, locale, timezone, is_active, settings, created_time, updated_time)
SELECT '00000000-0000-0000-0000-000000000001', '默认站点', 'default', '系统默认站点', 'zh-CN', 'Asia/Shanghai', '1', '{}', SYSTIMESTAMP, SYSTIMESTAMP FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_sites WHERE id = '00000000-0000-0000-0000-000000000001');

-- 2. 系统角色
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
SELECT '00000000-0000-0000-0000-000000000001', 'admin@contful.com', q'[$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu]', 'Administrator', 'active', '1', SYSTIMESTAMP, SYSTIMESTAMP
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
