-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Migration 000003: Rollback advanced database features
-- =============================================================================

-- =============================================================================
-- 1. 删除触发器
-- =============================================================================

DROP TRIGGER IF EXISTS update_tokens_updated_at ON tokens;
DROP TRIGGER IF EXISTS update_assets_updated_at ON assets;
DROP TRIGGER IF EXISTS update_asset_folders_updated_at ON asset_folders;
DROP TRIGGER IF EXISTS update_entry_versions_updated_at ON entry_versions;
DROP TRIGGER IF EXISTS update_entry_values_updated_at ON entry_values;
DROP TRIGGER IF EXISTS update_entries_updated_at ON entries;
DROP TRIGGER IF EXISTS update_fields_updated_at ON fields;
DROP TRIGGER IF EXISTS update_schemas_updated_at ON schemas;
DROP TRIGGER IF EXISTS update_sites_updated_at ON sites;
DROP TRIGGER IF EXISTS update_audit_logs_updated_at ON audit_logs;
DROP TRIGGER IF EXISTS update_system_config_updated_at ON system_config;
DROP TRIGGER IF EXISTS update_system_user_roles_updated_at ON system_user_roles;
DROP TRIGGER IF EXISTS update_system_roles_updated_at ON system_roles;
DROP TRIGGER IF EXISTS update_system_users_updated_at ON system_users;
DROP TRIGGER IF EXISTS prevent_audit_logs_update ON audit_logs;

-- =============================================================================
-- 2. 删除触发器函数
-- =============================================================================

DROP FUNCTION IF EXISTS prevent_audit_log_update();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- =============================================================================
-- 3. 删除复杂索引
-- =============================================================================

DROP INDEX IF EXISTS idx_audit_logs_category_time;
DROP INDEX IF EXISTS idx_audit_logs_site_user_time;
DROP INDEX IF EXISTS idx_assets_path;
DROP INDEX IF EXISTS idx_assets_site_created;
DROP INDEX IF EXISTS idx_assets_site_type;
DROP INDEX IF EXISTS idx_entries_site_locale;
DROP INDEX IF EXISTS idx_entries_list;
DROP INDEX IF EXISTS idx_entry_values_text_gin;

-- =============================================================================
-- 4. 删除数据完整性签名列
-- =============================================================================

ALTER TABLE schemas DROP COLUMN IF EXISTS signature_enabled;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS data_signature;
ALTER TABLE assets DROP COLUMN IF EXISTS data_signature;
ALTER TABLE entry_values DROP COLUMN IF EXISTS data_signature;
ALTER TABLE entries DROP COLUMN IF EXISTS data_signature;

-- =============================================================================
-- 5. 删除表注释和列注释
-- =============================================================================

-- ENUM 类型注释
COMMENT ON TYPE audit_type IS NULL;
COMMENT ON TYPE audit_level IS NULL;
COMMENT ON TYPE token_status IS NULL;
COMMENT ON TYPE asset_status IS NULL;
COMMENT ON TYPE asset_visibility IS NULL;
COMMENT ON TYPE asset_type IS NULL;
COMMENT ON TYPE field_type IS NULL;
COMMENT ON TYPE content_type_kind IS NULL;
COMMENT ON TYPE entry_status IS NULL;
COMMENT ON TYPE user_status IS NULL;

-- tokens 表注释
COMMENT ON COLUMN tokens.updated_at IS NULL;
COMMENT ON COLUMN tokens.created_at IS NULL;
COMMENT ON COLUMN tokens.created_by IS NULL;
COMMENT ON COLUMN tokens.last_used_at IS NULL;
COMMENT ON COLUMN tokens.status IS NULL;
COMMENT ON COLUMN tokens.expires_at IS NULL;
COMMENT ON COLUMN tokens.rate_limits IS NULL;
COMMENT ON COLUMN tokens.permissions IS NULL;
COMMENT ON COLUMN tokens.token_hash IS NULL;
COMMENT ON COLUMN tokens.token_prefix IS NULL;
COMMENT ON COLUMN tokens.description IS NULL;
COMMENT ON COLUMN tokens.name IS NULL;
COMMENT ON COLUMN tokens.site_id IS NULL;
COMMENT ON COLUMN tokens.id IS NULL;
COMMENT ON TABLE tokens IS NULL;

-- assets 表注释
COMMENT ON COLUMN assets.updated_at IS NULL;
COMMENT ON COLUMN assets.created_at IS NULL;
COMMENT ON COLUMN assets.created_by IS NULL;
COMMENT ON COLUMN assets.used_count IS NULL;
COMMENT ON COLUMN assets.download_count IS NULL;
COMMENT ON COLUMN assets.disk IS NULL;
COMMENT ON COLUMN assets.file_hash IS NULL;
COMMENT ON COLUMN assets.visibility IS NULL;
COMMENT ON COLUMN assets.metadata IS NULL;
COMMENT ON COLUMN assets.tags IS NULL;
COMMENT ON COLUMN assets.description IS NULL;
COMMENT ON COLUMN assets.alt_text IS NULL;
COMMENT ON COLUMN assets.caption IS NULL;
COMMENT ON COLUMN assets.title IS NULL;
COMMENT ON COLUMN assets.alt IS NULL;
COMMENT ON COLUMN assets.thumbnail_url IS NULL;
COMMENT ON COLUMN assets.url IS NULL;
COMMENT ON COLUMN assets.path IS NULL;
COMMENT ON COLUMN assets.duration IS NULL;
COMMENT ON COLUMN assets.height IS NULL;
COMMENT ON COLUMN assets.width IS NULL;
COMMENT ON COLUMN assets.size IS NULL;
COMMENT ON COLUMN assets.extension IS NULL;
COMMENT ON COLUMN assets.mime_type IS NULL;
COMMENT ON COLUMN assets.type IS NULL;
COMMENT ON COLUMN assets.slug IS NULL;
COMMENT ON COLUMN assets.original_name IS NULL;
COMMENT ON COLUMN assets.name IS NULL;
COMMENT ON COLUMN assets.uuid IS NULL;
COMMENT ON COLUMN assets.folder_id IS NULL;
COMMENT ON COLUMN assets.site_id IS NULL;
COMMENT ON COLUMN assets.id IS NULL;
COMMENT ON COLUMN assets.data_signature IS NULL;
COMMENT ON TABLE assets IS NULL;

-- asset_folders 表注释
COMMENT ON COLUMN asset_folders.updated_at IS NULL;
COMMENT ON COLUMN asset_folders.created_at IS NULL;
COMMENT ON COLUMN asset_folders.created_by IS NULL;
COMMENT ON COLUMN asset_folders.sort_order IS NULL;
COMMENT ON COLUMN asset_folders.path IS NULL;
COMMENT ON COLUMN asset_folders.slug IS NULL;
COMMENT ON COLUMN asset_folders.name IS NULL;
COMMENT ON COLUMN asset_folders.parent_id IS NULL;
COMMENT ON COLUMN asset_folders.site_id IS NULL;
COMMENT ON COLUMN asset_folders.id IS NULL;
COMMENT ON TABLE asset_folders IS NULL;

-- entry_versions 表注释
COMMENT ON COLUMN entry_versions.change_summary IS NULL;
COMMENT ON COLUMN entry_versions.created_at IS NULL;
COMMENT ON COLUMN entry_versions.created_by IS NULL;
COMMENT ON COLUMN entry_versions.values_snapshot IS NULL;
COMMENT ON COLUMN entry_versions.version IS NULL;
COMMENT ON COLUMN entry_versions.entry_id IS NULL;
COMMENT ON COLUMN entry_versions.id IS NULL;
COMMENT ON TABLE entry_versions IS NULL;

-- entry_values 表注释
COMMENT ON COLUMN entry_values.updated_at IS NULL;
COMMENT ON COLUMN entry_values.created_at IS NULL;
COMMENT ON COLUMN entry_values.datetime_value IS NULL;
COMMENT ON COLUMN entry_values.date_value IS NULL;
COMMENT ON COLUMN entry_values.bool_value IS NULL;
COMMENT ON COLUMN entry_values.number_value IS NULL;
COMMENT ON COLUMN entry_values.text_value IS NULL;
COMMENT ON COLUMN entry_values.value IS NULL;
COMMENT ON COLUMN entry_values.field_id IS NULL;
COMMENT ON COLUMN entry_values.entry_id IS NULL;
COMMENT ON COLUMN entry_values.id IS NULL;
COMMENT ON COLUMN entry_values.data_signature IS NULL;
COMMENT ON TABLE entry_values IS NULL;

-- entries 表注释
COMMENT ON COLUMN entries.updated_at IS NULL;
COMMENT ON COLUMN entries.created_at IS NULL;
COMMENT ON COLUMN entries.created_by IS NULL;
COMMENT ON COLUMN entries.sort_weight IS NULL;
COMMENT ON COLUMN entries.seo_keywords IS NULL;
COMMENT ON COLUMN entries.seo_description IS NULL;
COMMENT ON COLUMN entries.seo_title IS NULL;
COMMENT ON COLUMN entries.relations IS NULL;
COMMENT ON COLUMN entries.published_by IS NULL;
COMMENT ON COLUMN entries.published_at IS NULL;
COMMENT ON COLUMN entries.version_history IS NULL;
COMMENT ON COLUMN entries.version IS NULL;
COMMENT ON COLUMN entries.status IS NULL;
COMMENT ON COLUMN entries.locale IS NULL;
COMMENT ON COLUMN entries.site_id IS NULL;
COMMENT ON COLUMN entries.schema_id IS NULL;
COMMENT ON COLUMN entries.id IS NULL;
COMMENT ON COLUMN entries.data_signature IS NULL;
COMMENT ON TABLE entries IS NULL;

-- fields 表注释
COMMENT ON COLUMN fields.updated_at IS NULL;
COMMENT ON COLUMN fields.created_at IS NULL;
COMMENT ON COLUMN fields.deleted_at IS NULL;
COMMENT ON COLUMN fields.conditional_display IS NULL;
COMMENT ON COLUMN fields.sort_order IS NULL;
COMMENT ON COLUMN fields.default_value IS NULL;
COMMENT ON COLUMN fields.display IS NULL;
COMMENT ON COLUMN fields.validation IS NULL;
COMMENT ON COLUMN fields.config IS NULL;
COMMENT ON COLUMN fields.field_type IS NULL;
COMMENT ON COLUMN fields.description IS NULL;
COMMENT ON COLUMN fields.label IS NULL;
COMMENT ON COLUMN fields.name IS NULL;
COMMENT ON COLUMN fields.schema_id IS NULL;
COMMENT ON COLUMN fields.id IS NULL;
COMMENT ON TABLE fields IS NULL;

-- schemas 表注释
COMMENT ON COLUMN schemas.updated_at IS NULL;
COMMENT ON COLUMN schemas.created_at IS NULL;
COMMENT ON COLUMN schemas.created_by IS NULL;
COMMENT ON COLUMN schemas.sort_order IS NULL;
COMMENT ON COLUMN schemas.is_active IS NULL;
COMMENT ON COLUMN schemas.draft_autosave_interval IS NULL;
COMMENT ON COLUMN schemas.versioning_enabled IS NULL;
COMMENT ON COLUMN schemas.preview_config IS NULL;
COMMENT ON COLUMN schemas.api_config IS NULL;
COMMENT ON COLUMN schemas.display_config IS NULL;
COMMENT ON COLUMN schemas.kind IS NULL;
COMMENT ON COLUMN schemas.description IS NULL;
COMMENT ON COLUMN schemas.slug IS NULL;
COMMENT ON COLUMN schemas.name IS NULL;
COMMENT ON COLUMN schemas.site_id IS NULL;
COMMENT ON COLUMN schemas.id IS NULL;
COMMENT ON COLUMN schemas.signature_enabled IS NULL;
COMMENT ON TABLE schemas IS NULL;

-- sites 表注释
COMMENT ON COLUMN sites.updated_at IS NULL;
COMMENT ON COLUMN sites.created_at IS NULL;
COMMENT ON COLUMN sites.created_by IS NULL;
COMMENT ON COLUMN sites.deleted_at IS NULL;
COMMENT ON COLUMN sites.is_active IS NULL;
COMMENT ON COLUMN sites.settings IS NULL;
COMMENT ON COLUMN sites.seo_keywords IS NULL;
COMMENT ON COLUMN sites.seo_description IS NULL;
COMMENT ON COLUMN sites.seo_title IS NULL;
COMMENT ON COLUMN sites.timezone IS NULL;
COMMENT ON COLUMN sites.locale IS NULL;
COMMENT ON COLUMN sites.site_url IS NULL;
COMMENT ON COLUMN sites.description IS NULL;
COMMENT ON COLUMN sites.slug IS NULL;
COMMENT ON COLUMN sites.name IS NULL;
COMMENT ON COLUMN sites.id IS NULL;
COMMENT ON TABLE sites IS NULL;

-- audit_logs 表注释
COMMENT ON COLUMN audit_logs.created_at IS NULL;
COMMENT ON COLUMN audit_logs.user_agent IS NULL;
COMMENT ON COLUMN audit_logs.ip_address IS NULL;
COMMENT ON COLUMN audit_logs.details IS NULL;
COMMENT ON COLUMN audit_logs.category IS NULL;
COMMENT ON COLUMN audit_logs.level IS NULL;
COMMENT ON COLUMN audit_logs.resource_id IS NULL;
COMMENT ON COLUMN audit_logs.resource_type IS NULL;
COMMENT ON COLUMN audit_logs.action IS NULL;
COMMENT ON COLUMN audit_logs.user_id IS NULL;
COMMENT ON COLUMN audit_logs.site_id IS NULL;
COMMENT ON COLUMN audit_logs.id IS NULL;
COMMENT ON COLUMN audit_logs.data_signature IS NULL;
COMMENT ON TABLE audit_logs IS NULL;

-- system_config 表注释
COMMENT ON COLUMN system_config.updated_at IS NULL;
COMMENT ON COLUMN system_config.created_at IS NULL;
COMMENT ON COLUMN system_config.is_public IS NULL;
COMMENT ON COLUMN system_config.description IS NULL;
COMMENT ON COLUMN system_config.value_type IS NULL;
COMMENT ON COLUMN system_config.config_value IS NULL;
COMMENT ON COLUMN system_config.config_key IS NULL;
COMMENT ON COLUMN system_config.id IS NULL;
COMMENT ON TABLE system_config IS NULL;

-- system_user_roles 表注释
COMMENT ON COLUMN system_user_roles.created_at IS NULL;
COMMENT ON COLUMN system_user_roles.role_id IS NULL;
COMMENT ON COLUMN system_user_roles.user_id IS NULL;
COMMENT ON COLUMN system_user_roles.id IS NULL;
COMMENT ON TABLE system_user_roles IS NULL;

-- system_roles 表注释
COMMENT ON COLUMN system_roles.updated_at IS NULL;
COMMENT ON COLUMN system_roles.created_at IS NULL;
COMMENT ON COLUMN system_roles.deleted_at IS NULL;
COMMENT ON COLUMN system_roles.permissions IS NULL;
COMMENT ON COLUMN system_roles.is_system IS NULL;
COMMENT ON COLUMN system_roles.description IS NULL;
COMMENT ON COLUMN system_roles.name IS NULL;
COMMENT ON COLUMN system_roles.id IS NULL;
COMMENT ON TABLE system_roles IS NULL;

-- system_users 表注释
COMMENT ON COLUMN system_users.updated_at IS NULL;
COMMENT ON COLUMN system_users.created_at IS NULL;
COMMENT ON COLUMN system_users.deleted_at IS NULL;
COMMENT ON COLUMN system_users.last_login_ip IS NULL;
COMMENT ON COLUMN system_users.last_login_at IS NULL;
COMMENT ON COLUMN system_users.is_super_admin IS NULL;
COMMENT ON COLUMN system_users.status IS NULL;
COMMENT ON COLUMN system_users.avatar_url IS NULL;
COMMENT ON COLUMN system_users.nickname IS NULL;
COMMENT ON COLUMN system_users.password_hash IS NULL;
COMMENT ON COLUMN system_users.email IS NULL;
COMMENT ON COLUMN system_users.id IS NULL;
COMMENT ON TABLE system_users IS NULL;

-- =============================================================================
-- 6. 删除预置角色数据（可选，谨慎操作）
-- =============================================================================

-- 注意：如果角色已经被使用，不应该删除
-- DELETE FROM system_roles WHERE id IN (
--     '00000000-0000-0000-0000-000000000101',
--     '00000000-0000-0000-0000-000000000103'
-- );
