# Spike 001: environment, R's trimmed mean, and publishing a module

| Field | Value |
|---|---|
| Date | 2026-11-22 |
| Phase | 1 (spike), started during phase 0 under the same-day deadline |
| Parts | **Workspace part** (Claude, done) · **Mac part** (`scripts/env-probe.sh`, `scripts/capture-r.sh`; results in `docs/spikes/env-mac.txt` and `testdata/r-reference/`) |

## 1. Environment probe, workspace part (SFEDD §5.7)

| Check | Result | Consequence |
|---|---|---|
| Go | go1.24.7 linux/amd64 | Workspace builds and tests; `go.mod` says `go 1.24` |
| R | **absent** | R reference values come from the Mac as saved files (L-W8-12c) |
| GitHub read (`git clone`) | ok | Can read last week's repository and our own |
| GitHub push | pending: repositories not created yet | Phase 0 gate |
| `proxy.golang.org`, `sum.golang.org` | **refused (HTTP 403)** | The workspace cannot `go get` the published package; the program's `go.sum` is generated on the Mac (L-W7-06) |
| tutorialspoint (shell) | refused (HTTP 403); read through the web-fetch tool instead | See §2 |

The Mac part (last probed 2026-11-14: Go 1.27.2, R 4.4.2, git, **no `gh`**) is refreshed by
`scripts/env-probe.sh`.

## 2. The tutorialspoint instructions (S1)

Source (given by Luis): <https://www.tutorialspoint.com/article/how-to-create-your-own-package-in-golang>,
"How to Create Your Own Package in Golang?", by Siva Sai, updated 2023-05-05. Read 2026-11-22 through the
web-fetch tool (the shell cannot reach the site). Its seven steps, and what we do with each:

| # | Tutorial step | We | Reason |
|---|---|---|---|
| 1 | `mkdir $GOPATH/src/mathutils` | **Depart**: the package lives at the root of its own repository | GOPATH layout is legacy; with modules (default since Go 1.16) code may live anywhere |
| 2 | Source files named after the package, e.g. `average.go` with `package mathutils` and an exported function | **Follow**: `trimmedmean.go`, `package trimmedmean`, exported functions | — |
| 3 | `go mod init mathutils` | **Adapt**: `go mod init github.com/mioluigi/trimmedmean` | A bare module path cannot be fetched by the `go get` of step 7; the path must be the repository address |
| 4 | `package` keyword at the top of each file | **Follow** | — |
| 5 | `go build` (said to produce an executable) | **Follow, corrected**: `go build ./...` checks the library compiles; the executable comes from the program repository | A non-`main` package does not produce an executable |
| 6 | `_test.go` files, `Test…` functions, `go test` | **Follow**, plus `Example…` and `Benchmark…` functions | Rubric [T]; runnable examples document the API |
| 7 | Publish on GitHub; consumers run `go get github.com/your-username/your-package` | **Follow, completed**: push, then a semantic-version tag (e.g. `v1.0.0`); `go get github.com/mioluigi/trimmedmean@v1.0.0` | The tutorial does not mention tags; a tag gives `go get` an exact, verifiable version |

The tutorial does not explain that capitalisation exports a name, does not show importing the package from
another program, and does not mention `go install`. The package README covers all three, because S7 asks
for complete instructions for getting and using the package.

## 3. R's trimmed mean: what we expect, and the conditions (to be checked on the Mac)

From R's `mean.default` (to be confirmed by running it, constitution 1.2 and 9.3):

| # | Prediction | Condition it assumes | Checked by |
|---|---|---|---|
| P1 | R drops `floor(n * trim)` observations from **each** end, computed in float64, so `n = 100, trim = 0.29` drops **28**, not 29 (Go's float64 gives `100 * 0.29 = 28.999999999999996`, measured in the workspace) | R 4.4.2 `mean.default` | `semantics.tsv` rows `floor(...)`; `dropped_low` column |
| P2 | `trim >= 0.5` returns the **median**; a **negative** trim is silently ignored (no trimming); `NA` in the data gives `NA`; empty data gives `NaN`; a trim of length 2 is an error | same | `semantics.tsv` |
| P3 | **Integers agree exactly.** For integer samples the kept values sum exactly in float64, so the order of summation cannot matter | values well below 2^53 | `trimmed-means.tsv`, integer rows: Go value identical to R value |
| P4 | **Doubles may not agree to the last bit**, which is the riskiest assumption. R takes the kept values from a **partial sort** (`sort.int(x, partial = c(lo, hi))[lo:hi]`), so it adds them in an order that is not fully sorted; a Go implementation that sorts fully adds them in another order, and a floating-point sum depends on order. R's second pass (adding the mean deviation) should shrink any difference to a few units in the last place | arm64 Mac, where R's long double is 64 bits (`sizeof long double` in `r-session.txt`); Go uses R's two-pass mean | column `identical_to_full_sort` (R's built-in against R's own mean of the fully sorted kept values) |
| P5 | Integer and double input give the same R result for the same values | same | column `identical_as_double` |
| P6 | `mean(c(1e308, 1e308))` is `Inf` on the Mac (64-bit long double overflows) but would be `1e308` on an x86 machine with 80-bit long double: a platform difference, and a documented limit | arm64 | `semantics.tsv` |

**Decision rule for the spec, fixed now (before seeing the results).** If every double row has
`identical_to_full_sort = TRUE`, the comparison rule is **exact agreement** for all samples. If any row is
`FALSE`, the rule is exact for integers and, for doubles, agreement within a tolerance **measured** from those
rows (the largest relative difference found, times a safety factor stated in the spec), unless the spec
chooses to reproduce R's partial sort (open decision for the spec).

## 4. Publishing a module (S1, S5)

| # | Prediction | Condition | Checked by |
|---|---|---|---|
| P7 | After `git tag v… && git push --tags` on the Mac, `go get github.com/mioluigi/trimmedmean@v…` works from a fresh module on the Mac within a minute | public repository; default `GOPROXY`, `GOSUMDB`; no `GOPRIVATE` match | Build phase, task "release"; acceptance |
| P8 | The Mac already fetches `mioluigi` modules through the public proxy | as P7 | `env-mac.txt`: `go list -m github.com/mioluigi/go-applied-statistics@latest` |

## Addendum A (2026-11-22 17:55): the first R capture stopped

- **What happened.** `capture-r.sh` stopped at `stopifnot(identical(values_read, values))`: a sample written
  with `%.17g` and read back by `as.numeric` was not identical to the values R generated. Nothing was committed.
- **Hypothesis, not yet shown.** R reads numbers with its own parser (`R_strtod`), which accumulates digits in
  a long double. The Mac probe shows `sizeof long double: 8`, so on this arm64 Mac the accumulation is plain
  double and a 17-digit string can come back one unit in the last place away. `testdata/r-reference/parse-check.tsv`
  counts, per float sample, the values whose 17-digit round trip fails, as evidence.
- **Change.** Floats are written with 4 decimals, as in Week 8; R computes from what it **read**; and R saves
  those doubles at 17 digits (`sprintf` is exact) in `testdata/r-reference/*-as-read-by-r.txt`, so a Go test
  proves Go's `strconv.ParseFloat` reads the same doubles from the sample files. Constitution 8.3 amended.
- **Lesson candidate.** A round-trip check that compares with the generator's memory tests the generator,
  not the shared file; compare what each side reads.

## Addendum B (2026-11-22 18:05): the Mac results, prediction by prediction

Evidence: commit `835c248` (`docs/spikes/env-mac.txt`, `testdata/r-reference/`), analysed in the workspace
with Python (correctly rounded parsing, like Go's `strconv`).

| # | Prediction | Outcome | Evidence |
|---|---|---|---|
| A | (addendum A) R's parser fails some 17-digit round trips on this Mac | **Observation held** (the cause, R's 8-byte long double in `R_strtod`, remains a hypothesis: not shown from R's source; R1 §5): 28 of 100, 23 of 101, 268 of 1,000 values. With 4 decimals, R reads every value exactly as a correctly rounded parser does (0 differences of 1,201) | `parse-check.tsv`; `*-as-read-by-r.txt` |
| P1 | `floor(n * trim)` in float64; `100 × 0.29` drops 28 | **Held** (28, 7, 5, 5) | `semantics.tsv`; `dropped_low` |
| P2 | trim ≥ 0.5 → median; negative ignored; NA → NA; empty → NaN; length-2 trim → error | **Held**: `trim = -0.1` gives 22 (the untrimmed mean); 0.5 and 0.7 give the median 3; NA and NaN give NA; empty gives NaN; `c(0.1, 0.2)` and `NA` give "'trim' must be numeric of length one" | `semantics.tsv` |
| P3 | Integers: order cannot matter | **Held**: 24 of 24 integer rows identical to the full-sort mean | `identical_to_full_sort` |
| P4 | Doubles: R's partial sort changes the last bits | **Held** (Go's own agreement is shown in the workspace by `trimmedmean` T03; on the Mac by acceptance AC-8): 20 of 24 double rows differ from R's own full-sort mean, by up to 303 ULPs (the mean is near zero, so ULPs are small); **scaled by the mean absolute kept value, the worst gap is 4.7 × 10⁻¹⁶**, about two machine epsilons | `trimmed-means.tsv`; analysis above |
| P5 | Integer and double input give the same R result | **Failed**: 12 of 24 integer rows differ from the same values as doubles. R averages integers with **one exact sum divided by n**, doubles with its two-pass method. The one-pass rule reproduces all 42 integer rows exactly | `identical_as_double`; single-pass check |
| P6 | `mean(c(1e308, 1e308))` overflows to Inf on this Mac | **Failed**: R returns the value (`1.0000000000000006e+308`, which is how R's parser reads `1e308`). R has an **overflow fallback**: when the sum is infinite it re-adds `x/n` term by term | `semantics.tsv` |
| P7 | Tag → `go get` from a fresh module works | pending (T07) | — |
| P8 | The Mac fetches `mioluigi` modules through the proxy | **Held**: `own module through proxy (go list -m …/go-applied-statistics@latest): ok · github.com/mioluigi/go-applied-statistics v0.0.0-20261115112517-44d3b0355db7` | `env-mac.txt` |

**Consequences for the spec (v1.1, before R1).** Integers use R's one-pass exact mean; doubles use R's
two-pass mean with R's overflow fallback. Comparison rule: **exact** for integers, for untrimmed doubles (Go,
like R, does not sort when nothing is trimmed) and for every asymmetric row (both sides sort fully); trimmed
symmetric doubles must equal **R's full-sort mean exactly** and R's built-in within **10⁻¹⁴ × the mean absolute
kept value** (20 times the measured worst gap). Reproducing R's partial sort is not worth tying the package to
R's internals for gaps of about two epsilons (Q7).

## 5. Pre-mortem: "It is after midnight and this project failed. Why?"

| # | Cause | Guard |
|---|---|---|
| 1 | **Ran out of time**: two repositories and two reviews in about six hours | Optional S8 (bootstrap) deferred until the required work is done; a phase clock in the task list; a **minimum submittable state by 22:30** (both repositories public, the program runs, both READMEs complete) before R2 |
| 2 | **The program repository cannot fetch the package**: wrong module path, tag not pushed, `go.sum` from the wrong machine, a leftover `replace` | Module path equals the repository address; the release is a task with its own check; `go.sum` from the Mac; acceptance runs `go build` on a **fresh clone with an empty module cache** and `-mod=readonly` |
| 3 | **Go and R disagree, and we can't say why** (P1, P4) | Spike evidence decides the comparison rule before any code; a comparison program reports exact or ulp differences per row |
| 4 | **Grader friction**: missing "get and use" instructions, no README in one repository, the departure from tutorialspoint unexplained, no `.exe` or `.app` instructions | Traceability to S1–S9 and the rubric; acceptance checks the README sections exist; R2 asks for grader friction |
| 5 | **Bad input panics or gives a silent wrong answer** (empty file, NaN, missing header, a proportion out of range, too many proportions) | Limits and Consequences section; each item is an acceptance criterion tested on the **built binary** |
