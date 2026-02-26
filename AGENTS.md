# AGENTS.md

## Project Overview
- Language: Go
- Module: `m4o.io/pbf/v2`
- Primary goals: stable PBF encoder/decoder library and `pbf` CLI

## Environment
- Preferred Go version: `1.23.x` (see `go.mod` toolchain)
- OS support: macOS, Linux, Windows

## Working Agreement
- Keep edits minimal and targeted to the task.
- Preserve public API behavior unless change is explicitly requested.
- Prefer table-driven tests for new behavior.
- Avoid introducing new dependencies without clear need.

## Fast Start
- Format: `make fmt`
- Unit tests: `make test`
- Race tests: `make test-race`
- Integration tests: `make test-integration`
- Lint: `make lint`
- Full local CI parity: `make verify`

## Code Style
- Follow standard `gofmt` formatting.
- Keep package boundaries explicit (`internal/` remains non-public).
- Return wrapped errors when adding new error paths (`fmt.Errorf("...: %w", err)`).

## Validation Before Merge
- Run `make verify`.
- If behavior changes, add or update tests in the same change.
- Keep README/docs in sync with user-facing CLI/library changes.

## Execution Defaults
- Make one focused change at a time and stop for review.
- Do not bundle unrelated edits in a single change.
- Do not use staging/cherry-pick style split workflows unless explicitly requested.
- Preserve behavior unless an explicit behavior change is approved.
- For concurrent close/error paths, default to first-wins semantics unless explicitly changed.
