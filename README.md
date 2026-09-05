# Sphere Bun Layout

`sphere-bun-layout` is the Sphere project template for applications that use
Bun instead of Ent. It keeps the same Proto-first HTTP, Wire, Swagger, Docker,
and Makefile contracts as the other official layouts while limiting the example
domain to a small authenticated admin API.

## Capabilities

- Protobuf and Buf API contracts with generated Gin-compatible HTTP handlers.
- Bun models and SQLite through `sqliteshim`.
- JWT-protected admin CRUD example.
- Wire dependency injection and Swagger/OpenAPI generation.
- Docker and multi-architecture build targets.

## Workflow

```shell
make init
make run
```

During development use `make gen/all`, `make check`, and `make build`. Run
`make help` for the exact targets supported by this layout. Deployment is not a
layout capability; connect the generated image to the project's own delivery
system.

## Structure and Ownership

- `proto/**` contains handwritten API and Bun model contracts.
- `api/**` and `swagger/**` are generated.
- `internal/service/**` contains handler implementations.
- `internal/biz/**` contains initialization and business tasks.
- `internal/pkg/database` owns Bun connection setup.
- `cmd/app` composes the application through Wire.

Read `.sphere/layout.json` and `AGENTS.md` before extending or synchronizing the
layout. Unclassified paths are project-owned by default.
