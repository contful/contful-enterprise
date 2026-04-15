-- =============================================================================
-- Contful - Headless CMS 数据库架构
-- =============================================================================
-- 版本: v1.0.0
-- 数据库: PostgreSQL 18
-- 最后更新: 2026-04-15
--
-- 架构设计理念:
-- 1. 多站点隔离 - 每个表通过 site_id 实现数据隔离
-- 2. 软删除 - 所有业务表支持软删除 (deleted_at)
-- 3. UUID 主键 - 使用 UUID v7 便于分布式生成和时序排序
-- 4. 时序追踪 - created_at/updated_at 记录数据生命周期
-- 5. 动态内容 - 内容字段使用 JSONB 存储，支持灵活扩展
-- =============================================================================

-- =============================================================================
-- 扩展启用
-- =============================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- 枚举类型定义
-- =============================================================================

-- 用户状态
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended');

-- 内容状态
CREATE TYPE entry_status AS ENUM ('draft', 'published', 'archived');

-- 内容类型类型
CREATE TYPE content_type_kind AS ENUM ('collection', 'single');

-- 字段类型
CREATE TYPE field_type AS ENUM (
    'text',           -- 单行文本
    'rich_text',      -- 富文本 (HTML)
    'number',         -- 数字
    'boolean',        -- 布尔值
    'date',           -- 日期
    'datetime',       -- 日期时间
    'email',          -- 邮箱
    'url',            -- URL
    'json',           -- JSON 对象
    'media',          -- 媒体引用 (关联 assets)
    'relation',       -- 关联关系 (关联其他 entry)
    'enum',           -- 枚举选项
    'password'        -- 密码 (加密存储)
);

-- 资产类型
CREATE TYPE asset_type AS ENUM ('image', 'video', 'audio', 'document', 'other');

-- 资产状态
CREATE TYPE asset_status AS ENUM ('active', 'processing', 'failed', 'deleted');

-- API Token 状态
CREATE TYPE token_status AS ENUM ('active', 'expired', 'revoked');

-- 审计日志级别
CREATE TYPE audit_level AS ENUM ('debug', 'info', 'warn', 'error');

-- 审计日志类型
CREATE TYPE audit_type AS ENUM (
    'auth',           -- 认证相关
    'content',        -- 内容操作
    'media',          -- 媒体操作
    'settings',       -- 设置变更
    'user',           -- 用户管理
    'system'          -- 系统操作
);

-- =============================================================================
-- 0. 全局配置表 (不区分站点)
-- =============================================================================

-- 超级管理员用户表 (跨站点)
CREATE TABLE global_users (
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

CREATE INDEX idx_global_users_email ON global_users(email);
CREATE INDEX idx_global_users_status ON global_users(status);

-- 全局角色表
CREATE TABLE global_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,  -- 系统内置角色不可删除
    permissions JSONB NOT NULL DEFAULT '[]',   -- 全局权限数组
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_global_roles_name ON global_roles(name);

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
    config JSONB NOT NULL DEFAULT '{}',        -- 插件配置
    hooks JSONB NOT NULL DEFAULT '[]',         -- 注册的钩子列表
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_plugins_enabled ON plugins(is_enabled);

-- 审计日志表
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID,                              -- 可为空，表示全局操作
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    level audit_level NOT NULL DEFAULT 'info',
    category audit_type NOT NULL,
    details JSONB,                              -- 操作详情
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_site ON audit_logs(site_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_category ON audit_logs(category);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

-- =============================================================================
-- 1. 站点层 (Site)
-- =============================================================================

-- 站点表
CREATE TABLE sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,         -- URL 友好标识
    description TEXT,
    logo_url TEXT,
    favicon_url TEXT,

    -- 站点配置
    config JSONB NOT NULL DEFAULT '{
        "timezone": "Asia/Shanghai",
        "locale": "zh-CN",
        "dateFormat": "YYYY-MM-DD",
        "datetimeFormat": "YYYY-MM-DD HH:mm:ss"
    }',

    -- SEO 默认配置
    seo JSONB NOT NULL DEFAULT '{
        "defaultTitle": "",
        "defaultDescription": "",
        "ogImage": "",
        "twitterCard": "summary_large_image"
    }',

    -- 自定义域名
    custom_domains JSONB NOT NULL DEFAULT '[]',

    -- 状态
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- 租户信息 (Phase 3 多租户)
    tenant_id UUID,
    plan VARCHAR(50) DEFAULT 'free',

    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_sites_slug ON sites(slug);
CREATE INDEX idx_sites_active ON sites(is_active);
CREATE INDEX idx_sites_tenant ON sites(tenant_id);

-- 渠道表 (终端)
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,

    -- 渠道类型
    channel_type VARCHAR(50) NOT NULL,         -- web, ios, android, miniprogram, other

    -- 渠道特定配置
    config JSONB NOT NULL DEFAULT '{}',         -- API 密钥、CDN 设置、推送配置等

    -- 路由配置
    routing JSONB NOT NULL DEFAULT '{
        "pathPrefix": "",
        "homepageSlug": "home"
    }',

    -- 缓存配置
    cache JSONB NOT NULL DEFAULT '{
        "enabled": false,
        "ttl": 300,
        "strategy": "no-cache"
    }',

    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(site_id, slug)
);

CREATE INDEX idx_channels_site ON channels(site_id);
CREATE INDEX idx_channels_type ON channels(channel_type);
CREATE INDEX idx_channels_enabled ON channels(is_enabled);

-- 语言/本地化表
CREATE TABLE locales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    code VARCHAR(20) NOT NULL,                 -- zh-CN, en-US, ja-JP
    name VARCHAR(100) NOT NULL,                 -- 简体中文, English, 日本語
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(site_id, code)
);

CREATE INDEX idx_locales_site ON locales(site_id);
CREATE INDEX idx_locales_default ON locales(site_id, is_default) WHERE is_default = TRUE;

-- =============================================================================
-- 2. 用户权限层 (Site Users & Roles)
-- =============================================================================

-- 站点用户关联表 (用户-站点多对多，包含站点级角色)
CREATE TABLE site_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES global_users(id) ON DELETE CASCADE,

    role_id UUID NOT NULL REFERENCES site_roles(id),
    status user_status NOT NULL DEFAULT 'active',

    -- 站点级个人权限覆盖
    extra_permissions JSONB NOT NULL DEFAULT '[]',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(site_id, user_id)
);

CREATE INDEX idx_site_users_site ON site_users(site_id);
CREATE INDEX idx_site_users_user ON site_users(user_id);
CREATE INDEX idx_site_users_role ON site_users(role_id);

-- 站点角色表
CREATE TABLE site_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,

    -- 权限配置
    permissions JSONB NOT NULL DEFAULT '[]',   -- 权限数组

    -- 内容类型级权限
    content_permissions JSONB NOT NULL DEFAULT '[]',  -- {"article": ["read", "write"], "page": ["read"]}

    -- 渠道级权限
    channel_permissions JSONB NOT NULL DEFAULT '[]',

    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(site_id, name)
);

CREATE INDEX idx_site_roles_site ON site_roles(site_id);

-- =============================================================================
-- 3. 内容模型层 (Content Types & Fields)
-- =============================================================================

-- 内容类型表
CREATE TABLE content_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,

    name VARCHAR(200) NOT NULL,                -- 显示名称
    slug VARCHAR(100) NOT NULL,                -- API 标识
    description TEXT,

    -- 类型: collection (多条) vs single (单条，如首页配置)
    kind content_type_kind NOT NULL DEFAULT 'collection',

    -- 显示配置
    display_config JSONB NOT NULL DEFAULT '{
        "icon": "article",
        "color": "#1890ff",
        "sortable": true,
        "creatable": true,
        "editable": true,
        "deletable": true
    }',

    -- API 配置
    api_config JSONB NOT NULL DEFAULT '{
        "publicRead": false,
        "publicWrite": false,
        "draftVisible": false,
        "graphqlEnabled": true,
        "pagination": {"defaultLimit": 20, "maxLimit": 100}
    }',

    -- 预览配置
    preview_config JSONB NOT NULL DEFAULT '{
        "enabled": false,
        "urlTemplate": ""
    }',

    -- 版本控制
    versioning_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    draft_autosave_interval INT,                -- 秒，NULL 表示禁用

    -- 站点级唯一 (slug 在同一站点内唯一)
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,

    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(site_id, slug)
);

CREATE INDEX idx_content_types_site ON content_types(site_id);
CREATE INDEX idx_content_types_slug ON content_types(slug);
CREATE INDEX idx_content_types_active ON content_types(is_active);
CREATE INDEX idx_content_types_kind ON content_types(kind);

-- 字段定义表
CREATE TABLE fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type_id UUID NOT NULL REFERENCES content_types(id) ON DELETE CASCADE,

    name VARCHAR(100) NOT NULL,                 -- 字段名 (英文，API 用)
    label VARCHAR(200) NOT NULL,                -- 显示名称
    description TEXT,

    field_type field_type NOT NULL,

    -- 字段配置 (JSONB，根据 field_type 不同结构不同)
    config JSONB NOT NULL DEFAULT '{}',        -- 通用配置

    -- 字段验证规则
    validation JSONB NOT NULL DEFAULT '{}',    -- required, min, max, pattern 等

    -- 显示配置
    display JSONB NOT NULL DEFAULT '{
        "width": "100%",
        "position": "main",
        "visible": true,
        "readonly": false
    }',

    -- 默认值
    default_value JSONB,

    -- 排序
    sort_order INT NOT NULL DEFAULT 0,

    -- 条件显示
    conditional_display JSONB,                  -- {"field": "type", "operator": "equals", "value": "vip"}

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE(content_type_id, name)
);

CREATE INDEX idx_fields_content_type ON fields(content_type_id);
CREATE INDEX idx_fields_sort ON fields(content_type_id, sort_order);

-- 字段类型特定配置示例 (存于 config 字段)

-- text 类型:
-- {"maxLength": 255, "placeholder": "", "prefix": "", "suffix": ""}

-- rich_text 类型:
-- {"minHeight": 200, "maxHeight": 600, "toolbar": ["bold", "italic", "link"], "markdown": false}

-- number 类型:
-- {"min": 0, "max": 100, "step": 1, "decimals": 0, "unit": ""}

-- enum 类型:
-- {"options": [{"value": "a", "label": "选项A"}, {"value": "b", "label": "选项B"}], "multiple": false}

-- media 类型:
-- {"maxCount": 1, "allowedTypes": ["image"], "maxSize": 5242880}

-- relation 类型:
-- {"targetTypeId": "uuid", "relationType": "many_to_one", "maxCount": 10}

-- =============================================================================
-- 4. 内容条目层 (Content Entries)
-- =============================================================================

-- 内容条目表 (核心数据表)
CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_type_id UUID NOT NULL REFERENCES content_types(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES sites(id),
    locale VARCHAR(20) NOT NULL DEFAULT 'zh-CN',

    -- 状态管理
    status entry_status NOT NULL DEFAULT 'draft',

    -- 版本控制
    version INT NOT NULL DEFAULT 1,             -- 当前版本号
    version_history JSONB,                     -- {"1": "uuid1", "2": "uuid2"}

    -- 草稿 vs 发布的数据分离
    -- published_at: 发布时间
    published_at TIMESTAMPTZ,
    published_by UUID,

    -- 关联条目 (用于 relation 字段)
    relations JSONB NOT NULL DEFAULT '[]',      -- [{"fieldId": "uuid", "entryIds": ["uuid1", "uuid2"]}]

    -- SEO 字段 (可选)
    seo_title VARCHAR(255),
    seo_description TEXT,
    seo_keywords TEXT[],

    -- 排序权重 (用于 collection 类型的列表排序)
    sort_weight INT NOT NULL DEFAULT 0,

    -- 创建者
    created_by UUID,

    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- 复合索引
    UNIQUE(content_type_id, locale, id)  -- 允许同一类型同一语言多条
);

CREATE INDEX idx_entries_type ON entries(content_type_id);
CREATE INDEX idx_entries_site ON entries(site_id);
CREATE INDEX idx_entries_locale ON entries(locale);
CREATE INDEX idx_entries_status ON entries(status);
CREATE INDEX idx_entries_published ON entries(published_at DESC) WHERE status = 'published';
CREATE INDEX idx_entries_sort ON entries(content_type_id, sort_weight);
CREATE INDEX idx_entries_created_by ON entries(created_by);
CREATE INDEX idx_entries_deleted ON entries(deleted_at) WHERE deleted_at IS NULL;

-- 内容值表 (存储字段的实际值)
CREATE TABLE entry_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,

    -- 字段值 (根据字段类型不同，存储格式不同)
    value JSONB NOT NULL,                       -- 灵活存储各种类型值

    -- 索引优化: 对于需要精确查询的字段，建立额外索引
    -- text_value 存储文本类型用于全文搜索
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
CREATE INDEX idx_entry_values_text ON entry_values USING gin(to_tsvector('simple', text_value)) WHERE text_value IS NOT NULL;
CREATE INDEX idx_entry_values_number ON entry_values(number_value) WHERE number_value IS NOT NULL;
CREATE INDEX idx_entry_values_bool ON entry_values(bool_value) WHERE bool_value IS NOT NULL;
CREATE INDEX idx_entry_values_date ON entry_values(date_value) WHERE date_value IS NOT NULL;

-- 内容版本历史表 (可选，versioning_enabled = true 时使用)
CREATE TABLE entry_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    version INT NOT NULL,

    -- 快照该版本的完整数据
    values_snapshot JSONB NOT NULL,             -- 该版本的所有字段值

    -- 版本元信息
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    change_summary TEXT,                        -- 变更摘要

    UNIQUE(entry_id, version)
);

CREATE INDEX idx_entry_versions_entry ON entry_versions(entry_id);
CREATE INDEX idx_entry_versions_created ON entry_versions(created_at DESC);

-- =============================================================================
-- 5. 媒体资产层 (Media Assets)
-- =============================================================================

-- 媒体资产表
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),

    -- 文件元信息
    filename VARCHAR(255) NOT NULL,             -- 原始文件名
    original_filename VARCHAR(255),              -- 用户上传时的文件名
    mimetype VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,

    -- 存储信息
    storage_provider VARCHAR(50) NOT NULL DEFAULT 'local',  -- local, s3, oss, cos
    storage_path VARCHAR(512) NOT NULL,         -- 存储路径
    storage_url TEXT,                           -- 访问 URL

    -- 文件哈希 (用于去重)
    file_hash VARCHAR(64),                      -- SHA-256

    -- 资产类型
    asset_type asset_type NOT NULL,

    -- 媒体特有属性
    width INT,                                  -- 图片/视频宽度
    height INT,                                 -- 图片/视频高度
    duration_sec NUMERIC(10,3),                 -- 视频/音频时长
    color_space VARCHAR(50),
    has_alpha BOOLEAN DEFAULT FALSE,

    -- 缩略图
    thumbnail_path VARCHAR(512),
    thumbnail_url TEXT,

    -- 元数据
    metadata JSONB NOT NULL DEFAULT '{}',       -- EXIF、分辨率、编码等

    -- 状态
    status asset_status NOT NULL DEFAULT 'active',

    -- 使用统计
    usage_count INT NOT NULL DEFAULT 0,

    -- 上传信息
    uploaded_by UUID,
    ip_address INET,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_assets_site ON assets(site_id);
CREATE INDEX idx_assets_type ON assets(asset_type);
CREATE INDEX idx_assets_mimetype ON assets(mimetype);
CREATE INDEX idx_assets_hash ON assets(file_hash) WHERE file_hash IS NOT NULL;
CREATE INDEX idx_assets_status ON assets(status);
CREATE INDEX idx_assets_created ON assets(created_at DESC);
CREATE INDEX idx_assets_uploaded_by ON assets(uploaded_by);

-- =============================================================================
-- 6. API Token 层 (Developer Access)
-- =============================================================================

-- API Token 表
CREATE TABLE api_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),

    name VARCHAR(200) NOT NULL,                -- Token 名称
    description TEXT,

    -- Token 值 (仅创建时显示一次)
    token_prefix VARCHAR(20) NOT NULL,          -- ctf_xxxx
    token_hash VARCHAR(64) NOT NULL UNIQUE,     -- SHA-256 Hash

    -- 权限范围 (JSONB)
    scopes JSONB NOT NULL DEFAULT '[]',        -- ["read:article", "write:article", "read:page"]

    -- 站点范围 (可以访问哪些站点)
    site_scope JSONB NOT NULL DEFAULT '[]',    -- ["site_id_1", "site_id_2"] 或 ["*"] 表示全部

    -- 渠道范围
    channel_scope JSONB NOT NULL DEFAULT '[]', -- ["channel_slug_1"]

    -- 限制
    allowed_ips INET[],                         -- IP 白名单，NULL 表示不限
    rate_limit INT,                             -- 每分钟请求数限制

    -- 有效期
    expires_at TIMESTAMPTZ,                     -- NULL 表示永不过期

    -- 状态
    status token_status NOT NULL DEFAULT 'active',

    -- 使用统计
    last_used_at TIMESTAMPTZ,
    last_used_ip INET,
    request_count BIGINT NOT NULL DEFAULT 0,

    -- 创建信息
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_api_tokens_site ON api_tokens(site_id);
CREATE INDEX idx_api_tokens_hash ON api_tokens(token_hash);
CREATE INDEX idx_api_tokens_status ON api_tokens(status);
CREATE INDEX idx_api_tokens_expires ON api_tokens(expires_at) WHERE expires_at IS NOT NULL;

-- =============================================================================
-- 7. Webhook 表 (Phase 2)
-- =============================================================================

CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),

    name VARCHAR(200) NOT NULL,
    description TEXT,

    -- 触发条件
    trigger_events JSONB NOT NULL DEFAULT '[]',  -- ["entry.publish", "entry.delete", "media.upload"]
    content_type_ids JSONB,                        -- NULL 表示所有类型

    -- 目标
    url TEXT NOT NULL,
    http_method VARCHAR(10) NOT NULL DEFAULT 'POST',
    headers JSONB NOT NULL DEFAULT '{}',

    -- 安全
    secret VARCHAR(255),                          -- 用于签名验证

    -- 重试配置
    retry_config JSONB NOT NULL DEFAULT '{
        "enabled": true,
        "maxRetries": 3,
        "retryDelay": 60
    }',

    -- 状态
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    -- 统计
    success_count INT NOT NULL DEFAULT 0,
    failure_count INT NOT NULL DEFAULT 0,
    last_triggered_at TIMESTAMPTZ,
    last_error TEXT,

    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_webhooks_site ON webhooks(site_id);
CREATE INDEX idx_webhooks_enabled ON webhooks(is_enabled);

-- Webhook 投递日志
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,

    -- 投递信息
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,

    -- HTTP 响应
    response_status INT,
    response_body TEXT,
    response_time_ms INT,

    -- 重试
    attempt INT NOT NULL DEFAULT 1,
    next_retry_at TIMESTAMPTZ,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, success, failed, retrying

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_created ON webhook_deliveries(created_at DESC);

-- =============================================================================
-- 8. 缓存表 (Redis 补充，用于分布式锁等)
-- =============================================================================

-- PostgreSQL 实现的轻量级分布式锁 (Redis 不可用时的备选)
CREATE TABLE distributed_locks (
    lock_key VARCHAR(255) PRIMARY KEY,
    lock_value UUID NOT NULL,                    -- 持有者的 UUID
    expires_at TIMESTAMPTZ NOT NULL,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_distributed_locks_expires ON distributed_locks(expires_at);

-- =============================================================================
-- 初始化数据
-- =============================================================================

-- 插入全局内置角色
INSERT INTO global_roles (id, name, description, is_system, permissions) VALUES
    (gen_random_uuid(), 'Super Admin', '超级管理员，拥有所有权限', TRUE, '["*"]'),
    (gen_random_uuid(), 'Plugin Manager', '插件管理员', TRUE, '["plugins:read", "plugins:write", "plugins:install", "plugins:uninstall"]'),
    (gen_random_uuid(), 'Auditor', '审计员，只读访问', TRUE, '["audit:read"]');

-- =============================================================================
-- 触发器函数
-- =============================================================================

-- 自动更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有业务表创建 updated_at 触发器
CREATE TRIGGER update_global_users_updated_at BEFORE UPDATE ON global_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_global_roles_updated_at BEFORE UPDATE ON global_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sites_updated_at BEFORE UPDATE ON sites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_channels_updated_at BEFORE UPDATE ON channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_locales_updated_at BEFORE UPDATE ON locales
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_site_users_updated_at BEFORE UPDATE ON site_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_site_roles_updated_at BEFORE UPDATE ON site_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_content_types_updated_at BEFORE UPDATE ON content_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_fields_updated_at BEFORE UPDATE ON fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entries_updated_at BEFORE UPDATE ON entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_entry_values_updated_at BEFORE UPDATE ON entry_values
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_assets_updated_at BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_api_tokens_updated_at BEFORE UPDATE ON api_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 清理函数 (用于测试或数据迁移)
-- =============================================================================

-- 清理过期锁
CREATE OR REPLACE FUNCTION cleanup_expired_locks()
RETURNS INT AS $$
DECLARE
    deleted_count INT;
BEGIN
    DELETE FROM distributed_locks WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 清理软删除超过 30 天的数据 (可配置)
CREATE OR REPLACE FUNCTION cleanup_soft_deleted(days_to_keep INT DEFAULT 30)
RETURNS TABLE(table_name TEXT, deleted_count BIGINT) AS $$
DECLARE
    row_count BIGINT;
BEGIN
    -- 注意: 实际删除需要考虑外键约束和级联删除
    -- 这里仅作为示例，实际使用需要更谨慎的设计
    RAISE NOTICE 'Cleanup function called with days_to_keep = %', days_to_keep;
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 注释说明
-- =============================================================================

COMMENT ON TABLE sites IS '站点表 - 多站点架构的核心，每条记录代表一个独立站点';
COMMENT ON TABLE channels IS '渠道表 - 每个站点的展示终端，如 Web、iOS、Android、小程序';
COMMENT ON TABLE locales IS '多语言表 - 每个站点支持的语言版本';
COMMENT ON TABLE content_types IS '内容类型表 - 定义内容的结构，如文章、产品、页面';
COMMENT ON TABLE fields IS '字段定义表 - 每个内容类型的字段配置';
COMMENT ON TABLE entries IS '内容条目表 - 实际的内容数据';
COMMENT ON TABLE entry_values IS '字段值表 - 每个条目的字段实际值';
COMMENT ON TABLE assets IS '媒体资产表 - 上传的文件元数据';
COMMENT ON TABLE api_tokens IS 'API Token 表 - 第三方开发者接入凭证';
COMMENT ON TABLE webhooks IS 'Webhook 表 - 事件通知配置';
