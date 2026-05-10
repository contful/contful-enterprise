-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Migration 000002: Rollback table renaming
-- =============================================================================

-- Rollback: Rename tables from system_* back to global_*
DO $$
BEGIN
    -- Rename system_users back to global_users
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'system_users') THEN
        ALTER TABLE system_users RENAME TO global_users;
        ALTER INDEX idx_system_users_email RENAME TO idx_global_users_email;
        ALTER INDEX idx_system_users_status RENAME TO idx_global_users_status;
    END IF;

    -- Rename system_roles back to global_roles
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'system_roles') THEN
        ALTER TABLE system_roles RENAME TO global_roles;
        ALTER INDEX idx_system_roles_name RENAME TO idx_global_roles_name;
    END IF;

    -- Rename system_user_roles back to global_user_roles
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'system_user_roles') THEN
        ALTER TABLE system_user_roles RENAME TO global_user_roles;
        ALTER INDEX idx_system_user_roles_user RENAME TO idx_global_user_roles_user;
        ALTER INDEX idx_system_user_roles_role RENAME TO idx_global_user_roles_role;
    END IF;

    -- Rename tokens back to api_tokens
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'tokens') THEN
        ALTER TABLE tokens RENAME TO api_tokens;
        ALTER INDEX idx_tokens_site RENAME TO idx_api_tokens_site;
        ALTER INDEX idx_tokens_hash RENAME TO idx_api_tokens_hash;
        ALTER INDEX idx_tokens_status RENAME TO idx_api_tokens_status;
        ALTER INDEX idx_tokens_expires RENAME TO idx_api_tokens_expires;
    END IF;
END $$;
