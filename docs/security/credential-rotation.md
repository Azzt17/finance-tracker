# Credential Rotation Guide

This document outlines the standard operating procedure for rotating credentials if they are ever suspected of being exposed or compromised.

## General Principles

1. **Never commit secrets to the repository.**
2. **Use `.env.example`** to document required environment variables without including real values.
3. **If a secret is exposed**, consider it immediately compromised and rotate it.

## Rotation Steps

### 1. Identify the Exposed Secret
Identify which secret was exposed (e.g., `APP_PASSWORD`, `DATABASE_URL`, Cloudflare Token). Do not log the exposed secret.

### 2. Rotate the Credential at the Source
- **Database Passwords:** Access the database server and change the user password.
- **Application Passwords:** For `APP_PASSWORD`, simply update your `.env` file on the deployment server with a newly generated strong password. Restart the service.
- **Third-Party Tokens (e.g., Cloudflare):** Go to the third-party dashboard, revoke the old token, and generate a new one.

### 3. Update the Production Environment
Connect to the VPS or deployment server, update the `.env` file with the newly generated secret, and restart the application container.
```bash
# Example
nano .env
docker compose down
docker compose up -d
```

### 4. Remove the Secret from Git History (If Applicable)
If the secret was committed to the repository, use tools like `git filter-repo` or BFG Repo-Cleaner to scrub the secret from the repository's history, or rotate the secret and ignore the historical commit (since the secret is no longer valid).

### 5. Verify the Secret Scan
Ensure that the `gitleaks` step in CI passes after the rotation and cleanup to prevent future exposures.
