-- =============================================================================
-- Contful Enterprise DM8 升级脚本 v1.4.0 — 新增 Webhook 功能
-- =============================================================================
-- 用法: 以 CONTFUL_ENT 身份执行
-- =============================================================================

-- 1. Webhook 配置表（达梦版本）
BEGIN
  EXECUTE IMMEDIATE 'CREATE TABLE CONTFUL_ENT.contful_webhooks (
    id VARCHAR2(36),
    site_id VARCHAR2(36) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    url VARCHAR2(2000) NOT NULL,
    events CLOB DEFAULT ''{}'',
    secret VARCHAR2(255),
    is_active CHAR(1) DEFAULT ''1'',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_webhooks PRIMARY KEY (id),
    CONSTRAINT fk_webhooks_site FOREIGN KEY (site_id) REFERENCES CONTFUL_ENT.contful_sites(id)
  )';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

BEGIN
  EXECUTE IMMEDIATE 'CREATE INDEX idx_webhooks_site ON CONTFUL_ENT.contful_webhooks(site_id)';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

BEGIN
  EXECUTE IMMEDIATE 'CREATE INDEX idx_webhooks_active ON CONTFUL_ENT.contful_webhooks(is_active)';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

-- 2. Webhook 投递记录表
BEGIN
  EXECUTE IMMEDIATE 'CREATE TABLE CONTFUL_ENT.contful_webhook_deliveries (
    id VARCHAR2(36),
    webhook_id VARCHAR2(36) NOT NULL,
    event VARCHAR2(50),
    payload CLOB,
    response_status INT,
    response_body CLOB,
    status VARCHAR2(20) DEFAULT ''pending'',
    attempt INT DEFAULT 1,
    error_message CLOB,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_webhook_deliveries PRIMARY KEY (id),
    CONSTRAINT fk_wd_webhook FOREIGN KEY (webhook_id) REFERENCES CONTFUL_ENT.contful_webhooks(id)
  )';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

BEGIN
  EXECUTE IMMEDIATE 'CREATE INDEX idx_wd_webhook ON CONTFUL_ENT.contful_webhook_deliveries(webhook_id)';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

BEGIN
  EXECUTE IMMEDIATE 'CREATE INDEX idx_wd_status ON CONTFUL_ENT.contful_webhook_deliveries(status)';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

BEGIN
  EXECUTE IMMEDIATE 'CREATE INDEX idx_wd_created ON CONTFUL_ENT.contful_webhook_deliveries(created_time)';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

-- BEFORE INSERT 触发器（UUID 自动生成）
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_webhooks_bi
BEFORE INSERT ON CONTFUL_ENT.contful_webhooks FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/

CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_wd_bi
BEFORE INSERT ON CONTFUL_ENT.contful_webhook_deliveries FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/

-- 3. Webhook 权限注册（种子数据）
MERGE INTO CONTFUL_ENT.contful_system_permission_groups t
USING (SELECT '00000000-0000-0000-0000-000000000311' AS id, 'webhook' AS gkey, 'Webhook' AS label, 'Webhooks' AS label_en, 10 AS sort FROM DUAL) s
ON (t.group_key = s.gkey)
WHEN NOT MATCHED THEN INSERT (id, group_key, label, label_en, sort_order) VALUES (s.id, s.gkey, s.label, s.label_en, s.sort);

BEGIN
  EXECUTE IMMEDIATE 'INSERT INTO CONTFUL_ENT.contful_system_permissions (group_id, action, label, label_en, sort_order)
  SELECT ''00000000-0000-0000-0000-000000000311'', ''read'', ''查看 Webhook'', ''View Webhooks'', 0 FROM DUAL
  WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_permissions WHERE group_id=''00000000-0000-0000-0000-000000000311'' AND action=''read'')';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/

BEGIN
  EXECUTE IMMEDIATE 'INSERT INTO CONTFUL_ENT.contful_system_permissions (group_id, action, label, label_en, sort_order)
  SELECT ''00000000-0000-0000-0000-000000000311'', ''write'', ''管理 Webhook'', ''Manage Webhooks'', 1 FROM DUAL
  WHERE NOT EXISTS (SELECT 1 FROM CONTFUL_ENT.contful_system_permissions WHERE group_id=''00000000-0000-0000-0000-000000000311'' AND action=''write'')';
EXCEPTION WHEN OTHERS THEN NULL;
END;
/
COMMIT;
