-- =============================================================================
-- Contful Enterprise DM8 数据库初始化脚本
-- 达梦数据库 (Oracle 兼容模式) — UTF-8, 不区分大小写
-- 用法: 以 SYSDBA 身份在 DM 管理工具中执行
-- =============================================================================

-- 创建用户（schema）— 使用默认表空间
CREATE USER CONTFUL_ENT IDENTIFIED BY "Contful@2024";

GRANT DBA TO CONTFUL_ENT;
GRANT CREATE TABLE TO CONTFUL_ENT;
GRANT CREATE VIEW TO CONTFUL_ENT;
GRANT CREATE PROCEDURE TO CONTFUL_ENT;
GRANT CREATE SEQUENCE TO CONTFUL_ENT;
GRANT CREATE TRIGGER TO CONTFUL_ENT;

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
CREATE TABLE CONTFUL_ENT.system_users (
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
CREATE INDEX idx_system_users_status ON CONTFUL_ENT.system_users(status);
CREATE INDEX idx_system_users_deleted ON CONTFUL_ENT.system_users(deleted_time);

-- 触发器：自动更新 updated_time
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_system_users_ut
BEFORE UPDATE ON CONTFUL_ENT.system_users FOR EACH ROW
BEGIN :NEW.updated_time := SYSTIMESTAMP; END;
/

-- =============================================================================
-- 2. 站点表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.sites (
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
CREATE INDEX idx_sites_active ON CONTFUL_ENT.sites(is_active);

-- =============================================================================
-- 3. 内容模型表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.schemas (
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
    CONSTRAINT fk_schemas_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.sites(id),
    CONSTRAINT uq_schemas_slug UNIQUE (site_id, slug)
);
CREATE INDEX idx_schemas_site ON CONTFUL_ENT.schemas(site_id);

-- =============================================================================
-- 4. 字段定义表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.fields (
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
    CONSTRAINT fk_fields_schema FOREIGN KEY (schema_id) REFERENCES CONTFUL_ENT.schemas(id)
);
CREATE INDEX idx_fields_schema ON CONTFUL_ENT.fields(schema_id);

-- =============================================================================
-- 5. 内容条目表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.entries (
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
    CONSTRAINT fk_entries_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.sites(id),
    CONSTRAINT fk_entries_schema FOREIGN KEY (schema_id) REFERENCES CONTFUL_ENT.schemas(id)
);
CREATE INDEX idx_entries_site ON CONTFUL_ENT.entries(site_id);
CREATE INDEX idx_entries_schema ON CONTFUL_ENT.entries(schema_id);
CREATE INDEX idx_entries_status ON CONTFUL_ENT.entries(status);
CREATE INDEX idx_entries_slug ON CONTFUL_ENT.entries(site_id, schema_id, slug);
CREATE INDEX idx_entries_scheduled_publish ON CONTFUL_ENT.entries(scheduled_publish_time);
CREATE INDEX idx_entries_scheduled_unpublish ON CONTFUL_ENT.entries(scheduled_unpublish_time);

-- =============================================================================
-- 6. 条目值表（EAV 模式）
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_entry_values (
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
    CONSTRAINT fk_entry_values_entry FOREIGN KEY (entry_id) REFERENCES CONTFUL_ENT.entries(id)
);
CREATE INDEX idx_entry_values_entry ON CONTFUL_ENT.t_entry_values(entry_id);
CREATE INDEX idx_entry_values_field ON CONTFUL_ENT.t_entry_values(entry_id, field_id);

-- =============================================================================
-- 7. 资产文件夹
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_asset_folders (
    id VARCHAR2(36) ,
    site_id VARCHAR2(36) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    parent_id VARCHAR2(36),
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    deleted_time TIMESTAMP,
    CONSTRAINT pk_asset_folders PRIMARY KEY (id),
    CONSTRAINT fk_asset_folders_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.sites(id)
);
CREATE INDEX idx_asset_folders_site ON CONTFUL_ENT.t_asset_folders(site_id);

-- =============================================================================
-- 8. 资产表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_assets (
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
    CONSTRAINT fk_assets_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.sites(id),
    CONSTRAINT fk_assets_folder FOREIGN KEY (folder_id) REFERENCES CONTFUL_ENT.t_asset_folders(id)
);
CREATE INDEX idx_assets_site ON CONTFUL_ENT.t_assets(site_id);
CREATE INDEX idx_assets_folder ON CONTFUL_ENT.t_assets(folder_id);

-- =============================================================================
-- 9. API Token 表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_tokens (
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
    CONSTRAINT fk_tokens_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.sites(id)
);
CREATE INDEX idx_tokens_site ON CONTFUL_ENT.t_tokens(site_id);

-- =============================================================================
-- 10. 审计日志表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_audit_logs (
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
CREATE INDEX idx_audit_logs_category ON CONTFUL_ENT.t_audit_logs(category, created_time);
CREATE INDEX idx_audit_logs_level ON CONTFUL_ENT.t_audit_logs(level, created_time);
CREATE INDEX idx_audit_logs_user ON CONTFUL_ENT.t_audit_logs(user_id);

-- =============================================================================
-- 11. 系统角色
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_system_roles (
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
CREATE TABLE CONTFUL_ENT.t_system_user_roles (
    user_id VARCHAR2(36) NOT NULL,
    role_id VARCHAR2(36) NOT NULL,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_system_user_roles PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_sur_user FOREIGN KEY (user_id) REFERENCES CONTFUL_ENT.system_users(id),
    CONSTRAINT fk_sur_role FOREIGN KEY (role_id) REFERENCES CONTFUL_ENT.t_system_roles(id)
);

-- =============================================================================
-- 13. 系统配置
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_system_config (
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
CREATE TABLE CONTFUL_ENT.t_system_permission_groups (
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
CREATE TABLE CONTFUL_ENT.t_system_permissions (
    group_id VARCHAR2(36) NOT NULL,
    action VARCHAR2(100) NOT NULL,
    label VARCHAR2(255),
    label_en VARCHAR2(255),
    sort_order NUMBER DEFAULT 0,
    CONSTRAINT pk_permissions PRIMARY KEY (group_id, action),
    CONSTRAINT fk_perms_group FOREIGN KEY (group_id) REFERENCES CONTFUL_ENT.t_system_permission_groups(id)
);

-- =============================================================================
-- 16. Webhook 配置表
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_webhooks (
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
    CONSTRAINT fk_webhooks_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.sites(id)
);
CREATE INDEX idx_webhooks_site ON CONTFUL_ENT.t_webhooks(site_id);
CREATE INDEX idx_webhooks_active ON CONTFUL_ENT.t_webhooks(is_active);

-- =============================================================================
-- 17. Webhook 投递记录
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_webhook_deliveries (
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
    CONSTRAINT fk_wd_webhook FOREIGN KEY (webhook_id) REFERENCES CONTFUL_ENT.t_webhooks(id)
);
CREATE INDEX idx_wd_webhook ON CONTFUL_ENT.t_webhook_deliveries(webhook_id);
CREATE INDEX idx_wd_status ON CONTFUL_ENT.t_webhook_deliveries(status);
CREATE INDEX idx_wd_created ON CONTFUL_ENT.t_webhook_deliveries(created_time);

-- =============================================================================
-- 18. 审计导出任务（企业版）
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_audit_report_exports (
    id VARCHAR2(36) ,
    user_id VARCHAR2(36),
    status VARCHAR2(20) DEFAULT 'pending',
    formats CLOB DEFAULT '[]',
    filter CLOB,
    file_path VARCHAR2(500),
    file_size NUMBER,
    expires_time TIMESTAMP,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_audit_exports PRIMARY KEY (id)
);

-- =============================================================================
-- 19. 审计保留策略（企业版）
-- =============================================================================
CREATE TABLE CONTFUL_ENT.t_audit_retention_policies (
    id VARCHAR2(36) ,
    category VARCHAR2(50),
    level VARCHAR2(20),
    retention_days NUMBER NOT NULL DEFAULT 180,
    archive_enabled CHAR(1) DEFAULT '0',
    archive_storage VARCHAR2(500),
    is_active CHAR(1) DEFAULT '1',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_retention_policies PRIMARY KEY (id),
    CONSTRAINT uq_retention_clevel UNIQUE (category, level)
);

-- =============================================================================
-- UUID 主键 BEFORE INSERT 触发器
-- DM8 DEFAULT 不支持函数调用，改用触发器
-- =============================================================================
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_system_users_bi
BEFORE INSERT ON CONTFUL_ENT.system_users FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_sites_bi
BEFORE INSERT ON CONTFUL_ENT.sites FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_schemas_bi
BEFORE INSERT ON CONTFUL_ENT.schemas FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_fields_bi
BEFORE INSERT ON CONTFUL_ENT.fields FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_entries_bi
BEFORE INSERT ON CONTFUL_ENT.entries FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_entry_values_bi
BEFORE INSERT ON CONTFUL_ENT.t_entry_values FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_asset_folders_bi
BEFORE INSERT ON CONTFUL_ENT.t_asset_folders FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_assets_bi
BEFORE INSERT ON CONTFUL_ENT.t_assets FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_tokens_bi
BEFORE INSERT ON CONTFUL_ENT.t_tokens FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_audit_logs_bi
BEFORE INSERT ON CONTFUL_ENT.t_audit_logs FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_system_roles_bi
BEFORE INSERT ON CONTFUL_ENT.t_system_roles FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_perm_groups_bi
BEFORE INSERT ON CONTFUL_ENT.t_system_permission_groups FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_webhooks_bi
BEFORE INSERT ON CONTFUL_ENT.t_webhooks FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_wd_bi
BEFORE INSERT ON CONTFUL_ENT.t_webhook_deliveries FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_audit_exports_bi
BEFORE INSERT ON CONTFUL_ENT.t_audit_report_exports FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_retention_bi
BEFORE INSERT ON CONTFUL_ENT.t_audit_retention_policies FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/

-- =============================================================================
-- 初始化完成
-- =============================================================================
