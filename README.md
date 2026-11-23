# go-trimmed-mean-demo

A Go program that uses the trimmed-mean package
**[github.com/mioluigi/trimmedmean](https://github.com/mioluigi/trimmedmean)** and checks it against R's
`mean(x, trim = 0.05)` on samples of at least 100 integers and 100 floats.

This is the second of two repositories for the MSDS 431 Week 9 assignment *Create a Go Package*:

| Repository | Role |
|---|---|
| [`mioluigi/trimmedmean`](https://github.com/mioluigi/trimmedmean) | **The package**: `trimmedmean.Mean(values, proportions...)`, symmetric or asymmetric trimming, for any Go integer or float slice. No dependencies. MIT licence. |
| `mioluigi/go-trimmed-mean-demo` (this one) | **The example program**: imports the package, compares it with R, and holds the data, the R script, the results and the project documents. |

## Summary

<!-- BEGIN generated:headline -->
- **84 of 84** comparisons with R agree under the rule of spec §7.
- At `trim = 0.05`, Go equals R **exactly** on **3 of 3** integer samples; on the float samples the largest difference from R is **2.2e-16**.
- Slowest benchmark (`floats/n=1000000`): **87.98 ms** per call, **7.63 MiB** allocated in **1** allocation(s).
<!-- END generated:headline -->

The integer results equal R's exactly. The float results equal R's own mean of the fully sorted kept values
exactly, and differ from R's `mean(x, trim)` only in the last digits, because R adds the kept values in the
order its partial sort leaves them ([why](#how-the-comparison-with-r-works)).

## Contents

| Path | Role |
|---|---|
| `main.go`, `comparison.go`, `profile.go` | the program: comparison mode, one-file mode, profiles |
| `internal/sample/` | reads a file of numbers (integers or floats) |
| `internal/compare/` | the rule that decides whether a Go result agrees with R's ([spec §7](specs/001-trimmed-mean/spec.md)) |
| `cmd/tables/` | turns `results/` into the generated tables of this README |
| `cmd/bootstrap-study/`, `internal/study/`, `internal/rrand/` | the optional bootstrap study: the course's `run-bootstrap-robust.R` reproduced in Go draw for draw, with R's random numbers ([below](#bootstrap-study-optional-step-s8)) |
| `data/` | the six samples: 100, 101 and 1,000 integers; 100, 101 and 1,000 floats (`MANIFEST.tsv`: size, SHA-256, seed) |
| `r/capture-reference.R` | the R script that generated the samples and R's results (seed 9431) |
| `testdata/r-reference/` | R's saved results: 84 trimmed means, R's behaviour at the edges, the doubles R read, the R session |
| `testdata/bootstrap-reference/`, `r/capture-bootstrap-reference.R` | R's results for the bootstrap study, from the course file run unchanged on the Mac (the course file itself is not committed) |
| `results/` | the official run on the Mac: the comparison, benchmarks, profiles, machine, generated tables |
| `scripts/` | build, `.app`, data gate, measurement, README tables, acceptance, review bundle, R capture |
| `docs/`, `specs/` | assignment and rubric, constitution, spike, spec, plan, tasks, reviews, decisions, AI dialogs |

## Getting and using the package

These steps use the package in **your own** Go program; nothing from this repository is needed.

1. Install Go 1.24 or later (<https://go.dev/dl/>), and check it with `go version`.
2. Make a folder for your program and start a module in it:

   ```bash
   mkdir mystats && cd mystats
   go mod init example.com/mystats
   ```

3. Get the package (this writes it into `go.mod` and its checksum into `go.sum`):

   ```bash
   go get github.com/mioluigi/trimmedmean@v1.0.0
   ```

4. Write `main.go`:

   ```go
   package main

   import (
   	"errors"
   	"fmt"
   	"log"

   	"github.com/mioluigi/trimmedmean"
   )

   func main() {
   	values := []float64{2.1, 2.4, 2.2, 2.3, 9.9, 2.5, 2.2, 0.1, 2.4, 2.3}

   	symmetric, err := trimmedmean.Mean(values, 0.1) // 10% from each end
   	if err != nil {
   		log.Fatal(err)
   	}
   	asymmetric, err := trimmedmean.Mean(values, 0.1, 0.2) // 10% from the bottom, 20% from the top
   	if err != nil {
   		log.Fatal(err)
   	}
   	fmt.Println(symmetric, asymmetric)

   	if _, err := trimmedmean.Mean(values, 0.5); errors.Is(err, trimmedmean.ErrProportion) {
   		fmt.Println("rejected:", err)
   	}
   }
   ```

5. Run it with `go run .`. It prints the two trimmed means, then the reason a proportion of 0.5 was
   rejected.

The package's own [README](https://github.com/mioluigi/trimmedmean#readme) documents every call, error and
limit. In short: no proportion gives the ordinary mean; **one** proportion trims **symmetrically**, like R's
`mean(x, trim)`; **two** trim the bottom and the top separately; each proportion must be in `[0, 0.5)`.
Integers (`[]int`, `[]int64`, …) and floats (`[]float64`, `[]float32`, …) are both accepted.

## Building

You need Go 1.24 or later and this repository:

```bash
git clone https://github.com/mioluigi/go-trimmed-mean-demo.git
cd go-trimmed-mean-demo
go build -o trimmed-mean-demo .          # this machine
scripts/build.sh                         # dist/: macOS (Apple silicon, Intel) and Windows .exe
scripts/build-macos-app.sh               # dist/TrimmedMean.app
```

`go build` downloads the package at the version in `go.mod` and checks it against `go.sum`.
`scripts/build.sh` cross-compiles from any machine, producing `trimmed-mean-demo-macos-arm64`,
`trimmed-mean-demo-macos-amd64` and `trimmed-mean-demo.exe`. **`TrimmedMean.app`** can be double-clicked
in Finder: it opens Terminal, runs the comparison with R on samples bundled inside the app, and shows how
to analyse your own file. (Unsigned: the first time, right-click it and choose *Open*.)

## Running

Run from the repository folder, where `data/` and `testdata/` are.

```bash
./trimmed-mean-demo                                       # compare with R on every committed sample
./trimmed-mean-demo -data data/integers-100.csv -trim 0.05
./trimmed-mean-demo -data data/floats-1000.csv -trim 0.05,0.10
./trimmed-mean-demo -h                                    # every option
```

On Windows: `trimmed-mean-demo.exe -data data\floats-100.csv -trim 0.05`.

**With no arguments** the program computes all 84 cases R saved (8 symmetric proportions and 6 asymmetric
pairs for each of the 6 samples), judges each by the rule below, prints the `trim = 0.05` comparison and a
one-line verdict, and writes the full table to `results/comparison.tsv`. **With `-data FILE`** it computes
one trimmed mean; `-trim P` trims `P` from each end (default 0.05), `-trim LOW,HIGH` trims the two ends
separately. The input file has one number per line, optionally under a first line `value`; blank lines are
ignored; a file of whole numbers is read as integers.

| Option | Meaning |
|---|---|
| `-data FILE` | compute one trimmed mean of FILE |
| `-trim P` or `-trim LOW,HIGH` | proportions trimmed (with `-data`) |
| `-data-dir`, `-reference`, `-out` | comparison mode: samples, R's results, output table |
| `-cpuprofile FILE`, `-memprofile FILE` | write profiles |

Logs (what was read and computed) go to standard error; results to standard output.

| Exit code | Meaning | Examples |
|---|---|---|
| 0 | success; in comparison mode, every row agrees with R | |
| 1 | a comparison with R failed, or the mean is not finite | `-data f.csv -trim 0` on a file holding both `-Inf` and `Inf` |
| 2 | a usage or input error, with a message saying what to fix | `-trim 0.6` → `trim proportion must be in [0, 0.5): got 0.6`; a line `12abc` → `line 3: "12abc" is not a number`; an empty file → `no numbers in FILE`; `-trim` without `-data` |

## Testing

```bash
go test ./...                    # all tests; they need neither R nor the internet
scripts/check-inputs.sh          # the data gate: every sample matches data/MANIFEST.tsv
scripts/acceptance.sh            # every acceptance criterion, on fresh clones of both repositories (needs GitHub)
```

- `main_test.go` builds the program and tests the **built binary**: the comparison (84 of 84), a planted wrong
  R value (exit 1, `FAIL`), one-file mode, twelve input errors with their exact message and exit code, and
  the profile flags.
- `internal/compare` tests the rule with planted one-ULP differences; `internal/sample` the reader;
  `cmd/tables` the generated tables, including a stale README.
- The package's own tests ([trimmedmean](https://github.com/mioluigi/trimmedmean#tests)) include the same 84
  R results, examples, a fuzz test and benchmarks.
- [`docs/mutation-checks.txt`](docs/mutation-checks.txt): deliberate defects in the package, the program and
  the bootstrap study; the tests caught every one.
- `internal/rrand`, `internal/study`: R's random draws, the course study's 400 rows and 16 standard errors, and
  its printed output, all against R's saved values.

To regenerate R's results (needs R; changes the committed samples only if the R script changes):
`scripts/capture-r.sh`.

## Results

Official run on the Mac (`scripts/measure.sh`). Every number below is generated from `results/` by
`scripts/readme.sh`; none is typed.

**Go against R at `trim = 0.05`**, the comparison the assignment asks for:

<!-- BEGIN generated:comparison -->
| Sample | n | Kind | Go `Mean(x, 0.05)` | R `mean(x, trim = 0.05)` | Difference | ULPs | Rule | Verdict |
|---|---:|---|---:|---:|---:|---:|---|---|
| integers-100 | 100 | integer | 50.81111111111111 | 50.81111111111111 | 0 | 0 | exact | PASS |
| integers-101 | 101 | integer | 51.032967032967036 | 51.032967032967036 | 0 | 0 | exact | PASS |
| integers-1000 | 1000 | integer | 51.317777777777778 | 51.317777777777778 | 0 | 0 | exact | PASS |
| floats-100 | 100 | double | 0.082987777777777605 | 0.082987777777777827 | 2.2e-16 | 16 | exact full sort + tolerance | PASS |
| floats-101 | 101 | double | -0.042767032967032943 | -0.042767032967032916 | 2.8e-17 | 4 | exact full sort + tolerance | PASS |
| floats-1000 | 1000 | double | -0.031235555555555555 | -0.031235555555555468 | 8.7e-17 | 25 | exact full sort + tolerance | PASS |
<!-- END generated:comparison -->

**All 84 comparisons**, by the rule that applies:

<!-- BEGIN generated:agreement -->
| Rule · kind | Rows | Agree | Largest distance from R's built-in (ULPs) |
|---|---:|---:|---:|
| exact full sort + tolerance · double | 21 | 21 | 303 |
| exact · double | 21 | 21 | 0 |
| exact · integer | 42 | 42 | 0 |
<!-- END generated:agreement -->

**Time and memory** of `trimmedmean.Mean` (package benchmarks, `go test -bench -benchmem`, 5 runs each,
medians):

<!-- BEGIN generated:benchmarks -->
| Benchmark (`trim = 0.05`) | Runs | Median time per call | Memory per call | Allocations |
|---|---:|---:|---:|---:|
| floats/n=100 | 5 | 1.10 µs | 896 B | 1 |
| ints/n=100 | 5 | 1.00 µs | 896 B | 1 |
| floats/n=10000 | 5 | 573.54 µs | 80.0 KiB | 1 |
| ints/n=10000 | 5 | 471.19 µs | 80.0 KiB | 1 |
| floats/n=1000000 | 5 | 87.98 ms | 7.63 MiB | 1 |
| ints/n=1000000 | 5 | 44.32 ms | 7.63 MiB | 1 |
<!-- END generated:benchmarks -->

The package makes one allocation per call: a float64 copy of the input, which it sorts, so the caller's
slice is never changed. The integer benchmarks are faster than the float ones of the same size because
their values are drawn from only 1,000 distinct integers, which sort faster, and integers need one pass
instead of two.

Profiles of the program on a file of 1,000,000 values are in [`results/cpu-profile.txt`](results/cpu-profile.txt)
and [`results/mem-profile.txt`](results/mem-profile.txt). They show that, for the **program**, reading the
file costs more than the trimmed mean: most of the memory allocated is the reader's line buffer
(`sample.readValueLines`), and the sort (`slices.partitionOrdered`) is the largest single CPU cost of the
package. A reader that parses as it scans would cut the program's memory; the package is not affected.

<!-- BEGIN generated:machine -->
```
measured: 2026-11-22 19:53 EST
commit: 7862922
machine: MacBookPro18,2 · Apple M1 Max · 64 GB · macOS 27.2.1
go: go version go1.27.2 darwin/arm64
trimmedmean: v1.0.0
R reference: R version 4.4.2 (2024-10-31), aarch64-apple-darwin20
```
<!-- END generated:machine -->

## Bootstrap study (optional step S8)

The assignment's optional step asks for a bootstrap study showing the value of the trimmed mean as a robust
estimator, based on the course file `run-bootstrap-robust.R` (Thomas W. Miller, May 23, 2024). Rather than
write a new study, **`cmd/bootstrap-study` reproduces the course's study in Go, draw for draw**, with the
trimmed mean computed by this package, and checks every result against R
([spec 002](specs/002-bootstrap-study/spec.md)).

```bash
go run ./cmd/bootstrap-study            # the course's settings: B = 100; prints what the course prints
go run ./cmd/bootstrap-study -B 1000    # more resamples (2 to 10,000)
```

The study draws 100 samples at each size n = 25, 100, 225, 400 from a population with **1% outliers**
(normal with mean 100 and SD 10, contaminated by values near 200), and bootstraps each sample B times.

<!-- BEGIN generated:bootstrap -->
<!-- END generated:bootstrap -->

**Reading it.** With even 1% outliers, the standard error of the mean rises well above what the Central
Limit Theorem gives for a clean sample, because an outlier in a sample pulls the mean far. A 5% trimmed mean
removes those few extreme values, so its standard error stays close to the clean-sample value; it is also
below the median's, because the median discards far more of the data. At the smallest size the median can
do better by the course's measure. That is the case for the trimmed mean: most of the mean's efficiency, most
of the median's robustness.

**Two things found by reading the course file** (spec 002 §6). Its "bootstrap" standard errors are the
standard deviation, over the 100 samples, of each statistic **averaged over its B resamples**, which is close
to the statistic's own sampling variability rather than the usual bootstrap estimate from one sample; so the
second table adds the usual estimate (the spread of each sample's B values), computed in Go only because the
course file does not keep those values. And its comment says the seed is set inside the loop, while the code
sets it once before the loops; the study is reproducible either way.

**How Go gets R's random numbers.** `internal/rrand` reuses the Week 8 generator, which reproduces R's
`set.seed`, `runif` and `sample` exactly, and adds R's `rnorm`: R's default method draws two uniforms and
inverts the normal distribution with Wichura's algorithm AS 241. It was checked bit for bit against Python's
pure-Python AS 241 on 3,998 probabilities and against R's own draws; the checks found one coefficient
transposed in the first draft, and that Go's `math.Log` can be one unit in the last place away from the C
library's correctly rounded `log` that R uses, so the tail uses a correctly rounded logarithm.

## How the comparison with R works

R's results come from [`r/capture-reference.R`](r/capture-reference.R), run once on the Mac: it generated
the samples (seed 9431; floats written with 4 decimals), read them back, and saved `mean(x, trim)` for each.
What R does was learned by running it ([spike 001](docs/spikes/001-environment-and-r-reference.md)):

- R removes `floor(n × trim)` values from each end, in float64: `100 × 0.29` removes 28, not 29. The
  package does the same.
- R averages **integers** with one exact sum divided by the count, and **doubles** with a two-pass method;
  the package does the same, so for the same numbers an integer and a float sample can differ in the last
  digit, in R and in Go alike.
- R sorts only **partly** before averaging, so it adds the kept values in a different order from a full sort.
  On a computer, a sum depends on the order, so R's own result changes in the last digits. Measured on these
  samples, the gap is at most about two machine epsilons, relative to the size of the values.

The rule ([spec §7](specs/001-trimmed-mean/spec.md)) follows from this. Results must be **identical** for
integer samples, untrimmed means and the asymmetric definition. For trimmed float samples, Go must be
identical to R's mean of the fully sorted kept values, **and** within 10⁻¹⁴ × the mean absolute kept value of
R's `mean(x, trim)` (about 20 times the largest gap measured). Asymmetric trimming has no R function, so R
evaluated the package's definition for the 36 asymmetric cases.

## Limits

| Limit | Behaviour |
|---|---|
| Proportions | each in `[0, 0.5)`, one or two of them; anything else is an error (exit 2). R instead returns the median for 0.5 or more and ignores a negative trim |
| Values | NaN is rejected (exit 2); `±Inf` are accepted if they are trimmed away; a NaN or infinite result is an error (exit 1) |
| Integers | exact up to 2^53 in magnitude (values and their sum); larger integers lose precision as float64 |
| Floats against R | identical to R's full-sort mean; within about two machine epsilons of R's `mean(x, trim)` (see above) |
| Size | memory for one float64 copy (8 bytes per value) plus the file being read; time `O(n log n)` |
| Input files | one number per line, optional `value` header; a line that is not a number stops the run with its line number |
| R results | saved from R 4.4.2 on an Apple-silicon Mac; regenerating them on another platform could change the last digits of the float rows |

## Development process

Built with *Spec-First, Evidence-Driven Development* (SFEDD): assignment and rubric
([`docs/assignment-and-rubric.md`](docs/assignment-and-rubric.md)), constitution
([`docs/constitution.md`](docs/constitution.md)), a spike that ran R before planning
([`docs/spikes/`](docs/spikes/)), a spec reviewed by Gemini ([`docs/reviews/`](docs/reviews/)) and approved,
plan and tasks ([`specs/001-trimmed-mean/`](specs/001-trimmed-mean/)), one task per commit, and every
change after approval in [`docs/decisions.md`](docs/decisions.md).

## GenAI Tools

| Tool | Plan | Used for |
|---|---|---|
| **Claude** (Anthropic), with Claude Code | Claude Pro | the **builder**: drafted every document, ran the spike, wrote the tests first and then the code of both repositories, the R script and the scripts; triaged both reviews; committed and pushed through GitHub |
| **Gemini** (Google) | Gemini Ultra | the **independent reviewer**, twice: R1 on the spec before approval, R2 on both finished repositories |

How they were used: Luis Varela, the owner, set the intent and the process, answered every open decision,
approved the spec, plan and tasks, ran everything that needed his Mac or the internet (R, the release tag,
the official measurements, the acceptance run), and submitted. Claude never ran R: R's results reached the
tests only as files committed from the Mac. Every reviewer finding received a recorded decision; claims
that could be tested were tested first, and two inaccurate reviewer statements are recorded in
[`docs/reviews/001-spec-review.md`](docs/reviews/001-spec-review.md). The conversation logs are in
[`docs/ai-dialogs/`](docs/ai-dialogs/README.md).
