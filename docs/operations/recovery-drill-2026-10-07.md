# Isolated recovery verification — 2026-10-07

No production database, attachment volume, or regulatory endpoint was used.
All PostgreSQL targets were on the temporary local cluster at port 55433.

- Source: `erp_compliance_test`, migrated through version 410 at drill time.
- Native Bash restore target: `erp_restore_drill`.
- Container restore target: `erp_restore_container_drill`.
- Backup runtime: rebuilt `erp-backup:remediation20261007`, non-root user.
- Bundle: database custom archive plus two attachments, including UTF-8 Arabic
  text and an attachment named `manifest.json`.
- Encryption: AES-256-CBC / PBKDF2, 200,000 iterations, disposable drill key.
- Native and container restores verified the database marker and all attachment
  hashes. Arabic attachment SHA-256:
  `e65937f9b71ef82a72e4571b6b642e72a53d0d6bf5728c5b75d31576ff0f3376`.
- Restic uploaded the artifact and checksum into a separate local repository;
  `restic check` passed. No actual cloud credentials were supplied.
- Backup metrics were written with last-success timestamp and artifact size.
- Bundle tests reject tampered/missing files, traversal members, and symlinks.
- Production and monitoring Compose overlays passed `config --quiet` using
  placeholder validation values; no production services were started.

The first container restore exposed a bind-mount timestamp permission error in
`cp -a`. Restore now copies document contents without attempting to preserve the
destination mount's ownership/timestamps; the subsequent full container restore
passed. Operators must assign restored directory ownership to the API user.

These are correctness drills on a small test database. They do not establish
production RTO, activate the systemd timer or monitoring stack, verify incident
routing, provision immutable remote retention, or configure PITR.
