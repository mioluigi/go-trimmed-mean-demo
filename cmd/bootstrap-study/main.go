// Command bootstrap-study runs the course's bootstrap study (run-bootstrap-robust.R) in Go, draw for draw,
// with the trimmed mean computed by github.com/mioluigi/trimmedmean (spec 002, assignment step S8).
//
// It prints exactly what the course file prints, then the added textbook bootstrap standard errors, and
// writes every standard error at full precision to -out. Exit codes: 0 success; 1 a run-time failure;
// 2 a usage error.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/mioluigi/go-trimmed-mean-demo/internal/study"
)

// maxReplicates bounds the run time (spec 002 §5): the work grows as iterations × Σn × B.
const maxReplicates = 10_000

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("bootstrap-study", flag.ContinueOnError)
	flags.SetOutput(stderr)
	replicates := flags.Int("B", 100, "bootstrap resamples per sample (the course uses 100)")
	out := flags.String("out", "results/bootstrap-study.tsv", "standard errors at full precision")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *replicates < 2 || *replicates > maxReplicates || flags.NArg() > 0 {
		fmt.Fprintf(stderr, "error: -B must be between 2 and %d, with no other arguments\n", maxReplicates)
		return 2
	}
	config := study.Course()
	config.Replicates = *replicates
	logger := slog.New(slog.NewTextHandler(stderr, nil))
	start := time.Now()
	rows, err := study.Run(config)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	summaries := study.Summarize(rows, config)
	logger.Info("study done", "B", config.Replicates, "samples", len(rows), "elapsed", time.Since(start).Round(time.Millisecond))
	fmt.Fprint(stdout, study.Listing(config, summaries))
	fmt.Fprint(stdout, study.TextbookListing(summaries))
	if err := writeTable(*out, config, summaries); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

// writeTable writes the standard errors at full precision, through a temporary file and a rename.
func writeTable(path string, config study.Config, summaries []study.Summary) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".bootstrap-*.tsv")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	fmt.Fprintln(temporary, "n\tB\tclt\tsampleMean\tbootMean\tbootMedian\tbootTrimmedMean\twithinMean\twithinMedian\twithinTrimmedMean")
	for _, s := range summaries {
		fmt.Fprintf(temporary, "%d\t%d\t%.17g\t%.17g\t%.17g\t%.17g\t%.17g\t%.17g\t%.17g\t%.17g\n", s.N, config.Replicates,
			s.CLT, s.SampleMean, s.BootMean, s.BootMedian, s.BootTrimmed, s.WithinMean, s.WithinMedian, s.WithinTrimmed)
	}
	if err := temporary.Close(); err != nil {
		return errors.Join(fmt.Errorf("writing %s", path), err)
	}
	return os.Rename(temporary.Name(), path)
}
