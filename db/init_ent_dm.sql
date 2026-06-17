-- =============================================================================
-- Contful Enterprise — 企业版增量初始化脚本（达梦 DM8 版）
-- 版本: v2.0.0-ent
-- 数据库: 达梦 DM8+
--
-- 使用方式（开发环境）：
--   1. 创建 contful_ent 模式
--   2. 先执行 init_dm.sql（社区版 DDL）
--   3. 执行本脚本（ALTER + 新建企业版表）
--   4. 执行 seed_data.sql（种子数据）
--
-- 设计原则：
--   - 不修改 init_dm.sql（社区版）
--   - 对开源表仅用 ALTER TABLE ADD（不删不改现有列）
--   - PostgreSQL 专有特性（tsvector/GIN/JSONB）不在此文件中
--   - 全文搜索和 JSONB 替代方案在应用层处理
-- =============================================================================

-- =============================================================================
-- 一、修改开源表结构（ALTER TABLE — 仅追加）
-- =============================================================================

-- 1.1 entries 表：定时发布排期字段（字段和索引已在 init_dm.sql 中定义，此处仅补充注释）
COMMENT ON COLUMN contful_entries.scheduled_publish_time IS '[企业版] 计划发布时间，非空时表示已排期发布';
COMMENT ON COLUMN contful_entries.scheduled_unpublish_time IS '[企业版] 计划下架时间，非空时表示已排期下架';

-- 1.2 contful_audit_logs 表：企业版审计日志增强字段（达梦不支持 tsvector/GIN/JSONB）
ALTER TABLE contful_audit_logs ADD request_body TEXT;
ALTER TABLE contful_audit_logs ADD response_status SMALLINT;
ALTER TABLE contful_audit_logs ADD duration_ms INTEGER;
ALTER TABLE contful_audit_logs ADD session_id VARCHAR(64);
ALTER TABLE contful_audit_logs ADD geo_ip_info TEXT;

COMMENT ON COLUMN contful_audit_logs.request_body IS '[企业版] 请求体内容';
COMMENT ON COLUMN contful_audit_logs.response_status IS '[企业版] 响应 HTTP 状态码';
COMMENT ON COLUMN contful_audit_logs.duration_ms IS '[企业版] 请求耗时（毫秒）';
COMMENT ON COLUMN contful_audit_logs.session_id IS '[企业版] 会话 ID';
COMMENT ON COLUMN contful_audit_logs.geo_ip_info IS '[企业版] IP 地理位置信息';

-- =============================================================================
-- 二、企业版独有表
-- =============================================================================

-- 2.1 contful_ent_schedule_logs — 排期执行记录
CREATE TABLE contful_ent_schedule_logs (
    id VARCHAR(36) NOT NULL,
    entry_id VARCHAR(36) NOT NULL,
    action VARCHAR(20) NOT NULL,
    scheduled_time TIMESTAMP NOT NULL,
    executed_time TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    executed_by VARCHAR(36),
    error_message TEXT,
    audit_log_id VARCHAR(36),
    created_time TIMESTAMP NOT NULL DEFAULT SYSDATE,
    PRIMARY KEY (id)
);

CREATE INDEX idx_ent_schedule_logs_entry ON contful_ent_schedule_logs(entry_id);
CREATE INDEX idx_ent_schedule_logs_status ON contful_ent_schedule_logs(status);
CREATE INDEX idx_ent_schedule_logs_scheduled ON contful_ent_schedule_logs(scheduled_time DESC);
CREATE INDEX idx_ent_schedule_logs_created ON contful_ent_schedule_logs(created_time DESC);

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

-- 2.2 contful_audit_anomalies — 审计异常事件（达梦版：JSONB → TEXT）
CREATE TABLE contful_audit_anomalies (
    id VARCHAR(36) NOT NULL,
    audit_log_id VARCHAR(36),
    anomaly_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    score DECIMAL(5,2) NOT NULL,
    baseline_value TEXT,
    actual_value TEXT,
    description TEXT NOT NULL,
    detected_time TIMESTAMP NOT NULL,
    resolved_time TIMESTAMP,
    resolution_note TEXT,
    created_time TIMESTAMP NOT NULL DEFAULT SYSDATE,
    PRIMARY KEY (id)
);

CREATE INDEX idx_audit_anomalies_type ON contful_audit_anomalies(anomaly_type);
CREATE INDEX idx_audit_anomalies_detected ON contful_audit_anomalies(detected_time DESC);
CREATE INDEX idx_audit_anomalies_log ON contful_audit_anomalies(audit_log_id);
CREATE INDEX idx_audit_anomalies_severity ON contful_audit_anomalies(severity);

COMMENT ON TABLE contful_audit_anomalies IS '[企业版] 审计异常事件表：存储异常检测结果';
COMMENT ON COLUMN contful_audit_anomalies.id IS '异常事件唯一标识符';
COMMENT ON COLUMN contful_audit_anomalies.audit_log_id IS '关联的审计日志 ID';
COMMENT ON COLUMN contful_audit_anomalies.anomaly_type IS '异常类型：abnormal_login/high_frequency/permission_escalation/time_series_anomaly/behavior_deviation';
COMMENT ON COLUMN contful_audit_anomalies.severity IS '严重程度：low/medium/high/critical';
COMMENT ON COLUMN contful_audit_anomalies.score IS '异常评分（0-100）';
COMMENT ON COLUMN contful_audit_anomalies.baseline_value IS '基线值';
COMMENT ON COLUMN contful_audit_anomalies.actual_value IS '实际值';
COMMENT ON COLUMN contful_audit_anomalies.description IS '异常描述';
COMMENT ON COLUMN contful_audit_anomalies.detected_time IS '检测时间';
COMMENT ON COLUMN contful_audit_anomalies.resolved_time IS '解决时间';
COMMENT ON COLUMN contful_audit_anomalies.resolution_note IS '解决备注';
COMMENT ON COLUMN contful_audit_anomalies.created_time IS '创建时间';

-- 2.3 contful_audit_query_templates — 审计查询模板（达梦版：JSONB → TEXT）
CREATE TABLE contful_audit_query_templates (
    id VARCHAR(36) NOT NULL,
    name VARCHAR(200) NOT NULL,
    conditions TEXT NOT NULL,
    created_by VARCHAR(36),
    created_time TIMESTAMP NOT NULL DEFAULT SYSDATE,
    PRIMARY KEY (id)
);

CREATE INDEX idx_audit_query_templates_created_by ON contful_audit_query_templates(created_by);

COMMENT ON TABLE contful_audit_query_templates IS '[企业版] 审计查询模板表：保存用户查询条件组合';
COMMENT ON COLUMN contful_audit_query_templates.id IS '模板唯一标识符';
COMMENT ON COLUMN contful_audit_query_templates.name IS '模板名称';
COMMENT ON COLUMN contful_audit_query_templates.conditions IS '查询条件（JSON 字符串）';
COMMENT ON COLUMN contful_audit_query_templates.created_by IS '创建者用户 ID';
COMMENT ON COLUMN contful_audit_query_templates.created_time IS '创建时间';

-- =============================================================================
-- 变更摘要
-- =============================================================================
--
-- 修改开源表（ALTER，不删不改现有列）：
--   entries:  +scheduled_publish_time, +scheduled_unpublish_time, +2 索引
--   contful_audit_logs:  +request_body, +response_status, +duration_ms, +session_id, +geo_ip_info
--
-- 新增企业版表：
--   contful_ent_schedule_logs — 排期执行记录
--   contful_audit_anomalies — 审计异常事件
--   contful_audit_query_templates — 审计查询模板
--
-- 达梦限制（与 PostgreSQL 版差异）：
--   不支持 tsvector → 无全文搜索向量列和触发器
--   不支持 GIN 索引 → 无 idx_audit_logs_search
--   不支持 JSONB → geo_ip_info/baseline_value/actual_value/conditions 使用 TEXT
--   不支持 gen_random_uuid() → 使用 VARCHAR(36) + 应用层生成 UUID
--   不支持 TIMESTAMPTZ → 使用 TIMESTAMP + SYSDATE
