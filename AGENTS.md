# Engineering guidelines

- Write idiomatic Go and use the standard library when it is sufficient.
- Keep the architecture simple. Avoid overengineering and speculative layers.
- Use a monorepo, starting with the order service only (docs/adr/001-monorepo.md). Add other services only for concrete use cases; do not create them in advance.
- Handle errors explicitly. Add useful context when wrapping errors and avoid logging the same error at every layer.
- Propagate context.Context across request and I/O boundaries. Accept it as the first parameter; do not store it in structs.
- Inject dependencies through constructors. Avoid mutable global dependencies.
- Use table-driven tests when multiple cases benefit from shared test structure.
- Introduce interfaces at the consumer boundary only when there is a concrete need. Avoid unnecessary interfaces and abstractions.
- Justify new libraries before adding them. Prefer fewer dependencies.
- Write documentation and commit messages in English. Maintain README.ru.md as the Russian translation of README.md and keep both versions aligned when changing project documentation.
- Before significant architectural changes, explain the trade-offs and record lasting decisions in docs/adr/.
- PostgreSQL, SQLC, Goose, gRPC, Kafka, and Transactional Outbox are planned. Add them incrementally when a concrete use case requires them.
- Do not connect databases, brokers, or other infrastructure during foundation work.
- Run formatting, build, vet, and race-enabled tests before handing off changes. Report checks that could not run.
- Do not commit, push, or create pull requests without explicit user authorization.

# Local working notes

- If `.project-notes/` exists, read its roadmap and progress notes before continuing stage work. These files are local context, not project specifications.
- Keep discussion history and handoff notes there. Keep requirements and lasting architectural decisions in `docs/`.
- Do not mark a stage complete until its acceptance criteria are met. Keep proposed requirements separate from agreed decisions.
