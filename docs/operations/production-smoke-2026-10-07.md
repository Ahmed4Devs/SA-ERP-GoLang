# Isolated production smoke — 7 October 2026

The production Compose topology was exercised with synthetic credentials and a
one-day self-signed localhost certificate. The project was
`erp-smoke-3ea5b3bc23`; **every named volume** was overridden with that project
prefix. No existing application volume was mounted. HTTP and HTTPS used ports
18080 and 18443. This is local verification, not a deployment or external
compliance approval.

## Results

- PostgreSQL 18 started healthy. Provision, migration and grants jobs exited 0.
  All migrations through version 340 applied to a fresh database.
- The API ran with `erp_app`, the migration job with `erp_migrator` assuming
  `erp_owner`, and the backup with `erp_backup`.
- `tests/operations/database_roles.sql` passed. Runtime DDL, audit deletion,
  disabling financial triggers, and migration metadata writes were denied.
- API readiness and the web login returned HTTP 200 through Nginx HTTPS. The
  worker initialized both outbox and notification-delivery jobs without errors.
- Bootstrap login and authenticated `/api/v1/auth/me` passed. The session cookie
  had Secure, HttpOnly and SameSite attributes. The new administrator had the
  global SUPER_ADMIN assignment.
- Chromium completed the production login and loaded the dashboard without page
  errors. All browser API requests used the same HTTPS origin.
- The operations-profile backup produced an encrypted archive and a matching
  SHA-256 checksum, without retaining a plaintext dump. Decryption and
  `pg_restore --list` succeeded; the archive contained 204 table-data entries.
- `scripts/restore.sh` restored that archive to the separate
  `erp_smoke_restore` database. It retained owner `erp_owner`, schema version
  340, four reference currencies, and the bootstrap user. The same negative
  runtime privilege checks passed against the restored database.

## Final source verification

The API, worker, migrator, and web images were rebuilt from the final source
using local tags `review20261007`. Migration 350 applied to the existing isolated
340 database and grants were refreshed. Recreated API/web containers became
healthy; the worker stayed running. Nginx HTTPS readiness, web login, API login,
secure cookie attributes, and authenticated identity all passed again.

A migration-350 down/up exercise succeeded. `/readyz` returned HTTP 503 while
schema 340 was behind the new binary, then returned HTTP 200 after migration
350 was reapplied. This checks both rollback syntax and schema compatibility
reporting. The earlier full restore exercise above deliberately records its
original version 340; it was not relabeled as a 350 restore.

The final dashboard removes hard-coded growth percentages, decorative KPI
progress values, a fixed “Healthy Margin” assessment, and an unsupported ECDSA
Phase 2 badge. It shows the actual server metrics without invented assessments.

The database deliberately contained no demo business data. Organization,
accounting, tax, and business setup remain necessary for a real installation.
ZATCA was disabled and SMTP used an inert local test address. No external
invoice submission or email delivery was attempted.

Temporary configuration, browser evidence and screenshots are under the ignored
`.tmp/production-smoke-20261007` directory. That configuration contains only
synthetic smoke credentials and must not be published or reused in production.

## Image security and bills of materials

Trivy 0.74.0 found fixable HIGH findings in OpenSSL in the original API, worker,
and migrator images, and libexpat/pcre2 in the cached Nginx image. The Go runtime
Dockerfiles now refresh Alpine packages before installation, and Nginx's existing
upgrade layer was rebuilt without stale cache. All seven final images passed the
CI-equivalent HIGH/CRITICAL scan with `--ignore-unfixed`. Patched-runtime HTTPS
readiness, login, and authenticated session checks passed again.

A clean export of the 753 tracked/new source files also passed Trivy vulnerability,
secret, and misconfiguration checks under the same severity/fix policy. This
excludes ignored local data, test databases, binaries, and environment files that
are not part of the source release. SPDX 2.3 SBOMs for all seven images were
created in ignored `.tmp/sbom`. These local checks do not assert remote CI passed;
the full npm audit remains blocked by the unpatched braces development chain.
