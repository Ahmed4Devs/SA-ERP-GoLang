# ERP-GO Disaster-Recovery Restore Guide

## Scope

This guide restores the encrypted PostgreSQL custom-format archive produced by
`scripts/backup.sh`. The checked-in deployment does not implement WAL archiving
or point-in-time recovery. See the backup strategy before assigning RPO/RTO
commitments.

## Before restoring

1. Isolate incoming traffic and stop writers:

   ```bash
   docker compose --env-file deployments/compose/.env.production \
     -f deployments/compose/docker-compose.prod.yml stop erp-api erp-worker
   ```

2. Preserve a forensic snapshot of the damaged state when safe.
3. Confirm the target host, database name, owner, backup path, and encryption
   key with the incident commander. Restoration drops and recreates the target
   database.
4. Verify that both the artifact and its adjacent `.sha256` file are present.

## Restore an encrypted logical backup

Run from the repository root with the production database variables loaded from
the secret manager:

```bash
export BACKUP_ENCRYPTION_KEY='read-from-your-secret-manager'
export POSTGRES_HOST='database-host'
export POSTGRES_PORT='5432'
export POSTGRES_USER='erp_user'
export POSTGRES_PASSWORD='read-from-your-secret-manager'
export POSTGRES_DB='erp_db'
export COMPOSE_ENV_FILE='deployments/compose/.env.production'

./scripts/restore.sh /backups/erp_erp_db_20260918_020000.dump.enc --force
```

The utility verifies SHA-256 before decryption, recreates the target database,
runs `pg_restore --exit-on-error --no-owner`, and then runs the migration
service. Unencrypted legacy archives remain readable, but new production
backups are rejected unless encryption is configured.

## Post-restore verification

1. Confirm the schema version:

   ```bash
   docker compose --env-file deployments/compose/.env.production \
     -f deployments/compose/docker-compose.prod.yml run --rm erp-migrate status
   ```

2. Run the system doctor and the release smoke tests.
3. Verify audit-chain integrity through `GET /api/v1/audit/verify`.
4. Reconcile representative ledger, inventory, receivable, payable, payroll,
   document, and outbox counts against the incident record.
5. Record elapsed restore time and the newest recovered transaction timestamp.
6. Restart writers only after business and technical owners sign off:

   ```bash
   docker compose --env-file deployments/compose/.env.production \
     -f deployments/compose/docker-compose.prod.yml start erp-api erp-worker
   ```

## Point-in-time recovery

PITR is unavailable in the repository's standalone Compose baseline. If the
production database platform provides WAL archiving, follow that platform's
reviewed runbook and test it separately. Creating `recovery.signal` without a
verified base backup, WAL archive, restore command, and timeline plan is not a
valid recovery procedure.
