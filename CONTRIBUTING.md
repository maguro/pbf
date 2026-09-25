# Contributing

## Development

This repository includes a Codex-friendly workflow:

    $ make fmt
    $ make test
    $ make test-race
    $ make test-integration
    $ make lint
    $ make verify

Project-specific agent guidance is documented in `AGENTS.md`.
The Makefile uses a project-local `.cache/` directory for Go and lint caches to
keep local and sandboxed runs reproducible; `.cache/` is intentionally ignored
by git.
