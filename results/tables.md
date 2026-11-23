# Generated tables (cmd/tables; do not edit)

## agreement

| Rule · kind | Rows | Agree | Largest distance from R's built-in (ULPs) |
|---|---:|---:|---:|
| exact full sort + tolerance · double | 21 | 21 | 303 |
| exact · double | 21 | 21 | 0 |
| exact · integer | 42 | 42 | 0 |

## benchmarks

| Benchmark (`trim = 0.05`) | Runs | Median time per call | Memory per call | Allocations |
|---|---:|---:|---:|---:|
| floats/n=100 | 5 | 1.10 µs | 896 B | 1 |
| ints/n=100 | 5 | 1.00 µs | 896 B | 1 |
| floats/n=10000 | 5 | 573.75 µs | 80.0 KiB | 1 |
| ints/n=10000 | 5 | 469.53 µs | 80.0 KiB | 1 |
| floats/n=1000000 | 5 | 87.82 ms | 7.63 MiB | 1 |
| ints/n=1000000 | 5 | 44.00 ms | 7.63 MiB | 1 |

## bootstrap

B = 100 resamples, 100 samples per n, seed 9999 (the course's settings).

| n | CLT, no outliers | SE of the mean, from the samples | SE of the mean | SE of the median | SE of the trimmed mean (5%) | Largest relative difference from R |
|---:|---:|---:|---:|---:|---:|---:|
| 25 | 2.00 | 2.64 | 2.65 | 2.17 | **2.25** | 0 |
| 100 | 1.00 | 1.41 | 1.42 | 1.11 | **1.02** | 0 |
| 225 | 0.67 | 0.91 | 0.91 | 0.72 | **0.63** | 2e-16 |
| 400 | 0.50 | 0.75 | 0.75 | 0.68 | **0.57** | 2e-15 |

Added (Go only): the textbook bootstrap standard error, the SD of each sample's own B values, averaged over the 100 samples.

| n | mean | median | trimmed mean (5%) |
|---:|---:|---:|---:|
| 25 | 2.39 | 2.56 | **2.34** |
| 100 | 1.40 | 1.27 | **1.07** |
| 225 | 0.90 | 0.85 | **0.68** |
| 400 | 0.70 | 0.63 | **0.51** |

- Go reproduces R's 16 standard errors (4 per n) with a largest relative difference of **2e-15**.
- The trimmed mean has the smallest standard error of the three, by both measures, at n = **100, 225, 400**.

## comparison

| Sample | n | Kind | Go `Mean(x, 0.05)` | R `mean(x, trim = 0.05)` | Difference | ULPs | Rule | Verdict |
|---|---:|---|---:|---:|---:|---:|---|---|
| integers-100 | 100 | integer | 50.81111111111111 | 50.81111111111111 | 0 | 0 | exact | PASS |
| integers-101 | 101 | integer | 51.032967032967036 | 51.032967032967036 | 0 | 0 | exact | PASS |
| integers-1000 | 1000 | integer | 51.317777777777778 | 51.317777777777778 | 0 | 0 | exact | PASS |
| floats-100 | 100 | double | 0.082987777777777605 | 0.082987777777777827 | 2.2e-16 | 16 | exact full sort + tolerance | PASS |
| floats-101 | 101 | double | -0.042767032967032943 | -0.042767032967032916 | 2.8e-17 | 4 | exact full sort + tolerance | PASS |
| floats-1000 | 1000 | double | -0.031235555555555555 | -0.031235555555555468 | 8.7e-17 | 25 | exact full sort + tolerance | PASS |

## headline

- **84 of 84** comparisons with R agree under the rule of spec §7.
- At `trim = 0.05`, Go equals R **exactly** on **3 of 3** integer samples; on the float samples the largest difference from R is **2.2e-16**.
- Slowest benchmark (`floats/n=1000000`): **87.82 ms** per call, **7.63 MiB** allocated in **1** allocation(s).

## machine

```
measured: 2026-11-22 20:59 EST
commit: 600cbde
machine: MacBookPro18,2 · Apple M1 Max · 64 GB · macOS 27.2.1
go: go version go1.27.2 darwin/arm64
trimmedmean: v1.0.0
R reference: R version 4.4.2 (2024-10-31), aarch64-apple-darwin20
```
