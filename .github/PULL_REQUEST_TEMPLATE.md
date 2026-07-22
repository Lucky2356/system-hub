<!-- Thanks for contributing. Keep this short. -->

## What this changes

<!-- One or two sentences. Link the issue it fixes, if any: Fixes #123 -->

## Why

<!-- The problem this solves, not just the diff. -->

## Checklist

- [ ] `go build ./...` and `go vet ./...` pass
- [ ] `go test ./... -race` passes
- [ ] `golangci-lint run` is clean — and, if the change touches a platform
      provider, `GOOS=linux CGO_ENABLED=0 golangci-lint run ./internal/...` too
      (dead code under a build tag is otherwise invisible)
- [ ] New user-facing strings go through `internal/i18n` (the coverage test
      fails otherwise) and have a Russian translation
- [ ] Checked on Linux and/or Windows — say which
