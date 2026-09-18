#!/usr/bin/env bash
# ==============================================================================
# Saudi Enterprise ERP - Production Database Restore Utility
# Restores logical dumps into PostgreSQL with SHA-256 verification
# ==============================================================================
set -euo pipefail

BACKUP_FILE="${1:-}"
FORCE="${2:-}"

if [ -z "${BACKUP_FILE}" ]; then
    echo "Usage: $0 <path_to_backup_file> [--force]"
    echo "Example: $0 /backups/erp_erp_db_20260918_020000.sql.gz --force"
    exit 1
fi

if [ ! -f "${BACKUP_FILE}" ]; then
    echo "Error: Backup file does not exist: ${BACKUP_FILE}"
    exit 1
fi

echo "===================================================================="
echo " Starting ERP Production Database Restoration"
echo " Target File: ${BACKUP_FILE}"
echo "===================================================================="

# Check Checksum if available
CHECKSUM_FILE="${BACKUP_FILE}.sha256"
if [ -f "${CHECKSUM_FILE}" ]; then
    echo " Verifying SHA-256 checksum..."
    sha256sum -c "${CHECKSUM_FILE}"
    echo " Checksum verified successfully."
else
    echo " Warning: No .sha256 file found for checksum verification."
fi

# Decrypt if encrypted
WORKING_FILE="${BACKUP_FILE}"
if [[ "${BACKUP_FILE}" == *.enc ]]; then
    if [ -z "${BACKUP_ENCRYPTION_KEY:-}" ]; then
        echo "Error: Encrypted archive detected, but BACKUP_ENCRYPTION_KEY is not set."
        exit 1
    fi
    DECRYPTED_FILE="/tmp/$(basename "${BACKUP_FILE}" .enc)"
    echo " Decrypting archive to ${DECRYPTED_FILE}..."
    openssl enc -d -aes-256-cbc -pbkdf2 \
        -in "${BACKUP_FILE}" \
        -out "${DECRYPTED_FILE}" \
        -pass pass:"${BACKUP_ENCRYPTION_KEY}"
    WORKING_FILE="${DECRYPTED_FILE}"
fi

# Safety Confirmation
if [ "${FORCE}" != "--force" ]; then
    read -rp "WARNING: This will overwrite the target database. Are you sure? (y/N): " CONFIRM
    if [[ ! "${CONFIRM}" =~ ^[Yy]$ ]]; then
        echo "Restoration cancelled by operator."
        exit 0
    fi
fi

DB_USER="${POSTGRES_USER:-erp_user}"
DB_NAME="${POSTGRES_DB:-erp_db}"
COMPOSE_FILE="$(dirname "$0")/../deployments/compose/docker-compose.prod.yml"

echo " Terminating open connections and rebuilding database ${DB_NAME}..."
docker compose -f "${COMPOSE_FILE}" exec -T postgres psql -U "${DB_USER}" -d postgres << EOF
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB_NAME}' AND pid <> pg_backend_pid();
DROP DATABASE IF EXISTS ${DB_NAME};
CREATE DATABASE ${DB_NAME} OWNER ${DB_USER};
EOF

echo " Streaming backup into database..."
gunzip -c "${WORKING_FILE}" | docker compose -f "${COMPOSE_FILE}" exec -T postgres psql -U "${DB_USER}" -d "${DB_NAME}"

echo " Running schema migrations..."
docker compose -f "${COMPOSE_FILE}" up erp-migrate

# Cleanup decrypted temp file
if [[ "${WORKING_FILE}" == /tmp/* ]]; then
    rm -f "${WORKING_FILE}"
fi

echo "===================================================================="
echo " Restoration completed successfully at $(date +"%Y-%m-%d %H:%M:%S AST")"
echo "===================================================================="
