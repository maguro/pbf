---
name: go-testing
description: Choose the test framework, assertions, and test structure for Go tests in this repo. Use whenever writing, changing, or reviewing a Go test, example, fixture, or mock, and whenever a bug fix or behavior change needs a test.
---

# Writing Go tests

`AGENTS.md` adds the test rules that apply only to this repo.

## Match the framework to the test

Use Ginkgo (`ginkgo/v2`) for complicated behavioral tests: branching setup, a
fixture shared across many related cases, or timing and concurrency to pin down.
Its containers and Gomega's `Eventually` and `Consistently` earn their weight
there.

Use plain `testing.T` for simple features: a handful of independent assertions,
a table of input-to-output cases, a single fact. Prefer a table-driven test for
new behavior of this kind. Reaching for Ginkgo to assert one thing costs more
than it returns.

A repo that does not have Ginkgo in `go.mod` uses plain `testing.T`. Do not add
Ginkgo to it for one test.

Do not convert an existing test to match. Convert only when you are already
changing it for another reason.

## Assertions follow the framework

Use Gomega (`Expect`) inside Ginkgo specs. Use `testify` (`assert` and `require`)
in plain `testing.T` tests, where it clarifies intent. Do not mix them in one
test. `require` inside a spec bypasses Ginkgo's failure reporting.

## Examples are tests

`example_test.go` runs under `go test`, and its `// Output:` blocks are
assertions. Update them with the behavior they check.

## Prefer generated mocks to hand-written fakes

This rule applies to a repo that has `go.uber.org/mock` in `go.mod`, under
either framework. Stand a collaborator up with a `//go:generate ... mockgen`
directive on the file that declares the interface. Put behavior on the mock, not
in a bespoke struct. For example, `DoAndReturn` supplies a slow call. An
unexpected call already fails the test, so "must never be called" needs no
assertion. When a mock cannot carry something a test needs, such as a
`fmt.Stringer` identity, embed it and add only the missing method.

## Keep passing test output quiet

A suite that does not assert on log content silences `slog`. It does not print
to stderr on every green pass. A spec that checks what got logged installs a
capturing handler for its duration and restores the previous default after. A
suite that reaches logging paths only incidentally swaps in a discard handler for
the whole run, for example in a `BeforeSuite`.
