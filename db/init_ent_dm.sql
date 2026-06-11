-- =============================================================================
-- Contful Enterprise DM8 数据库初始化脚本
-- 达梦数据库 (Oracle 兼容模式) — UTF-8, 不区分大小写
-- 用法: 以 SYSDBA 身份在 DM 管理工具中执行
-- =============================================================================

-- 创建用户（schema）— 使用默认表空间
CREATE USER CONTFUL_ENT IDENTIFIED BY "Contful@2024";

GRANT DBA TO CONTFUL_ENT;
GRANT CREATE TABLE TO CONTFUL_ENT;
GRANT CREATE VIEW TO CONTFUL_ENT;
GRANT CREATE PROCEDURE TO CONTFUL_ENT;
GRANT CREATE SEQUENCE TO CONTFUL_ENT;
GRANT CREATE TRIGGER TO CONTFUL_ENT;

-- =============================================================================
-- UUID 辅助函数：SYS_GUID() → 标准 UUID 格式 (8-4-4-4-12)
-- =============================================================================
CREATE OR REPLACE FUNCTION CONTFUL_ENT.GEN_UUID RETURN VARCHAR2 IS
  v_raw VARCHAR2(32);
  v_hex VARCHAR2(36);
BEGIN
  SELECT RAWTOHEX(SYS_GUID()) INTO v_raw FROM DUAL;
  v_hex := LOWER(SUBSTR(v_raw,1,8) || '-' || SUBSTR(v_raw,9,4) || '-' ||
           SUBSTR(v_raw,13,4) || '-' || SUBSTR(v_raw,17,4) || '-' ||
           SUBSTR(v_raw,21,12));
  RETURN v_hex;
END;
/

-- =============================================================================
-- 企业版增量（审计导出 + 保留策略）
-- 前置条件: 已执行 init_dm.sql
-- =============================================================================
-- 18. 审计导出任务（企业版）
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_audit_report_exports (
    id VARCHAR2(36) ,
    user_id VARCHAR2(36),
    status VARCHAR2(20) DEFAULT 'pending',
    formats CLOB DEFAULT '[]',
    filter CLOB,
    file_path VARCHAR2(500),
    file_size NUMBER,
    expires_time TIMESTAMP,
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_audit_exports PRIMARY KEY (id)
);

-- =============================================================================
-- 19. 审计保留策略（企业版）
-- =============================================================================
CREATE TABLE CONTFUL_ENT.contful_audit_retention_policies (
    id VARCHAR2(36) ,
    category VARCHAR2(50),
    level VARCHAR2(20),
    retention_days NUMBER NOT NULL DEFAULT 180,
    archive_enabled CHAR(1) DEFAULT '0',
    archive_storage VARCHAR2(500),
    is_active CHAR(1) DEFAULT '1',
    created_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    updated_time TIMESTAMP DEFAULT SYSTIMESTAMP,
    CONSTRAINT pk_retention_policies PRIMARY KEY (id),
    CONSTRAINT uq_retention_clevel UNIQUE (category, level)
);

-- =============================================================================

-- =============================================================================
-- UUID 触发器（企业版专属表）
-- =============================================================================
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_audit_exports_bi
BEFORE INSERT ON CONTFUL_ENT.contful_audit_report_exports FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/
CREATE OR REPLACE TRIGGER CONTFUL_ENT.trg_retention_bi
BEFORE INSERT ON CONTFUL_ENT.contful_audit_retention_policies FOR EACH ROW
BEGIN IF :NEW.id IS NULL THEN :NEW.id := CONTFUL_ENT.GEN_UUID(); END IF; END;
/

-- =============================================================================
-- 初始化完成
-- =============================================================================
