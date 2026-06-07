# 2. Authentication Strategy

Date: 2026-06-07

## Status

Accepted

## Context

As part of the multi-user upgrade (v0.3.0), Finance Tracker must introduce user identities and a robust authentication mechanism to replace the current basic, single-user access control. The application consists of a Go backend and a browser-based SPA frontend with offline/PWA behavior.

Storing sensitive, long-lived authentication tokens (like JWTs) in `localStorage` exposes the application to XSS (Cross-Site Scripting) attacks, which is unacceptable for a security-sensitive financial application. Furthermore, implementing full JWT flows with refresh tokens adds unnecessary complexity for the initial multi-user implementation.

## Decision

We will implement **Server-Side Sessions with Secure HttpOnly Cookies** as the primary authentication strategy for the first browser-based implementation.

- The server will generate a secure session ID upon successful login.
- The session ID will be stored in an `HttpOnly`, `Secure`, `SameSite=Strict` cookie, rendering it inaccessible to frontend JavaScript.
- Password hashing will use a modern algorithm (e.g., bcrypt or argon2) and password hashes will never be exposed in API responses.
- The frontend will rely on backend validation to determine authentication state, rather than parsing tokens locally.

## Consequences

- **Positive:** Highly secure against XSS attacks since JavaScript cannot access the session token.
- **Positive:** Simplifies backend logic, as we can easily invalidate sessions on the server side (e.g., during a password reset or manual logout).
- **Negative:** State must be maintained on the server (e.g., in the database or an in-memory store), adding slight overhead compared to stateless JWTs.
- **Negative:** If we later introduce native mobile apps or external API integrations, we may need to reconsider or supplement this with a stateless token strategy (e.g., JWT) designed specifically for those clients.
