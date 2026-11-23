# Spec 002: Bootstrap study of the trimmed mean (assignment S8, optional)

| Field | Value |
|---|---|
| Status | **v1 — APPROVED** by Luis, ~20:30 ET (as recommended) · **all ACs met** 21:00 ET: `results/acceptance-mac.txt` (26/26 on fresh clones), official run at `600cbde` |
| Governs by | `docs/constitution.md`; spec 001 stays unchanged and approved |
| Source of the design | the course file `run-bootstrap-robust.R` (Thomas W. Miller, revised May 23, 2024), given by Luis; Luis's listing of a run with B = 100 (`listing-from-100-run-bootstrap-robust.txt`) |

## 1. Problem

S8 asks for a bootstrap study "showing the value of the trimmed mean as a robust estimator of central
tendency", based on the course's `run-bootstrap-robust.R`. Reproducing the **course study itself** in Go,
draw for draw, and checking it against R, turns "Go agrees with R" from one function into a whole
statistical study — and uses the package where an analyst would.

## 2. The course study, as the file actually does it (read, not assumed)

- Seed `set.seed(9999)`, **once, before the loops** (the comment says "set inside the for-loop"; the code does
  not: a finding, §6).
- For iteration 1…100 and each n in 25, 100, 225, 400: draw n observations from a mixture, each by
  `runif(1) >= 0.01 ? rnorm(1, 100, 10) : rnorm(1, 200, 1)` (1% outliers near 200).
- For b in 1…B (B = 100): `sample(thisSample, n, replace = TRUE)`; record its mean, median and
  `mean(x, trim = 0.05)`. Then average the B values of each statistic.
- Report, per n: the CLT value 10/√n; `sd` over the 100 iterations of the **sample mean**; and `sd` over the
  100 iterations of the **average of the B bootstrap means / medians / trimmed means**, each rounded to
  2 decimals.

## 3. Requirements

| ID | Requirement |
|---|---|
| FR-B1 | A Go program `cmd/bootstrap-study` reproduces §2 exactly: same seed, same draws, same order, B = 100 by default (`-B` to change), printing the course's five lines per n, rounded as the course rounds. |
| FR-B2 | Random numbers reproduce R 4.4.2's defaults: Mersenne-Twister `runif` and `sample` (Week 8's `rrand`, reused unchanged, its tests included) and **`rnorm` by inversion** (R's `norm_rand`: two uniforms, then `qnorm`), new in this spec. |
| FR-B3 | Statistics: the mean as R computes it; the median as R computes it; the trimmed mean **by `trimmedmean.Mean(x, 0.05)` v1.0.0** — the package under study. |
| FR-B4 | The program also reports, as an **added, labelled column (Go only: the course file does not keep the B values, so R cannot be compared here)**, the textbook bootstrap standard error: for each sample, the standard deviation of its B bootstrap values, averaged over the 100 samples (§6 explains why). |
| FR-B5 | Results to `results/bootstrap-study.tsv`; README section with a generated table and Go-versus-R verdict. |
| FR-B6 | The course file is **not committed** (it is course material); a committed wrapper `r/capture-bootstrap-reference.R` runs Luis's copy **unchanged** on the Mac and saves R's full-precision results, the file's SHA-256 and the R session. |

## 4. Comparison with R (fixed before measuring)

| Item | Rule | Why |
|---|---|---|
| `rnorm` draws (5,000 after `set.seed(9999)`, and 2,000 mixture draws as the study makes them) | **exact**, or the measured gap stated if not | inversion is deterministic; a fused multiply-add (FMA) in R's or Go's compiled `qnorm` could change the last bit, which only the Mac run can show |
| The 16 standard errors (4 per n: sample mean, bootstrap mean, median, trimmed mean; the CLT line is a formula), full precision | **relative difference ≤ 1 × 10⁻⁹** | resample indices depend only on uniforms (exact), so differences can only come from the trimmed means' partial-sort order and `qnorm` rounding, each ≤ a few ULPs per value |
| The same 16 values rounded to 2 decimals | **identical** to R's output and to Luis's listing | what the course prints |

## 5. Limits and consequences

- Reproduction holds for R's default generators (`RNGkind`: Mersenne-Twister, Inversion, Rejection); the R
  session is recorded and the Go tests name it.
- The study's run time grows as 100 × Σn × B; `-B` above 10,000 is rejected (exit 2) to bound run time; bad
  flags exit 2 with a message, as in spec 001.
- Nothing personal is read or written; the course file stays on Luis's Mac.

## 6. Findings from reading the course file (to be confirmed by the runs)

1. **What the "bootstrap SE" lines measure.** They are the standard deviation, across 100 independent
   samples, of each statistic **averaged over its B bootstrap resamples**. That is a Monte Carlo estimate of
   the sampling variability of a bootstrap-averaged estimator, close to the statistic's own sampling standard
   deviation, not the usual bootstrap estimate from one sample (the spread of the B values themselves).
   FR-B4 reports the usual one beside it; both should tell the same story if the reading is right.
2. **The seed comment does not match the code** (§2). Harmless for reproducibility; noted, not changed.
3. **Prediction (stated with its conditions):** with 1% outliers at +10 SD, the trimmed mean's standard
   error is below the mean's at every n and close to or below the median's from n = 100 on — as Luis's
   listing shows for B = 100 (n = 100: mean 1.42, median 1.11, trimmed 1.02).

## 7. Acceptance and evidence

| AC | Criterion | Test |
|---|---|---|
| AC-B1 | Go's `rnorm` reproduces R's 5,000 and 2,000 saved draws (exactly, or within the measured gap) | unit test on `testdata/bootstrap-reference/` |
| AC-B2 | Go's 16 standard errors meet §4 against R's saved values | unit test + program run on the Mac |
| AC-B3 | Rounded output identical to R's captured listing and to Luis's listing | program test |
| AC-B4 | `rrand`'s Week 8 tests pass unchanged in this repository | `go test` |
| AC-B5 | Mutation check: `qnorm` coefficient, inversion step, median for even n, trimming proportion — all caught | `docs/mutation-checks.txt` addendum |
| AC-B6 | spec 001's acceptance still passes, plus the study's checks | `scripts/acceptance.sh` on fresh clones |

**Acceptance status (2026-11-22 21:00 ET): AC-B1 … AC-B6 met.** AC-B1: 5,000 + 2,000 draws identical to R's on the
Mac. AC-B2: 400 rows (mean and median columns identical, trimmed within 1.4e-16), 16 SEs within 2e-15. AC-B3:
the official run's listing equals R's character for character. AC-B4: Week 8's `rrand` tests pass here. AC-B5:
9 mutants, 8 caught, 1 equivalent. AC-B6: acceptance 26/26. The first Mac acceptance failed AC-B1/B2 on
arm64 fusion; fixed (spike 001 addendum D) before this run.

Under the deadline, no third Gemini review is planned: the evidence is the R comparison itself, the mutation
check and acceptance. If time remains after acceptance, a short Gemini review of this addendum is optional.
