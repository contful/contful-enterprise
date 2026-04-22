-- =============================================================================
-- Contful - Headless CMS 数据库初始化脚本
-- 版本: v1.0.0
-- 数据库: PostgreSQL 18
-- 使用: psql -h <host> -U <user> -d <db> -f init_pg.sql
-- =============================================================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- 清理已有对象（支持重复执行）
-- =============================================================================

-- 删除触发器
DROP TRIGGER IF EXISTS update_system_users_updated_time ON system_users;
DROP TRIGGER IF EXISTS update_system_roles_updated_time ON system_roles;
DROP TRIGGER IF EXISTS update_plugins_updated_time ON plugins;
DROP TRIGGER IF EXISTS update_sites_updated_time ON sites;
DROP TRIGGER IF EXISTS update_channels_updated_time ON channels;
DROP TRIGGER IF EXISTS update_locales_updated_time ON locales;
DROP TRIGGER IF EXISTS update_site_roles_updated_time ON site_roles;
DROP TRIGGER IF EXISTS update_site_users_updated_time ON site_users;
DROP TRIGGER IF EXISTS update_content_types_updated_time ON content_types;
DROP TRIGGER IF EXISTS update_fields_updated_time ON fields;
DROP TRIGGER IF EXISTS update_entries_updated_time ON entries;
DROP TRIGGER IF EXISTS update_entry_values_updated_time ON entry_values;
DROP TRIGGER IF EXISTS update_entry_versions_updated_time ON entry_versions;
DROP TRIGGER IF EXISTS update_assets_updated_time ON assets;
DROP TRIGGER IF EXISTS update_api_tokens_updated_time ON api_tokens;
DROP TRIGGER IF EXISTS update_webhooks_updated_time ON webhooks;

-- 删除表（按依赖顺序）
DROP TABLE IF EXISTS webhook_deliveries CASCADE;
DROP TABLE IF EXISTS webhooks CASCADE;
DROP TABLE IF EXISTS api_tokens CASCADE;
DROP TABLE IF EXISTS assets CASCADE;
DROP TABLE IF EXISTS asset_folders CASCADE;
DROP TABLE IF EXISTS entry_versions CASCADE;
DROP TABLE IF EXISTS entry_values CASCADE;
DROP TABLE IF EXISTS entries CASCADE;
DROP TABLE IF EXISTS fields CASCADE;
DROP TABLE IF EXISTS content_types CASCADE;
DROP TABLE IF EXISTS site_users CASCADE;
DROP TABLE IF EXISTS site_roles CASCADE;
DROP TABLE IF EXISTS locales CASCADE;
DROP TABLE IF EXISTS channels CASCADE;
DROP TABLE IF EXISTS sites CASCADE;
DROP TABLE IF EXISTS plugins CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS system_roles CASCADE;
DROP TABLE IF EXISTS system_users CASCADE;
DROP TABLE IF EXISTS distributed_locks CASCADE;

-- 删除视图
DROP VIEW IF EXISTS active_locks;

-- 删除函数
DROP FUNCTION IF EXISTS cleanup_expired_locks();
DROP FUNCTION IF EXISTS update_updated_time_column();

-- 删除枚举类型
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS entry_status;
DROP TYPE IF EXISTS content_type_kind;
DROP TYPE IF EXISTS field_type;
DROP TYPE IF EXISTS asset_type;
DROP TYPE IF EXISTS asset_status;
DROP TYPE IF EXISTS asset_visibility;
DROP TYPE IF EXISTS token_status;
DROP TYPE IF EXISTS audit_level;
DROP TYPE IF EXISTS audit_type;

-- =============================================================================
-- 枚举类型
-- =============================================================================

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE entry_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE content_type_kind AS ENUM ('collection', 'single');
CREATE TYPE field_type AS ENUM (
    'text', 'rich_text', 'number', 'boolean', 'date', 'datetime',
    'email', 'url', 'json', 'media', 'relation', 'enum', 'password'
);
CREATE TYPE token_status AS ENUM ('active', 'expired', 'revoked');
CREATE TYPE audit_level AS ENUM ('debug', 'info', 'warn', 'error');
CREATE TYPE audit_type AS ENUM ('auth', 'content', 'media', 'settings', 'user', 'system');

COMMENT ON TYPE user_status IS '用户账户状态：active=正常、inactive=未激活、suspended=已停用';
COMMENT ON TYPE entry_status IS '内容条目状态：draft=草稿、published=已发布、archived=已归档';
COMMENT ON TYPE content_type_kind IS '内容模型类型：collection=集合（如文章列表）、single=单页（如关于页面）';
COMMENT ON TYPE field_type IS '字段类型：text=文本、rich_text=富文本、number=数字、boolean=布尔、date=日期、datetime=时间、email=邮箱、url=链接、json=JSON、media=媒体、relation=关联、enum=枚举、password=密码';
COMMENT ON TYPE token_status IS 'API Token 状态：active=有效、expired=已过期、revoked=已撤销';
COMMENT ON TYPE audit_level IS '审计日志级别：debug=调试、info=信息、warn=警告、error=错误';
COMMENT ON TYPE audit_type IS '审计日志类别：auth=认证、content=内容、media=媒体、settings=设置、user=用户、system=系统';

-- =============================================================================
-- 0. 全局配置表
-- =============================================================================

-- 超级管理员用户表
-- 说明：全局用户表，存放所有系统用户。用户与站点的关系通过 site_users 关联表维护（多对多）
CREATE TABLE system_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
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
COMMENT ON TABLE system_users IS '系统用户表：存放所有用户账户，用户与站点关系通过 site_users 关联表维护';
COMMENT ON COLUMN system_users.id IS '用户唯一标识符(UUID)';
COMMENT ON COLUMN system_users.email IS '用户邮箱（全局唯一）';
COMMENT ON COLUMN system_users.password_hash IS 'bcrypt 加密的密码哈希';
COMMENT ON COLUMN system_users.nickname IS '用户昵称';
COMMENT ON COLUMN system_users.avatar_url IS '头像 URL';
COMMENT ON COLUMN system_users.status IS '账户状态：active/inactive/suspended';
COMMENT ON COLUMN system_users.is_super_admin IS '是否超级管理员（拥有全部权限，不受站点限制）';
COMMENT ON COLUMN system_users.last_login_time IS '最后登录时间';
COMMENT ON COLUMN system_users.last_login_ip IS '最后登录 IP 地址';
COMMENT ON COLUMN system_users.created_time IS '创建时间';
COMMENT ON COLUMN system_users.updated_time IS '更新时间';
COMMENT ON COLUMN system_users.deleted_time IS '软删除时间（非空表示已删除）';

CREATE INDEX idx_system_users_email ON system_users(email);
CREATE INDEX idx_system_users_status ON system_users(status);

-- 全局角色表
CREATE TABLE system_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);
COMMENT ON TABLE system_roles IS '系统角色表：全局角色定义，系统级角色不可删除';
COMMENT ON COLUMN system_roles.id IS '角色唯一标识符';
COMMENT ON COLUMN system_roles.name IS '角色名称（全局唯一）';
COMMENT ON COLUMN system_roles.description IS '角色描述';
COMMENT ON COLUMN system_roles.is_system IS '是否系统角色（系统角色不可删除）';
COMMENT ON COLUMN system_roles.permissions IS '权限列表 JSON：["user:read", "user:write", ...]';
COMMENT ON COLUMN system_roles.created_time IS '创建时间';
COMMENT ON COLUMN system_roles.updated_time IS '更新时间';
COMMENT ON COLUMN system_roles.deleted_time IS '软删除时间';

CREATE INDEX idx_system_roles_name ON system_roles(name);

-- 插件表
CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    version VARCHAR(50) NOT NULL,
    author VARCHAR(200),
    homepage_url TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    config JSONB NOT NULL DEFAULT '{}',
    hooks JSONB NOT NULL DEFAULT '[]',
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);
COMMENT ON TABLE plugins IS '插件表：系统插件和应用插件';
COMMENT ON COLUMN plugins.id IS '插件唯一标识符';
COMMENT ON COLUMN plugins.name IS '插件名称（唯一标识）';
COMMENT ON COLUMN plugins.display_name IS '插件显示名称';
COMMENT ON COLUMN plugins.description IS '插件描述';
COMMENT ON COLUMN plugins.version IS '插件版本号';
COMMENT ON COLUMN plugins.author IS '插件作者';
COMMENT ON COLUMN plugins.homepage_url IS '插件主页 URL';
COMMENT ON COLUMN plugins.is_enabled IS '是否启用';
COMMENT ON COLUMN plugins.is_system IS '是否系统插件（系统插件不可禁用）';
COMMENT ON COLUMN plugins.config IS '插件配置 JSON';
COMMENT ON COLUMN plugins.hooks IS '插件钩子列表 JSON';
COMMENT ON COLUMN plugins.created_time IS '安装时间';
COMMENT ON COLUMN plugins.updated_time IS '更新时间';
COMMENT ON COLUMN plugins.deleted_time IS '软删除时间';

CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_plugins_enabled ON plugins(is_enabled);

-- 审计日志表
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
COMMENT ON TABLE audit_logs IS '审计日志表：记录所有操作行为';
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
COMMENT ON COLUMN audit_logs.created_time IS '日志创建时间';

CREATE INDEX idx_audit_logs_site ON audit_logs(site_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_category ON audit_logs(category);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_time DESC);

-- =============================================================================
-- 1. 站点层
-- =============================================================================

CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    logo_url TEXT,
    favicon_url TEXT,
    config JSONB NOT NULL DEFAULT '{"timezone":"Asia/Shanghai","locale":"zh-CN"}',
    seo JSONB NOT NULL DEFAULT '{}',
    custom_domains JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    tenant_id UUID,
    plan VARCHAR(50) DEFAULT 'free',
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);
COMMENT ON TABLE sites IS '站点表：每个站点是一个独立的 CMS 实例';
COMMENT ON COLUMN sites.id IS '站点唯一标识符';
COMMENT ON COLUMN sites.name IS '站点名称';
COMMENT ON COLUMN sites.slug IS '站点别名（URL 中使用）';
COMMENT ON COLUMN sites.description IS '站点描述';
COMMENT ON COLUMN sites.logo_url IS '站点 Logo URL';
COMMENT ON COLUMN sites.favicon_url IS '站点 Favicon URL';
COMMENT ON COLUMN sites.config IS '站点配置 JSON：timezone 时区、locale 语言等';
COMMENT ON COLUMN sites.seo IS 'SEO 配置 JSON：title、description、keywords 等';
COMMENT ON COLUMN sites.custom_domains IS '自定义域名列表 JSON';
COMMENT ON COLUMN sites.is_active IS '是否激活';
COMMENT ON COLUMN sites.tenant_id IS '租户 ID（多租户模式使用）';
COMMENT ON COLUMN sites.plan IS '套餐：free/pro/enterprise';
COMMENT ON COLUMN sites.created_by IS '创建者用户 ID';
COMMENT ON COLUMN sites.created_time IS '创建时间';
COMMENT ON COLUMN sites.updated_time IS '更新时间';
COMMENT ON COLUMN sites.deleted_time IS '软删除时间';

CREATE INDEX idx_sites_slug ON sites(slug);
CREATE INDEX idx_sites_active ON sites(is_active);
CREATE INDEX idx_sites_tenant ON sites(tenant_id);

CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    channel_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    routing JSONB NOT NULL DEFAULT '{}',
    cache JSONB NOT NULL DEFAULT '{}',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    UNIQUE(site_id, slug)
);
COMMENT ON TABLE channels IS '渠道表：站点的发布渠道（如默认、PC、移动端、API 等）';
COMMENT ON COLUMN channels.id IS '渠道唯一标识符';
COMMENT ON COLUMN channels.site_id IS '所属站点';
COMMENT ON COLUMN channels.name IS '渠道名称';
COMMENT ON COLUMN channels.slug IS '渠道别名';
COMMENT ON COLUMN channels.description IS '渠道描述';
COMMENT ON COLUMN channels.channel_type IS '渠道类型：default/pc/mobile/api 等';
COMMENT ON COLUMN channels.config IS '渠道配置 JSON';
COMMENT ON COLUMN channels.routing IS '路由配置 JSON';
COMMENT ON COLUMN channels.cache IS '缓存配置 JSON';
COMMENT ON COLUMN channels.is_enabled IS '是否启用';
COMMENT ON COLUMN channels.sort_order IS '排序权重';
COMMENT ON COLUMN channels.created_time IS '创建时间';
COMMENT ON COLUMN channels.updated_time IS '更新时间';
COMMENT ON COLUMN channels.deleted_time IS '软删除时间';

CREATE INDEX idx_channels_site ON channels(site_id);
CREATE INDEX idx_channels_type ON channels(channel_type);
CREATE INDEX idx_channels_enabled ON channels(is_enabled);

CREATE TABLE locales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    UNIQUE(site_id, code)
);
COMMENT ON TABLE locales IS '本地化语言表：站点支持的语言/地区';
COMMENT ON COLUMN locales.id IS '语言唯一标识符';
COMMENT ON COLUMN locales.site_id IS '所属站点';
COMMENT ON COLUMN locales.code IS '语言代码：zh-CN/en-US/ja-JP 等';
COMMENT ON COLUMN locales.name IS '语言显示名称';
COMMENT ON COLUMN locales.is_default IS '是否为默认语言';
COMMENT ON COLUMN locales.is_enabled IS '是否启用';
COMMENT ON COLUMN locales.sort_order IS '排序权重';
COMMENT ON COLUMN locales.created_time IS '创建时间';
COMMENT ON COLUMN locales.updated_time IS '更新时间';
COMMENT ON COLUMN locales.deleted_time IS '软删除时间';

CREATE INDEX idx_locales_site ON locales(site_id);
CREATE INDEX idx_locales_default ON locales(site_id, is_default) WHERE is_default = TRUE;

-- =============================================================================
-- 2. 用户权限层
-- =============================================================================

CREATE TABLE site_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]',
    content_permissions JSONB NOT NULL DEFAULT '[]',
    channel_permissions JSONB NOT NULL DEFAULT '[]',
    sort_order INT NOT NULL DEFAULT 0,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    UNIQUE(site_id, name)
);
COMMENT ON TABLE site_roles IS '站点角色表：站点内的权限角色定义';
COMMENT ON COLUMN site_roles.id IS '角色唯一标识符';
COMMENT ON COLUMN site_roles.site_id IS '所属站点';
COMMENT ON COLUMN site_roles.name IS '角色名称';
COMMENT ON COLUMN site_roles.description IS '角色描述';
COMMENT ON COLUMN site_roles.is_system IS '是否系统角色';
COMMENT ON COLUMN site_roles.permissions IS '通用权限列表 JSON';
COMMENT ON COLUMN site_roles.content_permissions IS '内容权限列表 JSON';
COMMENT ON COLUMN site_roles.channel_permissions IS '渠道权限列表 JSON';
COMMENT ON COLUMN site_roles.sort_order IS '排序权重';
COMMENT ON COLUMN site_roles.created_time IS '创建时间';
COMMENT ON COLUMN site_roles.updated_time IS '更新时间';
COMMENT ON COLUMN site_roles.deleted_time IS '软删除时间';

CREATE INDEX idx_site_roles_site ON site_roles(site_id);

CREATE TABLE site_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES system_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES site_roles(id),
    status user_status NOT NULL DEFAULT 'active',
    extra_permissions JSONB NOT NULL DEFAULT '[]',
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ,
    UNIQUE(site_id, user_id)
);
COMMENT ON TABLE site_users IS '用户站点关联表：用户与站点的多对多关系，记录用户在站点中的角色';
COMMENT ON COLUMN site_users.id IS '关联记录唯一标识符';
COMMENT ON COLUMN site_users.site_id IS '站点 ID';
COMMENT ON COLUMN site_users.user_id IS '用户 ID';
COMMENT ON COLUMN site_users.role_id IS '站点角色 ID';
COMMENT ON COLUMN site_users.status IS '在站点中的状态';
COMMENT ON COLUMN site_users.extra_permissions IS '额外权限 JSON（可覆盖角色权限）';
COMMENT ON COLUMN site_users.created_time IS '创建时间';
COMMENT ON COLUMN site_users.updated_time IS '更新时间';
COMMENT ON COLUMN site_users.deleted_time IS '软删除时间';

CREATE INDEX idx_site_users_site ON site_users(site_id);
CREATE INDEX idx_site_users_user ON site_users(user_id);
CREATE INDEX idx_site_users_role ON site_users(role_id);

-- =============================================================================
-- 3. 内容模型层
-- =============================================================================

CREATE TABLE content_types (
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
    deleted_time TIMESTAMPTZ,
    UNIQUE(site_id, slug)
);
COMMENT ON TABLE content_types IS '内容模型表：定义内容的结构（类似 WordPress 的自定义内容类型）';
COMMENT ON COLUMN content_types.id IS '内容模型唯一标识符';
COMMENT ON COLUMN content_types.site_id IS '所属站点';
COMMENT ON COLUMN content_types.name IS '模型名称';
COMMENT ON COLUMN content_types.slug IS '模型别名（API 中使用）';
COMMENT ON COLUMN content_types.description IS '模型描述';
COMMENT ON COLUMN content_types.kind IS '类型：collection=集合、single=单页';
COMMENT ON COLUMN content_types.display_config IS '前端显示配置 JSON';
COMMENT ON COLUMN content_types.api_config IS 'API 配置 JSON：publicRead/publicWrite 等';
COMMENT ON COLUMN content_types.preview_config IS '预览配置 JSON';
COMMENT ON COLUMN content_types.versioning_enabled IS '是否启用版本历史';
COMMENT ON COLUMN content_types.draft_autosave_interval IS '草稿自动保存间隔（秒）';
COMMENT ON COLUMN content_types.is_active IS '是否启用';
COMMENT ON COLUMN content_types.sort_order IS '排序权重';
COMMENT ON COLUMN content_types.created_by IS '创建者';
COMMENT ON COLUMN content_types.created_time IS '创建时间';
COMMENT ON COLUMN content_types.updated_time IS '更新时间';
COMMENT ON COLUMN content_types.deleted_time IS '软删除时间';

CREATE INDEX idx_content_types_site ON content_types(site_id);
CREATE INDEX idx_content_types_slug ON content_types(slug);
CREATE INDEX idx_content_types_active ON content_types(is_active);
CREATE INDEX idx_content_types_kind ON content_types(kind);

CREATE TABLE fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type_id UUID NOT NULL REFERENCES content_types(id) ON DELETE CASCADE,
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
    deleted_time TIMESTAMPTZ,
    UNIQUE(content_type_id, name)
);
COMMENT ON TABLE fields IS '字段表：内容模型的字段定义';
COMMENT ON COLUMN fields.id IS '字段唯一标识符';
COMMENT ON COLUMN fields.content_type_id IS '所属内容模型';
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

CREATE INDEX idx_fields_content_type ON fields(content_type_id);
CREATE INDEX idx_fields_sort ON fields(content_type_id, sort_order);

-- =============================================================================
-- 4. 内容条目层
-- =============================================================================

CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type_id UUID NOT NULL REFERENCES content_types(id) ON DELETE CASCADE,
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
    UNIQUE(content_type_id, locale, id)
);
COMMENT ON TABLE entries IS '内容条目表：实际的内容数据';
COMMENT ON COLUMN entries.id IS '条目唯一标识符';
COMMENT ON COLUMN entries.content_type_id IS '所属内容模型';
COMMENT ON COLUMN entries.site_id IS '所属站点';
COMMENT ON COLUMN entries.locale IS '语言/地区代码';
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
COMMENT ON COLUMN entries.created_by IS '创建者';
COMMENT ON COLUMN entries.created_time IS '创建时间';
COMMENT ON COLUMN entries.updated_time IS '更新时间';
COMMENT ON COLUMN entries.deleted_time IS '软删除时间';

CREATE INDEX idx_entries_type ON entries(content_type_id);
CREATE INDEX idx_entries_site ON entries(site_id);
CREATE INDEX idx_entries_locale ON entries(locale);
CREATE INDEX idx_entries_status ON entries(status);
CREATE INDEX idx_entries_published ON entries(published_time DESC) WHERE status = 'published';
CREATE INDEX idx_entries_sort ON entries(content_type_id, sort_weight);
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
COMMENT ON COLUMN entry_values.created_time IS '创建时间';
COMMENT ON COLUMN entry_values.updated_time IS '更新时间';

CREATE INDEX idx_entry_values_entry ON entry_values(entry_id);
CREATE INDEX idx_entry_values_field ON entry_values(field_id);
CREATE INDEX idx_entry_values_text ON entry_values USING gin(to_tsvector('simple', text_value)) WHERE text_value IS NOT NULL;
CREATE INDEX idx_entry_values_number ON entry_values(number_value) WHERE number_value IS NOT NULL;
CREATE INDEX idx_entry_values_bool ON entry_values(bool_value) WHERE bool_value IS NOT NULL;
CREATE INDEX idx_entry_values_date ON entry_values(date_value) WHERE date_value IS NOT NULL;

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

-- 中文分词支持（可选）
COMMENT ON TABLE entry_versions IS '条目版本表：内容的历史版本记录';
COMMENT ON COLUMN entry_versions.id IS '版本唯一标识符';
COMMENT ON COLUMN entry_versions.entry_id IS '所属条目';
COMMENT ON COLUMN entry_versions.version IS '版本号';
COMMENT ON COLUMN entry_versions.values_snapshot IS '值快照 JSON';
COMMENT ON COLUMN entry_versions.created_by IS '创建者';
COMMENT ON COLUMN entry_versions.created_time IS '创建时间';
COMMENT ON COLUMN entry_versions.change_summary IS '变更摘要';


-- =============================================================================
-- 5. 媒体资产层
-- =============================================================================

CREATE TYPE asset_type AS ENUM ('image', 'video', 'audio', 'document', 'file');
CREATE TYPE asset_visibility AS ENUM ('public', 'private');
CREATE TYPE asset_status AS ENUM ('active', 'inactive', 'deleted');

COMMENT ON TYPE asset_type IS '资源类型：image=图片、video=视频、audio=音频、document=文档、file=文件';
COMMENT ON TYPE asset_visibility IS '资源可见性：public=公开、private=私有';
COMMENT ON TYPE asset_status IS '资源状态：active=正常、inactive=禁用、deleted=已删除';

CREATE TABLE asset_folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
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

CREATE INDEX idx_asset_folders_site ON asset_folders(site_id);
CREATE INDEX idx_asset_folders_parent ON asset_folders(parent_id);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
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
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);
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
COMMENT ON COLUMN assets.created_by IS '上传者';
COMMENT ON COLUMN assets.created_time IS '上传时间';
COMMENT ON COLUMN assets.updated_time IS '更新时间';
COMMENT ON COLUMN assets.deleted_time IS '软删除时间';

CREATE INDEX idx_assets_site ON assets(site_id);
CREATE INDEX idx_assets_folder ON assets(folder_id);
CREATE INDEX idx_assets_type ON assets(type);
CREATE INDEX idx_assets_extension ON assets(extension);
CREATE INDEX idx_assets_slug ON assets(slug);
CREATE INDEX idx_assets_hash ON assets(file_hash) WHERE file_hash IS NOT NULL;
CREATE INDEX idx_assets_created ON assets(created_time DESC);
CREATE INDEX idx_assets_download ON assets(download_count);
CREATE INDEX idx_assets_used ON assets(used_count);

-- =============================================================================
-- 6. API Token 层
-- =============================================================================

CREATE TABLE api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    token_prefix VARCHAR(20) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    scopes JSONB NOT NULL DEFAULT '[]',
    site_scope JSONB NOT NULL DEFAULT '[]',
    channel_scope JSONB NOT NULL DEFAULT '[]',
    allowed_ips INET[],
    rate_limit INT,
    expires_time TIMESTAMPTZ,
    status token_status NOT NULL DEFAULT 'active',
    last_used_time TIMESTAMPTZ,
    last_used_ip INET,
    request_count BIGINT NOT NULL DEFAULT 0,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);
COMMENT ON TABLE api_tokens IS 'API Token 表：用于 Open API 认证';
COMMENT ON COLUMN api_tokens.id IS 'Token 唯一标识符';
COMMENT ON COLUMN api_tokens.site_id IS '所属站点';
COMMENT ON COLUMN api_tokens.name IS 'Token 名称';
COMMENT ON COLUMN api_tokens.description IS 'Token 描述';
COMMENT ON COLUMN api_tokens.token_prefix IS 'Token 前缀（显示用）';
COMMENT ON COLUMN api_tokens.token_hash IS 'Token 哈希（存储用）';
COMMENT ON COLUMN api_tokens.scopes IS '权限范围 JSON';
COMMENT ON COLUMN api_tokens.site_scope IS '站点权限范围 JSON';
COMMENT ON COLUMN api_tokens.channel_scope IS '渠道权限范围 JSON';
COMMENT ON COLUMN api_tokens.allowed_ips IS '允许的 IP 列表';
COMMENT ON COLUMN api_tokens.rate_limit IS '速率限制（请求/分钟）';
COMMENT ON COLUMN api_tokens.expires_time IS '过期时间';
COMMENT ON COLUMN api_tokens.status IS '状态：active/expired/revoked';
COMMENT ON COLUMN api_tokens.last_used_time IS '最后使用时间';
COMMENT ON COLUMN api_tokens.last_used_ip IS '最后使用 IP';
COMMENT ON COLUMN api_tokens.request_count IS '累计请求次数';
COMMENT ON COLUMN api_tokens.created_by IS '创建者';
COMMENT ON COLUMN api_tokens.created_time IS '创建时间';
COMMENT ON COLUMN api_tokens.updated_time IS '更新时间';
COMMENT ON COLUMN api_tokens.deleted_time IS '软删除时间';

CREATE INDEX idx_api_tokens_site ON api_tokens(site_id);
CREATE INDEX idx_api_tokens_hash ON api_tokens(token_hash);
CREATE INDEX idx_api_tokens_status ON api_tokens(status);
CREATE INDEX idx_api_tokens_expires ON api_tokens(expires_time) WHERE expires_time IS NOT NULL;

-- =============================================================================
-- 7. Webhook 层
-- =============================================================================

CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    trigger_events JSONB NOT NULL DEFAULT '[]',
    content_type_ids JSONB,
    url TEXT NOT NULL,
    http_method VARCHAR(10) NOT NULL DEFAULT 'POST',
    headers JSONB NOT NULL DEFAULT '{}',
    secret VARCHAR(255),
    retry_config JSONB NOT NULL DEFAULT '{}',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    success_count INT NOT NULL DEFAULT 0,
    failure_count INT NOT NULL DEFAULT 0,
    last_triggered_time TIMESTAMPTZ,
    last_error TEXT,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_time TIMESTAMPTZ
);
COMMENT ON TABLE webhooks IS 'Webhook 表：事件通知配置';
COMMENT ON COLUMN webhooks.id IS 'Webhook 唯一标识符';
COMMENT ON COLUMN webhooks.site_id IS '所属站点';
COMMENT ON COLUMN webhooks.name IS 'Webhook 名称';
COMMENT ON COLUMN webhooks.description IS '描述';
COMMENT ON COLUMN webhooks.trigger_events IS '触发事件列表 JSON';
COMMENT ON COLUMN webhooks.content_type_ids IS '关联的内容模型 ID 列表';
COMMENT ON COLUMN webhooks.url IS '回调 URL';
COMMENT ON COLUMN webhooks.http_method IS 'HTTP 方法';
COMMENT ON COLUMN webhooks.headers IS '自定义请求头 JSON';
COMMENT ON COLUMN webhooks.secret IS '签名密钥';
COMMENT ON COLUMN webhooks.retry_config IS '重试配置 JSON';
COMMENT ON COLUMN webhooks.is_enabled IS '是否启用';
COMMENT ON COLUMN webhooks.success_count IS '成功次数';
COMMENT ON COLUMN webhooks.failure_count IS '失败次数';
COMMENT ON COLUMN webhooks.last_triggered_time IS '最后触发时间';
COMMENT ON COLUMN webhooks.last_error IS '最后错误信息';
COMMENT ON COLUMN webhooks.created_by IS '创建者';
COMMENT ON COLUMN webhooks.created_time IS '创建时间';
COMMENT ON COLUMN webhooks.updated_time IS '更新时间';
COMMENT ON COLUMN webhooks.deleted_time IS '软删除时间';

CREATE INDEX idx_webhooks_site ON webhooks(site_id);
CREATE INDEX idx_webhooks_enabled ON webhooks(is_enabled);

CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    response_status INT,
    response_body TEXT,
    response_time_ms INT,
    attempt INT NOT NULL DEFAULT 1,
    next_retry_time TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE webhook_deliveries IS 'Webhook 投递记录表：事件发送历史';
COMMENT ON COLUMN webhook_deliveries.id IS '投递记录唯一标识符';
COMMENT ON COLUMN webhook_deliveries.webhook_id IS '所属 Webhook';
COMMENT ON COLUMN webhook_deliveries.event_type IS '事件类型';
COMMENT ON COLUMN webhook_deliveries.payload IS '请求载荷 JSON';
COMMENT ON COLUMN webhook_deliveries.response_status IS '响应状态码';
COMMENT ON COLUMN webhook_deliveries.response_body IS '响应体';
COMMENT ON COLUMN webhook_deliveries.response_time_ms IS '响应时间（毫秒）';
COMMENT ON COLUMN webhook_deliveries.attempt IS '尝试次数';
COMMENT ON COLUMN webhook_deliveries.next_retry_time IS '下次重试时间';
COMMENT ON COLUMN webhook_deliveries.status IS '状态：pending/success/failed';
COMMENT ON COLUMN webhook_deliveries.created_time IS '创建时间';

CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_created ON webhook_deliveries(created_time DESC);

-- =============================================================================
-- 8. 分布式锁
-- =============================================================================

CREATE TABLE distributed_locks (
    lock_key VARCHAR(255) PRIMARY KEY,
    lock_value UUID NOT NULL,
    expires_time TIMESTAMPTZ NOT NULL,
    acquired_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE distributed_locks IS '分布式锁表：用于分布式环境下的并发控制';
COMMENT ON COLUMN distributed_locks.lock_key IS '锁键';
COMMENT ON COLUMN distributed_locks.lock_value IS '锁值（持有者标识）';
COMMENT ON COLUMN distributed_locks.expires_time IS '过期时间';
COMMENT ON COLUMN distributed_locks.acquired_time IS '获取时间';

CREATE INDEX idx_distributed_locks_expires ON distributed_locks(expires_time);

-- =============================================================================
-- 性能优化索引
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_entries_list ON entries(content_type_id, locale, status, published_time DESC);
CREATE INDEX IF NOT EXISTS idx_entries_site_locale ON entries(site_id, locale, status);
CREATE INDEX IF NOT EXISTS idx_assets_site_type ON assets(site_id, type);
CREATE INDEX IF NOT EXISTS idx_assets_site_created ON assets(site_id, created_time DESC);
CREATE INDEX IF NOT EXISTS idx_assets_path ON assets(path);
CREATE INDEX IF NOT EXISTS idx_audit_logs_site_user_time ON audit_logs(site_id, user_id, created_time DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_category_time ON audit_logs(category, created_time DESC);

-- =============================================================================
-- 分布式锁自动清理
-- =============================================================================

CREATE OR REPLACE FUNCTION cleanup_expired_locks()
RETURNS void AS $$
BEGIN
    DELETE FROM distributed_locks WHERE expires_time < NOW();
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS pg_cron;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pg_cron not available: %', SQLERRM;
END;
$$;

CREATE OR REPLACE VIEW active_locks AS
SELECT
    lock_key,
    lock_value,
    acquired_time,
    expires_time,
    EXTRACT(EPOCH FROM (expires_time - NOW()))::INT as remaining_seconds
FROM distributed_locks
WHERE expires_time > NOW();

COMMENT ON VIEW active_locks IS '当前活跃的分布式锁';
COMMENT ON FUNCTION cleanup_expired_locks() IS '清理过期的分布式锁';

-- =============================================================================
-- 初始化数据
-- =============================================================================

INSERT INTO system_roles (id, name, description, is_system, permissions) VALUES
    (gen_random_uuid(), 'Super Admin', '超级管理员，拥有所有权限', TRUE, '["*"]'),
    (gen_random_uuid(), 'Plugin Manager', '插件管理员', TRUE, '["plugins:read", "plugins:write", "plugins:install", "plugins:uninstall"]'),
    (gen_random_uuid(), 'Auditor', '审计员，只读访问', TRUE, '["audit:read"]');

INSERT INTO system_users (id, email, password_hash, nickname, status, is_super_admin) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@contful.com', '$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu', 'Administrator', 'active', TRUE);

-- =============================================================================
-- 触发器
-- =============================================================================

CREATE OR REPLACE FUNCTION update_updated_time_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_time = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_time_column() IS '自动更新 updated_time 时间戳的触发器函数';

CREATE TRIGGER update_system_users_updated_time BEFORE UPDATE ON system_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_system_roles_updated_time BEFORE UPDATE ON system_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_plugins_updated_time BEFORE UPDATE ON plugins
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_sites_updated_time BEFORE UPDATE ON sites
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_channels_updated_time BEFORE UPDATE ON channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_locales_updated_time BEFORE UPDATE ON locales
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_site_roles_updated_time BEFORE UPDATE ON site_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_site_users_updated_time BEFORE UPDATE ON site_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_content_types_updated_time BEFORE UPDATE ON content_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_fields_updated_time BEFORE UPDATE ON fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_entries_updated_time BEFORE UPDATE ON entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_entry_values_updated_time BEFORE UPDATE ON entry_values
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_entry_versions_updated_time BEFORE UPDATE ON entry_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_assets_updated_time BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_api_tokens_updated_time BEFORE UPDATE ON api_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
CREATE TRIGGER update_webhooks_updated_time BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();

-- =============================================================================
-- M2: 配置管理中心（PRE-001）
-- =============================================================================

CREATE TABLE IF NOT EXISTS site_configs (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         UUID         NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    config_key      VARCHAR(128) NOT NULL,
    config_value    TEXT,
    config_type     VARCHAR(32)  NOT NULL DEFAULT 'string',   -- string | number | boolean | json | encrypted
    config_group    VARCHAR(64)  NOT NULL DEFAULT 'default', -- default | storage | mail | oauth | payment | feature | integrity
    is_encrypted    BOOLEAN      NOT NULL DEFAULT FALSE,
    is_readonly     BOOLEAN      NOT NULL DEFAULT FALSE,
    description     VARCHAR(255),
    created_time    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_time    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_by      UUID         REFERENCES system_users(id),
    CONSTRAINT uk_site_config_key UNIQUE (site_id, config_key)
);

CREATE INDEX idx_site_configs_site_id ON site_configs(site_id);
CREATE INDEX idx_site_configs_group ON site_configs(site_id, config_group);
CREATE INDEX idx_site_configs_key ON site_configs(site_id, config_key);

CREATE TRIGGER update_site_configs_updated_time BEFORE UPDATE ON site_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
