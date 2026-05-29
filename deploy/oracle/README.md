# Oracle Cloud Deployment

This deployment path targets an Oracle Cloud Always Free VM with Docker Compose.

## Architecture

- `finance-tracker`: Go app container.
- `finance_data`: persistent Docker volume mounted at `/data` for SQLite.
- `caddy`: HTTPS reverse proxy with automatic certificates.
- Public ports: `80` and `443`.

## Oracle VM Checklist

1. Create an Oracle Cloud account.
2. Create an Always Free eligible VM in your home region.
   - Ubuntu 24.04 or Oracle Linux 9 are both fine.
   - Ampere A1 is preferred if capacity is available.
3. Open ingress ports in the VM subnet security list or network security group:
   - TCP `22` for SSH.
   - TCP `80` for HTTP certificate challenge.
   - TCP `443` for HTTPS.
4. Point a DNS `A` record for your domain to the VM public IPv4 address.
5. SSH into the VM and install Docker Engine plus the Compose plugin.

## Deploy

Clone the repo on the VM, then copy and edit the environment file:

```sh
cd finance-tracker/deploy/oracle
cp .env.example .env
nano .env
```

Set:

- `DOMAIN`: the domain pointing to this VM.
- `APP_USERNAME`: Basic Auth username.
- `APP_PASSWORD`: long random Basic Auth password.

Start the stack:

```sh
docker compose up -d --build
```

Check status:

```sh
docker compose ps
docker compose logs -f finance-tracker
docker compose logs -f caddy
```

The app should be available at `https://$DOMAIN`.

## Backup SQLite

Create a backup from the running app volume:

```sh
docker compose exec finance-tracker sqlite3 /data/finance-tracker.db ".backup '/data/finance-tracker-backup.db'"
docker compose cp finance-tracker:/data/finance-tracker-backup.db ./finance-tracker-backup.db
```

Store backups outside the VM as well.
