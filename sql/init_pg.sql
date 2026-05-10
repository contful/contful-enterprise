-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Contful - Headless CMS 数据库初始化脚本（GORM AutoMigrate 兼容版）
-- 版本: v1.3.0
-- 数据库: PostgreSQL 18
-- 
-- 使用方式：
-- 1. 首次部署：执行此脚本创建基础表结构
-- 2. GORM AutoMigrate：自动同步 GORM 模型变更
-- 
-- 兼容 GORM AutoMigrate 注意事项：
-- - 表名、字段名、类型与 GORM 模型完全一致
-- - ENUM 类型需预先创建（GORM 不会自动创建）
-- - 触发器、函数、COMMENT、复杂索引需手动执行（见下文"高级特性"部分）
-- =============================================================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- 清理已有对象（支持重复执行）
-- =============================================================================

-- 删除触发器（高级特性，在 advanced_features.sql 中创建）
DROP TRIGGER IF EXISTS update_system_users_updated_at ON system_users;
DROP TRIGGER IF EXISTS update_system_roles_updated_at ON system_roles;
DROP TRIGGER IF EXISTS update_sites_updated_at ON sites;
DROP TRIGGER IF EXISTS update_schemas_updated_at ON schemas;
DROP TRIGGER IF EXISTS update_fields_updated_at ON fields;
DROP TRIGGER IF EXISTS update_entries_updated_at ON entries;
DROP TRIGGER IF EXISTS update_entry_values_updated_at ON entry_values;
DROP TRIGGER IF EXISTS update_entry_versions_updated_at ON entry_versions;
DROP TRIGGER IF EXISTS update_assets_updated_at ON assets;
DROP TRIGGER IF EXISTS update_tokens_updated_at ON tokens;
DROP TRIGGER IF EXISTS prevent_audit_logs_update ON audit_logs;

-- 删除函数
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS prevent_audit_log_update();

-- 删除表（按依赖顺序，CASCADE 自动处理外键）
DROP TABLE IF EXISTS tokens CASCADE;
DROP TABLE IF EXISTS assets CASCADE;
DROP TABLE IF EXISTS asset_folders CASCADE;
DROP TABLE IF EXISTS entry_versions CASCADE;
DROP TABLE IF EXISTS entry_values CASCADE;
DROP TABLE IF EXISTS entries CASCADE;
DROP TABLE IF EXISTS fields CASCADE;
DROP TABLE IF EXISTS schemas CASCADE;
DROP TABLE IF EXISTS sites CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS system_user_roles CASCADE;
DROP TABLE IF EXISTS system_roles CASCADE;
DROP TABLE IF EXISTS system_users CASCADE;
DROP TABLE IF EXISTS system_config CASCADE;

-- 删除 ENUM 类型（需先删除依赖的表）
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS entry_status;
DROP TYPE IF EXISTS content_type_kind;
DROP TYPE IF EXISTS field_type;
DROP TYPE IF EXISTS asset_type;
DROP TYPE IF EXISTS asset_visibility;
DROP TYPE IF EXISTS asset_status;
DROP TYPE IF EXISTS token_status;
DROP TYPE IF EXISTS audit_level;
DROP TYPE IF EXISTS audit_type;

-- =============================================================================
-- ENUM 类型（必须在表创建之前定义）
-- =============================================================================

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE entry_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE content_type_kind AS ENUM ('collection', 'single');
CREATE TYPE field_type AS ENUM (
    'text', 'rich_text', 'number', 'boolean', 'date', 'datetime',
    'email', 'url', 'json', 'media', 'relation', 'enum', 'password'
);
CREATE TYPE asset_type AS ENUM ('image', 'video', 'audio', 'document', 'file');
CREATE TYPE asset_visibility AS ENUM ('public', 'private');
CREATE TYPE asset_status AS ENUM ('active', 'inactive', 'deleted');
CREATE TYPE token_status AS ENUM ('active', 'expired', 'revoked');
CREATE TYPE audit_level AS ENUM ('debug', 'info', 'warn', 'error');
CREATE TYPE audit_type AS ENUM ('auth', 'content', 'media', 'settings', 'user', 'system');

-- =============================================================================
-- 0. 全局用户与角色（系统级）
-- =============================================================================

-- 全局用户表
CREATE TABLE system_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url TEXT,
    status user_status NOT NULL DEFAULT 'active',
    is_super_admin BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_time TIMESTAMPTZ,
    last_login_ip INET,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- 部分唯一索引：仅对未删除的记录强制 email 唯一，软删除后允许复用 email
CREATE UNIQUE INDEX idx_system_users_email_active ON system_users(email) WHERE deleted_time IS NULL;
CREATE INDEX idx_system_users_email ON system_users(email);
CREATE INDEX idx_system_users_status ON system_users(status);
CREATE INDEX idx_system_users_deleted_time ON system_users(deleted_time) WHERE deleted_time IS NOT NULL;

-- 全局角色表
CREATE TABLE system_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

-- 部分唯一索引：仅对未删除的记录强制 name 唯一
CREATE UNIQUE INDEX idx_system_roles_name_active ON system_roles(name) WHERE deleted_time IS NULL;
CREATE INDEX idx_system_roles_name ON system_roles(name);

-- 全局用户-角色关联表
CREATE TABLE system_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES system_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_system_user_roles_user ON system_user_roles(user_id);
CREATE INDEX idx_system_user_roles_role ON system_user_roles(role_id);

-- =============================================================================
-- 0.5 系统配置表
-- =============================================================================

CREATE TABLE system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    value_type VARCHAR(20) NOT NULL DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_system_config_key ON system_config(config_key);

-- =============================================================================
-- 0.7 审计日志表
-- =============================================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID,
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    level audit_level NOT NULL DEFAULT 'info',
    category audit_type NOT NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_site ON audit_logs(site_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_category ON audit_logs(category);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_time DESC);

-- =============================================================================
-- 1. 站点表
-- =============================================================================

CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    site_url TEXT,
    locale VARCHAR(20) NOT NULL DEFAULT 'zh-CN',
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai',
    seo_title VARCHAR(255),
    seo_description TEXT,
    seo_keywords JSONB DEFAULT '[]',
    settings JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

CREATE INDEX idx_sites_slug ON sites(slug);
CREATE INDEX idx_sites_active ON sites(is_active);
CREATE INDEX idx_sites_locale ON sites(locale);

-- =============================================================================
-- 2. 插件表
-- =============================================================================

-- =============================================================================
-- 3. 站点用户与角色（站点级）
-- =============================================================================

-- 站点角色表
-- =============================================================================
-- 4. 内容模型层（使用 schemas 作为表名）
-- =============================================================================

CREATE TABLE schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    kind content_type_kind NOT NULL DEFAULT 'collection',
    display_config JSONB NOT NULL DEFAULT '{}',
    api_config JSONB NOT NULL DEFAULT '{"publicRead":false,"publicWrite":false}',
    preview_config JSONB NOT NULL DEFAULT '{}',
    versioning_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    draft_autosave_interval INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

-- 部分唯一索引：仅对未删除的记录强制 slug 唯一
CREATE UNIQUE INDEX idx_schemas_slug_active ON schemas(site_id, slug) WHERE deleted_time IS NULL;
CREATE INDEX idx_schemas_site ON schemas(site_id);
CREATE INDEX idx_schemas_slug ON schemas(slug);
CREATE INDEX idx_schemas_active ON schemas(is_active);
CREATE INDEX idx_schemas_kind ON schemas(kind);
CREATE INDEX idx_schemas_deleted_time ON schemas(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TABLE fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    label VARCHAR(200) NOT NULL,
    description TEXT,
    field_type field_type NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    validation JSONB NOT NULL DEFAULT '{}',
    display JSONB NOT NULL DEFAULT '{}',
    default_value JSONB,
    sort_order INT NOT NULL DEFAULT 0,
    conditional_display JSONB,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

-- 部分唯一索引：仅对未删除的记录强制 name 唯一
CREATE UNIQUE INDEX idx_fields_name_active ON fields(schema_id, name) WHERE deleted_time IS NULL;
CREATE INDEX idx_fields_schema ON fields(schema_id);
CREATE INDEX idx_fields_sort ON fields(schema_id, sort_order);
CREATE INDEX idx_fields_deleted_time ON fields(deleted_time) WHERE deleted_time IS NOT NULL;

-- =============================================================================
-- 5. 内容条目层
-- =============================================================================

CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES schemas(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES sites(id),
    locale VARCHAR(20) NOT NULL DEFAULT 'zh-CN',
    status entry_status NOT NULL DEFAULT 'draft',
    version INT NOT NULL DEFAULT 1,
    version_history JSONB,
    published_time TIMESTAMPTZ,
    published_by UUID,
    relations JSONB NOT NULL DEFAULT '[]',
    seo_title VARCHAR(255),
    seo_description TEXT,
    seo_keywords TEXT[],
    sort_weight INT NOT NULL DEFAULT 0,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    UNIQUE(schema_id, locale, id)
);

CREATE INDEX idx_entries_schema ON entries(schema_id);
CREATE INDEX idx_entries_site ON entries(site_id);
CREATE INDEX idx_entries_locale ON entries(locale);
CREATE INDEX idx_entries_status ON entries(status);
CREATE INDEX idx_entries_published ON entries(published_time DESC) WHERE status = 'published';
CREATE INDEX idx_entries_sort ON entries(schema_id, sort_weight);
CREATE INDEX idx_entries_created_by ON entries(created_by);
CREATE INDEX idx_entries_deleted ON entries(deleted_time) WHERE deleted_time IS NULL;

CREATE TABLE entry_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    value JSONB NOT NULL,
    text_value TEXT,
    number_value NUMERIC,
    bool_value BOOLEAN,
    date_value DATE,
    datetime_value TIMESTAMPTZ,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(entry_id, field_id)
);

CREATE INDEX idx_entry_values_entry ON entry_values(entry_id);
CREATE INDEX idx_entry_values_field ON entry_values(field_id);

CREATE TABLE entry_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    version INT NOT NULL,
    values_snapshot JSONB NOT NULL,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    change_summary TEXT,
    UNIQUE(entry_id, version)
);

CREATE INDEX idx_entry_versions_entry ON entry_versions(entry_id);
CREATE INDEX idx_entry_versions_created ON entry_versions(created_time DESC);

-- =============================================================================
-- 6. 媒体资产层
-- =============================================================================

CREATE TABLE asset_folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),
    parent_id UUID REFERENCES asset_folders(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

CREATE INDEX idx_asset_folders_site ON asset_folders(site_id);
CREATE INDEX idx_asset_folders_parent ON asset_folders(parent_id);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),
    folder_id UUID REFERENCES asset_folders(id),
    uuid VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    type asset_type NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    extension VARCHAR(20) NOT NULL,
    size BIGINT NOT NULL,
    width INT,
    height INT,
    duration DOUBLE PRECISION,
    path VARCHAR(500) NOT NULL,
    url VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    alt TEXT,
    title VARCHAR(255),
    caption TEXT,
    alt_text TEXT,
    description TEXT,
    tags TEXT[],
    metadata JSONB NOT NULL DEFAULT '{}',
    visibility asset_visibility NOT NULL DEFAULT 'private',
    file_hash VARCHAR(64),
    disk VARCHAR(50) NOT NULL DEFAULT 'local',
    download_count INT NOT NULL DEFAULT 0,
    used_count INT NOT NULL DEFAULT 0,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

-- 部分唯一索引：仅对未删除的记录强制 uuid 唯一
CREATE UNIQUE INDEX idx_assets_uuid_active ON assets(uuid) WHERE deleted_time IS NULL;
CREATE INDEX idx_assets_site ON assets(site_id);
CREATE INDEX idx_assets_folder ON assets(folder_id);
CREATE INDEX idx_assets_type ON assets(type);
CREATE INDEX idx_assets_extension ON assets(extension);
CREATE INDEX idx_assets_slug ON assets(slug);
CREATE INDEX idx_assets_hash ON assets(file_hash) WHERE file_hash IS NOT NULL;
CREATE INDEX idx_assets_created ON assets(created_time DESC);
CREATE INDEX idx_assets_download ON assets(download_count);
CREATE INDEX idx_assets_used ON assets(used_count);
CREATE INDEX idx_assets_deleted_time ON assets(deleted_time) WHERE deleted_time IS NOT NULL;

-- =============================================================================
-- 7. API Token 表
-- =============================================================================

CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    token_prefix VARCHAR(10) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    permissions JSONB NOT NULL DEFAULT '{}',
    rate_limits JSONB NOT NULL DEFAULT '{"requests_per_minute": 60, "requests_per_day": 10000}',
    usage JSONB NOT NULL DEFAULT '{"request_count": 0}',
    expires_time TIMESTAMPTZ,
    status token_status NOT NULL DEFAULT 'active',
    last_used_time TIMESTAMPTZ,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

CREATE INDEX idx_tokens_site ON tokens(site_id);
CREATE INDEX idx_tokens_hash ON tokens(token_hash);
CREATE INDEX idx_tokens_status ON tokens(status);
CREATE INDEX idx_tokens_expires ON tokens(expires_time) WHERE expires_time IS NOT NULL;

-- =============================================================================
-- 8. Webhook 表
-- =============================================================================

-- =============================================================================
-- 初始化数据
-- =============================================================================

-- 默认站点
INSERT INTO sites (id, name, slug, description, locale, timezone, is_active, settings) VALUES
    ('00000000-0000-0000-0000-000000000001', '默认站点', 'default', '系统默认站点', 'zh-CN', 'Asia/Shanghai', TRUE, '{}');

-- 系统角色
INSERT INTO system_roles (id, name, description, is_system, permissions) VALUES
    ('00000000-0000-0000-0000-000000000101', 'Super Admin', '超级管理员，拥有所有权限', TRUE, '["*"]'),
    ('00000000-0000-0000-0000-000000000102', 'Plugin Manager', '插件管理员', TRUE, '["plugins:read", "plugins:write", "plugins:install", "plugins:uninstall"]'),
    ('00000000-0000-0000-0000-000000000103', 'Auditor', '审计员，只读访问', TRUE, '["audit:read"]');

-- 系统用户
INSERT INTO system_users (id, email, password_hash, nickname, status, is_super_admin) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@contful.com', '$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu', 'Administrator', 'active', TRUE);

-- 关联 admin 用户与 Super Admin 角色
INSERT INTO system_user_roles (user_id, role_id) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000101');

-- 初始化默认配置
INSERT INTO system_config (config_key, config_value, value_type, description, is_public) VALUES
    ('password_expire_days', '90', 'number', '密码有效期（天），0 表示永不过期', FALSE),
    ('site_name', 'Contful', 'string', '系统名称', TRUE),
    ('logo_url', '', 'string', '系统 Logo 图片地址', TRUE),
    ('login_background_url', '', 'string', '登录页背景图片地址', TRUE),
    ('password_min_length', '8', 'number', '密码最小长度', FALSE),
    ('password_require_uppercase', 'true', 'boolean', '密码必须包含大写字母', FALSE),
    ('password_require_lowercase', 'true', 'boolean', '密码必须包含小写字母', FALSE),
    ('password_require_number', 'true', 'boolean', '密码必须包含数字', FALSE)
ON CONFLICT (config_key) DO NOTHING;
