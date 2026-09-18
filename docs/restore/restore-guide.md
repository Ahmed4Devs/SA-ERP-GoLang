# Saudi Enterprise ERP - Disaster Recovery & Restore Guide

## 1. Overview & Recovery Objectives
This document provides technical instructions for restoring the Saudi Enterprise ERP relational database and application state during a disaster recovery (DR) event, hardware failure, or data corruption incident.

### Target SLA
- **RTO (Recovery Time Objective)**: `< 30 minutes` for complete logical restore.
- **RPO (Recovery Point Objective)**: `< 5 minutes` with continuous WAL stream archiving.

---

## 2. Pre-Restoration Checklist
Before initiating any restore procedure:
1. **Quarantine Application Traffic**: Stop or isolate incoming traffic at Nginx to prevent concurrent writes:
   ```bash
   docker compose -f deployments/compose/docker-compose.prod.yml stop erp-api erp-worker
   ```
2. **Verify Backup Integrity**: Validate the SHA-256 checksum against the `.sha256` manifest:
   ```bash
   sha256sum -c erp_erp_db_20260918_020000.dump.sha256
   ```
3. **Backup Current Broken State**: Take a snapshot of the damaged database state before overwriting, in case forensic analysis is required.

---

## 3. Disaster Recovery Procedures

### Scenario A: Logical Restore from Dump Archive
Use the tested restore utility for an unencrypted or encrypted custom-format dump:

```bash
./scripts/restore.sh /backups/erp_erp_db_20260918_020000.dump --force
```

The utility verifies the required checksum, recreates the target database,
restores through `pg_restore --exit-on-error --no-owner`, then runs migrations.

### Scenario B: Decrypting an Encrypted Backup
If the backup was encrypted with AES-256-CBC:
```bash
export BACKUP_ENCRYPTION_KEY='read-from-your-secret-manager'
./scripts/restore.sh /backups/erp_erp_db_20260918_020000.dump.enc --force
```

### Scenario C: Point-In-Time Recovery (PITR)
For rolling back immediately prior to a catastrophic incident (e.g. accidental bulk deletion at 14:32:10):
1. Stop PostgreSQL:
   ```bash
   docker compose -f deployments/compose/docker-compose.prod.yml stop postgres
   ```
2. Place a `recovery.signal` file in the data directory.
3. Configure `recovery_target_time = '2026-09-18 14:32:00+03'` in `postgresql.conf`.
4. Start PostgreSQL and monitor recovery logs:
   ```bash
   docker compose -f deployments/compose/docker-compose.prod.yml logs -f postgres
   ```

---

## 4. Post-Restore Integrity Verification
Once the database is restored:
1. **Verify Schema Version**:
   ```bash
   docker compose -f deployments/compose/docker-compose.prod.yml run --rm erp-migrate goose -dir /database/migrations postgres "$DATABASE_URL" status
   ```
2. **Reconcile General Ledger Balances**:
   Run database integrity checks to confirm debits match credits across all journal entries:
   ```sql
   SELECT sum(debit_amount) - sum(credit_amount) AS ledger_imbalance FROM journal_entry_lines;
   ```
   *Expected output: `0.0000`*.
3. **Run System Doctor**:
   ```bash
   ./scripts/doctor.sh
   ```
4. **Restart Application Services**:
   ```bash
   docker compose -f deployments/compose/docker-compose.prod.yml start erp-api erp-worker
   ```
