# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/). The project has no
tagged releases yet, so entries are grouped by the date they were merged into
`trunk`.

## [Unreleased]

## 2026-09-23

### Changed

- Local development ports moved to the 8200 block (#30).

## 2026-08-25 — hardening, observability and tests

### Added

- Business metrics, and tracing of Claude calls (#23).
- Grafana dashboard and Prometheus alert rules (#24).
- Console mobile navigation, touch actions and optimistic toggle (#28).
- Backend tests for the evaluator read path and the analyzer (#27).
- Console tests with Vitest, and ESLint (#29).
- Logo, icon and a rewritten README (#25, #26).
- MIT license file (#26).

### Fixed

- Evaluator cache can no longer go permanently stale (#17).
- Evaluator graceful shutdown and a cap on Watch subscribers (#18).
- Deployment: Terraform remote state, ECS health checks and rollback (#19).

### Security

- Gateway `/api` authentication fails closed (#20).
- Evaluator gRPC authentication, with reflection off in production (#21).
- Terraform: Secrets Manager, database encryption and HTTPS (#22).

## 2026-08-25 — initial build

### Added

- Gateway REST flag CRUD with an audit trail (#5).
- gRPC evaluator, protobuf contract and Go SDK (#6).
- Redis evaluation cache and NATS change stream (#7).
- Streaming `Watch` RPC and SDK `Watch` (#8).
- React console (#9).
- AI rollout analysis: A/B statistics and Claude (#10).
- MCP server (#11).
- OpenTelemetry distributed tracing (#12).
- AWS infrastructure as code with Terraform (#13).
- Continuous delivery: image build/push and a gated deploy (#14).

### Changed

- Dependency bumps: `actions/checkout` to v7 (#1),
  `prometheus/client_golang` to 1.24.1 (#3), `rs/zerolog` to 1.35.1 (#4),
  `actions/setup-go` to v7 (#16).

### Removed

- Dependabot configuration (#15).
