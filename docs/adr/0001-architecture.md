# ADR-0001: DDD-inspired layered architecture

## Status
Accepted

## Context
Project requires long-term maintainability and explicit business rules around
transactions, splits, imports, FX conversion, and budgeting.

## Decision
Adopt a DDD-inspired architecture with explicit use-case layer and repository ports.

## Consequences
- Better separation of concerns
- Easier testing at use-case level
- Slightly higher upfront structure cost
