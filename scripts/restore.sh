#!/usr/bin/env bash
# Restore a verified database+documents bundle. Keep all writers stopped until verification succeeds.
set -euo pipefail
umask 077
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PYTHON_BIN="${PYTHON_BIN:-python3}"
BACKUP_FILE="${1:-}"
[[ -f "$BACKUP_FILE" && -f "$BACKUP_FILE.sha256" ]] || { echo "Usage: restore.sh <bundle.tar.gz[.enc]> [--force]; checksum sidecar required" >&2; exit 1; }
[[ "${RESTORE_QUIESCED:-false}" == true ]] || { echo "Stop and isolate all application/database/storage writers, then set RESTORE_QUIESCED=true" >&2; exit 1; }
DB_HOST="${POSTGRES_HOST:-postgres}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-erp_admin}"
DB_NAME="${POSTGRES_DB:-erp_restore}"
DB_OWNER="${RESTORE_OWNER:-$DB_USER}"
STORAGE_LOCAL_DIR="${STORAGE_LOCAL_DIR:?Set STORAGE_LOCAL_DIR to an EMPTY restore destination}"
for name in "$DB_USER" "$DB_NAME" "$DB_OWNER"; do [[ "$name" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || { echo "Invalid database identifier" >&2; exit 1; }; done
for command in psql pg_restore "$PYTHON_BIN" sha256sum; do command -v "$command" >/dev/null || { echo "$command is required" >&2; exit 1; }; done
# Existing attachments are preserved separately by the operator; never delete a storage tree.
mkdir -p "$STORAGE_LOCAL_DIR"
[[ -z "$(find "$STORAGE_LOCAL_DIR" -mindepth 1 -print -quit)" ]] || { echo "Restore destination must be empty; preserve current documents in a separate location first" >&2; exit 1; }
EXPECTED="$(awk 'NR==1 {print $1}' "$BACKUP_FILE.sha256")"
[[ "$EXPECTED" =~ ^[0-9a-fA-F]{64}$ && "$(sha256sum "$BACKUP_FILE" | cut -d' ' -f1)" == "$EXPECTED" ]] || { echo "Backup checksum mismatch" >&2; exit 1; }
STAGE="$(mktemp -d "${TMPDIR:-/tmp}/erp-restore.XXXXXX")"
cleanup() { rm -rf -- "$STAGE"; }
trap cleanup EXIT
ARCHIVE="$BACKUP_FILE"
if [[ "$BACKUP_FILE" == *.enc ]]; then
    : "${BACKUP_ENCRYPTION_KEY:?Encryption key is required}"
    export BACKUP_ENCRYPTION_KEY
    ARCHIVE="$STAGE/decrypted.tar.gz"
    openssl enc -d -aes-256-cbc -pbkdf2 -iter 200000 -in "$BACKUP_FILE" -out "$ARCHIVE" -pass env:BACKUP_ENCRYPTION_KEY
fi
mkdir "$STAGE/verified"
"$PYTHON_BIN" "$SCRIPT_DIR/backup_bundle.py" verify "$STAGE/verified" "$ARCHIVE"
pg_restore --list "$STAGE/verified/database.dump" > /dev/null
if [[ "${2:-}" != --force ]]; then
    read -rp "Recreate database $DB_NAME on $DB_HOST:$DB_PORT and restore documents to $STORAGE_LOCAL_DIR? (y/N): " CONFIRM
    [[ "$CONFIRM" =~ ^[Yy]$ ]] || exit 0
fi
export PGPASSWORD="${POSTGRES_PASSWORD:-}"
PSQL=(psql --host="$DB_HOST" --port="$DB_PORT" --username="$DB_USER" --dbname=postgres --set=ON_ERROR_STOP=1)
"${PSQL[@]}" --command="SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();"
"${PSQL[@]}" --command="DROP DATABASE IF EXISTS \"$DB_NAME\";"
"${PSQL[@]}" --command="CREATE DATABASE \"$DB_NAME\" OWNER \"$DB_OWNER\";"
pg_restore --host="$DB_HOST" --port="$DB_PORT" --username="$DB_USER" --dbname="$DB_NAME" --role="$DB_OWNER" --exit-on-error --single-transaction --no-owner "$STAGE/verified/database.dump"
cp -R "$STAGE/verified/attachments/." "$STORAGE_LOCAL_DIR/"
# Verify every restored object independently of the archive extraction.
"$PYTHON_BIN" - "$STAGE/verified/manifest.json" "$STORAGE_LOCAL_DIR" <<'PY'
import hashlib, json, pathlib, sys
manifest = json.loads(pathlib.Path(sys.argv[1]).read_text())
root = pathlib.Path(sys.argv[2])
for name, expected in manifest['files'].items():
    if not name.startswith('attachments/'):
        continue
    path = root / name.removeprefix('attachments/')
    with path.open('rb') as source:
        if hashlib.file_digest(source, 'sha256').hexdigest() != expected['sha256']:
            raise ValueError('restored attachment checksum mismatch: ' + name)
print('Restored attachment hashes verified')
PY
# Migrations are a separate explicit operation against the restored target; never
# invoke a Compose migration job whose DATABASE_URL could refer to another database.
printf 'Restore completed: %s on %s:%s. Keep writers stopped until migrations, grants and readiness checks pass.\n' "$DB_NAME" "$DB_HOST" "$DB_PORT"
