# Review R2: both repositories — triage

| Field | Value |
|---|---|
| Reviewer | Gemini (web, Gemini Ultra), 2026-11-22 ~19:45 ET, new chat |
| Input | `docs/reviews/002-repository-review-bundle.txt` (22 files from both repositories; program `58db9ca`, package `6fd1a6e`), prompt `002-repository-review-prompt.md` |
| Raw answer | `docs/reviews/002-repository-review-gemini.txt`; PDF in `docs/ai-dialogs/gemini-r2-repository-review.pdf` |
| Receipt check | **Passed**: all 22 files and commit `58db9ca` listed |
| Scores | [C] 18 · [T] 17 · [D] 18 · [Doc] 17 · [A] 17 = **87 / 100** |
| Verdict | *Submit after the listed fixes* |

Most deductions are the steps already planned after R2 (spec Q6, tasks T15): tag `v1.0.0`, move the program
to it, re-measure, run acceptance on fresh clones. Testable claims were tested first.

| # | Sev. | Finding (short) | Decision | Reason / change |
|---|---|---|---|---|
| [C] | — | `parseReference` reads columns by magic positions | **Accept** | Columns are now found by header name; a missing column or a bad number is named in the error. 🔴 test first (`TestReadReferencesByColumnName`, a reordered file with an extra column) failed, then passed. The same pattern in `cmd/tables` is kept: it reads a file this program writes, tested on a real one. |
| [C] | — | gofmt / vet not confirmed across both repositories | **Already met** | `scripts/acceptance.sh` runs `gofmt -l` and `go vet` on fresh clones of both (not in R2's view because its result comes from the Mac run). |
| [T], [A] | — | profiles, build scripts, `docs/ai-dialogs/` not in the bundle | **Adjust** | They exist in the repository (left out to keep the bundle short). Acceptance now also checks that the CPU and memory profiles of the 1,000,000-value run are committed (tested on the real files and on an edited copy). |
| [T], [A] | — | "`TestInputErrors` tests floats-100.csv rather than a 1,000,000-row profile run" | **Reject** | Input-error tests need small inputs; profiling is `scripts/measure.sh`'s job (spec NFR-3), not a unit test's. |
| [D], 3 High, 4 High, 5 High | High | `go.mod` requires `v0.1.0` while both READMEs say `@v1.0.0` | **Accept (planned)** | T15 on the Mac: tag `v1.0.0`, `go get …@v1.0.0`, re-measure, acceptance. Acceptance already fails if `go.mod`, the README and `results/machine.txt` disagree on the version. |
| [Doc], 5 Med | Med | the acceptance headings "are not matched verbatim" in the package README | **Reject — inaccurate** | Every heading acceptance checks exists verbatim in `../trimmedmean/README.md`; the reviewer's own list in §5 shows it. An extra `## Definition` does not break the check. Recorded below. |
| 3 Med | Med | `isInteger` relies on truncation | **Already met** | Correct for every type in `Number`; `TestMeanIntegerAndFloatPaths` checks both paths and mutation M15 shows the test catches a wrong detection. The reviewer's suggested test is that existing test. |
| 3 Med | Med | full-sort exactness could fail by 1 ULP elsewhere; `meanAbsolute` of an empty slice is NaN | **Accept in part** | Exactness against R's full-sort mean is the rule by design (spec §7), shown on the Mac by both the package's R test and the official run; its four cases are already tested (`TestCheckToleranceRule`). The empty case now returns 0 (limit 0, never NaN): 🔴 `TestCheckWithNoKeptValues` failed, then passed. |
| 3 Low | Low | a failed run might leave a temporary file | **Accept, tested** | `TestComparisonFailure` now checks the output folder holds only the table after a failed comparison: passes. |
| 4 Low | Low | `results/machine.txt` names `2304f25`, the bundle `58db9ca` | **Accept (planned)** | Expected: the results were measured at `2304f25` and the later commits changed documents only (acceptance checks this). The final run on `v1.0.0` measures the submitted commit. |

**Inaccurate reviewer statements recorded** (SFEDD §5.4): the acceptance heading mismatch ([Doc] and §5),
contradicted by the reviewer's own listing; "`floatMean` converts values after `toFloat64` already
converted" (§3: `floatMean` receives `[]float64` and converts nothing).
