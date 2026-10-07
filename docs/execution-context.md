# Execution context — 2026-10-07

Canonical requirements: `context/project-prompt.md`. Branch: `master`; initial
commit: `9ed25d3`; the starting working tree was clean. Latest implementation
commit: `355aab2`. This context accompanies the final readiness-document commit;
use `git log -1` for the documentation commit itself.

## Current phase and completion status

Repaired gates in existing phases 3–9, 15, 27 and 31–34. The repository has broad
implementations through phase 34, but no phase-wide or production-complete claim
is justified while the canonical gaps below remain. No release tag or remote
push was created.

## Completed work and commits

- `c8fbf7a`: global IAM administration, strict/versioned commands, atomic security
  audit, last-admin/self-disable protection, branch-scoped grants and immediate
  revocation across replicas. Password reset/suspension/logout-all order against
  challenge/session issuance using a user lock and authentication version. MFA
  activation invalidates earlier verification; concurrent guesses share a
  committed five-attempt budget; logout failures are surfaced. OpenAPI resolves
  split handlers by package and receiver.
- `57accd9`: transaction-bound sales/manual-GL posting and reversal, balance and
  audit rollback, source/period locks, authenticated posting actors, server VAT,
  shared FX allocation, historical credit snapshots and over-credit prevention.
  The global audit chain is locked before document/balance locks. Manual drafts
  preserve identity; posted journal mutation is guarded even under concurrency.
- `2638f61`: real migration/seed CLI, development worker, production database
  identities, diagnostic/restore corrections, OS package updates, isolated
  production smoke and full encrypted restore evidence.
- `355aab2`: structured bilingual master forms, user/role administration,
  document versions/downloads, business invoice workflow, CORS idempotency
  support, mobile/theme/race fixes, dependency patches, removal of invented
  dashboard metrics and unsupported compliance badge.

## Migrations and interfaces

- 330: legal-entity/branch composite assignment constraint and branch scope rule.
- 340: journal dimension ownership foreign keys; depends on 330.
- 350: parent locking for journal line changes/reassignment; reversal metadata
  cannot rewrite original posted financial fields. Current schema constant: 350.
- Invalid legacy scopes deliberately fail migration instead of rewriting history.
- `PostingEngine.PostDraft` preserves an existing draft while using the same
  validation, balances and mandatory audit as source posting. `LockChain` must
  precede business locks in callers composing several financial operations.
- User administration API requires global IAM permissions and optimistic versions.
  Role administration requires global authorization permissions. Branch grants
  do not authorize legal-entity-wide operations.

## Validation on the final code

- Go build/vet/formatting and internal race tests passed.
- All 29 integration packages passed with `-race`; internal-code coverage 50.1%,
  above the 45% CI gate. sqlc/OpenAPI regeneration is reproducible.
- Web build/lint/typecheck, 18 unit tests and 23 mocked browser tests passed.
- Four real browser workflows passed: category creation, user/role/status/password
  management, byte-exact document upload/download, invoice draft/review/post.
- All seven Docker images built. Isolated Compose provision/migrate/grants,
  restricted roles, API/web health, worker startup, HTTPS login/secure cookies
  and encrypted backup/restore passed. 350 down/up passed and readiness correctly
  rejected schema 340 before recovering at 350. Patched images were smoke-tested.
- Trivy source vulnerability/secret/misconfiguration scan and all seven images
  passed HIGH/CRITICAL with the same `--ignore-unfixed` policy as CI. SPDX 2.3
  SBOMs are in ignored `.tmp/sbom`. Go reachable scan and npm runtime audit clean.
- Full npm audit remains blocked by seven HIGH development advisories rooted in
  unpatched braces GHSA-vfj7-8cjw-p6xm. CI was not weakened. No remote CI result
  is claimed.

## Environment and cleanup

Only synthetic data was used. The native PostgreSQL cluster was loopback port
55432 in ignored `.tmp/postgres-audit`; its API and database were stopped after
verification. The isolated Compose project's containers and three explicitly
prefixed test volumes were removed after checking resolved volume names. Source
images, ignored logs, test screenshots and SBOMs remain for review. Existing
application data and system PostgreSQL on 5432 were not modified.

## Architectural and regulatory decisions

Preserve modular Go/PostgreSQL/Next architecture and ADRs 0001–0004, including the
transactional PostgreSQL outbox. ADR 0005 documents owner/migrator/app/readonly/
backup roles. Never rewrite released migrations or posted financial records.
Existing NUMERIC precision is preserved pending a deliberate forward migration.
No real invoices, email or regulatory data were sent. ZATCA production onboarding
continues to fail closed without verified conformance material and credentials.

## Outstanding implementation and release gates

- Eleven mandatory accounting/inventory table names remain absent; versioned
  posting configuration, general dimensions, exchange rates, opening balances,
  reconciliation, item costs and valuation snapshots need implementation.
- Sales still lacks the full approval/reversal lifecycle and configurable
  maker/checker policy; sales account mapping remains fixed by code. Other
  modules still have direct posting SQL and need centralization/review.
- Canonical six-decimal money/quantity and twelve-decimal exchange-rate storage
  require a reviewed forward migration across existing four-decimal modules.
- Dedicated expenses/projects modules and remaining business UI workflows are
  absent/incomplete; several screens still use generic JSON actions and the
  prescribed frontend component stack is not fully adopted.
- The unpatched development dependency chain prevents the full npm security gate.
- Actual deployment needs customer configuration, credentials, TLS/mail/backup
  scheduling and official regulatory acceptance.

See `readiness-review-2026-10-07.md`, `financial-posting-review-2026-10-06.md`,
`operations/user-administration.md`, `operations/sales-invoice-workflow.md`,
and `operations/production-smoke-2026-10-07.md` for evidence and practical limits.
