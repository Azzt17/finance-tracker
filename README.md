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
- `DATABASE_URL`: SQLite DSN. On Fly, use `file:/data/finance-tracker.db?_foreign_keys=on&_busy_timeout=5000`.
- `APP_USERNAME` and `APP_PASSWORD`: enable Basic Auth when both are set.
- `CORS_ALLOWED_ORIGIN`: optional allowed origin for cross-origin API access. Leave empty for same-origin use.

## Fly.io Deployment

Create the persistent volume before first deploy:

```sh
fly volumes create finance_data --size 1 --region sin
fly secrets set APP_USERNAME='your-user' APP_PASSWORD='your-password'
fly deploy
```

The SQLite database is stored under `/data`, which is mounted from the Fly volume.
