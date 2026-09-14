# MASTER CODING SPECIFICATION
## Production-Grade Saudi ERP Platform
### Final Unified Architecture, Database Contract & Implementation Specification

**Technology:** Go + PostgreSQL + Next.js  
**Architecture:** Modular Monolith  
**Deployment:** Private VPS / On-Premise / Dedicated Installation per Customer  
**Languages:** Arabic + English  
**Direction:** RTL + LTR  
**Themes:** Light + Dark + System  
**Country Focus:** Saudi Arabia  
**Compliance:** ZATCA-ready + GOSI-ready  

---

# CODING AGENT OPERATING PROTOCOL
## Mandatory Instructions for Codex, Claude Code, Gemini, or Any Coding Agent

This section is mandatory and has higher execution priority than ordinary implementation convenience.

The Coding Agent MUST treat this file as the canonical source of truth for the project.

## A. Specification Reading Order

Before making any code change, the Coding Agent MUST:

1. Read this `MASTER_SPEC.md`.
2. Read the current phase completely.
3. Read all prerequisite phases referenced by the current phase.
4. Inspect the current repository state.
5. Inspect existing migrations, interfaces, tests, ADRs, and documentation relevant to the task.
6. Verify whether the requested feature already exists partially or fully.
7. Compare existing implementation against this specification before modifying anything.

Do not implement from memory if the canonical specification is available.

Do not infer architectural decisions that are explicitly defined here.

## B. Context Management

For long-running coding sessions, the Coding Agent MUST maintain a compact project execution context containing at minimum:

- current phase number and title
- completed phases
- current Git branch
- latest completed commit
- pending tasks in the current phase
- migrations already created
- architectural decisions already made
- canonical interfaces already established
- known test failures
- unresolved blockers
- relevant ADRs
- database schema changes
- security-sensitive decisions
- regulatory assumptions requiring verification

When context becomes large:

1. summarize completed work
2. preserve unresolved decisions
3. preserve exact phase position
4. preserve file paths that remain unfinished
5. preserve test status
6. preserve Git status
7. continue from the summary rather than restarting architecture design

The Agent MUST NOT use context compression as a reason to forget or weaken project invariants.

## C. Phase Gating

The Coding Agent MUST NOT skip phases.

A phase is complete only when:

- required files exist
- migrations succeed
- generated code is current
- domain rules are implemented
- interfaces are implemented
- authorization is enforced
- audit behavior is implemented where applicable
- frontend work is complete where applicable
- Arabic and English are implemented
- RTL and LTR are verified
- Light and Dark themes are verified
- required tests pass
- static analysis passes
- security checks pass
- OpenAPI is updated
- documentation is updated
- Git working tree is intentionally clean
- the phase has its required Git commit

The Agent MUST NOT begin the next phase while the current phase has unresolved failing tests unless the failure is explicitly documented as an external blocker and continuing is explicitly approved.

## D. Session Resumption

When resuming after a session interruption, token/context reset, or agent restart:

1. run `git status`
2. inspect the latest commits
3. read the current phase from this specification
4. inspect `docs/adr/`
5. inspect relevant migrations
6. inspect relevant tests
7. identify unfinished files
8. run the relevant test suite
9. continue from the exact unfinished point

Never restart a completed phase from scratch unless the implementation is proven incorrect.

Never duplicate migrations, modules, or interfaces because prior context was lost.

## E. Git Discipline

Every completed phase MUST end with a meaningful Git commit.

Do not bundle unrelated phases into one commit.

Before committing:

- run formatters
- run linters
- run tests
- inspect diff
- ensure no secrets are included
- ensure generated files are current
- ensure migration order is correct
- ensure documentation reflects the implementation

Commit messages must follow the conventions defined later in this specification.

## F. Decision Logging

Any material architecture deviation requires an ADR under:

```text
docs/adr/
```

An ADR is required when changing:

- database architecture
- transaction strategy
- authentication model
- authorization model
- regulatory integration architecture
- deployment model
- primary libraries
- persistence strategy
- message/job strategy
- financial invariants
- module boundaries

An ADR MUST include:

- context
- decision
- alternatives considered
- consequences
- migration impact
- security impact
- testing impact

The Agent may not silently alter architecture.

## G. Do Not Guess

If something is already defined in this specification, use it.

If a detail is not defined:

1. prefer existing project conventions
2. prefer the simplest production-safe solution
3. preserve security and financial invariants
4. avoid adding infrastructure
5. avoid adding dependencies unless justified
6. document material decisions in an ADR

Do not invent a new framework or subsystem simply because it is familiar.

## H. Test-Before-Progress Rule

Every implementation step that changes business behavior MUST include appropriate tests before the phase is considered complete.

Financial code MUST have deterministic tests.

Security boundaries MUST have negative tests.

Database constraints MUST have integration tests against real PostgreSQL.

Frontend critical flows MUST have Playwright tests.

## I. Regulatory Safety Rule

ZATCA and GOSI behavior MUST be isolated behind versioned rules and adapters.

Never assume a regulatory percentage, field, API contract, certificate requirement, or submission rule will remain unchanged forever.

Regulatory configuration that changes over time MUST be effective-dated and versioned.

## J. Security Preservation Rule

The Coding Agent MUST never fix a problem by:

- bypassing authorization
- weakening a database constraint
- disabling CSRF
- disabling authentication
- allowing cross-entity access
- logging secrets
- hard-coding credentials
- making posted financial data mutable
- disabling failing security tests
- removing audit behavior

## K. Final Agent Checklist Before Every Phase Commit

The Agent MUST answer these internally before committing:

- Does this implementation follow the current phase?
- Did I preserve all financial invariants?
- Did I preserve legal-entity isolation?
- Did I introduce unnecessary infrastructure?
- Are migrations safe?
- Are tests meaningful?
- Are Arabic and English covered?
- Does RTL work?
- Does LTR work?
- Does Light theme work?
- Does Dark theme work?
- Did I update OpenAPI?
- Did I update documentation?
- Did I scan the diff for secrets?
- Is the commit focused?

Only then commit.

---

# 0. AUTHORITY OF THIS DOCUMENT

This document is the highest-level technical authority and canonical implementation contract for the ERP project.

The Coding Agent MUST follow this specification.

The Coding Agent MUST NOT silently:

- replace technologies
- simplify financial invariants
- change canonical database structures
- introduce cloud services
- introduce microservices
- introduce an ORM
- weaken authorization
- bypass tests
- disable security protections
- use floating-point financial arithmetic
- change accounting semantics
- change inventory-ledger semantics
- hard-code regulatory values
- remove Arabic or English support
- remove RTL/LTR support
- skip Git commits after completed phases

If implementation requirements conflict with this specification, the Coding Agent MUST prefer this specification unless an explicit architectural decision has been approved.

---

# 1. ROLE

Act as a:

- Principal Software Architect
- Staff Go Engineer
- PostgreSQL Architect
- ERP Domain Architect
- Accounting Systems Engineer
- Saudi Compliance Integration Engineer
- Frontend Architect
- Security Engineer
- QA Engineer
- DevOps Engineer
- Technical Writer

Build a real commercial ERP.

This is NOT:

- a tutorial
- a prototype
- a demo
- a CRUD generator
- a hackathon implementation

---

# 2. PRODUCT GOAL

Build a production-grade ERP suitable primarily for medium-sized Saudi organizations.

The ERP must support:

- Organization Management
- Legal Entities
- Branches
- Departments
- Cost Centers
- IAM
- RBAC
- ABAC-style Scoping
- Approval Limits
- Segregation of Duties
- Accounting
- General Ledger
- Accounts Receivable
- Accounts Payable
- Sales
- Purchasing
- Products
- Inventory
- Warehouses
- Treasury
- Banking
- Reconciliation
- Fixed Assets
- Expenses
- Budgets
- Projects
- HR
- Attendance
- Leave
- Payroll
- GOSI
- VAT
- ZATCA Electronic Invoicing
- Workflow
- Documents
- Notifications
- Audit
- Reports
- Dashboards
- Arabic
- English
- RTL
- LTR
- Dark Theme
- Light Theme

---

# 3. CORE ARCHITECTURE

Mandatory:

```text
Modular Monolith
```

Do NOT start with microservices.

Logical architecture:

```text
                         Browser
                            │
                         HTTPS
                            │
                            ▼
                         Nginx
               ┌────────────┴────────────┐
               │                         │
               ▼                         ▼
         Next.js Web                  Go API
                                   Modular Monolith
                                        │
                   ┌────────────────────┼────────────────────┐
                   │                    │                    │
                   ▼                    ▼                    ▼
              PostgreSQL           Go Worker         Integrations
                                     River           ZATCA/GOSI/etc.
```

---

# 4. DEPLOYMENT MODEL

Every customer receives an independent deployment.

```text
Customer A
├── web
├── api
├── worker
├── PostgreSQL
├── storage
├── secrets
└── backups

Customer B
├── web
├── api
├── worker
├── PostgreSQL
├── storage
├── secrets
└── backups
```

Never mix separate customers inside a shared production database in the initial product.

One customer deployment can support:

```text
Organization
├── Legal Entity A
│   ├── Riyadh Branch
│   └── Jeddah Branch
│
└── Legal Entity B
    └── Dammam Branch
```

The principal financial isolation boundary inside a deployment is:

```text
legal_entity_id
```

---

# 5. MANDATORY BACKEND STACK

Use:

```text
Go >= 1.27
net/http
Chi

PostgreSQL >= 18

pgx/v5
pgxpool

sqlc
Goose
River

log/slog

OpenTelemetry
Prometheus

OpenAPI 3.1

Argon2id

Testcontainers
Go testing
httptest
```

Do NOT replace `pgx + sqlc` with an ORM without explicit approval.

---

# 6. FRONTEND STACK

Mandatory:

```text
Next.js
React
TypeScript

Tailwind CSS
shadcn/ui
Radix UI

TanStack Query
TanStack Table

React Hook Form
Zod

Recharts
Lucide React

next-intl
Playwright
```

The frontend must implement an ERP-specific Design System rather than relying on a large opinionated dashboard framework.

---

# 7. LANGUAGE AND DIRECTION — NON-NEGOTIABLE

The application MUST support:

```text
Arabic
English
```

Locales:

```text
ar-SA
en
```

The UI MUST support both:

```text
RTL
LTR
```

Arabic:

```html
<html lang="ar" dir="rtl">
```

English:

```html
<html lang="en" dir="ltr">
```

The user MUST be able to switch languages without logging out.

Direction must automatically change with language.

---

# 8. RTL REQUIREMENTS

Prefer logical CSS properties:

```text
margin-inline-start
margin-inline-end
padding-inline-start
padding-inline-end
inset-inline-start
inset-inline-end
```

The following must be tested under RTL:

- Sidebar
- Dropdowns
- Dialogs
- Drawers
- Tables
- Pagination
- Breadcrumbs
- Forms
- Tabs
- Charts
- Invoice views
- Approval timelines
- Audit history
- Navigation arrows
- Icons where directional meaning matters

---

# 9. TRANSLATION ARCHITECTURE

Use:

```text
apps/web/messages/
├── ar.json
└── en.json
```

Do NOT hard-code UI strings throughout React components.

Translation keys should be semantic:

```text
sales.invoice.create
sales.invoice.approve
sales.invoice.post
accounting.period.closed
inventory.stock.insufficient
workflow.approval.pending
```

---

# 10. BILINGUAL MASTER DATA

Where useful, master data supports:

```text
name_ar
name_en
```

Examples:

- Account
- Department
- Cost Center
- Item
- Category
- Warehouse
- Tax Code
- Salary Component
- Fixed Asset Category

---

# 11. LIGHT / DARK THEMES

Required:

```text
light
dark
system
```

Use CSS variables and semantic design tokens.

Theme choice is persisted per user.

---

# 12. ERP UI SHELL

```text
┌────────────────────────────────────────────────────────┐
│ Topbar                                                 │
│ Search | Entity | Branch | Notifications | Theme/User │
├──────────────┬─────────────────────────────────────────┤
│ Sidebar      │ Breadcrumbs                             │
│              ├─────────────────────────────────────────┤
│ Dashboard    │ Page Header                             │
│ Accounting   │                                         │
│ Sales        │ Main Workspace                          │
│ Purchasing   │                                         │
│ Inventory    │ Tables / Forms / Charts / Documents    │
│ HR           │                                         │
│ Payroll      │                                         │
│ Reports      │                                         │
└──────────────┴─────────────────────────────────────────┘
```

Sidebar:

```text
LTR → left
RTL → right
```

---

# 13. UI COMPONENT ARCHITECTURE

```text
apps/web/components/

├── ui/
├── layout/
├── data-table/
├── forms/
├── charts/
└── erp/
```

ERP reusable components include:

- DocumentHeader
- DocumentStatus
- DocumentActions
- ApprovalTimeline
- AuditDrawer
- JournalPreview
- TotalsPanel
- TaxSummary
- ZATCAStatus
- EntitySelector
- BranchSelector
- MoneyInput
- AccountSelector
- CustomerSelector
- SupplierSelector
- ItemSelector
- WarehouseSelector

---

# 14. REPOSITORY STRUCTURE

```text
erp/
│
├── apps/
│   └── web/
│       ├── app/
│       ├── components/
│       ├── features/
│       ├── lib/
│       ├── hooks/
│       ├── messages/
│       └── tests/
│
├── cmd/
│   ├── erp-api/
│   │   └── main.go
│   ├── erp-worker/
│   │   └── main.go
│   ├── erp-cli/
│   │   └── main.go
│   └── erp-migrate/
│       └── main.go
│
├── internal/
│   ├── platform/
│   │   ├── auth/
│   │   ├── authorization/
│   │   ├── database/
│   │   ├── transaction/
│   │   ├── config/
│   │   ├── http/
│   │   ├── middleware/
│   │   ├── crypto/
│   │   ├── jobs/
│   │   ├── events/
│   │   ├── outbox/
│   │   ├── idempotency/
│   │   ├── storage/
│   │   ├── mail/
│   │   ├── localization/
│   │   ├── observability/
│   │   └── health/
│   │
│   ├── iam/
│   ├── organization/
│   ├── workflow/
│   ├── accounting/
│   ├── receivables/
│   ├── payables/
│   ├── crm/
│   ├── sales/
│   ├── purchasing/
│   ├── products/
│   ├── pricing/
│   ├── inventory/
│   ├── warehouse/
│   ├── treasury/
│   ├── banking/
│   ├── expenses/
│   ├── fixedassets/
│   ├── budgeting/
│   ├── projects/
│   ├── hr/
│   ├── attendance/
│   ├── leave/
│   ├── payroll/
│   ├── tax/
│   ├── zatca/
│   ├── gosi/
│   ├── documents/
│   ├── notifications/
│   ├── reporting/
│   ├── audit/
│   └── integrations/
│
├── pkg/
│   ├── money/
│   ├── decimal/
│   ├── identifier/
│   ├── pagination/
│   ├── clock/
│   └── errors/
│
├── database/
│   ├── migrations/
│   ├── queries/
│   └── seeds/
│
├── api/
│   └── openapi/
│
├── configs/
├── deployments/
├── docs/
├── scripts/
├── tests/
│
├── go.mod
├── go.sum
├── sqlc.yaml
├── docker-compose.yml
├── Makefile
└── README.md
```

---

# 15. DOMAIN MODULE STRUCTURE

```text
internal/accounting/

├── domain/
│   ├── entities.go
│   ├── value_objects.go
│   ├── rules.go
│   └── errors.go
│
├── application/
│   ├── commands/
│   ├── queries/
│   ├── services/
│   └── dto/
│
├── ports/
│   ├── repositories.go
│   └── services.go
│
├── infrastructure/
│   ├── postgres/
│   └── adapters/
│
├── transport/
│   └── http/
│
└── module.go
```

Dependency direction:

```text
Transport
    ↓
Application
    ↓
Domain
    ↓
Ports

Infrastructure → Ports
```

Domain packages must not import infrastructure concerns.

---

# 16. PROGRAMMING BEST PRACTICES — MANDATORY

All code MUST follow idiomatic Go and sound frontend engineering practices.

Mandatory:

- small cohesive packages
- explicit dependencies
- clear domain naming
- `context.Context` propagation
- short consumer-driven interfaces
- `errors.Is`
- `errors.As`
- error wrapping
- structured logging
- deterministic business calculations
- dependency inversion only where useful
- composition over inheritance-style abstractions
- explicit transaction boundaries
- meaningful tests
- no duplicated financial logic
- no business logic in HTTP handlers
- no business logic in React presentation components
- no hidden global mutable state
- no premature abstractions
- no speculative infrastructure

---

# 17. SECURITY BEST PRACTICES — MANDATORY

Implement secure-by-default behavior following current best practices and OWASP principles.

Mandatory:

- least privilege
- secure session cookies
- CSRF protection
- rate limiting
- Argon2id
- MFA-ready authentication
- strict input validation
- safe output encoding
- permission checks
- resource-scope validation
- secure file handling
- secret management
- secure headers
- audit trail
- database role separation
- query parameterization
- SQL injection protection
- IDOR/BOLA protection
- mass-assignment protection
- XSS protection
- dependency scanning
- security testing
- secret scanning
- secure defaults
- restrictive CORS configuration
- careful redirect validation
- file content validation
- no sensitive values in logs

Never trust browser totals, client permissions, client scopes, client approval limits, client status values, or client tax values.

---

# 18. ABSOLUTELY FORBIDDEN

Do NOT:

- use `float32`/`float64` for money
- use generic CRUD for financial posting
- use generic `Repository[T]`
- create God services
- put SQL in HTTP handlers
- put business logic in React
- let frontend enforce security
- allow mutable posted accounting records
- allow mutable posted inventory ledger
- allow direct editing of posted payroll
- hard-code VAT/GOSI regulations
- store private keys in source control
- store passwords/API keys plaintext
- use shared SaaS database initially
- introduce microservices
- introduce Kubernetes
- introduce Kafka
- introduce Elasticsearch initially
- require managed cloud services
- disable failing tests
- skip migrations
- rewrite released migrations
- silently swallow errors

---

# 19. CANONICAL GO CONTRACTS

The following contracts establish expected architectural boundaries.

---

# 20. TX MANAGER

```go
package transaction

import "context"

type Manager interface {
    WithinTx(
        ctx context.Context,
        fn func(ctx context.Context) error,
    ) error
}
```

Implementation must bind repositories to the same underlying `pgx.Tx`.

---

# 21. ACCOUNTING POSTING ENGINE

```go
package accounting

import (
    "context"
    "time"

    "github.com/google/uuid"
)

type PostingRequest struct {
    LegalEntityID uuid.UUID

    SourceModule string
    SourceType   string
    SourceID     uuid.UUID

    JournalID uuid.UUID

    AccountingDate time.Time

    CurrencyCode string
    ExchangeRate Decimal

    Reference   string
    Description string

    Lines []PostingLine
}

type PostingLine struct {
    AccountID uuid.UUID

    Description string

    TransactionDebit  Decimal
    TransactionCredit Decimal

    BaseDebit  Decimal
    BaseCredit Decimal

    BranchID     *uuid.UUID
    DepartmentID *uuid.UUID
    CostCenterID *uuid.UUID
}

type PostingResult struct {
    JournalEntryID uuid.UUID
    EntryNumber    string
}

type PostingEngine interface {
    Post(
        ctx context.Context,
        request PostingRequest,
    ) (PostingResult, error)

    Reverse(
        ctx context.Context,
        journalEntryID uuid.UUID,
        reason string,
        actorID uuid.UUID,
    ) (PostingResult, error)
}
```

PostingEngine MUST ensure:

- balanced debit/credit
- open accounting period
- active accounts
- posting permissions
- idempotency
- atomic transaction
- audit records
- immutable posted journal

---

# 22. ACCOUNT REPOSITORY

```go
type AccountRepository interface {
    GetByID(
        ctx context.Context,
        legalEntityID uuid.UUID,
        id uuid.UUID,
    ) (*Account, error)

    GetForUpdate(
        ctx context.Context,
        legalEntityID uuid.UUID,
        id uuid.UUID,
    ) (*Account, error)

    IsPostable(
        ctx context.Context,
        legalEntityID uuid.UUID,
        id uuid.UUID,
    ) (bool, error)
}
```

---

# 23. JOURNAL REPOSITORY

```go
type JournalRepository interface {
    CreateDraft(
        ctx context.Context,
        entry JournalEntry,
    ) (*JournalEntry, error)

    GetByID(
        ctx context.Context,
        legalEntityID uuid.UUID,
        id uuid.UUID,
    ) (*JournalEntry, error)

    LockForPosting(
        ctx context.Context,
        legalEntityID uuid.UUID,
        id uuid.UUID,
    ) (*JournalEntry, error)

    MarkPosted(
        ctx context.Context,
        id uuid.UUID,
        totals JournalTotals,
        actorID uuid.UUID,
    ) error

    CreateReversal(
        ctx context.Context,
        original JournalEntry,
        reversal JournalEntry,
    ) (*JournalEntry, error)
}
```

---

# 24. SALES INVOICE REPOSITORY

```go
type InvoiceRepository interface {
    Create(
        ctx context.Context,
        invoice Invoice,
    ) (*Invoice, error)

    GetByID(
        ctx context.Context,
        legalEntityID uuid.UUID,
        id uuid.UUID,
    ) (*Invoice, error)

    LockForPosting(
        ctx context.Context,
        legalEntityID uuid.UUID,
        id uuid.UUID,
    ) (*Invoice, error)

    SaveDraft(
        ctx context.Context,
        invoice Invoice,
        expectedVersion int64,
    ) error

    MarkApproved(
        ctx context.Context,
        id uuid.UUID,
        approvedBy uuid.UUID,
    ) error

    MarkPosted(
        ctx context.Context,
        id uuid.UUID,
        journalEntryID uuid.UUID,
        postedBy uuid.UUID,
    ) error

    MarkReversed(
        ctx context.Context,
        id uuid.UUID,
        reversedBy uuid.UUID,
    ) error
}
```

---

# 25. INVENTORY REPOSITORY

```go
type StockRepository interface {
    LockBalance(
        ctx context.Context,
        legalEntityID uuid.UUID,
        itemID uuid.UUID,
        warehouseID uuid.UUID,
        binID uuid.UUID,
    ) (*StockBalance, error)

    InsertMovement(
        ctx context.Context,
        movement StockMovement,
    ) error

    ApplyBalanceDelta(
        ctx context.Context,
        key StockKey,
        quantityDelta Decimal,
        costDelta Decimal,
    ) error

    Reserve(
        ctx context.Context,
        reservation StockReservation,
    ) error

    ReleaseReservation(
        ctx context.Context,
        reservationID uuid.UUID,
    ) error
}
```

---

# 26. GOSI SERVICE

```go
type ContributionInput struct {
    LegalEntityID uuid.UUID
    EmployeeID    uuid.UUID

    PayrollDate time.Time
    NationalityCode string
    ContributionBasis Decimal
}

type ContributionResult struct {
    RuleVersionID uuid.UUID

    Basis Decimal

    EmployeeRate Decimal
    EmployerRate Decimal

    EmployeeAmount Decimal
    EmployerAmount Decimal

    CalculationSnapshot map[string]any
}

type GOSIService interface {
    Calculate(
        ctx context.Context,
        input ContributionInput,
    ) (ContributionResult, error)
}
```

Regulatory rates must come from versioned database rules.

---

# 27. GOSI EXTERNAL ADAPTER

```go
type GOSIAdapter interface {
    Validate(
        ctx context.Context,
        request ExportRequest,
    ) (ValidationResult, error)

    Export(
        ctx context.Context,
        request ExportRequest,
    ) (ExportResult, error)

    Submit(
        ctx context.Context,
        request SubmissionRequest,
    ) (SubmissionResult, error)

    Status(
        ctx context.Context,
        submissionID string,
    ) (SubmissionStatus, error)
}
```

Implementations may include:

```text
ManualAdapter
FileExportAdapter
OfficialIntegrationAdapter
```

---

# 28. ZATCA XML GENERATOR

```go
type XMLGenerator interface {
    GenerateInvoice(
        ctx context.Context,
        input InvoiceData,
    ) ([]byte, error)
}
```

---

# 29. ZATCA VALIDATOR

```go
type InvoiceValidator interface {
    Validate(
        ctx context.Context,
        xml []byte,
    ) (ValidationResult, error)
}
```

---

# 30. ZATCA SIGNER

```go
type InvoiceSigner interface {
    Sign(
        ctx context.Context,
        xml []byte,
        credentials SigningCredentials,
    ) (SignedDocument, error)
}
```

---

# 31. ZATCA CLIENT

```go
type ZATCAClient interface {
    Clearance(
        ctx context.Context,
        request ClearanceRequest,
    ) (ClearanceResponse, error)

    Report(
        ctx context.Context,
        request ReportingRequest,
    ) (ReportingResponse, error)
}
```

---

# 32. ZATCA ONBOARDING CLIENT

```go
type ZATCAOnboardingClient interface {
    RequestComplianceCSID(
        ctx context.Context,
        request ComplianceCSIDRequest,
    ) (ComplianceCSIDResponse, error)

    RunComplianceChecks(
        ctx context.Context,
        request ComplianceCheckRequest,
    ) (ComplianceCheckResponse, error)

    RequestProductionCSID(
        ctx context.Context,
        request ProductionCSIDRequest,
    ) (ProductionCSIDResponse, error)

    RenewProductionCSID(
        ctx context.Context,
        request RenewalRequest,
    ) (RenewalResponse, error)
}
```

---

# 33. BLOB STORAGE

```go
type BlobStore interface {
    Put(
        ctx context.Context,
        object Object,
    ) (StoredObject, error)

    Open(
        ctx context.Context,
        key string,
    ) (io.ReadCloser, error)

    Delete(
        ctx context.Context,
        key string,
    ) error
}
```

Initial:

```text
LocalFilesystemBlobStore
```

Optional later:

```text
S3CompatibleBlobStore
```

---

# 34. AUTHORIZATION SERVICE

```go
type AuthorizationRequest struct {
    UserID uuid.UUID

    Permission string

    LegalEntityID *uuid.UUID
    BranchID      *uuid.UUID

    ResourceType string
    ResourceID   *uuid.UUID

    Amount *Decimal
}

type AuthorizationDecision struct {
    Allowed bool
    Reason  string
}

type AuthorizationService interface {
    Authorize(
        ctx context.Context,
        request AuthorizationRequest,
    ) (AuthorizationDecision, error)
}
```

---

# 35. NUMBER SEQUENCE SERVICE

```go
type NumberSequenceService interface {
    Next(
        ctx context.Context,
        request NumberRequest,
    ) (string, error)
}
```

Allocation MUST be transactionally locked using PostgreSQL.

---

# 36. CLOCK

```go
type Clock interface {
    Now() time.Time
}
```

Production: `SystemClock`  
Tests: `FixedClock`

Financial/date calculations should not call `time.Now()` throughout domain code.

---

# 37. CANONICAL DATABASE BASELINE

Mandatory:

```text
PostgreSQL >= 18
```

Primary identifiers:

```sql
UUID DEFAULT uuidv7()
```

Required extension:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

---

# 38. POSTGRESQL SCHEMAS

Canonical schemas:

```text
ref

platform
iam
organization
workflow

accounting
tax

products
pricing

receivables
sales

payables
purchasing

warehouse
inventory

treasury
banking

expenses
fixedassets
budgeting
projects

hr
attendance
leave
payroll
gosi

zatca

documents
notifications
integration

reporting
audit
```

---

# 39. DATABASE TYPE CONTRACT

```text
Money          NUMERIC(24,6)
Quantity       NUMERIC(24,6)
Exchange Rate  NUMERIC(24,12)
Percentage     NUMERIC(12,8)
Timestamp      TIMESTAMPTZ
Business Date  DATE
Identifier     UUID
```

Never use floating database types for financial values.

---

# 40. CANONICAL POSTGRESQL SCHEMA CONTRACT

The following schemas and tables are mandatory. The migration implementation must reproduce these contracts and all necessary primary keys, foreign keys, CHECK constraints, uniqueness constraints, legal-entity scoping constraints, and indexes.

## 40.1 `ref.currencies`

```sql
CREATE TABLE ref.currencies (
    code                CHAR(3) PRIMARY KEY,
    name_en             TEXT NOT NULL,
    name_ar             TEXT,
    decimal_places      SMALLINT NOT NULL DEFAULT 2,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT ck_currency_code_upper
        CHECK (code = upper(code)),

    CONSTRAINT ck_currency_decimal_places
        CHECK (decimal_places BETWEEN 0 AND 6)
);
```

## 40.2 `ref.countries`

```sql
CREATE TABLE ref.countries (
    code                CHAR(2) PRIMARY KEY,
    name_en             TEXT NOT NULL,
    name_ar             TEXT,
    phone_prefix        TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT ck_country_code_upper
        CHECK (code = upper(code))
);
```

## 40.3 `platform.outbox_events`

```sql
CREATE TABLE platform.outbox_events (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    event_type          TEXT NOT NULL,
    aggregate_type      TEXT NOT NULL,
    aggregate_id        UUID,
    payload             JSONB NOT NULL,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    available_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at        TIMESTAMPTZ,
    attempt_count       INTEGER NOT NULL DEFAULT 0,
    last_error          TEXT,

    CONSTRAINT ck_outbox_attempt_count
        CHECK (attempt_count >= 0)
);

CREATE INDEX idx_outbox_pending
ON platform.outbox_events (available_at, occurred_at)
WHERE published_at IS NULL;

CREATE INDEX idx_outbox_aggregate
ON platform.outbox_events (aggregate_type, aggregate_id);
```

## 40.4 `platform.idempotency_keys`

```sql
CREATE TABLE platform.idempotency_keys (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    scope               TEXT NOT NULL,
    idempotency_key     TEXT NOT NULL,
    request_hash        TEXT NOT NULL,
    state               TEXT NOT NULL DEFAULT 'PROCESSING',
    response_status     INTEGER,
    response_body       JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ NOT NULL,

    CONSTRAINT uq_idempotency_scope_key
        UNIQUE (scope, idempotency_key),

    CONSTRAINT ck_idempotency_state
        CHECK (state IN ('PROCESSING','COMPLETED','FAILED'))
);
```

## 40.5 IAM Canonical Tables

Mandatory:

```text
iam.users
iam.password_credentials
iam.sessions
iam.login_attempts
iam.mfa_factors
iam.roles
iam.permissions
iam.role_permissions
iam.user_role_assignments
iam.service_accounts
iam.api_keys
```

`iam.users` must include:

```text
id
username
email
phone_number
display_name
preferred_locale
preferred_timezone
status
last_login_at
created_at
updated_at
disabled_at
version
```

Required uniqueness:

```text
case-insensitive username when not null
case-insensitive email when not null
phone_number when not null
```

Session and API tokens are stored hashed.

## 40.6 Organization Canonical Tables

Mandatory:

```text
organization.organizations
organization.legal_entities
organization.tax_registrations
organization.branches
organization.business_units
organization.departments
organization.cost_centers
organization.addresses
organization.number_sequences
```

Every major entity must expose legal-entity scope constraints where required.

## 40.7 Accounting Canonical Tables

Mandatory:

```text
accounting.chart_of_accounts
accounting.accounts
accounting.fiscal_years
accounting.fiscal_periods
accounting.journals
accounting.journal_entries
accounting.journal_lines
accounting.dimensions
accounting.dimension_values
accounting.journal_line_dimensions
accounting.posting_rules
accounting.posting_rule_versions
accounting.exchange_rates
accounting.opening_balances
accounting.account_balances
accounting.reconciliations
accounting.reconciliation_lines
```

### Accounts

Required fields:

```text
id
legal_entity_id
chart_id
parent_id
code
name_en
name_ar
account_type
normal_balance
allow_posting
is_control_account
is_active
created_at
updated_at
version
```

Account types:

```text
ASSET
LIABILITY
EQUITY
REVENUE
EXPENSE
```

### Fiscal Periods

States:

```text
OPEN
SOFT_CLOSED
CLOSED
LOCKED
```

### Journal Entries

Required fields:

```text
id
legal_entity_id
journal_id
fiscal_period_id
entry_number
accounting_date
status
source_module
source_type
source_id
reference
description
currency_code
exchange_rate
total_debit
total_credit
created_by
approved_by
posted_by
created_at
approved_at
posted_at
reversed_at
reversal_of_id
version
```

States:

```text
DRAFT
SUBMITTED
APPROVED
POSTED
REVERSED
CANCELLED
```

Invariant:

```text
POSTED or REVERSED => total_debit = total_credit
```

### Journal Lines

Required fields:

```text
id
journal_entry_id
legal_entity_id
line_number
account_id
description
currency_code
exchange_rate
transaction_debit
transaction_credit
base_debit
base_credit
branch_id
department_id
cost_center_id
```

Each line must contain debit XOR credit.

## 40.8 Accounting Immutability

Database triggers MUST prevent UPDATE/DELETE of posted/reversed journal entries and their financial child lines.

Corrections happen through reversal/correction.

## 40.9 Tax Canonical Tables

```text
tax.tax_categories
tax.tax_codes
tax.tax_rate_versions
tax.tax_transactions
tax.tax_transaction_lines
tax.tax_periods
tax.tax_returns
tax.exemption_reasons
tax.zero_rate_reasons
```

Tax rates are effective-dated and historically reproducible.

## 40.10 Products Canonical Tables

```text
products.categories
products.items
products.item_variants
products.skus
products.barcodes
products.units_of_measure
products.uom_conversions
products.item_uoms
products.item_tax_profiles
```

Item types:

```text
STOCK
SERVICE
NON_STOCK
ASSET
```

Tracking:

```text
NONE
LOT
SERIAL
```

## 40.11 Warehouse Canonical Tables

```text
warehouse.warehouses
warehouse.bin_locations
```

Every warehouse SHOULD have a default system bin.

## 40.12 Inventory Canonical Tables

```text
inventory.stock_movements
inventory.stock_balances
inventory.stock_reservations
inventory.transfer_orders
inventory.transfer_order_lines
inventory.stock_adjustments
inventory.stock_adjustment_lines
inventory.stock_counts
inventory.stock_count_lines
inventory.lots
inventory.serial_numbers
inventory.cost_layers
inventory.item_costs
inventory.valuation_snapshots
```

`inventory.stock_movements` is the source-of-truth movement ledger.

Movement types:

```text
OPENING
PURCHASE_RECEIPT
SALES_ISSUE
CUSTOMER_RETURN
SUPPLIER_RETURN
TRANSFER_OUT
TRANSFER_IN
ADJUSTMENT_IN
ADJUSTMENT_OUT
PRODUCTION_IN
PRODUCTION_OUT
REVERSAL
```

Posted stock movements are immutable.

`inventory.stock_balances` is a projection, not original truth.

## 40.13 Customers / AR Canonical Tables

```text
receivables.customers
receivables.customer_accounts
receivables.payment_terms
receivables.customer_credit_limits
receivables.receipts
receivables.receipt_allocations
receivables.open_items
```

## 40.14 Sales Canonical Tables

```text
sales.quotations
sales.quotation_lines
sales.orders
sales.order_lines
sales.deliveries
sales.delivery_lines
sales.invoices
sales.invoice_lines
sales.credit_notes
sales.credit_note_lines
```

Invoice types:

```text
STANDARD
SIMPLIFIED
```

Invoice line snapshots MUST preserve:

```text
item_id
description
uom
quantity
unit_price
gross_amount
discount_amount
taxable_amount
tax_code
tax_rate_percent
tax_amount
net_amount
exemption reason
```

Historical invoices must not recalculate tax from current configuration.

## 40.15 Suppliers / AP Canonical Tables

```text
payables.suppliers
payables.supplier_accounts
payables.payment_terms
payables.supplier_invoices
payables.supplier_invoice_lines
payables.supplier_credit_notes
payables.payments
payables.payment_allocations
payables.open_items
```

## 40.16 Purchasing Canonical Tables

```text
purchasing.purchase_requests
purchasing.purchase_request_lines
purchasing.rfqs
purchasing.rfq_lines
purchasing.supplier_quotes
purchasing.supplier_quote_lines
purchasing.purchase_orders
purchasing.purchase_order_lines
purchasing.goods_receipts
purchasing.goods_receipt_lines
```

## 40.17 Treasury / Banking Canonical Tables

```text
treasury.cash_accounts
treasury.receipts
treasury.payments
treasury.transfers

banking.bank_accounts
banking.statements
banking.statement_lines
banking.reconciliations
banking.reconciliation_matches
```

Sensitive banking data must be encrypted where appropriate.

## 40.18 HR Canonical Tables

```text
hr.employees
hr.employee_contacts
hr.employee_documents
hr.employee_identifiers
hr.jobs
hr.positions
hr.contracts
hr.contract_versions
hr.bank_accounts
hr.qualifications
hr.emergency_contacts
```

Sensitive identity fields must be permission-restricted and encrypted where appropriate.

## 40.19 Attendance Canonical Tables

```text
attendance.shifts
attendance.shift_assignments
attendance.records
attendance.adjustments
attendance.overtime_requests
```

## 40.20 Leave Canonical Tables

```text
leave.types
leave.policies
leave.employee_balances
leave.requests
leave.transactions
```

Leave balances must remain reconcilable.

## 40.21 Payroll Canonical Tables

```text
payroll.salary_components
payroll.salary_structures
payroll.salary_structure_lines
payroll.employee_salary_assignments
payroll.periods
payroll.runs
payroll.run_employees
payroll.calculation_lines
payroll.payslips
payroll.loans
payroll.loan_installments
payroll.advances
payroll.end_of_service_calculations
```

Payroll states:

```text
DRAFT
CALCULATING
CALCULATED
UNDER_REVIEW
APPROVED
POSTED
PAID
CANCELLED
REVERSED
```

Posted payroll is financially immutable.

Payroll must preserve calculation breakdowns.

## 40.22 GOSI Canonical Tables

```text
gosi.contribution_rules
gosi.contribution_rule_versions
gosi.employee_insurance_profiles
gosi.contribution_calculations
gosi.contribution_calculation_lines
gosi.export_batches
gosi.integration_attempts
```

Rule versions contain:

```text
effective_from
effective_to
nationality scope
employee rate
employer rate
minimum basis
maximum basis
calculation config
legal reference
```

Never hard-code contribution percentages into Go source.

## 40.23 Fixed Assets Canonical Tables

```text
fixedassets.categories
fixedassets.assets
fixedassets.asset_locations
fixedassets.asset_assignments
fixedassets.depreciation_books
fixedassets.depreciation_schedules
fixedassets.depreciation_runs
fixedassets.depreciation_lines
fixedassets.transfers
fixedassets.impairments
fixedassets.disposals
```

Initial depreciation method:

```text
STRAIGHT_LINE
```

## 40.24 Budgeting Canonical Tables

```text
budgeting.budgets
budgeting.budget_versions
budgeting.budget_lines
```

Dimensions include:

```text
account
department
cost_center
project
period
```

## 40.25 Documents Canonical Tables

```text
documents.files
documents.file_versions
documents.entity_attachments
```

Store:

```text
storage_provider
storage_key
original_filename
mime_type
byte_size
sha256_hash
uploaded_by
created_at
```

## 40.26 ZATCA Canonical Tables

```text
zatca.egs_units
zatca.csids
zatca.documents
zatca.document_hashes
zatca.submissions
zatca.submission_attempts
zatca.validation_results
zatca.validation_messages
zatca.invoice_sequences
zatca.configuration_versions
```

Submission modes:

```text
CLEARANCE
REPORTING
```

Statuses:

```text
NOT_GENERATED
GENERATED
VALIDATED
SIGNED
QUEUED
SUBMITTING
CLEARED
REPORTED
WARNING
REJECTED
RETRY_PENDING
FAILED
```

Environments:

```text
SANDBOX
SIMULATION
PRODUCTION
```

Persist original compliance artifacts and submission history.

## 40.27 Workflow Canonical Tables

```text
workflow.definitions
workflow.definition_versions
workflow.steps
workflow.step_conditions
workflow.step_approvers
workflow.instances
workflow.instance_steps
workflow.approval_requests
workflow.approval_decisions
workflow.delegations
```

Workflow definitions are versioned.

## 40.28 Audit Canonical Tables

```text
audit.events
audit.entity_changes
audit.security_events
audit.access_events
```

Audit is append-only.

The normal app role may INSERT/SELECT but not UPDATE/DELETE audit history.

## 40.29 Notifications Canonical Tables

```text
notifications.notifications
```

## 40.30 Reporting Layer

Use views, materialized views, summary tables, and read models under:

```text
reporting.*
```

Examples:

```text
reporting.trial_balance
reporting.general_ledger
reporting.customer_aging
reporting.supplier_aging
reporting.inventory_valuation
reporting.sales_daily
reporting.payroll_summary
```

Reporting is never transactional source of truth.

---

# 41. COMPOSITE LEGAL-ENTITY FOREIGN KEYS

Where practical use:

```sql
FOREIGN KEY (resource_id, legal_entity_id)
REFERENCES target_table(id, legal_entity_id)
```

Parent tables must expose:

```sql
UNIQUE(id, legal_entity_id)
```

This prevents cross-company references at the database layer.

---

# 42. DATABASE STATUS POLICY

Prefer:

```text
TEXT + CHECK
```

for evolving workflow statuses rather than PostgreSQL native ENUMs unless there is a strong reason otherwise.

---

# 43. NO GENERIC SOFT DELETE

Do not add `deleted_at` to every table.

Master data uses:

```text
is_active
status
effective_to
```

Financial transactions use reversal/cancellation according to domain rules.

---

# 44. IMMUTABLE BUSINESS RECORDS

Database-level protection is mandatory for:

- posted accounting entries
- posted accounting lines
- stock movements
- posted payroll results
- final ZATCA signed artifacts/hashes
- audit events

---

# 45. OPTIMISTIC CONCURRENCY

Mutable aggregate roots include:

```sql
version BIGINT NOT NULL DEFAULT 1
```

Update with expected version.

Stale update returns:

```text
409 CONCURRENCY_CONFLICT
```

---

# 46. PESSIMISTIC LOCKING

Use PostgreSQL row locks where required:

```sql
SELECT ... FOR UPDATE;
```

Examples:

- number allocation
- invoice posting
- payment allocation
- stock reservation
- cost-layer consumption
- payroll posting
- fiscal close

---

# 47. NUMBERING

Business document numbers must be transactionally generated.

Examples:

```text
INV-RYD-2026-000001
PO-RYD-2026-000001
PAY-2026-000001
JE-2026-000001
```

---

# 48. TRANSACTIONAL CONSISTENCY

Critical operations use one PostgreSQL transaction.

Example invoice posting:

```text
BEGIN

Lock Invoice
Validate permissions
Validate workflow
Validate period
Validate tax
Create journal
Create tax transaction
Create stock movements
Update AR
Mark invoice POSTED
Insert audit event
Insert outbox event
Enqueue transactional River job

COMMIT
```

Any mandatory failure:

```text
ROLLBACK
```

---

# 49. TRANSACTIONAL OUTBOX

External side effects such as:

- Email
- Notifications
- Document generation
- ZATCA
- non-critical integrations
- report refresh

must not weaken internal transactional consistency.

Use `platform.outbox_events` and/or River transactional jobs.

---

# 50. IDEMPOTENCY

Mandatory for:

- financial posting
- payments
- payroll posting
- ZATCA submission
- webhooks
- critical integrations

HTTP uses:

```text
Idempotency-Key
```

Workers use stable business uniqueness keys.

---

# 51. API ARCHITECTURE

Base:

```text
/api/v1
```

Use explicit domain commands for protected state transitions.

Correct:

```http
POST /sales/invoices/{id}/submit
POST /sales/invoices/{id}/approve
POST /sales/invoices/{id}/post
POST /sales/invoices/{id}/reverse
```

Do not expose unrestricted generic status PATCH operations.

---

# 52. API ERROR CONTRACT

```json
{
  "error": {
    "code": "ACCOUNTING_PERIOD_CLOSED",
    "message": "The accounting period is closed.",
    "details": {},
    "request_id": "019..."
  }
}
```

Never expose raw internal errors, SQL, stack traces, internal paths, or secrets.

---

# 53. AUTHENTICATION

Browser authentication uses secure HttpOnly session cookies.

Required:

- Secure
- HttpOnly
- SameSite
- CSRF protection
- rotation
- expiration
- idle timeout
- revocation

Do not store long-lived browser auth tokens in `localStorage`.

---

# 54. AUTHORIZATION

Server-side authorization is mandatory.

Model:

```text
RBAC
+
Resource Scope
+
ABAC-style Conditions
```

Permission format:

```text
<module>.<resource>.<action>
```

Examples:

```text
accounting.journal.post
sales.invoice.approve
inventory.adjustment.approve
payroll.run.post
zatca.configuration.manage
audit.read
```

---

# 55. SEGREGATION OF DUTIES

Support policies:

```text
cannot_approve_own_document
cannot_post_own_document
cannot_approve_above_limit
cannot_post_without_approval
```

---

# 56. BACKGROUND JOBS

Use River.

Logical queues:

```text
critical
accounting
inventory
payroll
zatca
gosi
integrations
documents
reports
notifications
maintenance
```

Required jobs include:

```text
zatca.generate_document
zatca.validate_document
zatca.submit_clearance
zatca.submit_reporting
zatca.retry_submission

inventory.recalculate_cost
inventory.refresh_balance

accounting.refresh_balances

payroll.calculate
payroll.generate_payslip

gosi.calculate
gosi.build_export

report.generate_pdf
report.generate_excel

email.send
notification.dispatch

bank.import_statement

maintenance.cleanup_sessions
maintenance.verify_integrity
```

---

# 57. OBSERVABILITY

Every request receives:

```text
request_id
trace_id
```

Use:

```text
slog
OpenTelemetry
Prometheus
```

Never log secrets.

---

# 58. DATABASE SECURITY

Production roles:

```text
erp_owner
erp_migrator
erp_app
erp_readonly
erp_backup
```

Runtime does not connect as PostgreSQL superuser.

---

# 59. SECRETS

Never commit real:

- database passwords
- API keys
- ZATCA private keys
- certificates
- SMTP credentials
- banking credentials
- session secrets

Use dummy values in examples.

---

# 60. `.env.example`

Provide development-safe dummy configuration covering:

```text
ERP_ENV
ERP_HTTP_ADDR
PUBLIC_APP_URL
PUBLIC_API_URL
DATABASE_URL
SESSION_SECRET
CSRF_SECRET
STORAGE_DRIVER
STORAGE_LOCAL_PATH
REDIS_ENABLED
REDIS_URL
OTEL_ENABLED
OTEL_EXPORTER_OTLP_ENDPOINT
ZATCA_ENV
ZATCA_BASE_URL
ZATCA_COMPLIANCE_CSID
ZATCA_PRODUCTION_CSID
ZATCA_PRIVATE_KEY_PATH
GOSI_INTEGRATION_MODE
SMTP_HOST
SMTP_PORT
SMTP_USERNAME
SMTP_PASSWORD
DEFAULT_TIMEZONE
DEFAULT_LOCALE
```

No real production credentials.

---

# 61. TESTING STRATEGY

Mandatory:

- Domain Unit Tests
- Application Tests
- Repository Tests
- PostgreSQL Integration Tests
- HTTP Tests
- Authorization Tests
- Accounting Invariant Tests
- Inventory Invariant Tests
- Payroll Tests
- GOSI Tests
- ZATCA Tests
- Security Tests
- Playwright Browser E2E
- RTL E2E
- LTR E2E
- Light Theme E2E
- Dark Theme E2E
- Performance Tests

Use real PostgreSQL through Testcontainers.

---

# 62. ACCOUNTING INVARIANTS

Tests MUST prove:

- Debit = Credit
- Posted journal immutable
- Closed period blocks posting
- Disabled account rejected
- Duplicate posting prevented
- Reversal balanced
- Foreign currency values correct
- Cross-legal-entity posting rejected

---

# 63. INVENTORY INVARIANTS

Tests MUST prove:

- movement updates balance
- duplicate movement blocked
- reservation follows policy
- negative stock follows configured rules
- posted movement immutable
- transfer out/in matches
- valuation reconciles

---

# 64. PAYROLL INVARIANTS

Tests MUST prove:

- same inputs produce same result
- posted payroll immutable
- calculation explainable
- GOSI version preserved
- journal balanced
- duplicate post impossible

---

# 65. SECURITY TESTS

Mandatory coverage:

- unauthenticated access
- permission bypass
- cross-company access
- cross-branch access
- IDOR/BOLA
- CSRF
- session fixation
- revoked session
- SQL injection attempts
- XSS handling
- mass assignment
- file upload abuse
- oversized requests
- rate limiting
- privilege escalation

---

# 66. BILINGUAL UI TESTS

Playwright must test critical workflows under:

```text
English + LTR
Arabic + RTL
```

Important scenarios include:

- Login
- Dashboard
- Customer List
- Invoice List
- Invoice Detail
- Sales Invoice Form
- Purchase Order
- Inventory
- Journal Entry
- Payroll
- ZATCA
- Reports
- Settings

---

# 67. THEME TESTS

Critical UI must be verified under:

```text
Light
Dark
```

Theme switching must not break authentication, form state, chart contrast, table readability, or RTL.

---

# 68. DATABASE MIGRATION ORDER

Initial logical migration order:

```text
000001_reference
000010_platform
000020_iam
000030_organization
000040_workflow
000050_accounting
000060_tax
000070_products
000080_warehouse
000090_inventory
000100_receivables
000110_sales
000120_payables
000130_purchasing
000140_treasury
000150_banking
000160_hr
000170_attendance
000180_leave
000190_payroll
000200_gosi
000210_fixed_assets
000220_budgeting
000230_zatca
000240_documents
000250_notifications
000260_audit
000270_reporting
000280_indexes
000290_immutability
000300_seed_permissions
```

Use a consistent Goose filename convention.

Do not place the entire schema in one migration.

---

# 69. SQLC STRUCTURE

```text
database/queries/

├── iam/
├── organization/
├── workflow/
├── accounting/
├── tax/
├── products/
├── inventory/
├── receivables/
├── sales/
├── payables/
├── purchasing/
├── treasury/
├── banking/
├── hr/
├── payroll/
├── gosi/
├── fixedassets/
├── zatca/
├── audit/
└── reporting/
```

---

# 70. IMPLEMENTATION PHASE RULE

Every phase must follow:

1. Review specification.
2. Inspect existing implementation.
3. Create database migration when needed.
4. Add sqlc queries.
5. Run sqlc generation.
6. Implement domain model.
7. Implement domain rules.
8. Implement ports/interfaces.
9. Implement application service.
10. Implement PostgreSQL adapter.
11. Implement API transport.
12. Implement authorization.
13. Implement audit behavior.
14. Implement frontend.
15. Add Arabic translations.
16. Add English translations.
17. Verify RTL.
18. Verify LTR.
19. Verify Light theme.
20. Verify Dark theme.
21. Add unit tests.
22. Add integration tests.
23. Add browser tests where applicable.
24. Run all relevant tests.
25. Run static analysis.
26. Run security checks.
27. Update OpenAPI.
28. Update documentation.
29. Git commit.
30. Only then move to the next phase.

---

# 71. PHASE 0 — BOOTSTRAP

Create:

```text
go.mod
go.sum
Makefile
README.md
docker-compose.yml
sqlc.yaml

cmd/erp-api/main.go
cmd/erp-worker/main.go
cmd/erp-cli/main.go

internal/platform/config/
internal/platform/health/

apps/web/

docs/architecture/
docs/adr/
```

Frontend:

```text
Next.js
TypeScript
Tailwind
shadcn/ui
next-intl
theme provider
```

Acceptance:

- `go test ./...`
- frontend lint
- frontend typecheck
- frontend build
- Docker Compose starts
- health endpoint works
- Arabic page works
- English page works
- RTL works
- LTR works
- Light/Dark switch works

Commit:

```text
chore: bootstrap production ERP platform
```

---

# 72. PHASE 1 — DATABASE FOUNDATION

Create:

```text
database/migrations/000001_reference.sql
database/migrations/000010_platform.sql

internal/platform/database/db.go
internal/platform/transaction/manager.go

database/queries/platform/

tests/integration/database/
```

Implement:

```text
pgxpool
Goose
sqlc
Testcontainers
TxManager
```

Commit:

```text
feat(database): establish PostgreSQL persistence foundation
```

---

# 73. PHASE 2 — OBSERVABILITY / HTTP

Files:

```text
internal/platform/http/server.go
internal/platform/middleware/request_id.go
internal/platform/middleware/recovery.go
internal/platform/middleware/security_headers.go
internal/platform/observability/logging.go
internal/platform/observability/tracing.go
internal/platform/observability/metrics.go
```

Commit:

```text
feat(platform): add secure HTTP and observability foundation
```

---

# 74. PHASE 3 — IAM

Files:

```text
database/migrations/000020_iam.sql
database/queries/iam/

internal/iam/domain/
internal/iam/application/
internal/iam/ports/
internal/iam/infrastructure/postgres/
internal/iam/transport/http/

apps/web/features/auth/

apps/web/messages/ar.json
apps/web/messages/en.json
```

Implement:

```text
User
Password
Session
Login
Logout
Logout All
Session Rotation
Argon2id
Login Rate Limit
```

Commit:

```text
feat(iam): implement secure ERP authentication
```

---

# 75. PHASE 4 — ORGANIZATION

Files:

```text
database/migrations/000030_organization.sql
database/queries/organization/
internal/organization/...
apps/web/features/organization/
```

Implement:

```text
Organization
Legal Entities
Branches
Departments
Cost Centers
Number Sequences
```

Commit:

```text
feat(organization): implement organization hierarchy and numbering
```

---

# 76. PHASE 5 — AUTHORIZATION

Files:

```text
internal/platform/authorization/
internal/iam/application/authorization/
tests/security/authorization/
```

Implement:

```text
RBAC
Legal Entity Scope
Branch Scope
Approval Limits
Permission middleware
```

Commit:

```text
feat(authz): implement scoped ERP authorization engine
```

---

# 77. PHASE 6 — AUDIT

Files:

```text
database/migrations/000260_audit.sql
database/queries/audit/
internal/audit/
apps/web/features/audit/
```

Commit:

```text
feat(audit): implement append-only ERP audit trail
```

---

# 78. PHASE 7 — WORKFLOW

Create:

```text
database/migrations/000040_workflow.sql
internal/workflow/domain/
internal/workflow/application/
internal/workflow/ports/
internal/workflow/infrastructure/
apps/web/features/workflow/
```

Implement:

```text
workflow definitions
versioning
approval steps
conditions
delegation
maker/checker
approval history
```

Commit:

```text
feat(workflow): implement versioned approval workflow engine
```

---

# 79. PHASE 8 — ACCOUNTING FOUNDATION

Create:

```text
database/migrations/000050_accounting.sql
database/queries/accounting/

internal/accounting/domain/
internal/accounting/application/
internal/accounting/ports/
internal/accounting/infrastructure/postgres/
internal/accounting/transport/http/

apps/web/features/accounting/
```

Implement:

```text
Chart of Accounts
Fiscal Years
Fiscal Periods
Journals
Journal Entries
Journal Lines
```

Commit:

```text
feat(accounting): implement general ledger foundation
```

---

# 80. PHASE 9 — POSTING ENGINE

Files:

```text
internal/accounting/application/posting/
internal/accounting/domain/posting.go
internal/accounting/ports/posting.go
tests/integration/accounting/posting_test.go
```

Implement:

```text
PostingEngine
immutability triggers
reversal
idempotency
period locking
```

Commit:

```text
feat(accounting): implement transactional posting engine
```

---

# 81. PHASE 10 — TAX

Files:

```text
database/migrations/000060_tax.sql
database/queries/tax/
internal/tax/
apps/web/features/tax/
```

Implement:

```text
tax codes
tax categories
rate versions
VAT calculation
tax transaction
```

Commit:

```text
feat(tax): implement effective-dated VAT engine
```

---

# 82. PHASE 11 — PRODUCTS

Files:

```text
database/migrations/000070_products.sql
database/queries/products/
internal/products/
apps/web/features/products/
```

Commit:

```text
feat(products): implement ERP product master
```

---

# 83. PHASE 12 — WAREHOUSE / INVENTORY

Files:

```text
database/migrations/000080_warehouse.sql
database/migrations/000090_inventory.sql

database/queries/inventory/

internal/warehouse/
internal/inventory/

apps/web/features/inventory/
```

Implement:

```text
Warehouses
Bins
Movement Ledger
Balances
Reservations
Transfers
Adjustments
Counts
```

Commit:

```text
feat(inventory): implement immutable inventory ledger
```

---

# 84. PHASE 13 — COSTING

Files:

```text
internal/inventory/domain/costing.go
internal/inventory/application/costing/
database/queries/inventory/costing.sql
```

Initial implementation:

```text
Weighted Average
```

Structure for:

```text
FIFO
Standard
```

Commit:

```text
feat(inventory): implement inventory costing engine
```

---

# 85. PHASE 14 — AR / CUSTOMERS

Files:

```text
database/migrations/000100_receivables.sql
internal/receivables/
apps/web/features/customers/
apps/web/features/receivables/
```

Commit:

```text
feat(receivables): implement accounts receivable foundation
```

---

# 86. PHASE 15 — SALES

Files:

```text
database/migrations/000110_sales.sql
database/queries/sales/
internal/sales/
apps/web/features/sales/
```

Implement:

```text
Quotation
Order
Delivery
Invoice
Credit Note
Posting integration
Inventory integration
Tax integration
```

Commit:

```text
feat(sales): implement order-to-cash workflow
```

---

# 87. PHASE 16 — AP / SUPPLIERS

Files:

```text
database/migrations/000120_payables.sql
internal/payables/
apps/web/features/suppliers/
apps/web/features/payables/
```

Commit:

```text
feat(payables): implement accounts payable foundation
```

---

# 88. PHASE 17 — PURCHASING

Files:

```text
database/migrations/000130_purchasing.sql
internal/purchasing/
apps/web/features/purchasing/
```

Implement:

```text
Purchase Request
RFQ
Supplier Quote
PO
Receipt
Supplier Invoice
```

Commit:

```text
feat(purchasing): implement procure-to-pay workflow
```

---

# 89. PHASE 18 — TREASURY / BANKING

Files:

```text
database/migrations/000140_treasury.sql
database/migrations/000150_banking.sql

internal/treasury/
internal/banking/

apps/web/features/treasury/
apps/web/features/banking/
```

Commit:

```text
feat(treasury): implement treasury and bank reconciliation
```

---

# 90. PHASE 19 — HR

Files:

```text
database/migrations/000160_hr.sql
internal/hr/
apps/web/features/hr/
```

Commit:

```text
feat(hr): implement employee and contract management
```

---

# 91. PHASE 20 — ATTENDANCE / LEAVE

Files:

```text
database/migrations/000170_attendance.sql
database/migrations/000180_leave.sql

internal/attendance/
internal/leave/

apps/web/features/attendance/
apps/web/features/leave/
```

Commit:

```text
feat(hr): implement attendance and leave management
```

---

# 92. PHASE 21 — PAYROLL

Files:

```text
database/migrations/000190_payroll.sql
database/queries/payroll/

internal/payroll/domain/
internal/payroll/application/
internal/payroll/ports/
internal/payroll/infrastructure/

apps/web/features/payroll/
```

Implement:

```text
Salary Components
Structures
Payroll Runs
Calculation Engine
Calculation Breakdown
Payslips
```

Commit:

```text
feat(payroll): implement deterministic payroll engine
```

---

# 93. PHASE 22 — GOSI

Files:

```text
database/migrations/000200_gosi.sql
database/queries/gosi/

internal/gosi/domain/
internal/gosi/application/
internal/gosi/ports/
internal/gosi/infrastructure/

apps/web/features/gosi/
```

Implement:

```text
GOSIService
Rule Versions
Contribution Calculation
Export Adapter
```

Commit:

```text
feat(gosi): implement versioned social insurance engine
```

---

# 94. PHASE 23 — FIXED ASSETS

Files:

```text
database/migrations/000210_fixed_assets.sql
internal/fixedassets/
apps/web/features/fixed-assets/
```

Commit:

```text
feat(fixedassets): implement fixed asset lifecycle
```

---

# 95. PHASE 24 — BUDGETING

Files:

```text
database/migrations/000220_budgeting.sql
internal/budgeting/
apps/web/features/budgeting/
```

Commit:

```text
feat(budgeting): implement budget management
```

---

# 96. PHASE 25 — ZATCA CORE

Files:

```text
database/migrations/000230_zatca.sql

internal/zatca/domain/
internal/zatca/application/
internal/zatca/ports/
internal/zatca/xml/
internal/zatca/signing/
internal/zatca/validation/

apps/web/features/zatca/
```

Implement:

```text
XML Generation
Hashing
Validation
Signing
QR
CSID model
EGS model
```

Commit:

```text
feat(zatca): implement Saudi e-invoice compliance core
```

---

# 97. PHASE 26 — ZATCA INTEGRATION

Files:

```text
internal/zatca/infrastructure/client/
internal/zatca/application/onboarding/
internal/zatca/application/submission/
internal/zatca/jobs/
```

Implement:

```text
Sandbox
Onboarding
Compliance
Production CSID
Clearance
Reporting
Retries
Submission History
```

Commit:

```text
feat(zatca): integrate FATOORA workflows
```

---

# 98. PHASE 27 — DOCUMENTS

Files:

```text
database/migrations/000240_documents.sql
internal/documents/
internal/platform/storage/
apps/web/features/documents/
```

Commit:

```text
feat(documents): implement secure document storage
```

---

# 99. PHASE 28 — NOTIFICATIONS

Files:

```text
database/migrations/000250_notifications.sql
internal/notifications/
internal/platform/mail/
apps/web/features/notifications/
```

Commit:

```text
feat(notifications): implement ERP notification services
```

---

# 100. PHASE 29 — REPORTING

Files:

```text
database/migrations/000270_reporting.sql
database/queries/reporting/
internal/reporting/
apps/web/features/reports/
```

Implement:

- Trial Balance
- General Ledger
- Balance Sheet
- Income Statement
- Cash Flow
- AR Aging
- AP Aging
- Customer Statement
- Supplier Statement
- Inventory Valuation
- Sales
- Purchasing
- VAT
- Payroll
- GOSI
- Budget vs Actual

Commit:

```text
feat(reporting): implement ERP reporting platform
```

---

# 101. PHASE 30 — ERP DASHBOARD

Files:

```text
apps/web/app/[locale]/(erp)/dashboard/
apps/web/features/dashboard/
apps/web/components/charts/
```

Dashboard:

```text
Revenue
Expenses
Net Profit
AR
AP
Cash Position
Stock Warnings
Pending Approvals
ZATCA Status
Recent Documents
Activity Feed
```

Commit:

```text
feat(ui): implement bilingual ERP dashboard
```

---

# 102. PHASE 31 — UI HARDENING

Verify every key screen in:

```text
Arabic RTL Light
Arabic RTL Dark
English LTR Light
English LTR Dark
```

Commit:

```text
feat(ui): finalize bilingual RTL LTR ERP design system
```

---

# 103. PHASE 32 — SECURITY HARDENING

Perform:

```text
OWASP review
authorization matrix tests
IDOR tests
CSRF tests
session security tests
file security tests
rate limiting
secret scanning
dependency vulnerability scanning
SQL injection review
XSS review
```

Commit:

```text
security: harden ERP production boundaries
```

---

# 104. PHASE 33 — PRODUCTION OPERATIONS

Files:

```text
deployments/docker/
deployments/compose/
deployments/nginx/

docs/operations/
docs/backup/
docs/restore/
docs/upgrade/
```

Implement:

```text
production images
Compose
Nginx
TLS documentation
backup
PITR guidance
restore
upgrades
system doctor
```

Commit:

```text
chore(ops): complete ERP production operations
```

---

# 105. PHASE 34 — FINAL PRODUCTION READINESS

Before version 1:

```text
go test ./...
frontend lint passes
frontend typecheck passes
frontend build passes
Playwright passes

Arabic passes
English passes
RTL passes
LTR passes
Light passes
Dark passes

ZATCA tests pass
GOSI tests pass
security tests pass

backup restore tested
migration upgrade tested

OpenAPI current
documentation current

no real secrets
no ignored financial failures
```

Commit:

```text
release: prepare ERP v1.0.0
```

Tag:

```text
v1.0.0
```

---

# 106. GIT POLICY

Commit after every complete phase or meaningful independent feature.

Examples:

```text
feat(accounting): implement general ledger foundation
feat(accounting): add transactional posting engine
feat(inventory): implement stock movement ledger
feat(sales): implement sales invoice posting
feat(payroll): implement payroll calculation engine
feat(zatca): integrate clearance and reporting
test(accounting): add ledger invariants
security(authz): enforce legal entity scopes
```

Never leave the ERP as one giant commit.

---

# 107. DEFINITION OF DONE

A feature is complete only when applicable work includes:

- Database migration
- Constraints
- Indexes
- sqlc queries
- Domain model
- Business rules
- Interfaces
- Application service
- PostgreSQL adapter
- Authorization
- Audit
- API
- OpenAPI
- Frontend
- Arabic
- English
- RTL
- LTR
- Light
- Dark
- Unit tests
- Integration tests
- E2E tests
- Security review
- Documentation
- Git commit

---

# 108. DATABASE INVARIANTS

Automated tests must prove:

1. foreign keys work
2. cross-entity references are rejected
3. duplicate document numbers are rejected
4. journals balance
5. posted journals are immutable
6. stock movements are immutable
7. duplicate stock source is blocked
8. historical VAT is preserved
9. GOSI rule version is preserved
10. payroll calculation is preserved
11. ZATCA sequence is unique
12. audit is immutable
13. optimistic locking works
14. number sequences are concurrency safe
15. allocations cannot over-settle
16. financial posting is atomic

---

# 109. FINAL FINANCIAL INVARIANTS

Never violate:

- Every posted journal balances.
- Posted journals are immutable.
- Inventory is movement-led.
- Posted stock movements are immutable.
- Posted payroll is immutable.
- Corrections use reversal.
- Fiscal-period rules are enforced.
- Money uses exact decimal arithmetic.
- Core financial posting is transactional.
- Retryable financial operations are idempotent.
- Historical regulatory rules remain reproducible.

---

# 110. FINAL SECURITY INVARIANTS

- Authorization is always server-side.
- Every financial request is legal-entity scoped.
- Sensitive HR data uses explicit permission checks.
- Session secrets never reach logs.
- Private keys never reach source control.
- Audit cannot be modified through normal application access.
- External APIs never control internal accounting truth.
- Frontend visibility does not equal authorization.
- All database writes are parameterized.
- Production app database role is least privileged.

---

# 111. FINAL UI INVARIANTS

- Arabic is first-class.
- English is first-class.
- RTL is first-class.
- LTR is first-class.
- Dark theme is first-class.
- Light theme is first-class.
- No feature is complete if it only works correctly in English/LTR.
- No feature is complete if Dark Mode is visually broken.
- No feature is complete if Arabic layout overflows or misaligns.
- The UI must remain usable on desktop and tablet.
- Critical workflows must support mobile-friendly responsive behavior.

---

# 112. CODING AGENT BEHAVIOR

The Coding Agent MUST:

- inspect before editing
- avoid unnecessary rewrites
- preserve architecture
- write tests with implementation
- run tests after meaningful changes
- fix failures before continuing
- commit after each completed phase
- use meaningful commit messages
- write migrations before repository implementation
- avoid speculative dependencies
- avoid premature optimization
- prefer explicit behavior over magic
- document important architectural decisions

---

# 113. WHEN A TECHNICAL PROBLEM OCCURS

Do NOT create an unsafe workaround.

Instead:

1. identify the problem
2. determine whether it conflicts with this specification
3. preserve financial/security invariants
4. implement the safest standards-compliant solution
5. add an ADR if architecture changes materially
6. add tests
7. document the decision

---

# 114. FINAL ARCHITECTURE

```text
                    ┌──────────────────────────────┐
                    │       Next.js ERP UI         │
                    │                              │
                    │ Arabic / English             │
                    │ RTL / LTR                    │
                    │ Light / Dark                 │
                    │                              │
                    │ shadcn + Tailwind            │
                    └──────────────┬───────────────┘
                                   │ REST/JSON
                                   ▼
                    ┌──────────────────────────────┐
                    │          Go ERP API          │
                    │                              │
                    │ Modular Monolith             │
                    │                              │
                    │ IAM                          │
                    │ Organization                 │
                    │ Workflow                     │
                    │ Accounting                   │
                    │ Tax                          │
                    │ Sales / AR                   │
                    │ Purchasing / AP              │
                    │ Inventory                    │
                    │ Treasury                     │
                    │ HR / Payroll                 │
                    │ GOSI                         │
                    │ ZATCA                        │
                    │ Reporting                    │
                    └──────────────┬───────────────┘
                                   │
                                   ▼
                    ┌──────────────────────────────┐
                    │        PostgreSQL 18         │
                    │                              │
                    │ pgx + sqlc                   │
                    │                              │
                    │ Transactional Truth          │
                    └──────────────┬───────────────┘
                                   │
                                   ▼
                    ┌──────────────────────────────┐
                    │        Go ERP Worker         │
                    │            River             │
                    └──────────────┬───────────────┘
                                   │
                       ┌───────────┼────────────┐
                       │           │            │
                       ▼           ▼            ▼
                     ZATCA        GOSI      Integrations
```

---

# 115. FINAL SUCCESS CRITERIA

The system is ready when an independent Saudi customer deployment can reliably perform:

- organization setup
- user and permission management
- Arabic and English operation
- RTL and LTR operation
- Light and Dark operation
- customer management
- supplier management
- product management
- sales
- purchasing
- inventory
- accounts receivable
- accounts payable
- general ledger
- VAT
- ZATCA
- employee management
- attendance
- leave
- payroll
- GOSI
- treasury
- banking
- fixed assets
- workflow approvals
- audit
- financial reports
- backups
- restore
- upgrades
- production monitoring

while maintaining:

- financial integrity
- data integrity
- security
- auditability
- Saudi regulatory adaptability
- operational simplicity
- maintainability
- testability

---

# 116. FINAL DIRECTIVE

Correctness outranks cleverness.

Security outranks convenience.

Auditability outranks shortcuts.

Explicit architecture outranks framework magic.

Financial integrity outranks development speed.

The Coding Agent MUST build the ERP as a real long-lived commercial financial system.

# END OF MASTER SPECIFICATION
