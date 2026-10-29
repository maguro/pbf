# AGENTS.md

## Required Reads
- Before making any code or test edits, read `CODING_STANDARDS.md`.
- If `AGENTS.md` and `CODING_STANDARDS.md` differ, follow the stricter rule.

## Project Overview
- Language: Go
- Module: `m4o.io/pbf/v2`
- Primary goals: stable PBF encoder/decoder library and `pbf` CLI

## Environment
- Preferred Go version: `1.25.x` (see `go.mod`)
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
- Benchmarks: `make bench`
- Benchmark comparison: `make bench-compare BASE=<commit>` (default `HEAD`)
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

## Benchmarks for changes that affect decode and encode performance

A change that can alter the work done for each blob or entity is measured
before and after the change, on the same machine in the same session.
This includes:

- Code on the decode or encode path: `Decoder.Decode`, `Encoder.Encode`,
  `Encoder.EncodeBatch`, and the packages they call (`internal/decoder`,
  `internal/encoder`, `internal/encoder/packers`, `internal/pb`, `model`).
- Default option values for the decoder and encoder.
- Updates to dependencies on these paths and to the Go version.

When unsure whether a change qualifies, measure it. A change that does not
alter the work done on the success path, such as an error message, error
wrapping, a comment, or a rename, is not measured.

```sh
make bench-compare             # HEAD against the working tree
make bench-compare BASE=HEAD~1 # HEAD~1 against the working tree
```

`bench-compare` alternates the base runs and the working-tree runs, one
sample at a time. Do not compare two `make bench` outputs that ran one
after the other. Drift of the machine during the first run makes a false
difference in time.

If no benchmark covers the path that changed, add one in the same change.
A performance change is not done until the numbers show the gain. A
correctness change on these paths reports its cost.

The commit message states the result in a separate paragraph, for
allocations and bytes per op only, with no time figures:

- No change: `Benchmarks: no change (BenchmarkDecodeLondon, BenchmarkEncodeLondon).`
- A change, where benchstat reports a significant difference and the
  values differ at three significant figures: before → after and the
  percentage, for each benchmark that changed. For example:
  `BenchmarkEncodeLondon: allocs/op 8.37M → 7.10M (−15.2%), B/op 1.90 GB → 1.62 GB (−14.7%).`

A change that is not measured has no benchmark paragraph.

## Execution Defaults
- Make one focused change at a time and stop for review.
- Do not bundle unrelated edits in a single change.
- Do not use staging/cherry-pick style split workflows unless explicitly requested.
- Preserve behavior unless an explicit behavior change is approved.
- For concurrent close/error paths, default to first-wins semantics unless explicitly changed.
- When refactoring a proposed solution, re-scan pending changes and remove code that is no longer relevant.

## Commits

- Write commit messages as `.claude/skills/git-commit-message/SKILL.md` describes.
