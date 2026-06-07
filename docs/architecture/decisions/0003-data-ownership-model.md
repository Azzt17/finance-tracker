# 3. Data Ownership Model

Date: 2026-06-07

## Status

Accepted

## Context

In the current single-user MVP, all financial data (transactions, categories, budgets, savings goals) exists globally in the database. Moving to a multi-user environment requires that every piece of financial data belongs to a specific user, and cross-user data leakage must be strictly prevented. Additionally, we need a strategy to handle the existing data from the single-user deployment without data loss.

## Decision

We will implement a strict **User-Scoped Data Ownership Model** (v0.4.0):

1. **Schema Changes:** All financial entity tables (`transactions`, `categories`, `budgets`, `savings_goals`) will receive a mandatory `user_id` foreign key referencing the new `users` table.
2. **Query Scoping:** Every repository query that reads, updates, or deletes financial data must require the authenticated user's `user_id` to guarantee isolation. Global queries will be strictly prohibited unless executed within an administrative context (RBAC).
3. **Migration Strategy:** During the database migration, a default "owner" account will be created. All existing orphaned financial records will be automatically assigned to this initial owner account to preserve data integrity.
4. **Constraint Isolation:** Existing global unique constraints (e.g., global category names or global monthly budgets) will be redefined as composite unique constraints tied to the `user_id` (e.g., `UNIQUE(user_id, name)`).

## Consequences

- **Positive:** Strong guarantees against cross-tenant data leaks.
- **Positive:** Existing users migrating from v0.1.0 will not lose their financial history.
- **Negative:** Repository interfaces must be updated across the entire codebase to accept `user_id` as a parameter, requiring a widespread (though mechanical) refactor.
- **Negative:** Frontend caching strategies (IndexedDB, service workers) must be overhauled to isolate data per user and clear caches completely upon user logout to prevent offline data leaks.
