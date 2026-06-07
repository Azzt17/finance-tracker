# Finance Tracker Upgrade Analysis

## Status

Draft

## Purpose

This document analyzes the planned upgrade of Finance Tracker from a post-MVP personal finance tracker into a complete multi-user finance tracking application.

This is not an implementation document yet. Its purpose is to clarify scope, risks, architectural pressure points, and execution order before issues, ADRs, and pull requests are created.

## Source Documents

- `upgrade-plan.md`
- `v0.1.0` post-MVP release baseline
- Existing repository structure and CI/CD workflow

## Executive Summary

Finance Tracker has reached a stable post-MVP baseline for personal use. The current application is already useful, deployed, and validated through real usage.

The next upgrade is not a simple feature addition. Moving from a personal single-user application to a multi-user application requires changes to identity, authentication, authorization, data ownership, database constraints, frontend state, offline behavior, deployment security, and release workflow.

The upgrade must be executed incrementally through issues, small branches, pull requests, CI checks, and releases.

## Current Architecture Summary

The current application uses a relatively simple architecture:

```text
Frontend SPA
  -> Go HTTP handlers
  -> Repository layer
  -> SQLite database
```

Important current characteristics:

- Frontend uses Alpine.js, Pico CSS, Chart.js, IndexedDB, and a service worker.
- Backend is a Go binary with embedded frontend assets.
- Middleware stack is still simple: recovery, logging, and Basic Auth.
- Handler layer calls repository interfaces directly.
- There is almost no service layer except backup-related service code.
- SQLite is used with WAL mode and a single open connection.
- CI/CD already exists and includes quality, security, build, image, and deployment jobs.
- The application has already been deployed and used as a personal finance tracker.

## Main Gaps for Multi-User Support

The current design has several critical gaps for multi-user usage.

### 1. No User Identity Model

There is no `users` table yet. The application has no internal concept of a user account, user profile, user status, user role, or user ownership.

### 2. Basic Auth Is Not Enough

Basic Auth is acceptable as a simple protection layer for personal self-hosted usage, but it is not enough for a multi-user application.

A multi-user application needs per-user authentication, logout behavior, credential storage, password hashing, session or token lifecycle, and account state.

### 3. No Data Ownership

Current data tables do not have `user_id`.

This affects:

- transactions
- categories
- budget allocation
- savings goals
- analytics
- sync
- export

For multi-user safety, every financial record must belong to a user, and every protected query must be scoped by authenticated user identity.

### 4. Global Unique Constraints Conflict With Multi-User Usage

Current constraints such as global category name uniqueness and global monthly budget uniqueness are not compatible with multi-user usage.

Examples:

- Two users should be able to create a category named `Food`.
- Two users should be able to have a budget for the same month.
- Offline sync IDs should not collide globally if they are user-specific.

### 5. Authorization Is Not Yet Defined

The application does not yet distinguish between owner, admin, member, or viewer capabilities.

RBAC should not be implemented before user identity and data ownership are stable.

### 6. Offline-First Behavior Becomes Security-Sensitive

IndexedDB, service worker cache, and offline sync queue are safe enough for a personal app, but they become sensitive in a multi-user app.

The frontend must prevent cached or queued data from one user being visible or submitted under another user account.

### 7. Service Layer Is Needed Gradually

The current Handler -> Repository flow is simple and productive. However, multi-user logic introduces more business rules:

- ownership checks
- auth flow
- session lifecycle
- password changes
- budget alerts
- recurring transactions
- notifications
- import/export validation

A service layer should be introduced gradually around new multi-user and business logic, not as a massive refactor.

### 8. Security Cleanup Must Precede Feature Work

Before building new auth features, exposed credentials and `.env` handling must be fixed.

Credential rotation and repository hygiene are urgent because new auth work should not be built on top of a repository with known secret-management problems.

## Scope Correction

The master upgrade plan is valuable, but its current scope is too large to execute as a single milestone.

The plan includes:

- authentication
- refresh tokens
- RBAC
- admin APIs
- user-owned data migration
- service layer
- frontend auth overhaul
- user-scoped PWA behavior
- recurring transactions
- notifications
- settings
- audit logs
- import/export
- security headers
- rate limiting
- PostgreSQL option
- API v2 specification

This should not be implemented as one branch, one issue, or one pull request.

The correct execution model is incremental:

```text
baseline -> security cleanup -> workflow baseline -> architecture decisions -> auth foundation -> data ownership -> RBAC -> frontend integration -> feature upgrades
```

## Recommended Milestone Breakdown

### v0.1.0 — Post-MVP Baseline

Status: Done.

Purpose:

- freeze the stable personal-use version
- create rollback/reference point before major architectural changes

### v0.1.1 — Urgent Security Cleanup

Purpose:

- rotate exposed credentials
- stop tracking `.env` files
- update `.gitignore`
- add or update `.env.example`
- verify secret scan
- document credential rotation

This should happen before new auth work.

### v0.2.0 — Repository Workflow and Planning Baseline

Purpose:

- add pull request template
- add issue templates
- add changelog
- add architecture analysis documents
- establish issue-driven and PR-based workflow

This milestone changes repository process, not application behavior.

### v0.3.0 — Authentication Foundation

Purpose:

- introduce user model
- introduce password hashing
- introduce login/logout/session strategy
- protect routes
- add auth tests

Recommended direction:

- prefer server-side session with secure HttpOnly cookies for the initial browser-based implementation
- avoid storing long-lived authentication secrets in localStorage
- document any token-based alternative in a separate ADR before implementation

### v0.4.0 — Multi-User Data Ownership

Purpose:

- add `user_id` to financial tables
- migrate existing data into an initial owner account
- scope repository queries by user ID
- prevent cross-user access
- add repository and handler tests for data isolation

### v0.5.0 — RBAC and Admin Capabilities

Purpose:

- define roles and permissions
- protect admin routes
- move system backup/admin operations behind admin authorization
- add authorization tests

RBAC should come after authentication and data ownership.

### v0.6.0 — Frontend Auth and User-Scoped Offline Behavior

Purpose:

- add login/register/profile UI
- centralize API client
- handle unauthenticated state
- clear or isolate user-specific local state
- update service worker and IndexedDB behavior

### v0.7.0+ — Feature Expansion

Purpose:

- dashboard upgrades
- recurring transactions
- notifications
- import/export
- settings
- audit logs
- analytics improvements

These should not block the core multi-user security model.

## Initial Architectural Direction

The following points are initial direction, not final ADR decisions:

1. Multi-user support is accepted as the long-term direction.
2. The upgrade must be split into small reviewable pull requests.
3. Security cleanup must happen before auth implementation.
4. Data ownership is a prerequisite for safe multi-user behavior.
5. RBAC should be implemented after authentication and user-owned data are stable.
6. Service layer should be introduced gradually where business logic requires it.
7. Offline-first behavior must be reviewed as a security-sensitive subsystem.
8. API v2 should be introduced carefully and should not force a full rewrite in one milestone.

## Open Questions

These questions must be answered before implementation issues are created:

1. Should the first auth implementation use server-side sessions or JWT?
2. If JWT is used, where will refresh tokens be stored safely?
3. Should the first multi-user release allow public registration, or only admin-created users?
4. How should existing data be assigned to the first owner account?
5. Should v1 API remain available during transition?
6. What is the expected user count for the first multi-user version?
7. Should SQLite remain the only supported database for now?
8. What frontend data must be cleared on logout?
9. Which operations require admin role?
10. What is the minimum acceptable test coverage before v1.0.0?

## Execution Rules

During the upgrade:

- no direct push to `main`
- every change starts from an issue
- every issue maps to a small branch
- every branch is merged through a pull request
- every PR must pass CI
- security-sensitive PRs must include explicit security notes
- migration PRs must include rollback notes
- auth and data ownership PRs must include tests
- large refactors must not be mixed with feature changes

## Immediate Next Actions

1. Complete repository workflow baseline PR.
2. Create a dedicated security cleanup issue.
3. Rotate exposed credentials outside the codebase.
4. Add secret hygiene documentation.
5. Create ADR 0001 for the multi-user upgrade direction.
6. Create ADR 0002 for authentication strategy.
7. Create ADR 0003 for data ownership model.
8. Only then start implementation issues.
