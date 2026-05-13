-- Copyright © 2026-present reepu.com
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Migration 000002: Rename tables for better naming consistency
-- =============================================================================

-- Rename tables from global_* to system_*
DO $$
BEGIN
    -- Rename global_users to system_users
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'global_users') THEN
        ALTER TABLE global_users RENAME TO system_users;
        ALTER INDEX idx_global_users_email RENAME TO idx_system_users_email;
        ALTER INDEX idx_global_users_status RENAME TO idx_system_users_status;
    END IF;

    -- Rename global_roles to system_roles
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'global_roles') THEN
        ALTER TABLE global_roles RENAME TO system_roles;
        ALTER INDEX idx_global_roles_name RENAME TO idx_system_roles_name;
    END IF;

    -- Rename global_user_roles to system_user_roles
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'global_user_roles') THEN
        ALTER TABLE global_user_roles RENAME TO system_user_roles;
        ALTER INDEX idx_global_user_roles_user RENAME TO idx_system_user_roles_user;
        ALTER INDEX idx_global_user_roles_role RENAME TO idx_system_user_roles_role;
    END IF;

    -- Rename api_tokens to tokens
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'api_tokens') THEN
        ALTER TABLE api_tokens RENAME TO tokens;
        ALTER INDEX idx_api_tokens_site RENAME TO idx_tokens_site;
        ALTER INDEX idx_api_tokens_hash RENAME TO idx_tokens_hash;
        ALTER INDEX idx_api_tokens_status RENAME TO idx_tokens_status;
        ALTER INDEX idx_api_tokens_expires RENAME TO idx_tokens_expires;
    END IF;
END $$;
