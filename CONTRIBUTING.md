# Contributing to pbf

## Get pbf

Add pbf to a Go module with `go get m4o.io/pbf/v2`. The
[README](README.md#library-usage) shows how to decode and encode a file.
Install the `pbf` command with `go install m4o.io/pbf/v2/cmd/pbf@latest`.

## Report a bug

Report a bug in a GitHub issue:

https://github.com/maguro/pbf/issues/new

Include these items in the report:

- The version of pbf and the version of Go.
- The steps that show the problem. If you can, name the PBF file that causes
  it.
- The result that you expected and the result that you got.

Do not report a security vulnerability in a public issue. Use the process in
[SECURITY.md](SECURITY.md).

## Ask for a feature

Open a GitHub issue that describes the problem to solve. Open the issue
before you write the code. A change to the public API or to the output of
the encoder needs agreement before work starts.

## Send a change

1. Fork the repository and make a branch from `master`.
2. Make the change.
3. Run `make check`. This command runs the format check, the tests with the
   race detector, the integration tests, and the linter.
4. Open a pull request against `master`. Refer to the related issue in the
   description.

To round-trip a PBF file that is not in `testdata`, such as a planet file,
give its path to the integration test with the `-pbf` flag:

```sh
go test -tags=integration -run TestRoundTripOSMPBFDatasets -timeout 0 . -pbf /path/to/planet.osm.pbf
```

The test writes two copies of the file to the directory that `TMPDIR`
names. Set `TMPDIR` to a disk with enough free space.

The maintainer reviews each pull request. CI must pass before a merge.

The Makefile keeps the Go and lint caches in a `.cache/` directory in the
repository, so that local runs and sandboxed runs give the same result. Git
ignores `.cache/`.

## Requirements for a change

A change must obey the project conventions in
[CODING_STANDARDS.md](CODING_STANDARDS.md) and [AGENTS.md](AGENTS.md). These
conventions apply to all contributors. Their baseline is the
[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

- Add tests for new behavior. Add a test that shows the bug for a bug fix.
- Keep `make check` passing.
- Write code comments in ASD-STE100 Simplified Technical English.
- Measure a change to the decode or encode path, and report the result in the
  commit message. The
  [`go-benchmarks` skill](.claude/skills/go-benchmarks/SKILL.md) gives the
  procedure, and [AGENTS.md](AGENTS.md#hot-paths) lists the paths.
- Write commit messages that obey the
  [`git-commit-message` skill](.claude/skills/git-commit-message/SKILL.md).

## License

pbf uses the [Apache License 2.0](LICENSE). When you send a contribution, you
release it under this license.
