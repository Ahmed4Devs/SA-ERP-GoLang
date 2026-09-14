# ADR 0003: Bilingual and RTL-First UI Architecture

## Status
Accepted

## Context
Saudi enterprise users operate primarily in Arabic (`ar-SA`) with right-to-left (RTL) reading flow, while expatriate staff and multinational partners frequently use English (`en`) with left-to-right (LTR) flow. Many ERP applications treat Arabic as an afterthought, leading to visual clipping and broken data entry forms.

## Decision
1. **First-Class Arabic & RTL:** Arabic is treated as the primary citizen alongside English.
2. **Dynamic Direction:** The document root (`<html dir="rtl" lang="ar">` vs `<html dir="ltr" lang="en">`) dynamically toggles based on locale without requiring session re-authentication.
3. **Logical CSS Properties:** UI layouts must rely on CSS logical properties (`margin-inline-start`, `padding-inline-start`, `inset-inline-start`, etc.) to prevent inverted visual order.
4. **Theme Tokens:** Provide Light, Dark, and System themes powered by Tailwind CSS variables, ensuring consistent readability in both day and night operational environments.

## Consequences
- Every new component must be validated in both RTL and LTR modes.
- Hardcoded directional styles (`mr-4`, `pl-2`) are strictly prohibited in favor of logical variants (`ms-4`, `ps-2`).
