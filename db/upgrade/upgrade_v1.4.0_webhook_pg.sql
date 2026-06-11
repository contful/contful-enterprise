-- =============================================================================
-- Contful 升级脚本 v1.4.0 — 新增 Webhook 功能
-- =============================================================================
-- 用法: psql -d <database> -f db/upgrade_v1.4.0_webhook.sql
-- 幂等设计: 全部使用 IF NOT EXISTS / WHERE NOT EXISTS
-- =============================================================================

-- =============================================================================
-- 1. Webhook 配置表
-- =============================================================================
CREATE TABLE IF NOT EXISTS contful_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id UUID NOT NULL REFERENCES contful_sites(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(2000) NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    secret VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_site ON contful_webhooks(site_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON contful_webhooks(is_active);

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_contful_webhooks_updated_time'
    ) THEN
        CREATE TRIGGER update_contful_webhooks_updated_time
            BEFORE UPDATE ON contful_webhooks
            FOR EACH ROW EXECUTE FUNCTION update_updated_time_column();
    END IF;
END $$;

COMMENT ON TABLE contful_webhooks IS 'Webhook 配置表：内容事件通知的目标 URL 配置';
COMMENT ON COLUMN contful_webhooks.id IS 'Webhook 唯一标识符';
COMMENT ON COLUMN contful_webhooks.site_id IS '所属站点';
COMMENT ON COLUMN contful_webhooks.name IS 'Webhook 名称';
COMMENT ON COLUMN contful_webhooks.url IS '目标 URL';
COMMENT ON COLUMN contful_webhooks.events IS '订阅的事件类型列表';
COMMENT ON COLUMN contful_webhooks.secret IS 'HMAC-SHA256 签名密钥（可选）';
COMMENT ON COLUMN contful_webhooks.is_active IS '是否启用';
COMMENT ON COLUMN contful_webhooks.created_time IS '创建时间';
COMMENT ON COLUMN contful_webhooks.updated_time IS '更新时间';

-- =============================================================================
-- 2. Webhook 投递记录表
-- =============================================================================
CREATE TABLE IF NOT EXISTS contful_webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES contful_webhooks(id) ON DELETE CASCADE,
    event VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    response_status INT,
    response_body TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt INT NOT NULL DEFAULT 1,
    error_message TEXT,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON contful_webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON contful_webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_created ON contful_webhook_deliveries(created_time DESC);

COMMENT ON TABLE contful_webhook_deliveries IS 'Webhook 投递记录表：每次事件通知的发送记录和结果';
COMMENT ON COLUMN contful_webhook_deliveries.id IS '投递记录唯一标识符';
COMMENT ON COLUMN contful_webhook_deliveries.webhook_id IS '关联的 Webhook 配置';
COMMENT ON COLUMN contful_webhook_deliveries.event IS '事件类型';
COMMENT ON COLUMN contful_webhook_deliveries.payload IS '发送的 JSON 请求体快照';
COMMENT ON COLUMN contful_webhook_deliveries.response_status IS '目标服务器响应状态码';
COMMENT ON COLUMN contful_webhook_deliveries.response_body IS '响应内容（最多 1KB）';
COMMENT ON COLUMN contful_webhook_deliveries.status IS '投递状态：pending/success/failed';
COMMENT ON COLUMN contful_webhook_deliveries.attempt IS '第几次尝试';
COMMENT ON COLUMN contful_webhook_deliveries.error_message IS '失败原因';
COMMENT ON COLUMN contful_webhook_deliveries.created_time IS '创建时间';

-- =============================================================================
-- 3. Webhook 权限注册（种子数据）
-- =============================================================================

-- 权限分组
INSERT INTO contful_system_permission_groups (id, group_key, label, label_en, sort_order)
SELECT '00000000-0000-0000-0000-000000000311'::uuid, 'webhook', 'Webhook', 'Webhooks', 10
WHERE NOT EXISTS (
    SELECT 1 FROM contful_system_permission_groups WHERE group_key = 'webhook'
);

-- 权限项
INSERT INTO contful_system_permissions (group_id, action, label, label_en, sort_order)
SELECT g.id, t.action, t.label, t.label_en, t.sort_order
FROM (VALUES
    ('00000000-0000-0000-0000-000000000311'::uuid, 'read',  '查看 Webhook',  'View Webhooks',  0),
    ('00000000-0000-0000-0000-000000000311'::uuid, 'write', '管理 Webhook',  'Manage Webhooks', 1)
) AS t(group_id, action, label, label_en, sort_order)
JOIN contful_system_permission_groups g ON g.id = t.group_id::uuid
WHERE NOT EXISTS (
    SELECT 1 FROM contful_system_permissions p WHERE p.group_id = g.id AND p.action = t.action
);

-- 超级管理员角色追加 webhook 权限
UPDATE contful_system_roles
SET permissions = (
    SELECT jsonb_agg(DISTINCT p) FROM (
        SELECT jsonb_array_elements_text(permissions) AS p
        UNION ALL
        SELECT 'webhook:read' UNION ALL SELECT 'webhook:write'
    ) t
),
updated_time = NOW()
WHERE id = '00000000-0000-0000-0000-000000000101'::uuid
AND NOT (permissions::text LIKE '%webhook:read%');

-- =============================================================================
-- 验证
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '✅ Webhook 升级 v1.4.0 完成';
    RAISE NOTICE '   - contful_webhooks 表';
    RAISE NOTICE '   - contful_webhook_deliveries 表';
    RAISE NOTICE '   - webhook:read / webhook:write 权限';
END $$;
