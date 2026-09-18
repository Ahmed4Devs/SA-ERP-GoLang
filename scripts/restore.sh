#!/usr/bin/env bash
# Restores a PostgreSQL custom-format archive produced by backup.sh.
set -euo pipefail
umask 077

BACKUP_FILE="${1:-}"
FORCE="${2:-}"

if [ -z "${BACKUP_FILE}" ] || [ ! -f "${BACKUP_FILE}" ]; then
    echo "Usage: $0 <backup.dump[.enc]> [--force]" >&2
    exit 1
fi

CHECKSUM_FILE="${BACKUP_FILE}.sha256"
if [ ! -f "${CHECKSUM_FILE}" ]; then
    echo "Error: required checksum file is missing: ${CHECKSUM_FILE}" >&2
    exit 1
fi
sha256sum -c "${CHECKSUM_FILE}"

WORKING_FILE="${BACKUP_FILE}"
TEMP_FILE=""
cleanup() {
    if [ -n "${TEMP_FILE}" ] && [ -f "${TEMP_FILE}" ]; then
        rm -f -- "${TEMP_FILE}"
    fi
}
trap cleanup EXIT

if [[ "${BACKUP_FILE}" == *.enc ]]; then
    if [ -z "${BACKUP_ENCRYPTION_KEY:-}" ]; then
        echo "Error: BACKUP_ENCRYPTION_KEY is required for this encrypted backup." >&2
        exit 1
    fi
    command -v openssl >/dev/null 2>&1 || { echo "Error: openssl is required for decryption." >&2; exit 1; }
    TEMP_FILE="$(mktemp "${TMPDIR:-/tmp}/erp-restore.XXXXXX.dump")"
    export BACKUP_ENCRYPTION_KEY
    openssl enc -d -aes-256-cbc -pbkdf2 \
        -in "${BACKUP_FILE}" \
        -out "${TEMP_FILE}" \
        -pass env:BACKUP_ENCRYPTION_KEY
    WORKING_FILE="${TEMP_FILE}"
fi

if [ "${FORCE}" != "--force" ]; then
    read -rp "WARNING: this recreates the target database. Continue? (y/N): " CONFIRM
    [[ "${CONFIRM}" =~ ^[Yy]$ ]] || { echo "Restoration cancelled."; exit 0; }
fi

DB_HOST="${POSTGRES_HOST:-postgres}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_USER="${POSTGRES_USER:-erp_user}"
DB_NAME="${POSTGRES_DB:-erp_db}"
COMPOSE_FILE="${COMPOSE_FILE:-$(dirname "$0")/../deployments/compose/docker-compose.prod.yml}"
POSTGRES_SERVICE="${POSTGRES_SERVICE:-postgres}"
MIGRATE_SERVICE="${MIGRATE_SERVICE:-erp-migrate}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-}"

if [[ ! "${DB_USER}" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || [[ ! "${DB_NAME}" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
    echo "Error: POSTGRES_USER and POSTGRES_DB must be simple PostgreSQL identifiers." >&2
    exit 1
fi

TERMINATE_SQL="SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB_NAME}' AND pid <> pg_backend_pid();"

if [ -n "${POSTGRES_CONTAINER}" ]; then
    docker exec "${POSTGRES_CONTAINER}" \
        psql --username="${DB_USER}" --dbname=postgres --set=ON_ERROR_STOP=1 --command="${TERMINATE_SQL}"
    docker exec "${POSTGRES_CONTAINER}" \
        psql --username="${DB_USER}" --dbname=postgres --set=ON_ERROR_STOP=1 --command="DROP DATABASE IF EXISTS \"${DB_NAME}\";"
    docker exec "${POSTGRES_CONTAINER}" \
        psql --username="${DB_USER}" --dbname=postgres --set=ON_ERROR_STOP=1 --command="CREATE DATABASE \"${DB_NAME}\" OWNER \"${DB_USER}\";"
    docker exec -i "${POSTGRES_CONTAINER}" \
        pg_restore --username="${DB_USER}" --dbname="${DB_NAME}" --exit-on-error --no-owner < "${WORKING_FILE}"
elif command -v psql >/dev/null 2>&1 && command -v pg_restore >/dev/null 2>&1; then
    export PGPASSWORD="${POSTGRES_PASSWORD:-}"
    psql --host="${DB_HOST}" --port="${DB_PORT}" --username="${DB_USER}" --dbname=postgres \
        --set=ON_ERROR_STOP=1 --command="${TERMINATE_SQL}"
    psql --host="${DB_HOST}" --port="${DB_PORT}" --username="${DB_USER}" --dbname=postgres \
        --set=ON_ERROR_STOP=1 --command="DROP DATABASE IF EXISTS \"${DB_NAME}\";"
    psql --host="${DB_HOST}" --port="${DB_PORT}" --username="${DB_USER}" --dbname=postgres \
        --set=ON_ERROR_STOP=1 --command="CREATE DATABASE \"${DB_NAME}\" OWNER \"${DB_USER}\";"
    pg_restore --host="${DB_HOST}" --port="${DB_PORT}" --username="${DB_USER}" \
        --dbname="${DB_NAME}" --exit-on-error --no-owner "${WORKING_FILE}"
elif command -v docker >/dev/null 2>&1; then
    docker compose -f "${COMPOSE_FILE}" exec -T "${POSTGRES_SERVICE}" \
        psql --username="${DB_USER}" --dbname=postgres --set=ON_ERROR_STOP=1 --command="${TERMINATE_SQL}"
    docker compose -f "${COMPOSE_FILE}" exec -T "${POSTGRES_SERVICE}" \
        psql --username="${DB_USER}" --dbname=postgres --set=ON_ERROR_STOP=1 --command="DROP DATABASE IF EXISTS \"${DB_NAME}\";"
    docker compose -f "${COMPOSE_FILE}" exec -T "${POSTGRES_SERVICE}" \
        psql --username="${DB_USER}" --dbname=postgres --set=ON_ERROR_STOP=1 --command="CREATE DATABASE \"${DB_NAME}\" OWNER \"${DB_USER}\";"
    docker compose -f "${COMPOSE_FILE}" exec -T "${POSTGRES_SERVICE}" \
        pg_restore --username="${DB_USER}" --dbname="${DB_NAME}" --exit-on-error --no-owner < "${WORKING_FILE}"
else
    echo "Error: PostgreSQL client tools or Docker are required." >&2
    exit 1
fi

if [ "${RUN_MIGRATIONS_AFTER_RESTORE:-true}" = "true" ]; then
    command -v docker >/dev/null 2>&1 || {
        echo "Error: Docker is required to run post-restore migrations. Set RUN_MIGRATIONS_AFTER_RESTORE=false only for a controlled drill." >&2
        exit 1
    }
    docker compose -f "${COMPOSE_FILE}" run --rm "${MIGRATE_SERVICE}"
fi

echo "Restore completed successfully: ${DB_NAME}"
