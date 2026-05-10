-- 000004 Down: 恢复原始 UNIQUE 约束

-- system_users
DROP INDEX IF EXISTS idx_system_users_deleted_time;
DROP INDEX IF EXISTS idx_system_users_email;
DROP INDEX IF EXISTS idx_system_users_email_active;
ALTER TABLE system_users ADD CONSTRAINT system_users_email_key UNIQUE (email);
CREATE INDEX idx_system_users_email ON system_users(email);

-- system_roles
DROP INDEX IF EXISTS idx_system_roles_deleted_time;
DROP INDEX IF EXISTS idx_system_roles_name;
DROP INDEX IF EXISTS idx_system_roles_name_active;
ALTER TABLE system_roles ADD CONSTRAINT system_roles_name_key UNIQUE (name);
CREATE INDEX idx_system_roles_name ON system_roles(name);

-- schemas
DROP INDEX IF EXISTS idx_schemas_deleted_time;
DROP INDEX IF EXISTS idx_schemas_slug;
DROP INDEX IF EXISTS idx_schemas_slug_active;
ALTER TABLE schemas ADD CONSTRAINT schemas_site_id_slug_key UNIQUE (site_id, slug);
CREATE INDEX idx_schemas_slug ON schemas(slug);

-- fields
DROP INDEX IF EXISTS idx_fields_deleted_time;
DROP INDEX IF EXISTS idx_fields_sort;
DROP INDEX IF EXISTS idx_fields_name_active;
ALTER TABLE fields ADD CONSTRAINT fields_schema_id_name_key UNIQUE (schema_id, name);
CREATE INDEX idx_fields_sort ON fields(schema_id, sort_order);

-- assets
DROP INDEX IF EXISTS idx_assets_deleted_time;
DROP INDEX IF EXISTS idx_assets_uuid_active;
ALTER TABLE assets ADD CONSTRAINT assets_uuid_key UNIQUE (uuid);

-- entries
DROP INDEX IF EXISTS idx_entries_deleted_time;
