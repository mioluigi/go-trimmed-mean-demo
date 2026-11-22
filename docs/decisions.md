# Decisions log

Newest first. Each entry: date, decision, reason, documents updated. Entries marked **(on Luis's behalf)**
were taken under the deadline with Claude's recommendation, for Luis to overturn.

## 2026-11-22

- **Float samples use 4 decimals, not 17 digits** (spike 001 addendum A): R's 17-digit read-back failed on the
  Mac. R computes from what it reads; its doubles are saved at 17 digits for a Go parity test. *Docs:*
  constitution 8.3; `r/capture-reference.R`; `scripts/check-inputs.sh`.
- **The workspace cannot reach the Go module proxy or checksum database (HTTP 403).** The package will have
  no dependencies; the program's `go.mod` and `go.sum` are produced on the Mac from the released tag. No
  `go.work` or `replace` line is committed. *Docs:* constitution 4.4; spike 001 §1.
- **tutorialspoint is followed with modules, not GOPATH.** Its `go mod init mathutils` path cannot be
  fetched by `go get`; the module path is the repository address. *Docs:* constitution 9.4; spike 001 §2.
- **Process documents (constitution, spikes, spec, reviews, decisions, AI dialogs) live in the program
  repository**, so the package repository stays a clean library a firm would publish. **(on Luis's behalf)**
- **Week 8's draft lessons L-W8-01…16 and proposals P1–P9 are applied this week** although not yet confirmed
  into `SFEDD/lessons-learned.md`. **(on Luis's behalf)**
- **The optional bootstrap study (S8) is deferred until every required step and rubric criterion is done**
  (Luis, 17:45 ET). It starts only if the minimum submittable state is reached with time left, and it must not
  change anything already measured or reviewed.
- **Repository names: `mioluigi/trimmedmean` (package) and `mioluigi/go-trimmed-mean-demo` (program)**, created
  by Luis 17:45 ET. The package's import path ends in its package name, the Go convention.
- **The package is released under the MIT licence**, the usual choice for a firm promoting itself with
  open-source code. **(on Luis's behalf)**
- **Due date confirmed by Luis: 2026-11-22, 11:59 PM ET.** *Docs:* `docs/assignment-and-rubric.md`.
