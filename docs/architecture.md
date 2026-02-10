# Architecture

## Principles
- DDD-inspired layered architecture
- Use-case driven application layer
- Domain logic isolated from transport and persistence

## Backend layering
- `domain/`: entities, value objects, domain services
- `application/`: use cases, commands/queries, ports
- `infrastructure/`: sqlite repositories, external clients (ECB), CSV parsers
- `interfaces/http/`: handlers, DTO mapping, middleware
- `platform/`: config, logging, clock, tx manager
