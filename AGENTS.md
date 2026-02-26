# AGENTS.md

## Required Reads
- Before making any code or test edits, read `CODING_STANDARDS.md`.
- If `AGENTS.md` and `CODING_STANDARDS.md` differ, follow the stricter rule.

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

## Analysis Depth
- When collaboration level is `High` or `Extra High`, perform deep analysis before editing.
- Do not stop at first-order reasoning; check invariants, race/close paths, and edge cases.
- State assumptions explicitly and verify them against code before implementation.

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
- When refactoring a proposed solution, re-scan pending changes and remove code that is no longer relevant.
