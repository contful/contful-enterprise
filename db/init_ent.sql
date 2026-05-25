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
--   - 企业版独有表以 ent_ 前缀区分
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

DO $$ BEGIN
    CREATE TYPE report_format AS ENUM ('csv', 'xlsx', 'pdf');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE report_status AS ENUM ('pending', 'processing', 'completed', 'failed');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

COMMENT ON TYPE schedule_action IS '排期操作类型：publish=发布、unpublish=下架';
COMMENT ON TYPE schedule_status IS '排期执行状态：pending=待执行、running=执行中、completed=已完成、failed=失败、cancelled=已取消';
COMMENT ON TYPE report_format IS '审计报告导出格式：csv=CSV、xlsx=Excel、pdf=PDF';
COMMENT ON TYPE report_status IS '报告生成状态：pending=待处理、processing=生成中、completed=已完成、failed=失败';

-- =============================================================================
-- 一、修改开源表结构（ALTER TABLE — 仅追加，不删除不修改）
-- =============================================================================

-- 1.1 entries 表：定时发布排期字段
ALTER TABLE entries ADD COLUMN IF NOT EXISTS scheduled_publish_time TIMESTAMPTZ;
ALTER TABLE entries ADD COLUMN IF NOT EXISTS scheduled_unpublish_time TIMESTAMPTZ;

COMMENT ON COLUMN entries.scheduled_publish_time IS '[企业版] 计划发布时间，非空时表示已排期发布';
COMMENT ON COLUMN entries.scheduled_unpublish_time IS '[企业版] 计划下架时间，非空时表示已排期下架';

-- 排期轮询专用索引（cron 每 30 秒查询到期条目）
CREATE INDEX IF NOT EXISTS idx_entries_scheduled_publish
    ON entries(scheduled_publish_time)
    WHERE scheduled_publish_time IS NOT NULL AND status = 'draft' AND deleted_time IS NULL;

CREATE INDEX IF NOT EXISTS idx_entries_scheduled_unpublish
    ON entries(scheduled_unpublish_time)
    WHERE scheduled_unpublish_time IS NOT NULL AND status = 'published' AND deleted_time IS NULL;

-- 1.2 audit_logs 表：合规报告辅助索引（加速时间范围+类别联合查询）
CREATE INDEX IF NOT EXISTS idx_audit_logs_category_created
    ON audit_logs(category, created_time DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_level_created
    ON audit_logs(level, created_time DESC);

COMMENT ON INDEX idx_audit_logs_category_created IS '[企业版] 加速按类别+时间范围查询（合规报告常用）';
COMMENT ON INDEX idx_audit_logs_level_created IS '[企业版] 加速按级别+时间范围查询（异常行为检测常用）';

-- 1.3 system_config 表：企业版配置键无需改表，通过 seed_data 写入即可

-- =============================================================================
-- 二、企业版独有表（ent_ 前缀）
-- =============================================================================

-- 2.1 ent_schedule_logs — 排期执行记录
CREATE TABLE IF NOT EXISTS ent_schedule_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    action schedule_action NOT NULL,
    scheduled_time TIMESTAMPTZ NOT NULL,
    executed_time TIMESTAMPTZ,
    status schedule_status NOT NULL DEFAULT 'pending',
    executed_by UUID,                  -- 系统自动执行为 NULL
    error_message TEXT,
    audit_log_id UUID,                 -- 关联审计日志 ID
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ent_schedule_logs_entry ON ent_schedule_logs(entry_id);
CREATE INDEX IF NOT EXISTS idx_ent_schedule_logs_status ON ent_schedule_logs(status);
CREATE INDEX IF NOT EXISTS idx_ent_schedule_logs_scheduled ON ent_schedule_logs(scheduled_time DESC);
CREATE INDEX IF NOT EXISTS idx_ent_schedule_logs_created ON ent_schedule_logs(created_time DESC);

COMMENT ON TABLE ent_schedule_logs IS '[企业版] 内容排期执行记录：每次定时发布/下架操作的历史追踪';
COMMENT ON COLUMN ent_schedule_logs.id IS '记录唯一标识符';
COMMENT ON COLUMN ent_schedule_logs.entry_id IS '关联的内容条目';
COMMENT ON COLUMN ent_schedule_logs.action IS '排期操作类型：publish/unpublish';
COMMENT ON COLUMN ent_schedule_logs.scheduled_time IS '计划执行时间';
COMMENT ON COLUMN ent_schedule_logs.executed_time IS '实际执行时间';
COMMENT ON COLUMN ent_schedule_logs.status IS '执行状态：pending/running/completed/failed/cancelled';
COMMENT ON COLUMN ent_schedule_logs.executed_by IS '执行者用户 ID（系统自动执行为 NULL）';
COMMENT ON COLUMN ent_schedule_logs.error_message IS '失败原因';
COMMENT ON COLUMN ent_schedule_logs.audit_log_id IS '关联的审计日志 ID';
COMMENT ON COLUMN ent_schedule_logs.created_time IS '记录创建时间';

-- 2.2 ent_audit_report_exports — 审计报告导出记录
CREATE TABLE IF NOT EXISTS ent_audit_report_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,         -- 报告名称
    format report_format NOT NULL,      -- 导出格式
    template VARCHAR(100),              -- 模板名称（等保/SOC2/GDPR/custom）
    status report_status NOT NULL DEFAULT 'pending',
    filter JSONB NOT NULL DEFAULT '{}', -- 筛选条件快照（方便重新生成）
    total_records INT,                  -- 包含的审计记录数
    file_path TEXT,                     -- 导出文件存储路径
    file_size BIGINT,                   -- 文件大小（字节）
    requested_by UUID,                  -- 请求者用户 ID
    completed_time TIMESTAMPTZ,         -- 完成时间
    error_message TEXT,                 -- 失败原因
    expires_time TIMESTAMPTZ,           -- 文件过期时间（报告有有效期）
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ent_report_exports_status ON ent_audit_report_exports(status);
CREATE INDEX IF NOT EXISTS idx_ent_report_exports_requested ON ent_audit_report_exports(requested_by);
CREATE INDEX IF NOT EXISTS idx_ent_report_exports_created ON ent_audit_report_exports(created_time DESC);
CREATE INDEX IF NOT EXISTS idx_ent_report_exports_expires ON ent_audit_report_exports(expires_time)
    WHERE expires_time IS NOT NULL;

COMMENT ON TABLE ent_audit_report_exports IS '[企业版] 审计报告导出记录：追踪每次合规报告的生成和文件生命周期';
COMMENT ON COLUMN ent_audit_report_exports.id IS '报告唯一标识符';
COMMENT ON COLUMN ent_audit_report_exports.name IS '报告名称';
COMMENT ON COLUMN ent_audit_report_exports.format IS '导出格式：csv/xlsx/pdf';
COMMENT ON COLUMN ent_audit_report_exports.template IS '报告模板：等保/SOC2/GDPR/自定义';
COMMENT ON COLUMN ent_audit_report_exports.status IS '生成状态：pending/processing/completed/failed';
COMMENT ON COLUMN ent_audit_report_exports.filter IS '筛选条件 JSON 快照';
COMMENT ON COLUMN ent_audit_report_exports.total_records IS '包含的审计日志条数';
COMMENT ON COLUMN ent_audit_report_exports.file_path IS '导出文件路径';
COMMENT ON COLUMN ent_audit_report_exports.file_size IS '文件大小（字节）';
COMMENT ON COLUMN ent_audit_report_exports.requested_by IS '请求导出者';
COMMENT ON COLUMN ent_audit_report_exports.completed_time IS '生成完成时间';
COMMENT ON COLUMN ent_audit_report_exports.error_message IS '生成失败原因';
COMMENT ON COLUMN ent_audit_report_exports.expires_time IS '文件过期时间（过期后自动清理）';
COMMENT ON COLUMN ent_audit_report_exports.created_time IS '报告创建时间';

-- 2.3 ent_audit_retention_policies — 审计日志保留策略
CREATE TABLE IF NOT EXISTS ent_audit_retention_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category audit_type,               -- NULL 表示所有类别
    level audit_level,                 -- NULL 表示所有级别
    retention_days INT NOT NULL DEFAULT 180,
    archive_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    archive_storage VARCHAR(50) DEFAULT 'local',  -- local/s3/oss
    archive_path TEXT,
    note TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID,
    created_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(category, level)            -- 每个 类别-级别 组合仅一条策略
);

CREATE INDEX IF NOT EXISTS idx_ent_retention_active ON ent_audit_retention_policies(is_active);

COMMENT ON TABLE ent_audit_retention_policies IS '[企业版] 审计日志保留策略：按类别和级别配置不同的保留期和归档规则';
COMMENT ON COLUMN ent_audit_retention_policies.id IS '策略唯一标识符';
COMMENT ON COLUMN ent_audit_retention_policies.category IS '审计类别（NULL=所有类别）';
COMMENT ON COLUMN ent_audit_retention_policies.level IS '审计级别（NULL=所有级别）';
COMMENT ON COLUMN ent_audit_retention_policies.retention_days IS '保留天数（默认 180 天）';
COMMENT ON COLUMN ent_audit_retention_policies.archive_enabled IS '是否归档（而非直接删除）';
COMMENT ON COLUMN ent_audit_retention_policies.archive_storage IS '归档存储后端：local/s3/oss';
COMMENT ON COLUMN ent_audit_retention_policies.archive_path IS '归档路径';
COMMENT ON COLUMN ent_audit_retention_policies.note IS '策略说明';
COMMENT ON COLUMN ent_audit_retention_policies.is_active IS '是否启用';
COMMENT ON COLUMN ent_audit_retention_policies.created_by IS '创建者用户 ID';
COMMENT ON COLUMN ent_audit_retention_policies.created_time IS '创建时间';
COMMENT ON COLUMN ent_audit_retention_policies.updated_time IS '更新时间';

-- =============================================================================
-- 二、种子数据（企业版默认配置）
-- =============================================================================

-- 2.1 默认保留策略（180 天全保留）
INSERT INTO ent_audit_retention_policies (category, level, retention_days, archive_enabled, note)
VALUES (NULL, NULL, 180, FALSE, '默认保留策略：所有审计日志保留 180 天后删除')
ON CONFLICT (category, level) DO NOTHING;

-- 2.2 企业版 system_config 推荐项（注释形式，由应用代码按需写入）
-- config_key: 'enterprise.schedule_poll_interval'    → value: '60' (秒)
-- config_key: 'enterprise.crypto_algorithm'          → value: 'sm4-gcm' | 'aes-256-gcm'
-- config_key: 'enterprise.signing_algorithm'          → value: 'HMAC-SHA256' | 'SM3withSM2'
-- config_key: 'enterprise.asymmetric_algorithm'      → value: 'rsa' | 'sm2'
-- config_key: 'enterprise.report_max_days'           → value: '365' (报告最大时间跨度)
-- config_key: 'enterprise.report_export_expires_days' → value: '7' (导出文件保留天数)

-- =============================================================================
-- 变更摘要
-- =============================================================================
--
-- 修改开源表（ALTER，不删不改现有列）：
--   entries:  +scheduled_publish_time, +scheduled_unpublish_time, +2 索引
--   audit_logs: +2 辅助索引
--
-- 新增企业版独有表（ent_ 前缀）：
--   ent_schedule_logs          — 排期执行记录
--   ent_audit_report_exports   — 审计报告导出记录
--   ent_audit_retention_policies — 审计保留策略配置
--
-- 不需要改表的功能：
--   国密全套（SM2/SM3/SM4）— 纯应用层，配置通过 system_config 管理
--
-- 执行顺序：
--   1. psql -f db/init_pg.sql     ← 社区版基础表
--   2. psql -f db/init_ent.sql    ← 本文件（ALTER + 新建企业表）
--   3. psql -f db/seed_data.sql   ← 种子数据
