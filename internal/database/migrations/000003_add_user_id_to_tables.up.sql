-- Ensure a default user exists for existing data
INSERT OR IGNORE INTO users (id, username, password_hash) 
VALUES (1, 'admin', 'unset_password_hash_please_change');

-- 1. Create all new tables first
CREATE TABLE categories_new (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL DEFAULT 1,
	name TEXT NOT NULL,
	icon_emoji TEXT,
	is_quick_add INTEGER NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	UNIQUE(user_id, name),
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE budget_allocation_new (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL DEFAULT 1,
	year_month TEXT NOT NULL,
	total_budget INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	UNIQUE(user_id, year_month),
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE savings_goals_new (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL DEFAULT 1,
	name TEXT NOT NULL,
	target_amount INTEGER NOT NULL,
	current_saved INTEGER NOT NULL DEFAULT 0,
	year_month TEXT NOT NULL,
	is_achieved INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE transactions_new (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL DEFAULT 1,
	client_transaction_id TEXT NOT NULL,
	amount INTEGER NOT NULL,
	category_id INTEGER,
	note TEXT,
	transacted_at TEXT NOT NULL DEFAULT (datetime('now')),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	is_synced INTEGER NOT NULL DEFAULT 1,
	UNIQUE(user_id, client_transaction_id),
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
	FOREIGN KEY (category_id) REFERENCES categories_new (id) ON UPDATE CASCADE ON DELETE RESTRICT
);

-- 2. Copy data
INSERT INTO categories_new (id, user_id, name, icon_emoji, is_quick_add, sort_order, created_at)
SELECT id, 1, name, icon_emoji, is_quick_add, sort_order, created_at FROM categories;

INSERT INTO budget_allocation_new (id, user_id, year_month, total_budget, created_at, updated_at)
SELECT id, 1, year_month, total_budget, created_at, updated_at FROM budget_allocation;

INSERT INTO savings_goals_new (id, user_id, name, target_amount, current_saved, year_month, is_achieved, created_at)
SELECT id, 1, name, target_amount, current_saved, year_month, is_achieved, created_at FROM savings_goals;

INSERT INTO transactions_new (id, user_id, client_transaction_id, amount, category_id, note, transacted_at, created_at, is_synced)
SELECT id, 1, client_transaction_id, amount, category_id, note, transacted_at, created_at, is_synced FROM transactions;

-- 3. Drop old tables in reverse dependency order
DROP TABLE transactions;
DROP TABLE budget_allocation;
DROP TABLE savings_goals;
DROP TABLE categories;

-- 4. Rename new tables
ALTER TABLE categories_new RENAME TO categories;
ALTER TABLE budget_allocation_new RENAME TO budget_allocation;
ALTER TABLE savings_goals_new RENAME TO savings_goals;
ALTER TABLE transactions_new RENAME TO transactions;
