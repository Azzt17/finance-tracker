DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_google_id;
ALTER TABLE users DROP COLUMN email;
ALTER TABLE users DROP COLUMN google_id;
