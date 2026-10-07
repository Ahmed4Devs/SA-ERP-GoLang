# Coordinated ERP recovery backups

A recovery bundle contains a PostgreSQL custom-format dump, the document storage
tree, and a versioned manifest with the SHA-256 and size of every file. The bundle
is encrypted with AES-256-CBC and PBKDF2 (200,000 iterations); a SHA-256 sidecar
checks transport integrity. Store the encryption key separately in a secret
manager. SHA-256 is not a signature: protect the artifacts and their checksums
with restricted access and immutable offsite retention.

## Consistency and scheduling

Run `scripts/backup-coordinated.sh` on the Docker host. It takes a host lock,
stops API and worker writers gracefully, invokes the backup container with the
read-only document volume, and restarts only the services that were running,
even if backup fails. Keep all other database/storage writers disabled during
this maintenance window. Standalone `backup.sh` requires an explicit
`BACKUP_QUIESCED=true`; the flag is an operator assertion, not database locking.

```bash
export COMPOSE_ENV_FILE=/etc/erp-go/production.env
bash scripts/backup-coordinated.sh
```

Install `deployments/systemd/erp-backup.service` and `.timer` in
`/etc/systemd/system`, adjust `/opt/erp-go` and the environment-file path, then run:

```bash
systemctl daemon-reload
systemctl enable --now erp-backup.timer
systemctl start erp-backup.service
journalctl -u erp-backup.service
```

The timer runs at 02:00 Asia/Riyadh and catches missed runs. The maintenance
window duration grows with data size; measure it before deployment. Do not use
this procedure for an installation requiring uninterrupted writes without
adding a reviewed snapshot/coordinator implementation.

## Offsite destination and monitoring

The backup image contains restic. Initialize a repository in the approved
region once, set `RESTIC_REPOSITORY`, `RESTIC_PASSWORD` and restricted storage
credentials, and set `BACKUP_REQUIRE_OFFSITE=true`. Each job uploads the encrypted
bundle and checksum, runs `restic check`, and only then prunes the local cache
(default 30 days) and records success. An unavailable required destination fails
the job and does not advance the success timestamp. Keep remote immutable
retention and restic prune credentials separate from routine backup credentials.
Remote pruning is intentionally not run by the backup job.

`backupmetrics` contains `erp_backup.prom` for the optional monitoring Compose
overlay. The supplied alerts detect missing backups, stale backups (>26 hours),
and unavailable metrics. Configure and verify actual incident routing on the
chosen monitoring installation; no messages are automatically sent by setup.

## Recovery limits

Daily logical backups provide an RPO of up to approximately 24 hours. RTO must be
measured on representative data. This deployment does not configure WAL
archiving/PITR. Use a managed PostgreSQL recovery service or reviewed pgBackRest /
WAL-G setup for that requirement. A PITR database restore must be paired with a
compatible attachment recovery point; restoring only the database is incomplete.

A local drill on 2026-10-07 restored a fully migrated isolated test database on
port 55433 plus two document files (including an Arabic text file) from one
encrypted bundle. Both the database marker and attachment hashes were verified.
The rebuilt non-root backup container also completed encryption, restic upload
and repository checking against a separate local drill repository; a second
container restored the database and both documents and verified their hashes.
This proves mechanics, not production-volume RTO or an actual remote endpoint.
Repeat isolated restores monthly and after changes; CI also verifies a database
marker and attachment content using the same backup container.
