ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

-- Make the default initial user an admin
UPDATE users SET role = 'admin' WHERE id = 1;
