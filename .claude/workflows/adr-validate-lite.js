export const meta = {
  name: 'adr-validate-lite',
  description: 'Adversarial review of an ADR for provenance, self-consistency, and structural/self-containment issues, without the style dimension',
  whenToUse: 'Use in place of adr-validate when the token cost of the full run is not warranted. It drops the Style dimension, which produces the most candidates and so the most verification fan-out, and it verifies each candidate with two skeptics rather than three. Run the "asd-ste100" skill over the ADR separately: it is a single pass with no fan-out, so it costs a fraction of the Style dimension it replaces.',
  phases: [
    { title: 'Precheck', detail: 'cheap mechanical fact-check: ADR number references actually exist' },
    { title: 'Find', detail: 'three independent reviewers scan the ADR from different angles' },
    { title: 'Verify', detail: 'each finding checked by two independent skeptics, opus, high effort' },
  ],
}

const parsedArgs = typeof args === 'string' ? JSON.parse(args) : args
const ADR_PATH = parsedArgs && parsedArgs.adrPath
if (!ADR_PATH) {
  throw new Error('adr-validate-lite requires args.adrPath, a path to the ADR relative to the repository root, e.g. "docs/adr/ADR-013-my-decision.md"')
}

const ROOT_NOTE = `The path given, "${ADR_PATH}", is relative to the repository root. Locate the repository root yourself first (look for a .git directory, or a go.mod/package.json/similar project marker) starting from your current working directory, then read the file relative to that root. Do not assume any particular absolute path.`

// Named, citable categories a reviewer can refer to directly ("this fails the
// Evidence gate"), rather than an undifferentiated pile of prose findings.
const GATES = ['Evidence', 'Consistency', 'Self-Containment', 'Structure']

const FINDINGS_SCHEMA = {
  type: 'object',
  properties: {
    findings: {
      type: 'array',
      items: {
        type: 'object',
        properties: {
          gate: { type: 'string', enum: GATES, description: 'which named gate this finding belongs to' },
          summary: { type: 'string', description: 'one-sentence statement of the defect' },
          location: { type: 'string', description: 'section heading or approximate line' },
          quote: { type: 'string', description: 'short quote from the document showing the issue' },
          why_it_matters: { type: 'string', description: 'concrete consequence for a reader if this is not fixed' },
        },
        required: ['gate', 'summary', 'location', 'why_it_matters'],
      },
    },
  },
  required: ['findings'],
}

const VERDICT_SCHEMA = {
  type: 'object',
  properties: {
    refuted: { type: 'boolean' },
    reasoning: { type: 'string' },
  },
  required: ['refuted', 'reasoning'],
}

// Purely mechanical fact-check: does a referenced ADR number actually exist?
// This has no interpretive component, so it doesn't need three Opus skeptics
// arguing over a grep result — a cheap two-agent agreement check is enough,
// and disagreement between them is itself a signal to not trust the finding.
const PRECHECK_PROMPT = `${ROOT_NOTE} Read the ADR in full. This is a purely mechanical fact-check, not a judgment call: grep the file for every occurrence of "ADR-" followed by three digits. For each one, confirm whether a file matching that number exists in the same directory as this ADR. Report a finding only for a reference to an ADR number with no matching file, or a number higher than the current file's own number (a forward reference to a decision that has not been written). Do not report anything else — this check is scoped to ADR-number existence only, not to whether the citation is appropriate. For each finding, give gate: "Self-Containment", the summary, exact location, a short quote, and why_it_matters (state that a forward or nonexistent reference cannot be verified by a reader today).`

const DIMENSIONS = [
  {
    key: 'provenance',
    prompt: `${ROOT_NOTE} Read the ADR in full. For every factual claim, citation, code reference, or "established fact" the document asserts, verify it independently against the actual repository — grep for cited files, functions, or doc comments; read every other ADR it cites, in the same directory as this file. For a citation of an outside source (a format specification, or another project's source code or manual), fetch it at the pinned version that the ADR names and check the claim against it. Flag: any claim that does not check out against what you find, any claim presented as fact or history with no verifiable source, and any "rejected alternative" that reads as invented rather than something that actually happened in this project's code, docs, or history. State exactly what you checked each claim against. Tag every finding gate: "Evidence". For each finding, give the summary, exact location, a short quote, and why it matters.`,
  },
  {
    key: 'self-consistency',
    prompt: `${ROOT_NOTE} Read the ADR in full, and read its "Non-Goals" section closely first. Then reread the entire document (Context, Decision, Rationale, Consequences, Rejected Alternatives) against every Non-Goals bullet. Flag any claim in the body that the document's own Non-Goals section says is out of scope for this ADR. Tag every finding gate: "Consistency". For each finding, give the summary, exact location, a short quote, and which Non-Goals bullet it conflicts with as the why_it_matters.`,
  },
  {
    key: 'structure',
    prompt: `${ROOT_NOTE} Read the ADR in full. Check two separate things. Do NOT check whether cited ADR numbers exist — that is handled elsewhere.

First, self-containment (tag findings here gate: "Self-Containment") — this ADR must stand on its own. For every citation in the document to another document (any file, ticket, note, or artifact that is not itself an ADR in the same directory), ask: is this offered as evidence the decision is correct, or as a description of what results from the decision? Citing another document's current wording as support for the decision is always a finding, because that document's content can change independently of this ADR without anyone revisiting it — a reader has no way to tell from the ADR alone that the citation has drifted. Stating that something in another document is a downstream consequence of this decision is fine and not a finding. Exception: an outside source (a published format specification, or the source code or manual of another project) can be evidence for the decision if the ADR cites it by URL and pins the version that it read, such as a release, a commit, or a wiki revision. Such a citation is not a finding. An outside source that has no URL or no pinned version is a finding.

Second, structural soundness (tag findings here gate: "Structure") — a subsection whose heading claims an argument (e.g. "X is cheaper than Y") that its own body never actually makes; a subsection substantially redundant with another subsection, restating a point already made elsewhere without adding anything new; a property presented as if specific to this decision when it is actually a general property of every ADR in this project (check for a README or conventions file in the same directory as this ADR, and read its stated conventions before flagging this); a claim in one section undercut or contradicted by another section.

For each finding, give the gate, summary, exact location, a short quote, and why it matters (for self-containment findings, why_it_matters should state what happens if the cited document's content changes or is never written).`,
  },
]

phase('Precheck')

const precheckResult = await agent(PRECHECK_PROMPT, { label: 'precheck:adr-numbers', phase: 'Precheck', schema: FINDINGS_SCHEMA })
const precheckFindings = (precheckResult && precheckResult.findings) || []

log(`${precheckFindings.length} mechanical candidate(s) from precheck`)

const precheckVerified = await parallel(
  precheckFindings.map(f => () =>
    parallel(
      Array.from({ length: 2 }, () => () =>
        agent(
          `${ROOT_NOTE} A candidate finding claims: "${f.summary}" at ${f.location}, quoting: "${f.quote || ''}". This is a purely factual claim about whether a referenced ADR file exists at the expected path. Verify it yourself directly — list the directory, check the specific expected filename. Confirm or refute based only on what you actually find; do not guess. refuted=true if the finding turns out to be wrong (e.g. the file does exist, or it isn't actually a forward reference), refuted=false if you confirm the finding is accurate.`,
          { label: 'precheck-verify:self-containment', phase: 'Precheck', schema: VERDICT_SCHEMA }
        )
      )
    ).then(verdicts => {
      const valid = verdicts.filter(Boolean)
      const bothAgreeReal = valid.length === 2 && valid.every(v => !v.refuted)
      return { ...f, dimension: 'precheck', path: 'precheck', survives: bothAgreeReal, verdicts: valid }
    })
  )
)

phase('Find')

const dimensionResults = await parallel(
  DIMENSIONS.map(d => () => agent(d.prompt, { label: `find:${d.key}`, phase: 'Find', schema: FINDINGS_SCHEMA }))
)

const allFindings = dimensionResults
  .filter(Boolean)
  .flatMap((r, i) => (r.findings || []).map(f => ({ ...f, dimension: DIMENSIONS[i].key })))

log(`${allFindings.length} candidate findings across ${DIMENSIONS.length} dimensions`)

phase('Verify')

const verified = await parallel(
  allFindings.map(f => () =>
    parallel(
      Array.from({ length: 2 }, () => () =>
        agent(
          `${ROOT_NOTE} You are adversarially reviewing a candidate finding about this ADR. The finding claims: "${f.summary}" at ${f.location}, quoting: "${f.quote || ''}". Its stated consequence: "${f.why_it_matters}". Read the actual current file yourself, and check any code or docs it references, then genuinely try to REFUTE this finding — argue against it being a real, worthwhile issue in the document's CURRENT state (it may already have been fixed). If after a real attempt you cannot find a good reason to refute it, say refuted=false. Default to refuted=true if uncertain or if the file no longer shows the issue.`,
          { label: `verify:${f.dimension}`, phase: 'Verify', schema: VERDICT_SCHEMA, model: 'opus', effort: 'high' }
        )
      )
    ).then(verdicts => {
      // Two skeptics, so a single refusal kills the finding. The full run uses
      // three and a majority; with two, requiring both to refute would let a
      // finding survive on one skeptic's doubt, which is the wrong default for
      // a check whose whole purpose is to suppress plausible-but-wrong findings.
      const valid = verdicts.filter(Boolean)
      const refutedCount = valid.filter(v => v.refuted).length
      return { ...f, path: 'adversarial', survives: valid.length > 0 && refutedCount === 0, verdicts: valid }
    })
  )
)

const allVerified = [...precheckVerified, ...verified]
const confirmed = allVerified.filter(f => f.survives)
const killed = allVerified.filter(f => !f.survives)

log(`${confirmed.length} confirmed, ${killed.length} refuted out of ${allFindings.length + precheckFindings.length} candidates`)

return {
  adrPath: ADR_PATH,
  confirmed,
  killed,
  totalCandidates: allFindings.length + precheckFindings.length,
}
