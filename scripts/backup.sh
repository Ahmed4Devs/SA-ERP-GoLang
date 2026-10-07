#!/usr/bin/env bash
# Coordinated logical database + document recovery bundle; invoke via backup-coordinated.sh.
set -euo pipefail
umask 077
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PYTHON_BIN="${PYTHON_BIN:-python3}"
BACKUP_DIR="${BACKUP_DIR:-/backups}"
STORAGE_LOCAL_DIR="${STORAGE_LOCAL_DIR:-/app/data/storage}"
DB_HOST="${POSTGRES_HOST:-postgres}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-erp_backup}"
DB_NAME="${POSTGRES_DB:-erp_db}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%S)_$$"
[[ "$DB_NAME" =~ ^[A-Za-z_][A-Za-z0-9_]*$ && "$RETENTION_DAYS" =~ ^[0-9]+$ ]] || { echo "Invalid database name or retention" >&2; exit 1; }
[[ "${BACKUP_QUIESCED:-false}" == true ]] || { echo "Writers must be stopped; use backup-coordinated.sh or set BACKUP_QUIESCED=true after isolating all writers." >&2; exit 1; }
[[ -d "$STORAGE_LOCAL_DIR" ]] || { echo "Document storage directory is missing" >&2; exit 1; }
[[ -n "${BACKUP_ENCRYPTION_KEY:-}" || "${ALLOW_UNENCRYPTED_BACKUP:-false}" == true ]] || { echo "BACKUP_ENCRYPTION_KEY is required" >&2; exit 1; }
for command in pg_dump "$PYTHON_BIN" tar sha256sum; do command -v "$command" >/dev/null || { echo "$command is required" >&2; exit 1; }; done
mkdir -p "$BACKUP_DIR"
STAGE="$(mktemp -d "${BACKUP_DIR}/.erp-stage.XXXXXX")"
cleanup() { rm -rf -- "$STAGE"; }
trap cleanup EXIT
export PGPASSWORD="${POSTGRES_PASSWORD:-}"
pg_dump --host="$DB_HOST" --port="$DB_PORT" --username="$DB_USER" --dbname="$DB_NAME" --format=custom --compress=6 --file="$STAGE/database.dump"
mkdir "$STAGE/attachments"
# Reject symlinks and special files; documents are regular immutable storage objects.
if [[ -n "$(find "$STORAGE_LOCAL_DIR" ! -type d ! -type f -print -quit)" ]]; then echo "Unsupported special file in document storage" >&2; exit 1; fi
cp -a "$STORAGE_LOCAL_DIR/." "$STAGE/attachments/"
ARCHIVE="$STAGE/bundle.tar.gz"
"$PYTHON_BIN" "$SCRIPT_DIR/backup_bundle.py" create "$STAGE" "$ARCHIVE" --database "$DB_NAME"
FINAL_FILE="$BACKUP_DIR/erp_${DB_NAME}_${TIMESTAMP}.tar.gz"
if [[ -n "${BACKUP_ENCRYPTION_KEY:-}" ]]; then
    export BACKUP_ENCRYPTION_KEY
    FINAL_FILE="${FINAL_FILE}.enc"
    openssl enc -aes-256-cbc -salt -pbkdf2 -iter 200000 -in "$ARCHIVE" -out "$FINAL_FILE.partial" -pass env:BACKUP_ENCRYPTION_KEY
else
    cp "$ARCHIVE" "$FINAL_FILE.partial"
fi
mv "$FINAL_FILE.partial" "$FINAL_FILE"
(cd "$BACKUP_DIR"; sha256sum "$(basename "$FINAL_FILE")" > "$(basename "$FINAL_FILE").sha256"; sha256sum -c "$(basename "$FINAL_FILE").sha256")
if [[ -n "${RESTIC_REPOSITORY:-}" ]]; then
    command -v restic >/dev/null || { echo "restic is required for configured offsite replication" >&2; exit 1; }
    restic backup --tag erp-recovery "$FINAL_FILE" "$FINAL_FILE.sha256"
    restic check
elif [[ "${BACKUP_REQUIRE_OFFSITE:-false}" == true ]]; then
    echo "Offsite replication is required but RESTIC_REPOSITORY is not configured" >&2; exit 1
fi
# Retention runs only after all required destinations have accepted this bundle.
find "$BACKUP_DIR" -maxdepth 1 -type f -name "erp_${DB_NAME}_*.tar.gz*" ! -name '*.partial' -mtime "+$RETENTION_DAYS" -delete
printf '%s\n' "$(date +%s)" > "$BACKUP_DIR/.last-success.tmp"
mv "$BACKUP_DIR/.last-success.tmp" "$BACKUP_DIR/last-success"
if [[ -n "${BACKUP_METRICS_FILE:-}" ]]; then
    printf 'erp_backup_last_success_timestamp_seconds %s\nerp_backup_archive_bytes %s\n' "$(date +%s)" "$(wc -c < "$FINAL_FILE")" > "${BACKUP_METRICS_FILE}.tmp"
    chmod 0644 "${BACKUP_METRICS_FILE}.tmp"
    mv "${BACKUP_METRICS_FILE}.tmp" "$BACKUP_METRICS_FILE"
fi
printf 'Backup completed: %s\n' "$FINAL_FILE"
