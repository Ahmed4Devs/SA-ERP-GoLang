# Sales invoice workflow

The Sales page contains a dedicated invoice workspace. Select the legal entity,
choose **New invoice**, and select an active customer, sales product, unit of
measure, and output tax code. Add lines, dates, currency, exchange rate, and any
discount or exemption description, then save the draft.

The review displays the stored invoice returned by the API, including its line
snapshots, calculated tax, totals, and status. Review and acknowledge the draft
before choosing **Post invoice**. A successful post shows the journal reference
and removes the posting action. Backend permission, period, and accounting errors
remain visible; a failed post never changes the displayed status optimistically.

All financial inputs are transmitted as decimal strings. The client supplies
neither calculated totals nor tax rates; the server resolves effective-dated tax
rules. Entity switching discards the previous form and cancels pending lookups.
Lookups require the respective customer, product, and tax read permissions.

## Current persistence precision

The pre-existing sales migration defines most invoice numeric columns as
`NUMERIC(18,4)`, although the canonical specification calls for `NUMERIC(24,6)`.
A real database exercise confirmed that a unit price of `123.456789` is returned
as `123.4568`. The UI preserves the submitted string and explicitly calls out
changed saved quantities/prices before posting. Correcting persistence precision
requires a forward migration and an audit of calculation and posting precision
across the affected modules. No released migration was rewritten for this UI.

## Verification

`sales-invoices.spec.ts` tests draft creation, exact decimal transport, server
totals, posting confirmation, and posted state in Arabic/English and light/dark,
plus entity switching and denied lookup behavior. `sales-invoices-real.spec.ts`
exercises login, live lookups, draft save, and posting against the Go API and a
seeded disposable database when `E2E_REAL_BACKEND=1`. The real flow also ran as
the restricted `erp_app` PostgreSQL login.
