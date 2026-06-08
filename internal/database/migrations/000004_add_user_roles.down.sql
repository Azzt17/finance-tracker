-- SQLite does not support dropping columns directly without recreating the table.
-- Since this is a down migration for RBAC, we recreate the users table.

CREATE TABLE users_new (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO users_new (id, username, password_hash, created_at)
SELECT id, username, password_hash, created_at FROM users;

DROP TABLE users;
ALTER TABLE users_new RENAME TO users;
