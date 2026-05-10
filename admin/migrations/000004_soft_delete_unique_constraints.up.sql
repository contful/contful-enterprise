-- 000004: 软删除 - 唯一约束迁移
-- 将现有的 UNIQUE 约束替换为部分唯一索引（WHERE deleted_time IS NULL）
-- 这样软删除后可以复用 email/name/slug 等唯一字段

-- =============================================
-- system_users: email 唯一约束 → 部分唯一索引
-- =============================================
-- 删除旧的唯一约束和索引
ALTER TABLE system_users DROP CONSTRAINT IF EXISTS system_users_email_key;
DROP INDEX IF EXISTS idx_system_users_email;
-- 创建部分唯一索引（仅对未删除记录）
CREATE UNIQUE INDEX idx_system_users_email_active ON system_users(email) WHERE deleted_time IS NULL;
-- 普通索引保留
CREATE INDEX idx_system_users_email ON system_users(email);
-- 已删除记录索引（便于查询和管理）
CREATE INDEX idx_system_users_deleted_time ON system_users(deleted_time) WHERE deleted_time IS NOT NULL;

-- =============================================
-- system_roles: name 唯一约束 → 部分唯一索引
-- =============================================
ALTER TABLE system_roles DROP CONSTRAINT IF EXISTS system_roles_name_key;
DROP INDEX IF EXISTS idx_system_roles_name;
CREATE UNIQUE INDEX idx_system_roles_name_active ON system_roles(name) WHERE deleted_time IS NULL;
CREATE INDEX idx_system_roles_name ON system_roles(name);
CREATE INDEX idx_system_roles_deleted_time ON system_roles(deleted_time) WHERE deleted_time IS NOT NULL;

-- =============================================
-- schemas: (site_id, slug) 唯一约束 → 部分唯一索引
-- =============================================
ALTER TABLE schemas DROP CONSTRAINT IF EXISTS schemas_site_id_slug_key;
DROP INDEX IF EXISTS idx_schemas_slug;
CREATE UNIQUE INDEX idx_schemas_slug_active ON schemas(site_id, slug) WHERE deleted_time IS NULL;
CREATE INDEX idx_schemas_slug ON schemas(slug);
CREATE INDEX idx_schemas_deleted_time ON schemas(deleted_time) WHERE deleted_time IS NOT NULL;

-- =============================================
-- fields: (schema_id, name) 唯一约束 → 部分唯一索引
-- =============================================
ALTER TABLE fields DROP CONSTRAINT IF EXISTS fields_schema_id_name_key;
DROP INDEX IF EXISTS idx_fields_sort;
CREATE UNIQUE INDEX idx_fields_name_active ON fields(schema_id, name) WHERE deleted_time IS NULL;
CREATE INDEX idx_fields_sort ON fields(schema_id, sort_order);
CREATE INDEX idx_fields_deleted_time ON fields(deleted_time) WHERE deleted_time IS NOT NULL;

-- =============================================
-- assets: uuid 唯一约束 → 部分唯一索引
-- =============================================
ALTER TABLE assets DROP CONSTRAINT IF EXISTS assets_uuid_key;
CREATE UNIQUE INDEX idx_assets_uuid_active ON assets(uuid) WHERE deleted_time IS NULL;
CREATE INDEX idx_assets_deleted_time ON assets(deleted_time) WHERE deleted_time IS NOT NULL;

-- =============================================
-- entries: 已有唯一约束 (schema_id, locale, id) 不需要修改
-- 因为 id 是主键，该约束本身不会影响软删除
-- 添加已删除记录索引
-- =============================================
CREATE INDEX IF NOT EXISTS idx_entries_deleted_time ON entries(deleted_time) WHERE deleted_time IS NOT NULL;
