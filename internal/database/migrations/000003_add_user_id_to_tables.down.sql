-- Categories
CREATE TABLE categories_old (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	icon_emoji TEXT,
	is_quick_add INTEGER NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
INSERT INTO categories_old (id, name, icon_emoji, is_quick_add, sort_order, created_at)
SELECT id, name, icon_emoji, is_quick_add, sort_order, created_at FROM categories;
DROP TABLE categories;
ALTER TABLE categories_old RENAME TO categories;

-- Budget Allocation
CREATE TABLE budget_allocation_old (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	year_month TEXT NOT NULL UNIQUE,
	total_budget INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
INSERT INTO budget_allocation_old (id, year_month, total_budget, created_at, updated_at)
SELECT id, year_month, total_budget, created_at, updated_at FROM budget_allocation;
DROP TABLE budget_allocation;
ALTER TABLE budget_allocation_old RENAME TO budget_allocation;

-- Savings Goals
CREATE TABLE savings_goals_old (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	target_amount INTEGER NOT NULL,
	current_saved INTEGER NOT NULL DEFAULT 0,
	year_month TEXT NOT NULL,
	is_achieved INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
INSERT INTO savings_goals_old (id, name, target_amount, current_saved, year_month, is_achieved, created_at)
SELECT id, name, target_amount, current_saved, year_month, is_achieved, created_at FROM savings_goals;
DROP TABLE savings_goals;
ALTER TABLE savings_goals_old RENAME TO savings_goals;

-- Transactions
CREATE TABLE transactions_old (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	client_transaction_id TEXT NOT NULL UNIQUE,
	amount INTEGER NOT NULL,
	category_id INTEGER,
	note TEXT,
	transacted_at TEXT NOT NULL DEFAULT (datetime('now')),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	is_synced INTEGER NOT NULL DEFAULT 1,
	FOREIGN KEY (category_id) REFERENCES categories (id) ON UPDATE CASCADE ON DELETE RESTRICT
);
INSERT INTO transactions_old (id, client_transaction_id, amount, category_id, note, transacted_at, created_at, is_synced)
SELECT id, client_transaction_id, amount, category_id, note, transacted_at, created_at, is_synced FROM transactions;
DROP TABLE transactions;
ALTER TABLE transactions_old RENAME TO transactions;
