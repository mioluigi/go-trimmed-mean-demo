# Plan 001: Trimmed mean package and demonstration program

| Field | Value |
|---|---|
| Status | **DRAFT v1** (2026-11-22), drafted during review R1 under the same-day deadline; nothing is built until approval |
| Spec | `specs/001-trimmed-mean/spec.md` |

Every design choice has its reason next to it, so Luis can defend it.

## 1. Package `github.com/mioluigi/trimmedmean`

```
trimmedmean/
  go.mod              module github.com/mioluigi/trimmedmean · go 1.24 · no require (FR-R1)
  trimmedmean.go      Number, Mean, the errors, and three small helpers
  trimmedmean_test.go hand-computed cases, errors, input unchanged, at-least-one-kept (AC-1, 2, 4, 5, 6)
  reference_test.go   every R reference row (AC-3), reading testdata/
  example_test.go     Example, Example_asymmetric, Example_errors (FR-R3)
  fuzz_test.go        FuzzMean: never panics; result between the kept min and max (AC-4)
  bench_test.go       BenchmarkMean at n = 100, 10 000, 1 000 000, ints and floats (NFR-2, NFR-4)
  testdata/           copy of the six samples and R's reference, with a README naming their source commit
  README.md · LICENSE
```

| Choice | Reason |
|---|---|
| **Generics with a `Number` constraint** (`~int … ~float64`) | One function for every numeric slice, type-checked at compile time; the `~` admits named types such as `type Celsius float64`. |
| **Variadic proportions** | Mirrors the assignment's "one argument → symmetric" literally, and keeps one function name (Q1). |
| **Sentinel errors** (`var ErrEmpty = errors.New(…)`), wrapped with the offending value (`fmt.Errorf("%w: %v", ErrProportion, p)`) | Callers test with `errors.Is`; people read the value. |
| **Convert to float64, then `slices.Sort` a copy** | The copy protects the caller (FR-3); converting first means one sort routine and one mean routine for all types. `slices.Sort` orders floats correctly once NaN is rejected. |
| **Validate before sorting** | Bad input costs no work, and the checks read as one list at the top of `Mean`. |
| **`dropCounts(n, low, high)` uses `math.Floor(float64(n) * p)`** | The same float64 product R computes, so the counts agree with R, including `100 × 0.29 → 28` (spike P1). |
| **R's two-pass mean** (`mean := sum/n; mean += Σ(x − mean)/n`), reused in spirit from Week 8 `stats.Mean` | It is R's own algorithm; on an arm64 Mac R's long double is 64 bits, so the same steps give the same bits (spike P4 decides how far that holds). |
| **Integers get R's one-pass mean; floats R's two-pass mean with its overflow fallback** | Spike addendum B: R treats them differently (P5 failed, 12 of 24 rows), and the one-pass rule reproduces all 42 integer rows. The kind is known from `T` (`T(1)/T(2) == 0` only for integers), so no flag is needed. |
| **No sort when nothing is trimmed** | R doesn't sort then either, so untrimmed results match R exactly; it also saves the `O(n log n)` work. |
| **Guard `dropLow + dropHigh < n` even though `p < 0.5` implies it** | For very large `n` the product can round up to `n/2` (spec FR-5). |

Functions (each one task): `Mean` (validate → convert → sort → slice → average), `validateProportions`,
`dropCounts`, `average`. No function longer than about 25 lines.

## 2. Program `github.com/mioluigi/go-trimmed-mean-demo`

```
main.go                   flags → run(args, stdout, stderr) int → os.Exit (only here)
main_test.go              builds the binary once, then tests exit codes and messages (AC-9, AC-10)
internal/sample/          Read(path) → ints or floats, with line-numbered errors (FR-10)
internal/compare/         the §7 rule, ULP distance, PASS/FAIL rows (FR-8)
cmd/tables/               results/comparison.tsv → markdown table and headline for the README
scripts/                  env-probe · capture-r · check-inputs · build · build-macos-app · measure · readme · acceptance · review-bundle
data/ · testdata/r-reference/ · results/
```

| Choice | Reason |
|---|---|
| `run` returns an exit code; `main` only calls `os.Exit(run(…))` | Testable without a subprocess, and the built-binary tests still check the real codes (L-W6-14). |
| Kind detected from the content (every line an integer → `[]int`) | The user doesn't need a flag, and integer samples exercise the generic integer path (S2). |
| Comparison writes to a temporary file and renames | No half-written results on failure (spec §8.3). |
| `go.mod` requires `trimmedmean v1.0.0`; `go.sum` from the Mac | L-W7-06; the workspace builds with `GOPROXY=direct GOSUMDB=off` in its own environment only, never committed. |
| Measurement is benchmarks + one profiled comparison run, not R timing | The assignment asks for agreement with R, not speed against R; the rubric's "memory and processing requirements" is met by `-benchmem` and profiles. |

## 3. Release and hand-off sequence

1. Package tasks T01–T06 in the workspace; push.
2. **Mac:** `git tag v1.0.0 && git push origin v1.0.0` in the package (T07), then in the program
   `go get github.com/mioluigi/trimmedmean@v1.0.0 && go mod tidy`, commit `go.mod`/`go.sum`, push (T08).
3. Program tasks T09–T12 in the workspace; push.
4. **Mac:** `scripts/measure.sh` (freshness guard; comparison, benchmarks, profiles) → `results/`; push (T13).
5. READMEs with generated tables (T14); R2; acceptance on fresh clones (Mac).
