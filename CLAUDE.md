# CLAUDE.md

Guidance for Claude Code (and other agents) working in this repo.

## What this is

flagcast — a feature-flag & experimentation platform. Event-driven Go
microservices, a gRPC evaluation SDK, a React/TypeScript console, MongoDB +
Redis, Prometheus/Grafana, and Claude-powered experiment analysis with an MCP
server.

- Go 1.27, single module `github.com/mralaminahamed/flagcast`
- Web: React 19 + TypeScript (Vite), in `apps/console` (later phase)

## Layout

```
apps/{gateway,evaluator,streamer,ai,console}
packages/shared/   # store cache bus proto models config logger health metrics auth
infra/             # docker-compose, mongo, prometheus, grafana, k8s
```

Each `apps/<svc>/cmd` builds its own binary.

## Commands

```bash
make up         # local stack
make build      # binaries into ./bin
make test-race  # CGO_ENABLED=1 go test -race ./...
make lint       # go vet + gofmt check
```

## Verify before claiming done

```bash
gofmt -l apps packages && go vet ./... && go build ./... && CGO_ENABLED=1 go test -race ./...
```

Verify live where it matters (run the service, curl the endpoint).

## Architecture notes (target)

- **gRPC** is the evaluation path: SDK clients call the evaluator for flag
  decisions and subscribe to a server-stream of changes. REST (gateway) is the
  admin/control plane. All proto lives in `packages/shared/proto`.
- **NATS** carries flag-change events from the gateway to the streamer/evaluator.
- **Redis** caches evaluations; **Mongo** stores flag config + an audit log.
- **Graceful degradation:** the gateway runs standalone when a dep is absent.
- **Health vs readiness:** `/health` is liveness; `/ready` pings deps (503 on
  failure); `/metrics` is Prometheus. All served from `packages/shared/health`.
- **Claude:** Anthropic Go SDK, default `claude-opus-5`. Consult the `claude-api`
  guidance before touching Claude code — model ids/SDK shapes drift.

## Conventions

- **Comments:** minimal — only what's required, short, no over-explaining.
- **Git:** small single-scope Conventional Commits; branch from `trunk`; open a
  PR per change; merge with a merge commit (not squash); delete the branch.
  - Scopes: `gateway evaluator streamer ai console infra`.
- **Never commit** `.env`. `PLAN.md`/`ARCHITECTURE.md` stay untracked.
- Reuse `packages/shared` rather than duplicating logic.
