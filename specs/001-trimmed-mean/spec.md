# Spec 001: Trimmed mean package and demonstration program

| Field | Value |
|---|---|
| Status | **v2** (2026-11-22, 18:50) after review R1 (`docs/reviews/001-spec-review.md`), for approval |
| Repositories | `github.com/mioluigi/trimmedmean` (package) · `github.com/mioluigi/go-trimmed-mean-demo` (program, documents) |
| Governs by | `docs/constitution.md`; traced to `docs/assignment-and-rubric.md` (S1–S9, [C][T][D][Doc][A]) |
| Evidence | `docs/spikes/001-environment-and-r-reference.md`; R reference in `testdata/r-reference/` (Mac part of the spike) |

## 1. Problem

A Go consulting firm wants a first open-source package to show its craft: a **trimmed mean**, symmetric or
asymmetric, that any Go program can fetch with `go get` and trust. "Trust" means two things here: it agrees
with the reference implementation analysts already use (R's `mean(x, trim)`), and it refuses inputs where a
trimmed mean is undefined or misleading instead of returning a quiet wrong number.

## 2. Users

| User | Needs |
|---|---|
| **Go developer** (the firm's prospective client) | `go get` one module with no dependencies; one obvious function; clear errors; documentation with runnable examples |
| **Analyst moving from R** | Same answer as `mean(x, trim)` for symmetric trimming, and a written list of where the package deliberately differs |
| **Grader** | Two repositories, a README in each, instructions to get, use, build, run and test, and the Go-versus-R comparison |

## 3. Definitions

- **Proportions.** `low` and `high` are the proportions of the **sorted** observations removed from the lower
  and upper ends.
- **Counts.** For `n` observations: `dropLow = floor(n × low)`, `dropHigh = floor(n × high)`, computed in
  float64 exactly as R computes `floor(n * trim)`, so that `n = 100, trim = 0.29` drops 28 (spike P1).
- **Trimmed mean.** The arithmetic mean of `sorted[dropLow : n − dropHigh]`, computed **as R computes it**
  (spike addendum B): for **integer** types, one sum (exact below 2^53) divided by the count; for **float**
  types, R's two-pass mean (the first estimate plus the mean of the deviations from it), and, if the first
  sum overflows, R's fallback of adding each value divided by the count. When nothing is trimmed, the values
  are averaged in their given order, without sorting, as R does.
- **Symmetric** when only one proportion is given: `low = high = trim`. With `low = high`, this is R's
  `mean(x, trim)`. **Asymmetric** when two are given; R has no built-in, so this is **our** definition,
  evaluated in R the same way for cross-checking (spike §3).

## 4. Functional requirements — package (`trimmedmean`)

| ID | Requirement | Steps |
|---|---|---|
| FR-1 | `func Mean[T Number](values []T, proportions ...float64) (float64, error)`, where `Number` is every Go integer and floating-point type (including named types based on them). | S1, S2 |
| FR-2 | No proportion → the ordinary (untrimmed) mean; one → symmetric; two → `low`, `high`; more than two → `ErrProportionCount`. | S3, S4 |
| FR-3 | `values` is never modified; the function sorts a copy. | S2 |
| FR-4 | Errors are sentinel values usable with `errors.Is`: `ErrEmpty` (no values), `ErrNaN` (a NaN value), `ErrProportion` (a proportion that is NaN, negative, or not below 0.5), `ErrProportionCount`, `ErrNotFinite` (the **result** is NaN or infinite even after R's overflow fallback, e.g. `+Inf` and `−Inf` both kept; infinite **values** are accepted when they are trimmed away or the result is finite). Each package error message names the offending value or index (`trimmedmean: trim proportion must be in [0, 0.5): got 0.6`). | S2 |
| FR-5 | Because each proportion is below 0.5, `dropLow + dropHigh < n`: at least one observation is always kept: tested for every `n` in 1…105 and proportions up to `Nextafter(0.5, 0)`, and for `n` up to 2^53. *(v2 correction: v1 claimed the float64 product `n × p` could round up to `n/2` for very large `n`; analysis and tests show it cannot below 2^53.)* The code keeps a defensive check that returns `ErrProportion`. | S2, S3 |
| FR-6 | The package prints nothing, logs nothing, never panics on any input, never calls `os.Exit`, holds no state, and is safe for concurrent use. | — |

### Reuse requirements (the firm's offering)

| ID | Requirement |
|---|---|
| FR-R1 | The module depends on the standard library only (`go.mod` has no `require`). |
| FR-R2 | Released by a semantic-version tag; `go get github.com/mioluigi/trimmedmean@<tag>` works from a fresh module on the Mac. |
| FR-R3 | Every exported identifier has a doc comment; at least three runnable `Example` functions (symmetric, asymmetric, error handling) pass under `go test`. |
| FR-R4 | The package README alone is enough to install, call, and understand every error and limit. It contains: the `go get github.com/mioluigi/trimmedmean@<tag>` command; the `import` line; a call with one and with two proportions; error handling with `errors.Is`; the definition (§3); the differences from R (§8.2); the limits (§8.1); how to run the tests and benchmarks; the licence. Its code is the code of the `Example` functions, which `go test` compiles and checks. |

## 5. Functional requirements — program (`go-trimmed-mean-demo`)

| ID | Requirement | Steps |
|---|---|---|
| FR-7 | `main.go` imports `github.com/mioluigi/trimmedmean` at the released tag (no `replace`, no `go.work`). | S5 |
| FR-8 | **With no arguments**, the program reads every committed sample in `data/` — at least one integer sample and one float sample of **at least 100 values** each (the data gate enforces ≥ 100 per file) — with integers as `[]int` and floats as `[]float64`, computes every case in `testdata/r-reference/trimmed-means.tsv` (which includes R's `mean(x, trim = 0.05)` for every sample), compares each with R's saved value under §7, and writes `results/comparison.tsv`. It prints, in plain words: a table for **trim = 0.05** with sample, n, kind, Go value, R value, difference and PASS/FAIL; then one line saying how many of all rows agreed and where the full table is (constitution 6.4). | S5, S6 |
| FR-9 | `-data FILE -trim P` computes one trimmed mean for the user's file (`-trim "0.05"` symmetric, `-trim "0.05,0.10"` asymmetric) and prints it with the counts dropped from each end. | S5 |
| FR-10 | Input files: one number per line, optional header line `value`, blank lines ignored; the kind (integer or float) is detected from the content. | S5 |
| FR-11 | Logs (structured, `log/slog`) go to standard error; results go to standard output. `-cpuprofile FILE` and `-memprofile FILE` write profiles. | — |
| FR-12 | Exit codes: **0** success (and, in comparison mode, every row agrees); **1** a comparison disagrees or a run-time failure; **2** a usage or input error. | — |
| FR-13 | The program README gives, step by step, how to **get and use the package in your own Go program** (`go mod init`, `go get …@<tag>`, `import`, a call, error handling), then how to build, run and test this program; and links `docs/ai-dialogs/README.md`. | S7, S9 |

## 6. Non-functional requirements

| ID | Requirement | Rubric |
|---|---|---|
| NFR-1 | `gofmt`, `go vet` clean; functions short and single-purpose; names specific (constitution Art. 2). | [C] |
| NFR-2 | Package: table-driven unit tests with hand-computed answers; R reference tests from saved files; mutation check of the critical tests with `-count=1`; benchmarks at n = 100, 10,000, 1,000,000 with `-benchmem`. | [T] |
| NFR-3 | Program: tests of every exit code and message on the **built binary**; a CPU and memory profile of a run on the **largest committed sample** kept in `results/`, with its structured log. | [T][A] |
| NFR-4 | Time and memory: `O(n log n)` time and one copy of the input (8 bytes × n) of extra memory, confirmed by the benchmark's allocations. | [T][D] |
| NFR-5 | Build instructions and scripts for macOS (arm64, amd64), a Windows `.exe`, and a macOS `.app`. | [A] |
| NFR-6 | Tests need neither R nor the internet. | [T] |

## 7. Comparison rule with R (fixed before measuring; constitution 1.1)

Settled by spike 001 addendum B (R 4.4.2, aarch64-apple-darwin, 8-byte long double):

- **Exact agreement** (identical float64) for: every **integer** row; every **untrimmed** row; every
  **asymmetric** row (R evaluates our definition with a full sort, as Go does).
- **Trimmed symmetric float rows:** exact agreement with **R's mean of the fully sorted kept values**, **and**
  `|Go − R built-in| ≤ 10⁻¹⁴ × mean(|kept values|)`. Reason: R's built-in adds the kept values in the order its
  partial sort leaves them; the measured worst gap is 4.7 × 10⁻¹⁶ on that scale, so the limit is 20 times it.
- Go must read the samples as the same doubles R read: Go's parse of each sample equals R's saved
  `*-as-read-by-r.txt` value (a test).
- The comparison reports, per row: Go value, R value, difference in units in the last place (ULPs), and
  PASS/FAIL. A single FAIL makes the run exit 1.

## 8. Limits and consequences

### 8.1 Boundaries

Messages in this table are the **program's**; the package returns the matching error (FR-4).

| Item | Supported | Rejected, with message and exit code |
|---|---|---|
| Element types | all Go integer and float types | — (compile-time) |
| Number of values | 1 to memory limits (benchmarked to 10^6) | 0: `ErrEmpty` → program: `no numbers in FILE` · exit 2 |
| Proportions | each in [0, 0.5); one or two of them | NaN, < 0, ≥ 0.5: `ErrProportion` → `trim proportion 0.6 is outside [0, 0.5)` · exit 2; three or more: `ErrProportionCount` · exit 2 |
| Values | any non-NaN number; ±Inf values accepted when the **result** is finite (trimmed away) | NaN value: `ErrNaN` · exit 2; NaN or infinite **result**: `ErrNotFinite` · exit 1 |
| Integer magnitude | exact when every value and the sum are below 2^53 in magnitude | beyond: still computed, documented loss of precision (README limit) |
| Input file | one number per line, optional `value` header | missing file, unreadable line (`line 7: "12abc" is not a number`) · exit 2 |
| Flags | `-data`, `-trim`, `-cpuprofile`, `-memprofile` | unknown flag, `-trim` without `-data`, malformed `-trim` · usage text · exit 2 |

### 8.2 Deliberate differences from R (README table)

| Case | R | This package | Why |
|---|---|---|---|
| negative trim | ignored, untrimmed mean | `ErrProportion` | a negative proportion is a caller mistake, not a request |
| trim ≥ 0.5 | returns the median | `ErrProportion` | a mean function that silently returns a median surprises callers |
| NA / NaN in data | `NA` | `ErrNaN` | an error says where; `NA` propagates silently |
| empty data | `NaN` | `ErrEmpty` | same |

**Same as R** (not a difference, listed because it surprises): when the sum of very large values overflows, R re-adds each value divided by the count and returns a finite mean; the package does the same.
| ±Inf kept, result NaN | `NaN` | `ErrNotFinite` | no silent NaN |

(Every R behaviour in this table is confirmed by `testdata/r-reference/semantics.tsv`, spike addendum B.)

### 8.3 Failure behaviour

- **Bad input:** rejected before any computation, with the line number or value; nothing written.
- **Partial failure (comparison mode):** every row is still compared and reported; the exit code says whether
  all agreed. `results/comparison.tsv` is written to a temporary file and renamed, so an interrupted or failed
  write never leaves a half file.
- **Interruption (Ctrl-C), full disk:** the program has nothing to clean up beyond the temporary file; a write
  error is reported with the path and exit 1.
- **Timeout:** not applicable (no network, bounded work).

### 8.4 Unintended consequences

| For | Risk | Guard |
|---|---|---|
| Data | the caller's slice reordered by sorting | FR-3, tested by comparing the input before and after |
| Data | personal data in committed samples | samples are synthetic, from a seeded script (constitution 8.3); data gate |
| Users | a misleading number (NaN, a median presented as a mean, silent precision loss) | FR-4, §8.2, the integer limit in the README |
| Users | believing asymmetric results are "R's" | README says asymmetric is our definition, cross-checked in R |
| External systems | a mistaken release cached forever by the Go module proxy | a tag is pushed only after the package's acceptance passes; never moved; a fix gets a new version (constitution 9.2) |

## 9. Acceptance criteria

| AC | Criterion | Test |
|---|---|---|
| AC-1 | Hand-computed cases: `Mean([]int{1,2,3,4,100}, 0.2) == 3`; asymmetric `Mean([]float64{1,2,3,4,5,6,7,8,9,10}, 0.1, 0.2) == 5` (keeps 2…8); no proportion equals the plain mean | package unit test |
| AC-2 | `floor` rule: n = 100, trim 0.29 drops 28 from each end | package unit test |
| AC-2a | Integer path is R's one-pass mean (a case where one-pass and two-pass differ); float path has R's overflow fallback (`[1e308, 1e308]` gives 1e308) | package unit test |
| AC-3 | Every R reference row (symmetric built-in and asymmetric defined) agrees under §7 | package test over `testdata/` copy · program comparison mode |
| AC-4 | Every error in FR-4 is returned for its case (including NaN and ±Inf proportions and three proportions), matched with `errors.Is`; a 30 s fuzz run finds no panic and no result outside [min, max] (evidence, not proof) | package unit + fuzz test |
| AC-5 | The input slice is unchanged after the call | package unit test |
| AC-6 | Counts equal `floor(n·p)` and keep at least one value for every n in 1…105 and proportions up to `Nextafter(0.5, 0)`, and for n up to 2^53 | package unit test |
| AC-7 | `go.mod` of the package has no `require`; `Example` functions pass | acceptance script |
| AC-8 | Neither repository has a `replace` directive or a `go.work` file; a fresh clone of the program builds with an empty `GOMODCACHE` and `-mod=readonly`, fetching the released tag; package and program tests pass with `GOPROXY=off` and no `Rscript` on `PATH`; the package's R test (AC-3) passes **on the Mac** | acceptance script (Mac) |
| AC-9 | Program with no arguments exits 0; stdout names every sample, its n, the Go and R values at trim 0.05 and PASS, and the count of agreeing rows; `results/comparison.tsv` has every reference row | built-binary test |
| AC-10 | Each row of §8.1 gives its message and exit code | built-binary tests |
| AC-11 | Mutation check: at least 6 behavioural mutants (`floor`→`round`, upper index off by one, low/high swapped, single-pass mean, copy removed, validation inverted) all caught | `docs/mutation-checks.txt` |
| AC-12 | Binaries for macOS and Windows and a macOS `.app` are built by the scripts | acceptance script |
| AC-13 | Each README has the required sections (package: install, use, errors, limits, R differences, tests, licence; program: roles, get and use, build, run, test, results, limits, GenAI Tools) and contains the `go get github.com/mioluigi/trimmedmean@` command; `docs/ai-dialogs/README.md` exists and lists every dialog file | acceptance script checks headings and strings |
| AC-14 | README numbers come from generated blocks; `scripts/readme.sh --check` passes | acceptance script |
| AC-15 | Every script and guard (data gate, measurement freshness, acceptance checks) is tested on a passing and a failing case | script tests |
| AC-16 | Every exported identifier of the package has a doc comment; no function exceeds 40 lines; `gofmt -l` and `go vet` are clean | package test (`go/ast`) + acceptance |

## 10. Traceability

| Step | Requirements | ACs |
|---|---|---|
| S1 package in its own repository, tutorialspoint | FR-1, FR-R1, FR-R2; spike §2 mapping | AC-7 |
| S2 ints and/or floats, proportions after sorting | FR-1, FR-3, FR-4, FR-5 | AC-1, AC-4, AC-5, AC-6 |
| S3 asymmetric | FR-2 | AC-1, AC-3 |
| S4 one argument → symmetric | FR-2 | AC-1 |
| S5 second repository, `main.go` imports | FR-7 … FR-12 | AC-8, AC-9, AC-10 |
| S6 ≥ 100 ints and floats vs R trim = 0.05 | FR-8, §7 | AC-3, AC-9 |
| S7 README: get and use, Go vs R summary | FR-R4, §8.2 | AC-13, AC-14 |
| S8 bootstrap (optional, deferred) | — | — |
| S9 two URLs, README each, GenAI Tools | — | AC-13 |

| Rubric | Where |
|---|---|
| [C] | NFR-1; Art. 2 |
| [T] | NFR-2, NFR-3, NFR-6; AC-3, AC-4, AC-11 |
| [D] | FR-1…6, FR-R1…R4, NFR-4 |
| [Doc] | FR-R4; AC-13, AC-14; Art. 5 |
| [A] | FR-9…12, NFR-5; AC-9, AC-10, AC-12 |

## 11. Open decisions (recommendation first)

| Q | Decision | Recommendation and reason |
|---|---|---|
| Q1 | API shape | **One variadic `Mean(values, proportions...)`**: it reads exactly like the assignment ("if only one argument is provided … symmetric"). Cost: three or more proportions are caught at run time (`ErrProportionCount`), not compile time (R1 4.4). Alternative: separate `Symmetric` and `Asymmetric` functions — clearer types, but two names for one idea. R1 agrees. |
| Q2 | "Integers and/or floats" | **Generics (`[T Number]`)**: one function for every numeric slice, checked at compile time. A Go slice holds one type, so a mixed list is passed as `[]float64`. Alternative: `[]any`, which loses type safety. |
| Q3 | Infinite values | **Allow ±Inf; reject a NaN or infinite result (`ErrNotFinite`)**: trimming is meant to remove extreme values, and the error catches the only harmful case. Alternative: reject any non-finite value. |
| Q4 | trim ≥ 0.5 | **Reject (`ErrProportion`)** with a message that names R's median behaviour. Alternative: return the median as R does. |
| Q5 | Asymmetric flag syntax | **`-trim "low,high"`**, one flag mirroring the function. Alternative: `-low` and `-high`. |
| Q6 | First release version | **`v0.1.0`** as soon as the package passes its tests, so the program can be built against it; **`v1.0.0`** after R2's fixes, and the program moves to it in the final acceptance step (recommendation changed after R1). No tag is ever moved. Alternative: `v1.0.0` straight away, with `v1.0.1` for any R2 fix. |
| Q7 | Float comparison rule | **As §7**, settled by spike evidence; reproducing R's partial sort is **not** recommended: it would tie the package's arithmetic to R's internals for a gap of about two machine epsilons. |
