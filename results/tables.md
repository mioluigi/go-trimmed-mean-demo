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
| ints/n=100 | 5 | 1.01 µs | 896 B | 1 |
| floats/n=10000 | 5 | 576.64 µs | 80.0 KiB | 1 |
| ints/n=10000 | 5 | 475.85 µs | 80.0 KiB | 1 |
| floats/n=1000000 | 5 | 88.54 ms | 7.63 MiB | 1 |
| ints/n=1000000 | 5 | 44.34 ms | 7.63 MiB | 1 |

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
- Slowest benchmark (`floats/n=1000000`): **88.54 ms** per call, **7.63 MiB** allocated in **1** allocation(s).

## machine

```
measured: 2026-11-22 18:40 EST
commit: 3f4e47b
machine: MacBookPro18,2 · Apple M1 Max · 64 GB · macOS 27.2.1
go: go version go1.27.2 darwin/arm64
trimmedmean: v0.1.0
R reference: R version 4.4.2 (2024-10-31), aarch64-apple-darwin20
```
