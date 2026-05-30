#!/bin/bash

# Konfigurasi
BACKUP_DIR="$HOME/backups"
LOG_FILE="$HOME/finance-backup.log"
# Mencari nama container backend secara otomatis
CONTAINER_NAME=$(docker ps --format '{{.Names}}' | grep finance-tracker | grep -v cloudflared | head -n 1)

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

# Ekstrak path file dari response JSON tanpa menggunakan jq
BACKUP_PATH=$(echo $RESPONSE | grep -o '"backup_path":"[^"]*' | grep -o '[^"]*$')
FILENAME=$(basename "$BACKUP_PATH")

echo "Snapshot berhasil dibuat di container pada path: $BACKUP_PATH" >> "$LOG_FILE"

# 2. Salin file dari container ke host
echo "Menyalin file $FILENAME dari container..." >> "$LOG_FILE"
docker cp "${CONTAINER_NAME}:${BACKUP_PATH}" "${BACKUP_DIR}/${FILENAME}" 2>> "$LOG_FILE"

# 3. Sinkronisasi dengan Google Drive menggunakan rclone
echo "Mengunggah ke Google Drive..." >> "$LOG_FILE"
rclone copy "${BACKUP_DIR}/${FILENAME}" gdrive:finance-tracker-backup/ 2>> "$LOG_FILE"

# 4. Hapus file backup lokal di VPS yang usianya lebih dari 7 hari
find "$BACKUP_DIR" -name "finance-tracker_*.db" -type f -mtime +7 -delete

echo "=== Backup selesai: $(date) ===" >> "$LOG_FILE"
