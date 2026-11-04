# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for pbf.

ADRs document significant architectural decisions, the context that shaped them,
the alternatives considered, and the consequences accepted by the project. They
are intended to preserve design rationale for future maintainers and reviewers.

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [ADR-001](ADR-001-encoder-keeps-caller-order.md) | The Encoder Keeps the Caller's Entity Order | Accepted |

## Conventions

New ADRs should:

* Use the filename format `ADR-NNN-short-title.md`.
* Use a monotonically increasing three-digit number.
* Include a status section near the top of the document.
* Record the decision, rationale, consequences, and rejected alternatives.

An ADR starts life as `Proposed`, not `Accepted`. It moves to `Accepted` once a
prototype or spike has exercised the decision — a design that hasn't touched code
yet is not settled, however confident the reasoning. A `Proposed` ADR is expected
to change; an `Accepted` one is not, so promoting one before it has survived
contact with an implementation invites the rewrite this pre-release stability
rule elsewhere presumes it avoided.

Accepted ADRs represent the current architectural direction unless a later ADR
explicitly supersedes them.

An ADR here describes the decision as it currently stands, not the sequence of
revisions that produced it. When a decision changes, its ADR is rewritten to
state the new decision and the reasoning for it, and the superseded design moves
to that ADR's Rejected Alternatives if the reasoning against it is worth keeping.
There are no amendment sections and no change logs; `git log` is the history.

This holds while v2 is not released and has no external consumers relying on a
decision's stability. Once it does, a change that breaks them wants a new ADR
that supersedes the old one by number, so the old decision stays readable at the
version it applied to.
