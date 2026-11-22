package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// comparisonRow is one line of results/comparison.tsv, written by the program.
type comparisonRow struct {
	sample, kind, method, rule, verdict string
	n                                   int
	low, high, goValue, rValue, limit   float64
	ulps                                uint64
}

func readComparisons(path string) ([]comparisonRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("%s has no rows", path)
	}
	rows := make([]comparisonRow, 0, len(records)-1)
	for i, record := range records[1:] {
		row, err := parseComparison(record)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, i+2, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// parseComparison reads the columns sample kind n low high method go r r_full_sort ulps limit rule verdict.
func parseComparison(record []string) (comparisonRow, error) {
	if len(record) != 13 {
		return comparisonRow{}, fmt.Errorf("%d fields, want 13", len(record))
	}
	number := func(index int) float64 {
		value, _ := strconv.ParseFloat(record[index], 64)
		return value
	}
	ulps, err := strconv.ParseUint(record[9], 10, 64)
	if err != nil {
		return comparisonRow{}, err
	}
	return comparisonRow{
		sample: record[0], kind: record[1], n: int(number(2)), low: number(3), high: number(4),
		method: record[5], goValue: number(6), rValue: number(7), ulps: ulps, limit: number(10),
		rule: record[11], verdict: record[12],
	}, nil
}

// benchmark is the median of repeated `go test -bench` lines for one name.
type benchmark struct {
	name                string
	nsPerOp, bytesPerOp float64
	allocsPerOp, runs   int
}

// readBenchmarks reads `go test -bench -benchmem -count N` output and keeps, per benchmark, the
// median of the N runs (SFEDD §5.3: medians of repeated runs).
func readBenchmarks(path string) ([]benchmark, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	samples := map[string][][3]float64{}
	var order []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name, measurement, ok := parseBenchmarkLine(scanner.Text())
		if !ok {
			continue
		}
		if _, seen := samples[name]; !seen {
			order = append(order, name)
		}
		samples[name] = append(samples[name], measurement)
	}
	if len(order) == 0 {
		return nil, fmt.Errorf("%s has no benchmark lines", path)
	}
	benchmarks := make([]benchmark, 0, len(order))
	for _, name := range order {
		runs := samples[name]
		benchmarks = append(benchmarks, benchmark{
			name: name, runs: len(runs), nsPerOp: median(runs, 0), bytesPerOp: median(runs, 1),
			allocsPerOp: int(median(runs, 2)),
		})
	}
	return benchmarks, scanner.Err()
}

// parseBenchmarkLine reads "BenchmarkMean/floats/n=100-10  123  1710 ns/op  896 B/op  1 allocs/op".
func parseBenchmarkLine(line string) (string, [3]float64, bool) {
	fields := strings.Fields(line)
	if len(fields) < 8 || !strings.HasPrefix(fields[0], "Benchmark") {
		return "", [3]float64{}, false
	}
	var measurement [3]float64
	for i, unit := range []string{"ns/op", "B/op", "allocs/op"} {
		position := indexOf(fields, unit)
		if position < 1 {
			return "", [3]float64{}, false
		}
		value, err := strconv.ParseFloat(fields[position-1], 64)
		if err != nil {
			return "", [3]float64{}, false
		}
		measurement[i] = value
	}
	name := fields[0]
	if dash := strings.LastIndex(name, "-"); dash > 0 {
		name = name[:dash] // drop the GOMAXPROCS suffix
	}
	return strings.TrimPrefix(name, "BenchmarkMean/"), measurement, true
}

func indexOf(fields []string, want string) int {
	for i, field := range fields {
		if field == want {
			return i
		}
	}
	return -1
}
