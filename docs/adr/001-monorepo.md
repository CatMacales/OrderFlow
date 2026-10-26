# ADR 001: Keep OrderFlow services in a monorepo

[English](001-monorepo.md) | [Русский](001-monorepo.ru.md)

Status: Accepted. Date: 2026-10-26.

## Context

OrderFlow starts with one order service and targets a system of cooperating services. The repository currently contains one executable. Repository organization must support future services without adding unnecessary maintenance now.

## Alternatives

- Separate repository per service: independent access and repository configuration, but more coordination and repeated maintenance for this small project.
- Monorepo: related changes, documentation, and tooling stay together, while service boundaries require deliberate enforcement.

## Decision

Keep the order service and future services in the OrderFlow repository. Start with the order service only; add other services when concrete use cases require them. The order service owns orders, line quantities, and lifecycle rules; product and customer IDs refer to externally owned data.

Keep the current code layout for now. A future `services/orders`, `services/catalog`, and `services/customers` layout is illustrative, not a commitment to implement those services. Choose exact directories, Go module boundaries, and workspace configuration when introducing another service.

Services must remain independently buildable, testable, and deployable. They communicate through explicit API or event contracts rather than importing each other's internal business logic. Shared business models are not introduced just to reuse code.

## Consequences

Related changes can be reviewed together and repository tooling can be reused. A shared repository does not make coordinated deployment mandatory or remove the need for compatible contracts. Service-specific CI and dependency boundaries will need attention as the repository grows.

This replaces the original modular-monolith starting direction with an order service in a future microservice system. It does not add services, relocate code, connect infrastructure, or select Go module boundaries now. External ID existence checks remain outside the first version.
