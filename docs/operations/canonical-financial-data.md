# Canonical financial data and precision

Migrations 400 and 410 complete the accounting and inventory tables specified in sections 39–40 of `context/project-prompt.md` and widen the stored decimal contract. Apply them with the normal migration role during a maintenance window; changing the numeric columns takes table locks. They are forward-only. Reverting an installation requires the verified database and attachment restore procedure, not a precision-narrowing migration.

Migration 410 preserves the values and metadata of existing journals, invoices, tax snapshots and stock movements. Money and quantities use `NUMERIC(24,6)`, exchange rates and unit conversion ratios use `NUMERIC(24,12)`, and percentages use `NUMERIC(12,8)`. Existing calculations of payroll statutory amounts retain their existing rounding policy. General posting, purchase/sales calculations, inventory costing and reporting now retain six decimal places. The posting engine rejects amounts with more than six decimal places and exchange rates with more than twelve, preventing database rounding from changing a balanced request.

The seven standard reporting views are recreated, retaining their owners and explicit grants. Custom views depending on affected objects stop the migration atomically because view removal uses `RESTRICT`. Review and migrate customer database extensions alongside the upgrade. No historic financial row is updated to backfill the new reference tables.

## Accounting operations

All endpoints below are under `/api/v1/accounting`, require the normal accounting permissions and legal entity scope, and use the authenticated actor. Lists accept `limit` (1–200, default 50) and `offset` and return `data`, `total`, `limit`, and `offset`.

| Resource | Operations | Behavior |
| --- | --- | --- |
| `/dimensions` | GET, POST | Create company dimensions with code and Arabic/English names. |
| `/dimension-values` | GET, POST | Create values belonging to a dimension in the same company. |
| `/journal-lines/{id}/dimensions` | POST | Assign a value to a draft line. Posted and reversed assignments cannot change. |
| `/exchange-rates` | GET, POST | Append effective rates; a correction creates the next version for the same currency pair and date. |
| `/posting-rules` | GET, POST | Append effective account mapping versions for a module and role. |
| `/opening-balances` | GET, POST | Create a draft opening balance for an account and fiscal period. |
| `/opening-balances/{id}/post` | POST | A different actor posts using `offset_account_id` and an optional explicit FX rate. |
| `/reconciliations` | GET, POST | Select posted journal lines for one account, currency and reconciliation date. |
| `/reconciliations/{id}/complete` | POST | A different actor completes a balanced reconciliation and freezes its lines. |

Rates, posting mapping versions and completed financial records are protected by database triggers. All mutations write audit records in the same transaction. Custom journal dimensions are included in journal details and copied into reversal lines.

A posting mapping request includes `source_module`, `code`, `name_en`, `name_ar`, `effective_from` (`YYYY-MM-DD`) and `account_id`. The active mapping with the latest effective date, then the latest version on that date, is selected using the document's accounting date. An explicit account stored on the document takes precedence. When there is no configured rule, the existing chart account fallback remains available. A configured rule that is inactive or has no effective version fails rather than silently selecting the fallback.

Connected mapping roles:

| Module | Code | Existing fallback |
| --- | --- | --- |
| `PAYABLES` | `AP_CONTROL` | `211000` |
| `PAYABLES` | `PURCHASE_EXPENSE` | `523000` |
| `PAYABLES` | `INPUT_VAT` | `114000` |
| `RECEIVABLES` | `AR_CONTROL` | `112000` |
| `PAYROLL` | Account code used by the payroll posting component | The same chart account code |

Exchange-rate storage does not automatically select a rate for a payment or realize settlement gains/losses. Each posting still needs an explicit foreign-currency snapshot; settlement FX requires the separate settlement implementation.

## Inventory valuation

`inventory.item_costs` is a warehouse-level projection across all bins. Migration 400 builds it from the existing bin balances, and database triggers apply atomic quantity/value deltas for subsequent balance inserts, updates and deletes. Concurrent bin changes cannot overwrite each other's totals. It is derived data; stock movements remain the underlying inventory ledger.

`GET /api/v1/inventory/costing/item-costs` reads this projection. `POST /api/v1/inventory/costing/snapshots` captures all current company item costs from one database statement snapshot and returns `snapshot_batch_id` and the record count. `GET` on the same path lists immutable snapshot records and accepts an optional `snapshot_batch_id`. Both lists support pagination. Snapshots represent the time they were captured; they do not reconstruct an arbitrary historical date.

## Validation evidence

The accounting integration suite covers twelve-decimal rate history, effective account mappings, scope isolation, independent opening posting and reconciliation completion, six-decimal journal amounts, immutability and custom dimensions on reversals. The inventory suite covers six-decimal values, concurrent updates to different bins and unchanged snapshots after later movements. The AP FX test uses twelve-decimal exchange rates and six-decimal prices and verifies equality between VAT in the tax ledger and general ledger.

A separate PostgreSQL 18 upgrade rehearsal migrated a disposable database from version 390 to 410 with posted large-value journals. JSONB comparisons confirmed all original header/line values and metadata remained equal; report grants and posted-record mutation guards remained effective.
