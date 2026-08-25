# flagcast

A feature-flag & experimentation platform — event-driven Go microservices, a
low-latency gRPC evaluation SDK, a React/TypeScript console, and Claude-powered
experiment analysis with an MCP server.

> Sibling project to [sitemon](https://github.com/mralaminahamed/sitemon).
> flagcast deliberately covers what sitemon left out: a real gRPC surface, a
> cloud deploy, and distributed tracing.

## What it does

- Define flags, segments, and percentage rollouts in a web console.
- Services evaluate flags through a **gRPC SDK** with a Redis-backed local cache.
- Flag changes stream over NATS to every connected SDK in near-real-time.
- Mongo stores flag config plus an audit log; AI analyzes experiment results.

## Stack

- Go 1.27, single module `github.com/mralaminahamed/flagcast`
- gRPC (evaluation + streaming), REST (admin), NATS (change events)
- MongoDB (config + audit), Redis (eval cache), Prometheus + Grafana
- React 19 + TypeScript (Vite) console — later phase
- Anthropic SDK + MCP server — later phase

## Layout

```
apps/{gateway,evaluator,streamer,ai,console}   # console = web
packages/shared/   # store cache bus proto models config logger health metrics auth
infra/             # docker-compose, mongo, prometheus, grafana, k8s
```

## Quick start

```bash
cp .env.example .env
make up                       # mongo, redis, nats, gateway
curl localhost:8080/health
```

## Roadmap (phased, PR per phase)

- **Phase 0** — monorepo skeleton, backing stores, gateway health stub, CI ✅
- **Phase 1** — flag domain + REST admin CRUD (gateway + Mongo)
- **Phase 2** — gRPC evaluator + proto + Go SDK client
- **Phase 3** — Redis eval cache + NATS change stream → streaming SDK updates
- **Phase 4** — React console
- **Phase 5** — AI experiment analysis + MCP server
- **Phase 6** — observability (metrics + tracing) + cloud IaC + CD
```
