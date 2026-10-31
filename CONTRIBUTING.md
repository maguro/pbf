# Contributing

## Development

This repository includes a Codex-friendly workflow:

    $ make fmt
    $ make test
    $ make test-integration
    $ make lint
    $ make check

Project-specific agent guidance is documented in `AGENTS.md`. The rules for
every edit are in `CODING_STANDARDS.md`. Task procedures, such as how to write
tests, are in `.claude/skills/`.
The Makefile uses a project-local `.cache/` directory for Go and lint caches to
keep local and sandboxed runs reproducible; `.cache/` is intentionally ignored
by git.
