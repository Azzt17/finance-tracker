CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	icon_emoji TEXT,
	is_quick_add INTEGER NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS budget_allocation (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	year_month TEXT NOT NULL UNIQUE,
	total_budget INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS savings_goals (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	target_amount INTEGER NOT NULL,
	current_saved INTEGER NOT NULL DEFAULT 0,
	year_month TEXT NOT NULL,
	is_achieved INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	client_transaction_id TEXT NOT NULL UNIQUE,
	amount INTEGER NOT NULL,
	category_id INTEGER,
	note TEXT,
	transacted_at TEXT NOT NULL DEFAULT (datetime('now')),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	is_synced INTEGER NOT NULL DEFAULT 1,
	FOREIGN KEY (category_id) REFERENCES categories (id)
		ON UPDATE CASCADE
		ON DELETE RESTRICT
);

INSERT OR IGNORE INTO schema_migrations (version, name)
VALUES (1, 'initial_schema');
