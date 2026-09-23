# ADR 0004: PostgreSQL Transactional Outbox for Background Delivery

## Status
Accepted; supersedes the queue worker choice in ADR 0002.

## Context
ADR 0002 specified River for background jobs. The implemented notification and ZATCA workers use the existing `platform.outbox_events` table and a PostgreSQL-backed dispatcher instead. The architecture record must match the operational behavior and retry guarantees.

## Decision
Use the transactional outbox for durable event delivery. Producers insert events in their business transaction. Worker replicas lease rows with `FOR UPDATE SKIP LOCKED`, dispatch registered handlers, and record success, retry with backoff, or dead-letter status. Delivery is at least once; handlers must tolerate retries. ZATCA submission retries remain a separate worker loop.

## Alternatives Considered
- **River jobs:** Provides a mature job framework but adds a runtime dependency and another job schema. It remains an option if scheduling and queue needs outgrow this dispatcher.
- **External broker:** Adds infrastructure and makes atomic publishing with PostgreSQL harder.

## Consequences
The worker and database own queue operations without another service. A worker crash after an external side effect and before success recording can repeat delivery. Operations must monitor pending and dead-letter events.

## Migration Impact
The existing `platform.outbox_events` schema and `000310_outbox_dead_letters.sql` remain the persistence contract. No database migration is needed for this documentation correction.

## Security Impact
Outbox payloads and error text reside in PostgreSQL; producers and handlers must avoid putting secrets in them. Worker database credentials need only the privileges required to claim and update events.

## Testing Impact
Keep transaction rollback, concurrent claim, retry, dead-letter, and handler idempotency coverage. The integration suite must run against an isolated PostgreSQL database.
