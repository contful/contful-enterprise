-- Add missing columns to system_users table
ALTER TABLE system_users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
ALTER TABLE system_users ADD COLUMN IF NOT EXISTS department VARCHAR(100);
