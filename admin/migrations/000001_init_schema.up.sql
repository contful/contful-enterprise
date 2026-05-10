-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Migration 000001: Initial schema
-- =============================================================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

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
-- 0. 系统用户与角色（系统级）
-- =============================================================================

-- 系统用户表
CREATE TABLE system_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url TEXT,
    status user_status NOT NULL DEFAULT 'active',
    is_super_admin BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_system_users_email ON system_users(email);
CREATE INDEX idx_system_users_status ON system_users(status);

-- 系统角色表
CREATE TABLE system_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_system_roles_name ON system_roles(name);

-- 系统用户-角色关联表
CREATE TABLE system_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES system_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_site ON audit_logs(site_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_category ON audit_logs(category);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_sites_slug ON sites(slug);
CREATE INDEX idx_sites_active ON sites(is_active);
CREATE INDEX idx_sites_locale ON sites(locale);

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(site_id, slug)
);

CREATE INDEX idx_schemas_site ON schemas(site_id);
CREATE INDEX idx_schemas_slug ON schemas(slug);
CREATE INDEX idx_schemas_active ON schemas(is_active);
CREATE INDEX idx_schemas_kind ON schemas(kind);

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(schema_id, name)
);

CREATE INDEX idx_fields_schema ON fields(schema_id);
CREATE INDEX idx_fields_sort ON fields(schema_id, sort_order);

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
    published_at TIMESTAMPTZ,
    published_by UUID,
    relations JSONB NOT NULL DEFAULT '[]',
    seo_title VARCHAR(255),
    seo_description TEXT,
    seo_keywords TEXT[],
    sort_weight INT NOT NULL DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(schema_id, locale, id)
);

CREATE INDEX idx_entries_schema ON entries(schema_id);
CREATE INDEX idx_entries_site ON entries(site_id);
CREATE INDEX idx_entries_locale ON entries(locale);
CREATE INDEX idx_entries_status ON entries(status);
CREATE INDEX idx_entries_published ON entries(published_at DESC) WHERE status = 'published';
CREATE INDEX idx_entries_sort ON entries(schema_id, sort_weight);
CREATE INDEX idx_entries_created_by ON entries(created_by);
CREATE INDEX idx_entries_deleted ON entries(deleted_at) WHERE deleted_at IS NULL;

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    change_summary TEXT,
    UNIQUE(entry_id, version)
);

CREATE INDEX idx_entry_versions_entry ON entry_versions(entry_id);
CREATE INDEX idx_entry_versions_created ON entry_versions(created_at DESC);

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_asset_folders_site ON asset_folders(site_id);
CREATE INDEX idx_asset_folders_parent ON asset_folders(parent_id);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES sites(id),
    folder_id UUID REFERENCES asset_folders(id),
    uuid VARCHAR(36) NOT NULL UNIQUE,
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_assets_site ON assets(site_id);
CREATE INDEX idx_assets_folder ON assets(folder_id);
CREATE INDEX idx_assets_type ON assets(type);
CREATE INDEX idx_assets_extension ON assets(extension);
CREATE INDEX idx_assets_slug ON assets(slug);
CREATE INDEX idx_assets_hash ON assets(file_hash) WHERE file_hash IS NOT NULL;
CREATE INDEX idx_assets_created ON assets(created_at DESC);
CREATE INDEX idx_assets_download ON assets(download_count);
CREATE INDEX idx_assets_used ON assets(used_count);

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
    expires_at TIMESTAMPTZ,
    status token_status NOT NULL DEFAULT 'active',
    last_used_at TIMESTAMPTZ,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tokens_site ON tokens(site_id);
CREATE INDEX idx_tokens_hash ON tokens(token_hash);
CREATE INDEX idx_tokens_status ON tokens(status);
CREATE INDEX idx_tokens_expires ON tokens(expires_at) WHERE expires_at IS NOT NULL;
