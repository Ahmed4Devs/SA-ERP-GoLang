# Saudi Enterprise ERP Platform (v1.0.0)
### منصة تخطيط الموارد المؤسسية للشركات السعودية

[![Backend](https://img.shields.io/badge/Backend-Go%201.27+-00ADD8.svg?style=flat-square&logo=go)](https://go.dev/)
[![Database](https://img.shields.io/badge/Database-PostgreSQL%2018+-336791.svg?style=flat-square&logo=postgresql)](https://www.postgresql.org/)
[![Frontend](https://img.shields.io/badge/Frontend-Next.js%2016%20%7C%20TypeScript-black.svg?style=flat-square&logo=next.js)](https://nextjs.org/)
[![ZATCA](https://img.shields.io/badge/ZATCA-Integration%20Requires%20Certification-orange.svg?style=flat-square)](#complete-functional-modules-catalog)
[![Saudi Labor](https://img.shields.io/badge/HR-Qiwa%20%7C%20Mudad%20WPS-green.svg?style=flat-square)](#saudi-labor-law--payroll)
[![GOSI](https://img.shields.io/badge/GOSI-Social%20Insurance%20Engine-blue.svg?style=flat-square)](#gosi-social-insurance)
[![Release](https://img.shields.io/badge/Release-v1.0.0-gold.svg?style=flat-square)](#release-v100)

An Enterprise Resource Planning (ERP) platform for Saudi organizations. It provides production deployment controls and broad business modules, but regulatory use—especially ZATCA, WPS/Qiwa, GOSI, retention, and PDPL obligations—still requires customer-specific configuration, testing, security review, and approval by the relevant authority or qualified adviser.

---

## 🏛️ System Architecture

```text
                    ┌─────────────────────────────────────────┐
                    │        Next.js 16 Standalone UI         │
                    │                                         │
                    │   Arabic / English (Bilingual RTL/LTR)  │
                    │   Dark / Light Theme (Design System)    │
                    │   Tailwind CSS + Lucide Icons           │
                    └────────────────────┬────────────────────┘
                                         │ HTTPS / Reverse Proxy
                                         ▼
                    ┌─────────────────────────────────────────┐
                    │               Nginx Edge                │
                    │   Rate Limiting & Security Headers      │
                    └────────────────────┬────────────────────┘
                                         │ REST API
                                         ▼
                    ┌─────────────────────────────────────────┐
                    │               Go Core API               │
                    │                                         │
                    │   Modular Monolith (Hexagonal Arch)     │
                    │   Tenant Boundary & IDOR Isolation      │
                    │   Token-Bucket Rate Limiter & CSRF      │
                    └────────────┬──────────────────┬─────────┘
                                 │                  │
                                 ▼
                    ┌─────────────────────────────────────────┐
                    │              PostgreSQL 18              │
                    │ pgx + Goose migrations + exact NUMERIC │
                    └─────────────────────────────────────────┘
```

---

## 📦 Complete Functional Modules Catalog

| # | Module | Package Path | Regulatory / Business Function |
| :-: | :--- | :--- | :--- |
| **1** | **IAM & Auth** | `internal/iam` | Secure authentication, opaque session management, Argon2id hashing. MFA is not implemented. |
| **2** | **Organization** | `internal/organization` | Multi-tenant legal entity boundaries, branches, cost centers, and CR numbers. |
| **3** | **Authorization** | `internal/authorization` | Role-Based Access Control (RBAC) with granular functional permissions. |
| **4** | **Audit Trail** | `internal/audit` | Append-only tamper-evident audit logging of all data mutations. |
| **5** | **Workflows** | `internal/workflow` | Multi-tier approval workflows with state machine validation. |
| **6** | **General Ledger** | `internal/accounting` | Double-entry bookkeeping, Chart of Accounts, fiscal calendars, immutable journals. |
| **7** | **Tax & VAT** | `internal/tax` | Saudi standard 15% VAT, zero-rated, exempt categories, and ZATCA VAT Return (Boxes 1–16). |
| **8** | **ZATCA Integration** | `internal/zatca` | UBL XML generation, ECDSA signing, TLV QR codes, clearance/reporting client and simulator. Formal Phase 2 conformance is not certified. |
| **9** | **Products Master** | `internal/products` | Categories, UOMs, SKUs, barcode tracking, cost pricing. |
| **10**| **Inventory Ledger** | `internal/inventory` | Movement-led inventory ledger, multi-warehouse bins, transfers, perpetual stock valuation. |
| **11**| **Sales & AR** | `internal/sales` & `internal/receivables` | Order-to-cash lifecycle, credit limits, payment terms, debit notes, open item tracking. |
| **12**| **Purchasing & AP**| `internal/purchasing` & `internal/payables` | Procure-to-pay lifecycle, RFQs, purchase orders, 3-way matching, supplier aging. |
| **13**| **Treasury** | `internal/treasury` | Cash accounts, petty cash disbursements, liquidity forecasting. |
| **14**| **Banking** | `internal/banking` | Multi-currency bank accounts, statement import, automated reconciliation engine. |
| **15**| **HR & Qiwa** | `internal/hr` | Employee master, electronic Qiwa contracts, Saudization / Nitaqat compliance. |
| **16**| **Attendance** | `internal/attendance` | Shift schedules, biometric rosters, automated overtime computation. |
| **17**| **Leave Ledger** | `internal/leave` | Saudi Labor Law statutory leaves (Annual, Sick, Hajj, Maternity, Paternity, Bereavement). |
| **18**| **Payroll & WPS** | `internal/payroll` | Deterministic salary computation, deductions, end-of-service (EOS), Mudad SIF file generator. |
| **19**| **GOSI** | `internal/gosi` | Versioned social insurance engine (Annuity 18%, OHRSI 2%, SANED unemployment 1.5%). |
| **20**| **Fixed Assets** | `internal/fixedassets` | Asset lifecycle, tracking, monthly straight-line depreciation schedules, disposals. |
| **21**| **Budgeting** | `internal/budgeting` | Multi-dimensional operating/capital budgets, revisions, budget vs actual variance. |
| **22**| **Documents** | `internal/documents` | Secure multi-tenant blob storage, document versions, entity attachments. |
| **23**| **Notifications** | `internal/notifications` | In-app alerts, email dispatch, alert preference center. |
| **24**| **Reporting & BI**| `internal/reporting` | Financial statements (Balance Sheet, Income Statement, Trial Balance) and commercial reports. |
| **25**| **Executive Dashboard**| `internal/dashboard` | Unified C-suite command center aggregating live KPIs across all enterprise domains. |
| **26**| **Security Kernel**| `internal/platform/middleware` | Token-bucket rate limiting, CSRF validation, body payload bounds, IDOR tenant guard. |
| **27**| **Operations Stack**| `deployments/` | Production Docker images, Compose orchestration, Nginx reverse proxy, System Doctor. |

---

## 🔒 Financial & Regulatory Invariants

1. **Exact Decimal Arithmetic**: Monetary quantities use PostgreSQL `NUMERIC` columns with precision selected per domain and decimal arithmetic in Go.
2. **Double-Entry Balance Guarantee**: Every journal entry transaction requires $\sum \text{Debits} = \sum \text{Credits}$. Imbalanced entries are rejected at the database constraint level.
3. **Immutability of Historical Truth**: Posted journal entries, finalized stock movements, and cleared ZATCA invoices are immutable; corrections must execute through reversing transactions.
4. **Tenant & Legal Entity Isolation**: Every operational request is enforced server-side against the user's active tenant and legal entity boundary.
5. **Saudi Sovereign Timezone**: All legal, financial cutoffs, and e-invoicing timestamps are stamped in `Asia/Riyadh` (AST, UTC+3).

---

## 🚀 Getting Started

### Local Development Setup

#### Prerequisites
- **Go**: >= 1.27
- **Node.js**: >= 20 & npm
- **Docker & Docker Compose**
- **PostgreSQL**: 18+

#### Running the Backend
```bash
# 1. Run migrations
go run github.com/pressly/goose/v3/cmd/goose -dir database/migrations postgres "postgres://erp_user:erp_pass@localhost:5432/erp_db?sslmode=disable" up

# 2. Start Core API Server
go run ./cmd/erp-api
# Listening on :8080 (Health: http://localhost:8080/healthz)
```

#### Running the Frontend Web
```bash
cd apps/web
npm install
npm run dev
# Accessible at http://localhost:3000
```

---

## 🐳 Production Deployment (Docker Compose)

The production deployment orchestrates all services in isolated containers with automated database migrations, healthchecks, and Nginx edge routing:

```bash
# 1. Enter Compose directory
cd deployments/compose

# 2. Configure Production Secrets
cp ../../.env.production.example .env.production
# Edit .env.production and provide all required credentials and service URLs

# 3. Launch Full Production Stack
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build

# 4. Verify Stack Status
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

---

## 🩺 Operational Tooling & Diagnostic Utilities

### 1. System Doctor
Validates system dependencies, migrations, frontend assets, timezone, and live API connectivity:
```bash
# On Linux / macOS / Docker:
./scripts/doctor.sh

# On Windows (PowerShell):
pwsh -File ./scripts/doctor.ps1
```

### 2. Automated Production Backup
Performs encrypted PostgreSQL backups with SHA-256 checksums and automated retention pruning:
```bash
docker compose --env-file deployments/compose/.env.production \
  -f deployments/compose/docker-compose.prod.yml \
  --profile operations run --rm erp-backup
```

### 3. Disaster Recovery Restore
Performs a checksum-verified, encrypted logical-dump restoration. The checked-in
Compose baseline does not provide WAL archiving or point-in-time recovery:
```bash
BACKUP_ENCRYPTION_KEY='read-from-your-secret-manager' \
  ./scripts/restore.sh /backups/erp_erp_db_20260918_020000.dump.enc --force
```

---

## 🧪 Testing & Quality Assurance

Run the comprehensive test suite across all modules:
```bash
# Run all Go backend integration & domain test suites
go test -timeout 900s ./tests/...

# Frontend static type check and Next.js production build
cd apps/web
npm run type-check
npm run build
```

---

## 📄 API Documentation

The generated REST API specification is maintained in OpenAPI 3.1 format at:
- File: [`api/openapi/openapi.yaml`](api/openapi/openapi.yaml)
- Production Endpoint: `https://erp.example.sa/api/v1`

---

## 📜 License & Compliance
Commercial Proprietary. Designed for sovereign deployment within the Kingdom of Saudi Arabia. Deployment does not by itself establish regulatory compliance; operators must complete legal, security, privacy, and authority acceptance reviews.
