# Production-Grade Saudi ERP Platform

[![Architecture](https://img.shields.io/badge/Architecture-Modular%20Monolith-blue.svg)](docs/architecture/overview.md)
[![Backend](https://img.shields.io/badge/Backend-Go%201.27+-00ADD8.svg)](https://go.dev/)
[![Database](https://img.shields.io/badge/Database-PostgreSQL%2018-336791.svg)](https://www.postgresql.org/)
[![Frontend](https://img.shields.io/badge/Frontend-Next.js%20%7C%20TypeScript-black.svg)](https://nextjs.org/)
[![Compliance](https://img.shields.io/badge/Compliance-ZATCA%20%7C%20GOSI-green.svg)](#)

A commercial, production-ready Enterprise Resource Planning (ERP) platform designed specifically for medium-to-large Saudi organizations. Built strictly following the canonical architecture specification in `context/project-prompt.md`.

---

## Key Characteristics

- **Architecture:** Modular Monolith with clean Hexagonal / Ports-and-Adapters internal boundaries.
- **Backend:** Go >= 1.27, `Chi`, `pgx/v5`, `sqlc`, `Goose`, `River` job queue.
- **Frontend:** Next.js (App Router), React, TypeScript, Tailwind CSS, shadcn/ui.
- **Localization:** Bilingual (Arabic `ar-SA` + English `en`), First-class RTL and LTR, System/Light/Dark themes.
- **Compliance:** ZATCA Phase 2 E-Invoicing (Clearance & Reporting) and GOSI rules engine.
- **Financial Invariants:** Exact decimal arithmetic (`NUMERIC(24,6)`), immutable posted journals, balanced debit/credit validation, movement-led inventory ledger.

---

## Directory Layout

```text
.
├── apps/
│   └── web/                   # Next.js frontend application
├── cmd/
│   ├── erp-api/               # Main HTTP REST API service
│   ├── erp-worker/            # Background worker service (River)
│   └── erp-cli/               # Management CLI & database utilities
├── internal/
│   ├── platform/              # Platform kernel (config, health, http, db, tx)
│   └── [domain modules]/      # iam, organization, accounting, sales, inventory, etc.
├── pkg/                       # Shared utility packages (decimal, money, clock)
├── database/
│   ├── migrations/            # Goose SQL schema migrations
│   ├── queries/               # sqlc SQL queries
│   └── seeds/                 # Initial system seeds
├── docs/                      # Architecture blueprints and ADRs
├── docker-compose.yml         # Local containerized infrastructure
└── Makefile                   # Build and development commands
```

---

## Quickstart

### Prerequisites

- Go >= 1.27
- Node.js >= 20 & npm
- Docker & Docker Compose
- PostgreSQL >= 18 (or via Docker Compose)

### Starting Services

1. **Start PostgreSQL Infrastructure:**
   ```bash
   docker compose up -d postgres
   ```

2. **Run Backend API:**
   ```bash
   go run ./cmd/erp-api
   ```
   Health check: `http://localhost:8080/healthz`

3. **Run Frontend Web:**
   ```bash
   cd apps/web
   npm install
   npm run dev
   ```
   Web interface: `http://localhost:3000`

---

## License
Commercial Proprietary — All Rights Reserved.
