// Command go-trimmed-mean-demo shows the trimmedmean package in use.
//
// With no arguments it checks the package against R's saved results on the committed samples, prints
// the comparison at trim = 0.05 (R's mean(x, trim = 0.05)), and writes every comparison to
// results/comparison.tsv. With -data and -trim it computes one trimmed mean for a file of numbers.
//
// Exit codes: 0 success; 1 a comparison with R failed, or a run-time failure; 2 a usage or input error.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/mioluigi/go-trimmed-mean-demo/internal/sample"
	"github.com/mioluigi/trimmedmean"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

// Errors for command-line mistakes; they lead to exit code 2.
var (
	errTrimWithoutData = errors.New("-trim needs -data FILE")
	errBadTrim         = errors.New("-trim must be one proportion (0.05) or two separated by a comma (0.05,0.10)")
)

// options are the command-line settings.
type options struct {
	data, trim              string // one-file mode
	dataDir, reference, out string // comparison mode
	cpuProfile, memProfile  string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run is the whole program; it returns the exit code so that tests can call it.
func run(args []string, stdout, stderr io.Writer) int {
	opts, err := parseFlags(args, stderr)
	if err != nil {
		return exitUsage
	}
	logger := slog.New(slog.NewTextHandler(stderr, nil))
	stopProfiles, err := startProfiles(opts)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return exitFailure
	}
	defer stopProfiles(logger)

	switch {
	case opts.data != "":
		err = runOne(opts, stdout, logger)
	case opts.trim != "":
		err = errTrimWithoutData
	default:
		err = runComparison(opts, stdout, logger)
	}
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
	}
	return exitCode(err)
}

func parseFlags(args []string, stderr io.Writer) (options, error) {
	var opts options
	flags := flag.NewFlagSet("go-trimmed-mean-demo", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&opts.data, "data", "", "file of numbers, one per line (optional header \"value\")")
	flags.StringVar(&opts.trim, "trim", "0.05", "proportion trimmed from each end, or \"low,high\"")
	flags.StringVar(&opts.dataDir, "data-dir", "data", "comparison mode: folder of samples")
	flags.StringVar(&opts.reference, "reference", "testdata/r-reference/trimmed-means.tsv", "comparison mode: R's results")
	flags.StringVar(&opts.out, "out", "results/comparison.tsv", "comparison mode: table written here")
	flags.StringVar(&opts.cpuProfile, "cpuprofile", "", "write a CPU profile to this file")
	flags.StringVar(&opts.memProfile, "memprofile", "", "write a memory profile to this file")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: go-trimmed-mean-demo                       compare with R on the committed samples")
		fmt.Fprintln(stderr, "       go-trimmed-mean-demo -data FILE [-trim P | -trim LOW,HIGH]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "error: unexpected argument %q\n", flags.Arg(0))
		flags.Usage()
		return options{}, errors.New("unexpected argument")
	}
	trimSet := false
	flags.Visit(func(f *flag.Flag) { trimSet = trimSet || f.Name == "trim" })
	if !trimSet {
		opts.trim = "" // the default applies only with -data
	}
	return opts, nil
}

// exitCode maps an error to the program's exit code: input and usage mistakes are 2, the rest 1.
func exitCode(err error) int {
	if err == nil {
		return exitOK
	}
	usageErrors := []error{
		errTrimWithoutData, errBadTrim, os.ErrNotExist, sample.ErrNoNumbers, sample.ErrNotNumber,
		trimmedmean.ErrEmpty, trimmedmean.ErrNaN, trimmedmean.ErrProportion, trimmedmean.ErrProportionCount,
	}
	for _, usageError := range usageErrors {
		if errors.Is(err, usageError) {
			return exitUsage
		}
	}
	return exitFailure
}

// runOne computes one trimmed mean for the file named by -data and reports it in plain words.
func runOne(opts options, stdout io.Writer, logger *slog.Logger) error {
	trim := opts.trim
	if trim == "" {
		trim = "0.05"
	}
	proportions, err := parseProportions(trim)
	if err != nil {
		return err
	}
	values, err := sample.Read(opts.data)
	if err != nil {
		return err
	}
	mean, err := meanOf(values, proportions)
	if err != nil {
		return fmt.Errorf("%s: %w", opts.data, err)
	}
	low, high := proportions[0], proportions[len(proportions)-1]
	logger.Info("computed", "file", opts.data, "n", values.Len(), "low", low, "high", high)
	fmt.Fprintf(stdout, "%s: %d %s\n", opts.data, values.Len(), kindName(values))
	fmt.Fprintf(stdout, "trimmed %v from the bottom and %v from the top (%d and %d values dropped)\n",
		low, high, dropped(values.Len(), low), dropped(values.Len(), high))
	fmt.Fprintf(stdout, "trimmed mean: %v\n", mean)
	return nil
}

// parseProportions reads "p" or "low,high".
func parseProportions(text string) ([]float64, error) {
	parts := strings.Split(text, ",")
	proportions := make([]float64, len(parts))
	for i, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("%w: got %q", errBadTrim, text)
		}
		proportions[i] = value
	}
	return proportions, nil
}

// meanOf calls the package with the sample's own element type.
func meanOf(values sample.Sample, proportions []float64) (float64, error) {
	if values.IsInteger() {
		return trimmedmean.Mean(values.Integers, proportions...)
	}
	return trimmedmean.Mean(values.Floats, proportions...)
}

// dropped is the number of values removed from one end, for the report only: floor(n × p), the rule
// the package applies (spec §3).
func dropped(n int, proportion float64) int {
	return int(math.Floor(float64(n) * proportion))
}

func kindName(values sample.Sample) string {
	if values.IsInteger() {
		return "integers"
	}
	return "floats"
}
