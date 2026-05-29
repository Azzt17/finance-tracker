# Cloudflare Tunnel Deployment

This deployment path runs Finance Tracker on your own machine and exposes it through Cloudflare Tunnel.

## Architecture

- `finance-tracker`: Go app container.
- `finance_data`: persistent Docker volume mounted at `/data` for SQLite.
- `cloudflared`: Cloudflare Tunnel connector.
- No public inbound ports are required on your machine.

## Prerequisites

- A machine that can stay online.
- Docker Engine and Docker Compose.
- A Cloudflare account.
- A domain managed by Cloudflare.

## Cloudflare Setup

1. Open Cloudflare Zero Trust.
2. Go to `Networks` -> `Tunnels`.
3. Create a tunnel.
4. Choose Docker as the connector environment.
5. Copy the tunnel token.
6. Add a public hostname:
   - Hostname: your chosen subdomain, for example `finance.example.com`.
   - Service type: `HTTP`.
   - Service URL: `finance-tracker:8080`.

## Deploy

Copy and edit the environment file:

```sh
cd deploy/cloudflare
cp .env.example .env
nano .env
```

Set:

- `APP_USERNAME`: Basic Auth username.
- `APP_PASSWORD`: long random Basic Auth password.
- `CLOUDFLARE_TUNNEL_TOKEN`: token from Cloudflare Zero Trust.

Start the stack:

```sh
docker compose up -d --build
```

Check status:

```sh
docker compose ps
docker compose logs -f finance-tracker
docker compose logs -f cloudflared
```

The app should be available at your configured Cloudflare hostname over HTTPS.

## Backup SQLite

Create a backup from the running app volume:

```sh
docker compose exec finance-tracker sqlite3 /data/finance-tracker.db ".backup '/data/finance-tracker-backup.db'"
docker compose cp finance-tracker:/data/finance-tracker-backup.db ./finance-tracker-backup.db
```

Store backups outside this machine as well.
