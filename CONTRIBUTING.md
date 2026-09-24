# Contributing to flagcast

Thanks for helping out. This guide covers the local setup, the checks to run
before opening a pull request, and how changes land on `trunk`.

Please read the [Code of Conduct](CODE_OF_CONDUCT.md) first. Security issues go
through the [security policy](SECURITY.md), not public issues.

## Prerequisites

- Go 1.27+
- Docker (for the local stack: MongoDB, Redis, NATS)
- Node 24 + pnpm, via `corepack enable` (only to work on `apps/console`)
- `buf`, `protoc-gen-go` and `protoc-gen-go-grpc` on `PATH` (only to regenerate gRPC stubs)

## Local setup

```bash
git clone https://github.com/mralaminahamed/flagcast.git
cd flagcast
cp .env.example .env
make up      # MongoDB, Redis, NATS and every service
make down    # stop the stack
```

## Go services and packages

```bash
make build       # build gateway, evaluator, ai, mcp into ./bin
make test        # go test ./...
make test-race   # CGO_ENABLED=1 go test -race ./...
make lint        # go vet + gofmt check on apps/ and packages/
make fmt         # gofmt -w apps packages
make tidy        # go mod tidy
```

Full gate before committing:

```bash
gofmt -l apps packages && go vet ./... && go build ./... && CGO_ENABLED=1 go test -race ./...
```

## gRPC contract

The contract lives in `proto/flagcast/v1/`. Generated stubs in
`packages/shared/genproto/` are committed; do not edit them by hand. After
changing a `.proto` file, run:

```bash
make proto   # buf lint + buf generate
```

and commit the regenerated files with the proto change.

## Console (`apps/console`)

```bash
cd apps/console
pnpm install --frozen-lockfile
pnpm dev          # Vite dev server
pnpm type-check   # tsc --noEmit
pnpm lint         # eslint src
pnpm test         # vitest run
pnpm build        # type-check + vite build
```

## Branches, commits and pull requests

- Branch from `trunk`.
- Keep commits small and single-scope, using
  [Conventional Commits](https://www.conventionalcommits.org/):
  `type(scope): description`. Types: `feat` `fix` `docs` `refactor` `perf`
  `test` `build` `ci` `chore`. Scopes: `gateway` `evaluator` `ai` `mcp`
  `console` `infra`.
- Open one pull request per change, against `trunk`.
- Pull requests are merged with a merge commit (not squash), so the scoped
  commits stay in history.
- Never commit `.env`.
