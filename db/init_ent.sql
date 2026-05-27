-- =============================================================================
-- Contful Enterprise — 企业版增量初始化脚本
-- 版本: v2.0.0-ent
-- 数据库: PostgreSQL 14+
--
-- 使用方式（开发环境）：
--   1. createdb contful_ent                         ← 创建企业版独立数据库
--   2. psql -d contful_ent -f db/init_pg.sql        ← 社区版基础表结构
--   3. psql -d contful_ent -f db/init_ent.sql       ← 本脚本（ALTER + 新建企业版表）
--   4. psql -d contful_ent -f db/seed_data.sql      ← 种子数据
--
-- 设计原则：
--   - 不修改 init_pg.sql（社区版）
--   - 对开源表仅用 ALTER TABLE ADD COLUMN（不删不改现有列）
--   - 新增列全部 DEFAULT NULL，社区版代码不受影响
--   - 企业版独有表以 contful_ent_ 前缀区分
-- =============================================================================

-- =============================================================================
-- 新增 ENUM 类型（仅企业版使用）
-- =============================================================================

DO $$ BEGIN
    CREATE TYPE schedule_action AS ENUM ('publish', 'unpublish');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE schedule_status AS ENUM ('pending', 'running', 'completed', 'failed', 'cancelled');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE schedule_action IS '排期操作类型：publish=发布、unpublish=下架';
COMMENT ON TYPE schedule_status IS '排期执行状态：pending=待执行、running=执行中、completed=已完成、failed=失败、cancelled=已取消';

-- =============================================================================
-- 一、修改开源表结构（ALTER TABLE — 仅追加，不删除不修改）
-- =============================================================================

-- 1.1 entries 表：定时发布排期字段
ALTER TABLE contful_entries ADD COLUMN IF NOT EXISTS scheduled_publish_time TIMESTAMPTZ;
ALTER TABLE contful_entries ADD COLUMN IF NOT EXISTS scheduled_unpublish_time TIMESTAMPTZ;

COMMENT ON COLUMN contful_entries.scheduled_publish_time IS '[企业版] 计划发布时间，非空时表示已排期发布';
COMMENT ON COLUMN contful_entries.scheduled_unpublish_time IS '[企业版] 计划下架时间，非空时表示已排期下架';

-- 排期轮询专用索引（cron 每 30 秒查询到期条目）
CREATE INDEX IF NOT EXISTS idx_entries_scheduled_publish
    ON contful_entries(scheduled_publish_time)
    WHERE scheduled_publish_time IS NOT NULL AND status = 'draft' AND deleted_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_entries_scheduled_unpublish
    ON contful_entries(scheduled_unpublish_time)
    WHERE scheduled_unpublish_time IS NOT NULL AND status = 'published' AND deleted_time IS NULL;

-- 1.2 system_config 表：企业版配置键无需改表，通过 seed_data 写入即可

-- =============================================================================
-- 二、企业版独有表（contful_ent_ 前缀）
-- =============================================================================

-- 2.1 contful_ent_schedule_logs — 排期执行记录
CREATE TABLE IF NOT EXISTS contful_ent_schedule_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES contful_entries(id) ON DELETE CASCADE,
    action schedule_action NOT NULL,
    scheduled_time TIMESTAMPTZ NOT NULL,
    executed_time TIMESTAMPTZ,
    status schedule_status NOT NULL DEFAULT 'pending',
    executed_by UUID,                  -- 系统自动执行为 NULL
    error_message TEXT,
    audit_log_id UUID,                 -- 关联审计日志 ID
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contful_ent_schedule_logs_entry ON contful_ent_schedule_logs(entry_id);
CREATE INDEX IF NOT EXISTS idx_contful_ent_schedule_logs_status ON contful_ent_schedule_logs(status);
CREATE INDEX IF NOT EXISTS idx_contful_ent_schedule_logs_scheduled ON contful_ent_schedule_logs(scheduled_time DESC);
CREATE INDEX IF NOT EXISTS idx_contful_ent_schedule_logs_created ON contful_ent_schedule_logs(created_time DESC);

COMMENT ON TABLE contful_ent_schedule_logs IS '[企业版] 内容排期执行记录：每次定时发布/下架操作的历史追踪';
COMMENT ON COLUMN contful_ent_schedule_logs.id IS '记录唯一标识符';
COMMENT ON COLUMN contful_ent_schedule_logs.entry_id IS '关联的内容条目';
COMMENT ON COLUMN contful_ent_schedule_logs.action IS '排期操作类型：publish/unpublish';
COMMENT ON COLUMN contful_ent_schedule_logs.scheduled_time IS '计划执行时间';
COMMENT ON COLUMN contful_ent_schedule_logs.executed_time IS '实际执行时间';
COMMENT ON COLUMN contful_ent_schedule_logs.status IS '执行状态：pending/running/completed/failed/cancelled';
COMMENT ON COLUMN contful_ent_schedule_logs.executed_by IS '执行者用户 ID（系统自动执行为 NULL）';
COMMENT ON COLUMN contful_ent_schedule_logs.error_message IS '失败原因';
COMMENT ON COLUMN contful_ent_schedule_logs.audit_log_id IS '关联的审计日志 ID';
COMMENT ON COLUMN contful_ent_schedule_logs.created_time IS '记录创建时间';

-- =============================================================================
-- 三、种子数据（企业版默认配置）
-- =============================================================================

-- 企业版 system_config 推荐项（注释形式，由应用代码按需写入）
-- config_key: 'enterprise.schedule_poll_interval'    → value: '60' (秒)
-- config_key: 'enterprise.crypto_algorithm'          → value: 'sm4-gcm' | 'aes-256-gcm'
-- config_key: 'enterprise.signing_algorithm'          → value: 'HMAC-SHA256' | 'SM3withSM2'
-- config_key: 'enterprise.asymmetric_algorithm'      → value: 'rsa' | 'sm2'

-- =============================================================================
-- 变更摘要
-- =============================================================================
--
-- 修改开源表（ALTER，不删不改现有列）：
--   entries:  +scheduled_publish_time, +scheduled_unpublish_time, +2 索引
--
-- 新增企业版独有表（contful_ent_ 前缀）：
--   contful_ent_schedule_logs — 排期执行记录
--
-- 不需要改表的功能：
--   国密全套（SM2/SM3/SM4）— 纯应用层，配置通过 system_config 管理
--
-- 执行顺序：
--   1. psql -f db/init_pg.sql     ← 社区版基础表
--   2. psql -f db/init_ent.sql    ← 本文件（ALTER + 新建企业表）
--   3. psql -f db/seed_data.sql   ← 种子数据

