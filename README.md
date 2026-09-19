# Sphere Bun Layout

`sphere-bun-layout` is the Sphere project template for applications that use
Bun instead of Ent. It keeps the same Proto-first HTTP, Wire, Swagger, Docker,
and Makefile contracts as the other official layouts while limiting the example
domain to a small authenticated admin API.

## Capabilities

- Protobuf and Buf API contracts with generated HTTP handlers on the stdx
  (net/http) engine.
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

## Upgrade Notes

This revision is a breaking template change: generated handlers are served by
the `stdx` (net/http) engine from `httpx` / `httpx/stdx` v0.0.5, pinned together
with `github.com/go-sphere/sphere` v0.0.6. The Gin adapter is gone.

- Generated `api/**` code and its error envelopes use the `httpz.*` types
  instead of the former `ginx.*` names.
- `internal/pkg/httpsrv` exports `NewServer(name, addr) httpx.Engine`, backed by
  `stdx` over `net/http`, and `UseCORS` registers CORS on that engine.
- Middleware is registered on the engine rather than a Gin router, so CORS and
  panic recovery also cover paths no route matched.

A project generated from an earlier revision should merge
`internal/pkg/httpsrv/**`, `internal/server/*/web.go`, and the regenerated
`api/**` outputs at its next layout sync, then run `make gen/all` and
`make check`.
