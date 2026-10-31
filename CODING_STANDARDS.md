# Coding standards

These rules apply to anyone, human or AI, on any model, who changes this repo.
This file is identical in pbf, gslog, and yama. Change it in all three repos
together. `AGENTS.md` names the commands of this repo and adds its project
rules. If `AGENTS.md` and this file differ, follow the stricter rule.

Task procedures are skills in `.claude/skills/`. `AGENTS.md` lists the skills
that this repo has.

## Go style baseline

The Go style baseline is the
[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).
Follow it for anything this file does not cover. Where a rule in this file
conflicts with the guide, this file wins.

`.golangci.yml` enforces the part of the guide that a linter can check. The
rest depends on the review pass in "Before declaring work complete".

Do not use language or standard-library features newer than the Go version
that `go.mod` declares.

## Scope and review

- Keep each change to one concrete fix, or one directly required refactor.
- Do not combine multiple fixes in one change, even when they share a goal.
- Keep diffs small enough to review end-to-end in a single pass.
- Keep commit boundaries aligned with that single fix.
- Stop after each change for review before continuing.
- Do not use staging or cherry-pick style split workflows unless explicitly
  requested.
- When refactoring a proposed solution, re-scan pending changes and remove code
  that is no longer relevant.

## Analysis depth

- For nontrivial work, do deep analysis before editing.
- Do not stop at first-order reasoning. Check invariants, race and close paths,
  and edge cases.
- State assumptions explicitly and verify them against the code before
  implementation.

## Behavior and compatibility

- Preserve existing behavior unless a behavior change is explicitly approved.
- Prefer minimal, targeted edits over broad refactors.
- Every exported symbol outside `internal/` is part of the public API. Do not
  rename, remove, or change the signature of an exported symbol as a side
  effect of other work. A breaking change needs its own change and explicit
  approval, and the pull request describes it.

## Errors

- Prefer typed sentinel errors for public-facing failure modes.
- Use `errors.Is` in tests and call sites. Do not match error text.
- Wrap underlying errors with `%w` when returning new error paths.

## Concurrency and close semantics

- For concurrent close and error paths, default to first-wins semantics.
- If first-wins is changed, document the reason and add focused tests.

## Dependencies and architecture

- Do not add a dependency without a clear need.
- Keep `internal/` package boundaries explicit and non-public.
- For unexported constructors and helpers, assume callers uphold internal
  contracts.
- Do not add defensive branches for impossible internal states that public
  entry points guarantee.

## Before declaring work complete

- If behavior changes, add or update tests in the same change.
- Keep the README and docs in sync with user-facing changes.
- For any change of nontrivial size, run `/code-review` against the diff. Check
  comment content, test framework choice, formatting, and the Go style baseline.
  A dedicated review pass catches drift that a single pass of authoring misses.
- A change that touches code is not done until the gate target that `AGENTS.md`
  names passes in the actual working tree, not in a scratch copy or worktree.

## Code comments

All code comments follow ASD-STE100 (Simplified Technical English). Use the
`asd-ste100` skill to write or check them. Do not rely on general-purpose
simplifying.

Document all exported types, funcs, vars, and consts with Go doc comments.
Document exported interface methods with concise behavior contracts: inputs,
outputs and errors, and lifecycle and concurrency semantics where they apply.

A comment states what the code does, how to use it, and why it is the way it is
when that is not obvious: an invariant, a constraint of a format or a
dependency, a consequence the caller must know. Keep each reason to a sentence.
For non-obvious behavior, such as concurrency, ownership, lifecycle, and
invariants, add a concise comment at the point of use. Do not add comments that
only restate the code.

**Anything that is only true relative to an earlier state of the code belongs in
the commit message.** What the code did before, what was tried, which
alternative lost, and how the change was arrived at do not go in the source. The
test is time, not category. A reason that holds for a reader who never saw the
change is a comment. A reason that only makes sense next to the previous version
is commit history.

**Never note future or provisional state.** No "not implemented yet" and no
"will be cleaned up later". If a stub needs a note, put it in the function body
next to the `panic` or stub return, so it is deleted with the code it describes.

**Self-check before calling a file done.** For every comment you wrote or
touched, apply the audience test. Caller concerns (behavior, contract,
consequences of use) belong in the doc comment. Maintainer concerns (why
obvious-looking code is wrong) belong at the definition site as a bare fact, not
an essay. That covers invariants the compiler cannot catch, such as an implicit
lock order or a slice that must stay sorted. It does not cover exported
signatures, where the diff itself is the warning.

## Package doc comments

Every non-`main` package has a package comment, in exactly one file, starting
with the package name: `// Package foo ...`.

A package with more than one source file puts that comment in `doc.go`. That
file holds the package comment and nothing else, except `//go:generate`
directives. A package with exactly one source file puts the comment at the top
of that file. The filename does not matter, and a file named after the package
does not exempt a multi-file package from `doc.go`.

## Code formatting beyond gofmt

gofmt does not enforce these rules. Apply them by hand, in production and test
code alike.

- **Method ordering.** This intentionally deviates from the Uber guide, which
  sorts by call order and allows an unexported helper to sit next to its one
  caller. Group methods by receiver, with a `NewXYZ` constructor immediately
  after its type definition. Place every exported method before any unexported
  one, so a reader can take in the public surface of the type breadth-first.
  Order the exported methods in rough call order (for example, `Start` before
  `Stop`). If the type implements more than one interface, group its exported
  methods by the interface they satisfy. Unexported helpers follow, each near
  the exported method it supports. Unattached utility functions go last in the
  file.
- **Single-line bodies.** A function body shares the signature's line only when
  it is empty (`{}`) or a single `return` of a bare value: a field, identifier,
  or literal with no call. Anything that does work, such as a `return` or
  `panic` whose expression calls something, multiple statements, or control
  flow, goes on its own line.
- **Blank-line grouping.** Separate the logical steps of a function with a
  single blank line, so its structure is visible at a glance. For example, set a
  guard clause off from the work it protects. Keep tightly coupled statements
  together. For example, a loop's accumulator seed stays with its loop.
- **No calls inside call arguments.** Hoist a function or method call out of
  another call's argument list into a named local. There are two reasons, and
  either alone is enough. Stepping into the outer call in a debugger otherwise
  walks the inner calls first. The name given to the result documents what the
  call returns. Skip the hoist only where neither applies: the outer call is one
  you never step into, such as a stdlib helper
  (`filepath.Base(fset.Position(f.Pos()).Filename)`) or an assertion wrapper
  (`Expect(...)`, `require.NoError(...)`), or the inner call is a trivial
  accessor whose name already says what it returns.
