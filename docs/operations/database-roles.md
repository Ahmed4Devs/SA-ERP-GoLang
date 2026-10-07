# Production database identities

A dedicated customer deployment uses five database roles:

| Role | Use |
| --- | --- |
| `erp_owner` | Owns database objects; cannot log in. |
| `erp_migrator` | Runs schema changes as `erp_owner`. |
| `erp_app` | API and worker DML; no schema changes or audit mutation. |
| `erp_readonly` | Read-only group; cannot log in. |
| `erp_backup` | Read-only login for logical backups. |

The installation administrator (`POSTGRES_USER`, normally `erp_admin`) is
separate and is not used by the API, worker, or backup service. Use independent
passwords in `.env.production`, and make `DATABASE_URL` use `erp_app` and
`MIGRATION_DATABASE_URL` use `erp_migrator`. URL-encode password characters.

Fresh Compose installations run `erp-provision`, then `erp-migrate`, then
`erp-grants` before the API and worker. The provisioning SQL reads passwords
through psql environment variables, configures the migrator's default role, and
rejects runtime membership in the schema-owner role. Grants restrict audit to
SELECT/INSERT and reference tables/schema-version metadata to SELECT.

After every manual migration run, refresh privileges for new tables:

```sh
docker compose --env-file deployments/compose/.env.production \
  -f deployments/compose/docker-compose.prod.yml run --rm erp-grants
```

## Existing installations

Take and verify a backup first. Existing tables created by an older superuser
deployment keep their original owner; changing a database owner does not change
table ownership. In a maintenance window, use the installation administrator to
transfer application schemas, tables, sequences, views, and routines to
`erp_owner`, including the Goose metadata table. Review the exact object list
before changing ownership; do not use a cluster-wide role reassignment blindly.
Then run the migration and grant services, switch API/worker credentials, and
test login, posting, document storage, and backup/restore on a staging clone.

This automatic provisioning is for the dedicated PostgreSQL instance in Compose.
For an externally managed server, the DBA must apply the same reviewed role and
grant scripts; service host names and credentials must be adjusted consistently.

## Restore ownership

Restoration needs the installation administrator to recreate the target database.
The backup login deliberately cannot do this. After recreation, restore using
`pg_restore --role=erp_owner --no-owner` so restored objects retain the intended
owner, then run `erp-grants`. For the repository restore utility set
`RESTORE_OWNER=erp_owner`; see the restore guide.
