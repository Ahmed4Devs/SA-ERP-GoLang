# Financial posting review — 2026-10-06

The previous audit remediation did not prove that the central posting engine used a single database transaction. The engine's journal creation, posting, and balance updates could commit separately. Sales also bypassed that engine, selected an arbitrary open fiscal period, used an exchange rate of one, accepted browser-supplied VAT rates, and omitted the tax transaction ledger.

## Changes

- The posting engine now wraps posting and reversal in a transaction. Accounting and tax repositories honor the transaction carried by the caller, including reads, child inserts, posting, and balance projections.
- Manual GL posting and reversal now use that same engine. `PostDraft` posts the saved journal in place, preserving its ID, number, creator and lines; retries cannot duplicate balances or financial audit. Saved drafts are revalidated for active same-entity journals/accounts, their date-matching open period, dimension ownership, positive exchange rates, permitted status and line/header totals. Account, journal and period shared locks retain these checks through commit. Balance projections use the persisted line precision. Reversals preserve each original line's currency and exchange rate.
- A transaction advisory lock serializes attempts for the same source document. A shared fiscal-period row lock prevents closing the selected period before posting commits. Reversals lock the original journal.
- Posting and reversal append their financial audit event inside that same transaction. An audit write failure rolls back the journal, balance projections, and audit event. Duplicate attempts produce no additional financial audit events. Reversal events preserve the original entry, actor, and reason; the global hash-chain lock is retained through the business commit.
- The engine and sales posting commands acquire the audit-chain lock before their document and balance row locks. This supports multiple engine posts in one outer transaction without reversing the audit/balance lock order. The global chain serializes financial posting transactions; capacity testing must account for this. Future callers that acquire business locks before invoking these services must explicitly acquire `AuditService.LockChain` first, and propagate errors to roll back the outer transaction.
- Journal branches, departments, and cost centers are checked against the journal legal entity. Migration `000340_accounting_dimension_scope.sql` adds composite foreign keys so direct SQL and subsequent attempts to move a referenced dimension across entities are also rejected.
- Sales invoices and credit notes use the posting engine in the same transaction as AR, VAT, order quantities, and document status. The accounting date determines the period; foreign-currency amounts use the document exchange rate.
- Quotations, orders, and invoices resolve active output tax codes and effective-dated rates from the legal entity's tax configuration. Client rates cannot override them. Credit notes preserve the original invoice's currency, exchange rate, and tax rate. Unknown or inactive tax codes fail; legacy unconfigured strings such as `VAT-15` must be replaced by the configured code, such as the seeded `VAT-STD-15`.
- Posting copies the saved tax amounts into the VAT ledger in base currency. Credit notes produce negative output VAT amounts. An unavailable or closed tax period rolls back the whole posting.
- GL and VAT use the same FX allocation: each line receives the difference between consecutive rounded cumulative totals. This preserves the converted document totals while making line totals reconcile exactly; a zero-tax line receives no allocated tax. For two lines with tax `0.0015` each and FX `3.75`, the base-tax allocation is `0.0056` and `0.0057`, totaling `0.0113` in both ledgers. Credit notes negate the same allocation.
- Sales draft header/line creation is transactional. HTTP creation records the authenticated user instead of selecting an arbitrary database user. Credit notes require the original customer and entity and cannot cumulatively exceed the original invoice amount.
- Invoice and credit-note posting commands require the authenticated posting actor explicitly. Journal `posted_by`, financial audit, and VAT attribution use that actor while the sales document retains its original creator.

## Validation

Passed against a separate, disposable PostgreSQL 18.6 database:

```text
go test -race -p 1 -count=1 -timeout 10m ./tests/integration/accounting ./tests/integration/sales ./tests/integration/tax ./tests/integration/audit
go vet ./internal/accounting/... ./internal/sales/... ./internal/tax/... ./internal/audit/...
```

New integration tests cover a balance-write failure after journal posting, reversal rollback, eight concurrent attempts for the same source, malicious VAT input, USD/SAR conversion, date-specific fiscal periods, AR balance projection updates, positive invoice/negative credit-note VAT, excessive credit notes, closed accounting and tax periods, and a failed second draft line leaving no header or first line behind. Further regressions cover a failure after audit insertion, audit idempotency, reversal reason preservation, hash-chain verification after rollback, cross-entity dimensions, direct SQL bypass, and dimension ownership changes after posting.

Additional regressions verify that a different user can post without changing the creator, fractional multi-line invoice/credit-note VAT reconciles with GL, HTTP posting derives its actor from authentication, and two posts in one transaction can finish while another transaction waits to post against the same balance rows.

Manual GL regressions cover eight simultaneous posting attempts and eight reversal attempts, preserving the draft identity and creator, using the actual posting actor, exactly one balance/audit projection, rollback after balance and audit failures, account/journal deactivation after drafting, closed or mismatched periods, foreign-entity journal/period IDs, foreign dimensions, invalid status/rates, inconsistent header totals and missing actors. A barrier test pauses posting after validation and proves other connections cannot deactivate its account/journal or close its period before posting commits.

## Remaining specification gaps

This correction does not establish full completion of the canonical specification. A comparison of its 206 explicitly named tables/views with the migrated database found these 11 missing names:

```text
accounting.dimensions
accounting.dimension_values
accounting.journal_line_dimensions
accounting.exchange_rates
accounting.opening_balances
accounting.posting_rules
accounting.posting_rule_versions
accounting.reconciliations
accounting.reconciliation_lines
inventory.item_costs
inventory.valuation_snapshots
```

The current sales chart mapping still uses account codes `112000`, `410000`, and `213000`; a configurable, versioned posting-rule subsystem is absent. The sales API still lacks the complete submit/approve/reverse lifecycle and domain enforcement of configurable maker/checker policies. Other business modules still contain independent posting SQL and need separate review before claiming all financial operations pass through the posting engine. Existing financial database columns commonly use `NUMERIC(18,4)`, while the canonical contract specifies `NUMERIC(24,6)`; this work preserves the installed precision rather than silently rewriting released migrations. Dedicated expenses and projects business modules are also absent. The ZATCA production-onboarding restriction and operational launch gates in the audit-remediation document remain applicable.

Migration 340 depends on the branch uniqueness constraint introduced by migration 330. It validates existing journal dimensions and deliberately fails if legacy rows cross entity boundaries; an accountant must review such data before upgrading. It does not rewrite posted entries. The transactional posting/audit correction itself does not require data migration. No production compliance approval or external submission was performed. HTTP request auditing remains a separate access/business-request record; the financial event described here is atomic only for commands that use the shared posting engine.

Migration 350 strengthens database immutability: line INSERT/UPDATE/DELETE locks
the parent journal through commit, checks both parents on reassignment, and
rejects changes to original financial fields during the POSTED-to-REVERSED
transition. Its concurrency test holds an uncommitted posting while a second
connection edits a line, then verifies the edit fails after posting commits.
It also rejects moving a draft line into a posted journal and rewriting posted
amounts while supplying reversal metadata.
