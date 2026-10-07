# Restore a database and attachment recovery bundle

Stop incoming writes and all workers. Preserve the existing database and document
volume separately. Provision the required database roles on a new server.
`restore.sh` drops and recreates only the explicit target database, so double-check
the target host, port, and name. The document destination must be empty: the
script never deletes an existing attachment tree.

Use PostgreSQL 18 client tools, Bash, Python 3.12+, OpenSSL and coreutils, or the
built backup image with its `/usr/local/bin/restore.sh` entrypoint. Mount backups
read-only and a fresh writable document destination. Restore must run with a
role allowed to create databases and assign `RESTORE_OWNER`.

```bash
export POSTGRES_HOST=isolated-database-host
export POSTGRES_PORT=5432
export POSTGRES_USER=erp_admin
export POSTGRES_PASSWORD='from-secret-manager'
export POSTGRES_DB=erp_recovered
export RESTORE_OWNER=erp_migrator
export BACKUP_ENCRYPTION_KEY='from-separate-secret-manager'
export STORAGE_LOCAL_DIR=/srv/erp-recovered-documents
export RESTORE_QUIESCED=true
bash scripts/restore.sh /backups/erp_erp_db_TIMESTAMP.tar.gz.enc --force
```

The utility verifies the external checksum, decrypts, rejects unsafe tar members,
checks the full manifest inventory and every file hash, and validates the dump
before touching the database. It restores the database in one `pg_restore`
transaction, copies documents, and verifies their hashes again. Any failure
returns nonzero; keep writers stopped. In a failed recovery, the target database
may be empty or partially prepared; the original bundle remains untouched.

Version-1 bundles replace the old database-only `.dump.enc` format. Legacy
archives must follow the old `pg_restore` procedure with a separately recovered
attachment backup; they cannot meet the coordinated-bundle contract.

Migrations and grants are deliberately separate: set their connection strings
explicitly to the recovered target before running them. A restore script must
not silently invoke a Compose migration job pointing to a different database.
Then check schema version, readiness and document storage access, verify the audit
chain, and reconcile representative ledger, inventory, AR/AP, payroll and
attachment records. Match restored filesystem ownership to the API service user
before reattaching a document volume. Record elapsed time and newest recovered
transaction, and restart writers only after verification.

For PITR, use the deployed database platform's tested WAL/base-backup runbook and
the corresponding attachment recovery point. The standalone Compose stack does
not include PITR.
