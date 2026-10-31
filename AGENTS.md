# AGENTS.md

## Required reads

Before making any code or test edits, read `CODING_STANDARDS.md`. It holds the
rules that apply to every edit. If this file and `CODING_STANDARDS.md` differ,
follow the stricter rule.

## Skills

Use these skills from `.claude/skills/` when the task matches:

| Skill | Use it when |
|---|---|
| `go-testing` | Writing or changing a test |
| `go-benchmarks` | Changing a hot path listed below |
| `git-commit-message` | Writing a commit message |
| `asd-ste100` | Writing or checking a code comment or a commit body |

## Project overview

- Language: Go
- Module: `m4o.io/pbf/v2`
- Primary goals: stable PBF encoder/decoder library and `pbf` CLI
- Preferred Go version: `1.25.x` (see `go.mod`)
- OS support: macOS, Linux, Windows

## Commands

- Format: `make fmt`
- Unit tests: `make test`
- Race tests: `make test-race`
- Integration tests: `make test-integration`
- Benchmarks: `make bench`
- Benchmark comparison: `make bench-compare BASE=<commit>` (default `HEAD`)
- Lint: `make lint`
- Gate target: `make verify`

For each change, run `make fmt`, `make test`, and `make lint`. Run `make verify`
before merge.

## Hot paths

The `go-benchmarks` skill applies to these paths:

- Code on the decode or encode path: `Decoder.Decode`, `Encoder.Encode`,
  `Encoder.EncodeBatch`, and the packages they call (`internal/decoder`,
  `internal/encoder`, `internal/encoder/packers`, `internal/pb`, `model`).
- Default option values for the decoder and encoder.

The compare command takes the base as `BASE`:

```sh
make bench-compare             # HEAD against the working tree
make bench-compare BASE=HEAD~1 # HEAD~1 against the working tree
```
