---
name: asd-ste100
description: Write or convert prose to ASD-STE100 (Simplified Technical English) against a concrete rule checklist, then verify the conversion against its original for meaning drift, not just style. Use whenever asked to write, rewrite, simplify, or check anything (ADRs, technical docs, code comments, tool descriptions, agent output) in "ASD-STE100", "STE", "Simplified Technical English", or "plain English" style, or to verify STE conformance. Use proactively before treating any STE-style rewrite as done, even if not asked to check it.
---

# ASD-STE100 (Simplified Technical English)

"Write it simply" is not enough instruction to produce STE reliably. General
prose habits (em dashes, semicolons, nested subordination, metaphor) are
strongly trained in. Vague guidance to simplify either overcorrects into
choppy, meaningless fragments, or lets those habits back in without the
writer noticing. Run the checklist below explicitly, sentence by sentence.
Do not rely on a general sense of "simple."

## What the checklist contains, and why it has three parts

**Part A** paraphrases the rules of the standard itself. ASD-STE100 is a
controlled language built by the aerospace and defense industry so that a
technician on a tarmac, with no author to call, cannot misread an
instruction. Issue 9 (January 2025) has 53 rules in 9 sections, plus a
dictionary of about 900 approved words. See `references/writing-rules.md`
for the fuller summary and the sources.

**Part B** names failure modes the standard never had to name, because its
authors were not language models. An em dash joining two clauses, a
semicolon chaining independent assertions, a spatial metaphor for an
abstract relation: these are trained-in habits, and a rule list that omits
them will pass text that no human STE editor would pass.

**Part C** protects meaning. Every rule in Parts A and B can be satisfied by
a rewrite that says something the original did not say. Style conformance
and meaning fidelity are independent failure modes, and Part C is the only
one of the three that catches the second.

A pass that applies Part A alone under-detects. A pass that applies Part C
alone leaves violations in place. Both were observed in a real head-to-head
run of two earlier versions of this skill.

## Two situations, two risk profiles

**Drafting new prose in STE** is lower risk. Apply the checklist as you
write. Self-check before calling a passage done.

**Converting existing prose to STE** is higher risk, because the goal is not
just STE-conformant text. It is text that means exactly what the original
meant.

For conversion, always run four separate passes, in this order:

1. **Inventory** every passage in scope before you rewrite anything. See
   "Inventory first" below. Skipping this step is how a defective passage
   gets left alone and never reported.
2. **Draft** the rewrite, passage by passage.
3. **Re-read what you created.** The draft pass produces sentences that did
   not exist when you started, so no earlier check has ever seen them. See
   "Second pass: check the sentences you just made" below. Do not fold this
   into the draft pass.
4. **Verify** the rewrite against the original, sentence by sentence, as a
   distinct step from applying the checklist. Check specifically for meaning
   drift. See "Verification pass" below. Do this pass yourself if nothing
   else is available, but know that it is weaker than an independent check.
   You already believe the rewrite says what you meant it to say. That
   belief is the blind spot that lets drift through. For a document that
   matters, give the original and the rewrite to a fresh subagent with no
   visibility into how the draft was produced. Have it check only for
   meaning drift. It has no reason to trust the rewrite.

## Inventory first

Before you rewrite anything, list every passage in scope: every doc comment,
every paragraph, every field comment. Give each one a line, in source order,
with its identifier or line number.

Work down that list. A passage you decide to leave alone stays on the list
with the rules you checked it against, so that leaving it alone is a visible
decision and not an absence. A rewrite pass that reports only what it
changed cannot tell the difference between a passage that was clean and a
passage nobody read.

## Rule triggers are not equal, and the easy ones will crowd out the rest

Some rules fire on a pattern you can see without understanding the sentence.
A8 (insert the missing "that") is the clearest case. Others need you to read
the sentence and ask what it claims: A3, C1, C2, and C4 are all of this
kind.

A single sweep runs the cheap rules everywhere and the expensive rules
almost nowhere, because the cheap ones keep finding work. The result looks
thorough. It is not. If your edit count is dominated by one mechanical rule,
that is the symptom, and the fix is the second pass below.

---

# Part A: rules of the standard

## A1. One word, one meaning, held consistently

Pick one verb for one action and reuse it every time. Do not rotate
synonyms: "check", "verify", and "confirm" for the same action force the
reader to guess whether they mean the same thing.

Short-sentence pressure also collapses distinct concepts onto one word.
Where a project already distinguishes terms, keep the distinction. If a
codebase uses "emit" for what a generator produces, "generate" for what an
upstream tool produces, and "write" for what a person authors by hand, a
rewrite that unifies all three on "generate" has destroyed information while
appearing to improve consistency.

**Grep before you assume a term is free.** A term that feels generic is
often established somewhere specific. A search settles it in seconds where a
guess does not.

## A2. One part of speech per word

Use a word in one grammatical role only.

Bad: "Oil the valve." (oil as a verb)
Good: "Apply oil to the valve." (oil as a noun)

## A3. Active voice, with a named actor

Name who or what acts. Passive voice is allowed in descriptive text only
when the actor is genuinely unknown or irrelevant.

Bad: "The type is taken from the first declared result."
Good: "Yama takes the type from the first declared result."

Watch for the agentless participle, which hides the same problem: ", taken
from the first declared result" has no actor and no finite verb.

**Do not overcorrect.** Repeating one generic filler subject ("A person...",
"The project...") in sentence after sentence reads as robotic. That is its
own failure, not a virtue of being active. Vary the subject with whatever
actually acts in that sentence.

## A4. Simple tenses only

Use the infinitive, imperative, simple present, simple past, or simple
future. A past participle is allowed as an adjective. Do not use the present
perfect or the past perfect.

Bad: "We have received the report."
Good: "We received the report."

## A5. No "-ing" form as a verb

An "-ing" form is permitted only inside a technical noun. It is not
permitted as a verb, a gerund, or a participial modifier.

Bad: "the call stating the providers"
Good: "the call that states the providers"

Bad: "A sweep skips the package rather than reporting it."
Good: "A sweep skips the package. It does not report the package."

## A6. One instruction, or one idea, per sentence

Bad: "Open the file and read line 3, then check if it matches."
Good: "Open the file. Read line 3. Check whether it matches."

## A7. Sentence length: about 20 words, or 25 for description

Use 20 words as the cap for procedures and instructions, and 25 for
descriptive text. Treat a longer sentence as a signal to look for a
chainable clause, not as a place to add a comma.

## A8. No ellipsis: keep the parts of the sentence explicit

Do not drop a word to save space. The standard warns that omission creates
ambiguity rather than brevity. Three omissions matter most:

**The article.** Keep "the" and "a" wherever they fit.

**The relative pronoun.** Do not elide "that". "the name the stub bound to
the parameter" becomes "the name that the stub bound to the parameter".

**The repeated verb.** "when the stub bound no name or the blank identifier"
becomes "when the stub bound no name, or bound the blank identifier".

Two related constructions belong here, because both are the same fault in
another form:

**"whose" is not plain English for most readers.** Rewrite it. "a file whose
imports name the packages" becomes "This file's imports name the packages."

**Do not front the object of a relative clause and strand its preposition.**
"the type that a stub's parameter list may end with" becomes "the type that
can end a stub's parameter list".

## A9. Noun clusters: 3 words maximum

Bad: "the agent task queue priority handler"
Good: "the handler that sets task-queue priority"

## A10. Paragraphs and lists

One topic per paragraph, and about 6 sentences maximum. Use a numbered or
bulleted list for 3 or more steps or conditions. Do not bury a sequence
inside one prose sentence.

Safety-critical instructions open with the command or the condition. They
are never buried mid-sentence.

---

# Part B: habits the standard does not name

## B1. No em dash as a clause connector

An em dash joining two independent clauses is an A6 violation wearing
different punctuation. Split into separate sentences.

Bad: "Check your client version — an outdated client is the most common
cause."
Good: "Check your client version. An outdated client is the most common
cause."

## B2. No semicolon or colon chaining independent clauses

A colon introduces a genuine list of items. It never introduces another full
clause.

Bad: "This left a contradiction nobody noticed: X's Y skips Z in silence,
matching W — its doc comment states this, and a test asserts it." (one
sentence, four assertions, joined by colon, em dash, and "and")

Good: "This left a contradiction that nobody noticed. X's Y skips Z in
silence, and this matches W. Its doc comment states this. A test asserts
it."

## B3. Literal language only: no metaphor, no idiom

A reader who is not a native speaker cannot resolve a figure of speech. If a
sentence uses a spatial or physical image for an abstract relation, restate
it literally.

Bad: "the type that may close a parameter list"
Good: "the type that can end a parameter list"

Bad: "It splits the rule along a line nothing defines."
Good: "It has two rules. Nothing says which one applies to a new question."

## B4. "can" and "must", not "may"

"may" is ambiguous between permission and possibility. Use "can" for
ability or possibility, and "must" for obligation.

Bad: "The list may end with a variadic parameter."
Good: "The list can end with a variadic parameter."

## B5. Do not stack relative clauses

One relative clause per sentence. Each nested clause is a place where
meaning gets lost.

Bad: "a second command-line vocabulary that a Wire user must learn to use a
tool built on the assumption that they already know Wire" (four levels of
embedding)

Good: "Answering each question on taste builds a second vocabulary. A Wire
user must learn that vocabulary. The tool assumes that they already know
Wire."

---

# Part C: meaning fidelity

## C1. After splitting a sentence, check that each piece still asserts something

This is the most common way meaning is lost. A long sentence is cut at a
clause boundary, and the kept piece does not carry the clause that gave it
its point.

Bad: "Generation is one go:generate directive." (dropped "...that invokes
Yama, which runs Wire", the clause that made the sentence assert anything)

Good: "A project needs one go:generate directive to generate. That directive
invokes Yama, and Yama runs Wire."

After you split any sentence, reread the new first piece alone. Ask whether
it makes a real, complete claim, or whether it only parses.

## C2. Preserve the scope of negations and quantifiers

"Not every X" (some X are an exception) is a different claim from "not
always" or "doesn't want every", which stack ambiguously. When a sentence
has a negation and a quantifier together, restate the same logical shape in
different words. Do not reconstruct it from a general sense of what it
meant.

Bad: original "not every injector's graph is one an application wants
orchestrated" rewritten as "does not always want to orchestrate every
graph" (three readings, none clearly the original)

Good: "Some injectors build a graph that the application does not want
orchestrated."

## C3. Preserve modal strength

Do not drop or soften "necessarily", "always", or "overwhelming majority".
Do not swap a duration connector ("for as long as X") for an event connector
("until X"). These look like simplifications and change the claim.

Real example: "ran both at once **for as long as** nobody compared them" (a
state that holds while a condition holds) became "ran both at once **until**
nobody compared them" (an event that ends the state). "Nobody compared them"
was true from the start and never became newly true, so "until" has nothing
to trigger on.

## C4. Preserve a stated relation; do not silently delete the connective

Splitting a sentence at "because" or "so" removes a stated causal relation
and leaves two bare facts side by side. That is a content loss, not a style
gain. "because" and "as a result" are plain English and are allowed.

Bad: original "The last path element is a version suffix, **so** an
un-aliased import still resolves to yamaPkgName" rewritten as two unlinked
sentences.

Good: "The last element of the path is a version suffix. **Because of
this**, a file that imports the path without an alias still refers to the
package as yamaPkgName."

Where a house style forbids rationale (a code comment, for example), do not
substitute a bare juxtaposition and call it equivalent. Raise the conflict
instead.

## C5. Do not orphan a discourse adverb when you restructure

"otherwise", "then", "instead", "also", "such", and "this" point at
something in the sentence structure around them. Restructuring moves or
deletes what they point at, and the adverb stays behind aimed at nothing.

Real example. The original packed a condition into a relative clause: "An
un-aliased import whose declared name differs from its path's last element
is **otherwise** misnamed." Here "otherwise" means "without this map". A
rewrite turned the relative clause into an explicit conditional and kept the
adverb: "**If** its declared name differs from the last element of its path,
the derived injector file **otherwise** misnames this import." "If X, then
otherwise Y" points at nothing and parses to nothing.

Good: "Without this map, the derived injector file misnames such an import."

After you convert a relative clause to a conditional, split a sentence, or
reorder clauses, find every discourse adverb in the passage and name what it
now refers to. If you cannot name it, delete the adverb or restore the
structure it depended on.

## C6. Add nothing

Do not add a fact, an example, or an inference that the original did not
state, even when it reads as a reasonable one. In particular, do not add a
"because" clause that explains a decision the original only stated.

---

# Do not rewrite what you were not asked to rewrite

An STE pass changes wording. It does not change anything else.

- **Keep the wrap column of the source.** Do not rewrap lines whose text did
  not change. Gratuitous rewrapping buries the real edits in diff noise.
- **Keep established terminology** (see A1). Fixing a local inconsistency by
  choosing a word the project does not use is a regression.
- **Keep the structure.** Do not reorder sections, merge paragraphs, or add
  headings unless the rewrite requires it.

# Second pass: check the sentences you just made

Every sentence you wrote in the draft pass is new. No rule has been applied
to it, because it did not exist when you applied the rules. A split turns
one checked sentence into two unchecked ones, and the checklist has already
moved on.

So run a separate pass over your own output. Take each sentence you wrote or
changed and apply exactly these four checks. Do not do this while drafting.

**1. C1: does the sentence assert something real?** Read it alone, with the
sentences around it covered. Ask what claim it makes. A first piece left
behind by a split is the usual offender.

Bad: "StubError reports a lifecycle stub." (true of nothing; the dropped
clause was "that Yama cannot derive an injector from", which is what the
sentence was about)

Good: "StubError reports a lifecycle stub that Yama cannot derive an
injector from."

**2. A3: who acts?** Name the subject of every verb. If the sentence has no
actor, or the actor is a participle with nothing attached to it, rewrite it.

**3. C5: what does each pointing word refer to?** Name the referent of every
"it", "this", "one", "such", "otherwise", "then", and "instead". If you
cannot name it, the word is wrong. In a doc comment, prefer repeating the
identifier over a pronoun: "Call OptsName only when..." rather than "Call it
only when...".

**4. Agreement and predication.** A definition sentence of the form "X is Y"
must agree in number, and Y must actually be what X is.

Bad: "StubPackage **is** the lifecycle **stubs** that one package declares."
Good: "StubPackage **holds** the lifecycle stubs that one package declares."

If this pass changes a sentence, that new sentence has not been checked
either. Run the four checks on it before you move on.

# After the checklist, read it again as a reader

A checklist finds what it names, in the order it names it. It can also make
you stop looking once you have been through the list, including for the
things it covers. In one test run, an agent worked through this checklist
against a flawed passage and still missed a passive-voice violation, a rule
that was in the list it had just applied.

After the numbered pass, read the whole passage straight through once more,
as a reader would, with the checklist out of mind. Ask only this: does
anything here read badly, sound vague, feel overcomplicated, drop something
specific for something general, or contradict itself? Trace anything that
snags back to the rule it violates. There is almost always one, even if that
sentence passed the numbered pass. The checklist tells you *why* something
is wrong once you notice it. It is not a substitute for noticing.

# Verification pass (conversion mode)

Read the original and the rewrite side by side, sentence by sentence, or
clause by clause where sentences were split or merged. For each pairing, ask
one question: **does the rewrite assert the same thing the original did, no
more, no less, and with no different scope?**

This is a different question from "is the rewrite STE-conformant". A
sentence can pass every rule above and still fail this one. Flag:

- A claim that is stronger, weaker, or differently scoped than the original.
- A grammatical but vacuous sentence (see C1).
- A dropped qualifier, example, or fact that changes what a reader concludes.
- Added content not present in the original (see C6).
- A deleted causal or conditional connective (see C4).
- A discourse adverb that no longer refers to anything (see C5).

# Scale: single edit against batch rewrite

For one sentence or paragraph, apply the checklist inline and move on.

For a document-level or multi-document rewrite, the draft-then-verify
process matters more, not less. It also works better as its own focused task
(a dedicated session, or a workflow with a separate verification stage) than
folded into a session that is also doing unrelated content or code work.
Style discipline degrades under split attention. A rule such as "no em dash"
gets followed while STE is the active task, then drifts back in during a
later, unrelated edit to the same document, because attention is on the
content fix. Isolating the rewrite as its own task removes that competition.

# Two more checks, same review pass

These are not STE rules. A careful read against a stated reference catches
them, so run them in the same pass.

**Self-consistency against stated scope.** If the document states what it
does not cover (an ADR's "Non-Goals", for example), reread the whole
document against that section after any edit. Flag a body claim that the
document's own scope section puts out of bounds.

**Provenance.** Flag any claim, list, or "alternative considered" that is
stated as fact or history with no verifiable source behind it: no cited
code, document, or measurement. Do not let something pass because it fits
the expected shape of a section. A "Rejected Alternatives" section does not
need a third entry because similar documents have three, and an invented but
plausible alternative that nobody proposed misrepresents how deliberated the
decision was.

# Boundaries

**Will:**

- Rewrite dense or ambiguous English into short, single-meaning,
  active-voice sentences.
- Name the rule a sentence violates before rewriting it.
- Preserve every fact, condition, and scope qualifier.
- Suggest a one-line glossary entry for a domain term that must stay.

**Will not:**

- Reproduce ASD's official dictionary of about 900 words as if it were
  memorized. Treat the official download as the source of truth for exact
  approved wording.
- Simplify creative, marketing, or persuasive copy, where voice and nuance
  are the point.
- Drop a safety condition, an exception, or a scope qualifier to shorten a
  sentence. It flags the trade-off instead.
- Guarantee an aerospace-grade or defense-grade STE-compliant document. This
  is a clarity tool built on STE. It is not a certified STE authoring tool.

# Additional resources

- `references/writing-rules.md`: a fuller summary of the 9 rule sections and
  the dictionary structure, with citations to the official standard and to
  secondary sources.
- `references/examples.md`: worked before-and-after examples: illustrations
  of the standard's rules, and rewrites of agent-facing text. The commentary
  in that file still uses em dashes as prose punctuation. The rewritten
  examples themselves do not.
- `ATTRIBUTION.md`: which source skill each section came from.
- The official standard is free to download at https://www.asd-ste100.org/.
  Use it when exact approved wording matters.
