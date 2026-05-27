-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Contful - Headless CMS 数据库初始化脚本（DDL + 种子数据）
-- 版本: v2.0.0
-- 数据库: PostgreSQL 14+
--
-- 使用方式：
--   psql -h <host> -U <user> -d <db> -f init_pg.sql
--
-- 注意：此脚本为完整重建脚本（DDL + 默认数据），会删除所有已有对象，仅用于全新部署或开发环境重置。
-- 幂等设计：DDL 使用 IF NOT EXISTS / 种子数据使用 INSERT WHERE NOT EXISTS + ON CONFLICT DO NOTHING
-- =============================================================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- 清理已有对象（支持重复执行）
-- =============================================================================

-- 删除触发器
DROP TRIGGER IF EXISTS update_system_users_updated_time ON contful_system_users;
DROP TRIGGER IF EXISTS update_system_roles_updated_time ON contful_system_roles;
DROP TRIGGER IF EXISTS update_system_config_updated_time ON contful_system_config;
DROP TRIGGER IF EXISTS update_sites_updated_time ON contful_sites;
DROP TRIGGER IF EXISTS update_schemas_updated_time ON contful_schemas;
DROP TRIGGER IF EXISTS update_fields_updated_time ON contful_fields;
DROP TRIGGER IF EXISTS update_entries_updated_time ON contful_entries;
DROP TRIGGER IF EXISTS update_entry_values_updated_time ON contful_entry_values;
DROP TRIGGER IF EXISTS update_asset_folders_updated_time ON contful_asset_folders;
DROP TRIGGER IF EXISTS update_assets_updated_time ON contful_assets;
DROP TRIGGER IF EXISTS update_tokens_updated_time ON contful_tokens;
DROP TRIGGER IF EXISTS prevent_audit_logs_update ON contful_audit_logs;

-- 删除函数
DROP FUNCTION IF EXISTS update_updated_time_column();
DROP FUNCTION IF EXISTS prevent_audit_log_update();

-- 删除表（按依赖顺序，CASCADE 自动处理外键）
DROP TABLE IF EXISTS contful_tokens CASCADE;
DROP TABLE IF EXISTS contful_assets CASCADE;
DROP TABLE IF EXISTS contful_asset_folders CASCADE;
DROP TABLE IF EXISTS contful_entry_versions CASCADE;
DROP TABLE IF EXISTS contful_entry_values CASCADE;
DROP TABLE IF EXISTS contful_entries CASCADE;
DROP TABLE IF EXISTS contful_fields CASCADE;
DROP TABLE IF EXISTS contful_schemas CASCADE;
DROP TABLE IF EXISTS contful_sites CASCADE;
DROP TABLE IF EXISTS contful_audit_logs CASCADE;
DROP TABLE IF EXISTS contful_system_user_roles CASCADE;
DROP TABLE IF EXISTS contful_system_roles CASCADE;
DROP TABLE IF EXISTS contful_system_users CASCADE;
DROP TABLE IF EXISTS contful_system_permissions CASCADE;
DROP TABLE IF EXISTS contful_system_permission_groups CASCADE;
DROP TABLE IF EXISTS contful_system_config CASCADE;

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
-- ENUM 类型（必须在表创建之前定义）[PG-ONLY: 其他数据库使用 VARCHAR + CHECK 约束]
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

-- ENUM 类型注释
COMMENT ON TYPE user_status IS '用户账户状态：active=正常、inactive=未激活、suspended=已停用';
COMMENT ON TYPE entry_status IS '内容条目状态：draft=草稿、published=已发布、archived=已归档';
COMMENT ON TYPE content_type_kind IS '内容模型类型：collection=集合（如文章列表）、single=单页（如关于页面）';
COMMENT ON TYPE field_type IS '字段类型：text=文本、rich_text=富文本、number=数字、boolean=布尔、date=日期、datetime=时间、email=邮箱、url=链接、json=JSON、media=媒体、relation=关联、enum=枚举、password=密码';
COMMENT ON TYPE asset_type IS '资源类型：image=图片、video=视频、audio=音频、document=文档、file=文件';
COMMENT ON TYPE asset_visibility IS '资源可见性：public=公开、private=私有';
COMMENT ON TYPE asset_status IS '资源状态：active=正常、inactive=禁用、deleted=已删除';
COMMENT ON TYPE token_status IS 'API Token 状态：active=有效、expired=已过期、revoked=已撤销';
COMMENT ON TYPE audit_level IS '审计日志级别：debug=调试、info=信息、warn=警告、error=错误';
COMMENT ON TYPE audit_type IS '审计日志类别：auth=认证、content=内容、media=媒体、settings=设置、user=用户、system=系统';

-- =============================================================================
-- 触发器函数 [PG-ONLY: MySQL 使用 ON UPDATE CURRENT_TIMESTAMP 或触发器]
-- =============================================================================

CREATE OR REPLACE FUNCTION update_updated_time_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_time = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_time_column() IS '自动更新 updated_time 时间戳的触发器函数';

CREATE OR REPLACE FUNCTION prevent_audit_log_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION '审计日志禁止更新，仅允许 INSERT 和软删除';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION prevent_audit_log_update() IS '审计日志防篡改触发器函数';

-- =============================================================================
-- 系统用户表
-- =============================================================================

CREATE TABLE contful_system_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url TEXT,
    phone VARCHAR(20),
    department VARCHAR(100),
    status user_status NOT NULL DEFAULT 'active', -- [PG-ONLY: ENUM → MySQL VARCHAR(20) + CHECK]
    is_super_admin BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    totp_secret VARCHAR(512),                       -- AES-256-GCM 加密的 TOTP 密钥
    recovery_codes TEXT,                             -- AES-256-GCM 加密的恢复码 (JSON 数组)
    password_changed_time TIMESTAMPTZ,               -- 密码最后修改时间
    data_signature VARCHAR(256) NOT NULL DEFAULT '',   -- 防篡改签名（HMAC-SHA256 hex）
    last_login_time TIMESTAMPTZ, -- [PG-ONLY: → MySQL DATETIME]
    last_login_ip INET,           -- [PG-ONLY: → MySQL VARCHAR(45)]
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- 部分唯一索引：仅对未删除的记录强制 email 唯一 [PG-ONLY: MySQL 不支持部分索引]
CREATE UNIQUE INDEX idx_system_users_email_active ON contful_system_users(email) WHERE deleted_time IS NULL;
CREATE INDEX idx_system_users_email ON contful_system_users(email);
CREATE INDEX idx_system_users_status ON contful_system_users(status);
CREATE INDEX idx_system_users_deleted_time ON contful_system_users(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_system_users_updated_time
    BEFORE UPDATE ON contful_system_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_system_users IS '系统用户表：存放所有用户账户，用户全局管理，权限通过 contful_system_user_roles 关联';
COMMENT ON COLUMN contful_system_users.id IS '用户唯一标识符(UUID)';
COMMENT ON COLUMN contful_system_users.email IS '用户邮箱（未删除时全局唯一）';
COMMENT ON COLUMN contful_system_users.password_hash IS 'bcrypt 加密的密码哈希';
COMMENT ON COLUMN contful_system_users.nickname IS '用户昵称';
COMMENT ON COLUMN contful_system_users.avatar_url IS '头像 URL';
COMMENT ON COLUMN contful_system_users.phone IS '用户手机号';
COMMENT ON COLUMN contful_system_users.department IS '用户所属部门';
COMMENT ON COLUMN contful_system_users.status IS '账户状态：active/inactive/suspended';
COMMENT ON COLUMN contful_system_users.is_super_admin IS '是否超级管理员（拥有全部权限，不受站点限制）';
COMMENT ON COLUMN contful_system_users.last_login_time IS '最后登录时间';
COMMENT ON COLUMN contful_system_users.last_login_ip IS '最后登录 IP 地址';
COMMENT ON COLUMN contful_system_users.created_time IS '创建时间';
COMMENT ON COLUMN contful_system_users.updated_time IS '更新时间';
COMMENT ON COLUMN contful_system_users.deleted_time IS '软删除时间（非空表示已删除）';

-- =============================================================================
-- 系统角色表
-- =============================================================================

CREATE TABLE contful_system_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]', -- [PG-ONLY: → MySQL JSON; 达梦 CLOB]
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

-- 部分唯一索引：仅对未删除的记录强制 name 唯一
CREATE UNIQUE INDEX idx_system_roles_name_active ON contful_system_roles(name) WHERE deleted_time IS NULL;
CREATE INDEX idx_system_roles_name ON contful_system_roles(name);
CREATE INDEX idx_system_roles_deleted_time ON contful_system_roles(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_system_roles_updated_time
    BEFORE UPDATE ON contful_system_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_system_roles IS '系统角色表：系统级角色定义，系统角色不可删除';
COMMENT ON COLUMN contful_system_roles.id IS '角色唯一标识符';
COMMENT ON COLUMN contful_system_roles.name IS '角色名称（未删除时全局唯一）';
COMMENT ON COLUMN contful_system_roles.description IS '角色描述';
COMMENT ON COLUMN contful_system_roles.is_system IS '是否系统角色（系统角色不可删除）';
COMMENT ON COLUMN contful_system_roles.permissions IS '权限列表 JSON：["user:read", "user:write", ...]';
COMMENT ON COLUMN contful_system_roles.created_time IS '创建时间';
COMMENT ON COLUMN contful_system_roles.updated_time IS '更新时间';
COMMENT ON COLUMN contful_system_roles.deleted_time IS '软删除时间';

-- =============================================================================
-- 系统用户-角色关联表
-- =============================================================================

CREATE TABLE contful_system_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES contful_system_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES contful_system_roles(id) ON DELETE CASCADE,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_system_user_roles_user ON contful_system_user_roles(user_id);
CREATE INDEX idx_system_user_roles_role ON contful_system_user_roles(role_id);

COMMENT ON TABLE contful_system_user_roles IS '系统用户-角色关联表：多对多关系，一个用户可拥有多个系统角色';
COMMENT ON COLUMN contful_system_user_roles.id IS '关联记录唯一标识符';
COMMENT ON COLUMN contful_system_user_roles.user_id IS '用户 ID';
COMMENT ON COLUMN contful_system_user_roles.role_id IS '系统角色 ID';
COMMENT ON COLUMN contful_system_user_roles.created_time IS '创建时间';

-- =============================================================================
-- 系统权限管理表
-- =============================================================================

-- 权限分组
CREATE TABLE contful_system_permission_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_key VARCHAR(50) UNIQUE NOT NULL,
    label VARCHAR(100) NOT NULL,
    label_en VARCHAR(100),
    sort_order INT NOT NULL DEFAULT 0,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_perm_group_key ON contful_system_permission_groups(group_key);
CREATE INDEX idx_perm_group_sort ON contful_system_permission_groups(sort_order);

COMMENT ON TABLE contful_system_permission_groups IS '系统权限分组表';
COMMENT ON COLUMN contful_system_permission_groups.group_key IS '分组键（如 users、contful_sites、contful_tokens 等）';
COMMENT ON COLUMN contful_system_permission_groups.label IS '分组中文标签';
COMMENT ON COLUMN contful_system_permission_groups.label_en IS '分组英文标签';
COMMENT ON COLUMN contful_system_permission_groups.sort_order IS '排序权重';

-- 权限项
CREATE TABLE contful_system_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES contful_system_permission_groups(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    label VARCHAR(100) NOT NULL,
    label_en VARCHAR(100),
    sort_order INT NOT NULL DEFAULT 0,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(group_id, action)
);

CREATE INDEX idx_permissions_group ON contful_system_permissions(group_id);
CREATE INDEX idx_permissions_sort ON contful_system_permissions(sort_order);

COMMENT ON TABLE contful_system_permissions IS '系统权限项表';
COMMENT ON COLUMN contful_system_permissions.group_id IS '所属权限分组 ID';
COMMENT ON COLUMN contful_system_permissions.action IS '操作键（如 read、write、delete 等）';
COMMENT ON COLUMN contful_system_permissions.label IS '权限中文标签';
COMMENT ON COLUMN contful_system_permissions.label_en IS '权限英文标签';
COMMENT ON COLUMN contful_system_permissions.sort_order IS '排序权重';

-- =============================================================================
-- 系统配置表
-- =============================================================================

CREATE TABLE contful_system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    value_type VARCHAR(20) NOT NULL DEFAULT 'string',
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_system_config_key ON contful_system_config(config_key);

CREATE TRIGGER update_system_config_updated_time
    BEFORE UPDATE ON contful_system_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_system_config IS '系统配置表：存储全局配置项（密码策略、系统名称、Logo 等）';
COMMENT ON COLUMN contful_system_config.id IS '配置项唯一标识符';
COMMENT ON COLUMN contful_system_config.config_key IS '配置键名（唯一，如 password_expire_days、site_name、logo_url）';
COMMENT ON COLUMN contful_system_config.config_value IS '配置值（根据 value_type 解析）';
COMMENT ON COLUMN contful_system_config.value_type IS '值类型：string/number/boolean/json';
COMMENT ON COLUMN contful_system_config.description IS '配置项描述';
COMMENT ON COLUMN contful_system_config.is_public IS '是否公开（公开配置可未授权访问）';
COMMENT ON COLUMN contful_system_config.is_system IS '是否系统配置（系统配置不可删除，自定义配置可删除）';
COMMENT ON COLUMN contful_system_config.created_time IS '创建时间';
COMMENT ON COLUMN contful_system_config.updated_time IS '更新时间';

-- =============================================================================
-- 审计日志表
-- =============================================================================

CREATE TABLE contful_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID,
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    level audit_level NOT NULL DEFAULT 'info',
    category audit_type NOT NULL,
    details TEXT,
    ip_address INET,
    user_agent TEXT,
    data_signature VARCHAR(128) NOT NULL DEFAULT '',
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_site ON contful_audit_logs(site_id);
CREATE INDEX idx_audit_logs_user ON contful_audit_logs(user_id);
CREATE INDEX idx_audit_logs_category ON contful_audit_logs(category);
CREATE INDEX idx_audit_logs_created ON contful_audit_logs(created_time DESC);
CREATE INDEX idx_audit_logs_site_user_time ON contful_audit_logs(site_id, user_id, created_time DESC);
CREATE INDEX idx_audit_logs_category_time ON contful_audit_logs(category, created_time DESC);

-- 审计日志防篡改触发器
CREATE TRIGGER prevent_audit_logs_update
    BEFORE UPDATE ON contful_audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_update();

COMMENT ON TABLE contful_audit_logs IS '审计日志表：记录所有操作行为，不可修改';
COMMENT ON COLUMN contful_audit_logs.id IS '日志唯一标识符';
COMMENT ON COLUMN contful_audit_logs.site_id IS '关联站点 ID（全局操作时为空）';
COMMENT ON COLUMN contful_audit_logs.user_id IS '操作用户 ID';
COMMENT ON COLUMN contful_audit_logs.action IS '操作动作名称';
COMMENT ON COLUMN contful_audit_logs.resource_type IS '资源类型';
COMMENT ON COLUMN contful_audit_logs.resource_id IS '资源 ID';
COMMENT ON COLUMN contful_audit_logs.level IS '日志级别：debug/info/warn/error';
COMMENT ON COLUMN contful_audit_logs.category IS '操作类别：auth/content/media/settings/user/system';
COMMENT ON COLUMN contful_audit_logs.details IS '操作详情';
COMMENT ON COLUMN contful_audit_logs.ip_address IS '客户端 IP';
COMMENT ON COLUMN contful_audit_logs.user_agent IS '客户端 User-Agent';
COMMENT ON COLUMN contful_audit_logs.data_signature IS '审计日志完整性签名，防篡改';
COMMENT ON COLUMN contful_audit_logs.created_time IS '日志创建时间';

-- =============================================================================
-- 站点表
-- =============================================================================

CREATE TABLE contful_sites (
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

CREATE INDEX idx_sites_slug ON contful_sites(slug);
CREATE INDEX idx_sites_active ON contful_sites(is_active);
CREATE INDEX idx_sites_locale ON contful_sites(locale);

CREATE TRIGGER update_sites_updated_time
    BEFORE UPDATE ON contful_sites
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_sites IS '站点表：每个站点是一个独立的 CMS 实例';
COMMENT ON COLUMN contful_sites.id IS '站点唯一标识符';
COMMENT ON COLUMN contful_sites.name IS '站点名称';
COMMENT ON COLUMN contful_sites.slug IS '站点别名（URL 中使用）';
COMMENT ON COLUMN contful_sites.description IS '站点描述';
COMMENT ON COLUMN contful_sites.site_url IS '站点访问地址（前端展示端点）';
COMMENT ON COLUMN contful_sites.locale IS '站点默认语言（如 zh-CN、en-US）';
COMMENT ON COLUMN contful_sites.timezone IS '站点时区（如 Asia/Shanghai、America/New_York）';
COMMENT ON COLUMN contful_sites.seo_title IS '站点 SEO 标题';
COMMENT ON COLUMN contful_sites.seo_description IS '站点 SEO 描述';
COMMENT ON COLUMN contful_sites.seo_keywords IS '站点 SEO 关键词';
COMMENT ON COLUMN contful_sites.settings IS '动态配置 JSON（主题、第三方集成等）';
COMMENT ON COLUMN contful_sites.is_active IS '是否激活';
COMMENT ON COLUMN contful_sites.created_by IS '创建者用户 ID';
COMMENT ON COLUMN contful_sites.created_time IS '创建时间';
COMMENT ON COLUMN contful_sites.updated_time IS '更新时间';
COMMENT ON COLUMN contful_sites.deleted_time IS '软删除时间';

-- =============================================================================
-- 内容模型表（contful_schemas）
-- =============================================================================

CREATE TABLE contful_schemas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES contful_sites(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    kind content_type_kind NOT NULL DEFAULT 'collection',
    versioning_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    draft_autosave_interval INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    data_signature VARCHAR(256) NOT NULL DEFAULT '',   -- 防篡改签名（HMAC-SHA256 hex）
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

-- 部分唯一索引：仅对未删除的记录强制 slug 唯一
CREATE UNIQUE INDEX idx_schemas_slug_active ON contful_schemas(site_id, slug) WHERE deleted_time IS NULL;
CREATE INDEX idx_schemas_site ON contful_schemas(site_id);
CREATE INDEX idx_schemas_slug ON contful_schemas(slug);
CREATE INDEX idx_schemas_active ON contful_schemas(is_active);
CREATE INDEX idx_schemas_kind ON contful_schemas(kind);
CREATE INDEX idx_schemas_deleted_time ON contful_schemas(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_schemas_updated_time
    BEFORE UPDATE ON contful_schemas
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_schemas IS '内容模型表：定义内容的结构（类似 WordPress 的自定义内容类型）';
COMMENT ON COLUMN contful_schemas.id IS '内容模型唯一标识符';
COMMENT ON COLUMN contful_schemas.site_id IS '所属站点';
COMMENT ON COLUMN contful_schemas.name IS '模型名称';
COMMENT ON COLUMN contful_schemas.slug IS '模型别名（API 中使用）';
COMMENT ON COLUMN contful_schemas.description IS '模型描述';
COMMENT ON COLUMN contful_schemas.kind IS '类型：collection=集合、single=单页';
COMMENT ON COLUMN contful_schemas.versioning_enabled IS '是否启用版本历史';
COMMENT ON COLUMN contful_schemas.draft_autosave_interval IS '草稿自动保存间隔（秒）';
COMMENT ON COLUMN contful_schemas.is_active IS '是否启用';
COMMENT ON COLUMN contful_schemas.sort_order IS '排序权重';
COMMENT ON COLUMN contful_schemas.created_by IS '创建者';
COMMENT ON COLUMN contful_schemas.created_time IS '创建时间';
COMMENT ON COLUMN contful_schemas.updated_time IS '更新时间';
COMMENT ON COLUMN contful_schemas.deleted_time IS '软删除时间';

-- =============================================================================
-- 字段表
-- =============================================================================

CREATE TABLE contful_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES contful_schemas(id) ON DELETE CASCADE,
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
CREATE UNIQUE INDEX idx_fields_name_active ON contful_fields(schema_id, name) WHERE deleted_time IS NULL;
CREATE INDEX idx_fields_schema ON contful_fields(schema_id);
CREATE INDEX idx_fields_sort ON contful_fields(schema_id, sort_order);
CREATE INDEX idx_fields_deleted_time ON contful_fields(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_fields_updated_time
    BEFORE UPDATE ON contful_fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_fields IS '字段表：内容模型的字段定义';
COMMENT ON COLUMN contful_fields.id IS '字段唯一标识符';
COMMENT ON COLUMN contful_fields.schema_id IS '所属内容模型';
COMMENT ON COLUMN contful_fields.name IS '字段名（代码中使用）';
COMMENT ON COLUMN contful_fields.label IS '字段显示标签';
COMMENT ON COLUMN contful_fields.description IS '字段描述';
COMMENT ON COLUMN contful_fields.field_type IS '字段类型：text/rich_text/number/boolean/date/datetime/email/url/json/media/relation/enum/password';
COMMENT ON COLUMN contful_fields.config IS '字段配置 JSON：maxLength/minValue/options 等';
COMMENT ON COLUMN contful_fields.validation IS '验证规则 JSON：required/pattern/min/max 等';
COMMENT ON COLUMN contful_fields.display IS '显示配置 JSON：placeholder/readOnly 等';
COMMENT ON COLUMN contful_fields.default_value IS '默认值 JSON';
COMMENT ON COLUMN contful_fields.sort_order IS '排序权重';
COMMENT ON COLUMN contful_fields.conditional_display IS '条件显示 JSON';
COMMENT ON COLUMN contful_fields.created_time IS '创建时间';
COMMENT ON COLUMN contful_fields.updated_time IS '更新时间';
COMMENT ON COLUMN contful_fields.deleted_time IS '软删除时间';

-- =============================================================================
-- 内容条目表
-- =============================================================================

CREATE TABLE contful_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schema_id UUID NOT NULL REFERENCES contful_schemas(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES contful_sites(id),
    locale VARCHAR(20) NOT NULL DEFAULT 'zh-CN',
    status entry_status NOT NULL DEFAULT 'draft',
    version INT NOT NULL DEFAULT 1,
    version_history JSONB,
    published_time TIMESTAMPTZ,
    published_by UUID,
    scheduled_publish_time TIMESTAMPTZ,
    scheduled_unpublish_time TIMESTAMPTZ,
    relations JSONB NOT NULL DEFAULT '[]',
    seo_title VARCHAR(255),
    seo_description TEXT,
    seo_keywords TEXT[],
    sort_weight INT NOT NULL DEFAULT 0,
    data_signature JSONB,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    UNIQUE(schema_id, locale, id)
);

CREATE INDEX idx_entries_schema ON contful_entries(schema_id);
CREATE INDEX idx_entries_site ON contful_entries(site_id);
CREATE INDEX idx_entries_locale ON contful_entries(locale);
CREATE INDEX idx_entries_status ON contful_entries(status);
CREATE INDEX idx_entries_published ON contful_entries(published_time DESC) WHERE status = 'published';
CREATE INDEX idx_entries_sort ON contful_entries(schema_id, sort_weight);
CREATE INDEX idx_entries_created_by ON contful_entries(created_by);
CREATE INDEX idx_entries_deleted ON contful_entries(deleted_time) WHERE deleted_time IS NULL;
CREATE INDEX idx_entries_deleted_time ON contful_entries(deleted_time) WHERE deleted_time IS NOT NULL;
CREATE INDEX idx_entries_list ON contful_entries(schema_id, locale, status, published_time DESC);
CREATE INDEX idx_entries_site_locale ON contful_entries(site_id, locale, status);
CREATE INDEX idx_entries_scheduled_publish ON contful_entries(scheduled_publish_time) WHERE scheduled_publish_time IS NOT NULL AND status = 'draft' AND deleted_time IS NULL;
CREATE INDEX idx_entries_scheduled_unpublish ON contful_entries(scheduled_unpublish_time) WHERE scheduled_unpublish_time IS NOT NULL AND status = 'published' AND deleted_time IS NULL;

CREATE TRIGGER update_entries_updated_time
    BEFORE UPDATE ON contful_entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_entries IS '内容条目表：实际的内容数据';
COMMENT ON COLUMN contful_entries.id IS '条目唯一标识符';
COMMENT ON COLUMN contful_entries.schema_id IS '所属内容模型';
COMMENT ON COLUMN contful_entries.site_id IS '所属站点';
COMMENT ON COLUMN contful_entries.locale IS '语言/地区代码（如 zh-CN、en-US）';
COMMENT ON COLUMN contful_entries.status IS '状态：draft/published/archived';
COMMENT ON COLUMN contful_entries.version IS '当前版本号';
COMMENT ON COLUMN contful_entries.version_history IS '版本历史 JSON';
COMMENT ON COLUMN contful_entries.published_time IS '发布时间';
COMMENT ON COLUMN contful_entries.published_by IS '发布者';
COMMENT ON COLUMN contful_entries.scheduled_publish_time IS '定时发布时间';
COMMENT ON COLUMN contful_entries.scheduled_unpublish_time IS '定时下架时间';
COMMENT ON COLUMN contful_entries.relations IS '关联条目 JSON';
COMMENT ON COLUMN contful_entries.seo_title IS 'SEO 标题';
COMMENT ON COLUMN contful_entries.seo_description IS 'SEO 描述';
COMMENT ON COLUMN contful_entries.seo_keywords IS 'SEO 关键词';
COMMENT ON COLUMN contful_entries.sort_weight IS '排序权重';
COMMENT ON COLUMN contful_entries.data_signature IS '数据完整性签名: {"alg":"HMAC-SHA256","created_time":"...","payload_hash":"...","signature":"..."}';
COMMENT ON COLUMN contful_entries.created_by IS '创建者';
COMMENT ON COLUMN contful_entries.created_time IS '创建时间';
COMMENT ON COLUMN contful_entries.updated_time IS '更新时间';
COMMENT ON COLUMN contful_entries.deleted_time IS '软删除时间';

-- =============================================================================
-- 条目值表
-- =============================================================================

CREATE TABLE contful_entry_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES contful_entries(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES contful_fields(id) ON DELETE CASCADE,
    value JSONB NOT NULL,
    text_value TEXT,
    number_value NUMERIC,
    bool_value BOOLEAN,
    date_value DATE,
    datetime_value TIMESTAMPTZ,
    data_signature JSONB,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(entry_id, field_id)
);

CREATE INDEX idx_entry_values_entry ON contful_entry_values(entry_id);
CREATE INDEX idx_entry_values_field ON contful_entry_values(field_id);
-- GIN 全文搜索索引 [PG-ONLY: → MySQL FULLTEXT INDEX]
CREATE INDEX idx_entry_values_text_gin ON contful_entry_values USING gin(to_tsvector('simple', text_value)) WHERE text_value IS NOT NULL;

CREATE TRIGGER update_entry_values_updated_time
    BEFORE UPDATE ON contful_entry_values
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_entry_values IS '条目值表：字段的实际值（一行对应一个字段的值）';
COMMENT ON COLUMN contful_entry_values.id IS '值唯一标识符';
COMMENT ON COLUMN contful_entry_values.entry_id IS '所属条目';
COMMENT ON COLUMN contful_entry_values.field_id IS '所属字段';
COMMENT ON COLUMN contful_entry_values.value IS '值 JSONB（通用存储）';
COMMENT ON COLUMN contful_entry_values.text_value IS '文本值（索引用）';
COMMENT ON COLUMN contful_entry_values.number_value IS '数值（索引用）';
COMMENT ON COLUMN contful_entry_values.bool_value IS '布尔值（索引用）';
COMMENT ON COLUMN contful_entry_values.date_value IS '日期值（索引用）';
COMMENT ON COLUMN contful_entry_values.datetime_value IS '时间戳值（索引用）';
COMMENT ON COLUMN contful_entry_values.data_signature IS '字段值完整性签名';
COMMENT ON COLUMN contful_entry_values.created_time IS '创建时间';
COMMENT ON COLUMN contful_entry_values.updated_time IS '更新时间';

-- =============================================================================
-- 条目版本表
-- =============================================================================

CREATE TABLE contful_entry_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES contful_entries(id) ON DELETE CASCADE,
    version INT NOT NULL,
    values_snapshot JSONB NOT NULL,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    change_summary TEXT,
    UNIQUE(entry_id, version)
);

CREATE INDEX idx_entry_versions_entry ON contful_entry_versions(entry_id);
CREATE INDEX idx_entry_versions_created ON contful_entry_versions(created_time DESC);

COMMENT ON TABLE contful_entry_versions IS '条目版本表：内容的历史版本记录';
COMMENT ON COLUMN contful_entry_versions.id IS '版本唯一标识符';
COMMENT ON COLUMN contful_entry_versions.entry_id IS '所属条目';
COMMENT ON COLUMN contful_entry_versions.version IS '版本号';
COMMENT ON COLUMN contful_entry_versions.values_snapshot IS '值快照 JSON';
COMMENT ON COLUMN contful_entry_versions.created_by IS '创建者';
COMMENT ON COLUMN contful_entry_versions.created_time IS '创建时间';
COMMENT ON COLUMN contful_entry_versions.change_summary IS '变更摘要';

-- =============================================================================
-- 媒体文件夹表
-- =============================================================================

CREATE TABLE contful_asset_folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES contful_sites(id),
    parent_id UUID REFERENCES contful_asset_folders(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

CREATE INDEX idx_asset_folders_site ON contful_asset_folders(site_id);
CREATE INDEX idx_asset_folders_parent ON contful_asset_folders(parent_id);

CREATE TRIGGER update_asset_folders_updated_time
    BEFORE UPDATE ON contful_asset_folders
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_asset_folders IS '资产文件夹表：媒体库文件夹结构';
COMMENT ON COLUMN contful_asset_folders.id IS '文件夹唯一标识符';
COMMENT ON COLUMN contful_asset_folders.site_id IS '所属站点';
COMMENT ON COLUMN contful_asset_folders.parent_id IS '父文件夹';
COMMENT ON COLUMN contful_asset_folders.name IS '文件夹名称';
COMMENT ON COLUMN contful_asset_folders.slug IS '文件夹别名';
COMMENT ON COLUMN contful_asset_folders.path IS '完整路径';
COMMENT ON COLUMN contful_asset_folders.sort_order IS '排序权重';
COMMENT ON COLUMN contful_asset_folders.created_by IS '创建者';
COMMENT ON COLUMN contful_asset_folders.created_time IS '创建时间';
COMMENT ON COLUMN contful_asset_folders.updated_time IS '更新时间';
COMMENT ON COLUMN contful_asset_folders.deleted_time IS '软删除时间';

-- =============================================================================
-- 媒体资产表
-- =============================================================================

CREATE TABLE contful_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID REFERENCES contful_sites(id),
    folder_id UUID REFERENCES contful_asset_folders(id),
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
    data_signature JSONB,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

-- 部分唯一索引：仅对未删除的记录强制 uuid 唯一
CREATE UNIQUE INDEX idx_assets_uuid_active ON contful_assets(uuid) WHERE deleted_time IS NULL;
CREATE INDEX idx_assets_site ON contful_assets(site_id);
CREATE INDEX idx_assets_folder ON contful_assets(folder_id);
CREATE INDEX idx_assets_type ON contful_assets(type);
CREATE INDEX idx_assets_extension ON contful_assets(extension);
CREATE INDEX idx_assets_slug ON contful_assets(slug);
CREATE INDEX idx_assets_hash ON contful_assets(file_hash) WHERE file_hash IS NOT NULL;
CREATE INDEX idx_assets_created ON contful_assets(created_time DESC);
CREATE INDEX idx_assets_download ON contful_assets(download_count);
CREATE INDEX idx_assets_used ON contful_assets(used_count);
CREATE INDEX idx_assets_deleted_time ON contful_assets(deleted_time) WHERE deleted_time IS NOT NULL;
CREATE INDEX idx_assets_site_type ON contful_assets(site_id, type);
CREATE INDEX idx_assets_site_created ON contful_assets(site_id, created_time DESC);
CREATE INDEX idx_assets_path ON contful_assets(path);

CREATE TRIGGER update_assets_updated_time
    BEFORE UPDATE ON contful_assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_assets IS '资产表：图片/视频/文档等媒体文件';
COMMENT ON COLUMN contful_assets.id IS '资产唯一标识符';
COMMENT ON COLUMN contful_assets.site_id IS '所属站点';
COMMENT ON COLUMN contful_assets.folder_id IS '所属文件夹';
COMMENT ON COLUMN contful_assets.uuid IS '业务 UUID（用于引用）';
COMMENT ON COLUMN contful_assets.name IS '显示名称';
COMMENT ON COLUMN contful_assets.original_name IS '原始文件名';
COMMENT ON COLUMN contful_assets.slug IS 'URL 别名';
COMMENT ON COLUMN contful_assets.type IS '资源类型';
COMMENT ON COLUMN contful_assets.mime_type IS 'MIME 类型';
COMMENT ON COLUMN contful_assets.extension IS '文件扩展名';
COMMENT ON COLUMN contful_assets.size IS '文件大小（字节）';
COMMENT ON COLUMN contful_assets.width IS '宽度（图片/视频）';
COMMENT ON COLUMN contful_assets.height IS '高度（图片/视频）';
COMMENT ON COLUMN contful_assets.duration IS '时长（音视频，秒）';
COMMENT ON COLUMN contful_assets.path IS '存储路径';
COMMENT ON COLUMN contful_assets.url IS '访问 URL';
COMMENT ON COLUMN contful_assets.thumbnail_url IS '缩略图 URL';
COMMENT ON COLUMN contful_assets.alt IS 'Alt 文本';
COMMENT ON COLUMN contful_assets.title IS '标题';
COMMENT ON COLUMN contful_assets.caption IS '图注';
COMMENT ON COLUMN contful_assets.alt_text IS '备用描述';
COMMENT ON COLUMN contful_assets.description IS '描述';
COMMENT ON COLUMN contful_assets.tags IS '标签列表';
COMMENT ON COLUMN contful_assets.metadata IS '元数据 JSON：EXIF 等';
COMMENT ON COLUMN contful_assets.visibility IS '可见性：public/private';
COMMENT ON COLUMN contful_assets.file_hash IS '文件哈希（SHA256）';
COMMENT ON COLUMN contful_assets.disk IS '存储驱动：local/s3/oss/cos 等';
COMMENT ON COLUMN contful_assets.download_count IS '下载次数';
COMMENT ON COLUMN contful_assets.used_count IS '引用次数';
COMMENT ON COLUMN contful_assets.data_signature IS '媒体资源元信息完整性签名';
COMMENT ON COLUMN contful_assets.created_by IS '上传者';
COMMENT ON COLUMN contful_assets.created_time IS '上传时间';
COMMENT ON COLUMN contful_assets.updated_time IS '更新时间';
COMMENT ON COLUMN contful_assets.deleted_time IS '软删除时间';

-- =============================================================================
-- API Token 表
-- =============================================================================

CREATE TABLE contful_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES contful_sites(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    token_prefix VARCHAR(20) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    encrypted_token TEXT NOT NULL,
    expires_time TIMESTAMPTZ,
    status token_status NOT NULL DEFAULT 'active',
    last_used_time TIMESTAMPTZ,
    last_used_ip INET,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);

CREATE INDEX idx_tokens_site ON contful_tokens(site_id);
CREATE INDEX idx_tokens_hash ON contful_tokens(token_hash);
CREATE INDEX idx_tokens_status ON contful_tokens(status);
CREATE INDEX idx_tokens_expires ON contful_tokens(expires_time) WHERE expires_time IS NOT NULL;
CREATE INDEX idx_tokens_deleted_time ON contful_tokens(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_tokens_updated_time
    BEFORE UPDATE ON contful_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE contful_tokens IS 'API Token 表：用于 Open API 认证';
COMMENT ON COLUMN contful_tokens.id IS 'Token 唯一标识符';
COMMENT ON COLUMN contful_tokens.site_id IS '所属站点';
COMMENT ON COLUMN contful_tokens.name IS 'Token 名称';
COMMENT ON COLUMN contful_tokens.description IS 'Token 描述';
COMMENT ON COLUMN contful_tokens.token_prefix IS 'Token 前缀（ctf_，显示用，10 字符）';
COMMENT ON COLUMN contful_tokens.token_hash IS 'Token 哈希（SHA-256，验证用）';
COMMENT ON COLUMN contful_tokens.encrypted_token IS '加密存储的完整 Token（AES-256-GCM）';
COMMENT ON COLUMN contful_tokens.expires_time IS '过期时间（NULL 表示永不过期）';
COMMENT ON COLUMN contful_tokens.status IS '状态：active / expired / revoked';
COMMENT ON COLUMN contful_tokens.last_used_time IS '最后使用时间';
COMMENT ON COLUMN contful_tokens.last_used_ip IS '最后使用 IP';
COMMENT ON COLUMN contful_tokens.created_by IS '创建者用户 ID';
COMMENT ON COLUMN contful_tokens.created_time IS '创建时间';
COMMENT ON COLUMN contful_tokens.updated_time IS '更新时间';
COMMENT ON COLUMN contful_tokens.deleted_time IS '软删除时间';

-- =============================================================================
-- 生产环境迁移脚本 (v1.3.0 - 排期功能)
-- 用于已有数据库的增量迁移，每个语句都是幂等的（IF NOT EXISTS / IF EXISTS）
-- =============================================================================

ALTER TABLE contful_entries ADD COLUMN IF NOT EXISTS scheduled_publish_time TIMESTAMPTZ;
ALTER TABLE contful_entries ADD COLUMN IF NOT EXISTS scheduled_unpublish_time TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_entries_scheduled_publish
    ON contful_entries(scheduled_publish_time)
    WHERE scheduled_publish_time IS NOT NULL AND status = 'draft' AND deleted_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_entries_scheduled_unpublish
    ON contful_entries(scheduled_unpublish_time)
    WHERE scheduled_unpublish_time IS NOT NULL AND status = 'published' AND deleted_time IS NULL;


-- =============================================================================
-- 种子数据
-- 幂等设计：使用 INSERT ... WHERE NOT EXISTS / ON CONFLICT DO NOTHING
-- =============================================================================

-- =============================================================================
-- 1. 默认站点
-- =============================================================================

INSERT INTO contful_sites (id, name, slug, description, locale, timezone, is_active, settings)
SELECT
    '00000000-0000-0000-0000-000000000001'::uuid,
    '默认站点',
    'default',
    '系统默认站点',
    'zh-CN',
    'Asia/Shanghai',
    TRUE,
    '{}'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM contful_sites WHERE id = '00000000-0000-0000-0000-000000000001'::uuid);

-- =============================================================================
-- 2. 系统角色
-- =============================================================================

-- 超级管理员角色（拥有全部系统级权限）
INSERT INTO contful_system_roles (id, name, description, is_system, permissions, created_time, updated_time)
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
WHERE NOT EXISTS (SELECT 1 FROM contful_system_roles WHERE id = '00000000-0000-0000-0000-000000000101'::uuid);

-- 审计员角色（仅可查看和导出审计日志）
INSERT INTO contful_system_roles (id, name, description, is_system, permissions, created_time, updated_time)
SELECT
    '00000000-0000-0000-0000-000000000103'::uuid,
    '审计人员',
    '审计人员，仅可查看和导出审计日志',
    TRUE,
    '["users:read","sites:read","tokens:read",
      "settings:read","audit:read","audit:export"]'::jsonb,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM contful_system_roles WHERE id = '00000000-0000-0000-0000-000000000103'::uuid);

-- =============================================================================
-- 3. 权限元数据（分组 + 权限项，用于角色权限配置）
-- =============================================================================

-- 权限分组
INSERT INTO contful_system_permission_groups (id, group_key, label, label_en, sort_order)
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
WHERE NOT EXISTS (SELECT 1 FROM contful_system_permission_groups WHERE id = t.id::uuid);

-- 权限项
INSERT INTO contful_system_permissions (group_id, action, label, label_en, sort_order)
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
JOIN contful_system_permission_groups g ON g.id = t.group_id::uuid
WHERE NOT EXISTS (
    SELECT 1 FROM contful_system_permissions p WHERE p.group_id = g.id AND p.action = t.action
);

-- =============================================================================
-- 4. 系统用户
-- =============================================================================

-- 默认管理员用户（密码：contful@com）
INSERT INTO contful_system_users (id, email, password_hash, nickname, status, is_super_admin, created_time, updated_time)
SELECT
    '00000000-0000-0000-0000-000000000001'::uuid,
    'admin@contful.com',
    '$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu',  -- 密码：contful@com
    'Administrator',
    'active',
    TRUE,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM contful_system_users WHERE id = '00000000-0000-0000-0000-000000000001'::uuid);

-- =============================================================================
-- 5. 用户-角色关联
-- =============================================================================

-- 关联 admin 用户与 Super Admin 角色
INSERT INTO contful_system_user_roles (user_id, role_id, created_time)
SELECT
    '00000000-0000-0000-0000-000000000001'::uuid,
    '00000000-0000-0000-0000-000000000101'::uuid,
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM contful_system_user_roles 
    WHERE user_id = '00000000-0000-0000-0000-000000000001'::uuid 
    AND role_id = '00000000-0000-0000-0000-000000000101'::uuid
);

-- =============================================================================
-- 6. 系统配置
-- =============================================================================

INSERT INTO contful_system_config (config_key, config_value, value_type, description, is_public, is_system, created_time, updated_time)
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
