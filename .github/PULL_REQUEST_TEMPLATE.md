## Summary

<!-- What does this change and why? -->

## Linked issue

<!-- e.g. Closes #123, or "none" -->

## Checklist

- [ ] Branched from `trunk`; one change per pull request
- [ ] Commits are single-scope Conventional Commits (`type(scope): description`)
- [ ] Go gate passes: `gofmt -l apps packages && go vet ./... && go build ./... && CGO_ENABLED=1 go test -race ./...`
- [ ] Console changes: `pnpm type-check`, `pnpm lint`, `pnpm test` and `pnpm build` pass in `apps/console`
- [ ] Proto changes: ran `make proto` and committed the regenerated stubs
- [ ] Docs updated where behaviour or configuration changed
- [ ] No secrets or `.env` committed
