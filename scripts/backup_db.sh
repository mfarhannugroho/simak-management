#!/usr/bin/env bash
# ─── PT INKA SIMAK Database Backup Script ──────────────────────────────────────
# Backs up PostgreSQL database using pg_dump with gzip compression and retention

set -eo pipefail

BACKUP_DIR="${BACKUP_DIR:-/backups/simak}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-simak_db}"
DB_USER="${DB_USER:-simak_user}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILENAME="simak_backup_${TIMESTAMP}.sql.gz"
FILEPATH="${BACKUP_DIR}/${FILENAME}"

mkdir -p "${BACKUP_DIR}"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting backup of ${DB_NAME} to ${FILEPATH}..."

# Export password if set
if [[ -n "${DB_PASSWORD}" ]]; then
  export PGPASSWORD="${DB_PASSWORD}"
fi

# Execute pg_dump with compression
pg_dump -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -F c -b -v -f "${FILEPATH}"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup successfully created: ${FILEPATH} ($(du -h "${FILEPATH}" | cut -f1))"

# Purge backups older than retention window
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaning backups older than ${RETENTION_DAYS} days..."
find "${BACKUP_DIR}" -name "simak_backup_*.sql.gz" -type f -mtime +"${RETENTION_DAYS}" -delete

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup maintenance complete."
