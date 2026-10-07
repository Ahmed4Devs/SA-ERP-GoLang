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

## Pending

Allocation validation; transactional posting across AR/AP, treasury, payroll,
assets and inventory; authenticated actors and approval policies; business UI;
canonical accounting/inventory tables and precision; expenses/projects;
compliance adapters; attachment backup and release checks. This document tracks
work in progress and does not assert production readiness.
