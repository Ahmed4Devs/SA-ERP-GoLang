# Regulatory integration boundaries

## GOSI

`GOSI_INTEGRATION_MODE=file` is the default. `manual` and `file` generate an ERP
review report; neither can submit it or report portal confirmation. `simulation`
is allowed only outside production: its result is `SIMULATED_SUBMITTED`, explicitly
marks `external_submission=false`, and leaves the real batch `GENERATED`.
Unknown receipt IDs fail. `official` fails configuration validation and adapter
operations until a verified official API contract, credentials, and acceptance
tests are supplied and implemented. No guessed portal endpoints are used.

The XML namespace `urn:erp:gosi:export:v1` identifies the application's own format,
not an officially certified GOSI upload format. Exports include employee records,
rule version IDs, bases, rates, contribution amounts, and reconciled summaries.
Validation checks parsed XML, namespace, period, IDs, rule amounts and totals,
and rejects malformed XML, multiple roots, DTDs and missing employee details.
Previously generated summary-only files remain in the database for history but
are not accepted as valid detailed exports by the adapter.

Monthly calculation requires an explicit contribution basis on every insurance
profile. There is no invented default salary. Rates and rule versions remain
stored with calculation lines. Legal applicability and new regulatory versions
must be reviewed before activation; local arithmetic tests do not certify them.
The dashboard's existing `ComplianceRate` field measures calculation coverage,
not GOSI portal acceptance.

Manual operations: review employee/rule data, export, perform the authorized
portal process, and retain the actual receipt as a document. This release does
not turn receipt attachments into verified external confirmation automatically.

## ZATCA

Simulation submissions retain `environment=SIMULATION`; they are local test
outcomes. Automatic live onboarding remains disabled because the official
signed compliance sample suite is absent. Simulator keys and samples are not
conformance evidence. Live submissions additionally require active, unexpired
X.509 CSID tokens; production requires a production CSID. A retry cannot cross
adapter environments. Non-2xx gateway responses fail regardless of a body's
claimed success, including compliance and renewal requests.

Production activation still requires the official conformance suite, sandbox
acceptance, verified signing/canonicalization/QR behavior, actual credentials,
and deployment approval. These cannot be manufactured by software tests.
No real GOSI or ZATCA submissions were made during the remediation checks.
