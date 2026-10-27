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

# Git workflow

- Keep `main` buildable and tested. Use short-lived task branches such as `codex/of-202-memory-storage`; do not introduce a long-lived `develop` branch.
- Make each commit a coherent change, including its relevant tests and documentation. Commit generated code with its source and dependency changes with `go.mod`/`go.sum`; keep `.project-notes/` local.
- Use Conventional Commits in English: `type(scope): short action`, with scope optional. Use `feat`, `fix`, `refactor`, `test`, `docs`, `build`, `ci`, or `chore` as appropriate.
- Keep commit messages short: normally one subject line, aiming for 50 characters and no more than 72. Add a brief body only when necessary to explain a non-obvious reason; task references are optional.
- Before handing off a task, run formatting, build, vet, race-enabled tests, and `git diff --check`. Run generation when its inputs change and review the generated diff. Inspect the staged diff before committing.
- Prefer squash merging a completed task branch; retain separate commits when they represent independently useful changes. Creating a PR or merging also requires explicit user authorization.

# Local working notes

- If `.project-notes/` exists, read its roadmap and progress notes before continuing stage work. These files are local context, not project specifications.
- Keep discussion history and handoff notes there. Keep requirements and lasting architectural decisions in `docs/`.
- Do not mark a stage complete until its acceptance criteria are met. Keep proposed requirements separate from agreed decisions.
