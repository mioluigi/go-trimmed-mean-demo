# Review R1: spec 001 — triage

| Field | Value |
|---|---|
| Reviewer | Gemini (web, Gemini Ultra), 2026-11-22 ~18:10 ET |
| Input | `docs/reviews/001-spec-review-bundle.txt` (5 files, commit `bddebe3`), prompt `001-spec-review-prompt.md` |
| Raw answer | `docs/reviews/001-spec-review-gemini.txt` (text of the PDF Luis exported) |
| Receipt check | **Passed**: all 5 files and commit `bddebe3` listed |
| Verdict | *Approve with changes* |

Every finding gets a decision (SFEDD §2.5). Testable claims were tested first; tests are in `trimmedmean`
branch `pre-approval`, commit `ddb3e94`.

| # | Sev. | Finding (short) | Decision | Reason / change |
|---|---|---|---|---|
| 1.1 | High | READMEs: no concrete get/import/call/error instructions required; `docs/ai-dialogs/README.md` not required by the spec | **Accept** | FR-R4 and new FR-13 list the exact content (`go get …@<tag>`, `import`, a call, `errors.Is`); the README snippet is the same code as the `Example` functions, which `go test` compiles and checks. AC-13 adds `docs/ai-dialogs/README.md`. |
| 1.2 | High | S6 sample sizes and constitution 6.4 output not in FR-8 | **Accept** | FR-8 now requires at least one integer and one float sample of ≥ 100 values (the data gate already enforces ≥ 100 per file) and says what a finished run prints. AC-9 checks stdout for sample, value and R status. |
| 1.3 | Med | Profile "on the largest sample"; guards tested on pass and fail | **Accept** | NFR-3 names the largest sample; new AC-15: each script and guard has a pass and a fail test (constitution 3.5). |
| 1.4 | Med | Doc-comment coverage not checked | **Accept, tested** | New AC-16: a `go/ast` test requires a doc comment on every exported name (caught a planted undocumented function). |
| 2.1 | High | ±Inf "allowed if finite" contradicts `ErrNotFinite`; FR-4 "names the value" contradicts fixed messages in §8.1 | **Adjust** | Not a contradiction: ±Inf values are allowed and the error is about the *result* (already tested: `[-Inf, 1…8, +Inf]` trimmed 0.1 gives 4.5; untrimmed gives `ErrNotFinite`). Wording of FR-4 and §8.1 made explicit; §8.1 now says its messages are the **program's**, and the package's are `trimmedmean: …: got 0.6`. |
| 2.2 | Med | No test for `Nextafter(0.5, 0)` at large n | **Accept, tested** | Tests added; they show the rounding worry **cannot happen below 2^53** (n·p stays more than half an ulp below n/2). Spec v1 FR-5 claimed it could: **corrected** (our own error, logged). Guard kept as a defensive check. |
| 3.1 | High | "trust", "promote", "UX/UI", "short functions" untestable | **Adjust** | §1–§2 are context, not requirements. "Short functions" made checkable: AC-16 also limits functions to 40 lines (tested on a planted 44-line function). UX is checked through AC-9/AC-10 (exact messages and exit codes). |
| 3.2 | Med | "never panics", "no internet / no R" cannot be proved | **Accept as properties** | AC-4 states the fuzz evidence (30 s, no failure) rather than a proof; acceptance runs `go test` with `GOPROXY=off` and `Rscript` removed from `PATH`. |
| 4.1 | High | No `low + high < 1` requirement | **Reject, tested** | Each proportion < 0.5 implies the sum < 1, and the counts always keep a value: tested for every n in 1…105 and proportions up to `Nextafter(0.5, 0)`. The reviewer's side claim that counts "can be asymmetric even when low == high if float64 rounding differs" is **inaccurate**: the same expression on the same inputs gives the same result. |
| 4.2 | High | Integers > 2^53; exact full-sort agreement may fail by 1–2 ulps on aarch64 | **Accept the test, record one inaccuracy** | The > 2^53 limit is already in §8.1 and the README. Exactness on the Mac is now an acceptance check (the package's R test runs on the Mac, AC-3). The statement that R's integer mean "can overflow or lose precision differently" is **inaccurate**: R integers are 32-bit and R sums them in a long double (spike addendum B reproduces all 42 rows with one exact sum). |
| 4.3 | High | Module: no `replace`, no `go.work`, clean `GOMODCACHE` build | **Accept** | AC-8 now names the checks explicitly. |
| 4.4 | Med | Variadic arity at run time; NaN proportions; overflow row in §8.2 confusing | **Accept** | NaN, ±Inf and 3-argument cases are tested (`TestMeanErrors`). The overflow row moves from "differences" to a "same as R" note. Run-time arity is the cost of Q1, stated there. |
| 5 | — | Addendum A cause not shown from R's source; P4 not yet shown on the Mac; P8 output not quoted | **Accept** | Addendum B says "Held" for the *observation*; the *cause* stays a hypothesis (wording clarified). Go-on-Mac exactness is AC-3 on the Mac. P8's output line is quoted in addendum B. |
| Q6 | — | Prefer `v0.1.0` now, `v1.0.0` after R2 | **Accept (recommendation changed)** | Costs one extra `go get` on the Mac, folded into the final acceptance step; the firm's submitted offering is still `v1.0.0`, and no tag is ever moved. |
| 7 | — | Scope cuts | **Accept** | Matches the plan: bootstrap deferred; README polish last. |

**Inaccurate reviewer statements recorded** (SFEDD §5.4): 4.1 (asymmetric counts from identical inputs),
4.2 (R's integer mean overflowing).
