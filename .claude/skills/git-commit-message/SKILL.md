---
name: git-commit-message
description: Generate a Chris Beams–style Git commit message summarizing the agent’s changes.
version: 1.0
---

# Skill: Git Commit Message (Chris Beams style)

## Purpose

After making code changes, produce a **single Git commit message** that summarizes the work, following the guidelines in:
https://cbea.ms/git-commit/

## Inputs

- The staged diff or the full diff of the change.
- The list of files changed.
- Any user intent, task description, or ticket reference already present.

## Output (exact)

Return **only** the commit message text in this structure:

- Subject line (required)
- Blank line
- Body (optional, but required if the change is not self-explanatory)
- Optional footer lines for issue references

## Rules (must follow)

1. Separate subject from body with a blank line.
2. Subject line should be ≤ 50 characters (hard max 72).
3. Capitalize the subject line.
4. Do not end the subject with a period.
5. Use the imperative mood (“Add…”, “Fix…”, “Remove…”).
6. Wrap body text at 72 characters.
7. Body should explain **what and why**, not how.
8. Body defaults to one short paragraph, 2–4 sentences. A second
   paragraph is rare and must earn its place — it is not a default slot
   to fill.
9. Write the body in ASD-STE100. Use the `asd-ste100` skill to write
   and check it.

## Procedure

1. Identify the single primary purpose of the change.
2. Draft 2–3 candidate subject lines and choose the best.
3. Write the subject as an imperative outcome statement.
4. Include a body if rationale or behavior change needs explanation.
5. Add issue references only if explicitly present in context.

## Writing the body (what belongs there)

- Write for a reader running `git blame` years from now. Assume they have
  the diff and the code — but not the plan doc, the ticket, or your memory
  of today. A commit is immutable history: it's read as "what this was and
  where it was headed," so that orientation is worth giving.
- State the change's effect and role at an altitude the diff doesn't already
  show. The diff is the _how_; don't re-narrate the mechanics.
- Don't repeat what the code documents itself — a doc comment, README, or
  type. If it already lives in the tree, the body shouldn't echo it.
- Record what the diff _can't_ reveal: deliberate omissions (why there's no
  X), non-obvious trade-offs, the decision behind the shape.
- Several changes landing in one commit don't each need their own sentence —
  name what ties them together once, rather than explaining each one in
  turn. One unifying sentence beats a roll call of the files touched.
- Context about what the change enables, or where adjacent pieces live, is
  worth one sentence only if a reader would otherwise wonder about it — not
  because it's true. Phrase it as durable architecture ("recovery lives in
  the level runner that calls this"), never as ephemeral bookkeeping (phase /
  PR / ticket numbers, plan-relative "next step") a future reader can't
  resolve.

## Quality checks (regenerate if any fail)

- Subject starts lowercase → fix
- Subject ends with “.” → remove
- Subject is past tense → convert to imperative
- Subject > 72 chars → shorten
- Body lines > 72 chars → reflow
- Body fails an `asd-ste100` check → rewrite
- Body focuses on implementation detail instead of intent → rewrite
- Body restates the diff's mechanics → raise altitude to effect/intent
- Body repeats a fact already in a code/doc comment → cut
- Forward reference points at an ephemeral artifact (phase, PR, ticket) → rephrase as durable architecture, or cut
- Body could be shorter and still orient the reader → cut, even if every sentence is individually true
- Body has a sentence per subject-line component → merge into one sentence naming what ties them together

## Examples

Simple:

Fix quorum validation when closing session

With body:

Add typed failures for session open commands

Return explicit failure variants. Do not throw exceptions. This lets
the caller handle authorization errors and invariant violations in a
deterministic way.

Multi-part:

Refactor motion numbering input into segmented control

Split the field into three parts: the year, the type code, and the
sequence number. This makes the field match the UI specification. This
also prevents invalid partial values.

- Preserve existing styling
- Keep validation localized to the control

## Do not

- Do not output anything except the commit message text
- Do not include diffs or code
- Do not invent ticket numbers or motivations not supported by context
