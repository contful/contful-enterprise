-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Migration 000001: Rollback initial schema
-- =============================================================================

-- 删除索引
DROP INDEX IF EXISTS idx_tokens_expires;
DROP INDEX IF EXISTS idx_tokens_status;
DROP INDEX IF EXISTS idx_tokens_hash;
DROP INDEX IF EXISTS idx_tokens_site;

DROP INDEX IF EXISTS idx_assets_used;
DROP INDEX IF EXISTS idx_assets_download;
DROP INDEX IF EXISTS idx_assets_created;
DROP INDEX IF EXISTS idx_assets_hash;
DROP INDEX IF EXISTS idx_assets_slug;
DROP INDEX IF EXISTS idx_assets_extension;
DROP INDEX IF EXISTS idx_assets_type;
DROP INDEX IF EXISTS idx_assets_folder;
DROP INDEX IF EXISTS idx_assets_site;

DROP INDEX IF EXISTS idx_asset_folders_parent;
DROP INDEX IF EXISTS idx_asset_folders_site;

DROP INDEX IF EXISTS idx_entry_versions_created;
DROP INDEX IF EXISTS idx_entry_versions_entry;

DROP INDEX IF EXISTS idx_entry_values_field;
DROP INDEX IF EXISTS idx_entry_values_entry;

DROP INDEX IF EXISTS idx_entries_deleted;
DROP INDEX IF EXISTS idx_entries_created_by;
DROP INDEX IF EXISTS idx_entries_sort;
DROP INDEX IF EXISTS idx_entries_published;
DROP INDEX IF EXISTS idx_entries_status;
DROP INDEX IF EXISTS idx_entries_locale;
DROP INDEX IF EXISTS idx_entries_site;
DROP INDEX IF EXISTS idx_entries_schema;

DROP INDEX IF EXISTS idx_fields_sort;
DROP INDEX IF EXISTS idx_fields_schema;

DROP INDEX IF EXISTS idx_schemas_kind;
DROP INDEX IF EXISTS idx_schemas_active;
DROP INDEX IF EXISTS idx_schemas_slug;
DROP INDEX IF EXISTS idx_schemas_site;

DROP INDEX IF EXISTS idx_sites_locale;
DROP INDEX IF EXISTS idx_sites_active;
DROP INDEX IF EXISTS idx_sites_slug;

DROP INDEX IF EXISTS idx_audit_logs_created;
DROP INDEX IF EXISTS idx_audit_logs_category;
DROP INDEX IF EXISTS idx_audit_logs_user;
DROP INDEX IF EXISTS idx_audit_logs_site;

DROP INDEX IF EXISTS idx_system_config_key;

DROP INDEX IF EXISTS idx_system_user_roles_role;
DROP INDEX IF EXISTS idx_system_user_roles_user;

DROP INDEX IF EXISTS idx_system_roles_name;

DROP INDEX IF EXISTS idx_system_users_status;
DROP INDEX IF EXISTS idx_system_users_email;

-- 删除表（按依赖逆序）
DROP TABLE IF EXISTS tokens CASCADE;
DROP TABLE IF EXISTS assets CASCADE;
DROP TABLE IF EXISTS asset_folders CASCADE;
DROP TABLE IF EXISTS entry_versions CASCADE;
DROP TABLE IF EXISTS entry_values CASCADE;
DROP TABLE IF EXISTS entries CASCADE;
DROP TABLE IF EXISTS fields CASCADE;
DROP TABLE IF EXISTS schemas CASCADE;
DROP TABLE IF EXISTS sites CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS system_config CASCADE;
DROP TABLE IF EXISTS system_user_roles CASCADE;
DROP TABLE IF EXISTS system_roles CASCADE;
DROP TABLE IF EXISTS system_users CASCADE;

-- 删除 ENUM 类型
DROP TYPE IF EXISTS audit_type;
DROP TYPE IF EXISTS audit_level;
DROP TYPE IF EXISTS token_status;
DROP TYPE IF EXISTS asset_status;
DROP TYPE IF EXISTS asset_visibility;
DROP TYPE IF EXISTS asset_type;
DROP TYPE IF EXISTS field_type;
DROP TYPE IF EXISTS content_type_kind;
DROP TYPE IF EXISTS entry_status;
DROP TYPE IF EXISTS user_status;

-- 删除扩展
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
