# Remediation execution — 7 October 2026

User-authorized sequence: financial reporting, cross-module posting, allocations,
treasury/approvals (items 1–4), then business UI, canonical schema, compliance,
and operations/security (items 5–8). Changes are committed by related domain
and pushed after verification. Starting revision: `628815e`, branch `master`.

## Reporting and reversal

Financial reports, dashboard totals and budget actuals include both `POSTED`
and `REVERSED` original journals. Reversal is a compensating entry; changing
the original's status must not erase its historical contribution. Migration
360 corrects the reporting views and excludes draft lines from trial balance.
Released migrations and original journal amounts are unchanged.

Verified against isolated PostgreSQL 18 with the race detector: accounting,
reporting, dashboard and budgeting integration packages. The new regression
compares present and historical trial balance, balance sheet, income statement,
general ledger and SQL reporting views before and after reversal, including an
unposted draft that must not contribute.

## Allocation integrity

Receipt and supplier-payment allocations reject zero/negative amounts,
duplicate invoice IDs, and mismatched legal entities, counterparties or
currencies. Receipt header totals use exactly the validated set of allocations.
Invoice rows are locked in a stable order. Integration regressions verify
rollback and two competing receipts cannot over-settle the same available
amount. AR and AP integration suites passed with the race detector.

## AR/AP central posting and purchase tax

Receipts, supplier invoices and supplier payments use the common posting
engine in the source transaction, with the authenticated actor, the document
date's accounting period, account balances and mandatory audit. Migration 370
adds input-tax code snapshots to new supplier invoice lines; posting records
INPUT tax using the saved effective rate and matching GL currency conversion.
Posted history is preserved. Legacy drafts without a snapshot must be recreated.

AR/AP integration suites pass with the race detector. New regressions cover
foreign invoice conversion and input VAT equality, duplicate posting, closed
accounting/tax periods, audit-write failure, and receipt allocation rollback.
Targeted vet passes. Receipt/payment FX snapshots and complete approval policies
remain follow-up work; foreign-currency settlements currently fail safely.

## Treasury actors, approvals and posting

Treasury receipts, approved payments and transfers create central GL entries
with scoped account mappings, source references and atomic audit events.
HTTP commands derive actors from the authenticated session. Creators cannot
approve/post their own documents; payment posting requires approval. Transfers
include fees in the balanced journal. Existing retry behavior is preserved.

Treasury PostgreSQL integration tests passed with the race detector, including
HTTP actor forgery, maker/checker rejection, duplicate requests, closed-period
rollback, and an injected audit failure. Targeted vet passed. Configurable
approval limits and FX snapshots remain follow-up work.

## Payroll central posting and approvals

Payroll uses the central ledger engine and date-selected fiscal period.
Authenticated creators, approvers and posting actors are required. Approval
records its actor/time and audit transactionally (migration 380); approved
calculations cannot be overwritten, and creators cannot approve/post their run.
Posting verifies employee totals, creates the journal, balances, audit and
payslips within one transaction. Existing posted runs are unchanged.

Payroll integration tests pass with the race detector, including approval
bypass attempts, creator rejection, persisted approver, closed-period rollback,
balanced ledger amounts and mandatory posting audit. Targeted vet passes.

## Fixed asset posting

Depreciation and disposal now create central GL entries atomically with asset
carrying values. Depreciation requires an independent actor, locks the run and
assets, verifies calculation snapshots and posts on the month's final date.
Calculating the same period cannot overwrite an existing run. Assets entering
service after the period are excluded. Financial posting uses the primary book;
separate tax-book carrying balances are not yet implemented.

Disposal requires explicit proceeds, gain/loss and (when applicable) impairment
account selections. Account selections are retained on the immutable journal.
Disposing an already disposed asset is rejected. HTTP actors come from sessions.
Fixed-assets race integration tests and vet pass, including closed-period
rollback, independent actor checks and duplicate-calculation history protection.

## Inventory adjustments, counts and valuation

Migration 390 snapshots inventory/variance account selections and journal links
on adjustments/counts, plus count creators and optional found-stock unit costs.
Adjustments and counts use independent authenticated posting actors, actual
movement costs, central journals and atomic source audits. Counts reject stale
stock snapshots. Posting validates item/warehouse/bin entity consistency.

Weighted-average revaluation now totals all bins and updates each bin using its
own quantity; previously each bin could receive the same warehouse total.
Its variance posts to GL atomically. FIFO/standard revaluation requires its own
cost-layer treatment and is rejected explicitly. Inventory race integration
tests and vet pass, including multiple bins, closed-period rollback, count
maker-checker and duplicate reconciliation. Internal/cmd tests pass on rerun;
the first aggregate run reported a failure without retained diagnostic output.

## Canonical finance and invoice approvals

Migrations 400/410 add the eleven missing canonical tables and widen financial
storage without recomputing posted history. Audited APIs support dimensions,
exchange rates, effective posting mappings, opening balances, reconciliations
and inventory valuation snapshots. Source mappings connect to AP, AR and payroll.
Journal custom dimensions survive reversal. Inventory costs aggregate all bins.

Migration 420 adds optimistic, audited approval limits for sales, treasury and
payroll. Workflow thresholds compare exact decimals. Migration 430 binds sales
approval to server snapshots, preserves independent creators/approvers/posters,
and freezes submitted/posted invoice amounts and lines. Employee GOSI deductions
now use the configured insurance liability (seeded 215000), not output VAT.

After sqlc regeneration, internal/cmd tests and all 29 PostgreSQL integration
packages passed with the race detector. See `.tmp/remediation-integrated-current.log`.
The schema upgrade rehearsal preserved historical journals and report grants.

## Compliance and recovery

GOSI no longer fabricates a portal confirmation or salary basis. File/manual,
simulation and unavailable official behavior are distinct. ZATCA live paths
reject simulator credentials, expired certificates and failed HTTP responses.
Encrypted recovery bundles contain DB plus attachments, with verified paths,
inventory and checksums. Native and non-root container restore drills passed.
Scheduling, optional restic offsite delivery and monitoring templates are present.

## Work still in progress

Expense/project modules, remaining UI workflows and FX settlement snapshots with
realized gains/losses. Actual regulatory acceptance, offsite credentials, deployed
timers/alerts and PITR need customer infrastructure and external validation.
This log does not assert production readiness.
