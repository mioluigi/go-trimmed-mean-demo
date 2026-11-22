# Assignment Requirements and Rubric (summary)

> Summary of the Week 9 assignment, *Create a Go Package* (100 points, due **2026-11-22, 11:59 PM ET**,
> confirmed by Luis from Canvas at 17:27 ET; assignment page revised March 13, 2025), in our own words, so
> every tool and reviewer can check the spec against it. The Canvas page is the authoritative source.

## Scenario

A **Go consulting firm** wants to promote its services by publishing **open-source Go packages on GitHub**.
Its first offering is a package that computes a **symmetric or asymmetric trimmed mean**.

## Required steps (numbered for traceability)

| # | Step | Source wording, condensed |
|---|---|---|
| S1 | Create a **Go package** for the trimmed mean in **its own GitHub repository**, following the tutorialspoint instructions | "Follow these instructions for creating a Go package" (<https://www.tutorialspoint.com/article/how-to-create-your-own-package-in-golang>; mapping in spike 001 §2) |
| S2 | A function that takes a **slice of integers and/or floats** and computes a trimmed mean; the trimming arguments are **proportions** taken from the lower and upper ends **after sorting** from lowest to highest | "read a slice of integers and/or floats and compute a trimmed mean" |
| S3 | **Asymmetric trimming**: different proportions from the lower and upper ends | "different proportions of trimming are possible" |
| S4 | With **one** trimming argument, trimming is **symmetric** | "If only one degree-of-trimming argument is provided, assume … symmetric" |
| S5 | A **second GitHub repository** whose **`main.go` imports the package** and uses it to compute a trimmed mean | "Develop a main.go program that imports the trimmed mean package" |
| S6 | Using samples of **at least 100 integers and 100 floats**, test the Go function against R's **`mean(x, trim = 0.05)`** (symmetric, 0.05 from each end) | "test the Go trimmed mean function against symmetric trimming results from … R" |
| S7 | The second repository's README gives **complete instructions for getting and using the package** in a Go program, and **summarises Go results against R's** | "provide complete instructions for getting and using the package … Summarize results" |
| S8 | *(Optional)* A **bootstrap study** showing the value of the trimmed mean as a robust estimator of central tendency (course file `run-bootstrap-robust.R`) | "(Optional) Run a bootstrap study" |
| S9 | Submit **both repository URLs** in the Canvas comments; a **README.md in each** repository with instructions for getting, using, running and testing; **GenAI Tools** section with conversation logs | deliverables |

The page also suggests using the **same random number generator** across R and Go (Mersenne Twister,
`github.com/seehuhn/mt19937`) for Monte Carlo work. It applies to S8 only.

## Deliverables

- Two GitHub URLs in the Canvas comments: the **package** repository and the **example program** repository.
- A `README.md` in each: roles of programs and data, how to get, use, run and test.
- **GenAI Tools** section and conversation logs (plain text, PDFs or links).

## Rubric: five criteria, 20 points each

Ratings per criterion: Outstanding 20 (all criteria) · Meets 15 (most) · Needs Improvement 10 · Partially
Acceptable 5 · Unacceptable 0.

| Tag | Criterion | What earns full marks |
|---|---|---|
| **[C]** | Coding rules, organization, aesthetics | Go modules, idiomatic Go; meaningful, specific names; no useless temporaries; clean blocks; small functions; simple control flow; distribute the NOT; self-documenting code with comments explaining the logic (McConnell 2004, 777–817). |
| **[T]** | Testing and software metrics | Unit tests of critical components with synthetic data; logs and profiles where appropriate; memory and processing requirements monitored; Monte Carlo benchmark if required (here: optional, S8). |
| **[D]** | Design and development | Clean, efficient design meeting every requirement and the use case; general-purpose, reusable modules. |
| **[Doc]** | Documentation | Effective Git/GitHub; a README.md **for each repository** noting the roles of programs and data and how to test and use. |
| **[A]** | Application | Instructions for creating an executable (`.exe` for Windows, `.app` for macOS); runs to completion; **validates user input** with an explanation for incorrect input; clean UX/UI. |

## Who runs what

| Task | Where | Why |
|---|---|---|
| Documents, spikes in Go, tests, package and program code, scripts, review bundles, triage | **Claude** (workspace) | Go 1.24.7; GitHub read and push once attached |
| Creating the two repositories on github.com | **Mac** (Luis) | Only Luis can create them |
| Environment probe, R reference capture, data generation in R | **Mac** | The workspace has no R |
| Tagging the package release; `go get` and `go.sum` for the program | **Mac** | The workspace cannot reach `proxy.golang.org` / `sum.golang.org` (probed 2026-11-22, HTTP 403) |
| Measurement, `.app` build, final acceptance on fresh clones | **Mac** | Official machine |
| Gemini reviews R1 and R2; exporting AI dialogs; submission | **Luis** | Owner |
