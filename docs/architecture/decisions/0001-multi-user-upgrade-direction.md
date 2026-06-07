# 1. Multi-User Upgrade Direction

Date: 2026-06-07

## Status

Accepted

## Context

Finance Tracker has reached a stable post-MVP baseline (v0.1.0) as a single-user personal finance tool. The next major objective is to evolve the application to support multiple users securely, enabling shared tracking or independent accounts within a single deployment.

Upgrading to a multi-user architecture requires profound changes to authentication, database schema, data ownership, RBAC (Role-Based Access Control), and frontend offline synchronization. Attempting to implement these changes simultaneously poses significant risks to stability, security, and developer velocity.

## Decision

We will transition to a multi-user architecture using a phased, incremental approach rather than a "big bang" rewrite. The upgrade will strictly follow these milestones:

1. **v0.1.1**: Urgent Security Cleanup (removing committed secrets, rotating credentials)
2. **v0.2.0**: Repository Workflow and Planning Baseline (ADRs, CI/CD, PR templates)
3. **v0.3.0**: Authentication Foundation (user identity, login/logout)
4. **v0.4.0**: Multi-User Data Ownership (assigning financial records to users)
5. **v0.5.0**: RBAC and Admin Capabilities
6. **v0.6.0**: Frontend Auth and User-Scoped Offline Behavior
7. **v0.7.0+**: Feature Expansion

## Consequences

- **Positive:** Reduces risk by allowing continuous integration, isolated testing, and independent security reviews for each subsystem.
- **Positive:** Keeps the application deployable and functional at every milestone.
- **Negative:** Requires rigorous discipline to avoid merging incomplete features across milestone boundaries. Some features will temporarily sit behind feature flags or remain inaccessible until subsequent milestones are complete.
