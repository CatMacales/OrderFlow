# OrderFlow

[English](README.md) | [Русский](README.ru.md)

OrderFlow is a production-oriented order processing project intended to demonstrate Middle+/Senior Go backend engineering skills. It starts with an order service; future services will share a monorepo. See [ADR 001](docs/adr/001-monorepo.md).

The project contains one executable, structured JSON logging, development commands, CI, and a tested order domain. The domain validates orders and product lines and implements confirmation, completion, and cancellation, including repeated actions. Application operations and storage are not implemented yet; there is no network server, database, or broker. The executable still logs startup and exits successfully.

## Goals

- Build a clear, maintainable order processing system with explicit boundaries.
- Explore transactional consistency, reliable event delivery, and idempotent processing.
- Demonstrate testing, observability, and operational practices as features are introduced.
- Extract services only if concrete requirements justify the added complexity.

## Planned technology stack

| Technology | Intended role | Status |
| --- | --- | --- |
| Go and log/slog | Application and structured logging | Foundation |
| PostgreSQL | Transactional persistence | Planned |
| SQLC | Type-safe code generated from SQL | Planned |
| Goose | Database migrations | Planned |
| gRPC / Protocol Buffers | Typed API contracts | Planned |
| Kafka | Asynchronous event transport | Planned |
| Transactional Outbox | Atomic persistence of state and pending events | Planned |

## Quick Start

Prerequisites: Go 1.27.3 or later, Git, and Make. CI uses the version declared in `go.mod`. Race-enabled tests also require CGO and a C compiler (for example, clang on macOS or GCC on Linux).

```sh
git clone https://github.com/CatMacales/OrderFlow.git
cd OrderFlow
make run
```

Expected output is a JSON log with `msg` set to `OrderFlow foundation initialized` and `service` set to `orderflow`. The process then exits with status 0.

```sh
make build       # Build bin/orderflow
make test        # Run tests
make test-race   # Run tests with the race detector
make fmt         # Format Go packages
make vet         # Run static analysis
make generate    # Run code generators
```

If Go is installed outside PATH, pass its executable explicitly, for example `make GO=/path/to/go build`.

No environment variables or external services are required, so an `.env.example` is not needed yet. Domain tests cover validation, all 12 state/action pairs, unchanged order data, and item-slice ownership. CI runs formatting, build, vet, and race-enabled tests.

Order IDs use `github.com/google/uuid`; the domain accepts non-nil UUID v4 values with the RFC 4122 variant. `quantity` uses `int64` and must be positive. External customer and product IDs are preserved exactly. Constructors validate data; state changes go through order methods. `Items()` returns a copy.

Statuses are generated with the pinned `go-enum` tool from `internal/order/domain/status.go`. After changing its enum declaration, run:

```sh
make generate
```

Keep `status_enum.go` in the repository and do not edit it manually. Ordinary builds use this generated file without running the generator. SQL integration flags are not enabled; future persistence conversion belongs to the storage adapter.

## Project layout

```text
cmd/orderflow/       Executable entry point
internal/           Private application packages, added as needed
internal/order/domain/ Order model, line validation, lifecycle, and tests
api/proto/          Future Protocol Buffers contracts
migrations/         Future SQL migrations
docs/adr/           Architecture decision records
deploy/             Future deployment configuration
.github/workflows/  Continuous integration
```

Reserved directories contain only `.gitkeep`; there are no speculative architecture layers.

## Roadmap

Order behavior: [requirements](docs/order-requirements.md) and [acceptance examples](docs/order-acceptance.md). Domain rules are implemented and tested; the full application scenarios, in-memory storage, and executable demonstration remain pending in stage 2.

1. Define the first order use case, invariants, lifecycle, and module boundaries; document key trade-offs in an ADR.
2. Implement and test a minimal vertical slice with explicit dependency injection and context propagation.
3. Add PostgreSQL persistence, Goose migrations, and SQLC queries with integration tests.
4. Introduce a gRPC API, request validation, deadlines, and graceful shutdown.
5. Add a Transactional Outbox and Kafka publishing with retries and idempotent consumers.
6. Extend observability, failure testing, and deployment documentation.

Each stage should deliver working behavior before introducing further infrastructure or abstractions.
