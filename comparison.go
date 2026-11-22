package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"text/tabwriter"

	"github.com/mioluigi/go-trimmed-mean-demo/internal/compare"
	"github.com/mioluigi/go-trimmed-mean-demo/internal/sample"
)

// headlineTrim is the proportion the assignment compares with R: mean(x, trim = 0.05).
const headlineTrim = 0.05

// comparison is one reference row with Go's value and the verdict.
type comparison struct {
	reference compare.Reference
	got       float64
	result    compare.Result
}

// runComparison checks every row of R's reference against the package, writes the full table, and prints
// the trim = 0.05 rows and a one-line verdict. It returns an error if any row disagrees.
func runComparison(opts options, stdout io.Writer, logger *slog.Logger) error {
	references, err := compare.ReadReferences(opts.reference)
	if err != nil {
		return err
	}
	comparisons, err := compareAll(opts.dataDir, references, logger)
	if err != nil {
		return err
	}
	if err := writeTable(opts.out, comparisons); err != nil {
		return err
	}
	printHeadline(stdout, comparisons)

	failed := countFailed(comparisons)
	fmt.Fprintf(stdout, "\n%d of %d comparisons with R agree (rule: spec §7). Full table: %s\n",
		len(comparisons)-failed, len(comparisons), opts.out)
	if failed > 0 {
		return fmt.Errorf("%d comparisons with R failed; see %s", failed, opts.out)
	}
	return nil
}

// compareAll computes Go's trimmed mean for every reference row, reading each sample once.
func compareAll(dataDir string, references []compare.Reference, logger *slog.Logger) ([]comparison, error) {
	samples := map[string]sample.Sample{}
	comparisons := make([]comparison, 0, len(references))
	for _, reference := range references {
		values, ok := samples[reference.Sample]
		if !ok {
			var err error
			values, err = readMatching(dataDir, reference)
			if err != nil {
				return nil, err
			}
			samples[reference.Sample] = values
			logger.Info("read sample", "sample", reference.Sample, "n", values.Len(), "kind", kindName(values))
		}
		got, err := meanOf(values, []float64{reference.Low, reference.High})
		if err != nil {
			return nil, fmt.Errorf("%s %v/%v: %w", reference.Sample, reference.Low, reference.High, err)
		}
		result := compare.Check(reference, got, keptValues(values, reference))
		comparisons = append(comparisons, comparison{reference, got, result})
	}
	return comparisons, nil
}

// readMatching reads the sample a reference row names and checks it is the size and kind R saw.
func readMatching(dataDir string, reference compare.Reference) (sample.Sample, error) {
	values, err := sample.Read(filepath.Join(dataDir, reference.Sample+".csv"))
	if err != nil {
		return sample.Sample{}, err
	}
	sameKind := values.IsInteger() == (reference.Kind == "integer")
	if values.Len() != reference.N || !sameKind {
		return sample.Sample{}, fmt.Errorf("sample %s: %d %s, but R's reference has %d %s values",
			reference.Sample, values.Len(), kindName(values), reference.N, reference.Kind)
	}
	return values, nil
}

// keptValues returns the values R kept for a row (in sorted order), for the tolerance's scale.
func keptValues(values sample.Sample, reference compare.Reference) []float64 {
	sorted := make([]float64, 0, values.Len())
	for _, value := range values.Integers {
		sorted = append(sorted, float64(value))
	}
	sorted = append(sorted, values.Floats...)
	slices.Sort(sorted)
	return sorted[reference.DropLow : len(sorted)-reference.DropHigh]
}

// printHeadline prints the rows for R's mean(x, trim = 0.05), the comparison the assignment asks for.
func printHeadline(stdout io.Writer, comparisons []comparison) {
	fmt.Fprintln(stdout, "Go trimmedmean.Mean(x, 0.05) against R mean(x, trim = 0.05), committed samples:")
	table := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "sample\tn\tkind\tGo\tR\tdifference (ULPs)\tverdict")
	for _, c := range comparisons {
		symmetric := c.reference.Method == compare.BuiltIn && c.reference.Low == headlineTrim
		if symmetric {
			fmt.Fprintf(table, "%s\t%d\t%s\t%.17g\t%.17g\t%d\t%s\n", c.reference.Sample, c.reference.N,
				c.reference.Kind, c.got, c.reference.Value, c.result.ULPs, verdict(c.result))
		}
	}
	table.Flush()
}

// writeTable writes every comparison as a tab-separated table. It writes a temporary file in the same
// folder and renames it, so a failed or interrupted run never leaves a half-written table.
func writeTable(path string, comparisons []comparison) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".comparison-*.tsv")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name()) // no effect after a successful rename

	fmt.Fprintln(temporary, "sample\tkind\tn\tlow\thigh\tmethod\tgo\tr\tr_full_sort\tulps\tlimit\trule\tverdict")
	for _, c := range comparisons {
		r := c.reference
		fmt.Fprintf(temporary, "%s\t%s\t%d\t%v\t%v\t%s\t%.17g\t%.17g\t%.17g\t%d\t%.3g\t%s\t%s\n",
			r.Sample, r.Kind, r.N, r.Low, r.High, r.Method, c.got, r.Value, r.FullSortValue,
			c.result.ULPs, c.result.Limit, c.result.Rule, verdict(c.result))
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return os.Rename(temporary.Name(), path)
}

func countFailed(comparisons []comparison) int {
	failed := 0
	for _, c := range comparisons {
		if !c.result.Pass {
			failed++
		}
	}
	return failed
}

func verdict(result compare.Result) string {
	if result.Pass {
		return "PASS"
	}
	return "FAIL"
}
