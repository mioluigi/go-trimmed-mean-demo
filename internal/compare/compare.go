// Package compare holds the rule by which a Go trimmed mean is judged to agree with R's (spec §7), and
// reads R's saved reference results.
package compare

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
)

// Tolerance bounds |Go − R built-in| for trimmed symmetric float rows, as a multiple of the mean absolute
// kept value. R adds the kept values in the order its partial sort leaves them; the measured worst gap is
// 4.7e-16 on that scale (spike 001 addendum B), so this limit is about 20 times it.
const Tolerance = 1e-14

// Methods named in the reference file.
const (
	BuiltIn = "mean(x, trim)" // R's own trimmed mean (symmetric)
	Defined = "defined"       // our asymmetric definition, evaluated in R with a full sort
)

// Rules a comparison can be judged by.
const (
	RuleExact     = "exact"
	RuleTolerance = "exact full sort + tolerance"
)

// Reference is one row of R's results (testdata/r-reference/trimmed-means.tsv).
type Reference struct {
	Sample, Kind, Method string
	N                    int
	Low, High            float64
	Value                float64 // R's result
	FullSortValue        float64 // R's mean of the fully sorted kept values
	DropLow, DropHigh    int
}

// Result is the verdict on one Go value.
type Result struct {
	Rule  string
	ULPs  uint64  // distance from R's value in units in the last place
	Limit float64 // absolute limit used by the tolerance rule; 0 for the exact rule
	Pass  bool
}

// Check judges got against reference by spec §7: exact for integer rows, untrimmed rows and our
// asymmetric definition; for trimmed symmetric float rows, exact against R's full-sort mean and within
// Tolerance × mean(|kept|) of R's built-in. kept is the values the trimmed mean averaged.
func Check(reference Reference, got float64, kept []float64) Result {
	ulps := ULPDistance(got, reference.Value)
	if needsExact(reference) {
		return Result{Rule: RuleExact, ULPs: ulps, Pass: ulps == 0}
	}
	limit := Tolerance * meanAbsolute(kept)
	pass := got == reference.FullSortValue && math.Abs(got-reference.Value) <= limit
	return Result{Rule: RuleTolerance, ULPs: ulps, Limit: limit, Pass: pass}
}

// needsExact reports whether a row must agree exactly: integers sum exactly in any order, untrimmed
// values are added in the same order by R and Go, and our definition is sorted fully on both sides.
func needsExact(reference Reference) bool {
	untrimmed := reference.Low == 0 && reference.High == 0
	return reference.Kind == "integer" || untrimmed || reference.Method == Defined
}

// ULPDistance returns how many representable float64 values lie between a and b (0 when equal; +0 and
// −0 count as equal).
func ULPDistance(a, b float64) uint64 {
	ordered := func(x float64) int64 {
		bits := int64(math.Float64bits(x))
		if bits < 0 {
			return math.MinInt64 - bits // map negative floats below the positive ones, in order
		}
		return bits
	}
	difference := ordered(a) - ordered(b)
	if difference < 0 {
		return uint64(-difference)
	}
	return uint64(difference)
}

func meanAbsolute(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += math.Abs(value)
	}
	return total / float64(len(values))
}

// ReadReferences reads R's saved results.
func ReadReferences(path string) ([]Reference, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("%s has no reference rows", path)
	}
	references := make([]Reference, 0, len(records)-1)
	for line, record := range records[1:] {
		reference, err := parseReference(record)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, line+2, err)
		}
		references = append(references, reference)
	}
	return references, nil
}

// parseReference turns one record (sample, kind, n, low, high, method, value, dropped_low, dropped_high,
// identical_to_full_sort, identical_as_double, full_sort_value) into a Reference.
func parseReference(record []string) (Reference, error) {
	if len(record) != 12 {
		return Reference{}, fmt.Errorf("%d fields, want 12", len(record))
	}
	numbers := make([]float64, 0, 7)
	for _, index := range []int{2, 3, 4, 6, 7, 8, 11} {
		value, err := strconv.ParseFloat(record[index], 64)
		if err != nil {
			return Reference{}, fmt.Errorf("field %d: %w", index+1, err)
		}
		numbers = append(numbers, value)
	}
	return Reference{
		Sample: record[0], Kind: record[1], Method: record[5],
		N: int(numbers[0]), Low: numbers[1], High: numbers[2], Value: numbers[3],
		DropLow: int(numbers[4]), DropHigh: int(numbers[5]), FullSortValue: numbers[6],
	}, nil
}
