#!/usr/bin/env bash
# Run on the Docker host. The application maintenance window covers the whole bundle.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$SCRIPT_DIR/../deployments/compose/docker-compose.prod.yml}"
COMPOSE_ENV_FILE="${COMPOSE_ENV_FILE:?Set the production Compose environment file}"
compose() { docker compose --env-file "$COMPOSE_ENV_FILE" -f "$COMPOSE_FILE" "$@"; }
# Serialize host-scheduled/manual backups and restore maintenance windows.
exec 9>"${BACKUP_LOCK_FILE:-/var/lock/erp-backup.lock}"
flock -n 9 || { echo "Another ERP maintenance job is running" >&2; exit 1; }
RUNNING=()
for service in erp-api erp-worker; do
    if [[ -n "$(compose ps --status running -q "$service")" ]]; then RUNNING+=("$service"); fi
done
resume() {
    local result=$?
    if [[ ${#RUNNING[@]} -gt 0 ]]; then
        compose start "${RUNNING[@]}" || result=1
    fi
    exit "$result"
}
trap resume EXIT
if [[ ${#RUNNING[@]} -gt 0 ]]; then compose stop --timeout 60 "${RUNNING[@]}"; fi
compose --profile operations run --rm -e BACKUP_QUIESCED=true erp-backup
