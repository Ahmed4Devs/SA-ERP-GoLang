# Sales invoice workflow

The Sales page contains a dedicated invoice workspace. Select the legal entity,
choose **New invoice**, and select an active customer, sales product, unit of
measure, and output tax code. Add lines, dates, currency, exchange rate, and any
discount or exemption description, then save the draft.

The creator reviews the saved invoice and submits it for approval. The server
locks the invoice and supplies the workflow's amount from the stored total and
exchange rate. The approval inbox routes its versioned review steps, including
the seeded finance step for base amounts above 50,000. Client context cannot
lower that amount. Rejection returns the document to draft; a fresh submission
creates another workflow instance. Submitted, approved and posted financial
snapshots cannot be edited.

Only an approved invoice may be posted, by someone other than its creator.
Approval decisions and source transitions are atomic with mandatory audit.
A successful post displays the journal reference. Errors leave the document in
its actual persisted state. Posted invoice corrections use credit notes.

All financial inputs are transmitted as decimal strings. The client supplies
neither calculated totals nor tax rates; the server resolves effective-dated tax
rules. Entity switching discards the previous form and cancels pending lookups.
Lookups require the respective customer, product, and tax read permissions.

## Current persistence precision

Migration 410 widens money and quantities to `NUMERIC(24,6)` and exchange rates
to `NUMERIC(24,12)`. Financial conversion allocates cumulative six-decimal
rounding consistently between GL and VAT. Historical records are not recomputed.

## Approval limits

Global `workflow.definitions.manage` authority can list or save limits through
`GET/PUT /api/v1/workflow/approval-limits`. Fields are legal_entity_id, user_id,
resource_type (`SALES_INVOICE`, `TREASURY_PAYMENT`, `PAYROLL_RUN`),
maximum_base_amount and version. Use version 0 to create, or the returned version
to update. Stale writes fail. Changes are audited atomically. No row means the
assigned approver's authority is unlimited; zero blocks positive amounts.
Invoice and treasury limits use document base amounts; payroll uses gross
salary, including deductions. Existing maker/checker rules remain mandatory.

## Verification

`sales-invoices.spec.ts` tests draft creation, exact decimal transport, server
totals, posting confirmation, and posted state in Arabic/English and light/dark,
plus entity switching and denied lookup behavior. `sales-invoices-real.spec.ts`
exercises login, live lookups, submission, independent approval and posting
against the Go API when `E2E_REAL_BACKEND=1`; set E2E_APPROVER_USERNAME and
E2E_APPROVER_PASSWORD for the independent user. The combined PostgreSQL race
suites verify the lifecycle, forged amounts, approval thresholds, history
immutability, closed-period rollback and audit failure. The new live browser
approval path still requires its separately configured test environment.
