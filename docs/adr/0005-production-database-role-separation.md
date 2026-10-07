# ADR 0005: Separate production database identities

## Status
Implemented to satisfy the canonical production-role and least-privilege contract.

## Context
The production Compose deployment used one PostgreSQL installation login for
migrations, API, worker, and backups. The official image creates this login as a
superuser, allowing runtime credentials to disable financial and audit triggers.

## Decision
Provision the canonical owner, migrator, runtime, read-only, and backup roles as
setup jobs. Migration sessions assume the non-login owner role. Apply explicit
runtime grants after migrations, with audit limited to SELECT/INSERT and schema
metadata/reference records limited to SELECT. API and worker use `erp_app`;
logical backup uses the read-only `erp_backup` login. The installation login is
reserved for role provisioning and disaster recovery.

## Alternatives considered
Manual DBA-only setup preserves the same model but is error-prone for new
dedicated Compose installations. Sharing the migrator login with the runtime
does not preserve immutable-record guarantees against a compromised runtime.

## Consequences
Deployments need three additional independent passwords and a migration URL.
Fresh deployments provision automatically. Runtime grants must be refreshed after
manual schema upgrades. Existing installations require a reviewed ownership
transfer; automatic global reassignment of an old administrator's objects is
deliberately avoided.

## Migration impact
No released migration is changed. Role provisioning and grants are operations
scripts. Restore uses `--role=erp_owner` so recovered objects have the intended
ownership, then reapplies runtime grants.

## Security impact
The runtime cannot change schemas, disable triggers, modify migration metadata,
or update/delete audit. Passwords are passed through environment variables to
psql and must be sourced from the deployment's secret manager. Role passwords
and `.env.production` must not be committed.

## Testing impact
CI provisions an isolated PostgreSQL database, applies all migrations as the
migrator, and verifies required reads/writes and denied audit mutation, DDL,
trigger disabling, and migration-metadata changes as `erp_app`.
