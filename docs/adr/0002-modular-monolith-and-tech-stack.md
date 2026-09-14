# ADR 0002: Modular Monolith and Tech Stack

## Status
Accepted

## Context
Building an enterprise ERP system requires high consistency, strict transaction management, and low operational complexity for on-premise and dedicated private cloud deployments.

## Decision
1. **Architecture:** Modular Monolith in Go instead of microservices.
2. **Backend:** Go 1.27+, net/http with Chi router, pgxpool, and sqlc.
3. **Database:** PostgreSQL 18 with relational integrity constraints and schema partitioning per domain.
4. **Queue Worker:** River (PostgreSQL-backed transactional queue) to ensure transactional event publishing and processing without distributed transactions.

## Consequences
- Simplifies atomic multi-table operations across modules (e.g., invoice posting generating general ledger lines and stock movements in a single transaction).
- Reduces infrastructure requirements (no Kafka or microservice mesh required).
- Preserves clean boundaries via Go package structure and domain interfaces.
