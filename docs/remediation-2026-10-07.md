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

## Pending

Transactional posting across treasury, payroll,
assets and inventory; authenticated actors and approval policies; business UI;
canonical accounting/inventory tables and precision; expenses/projects;
compliance adapters; attachment backup and release checks. This document tracks
work in progress and does not assert production readiness.
