# Finance Tracker

Personal finance tracker built with Go, SQLite, and embedded static assets.

## Local Development

```sh
go test ./...
go run ./cmd/server
```

The app listens on `:8080` by default and stores local data in `finance-tracker.db`.

## Configuration

- `ADDR`: HTTP listen address. Default: `:8080`.
- `DATABASE_URL`: SQLite DSN. In Docker, use `file:/data/finance-tracker.db?_foreign_keys=on&_busy_timeout=5000`.
- `APP_USERNAME` and `APP_PASSWORD`: enable Basic Auth when both are set.
- `CORS_ALLOWED_ORIGIN`: optional allowed origin for cross-origin API access. Leave empty for same-origin use.

## Cloudflare Tunnel Deployment

For personal use, the recommended deployment path is Docker Compose plus Cloudflare Tunnel:

```sh
cd deploy/cloudflare
cp .env.example .env
docker compose up -d --build
```

See `deploy/cloudflare/README.md` for the full setup.
