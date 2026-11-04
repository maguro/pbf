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
| `adr-writer` | Writing an ADR, or finishing one drafted by hand |

`.claude/workflows/` holds the `adr-validate` and `adr-validate-lite` workflows.
They run a final adversarial review of an ADR in `docs/adr/`.

## Design docs

`docs/adr/` holds the design docs of this repo. Each design decision and its
reasons go in an ADR there.

## Project overview

- Language: Go
- Module: `m4o.io/pbf/v2`
- Primary goals: stable PBF encoder/decoder library and `pbf` CLI
- Preferred Go version: `1.25.x` (see `go.mod`)
- OS support: macOS, Linux, Windows

## Commands

- Format: `make fmt`
- Format check: `make check-fmt`
- Unit tests, with the race detector: `make test`
- Integration tests: `make test-integration`
- Benchmarks: `make bench`
- Benchmark comparison: `make bench-compare BASE=<commit>` (default `HEAD`)
- Coverage: `make coverage`
- Fuzz: `make fuzz FUZZ=<target> FUZZTIME=<duration>`
- Lint: `make lint`
- Gate target: `make check`

## Before each commit

Do all of these steps before each commit. Do not ask for permission to do
them.

1. Run `make fmt`, `make test`, and `make lint`.
2. Use the `asd-ste100` skill to check each comment that the change adds or
   changes. Also check the commit body.
3. Run `/code-review` against the diff, unless the change is trivial.
4. Use the `go-benchmarks` skill if the change touches a hot path that this
   file lists.
5. Run `make check` in the working tree.

Tell the user which steps you did and which steps you skipped. Give the reason
for each skipped step.

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
