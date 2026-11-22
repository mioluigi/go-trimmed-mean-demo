# CLAUDE.md: project context

Create a Go Package (Week 9): a reusable trimmed-mean module (`github.com/mioluigi/trimmedmean`, symmetric
and asymmetric) and this program repository, which imports it and checks it against R's
`mean(x, trim = 0.05)` on the same samples. Process: SFEDD v2.1 with the Week 8 draft lessons. Due
2026-11-22, 11:59 PM ET.

## Read these first (in order)

1. `docs/assignment-and-rubric.md`: steps S1–S9, rubric tags, who runs what
2. `docs/constitution.md`: non-negotiable rules (governs both repositories)
3. `docs/spikes/001-environment-and-r-reference.md`: what each machine can do; predictions; pre-mortem
4. `specs/001-trimmed-mean/` (from phase 2): spec, plan, tasks; work the next unchecked task

## Working rules

- One task per commit; message starts with the task ID (`T05: …`); phase 0–4 documents use `P0: …` etc.
- Test first; mutation-check important tests with `go test -count=1` (compile failures do not count).
- The workspace has **no R** and cannot reach `proxy.golang.org`: R runs on Luis's Mac, `go.sum` comes from
  the Mac, and workspace-only settings (`GOPROXY=direct`, `go.work`) are never committed.
- R's results enter Go tests as saved files with provenance (`testdata/r-reference/`).
- Never hand-type results; generate them from `results/`.
- No personal data in files (no e-mail, user name, host name, home path).
- Changes after approval go in `docs/decisions.md` (newest first).
