# CODING_STANDARDS.md

## Scope and Review
- Make one focused change at a time.
- Stop after each change for review before continuing.
- Do not bundle unrelated edits.
- Keep commit boundaries aligned with a single purpose.
- Do not split changes via staging/cherry-pick workflows unless explicitly requested.

## Behavior and Compatibility
- Preserve existing behavior unless a behavior change is explicitly approved.
- Preserve public API compatibility unless explicitly requested.
- Prefer minimal, targeted edits over broad refactors.

## Errors
- Prefer typed sentinel errors for public-facing failure modes.
- Use `errors.Is` in tests and call sites; avoid string-matching error text.
- Wrap underlying errors with `%w` when returning new error paths.

## Concurrency and Close Semantics
- For concurrent close/error paths, default to first-wins semantics.
- If first-wins is changed, document the reason and add focused tests.

## Testing and Validation
- For each change, run:
  - `make fmt`
  - `make test`
  - `make lint`
- If behavior changes, add or update tests in the same change.
- Run `make verify` before merge.

## Dependencies and Architecture
- Avoid adding dependencies without a clear need.
- Keep `internal/` package boundaries explicit and non-public.
