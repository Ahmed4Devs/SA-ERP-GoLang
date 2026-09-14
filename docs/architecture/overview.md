# System Architecture Blueprint

## 1. High-Level Architecture

The ERP system is constructed as a **Modular Monolith** using Go for the high-performance backend, PostgreSQL 18 for transactional persistence, and Next.js for a bilingual, RTL-first modern web presentation layer.

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
      (App Router, RTL)          Modular Monolith
                                        │
                   ┌────────────────────┼────────────────────┐
                   │                    │                    │
                   ▼                    ▼                    ▼
              PostgreSQL           Go Worker         Integrations
             (Multi-Schema)          River           ZATCA / GOSI
```

---

## 2. Deployment & Isolation Model

- **Tenant Isolation:** Dedicated installation per customer deployment. No noisy-neighbor issues or shared multi-tenant database risks at this layer.
- **Financial Boundary:** Within a single customer deployment, `legal_entity_id` is the primary financial and security isolation boundary. Composite foreign keys enforce multi-entity safety at the database engine level.

---

## 3. Technology Stack

- **Backend:** Go 1.27+, net/http with Chi router, pgx/v5 connection pool, sqlc compiled type-safe SQL, Goose migrations, River queue worker, log/slog structured logging.
- **Database:** PostgreSQL >= 18 with `pg_trgm`, `uuidv7()`, and strict schema separation.
- **Frontend:** Next.js 15+ (App Router), React 19, TypeScript, Tailwind CSS, Lucide icons, next-themes.
- **Localization:** Arabic (`ar-SA`) & English (`en`), first-class RTL & LTR styling, light/dark/system themes.
- **Compliance:** Built-in adapters for ZATCA E-Invoicing Phase 2 (Clearance/Reporting) and GOSI Social Insurance calculations.

---

## 4. Module & Hexagonal Structure

Each business domain resides in its own isolated package under `internal/<domain>/`:

```text
internal/<domain>/
├── domain/            # Entities, value objects, domain rules, errors
├── application/       # Command/Query handlers, use-cases, DTOs
├── ports/             # Repository interfaces and service contracts
├── infrastructure/    # PostgreSQL adapters, external clients
├── transport/         # HTTP handlers and routers
└── module.go          # Dependency injection wireup
```
