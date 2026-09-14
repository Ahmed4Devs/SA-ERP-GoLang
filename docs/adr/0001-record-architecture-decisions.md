# ADR 0001: Record Architecture Decisions

## Status
Accepted

## Context
We need a standard mechanism to document significant architectural, financial, and infrastructure decisions as the ERP platform evolves.

## Decision
We will use Architecture Decision Records (ADRs) as documented by Michael Nygard.
Each ADR will be stored as a markdown file in `docs/adr/` with the naming scheme `NNNN-short-title.md`.

Each record contains:
- Title and Number
- Status (Proposed, Accepted, Rejected, Deprecated, Superseded)
- Context
- Decision
- Consequences
- Compliance & Invariants Impact

## Consequences
All core team members and AI coding agents must document non-trivial architecture modifications or deviations using an ADR before committing breaking changes.
