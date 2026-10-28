# Attribution

This skill merges two skills that were installed side by side in this repo.
Both are now removed, and this skill replaces them.

**`asd-ste100-skill`** — from https://github.com/danyuchn/asd-ste100-skill,
MIT licensed, Copyright (c) 2026 Dustin Yuchen Teng. See `LICENSE`. It
supplies the rules of the standard (Part A of `SKILL.md`), the output-table
format, the Boundaries section, and `references/writing-rules.md`, which is
copied unchanged. It was never tracked in git. Fetch it from the URL above
to consult the original.

**`ste100`** — project-local. It supplies the draft-then-verify process, the
verification pass, the read-as-a-reader pass, the meaning-fidelity rules
(Part C), the model-habit rules (Part B), the scale guidance, and the
self-consistency and provenance checks. Its last version is in git history
at `.claude/skills/ste100/SKILL.md`.

Rules with no source in either skill, added from a head-to-head test of the
two on the same file:

- **A8**, extended to cover `whose` and stranded prepositions in
  object-fronted relative clauses.
- **B4** (`can` and `must`, not `may`).
- **C4** (preserve a stated causal relation).
- **"Do not rewrite what you were not asked to rewrite"** (wrap column,
  established terminology, structure).
- **"Report what you did not change, with the check it passed"**.

Rules added after a second test, in which this skill rewrote the same file
and left five defects that the two source skills had caught:

- **"Inventory first"** and **"Report what you did not change"**, which now
  requires a line per passage. A skipped passage was invisible before.
- **"Second pass: check the sentences you just made"**, which runs C1, A3,
  C5, and an agreement check over the draft's own output. A cheap mechanical
  rule had crowded out the rules that need the sentence to be read.
- **C5** (do not orphan a discourse adverb when you restructure).
