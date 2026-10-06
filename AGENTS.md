# AI Agent Guide

## Layout Profile

This layout uses Protobuf/Buf, generated HTTP handlers on stdx (net/http), Bun,
Wire, Swagger, and SQLite. It demonstrates a small JWT-protected admin CRUD
API. It does not include the dashboard, Ent, Telegram, WeChat, or deployment
scripts.

## Ownership and Extension

Read `.sphere/layout.json` before modifying files. Never hand-edit generated
paths. Treat mixed paths as three-way merge seams and assume every unclassified
path is project-owned.

Add contracts under `proto/<domain>/v1`, business logic under
`internal/biz/<domain>`, service implementations under
`internal/service/<domain>`, and Bun persistence code under
`internal/pkg/database`. Product logic must not be added to layout-owned CI,
generation, app-bootstrap, or HTTP-adapter files.

See `docs/LAYOUT_CONTRACT.md` for the complete authoring and synchronization
protocol, including legacy-project adoption and conflict handling.

## Workflow

- `make gen/all` regenerates Proto, Swagger, and Wire outputs.
- `make test` runs Go tests.
- `make lint` checks Go and Buf without rewriting files.
- `make check` verifies dependencies, formatting, lint, and tests.
- `make codegen-check` compares generated `api/**` with the tracked
  `codegen.sha256` baseline; refresh it with `make codegen-baseline` after
  Proto changes or a generator bump in `codegen.versions`, and commit it.
- `make build` builds the application.

After changing constructors, provider sets, Proto, or Bun models, regenerate
before testing. Delivery requires `make check`, `make build`, and review of all
tracked generated changes.
