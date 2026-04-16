-- =============================================================================
-- Contful - Headless CMS 数据库初始化脚本
-- 版本: v1.0.0
-- 数据库: PostgreSQL 18
-- 使用: psql -h <host> -U <user> -d <db> -f init.sql
-- =============================================================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

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
CREATE TYPE asset_type AS ENUM ('image', 'video', 'audio', 'document', 'other');
CREATE TYPE asset_status AS ENUM ('active', 'processing', 'failed', 'deleted');
CREATE TYPE token_status AS ENUM ('active', 'expired', 'revoked');
CREATE TYPE audit_level AS ENUM ('debug', 'info', 'warn', 'error');
CREATE TYPE audit_type AS ENUM ('auth', 'content', 'media', 'settings', 'user', 'system');

-- =============================================================================
-- 0. 全局配置表
-- =============================================================================

-- 超级管理员用户表
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
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    permissions JSONB NOT NULL DEFAULT '[]',
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
    config JSONB NOT NULL DEFAULT '{}',
    hooks JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_logs_site ON audit_logs(site_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_category ON audit_logs(category);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(site_id, slug)
);
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(site_id, code)
);
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(site_id, name)
);
CREATE INDEX idx_site_roles_site ON site_roles(site_id);

CREATE TABLE site_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES global_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES site_roles(id),
    status user_status NOT NULL DEFAULT 'active',
    extra_permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(site_id, user_id)
);
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(site_id, slug)
);
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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(content_type_id, name)
);
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
    UNIQUE(content_type_id, locale, id)
);
CREATE INDEX idx_entries_type ON entries(content_type_id);
CREATE INDEX idx_entries_site ON entries(site_id);
CREATE INDEX idx_entries_locale ON entries(locale);
CREATE INDEX idx_entries_status ON entries(status);
CREATE INDEX idx_entries_published ON entries(published_at DESC) WHERE status = 'published';
CREATE INDEX idx_entries_sort ON entries(content_type_id, sort_weight);
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
-- 默认使用 simple 分词器（不支持中文）
CREATE INDEX idx_entry_values_text ON entry_values USING gin(to_tsvector('simple', text_value)) WHERE text_value IS NOT NULL;
CREATE INDEX idx_entry_values_number ON entry_values(number_value) WHERE number_value IS NOT NULL;
CREATE INDEX idx_entry_values_bool ON entry_values(bool_value) WHERE bool_value IS NOT NULL;
CREATE INDEX idx_entry_values_date ON entry_values(date_value) WHERE date_value IS NOT NULL;

-- =============================================================================
-- 中文分词支持（可选，需安装扩展）
-- =============================================================================
-- 推荐方案:
-- 1. pg_jieba: SELECT * FROM jieba_char_dict(); -- 查看词典
-- 2. zhparser: 需要 SCWS 算法支持
--
-- 启用中文分词的步骤:
-- a) 安装扩展: CREATE EXTENSION pg_jieba; 或 CREATE EXTENSION zhparser;
-- b) 创建文本搜索配置:
--    CREATE TEXT SEARCH CONFIGURATION zh_cn (parser = pg_jieba);
-- c) 创建中文搜索索引:
--    CREATE INDEX idx_entry_values_text_zh ON entry_values
--        USING gin(to_tsvector('zh_cn', text_value)) WHERE text_value IS NOT NULL;
-- d) 应用层切换搜索配置
--
-- 注意: 中文分词扩展会增加存储和索引维护开销，按需启用

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
-- 5. 媒体资产层
-- =============================================================================

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES sites(id),
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255),
    mimetype VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_provider VARCHAR(50) NOT NULL DEFAULT 'local',
    storage_path VARCHAR(512) NOT NULL,
    storage_url TEXT,
    file_hash VARCHAR(64),
    asset_type asset_type NOT NULL,
    width INT,
    height INT,
    duration_sec NUMERIC(10,3),
    color_space VARCHAR(50),
    has_alpha BOOLEAN DEFAULT FALSE,
    thumbnail_path VARCHAR(512),
    thumbnail_url TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    status asset_status NOT NULL DEFAULT 'active',
    usage_count INT NOT NULL DEFAULT 0,
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
    expires_at TIMESTAMPTZ,
    status token_status NOT NULL DEFAULT 'active',
    last_used_at TIMESTAMPTZ,
    last_used_ip INET,
    request_count BIGINT NOT NULL DEFAULT 0,
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
    last_triggered_at TIMESTAMPTZ,
    last_error TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
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
    next_retry_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_created ON webhook_deliveries(created_at DESC);

-- =============================================================================
-- 8. 分布式锁
-- =============================================================================

CREATE TABLE distributed_locks (
    lock_key VARCHAR(255) PRIMARY KEY,
    lock_value UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_distributed_locks_expires ON distributed_locks(expires_at);

-- =============================================================================
-- 性能优化索引
-- =============================================================================

-- 内容列表查询（最常用）
CREATE INDEX IF NOT EXISTS idx_entries_list
    ON entries(content_type_id, locale, status, published_at DESC);

-- 站点内容查询
CREATE INDEX IF NOT EXISTS idx_entries_site_locale
    ON entries(site_id, locale, status);

-- assets 表索引优化
CREATE INDEX IF NOT EXISTS idx_assets_site_status_date
    ON assets(site_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_assets_site_type
    ON assets(site_id, asset_type);
CREATE INDEX IF NOT EXISTS idx_assets_storage_path
    ON assets(storage_path);

-- audit_logs 索引优化
CREATE INDEX IF NOT EXISTS idx_audit_logs_site_user_time
    ON audit_logs(site_id, user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_category_time
    ON audit_logs(category, created_at DESC);

-- =============================================================================
-- 分布式锁自动清理函数
-- =============================================================================

CREATE OR REPLACE FUNCTION cleanup_expired_locks()
RETURNS void AS $$
BEGIN
    DELETE FROM distributed_locks WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- 自动清理任务（需要 pg_cron 扩展支持）
-- 如果 pg_cron 不可用，可忽略此错误，或改用应用层定时清理
DO $$
BEGIN
    -- 每分钟清理一次过期锁
    CREATE EXTENSION IF NOT EXISTS pg_cron;
    -- 注意: pg_cron 需要配置 cron.database 连接数据库
    -- 如需启用，在 postgresql.conf 中添加: cron.database_name = 'contful'
    -- 任务创建可能失败，忽略即可（pg_cron 未配置时）
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'pg_cron not available, automatic lock cleanup disabled: %', SQLERRM;
END;
$$;

-- 可选：创建视图查看当前活跃锁
CREATE OR REPLACE VIEW active_locks AS
SELECT
    lock_key,
    lock_value,
    acquired_at,
    expires_at,
    EXTRACT(EPOCH FROM (expires_at - NOW()))::INT as remaining_seconds
FROM distributed_locks
WHERE expires_at > NOW();

COMMENT ON VIEW active_locks IS '当前活跃的分布式锁';
COMMENT ON FUNCTION cleanup_expired_locks() IS '清理过期的分布式锁，由 pg_cron 每分钟调用或应用层调用';

-- =============================================================================
-- 初始化数据
-- =============================================================================

INSERT INTO global_roles (id, name, description, is_system, permissions) VALUES
    (gen_random_uuid(), 'Super Admin', '超级管理员，拥有所有权限', TRUE, '["*"]'),
    (gen_random_uuid(), 'Plugin Manager', '插件管理员', TRUE, '["plugins:read", "plugins:write", "plugins:install", "plugins:uninstall"]'),
    (gen_random_uuid(), 'Auditor', '审计员，只读访问', TRUE, '["audit:read"]');

-- =============================================================================
-- 默认管理员账号
-- =============================================================================
-- 邮箱: admin@contful.com
-- 密码: contful@com
-- 密码哈希: bcrypt(cost=10) of "contful@com"

INSERT INTO global_users (id, email, password_hash, nickname, status, is_super_admin) VALUES
    ('00000000-0000-0000-0000-000000000001', 'admin@contful.com', '$2a$10$65v1ImEvTC/GCPqBctpsiuAy/J04X1BHX7AKBufYsSV7kvuNSfJMu', 'Administrator', 'active', TRUE);

-- =============================================================================
-- 触发器
-- =============================================================================

-- 统一 updated_at 自动更新触发器函数
-- 注意: 软删除表必须保留 updated_at 字段，否则触发器会报错
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

COMMENT ON FUNCTION update_updated_at_column() IS '自动更新 updated_at 时间戳的触发器函数';

-- 注意:
-- 1. 所有软删除表都有 updated_at 字段，保证触发器正常工作
-- 2. 触发器在 BEFORE UPDATE 时执行，不记录变更历史（由 entry_versions 表处理内容版本）
-- 3. 批量更新时触发器仍会生效，但性能影响可忽略（仅更新一列）

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
CREATE TRIGGER update_site_roles_updated_at BEFORE UPDATE ON site_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_site_users_updated_at BEFORE UPDATE ON site_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_content_types_updated_at BEFORE UPDATE ON content_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_fields_updated_at BEFORE UPDATE ON fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entries_updated_at BEFORE UPDATE ON entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entry_values_updated_at BEFORE UPDATE ON entry_values
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entry_versions_updated_at BEFORE UPDATE ON entry_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_assets_updated_at BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_api_tokens_updated_at BEFORE UPDATE ON api_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
