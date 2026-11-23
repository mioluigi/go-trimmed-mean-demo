# Prompt for Gemini: repository review R2 (SFEDD §7.2)

**Upload one file:** `docs/reviews/002-repository-review-bundle.txt`. It holds files from **both**
repositories, each after a `===== FILE: path =====` header. Paths starting `../trimmedmean/` are the package
repository `github.com/mioluigi/trimmedmean` at commit `6fd1a6e`; all other paths are this repository,
`github.com/mioluigi/go-trimmed-mean-demo`, at the commit on the bundle's first line.

Then paste everything below the line.

---

Before anything else, list every file you received inside the bundle, by name, and the commit on its
first line. The bundle must contain 22 files (listed at its top). If any is missing, stop and say which.

You are two reviewers at once: a **strict grader** applying the rubric in `docs/assignment-and-rubric.md`,
and a **senior Go engineer who has published Go modules and knows R's `mean(x, trim)`**. The project is
finished: a Go package computing symmetric and asymmetric trimmed means (`../trimmedmean/`), and a second
repository whose `main.go` imports it and checks it against R. **The deadline is tonight, 11:59 PM ET**; the
package will be released as `v1.0.0` after this review, and fixes must fit in about an hour.

Report, in this order:

1. **Rubric scores**: for each criterion ([C], [T], [D], [Doc], [A]), a score out of 20 with the evidence,
   quoting file and section. Say what would be needed for full marks where you give less.
2. **Assignment steps S1–S9**: for each, met or not, with the file and section that shows it. In particular:
   is the package in its own repository and built as the tutorialspoint instructions describe (with the
   departures the spike explains)? Are there **complete instructions for getting and using the package**?
   Does each repository have a README? Is the Go-versus-R summary there?
3. **Code review** of the risky parts: `../trimmedmean/trimmedmean.go` (counts, R's two means, the overflow
   fallback, validation, generics), `internal/compare/compare.go` (the rule of spec §7 and the ULP
   distance), `main.go` and `comparison.go` (exit codes, atomic write, error messages). Report any defect as
   a **test that would fail**.
4. **Evidence check**: check every statement and number in both READMEs against `results/tables.md`,
   `docs/mutation-checks.txt` and the spike; list any that do not match, are stale, or are claimed without
   evidence in the bundle.
5. **Grader friction**: anything a grader would find missing, confusing or hard to run (for example the
   package version named in the READMEs versus the one in `go.mod`).
6. **The three changes under 20 minutes each** that would raise the score most.

For each finding give: **severity** (High / Medium / Low), **where** (file and section; quote the text),
**problem**, **suggested change**, expressed as a property a test or a reading can check. **For every claim,
quote the file and section it relies on.** Finish with a verdict: *ready to submit*, *submit after the
listed fixes*, or *not ready*.
