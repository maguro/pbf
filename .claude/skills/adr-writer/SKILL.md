---
name: adr-writer
description: Draft a new Architecture Decision Record (ADR) for this project, or restructure a rough decision into one. Use whenever asked to write, draft, or create an ADR, to document an architectural decision, or to capture a design choice as a formal record. Also use before treating any hand-drafted ADR as done, even one not created through this skill. Pair with the "asd-ste100" skill for prose and the "adr-validate" workflow for a final adversarial check.
---

# Writing an ADR for this project

An ADR earns its place by outliving the conversation that produced it. It
must be readable, and checkable, by someone who was not in the room —
including a future agent that has forgotten everything discussed here. That
constraint drives every rule below.

## Interview before drafting

Don't draft from a one-line request. Ask what's actually needed to write a
decision someone else can evaluate cold:

- What is the decision, stated as a single sentence a reader could act on?
- What forces made this a decision rather than an obvious default — a
  constraint, a prior failure, a real trade-off?
- What alternatives were **actually** considered — not ones invented to fill
  out a section. If nothing else was seriously considered, say that; don't
  manufacture a rejected alternative just because the template has a place
  for one. A section existing to be filled is not a reason to invent its
  contents.
- What does this decision deliberately **not** cover? An ADR that doesn't
  state its own boundary invites scope creep in its own body later.

If the answers aren't available, ask the user rather than filling gaps with
plausible-sounding content. A confident invented answer is worse than an
honest gap, because it reads as settled when it isn't.

## The template

Use this section set and order. If `docs/adr/` already has ADRs, check two or
three of them for exact heading style before writing:

```
# ADR-NNN: <Title>

## Status
Proposed

## Context
<the forces, the problem, why this needed a decision>

## Decision
<the decision itself, stated plainly>

## Rationale
### <reason 1, as a heading>
### <reason 2, as a heading>
...

## Consequences
### Positive
### Negative
### Accepted Trade-Off

## Rejected Alternatives
### <alternative 1, as a heading>
...

## Non-Goals
<what this decision does not cover, and doesn't try to>
```

Status starts as **Proposed**, never Accepted — see `docs/adr/README.md`'s
own convention: an ADR moves to Accepted only after a prototype or spike has
exercised the decision, not on the strength of the argument alone.

Number it the next unused `ADR-NNN` (check `docs/adr/` for the highest
existing number), and add it to the index table in `docs/adr/README.md`.

## The ADR must stand on its own

This is the rule this skill exists to enforce, because it's the one that
erodes quietly during drafting rather than failing loudly.

**Never cite another document's current wording as support for the
decision, unless that document is itself an ADR.** For every citation, ask:
is this offered as *evidence the decision is correct*, or as a *description
of what results from the decision*? Only the second is safe. Any document
whose content can change independently of this ADR — a plan, a roadmap, a
ticket, a design note, anything not itself an ADR — is not a stable source
of rationale; an ADR that leans on one can go stale without anyone touching
the ADR itself, and a reader has no way to tell from the ADR alone that the
citation has drifted. If something in another document is genuinely a
*consequence* of this decision, say what the decision causes — don't cite
the other document's current text as why the decision is correct.

**Exception: an outside source can be evidence.** A published format
specification, or the source code or manual of another project, can support
the decision. Cite it by URL, and pin the version that you read: a release, a
commit, or a wiki revision. A reader can then check the source as it was when
the author wrote the ADR.

**Never reference an ADR that doesn't exist yet, or hasn't been written.**
An ADR is a record of a decision that has been made; a forward reference to
a future one is a promise about content nobody has written, and it can't be
verified by a reader today. Reference only ADRs already present in
`docs/adr/`.

**Don't let a Rationale heading claim more than its body argues, and don't
duplicate an argument another subsection already made.** Before finishing,
reread every Rationale, Consequences, and Rejected Alternatives heading
against its own body: does the body actually make the claim the heading
states, in this section specifically, or did that argument already happen
two subsections earlier under a different heading? A heading that promises
"X is cheaper than Y" needs a body that compares the cost of X to the cost
of Y — not a restatement of a point made elsewhere dressed in new words.

## Write the prose in STE from the start

Load the **`asd-ste100`** skill and write to its checklist as you draft,
rather than drafting normally and converting afterward. Converting after
the fact is the higher-risk mode — it risks losing meaning at exactly the
sentences that mattered — and it's avoidable here, since you're generating
the text rather than translating someone else's.

## Before calling it done

Run the **`adr-validate`** workflow (`.claude/workflows/adr-validate.js`)
against the file, passing its path as `args`. That workflow independently
checks STE conformance, verifies every factual claim against the actual
codebase and cited ADRs, checks the document against its own Non-Goals, and
adversarially verifies every candidate finding before reporting it. A draft
that looks finished is not evidence that it is — treat that appearance as
exactly the moment this check earns its keep, not a reason to skip it.
