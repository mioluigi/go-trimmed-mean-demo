# Tasks 001

Status: **v2** (2026-11-22, after R1); T01–T04 done on branch `pre-approval`. One task = one commit (`T07: …`). 🔴 = failing test first. Tag: **C** Claude
(workspace) · **M** Mac (Luis).

| ID | Tag | Task | Verify | Repo |
|---|---|---|---|---|
| T00 | M | Spike 001 Mac part: environment probe, samples and R reference | `check-inputs: ok`; files pushed | demo |
| T01 | C | 🔴 `go.mod`, `Number`, errors, `validateProportions` (FR-2, FR-4) | error tests pass with `errors.Is` | pkg |
| T02 | C | 🔴 `dropCounts`, sort copy, `average`, `Mean` (FR-1, 3, 5; AC-1, 2, 5, 6) | hand-computed tests pass | pkg |
| T03 | C | 🔴 R reference test over `testdata/` (AC-3) | every row agrees under spec §7 | pkg |
| T04 | C | Examples, fuzz (30 s), benchmarks (FR-R3, AC-4, NFR-2) | `go test -run Example`, `-fuzz`, `-bench` | pkg |
| T05 | C | Package README (FR-R4) | headings check | pkg |
| T06 | C | Mutation check, `-count=1`, ≥ 6 behavioural mutants (AC-11) | `docs/mutation-checks.txt` all caught | demo |
| T07 | M | Tag and push `v0.1.0` (FR-R2, Q6) | `go list -m github.com/mioluigi/trimmedmean@v0.1.0` on the Mac | pkg |
| T08 | M | Program `go.mod`/`go.sum` from `go get …@v0.1.0` (FR-7) | `go build ./...` on the Mac | demo |
| T09 | C | 🔴 `internal/sample` reader (FR-10; §8.1 file rows) | unit tests | demo |
| T10 | C | 🔴 `internal/compare`: rule, ULPs (FR-8, §7) | unit tests incl. a planted 1-ULP difference | demo |
| T11 | C | 🔴 `main.go`: modes, flags, logs, profiles, exit codes (FR-8…12; AC-9, 10) | built-binary tests | demo |
| T12 | C | Scripts: `build.sh`, `build-macos-app.sh`, `measure.sh`, `readme.sh`, `cmd/tables`, `acceptance.sh`, `review-bundle.sh` (AC-12, 14) | each script's pass and fail case | demo |
| T13 | M | `scripts/measure.sh` official run → `results/` | freshness guard passed; comparison all PASS | demo |
| T14 | C | Program README with generated blocks; GenAI Tools; re-check every sentence (AC-13, 14) | `readme.sh --check` | demo |
| T15 | M | Tag `v1.0.0`, move the program to it, acceptance on fresh clones of both repositories (AC-7, 8, 12, 13, 15, 16) | `acceptance.sh`: all PASS | both |
| T16 | M | Export AI dialogs to `docs/ai-dialogs/` + index | files present | demo |
| — | C/M | *(Deferred, optional)* S8 bootstrap study, only after T16 | — | demo |
