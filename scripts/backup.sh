#!/usr/bin/env bash
# PostgreSQL logical backup utility. Produces a pg_dump custom-format archive.
set -euo pipefail
umask 077

TIMESTAMP="$(date +"%Y%m%d_%H%M%S")"
BACKUP_DIR="${BACKUP_DIR:-/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DB_HOST="${POSTGRES_HOST:-postgres}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-erp_user}"
DB_NAME="${POSTGRES_DB:-erp_db}"
COMPOSE_FILE="${COMPOSE_FILE:-$(dirname "$0")/../deployments/compose/docker-compose.prod.yml}"
COMPOSE_ENV_FILE="${COMPOSE_ENV_FILE:-}"
POSTGRES_SERVICE="${POSTGRES_SERVICE:-postgres}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-}"
BACKUP_FILE="${BACKUP_DIR}/erp_${DB_NAME}_${TIMESTAMP}.dump"

mkdir -p "${BACKUP_DIR}"

if [ -z "${BACKUP_ENCRYPTION_KEY:-}" ] && [ "${ALLOW_UNENCRYPTED_BACKUP:-false}" != "true" ]; then
    echo "Error: BACKUP_ENCRYPTION_KEY is required. Set ALLOW_UNENCRYPTED_BACKUP=true only for an explicitly approved non-production drill." >&2
    exit 1
fi

echo "Starting PostgreSQL backup: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"

compose() {
    if [ -n "${COMPOSE_ENV_FILE}" ]; then
        docker compose --env-file "${COMPOSE_ENV_FILE}" -f "${COMPOSE_FILE}" "$@"
    else
        docker compose -f "${COMPOSE_FILE}" "$@"
    fi
}

if [ -n "${POSTGRES_CONTAINER}" ]; then
    docker exec "${POSTGRES_CONTAINER}" \
        pg_dump --username="${DB_USER}" --dbname="${DB_NAME}" --format=custom --compress=9 > "${BACKUP_FILE}"
elif command -v pg_dump >/dev/null 2>&1; then
    export PGPASSWORD="${POSTGRES_PASSWORD:-}"
    pg_dump \
        --host="${DB_HOST}" \
        --port="${DB_PORT}" \
        --username="${DB_USER}" \
        --dbname="${DB_NAME}" \
        --format=custom \
        --compress=9 \
        --file="${BACKUP_FILE}"
elif command -v docker >/dev/null 2>&1; then
    compose exec -T "${POSTGRES_SERVICE}" \
        pg_dump --username="${DB_USER}" --dbname="${DB_NAME}" --format=custom --compress=9 > "${BACKUP_FILE}"
else
    echo "Error: neither pg_dump nor Docker is available." >&2
    exit 1
fi

FINAL_FILE="${BACKUP_FILE}"
if [ -n "${BACKUP_ENCRYPTION_KEY:-}" ]; then
    command -v openssl >/dev/null 2>&1 || { echo "Error: openssl is required for encryption." >&2; exit 1; }
    export BACKUP_ENCRYPTION_KEY
    FINAL_FILE="${BACKUP_FILE}.enc"
    openssl enc -aes-256-cbc -salt -pbkdf2 \
        -in "${BACKUP_FILE}" \
        -out "${FINAL_FILE}" \
        -pass env:BACKUP_ENCRYPTION_KEY
    rm -f -- "${BACKUP_FILE}"
fi

FINAL_DIR="$(dirname "${FINAL_FILE}")"
FINAL_NAME="$(basename "${FINAL_FILE}")"
(
    cd "${FINAL_DIR}"
    sha256sum "${FINAL_NAME}" > "${FINAL_NAME}.sha256"
    sha256sum -c "${FINAL_NAME}.sha256"
)

find "${BACKUP_DIR}" -type f -name "erp_${DB_NAME}_*.dump*" -mtime "+${RETENTION_DAYS}" -delete

echo "Backup completed: ${FINAL_FILE}"
