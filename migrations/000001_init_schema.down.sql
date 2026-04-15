-- +migrate Down
-- +migrate StatementBegin

-- 删除触发器
DROP TRIGGER IF EXISTS update_global_users_updated_at ON global_users;
DROP TRIGGER IF EXISTS update_global_roles_updated_at ON global_roles;
DROP TRIGGER IF EXISTS update_plugins_updated_at ON plugins;
DROP TRIGGER IF EXISTS update_sites_updated_at ON sites;
DROP TRIGGER IF EXISTS update_channels_updated_at ON channels;
DROP TRIGGER IF EXISTS update_locales_updated_at ON locales;
DROP TRIGGER IF EXISTS update_site_roles_updated_at ON site_roles;
DROP TRIGGER IF EXISTS update_site_users_updated_at ON site_users;
DROP TRIGGER IF EXISTS update_content_types_updated_at ON content_types;
DROP TRIGGER IF EXISTS update_fields_updated_at ON fields;
DROP TRIGGER IF EXISTS update_entries_updated_at ON entries;
DROP TRIGGER IF EXISTS update_entry_values_updated_at ON entry_values;
DROP TRIGGER IF EXISTS update_entry_versions_updated_at ON entry_versions;
DROP TRIGGER IF EXISTS update_assets_updated_at ON assets;
DROP TRIGGER IF EXISTS update_api_tokens_updated_at ON api_tokens;
DROP TRIGGER IF EXISTS update_webhooks_updated_at ON webhooks;

-- 删除触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除表（按依赖顺序）
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS entry_versions;
DROP TABLE IF EXISTS entry_values;
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS fields;
DROP TABLE IF EXISTS content_types;
DROP TABLE IF EXISTS site_users;
DROP TABLE IF EXISTS site_roles;
DROP TABLE IF EXISTS locales;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS sites;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS plugins;
DROP TABLE IF EXISTS global_roles;
DROP TABLE IF EXISTS global_users;
DROP TABLE IF EXISTS distributed_locks;

-- 删除枚举类型
DROP TYPE IF EXISTS audit_type;
DROP TYPE IF EXISTS audit_level;
DROP TYPE IF EXISTS token_status;
DROP TYPE IF EXISTS asset_status;
DROP TYPE IF EXISTS asset_type;
DROP TYPE IF EXISTS field_type;
DROP TYPE IF EXISTS content_type_kind;
DROP TYPE IF EXISTS entry_status;
DROP TYPE IF EXISTS user_status;

-- 删除扩展（谨慎使用）
-- DROP EXTENSION IF EXISTS "uuid-ossp";
-- DROP EXTENSION IF EXISTS "pgcrypto";

-- +migrate StatementEnd
