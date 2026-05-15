-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Contful - Headless CMS 数据库初始化脚本
-- 版本: v2.0.0
-- 数据库: PostgreSQL 14+
--
-- 使用方式：
--   psql -h <host> -U <user> -d <db> -f init_pg.sql
--
-- 注意：此脚本为完整重建脚本，会删除所有已有对象，仅用于全新部署或开发环境重置。
--
-- [PG-ONLY] 标记说明：以下语法为 PostgreSQL 专有，迁移到其他数据库时需替换：
--   - CREATE EXTENSION      → MySQL/达梦不需要
--   - ENUM TYPE             → MySQL: ENUM 或 VARCHAR + CHECK; 达梦: VARCHAR + CHECK
--   - gen_random_uuid()     → MySQL: UUID(); 达梦: SYS_GUID()
--   - UUID 类型             → MySQL: VARCHAR(36); 达梦: VARCHAR(36)
--   - JSONB 类型            → MySQL: JSON; 达梦: CLOB
--   - TIMESTAMPTZ           → MySQL: DATETIME; 达梦: TIMESTAMP
--   - INET 类型             → MySQL: VARCHAR(45); 达梦: VARCHAR(45)
--   - 部分索引 (WHERE)      → MySQL 不支持，需用生成的列+普通索引替代
--   - CREATE TRIGGER        → MySQL: 语法略有不同; 达梦: 类似
--   - ON CONFLICT DO NOTHING → MySQL: ON DUPLICATE KEY UPDATE; 达梦: MERGE INTO
--   - GIN 索引              → MySQL: FULLTEXT; 达梦: 不支持
--   - plpgsql 函数          → MySQL: 无存储过程替代方案
-- =============================================================================

-- 启用扩展 [PG-ONLY]
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- 清理已有对象（支持重复执行）
-- =============================================================================

-- 删除触发器
DROP TRIGGER IF EXISTS update_system_users_updated_time ON system_users;
DROP TRIGGER IF EXISTS update_system_roles_updated_time ON system_roles;
DROP TRIGGER IF EXISTS update_system_config_updated_time ON system_config;
DROP TRIGGER IF EXISTS update_sites_updated_time ON sites;
DROP TRIGGER IF EXISTS update_schemas_updated_time ON schemas;
DROP TRIGGER IF EXISTS update_fields_updated_time ON fields;
DROP TRIGGER IF EXISTS update_entries_updated_time ON entries;
DROP TRIGGER IF EXISTS update_entry_values_updated_time ON entry_values;
DROP TRIGGER IF EXISTS update_asset_folders_updated_time ON asset_folders;
DROP TRIGGER IF EXISTS update_assets_updated_time ON assets;
DROP TRIGGER IF EXISTS update_tokens_updated_time ON tokens;
DROP TRIGGER IF EXISTS prevent_audit_logs_update ON audit_logs;

-- 删除函数
DROP FUNCTION IF EXISTS update_updated_time_column();
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

CREATE TABLE system_users (
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
    last_login_time TIMESTAMPTZ, -- [PG-ONLY: → MySQL DATETIME]
    last_login_ip INET,           -- [PG-ONLY: → MySQL VARCHAR(45)]
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- 部分唯一索引：仅对未删除的记录强制 email 唯一 [PG-ONLY: MySQL 不支持部分索引]
CREATE UNIQUE INDEX idx_system_users_email_active ON system_users(email) WHERE deleted_time IS NULL;
CREATE INDEX idx_system_users_email ON system_users(email);
CREATE INDEX idx_system_users_status ON system_users(status);
CREATE INDEX idx_system_users_deleted_time ON system_users(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_system_users_updated_time
    BEFORE UPDATE ON system_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE system_users IS '系统用户表：存放所有用户账户，用户全局管理，权限通过 system_user_roles 关联';
COMMENT ON COLUMN system_users.id IS '用户唯一标识符(UUID)';
COMMENT ON COLUMN system_users.email IS '用户邮箱（未删除时全局唯一）';
COMMENT ON COLUMN system_users.password_hash IS 'bcrypt 加密的密码哈希';
COMMENT ON COLUMN system_users.nickname IS '用户昵称';
COMMENT ON COLUMN system_users.avatar_url IS '头像 URL';
COMMENT ON COLUMN system_users.phone IS '用户手机号';
COMMENT ON COLUMN system_users.department IS '用户所属部门';
COMMENT ON COLUMN system_users.status IS '账户状态：active/inactive/suspended';
COMMENT ON COLUMN system_users.is_super_admin IS '是否超级管理员（拥有全部权限，不受站点限制）';
COMMENT ON COLUMN system_users.last_login_time IS '最后登录时间';
COMMENT ON COLUMN system_users.last_login_ip IS '最后登录 IP 地址';
COMMENT ON COLUMN system_users.created_time IS '创建时间';
COMMENT ON COLUMN system_users.updated_time IS '更新时间';
COMMENT ON COLUMN system_users.deleted_time IS '软删除时间（非空表示已删除）';

-- =============================================================================
-- 系统角色表
-- =============================================================================

CREATE TABLE system_roles (
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
CREATE UNIQUE INDEX idx_system_roles_name_active ON system_roles(name) WHERE deleted_time IS NULL;
CREATE INDEX idx_system_roles_name ON system_roles(name);
CREATE INDEX idx_system_roles_deleted_time ON system_roles(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_system_roles_updated_time
    BEFORE UPDATE ON system_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE system_roles IS '系统角色表：系统级角色定义，系统角色不可删除';
COMMENT ON COLUMN system_roles.id IS '角色唯一标识符';
COMMENT ON COLUMN system_roles.name IS '角色名称（未删除时全局唯一）';
COMMENT ON COLUMN system_roles.description IS '角色描述';
COMMENT ON COLUMN system_roles.is_system IS '是否系统角色（系统角色不可删除）';
COMMENT ON COLUMN system_roles.permissions IS '权限列表 JSON：["user:read", "user:write", ...]';
COMMENT ON COLUMN system_roles.created_time IS '创建时间';
COMMENT ON COLUMN system_roles.updated_time IS '更新时间';
COMMENT ON COLUMN system_roles.deleted_time IS '软删除时间';

-- =============================================================================
-- 系统用户-角色关联表
-- =============================================================================

CREATE TABLE system_user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES system_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_system_user_roles_user ON system_user_roles(user_id);
CREATE INDEX idx_system_user_roles_role ON system_user_roles(role_id);

COMMENT ON TABLE system_user_roles IS '系统用户-角色关联表：多对多关系，一个用户可拥有多个系统角色';
COMMENT ON COLUMN system_user_roles.id IS '关联记录唯一标识符';
COMMENT ON COLUMN system_user_roles.user_id IS '用户 ID';
COMMENT ON COLUMN system_user_roles.role_id IS '系统角色 ID';
COMMENT ON COLUMN system_user_roles.created_time IS '创建时间';

-- =============================================================================
-- 系统配置表
-- =============================================================================

CREATE TABLE system_config (
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

CREATE INDEX idx_system_config_key ON system_config(config_key);

CREATE TRIGGER update_system_config_updated_time
    BEFORE UPDATE ON system_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE system_config IS '系统配置表：存储全局配置项（密码策略、系统名称、Logo 等）';
COMMENT ON COLUMN system_config.id IS '配置项唯一标识符';
COMMENT ON COLUMN system_config.config_key IS '配置键名（唯一，如 password_expire_days、site_name、logo_url）';
COMMENT ON COLUMN system_config.config_value IS '配置值（根据 value_type 解析）';
COMMENT ON COLUMN system_config.value_type IS '值类型：string/number/boolean/json';
COMMENT ON COLUMN system_config.description IS '配置项描述';
COMMENT ON COLUMN system_config.is_public IS '是否公开（公开配置可未授权访问）';
COMMENT ON COLUMN system_config.is_system IS '是否系统配置（系统配置不可删除，自定义配置可删除）';
COMMENT ON COLUMN system_config.created_time IS '创建时间';
COMMENT ON COLUMN system_config.updated_time IS '更新时间';

-- =============================================================================
-- 审计日志表
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
    data_signature JSONB NOT NULL DEFAULT '{}',
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_site ON audit_logs(site_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_category ON audit_logs(category);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_time DESC);
CREATE INDEX idx_audit_logs_site_user_time ON audit_logs(site_id, user_id, created_time DESC);
CREATE INDEX idx_audit_logs_category_time ON audit_logs(category, created_time DESC);

-- 审计日志防篡改触发器
CREATE TRIGGER prevent_audit_logs_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_update();

COMMENT ON TABLE audit_logs IS '审计日志表：记录所有操作行为，不可修改';
COMMENT ON COLUMN audit_logs.id IS '日志唯一标识符';
COMMENT ON COLUMN audit_logs.site_id IS '关联站点 ID（全局操作时为空）';
COMMENT ON COLUMN audit_logs.user_id IS '操作用户 ID';
COMMENT ON COLUMN audit_logs.action IS '操作动作名称';
COMMENT ON COLUMN audit_logs.resource_type IS '资源类型';
COMMENT ON COLUMN audit_logs.resource_id IS '资源 ID';
COMMENT ON COLUMN audit_logs.level IS '日志级别：debug/info/warn/error';
COMMENT ON COLUMN audit_logs.category IS '操作类别：auth/content/media/settings/user/system';
COMMENT ON COLUMN audit_logs.details IS '操作详情 JSON';
COMMENT ON COLUMN audit_logs.ip_address IS '客户端 IP';
COMMENT ON COLUMN audit_logs.user_agent IS '客户端 User-Agent';
COMMENT ON COLUMN audit_logs.data_signature IS '审计日志完整性签名，防篡改';
COMMENT ON COLUMN audit_logs.created_time IS '日志创建时间';

-- =============================================================================
-- 站点表
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

CREATE TRIGGER update_sites_updated_time
    BEFORE UPDATE ON sites
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE sites IS '站点表：每个站点是一个独立的 CMS 实例';
COMMENT ON COLUMN sites.id IS '站点唯一标识符';
COMMENT ON COLUMN sites.name IS '站点名称';
COMMENT ON COLUMN sites.slug IS '站点别名（URL 中使用）';
COMMENT ON COLUMN sites.description IS '站点描述';
COMMENT ON COLUMN sites.site_url IS '站点访问地址（前端展示端点）';
COMMENT ON COLUMN sites.locale IS '站点默认语言（如 zh-CN、en-US）';
COMMENT ON COLUMN sites.timezone IS '站点时区（如 Asia/Shanghai、America/New_York）';
COMMENT ON COLUMN sites.seo_title IS '站点 SEO 标题';
COMMENT ON COLUMN sites.seo_description IS '站点 SEO 描述';
COMMENT ON COLUMN sites.seo_keywords IS '站点 SEO 关键词';
COMMENT ON COLUMN sites.settings IS '动态配置 JSON（主题、第三方集成等）';
COMMENT ON COLUMN sites.is_active IS '是否激活';
COMMENT ON COLUMN sites.created_by IS '创建者用户 ID';
COMMENT ON COLUMN sites.created_time IS '创建时间';
COMMENT ON COLUMN sites.updated_time IS '更新时间';
COMMENT ON COLUMN sites.deleted_time IS '软删除时间';

-- =============================================================================
-- 内容模型表（schemas）
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
    signature_enabled BOOLEAN NOT NULL DEFAULT FALSE,
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

CREATE TRIGGER update_schemas_updated_time
    BEFORE UPDATE ON schemas
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE schemas IS '内容模型表：定义内容的结构（类似 WordPress 的自定义内容类型）';
COMMENT ON COLUMN schemas.id IS '内容模型唯一标识符';
COMMENT ON COLUMN schemas.site_id IS '所属站点';
COMMENT ON COLUMN schemas.name IS '模型名称';
COMMENT ON COLUMN schemas.slug IS '模型别名（API 中使用）';
COMMENT ON COLUMN schemas.description IS '模型描述';
COMMENT ON COLUMN schemas.kind IS '类型：collection=集合、single=单页';
COMMENT ON COLUMN schemas.display_config IS '前端显示配置 JSON';
COMMENT ON COLUMN schemas.api_config IS 'API 配置 JSON：publicRead/publicWrite 等';
COMMENT ON COLUMN schemas.preview_config IS '预览配置 JSON';
COMMENT ON COLUMN schemas.versioning_enabled IS '是否启用版本历史';
COMMENT ON COLUMN schemas.draft_autosave_interval IS '草稿自动保存间隔（秒）';
COMMENT ON COLUMN schemas.is_active IS '是否启用';
COMMENT ON COLUMN schemas.sort_order IS '排序权重';
COMMENT ON COLUMN schemas.signature_enabled IS '是否启用数据签名（该内容类型下的条目自动签名）';
COMMENT ON COLUMN schemas.created_by IS '创建者';
COMMENT ON COLUMN schemas.created_time IS '创建时间';
COMMENT ON COLUMN schemas.updated_time IS '更新时间';
COMMENT ON COLUMN schemas.deleted_time IS '软删除时间';

-- =============================================================================
-- 字段表
-- =============================================================================

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

CREATE TRIGGER update_fields_updated_time
    BEFORE UPDATE ON fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE fields IS '字段表：内容模型的字段定义';
COMMENT ON COLUMN fields.id IS '字段唯一标识符';
COMMENT ON COLUMN fields.schema_id IS '所属内容模型';
COMMENT ON COLUMN fields.name IS '字段名（代码中使用）';
COMMENT ON COLUMN fields.label IS '字段显示标签';
COMMENT ON COLUMN fields.description IS '字段描述';
COMMENT ON COLUMN fields.field_type IS '字段类型：text/rich_text/number/boolean/date/datetime/email/url/json/media/relation/enum/password';
COMMENT ON COLUMN fields.config IS '字段配置 JSON：maxLength/minValue/options 等';
COMMENT ON COLUMN fields.validation IS '验证规则 JSON：required/pattern/min/max 等';
COMMENT ON COLUMN fields.display IS '显示配置 JSON：placeholder/readOnly 等';
COMMENT ON COLUMN fields.default_value IS '默认值 JSON';
COMMENT ON COLUMN fields.sort_order IS '排序权重';
COMMENT ON COLUMN fields.conditional_display IS '条件显示 JSON';
COMMENT ON COLUMN fields.created_time IS '创建时间';
COMMENT ON COLUMN fields.updated_time IS '更新时间';
COMMENT ON COLUMN fields.deleted_time IS '软删除时间';

-- =============================================================================
-- 内容条目表
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
    data_signature JSONB,
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
CREATE INDEX idx_entries_deleted_time ON entries(deleted_time) WHERE deleted_time IS NOT NULL;
CREATE INDEX idx_entries_list ON entries(schema_id, locale, status, published_time DESC);
CREATE INDEX idx_entries_site_locale ON entries(site_id, locale, status);

CREATE TRIGGER update_entries_updated_time
    BEFORE UPDATE ON entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE entries IS '内容条目表：实际的内容数据';
COMMENT ON COLUMN entries.id IS '条目唯一标识符';
COMMENT ON COLUMN entries.schema_id IS '所属内容模型';
COMMENT ON COLUMN entries.site_id IS '所属站点';
COMMENT ON COLUMN entries.locale IS '语言/地区代码（如 zh-CN、en-US）';
COMMENT ON COLUMN entries.status IS '状态：draft/published/archived';
COMMENT ON COLUMN entries.version IS '当前版本号';
COMMENT ON COLUMN entries.version_history IS '版本历史 JSON';
COMMENT ON COLUMN entries.published_time IS '发布时间';
COMMENT ON COLUMN entries.published_by IS '发布者';
COMMENT ON COLUMN entries.relations IS '关联条目 JSON';
COMMENT ON COLUMN entries.seo_title IS 'SEO 标题';
COMMENT ON COLUMN entries.seo_description IS 'SEO 描述';
COMMENT ON COLUMN entries.seo_keywords IS 'SEO 关键词';
COMMENT ON COLUMN entries.sort_weight IS '排序权重';
COMMENT ON COLUMN entries.data_signature IS '数据完整性签名: {"alg":"HMAC-SHA256","created_time":"...","payload_hash":"...","signature":"..."}';
COMMENT ON COLUMN entries.created_by IS '创建者';
COMMENT ON COLUMN entries.created_time IS '创建时间';
COMMENT ON COLUMN entries.updated_time IS '更新时间';
COMMENT ON COLUMN entries.deleted_time IS '软删除时间';

-- =============================================================================
-- 条目值表
-- =============================================================================

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
    data_signature JSONB,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(entry_id, field_id)
);

CREATE INDEX idx_entry_values_entry ON entry_values(entry_id);
CREATE INDEX idx_entry_values_field ON entry_values(field_id);
-- GIN 全文搜索索引 [PG-ONLY: → MySQL FULLTEXT INDEX]
CREATE INDEX idx_entry_values_text_gin ON entry_values USING gin(to_tsvector('simple', text_value)) WHERE text_value IS NOT NULL;

CREATE TRIGGER update_entry_values_updated_time
    BEFORE UPDATE ON entry_values
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE entry_values IS '条目值表：字段的实际值（一行对应一个字段的值）';
COMMENT ON COLUMN entry_values.id IS '值唯一标识符';
COMMENT ON COLUMN entry_values.entry_id IS '所属条目';
COMMENT ON COLUMN entry_values.field_id IS '所属字段';
COMMENT ON COLUMN entry_values.value IS '值 JSONB（通用存储）';
COMMENT ON COLUMN entry_values.text_value IS '文本值（索引用）';
COMMENT ON COLUMN entry_values.number_value IS '数值（索引用）';
COMMENT ON COLUMN entry_values.bool_value IS '布尔值（索引用）';
COMMENT ON COLUMN entry_values.date_value IS '日期值（索引用）';
COMMENT ON COLUMN entry_values.datetime_value IS '时间戳值（索引用）';
COMMENT ON COLUMN entry_values.data_signature IS '字段值完整性签名';
COMMENT ON COLUMN entry_values.created_time IS '创建时间';
COMMENT ON COLUMN entry_values.updated_time IS '更新时间';

-- =============================================================================
-- 条目版本表
-- =============================================================================

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

COMMENT ON TABLE entry_versions IS '条目版本表：内容的历史版本记录';
COMMENT ON COLUMN entry_versions.id IS '版本唯一标识符';
COMMENT ON COLUMN entry_versions.entry_id IS '所属条目';
COMMENT ON COLUMN entry_versions.version IS '版本号';
COMMENT ON COLUMN entry_versions.values_snapshot IS '值快照 JSON';
COMMENT ON COLUMN entry_versions.created_by IS '创建者';
COMMENT ON COLUMN entry_versions.created_time IS '创建时间';
COMMENT ON COLUMN entry_versions.change_summary IS '变更摘要';

-- =============================================================================
-- 媒体文件夹表
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

CREATE TRIGGER update_asset_folders_updated_time
    BEFORE UPDATE ON asset_folders
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE asset_folders IS '资产文件夹表：媒体库文件夹结构';
COMMENT ON COLUMN asset_folders.id IS '文件夹唯一标识符';
COMMENT ON COLUMN asset_folders.site_id IS '所属站点';
COMMENT ON COLUMN asset_folders.parent_id IS '父文件夹';
COMMENT ON COLUMN asset_folders.name IS '文件夹名称';
COMMENT ON COLUMN asset_folders.slug IS '文件夹别名';
COMMENT ON COLUMN asset_folders.path IS '完整路径';
COMMENT ON COLUMN asset_folders.sort_order IS '排序权重';
COMMENT ON COLUMN asset_folders.created_by IS '创建者';
COMMENT ON COLUMN asset_folders.created_time IS '创建时间';
COMMENT ON COLUMN asset_folders.updated_time IS '更新时间';
COMMENT ON COLUMN asset_folders.deleted_time IS '软删除时间';

-- =============================================================================
-- 媒体资产表
-- =============================================================================

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
    data_signature JSONB,
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
CREATE INDEX idx_assets_site_type ON assets(site_id, type);
CREATE INDEX idx_assets_site_created ON assets(site_id, created_time DESC);
CREATE INDEX idx_assets_path ON assets(path);

CREATE TRIGGER update_assets_updated_time
    BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE assets IS '资产表：图片/视频/文档等媒体文件';
COMMENT ON COLUMN assets.id IS '资产唯一标识符';
COMMENT ON COLUMN assets.site_id IS '所属站点';
COMMENT ON COLUMN assets.folder_id IS '所属文件夹';
COMMENT ON COLUMN assets.uuid IS '业务 UUID（用于引用）';
COMMENT ON COLUMN assets.name IS '显示名称';
COMMENT ON COLUMN assets.original_name IS '原始文件名';
COMMENT ON COLUMN assets.slug IS 'URL 别名';
COMMENT ON COLUMN assets.type IS '资源类型';
COMMENT ON COLUMN assets.mime_type IS 'MIME 类型';
COMMENT ON COLUMN assets.extension IS '文件扩展名';
COMMENT ON COLUMN assets.size IS '文件大小（字节）';
COMMENT ON COLUMN assets.width IS '宽度（图片/视频）';
COMMENT ON COLUMN assets.height IS '高度（图片/视频）';
COMMENT ON COLUMN assets.duration IS '时长（音视频，秒）';
COMMENT ON COLUMN assets.path IS '存储路径';
COMMENT ON COLUMN assets.url IS '访问 URL';
COMMENT ON COLUMN assets.thumbnail_url IS '缩略图 URL';
COMMENT ON COLUMN assets.alt IS 'Alt 文本';
COMMENT ON COLUMN assets.title IS '标题';
COMMENT ON COLUMN assets.caption IS '图注';
COMMENT ON COLUMN assets.alt_text IS '备用描述';
COMMENT ON COLUMN assets.description IS '描述';
COMMENT ON COLUMN assets.tags IS '标签列表';
COMMENT ON COLUMN assets.metadata IS '元数据 JSON：EXIF 等';
COMMENT ON COLUMN assets.visibility IS '可见性：public/private';
COMMENT ON COLUMN assets.file_hash IS '文件哈希（SHA256）';
COMMENT ON COLUMN assets.disk IS '存储驱动：local/s3/oss/cos 等';
COMMENT ON COLUMN assets.download_count IS '下载次数';
COMMENT ON COLUMN assets.used_count IS '引用次数';
COMMENT ON COLUMN assets.data_signature IS '媒体资源元信息完整性签名';
COMMENT ON COLUMN assets.created_by IS '上传者';
COMMENT ON COLUMN assets.created_time IS '上传时间';
COMMENT ON COLUMN assets.updated_time IS '更新时间';
COMMENT ON COLUMN assets.deleted_time IS '软删除时间';

-- =============================================================================
-- API Token 表
-- =============================================================================

CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
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

CREATE INDEX idx_tokens_site ON tokens(site_id);
CREATE INDEX idx_tokens_hash ON tokens(token_hash);
CREATE INDEX idx_tokens_status ON tokens(status);
CREATE INDEX idx_tokens_expires ON tokens(expires_time) WHERE expires_time IS NOT NULL;
CREATE INDEX idx_tokens_deleted_time ON tokens(deleted_time) WHERE deleted_time IS NOT NULL;

CREATE TRIGGER update_tokens_updated_time
    BEFORE UPDATE ON tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

COMMENT ON TABLE tokens IS 'API Token 表：用于 Open API 认证';
COMMENT ON COLUMN tokens.id IS 'Token 唯一标识符';
COMMENT ON COLUMN tokens.site_id IS '所属站点';
COMMENT ON COLUMN tokens.name IS 'Token 名称';
COMMENT ON COLUMN tokens.description IS 'Token 描述';
COMMENT ON COLUMN tokens.token_prefix IS 'Token 前缀（ctf_，显示用，10 字符）';
COMMENT ON COLUMN tokens.token_hash IS 'Token 哈希（SHA-256，验证用）';
COMMENT ON COLUMN tokens.encrypted_token IS '加密存储的完整 Token（AES-256-GCM）';
COMMENT ON COLUMN tokens.expires_time IS '过期时间（NULL 表示永不过期）';
COMMENT ON COLUMN tokens.status IS '状态：active / expired / revoked';
COMMENT ON COLUMN tokens.last_used_time IS '最后使用时间';
COMMENT ON COLUMN tokens.last_used_ip IS '最后使用 IP';
COMMENT ON COLUMN tokens.created_by IS '创建者用户 ID';
COMMENT ON COLUMN tokens.created_time IS '创建时间';
COMMENT ON COLUMN tokens.updated_time IS '更新时间';
COMMENT ON COLUMN tokens.deleted_time IS '软删除时间';
