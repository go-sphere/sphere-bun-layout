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

- `proto/**` contains handwritten API and Bun model contracts; the Bun models
  live in `proto/bunpb/bunpb.proto` (generated as `api/bunpb`).
- `api/**` and `swagger/**` are generated.
- `internal/service/**` contains handler implementations.
- `internal/biz/**` contains initialization and business tasks.
- `internal/pkg/database` owns Bun connection setup.
- `cmd/app` composes the application through Wire.

Read `.sphere/layout.json` and `AGENTS.md` before extending or synchronizing the
layout. Unclassified paths are project-owned by default.

## Generator Versions and Codegen Baseline

`codegen.versions` pins every tool `make install` installs: the go-sphere
protoc plugins, Buf, protoc-gen-go, Swag, Wire, golangci-lint, and sphere-cli.
Change versions only there; the Makefile and the codegen scripts both read it.

Generated `api/**` is not committed, so `codegen.sha256` records the SHA-256
of every generated `api/**` file as the tracked regression baseline:

- `make codegen-check` compares the current `api/**` with the baseline. The CI
  workflow runs it right after `make gen/all`.
- `make codegen-verify` regenerates `api/**` with the pinned plugins in a
  temporary tool directory and checks that the generated packages build, that
  generation is idempotent, and that the output matches the baseline. The
  Codegen workflow runs it on every push.
- `make codegen-baseline` rewrites the baseline from the current `api/**`.

After changing Proto files or bumping a generator, run
`make install && make gen/all && make codegen-baseline` and commit
`codegen.sha256` with the change, so reviewers see which generated files
moved. A failing check without such a change means the installed tools do not
match `codegen.versions`.

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

The process timezone is now set explicitly: `cmd/app/main.go` calls
`boot.InitTimezone(boot.DefaultTimezone)` (`Asia/Shanghai`) before startup and
exits if the zone cannot be loaded. Newer sphere releases no longer set it from
a package `init()`, so a project that keeps an older `main.go` runs in the host
timezone after upgrading sphere. Merge `cmd/app/main.go` (it is `mixed`), then
pass another IANA zone or delete the call to keep the host default.

The handwritten Bun model contract was renamed from `proto/entpb/entpb.proto`
(package `entpb`) to `proto/bunpb/bunpb.proto` (package `bunpb`); the old name
suggested Ent-generated output. This is a breaking template change: generated
Go moves from `api/entpb` to `api/bunpb`. A project synchronizing this revision
should move its model messages into `proto/bunpb`, change proto imports to
`"bunpb/bunpb.proto"` and `entpb.X` references to `bunpb.X`, update Go imports
of `.../api/entpb` to `.../api/bunpb`, delete the stale `api/entpb`, then run
`make gen/all` and `make codegen-baseline`.
