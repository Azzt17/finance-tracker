#!/bin/bash

# Konfigurasi
BACKUP_DIR="/home/deploy/backups"
LOG_FILE="/var/log/finance-backup.log"
DATE=$(date +%Y-%m-%d)
BACKUP_FILE="finance-tracker-${DATE}.db"
CONTAINER_NAME="cloudflare-finance-tracker-1"

# Buat direktori backup jika belum ada
mkdir -p "$BACKUP_DIR"

echo "=== Backup dimulai: $(date) ===" >> "$LOG_FILE"

# 1. Panggil endpoint internal untuk membuat VACUUM INTO snapshot
echo "Memicu endpoint /internal/backup..." >> "$LOG_FILE"
RESPONSE=$(curl -s -X POST http://localhost:8080/internal/backup)
if [[ $RESPONSE != *"\"status\":\"ok\""* ]]; then
    echo "ERROR: Gagal memicu backup internal. Response: $RESPONSE" >> "$LOG_FILE"
    exit 1
fi

echo "Snapshot berhasil dibuat di dalam container." >> "$LOG_FILE"

# 2. Salin file dari container ke host
echo "Menyalin file dari container..." >> "$LOG_FILE"
docker cp "${CONTAINER_NAME}:/data/backup/finance-tracker.db" "${BACKUP_DIR}/${BACKUP_FILE}" 2>> "$LOG_FILE"
# Wait, my backup code writes to /data/backup/finance-tracker_YYYYMMDD_HHMMSS.db.
# I need to handle the dynamic filename, or I can just copy the whole directory and sync it!
