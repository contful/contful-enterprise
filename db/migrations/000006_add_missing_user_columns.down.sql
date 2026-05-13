-- Remove columns added in 000006
ALTER TABLE system_users DROP COLUMN IF EXISTS phone;
ALTER TABLE system_users DROP COLUMN IF EXISTS department;
