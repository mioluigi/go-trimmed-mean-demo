# Prompt for Gemini: spec review R1 (SFEDD §7.1)

**Upload one file:** `docs/reviews/001-spec-review-bundle.txt`. It contains five files, each after a
`===== FILE: path =====` header: `docs/assignment-and-rubric.md`, `docs/constitution.md`,
`docs/spikes/001-environment-and-r-reference.md`, `testdata/r-reference/semantics.tsv`,
`specs/001-trimmed-mean/spec.md`.

Then paste everything below the line.

---

Before anything else, list every file you received inside the bundle, by name, and the commit named on its
first line. The bundle must contain these five: `docs/assignment-and-rubric.md`, `docs/constitution.md`,
`docs/spikes/001-environment-and-r-reference.md`, `testdata/r-reference/semantics.tsv`,
`specs/001-trimmed-mean/spec.md`. If any is missing, stop and say which.

You are two reviewers at once: a **strict grader** applying the rubric in the bundle, and a **senior Go
engineer who has published Go modules and knows R's `mean(x, trim)` and floating-point arithmetic**. Review the
draft spec (with its assignment summary, constitution, spike and R evidence) for a project that publishes a
Go package computing a symmetric or asymmetric trimmed mean, and a second repository whose `main.go` imports
it and checks it against R. **The deadline is tonight, 11:59 PM ET**: favour findings that change the grade
or prevent a defect.

Report, in this order:

1. **Coverage gaps**: any assignment step (S1–S9) or rubric criterion ([C], [T], [D], [Doc], [A]) not
   covered, or covered only weakly, including "complete instructions for getting and using the package" and
   "a README for each repository".
2. **Contradictions** between spec, constitution, spike and assignment summary.
3. **Untestable requirements**: anything a test, a script or a reading of a file cannot check.
4. **Domain risks**: the definition of the trimmed mean (spec §3) and the `floor(n × p)` rule; the
   asymmetric definition; integer versus float paths and precision; the comparison rule in spec §7 and
   whether its tolerance is justified by the spike; the deliberate differences from R (spec §8.2); the API
   (generics, variadic proportions, errors); module publishing and versioning (tags, the module proxy,
   `go.sum`); and the Limits and Consequences section (spec §8).
5. **Evidence check**: is each spike finding (addenda A and B) supported by the evidence it cites? Is any
   prediction unjustified or missing its conditions?
6. **A view on each open decision Q1–Q7**: agree with the recommendation or argue for another option.
7. **Scope cuts** if the deadline is at risk, in the order you would cut them.

For each finding give: **severity** (High / Medium / Low), **where** (file and section; quote the text),
**problem**, **suggested change**. Express worries as **properties a test can check**, not as a required
mechanism. For every claim you make about the documents, quote the file and section you rely on.

Finish with a verdict: *approve*, *approve with changes*, or *rework*.
