package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// bootstrapRow is one sample size of results/bootstrap-study.tsv (Go) or of R's standard-errors.tsv.
type bootstrapRow map[string]float64

// readBootstrap reads a tab-separated table of numbers with a header.
func readBootstrap(path string) ([]bootstrapRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return nil, fmt.Errorf("%s: no rows (%v)", path, err)
	}
	rows := make([]bootstrapRow, 0, len(records)-1)
	for _, record := range records[1:] {
		row := bootstrapRow{}
		for i, name := range records[0] {
			value, err := strconv.ParseFloat(record[i], 64)
			if err != nil {
				return nil, fmt.Errorf("%s column %s: %w", path, name, err)
			}
			row[name] = value
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// courseColumns are the standard errors the course prints and R's reference holds.
var courseColumns = []string{"sampleMean", "bootMean", "bootMedian", "bootTrimmedMean"}

// renderBootstrap shows Go's study beside R's: the course's standard errors, the largest relative
// difference from R for each sample size, the added textbook column, and a generated conclusion.
func renderBootstrap(goRows, rRows []bootstrapRow) (string, error) {
	if len(goRows) != len(rRows) {
		return "", fmt.Errorf("Go has %d sample sizes, R %d", len(goRows), len(rRows))
	}
	var text strings.Builder
	fmt.Fprintf(&text, "B = %d resamples, 100 samples per n, seed 9999 (the course's settings).\n\n", int(goRows[0]["B"]))
	text.WriteString("| n | CLT, no outliers | SE of the mean, from the samples | SE of the mean | SE of the median | SE of the trimmed mean (5%) | Largest relative difference from R |\n")
	text.WriteString("|---:|---:|---:|---:|---:|---:|---:|\n")
	worstOverall, trimmedSmallest := 0.0, []string{}
	for i, g := range goRows {
		if g["n"] != rRows[i]["n"] {
			return "", fmt.Errorf("row %d: Go n = %v, R n = %v", i+1, g["n"], rRows[i]["n"])
		}
		worst := 0.0
		for _, column := range courseColumns {
			worst = math.Max(worst, relativeDifference(g[column], rRows[i][column]))
		}
		worstOverall = math.Max(worstOverall, worst)
		fmt.Fprintf(&text, "| %d | %.2f | %.2f | %.2f | %.2f | **%.2f** | %.1g |\n", int(g["n"]), g["clt"],
			g["sampleMean"], g["bootMean"], g["bootMedian"], g["bootTrimmedMean"], worst)
		if smallest(g["bootTrimmedMean"], g["bootMean"], g["bootMedian"]) && smallest(g["withinTrimmedMean"], g["withinMean"], g["withinMedian"]) {
			trimmedSmallest = append(trimmedSmallest, strconv.Itoa(int(g["n"])))
		}
	}
	text.WriteString("\nAdded (Go only): the textbook bootstrap standard error, the SD of each sample's own B values, averaged over the 100 samples.\n\n")
	text.WriteString("| n | mean | median | trimmed mean (5%) |\n|---:|---:|---:|---:|\n")
	for _, g := range goRows {
		fmt.Fprintf(&text, "| %d | %.2f | %.2f | **%.2f** |\n", int(g["n"]), g["withinMean"], g["withinMedian"], g["withinTrimmedMean"])
	}
	fmt.Fprintf(&text, "\n- Go reproduces R's 16 standard errors (4 per n) with a largest relative difference of **%.1g**.\n", worstOverall)
	fmt.Fprintf(&text, "- The trimmed mean has the smallest standard error of the three, by both measures, at n = **%s**.\n", strings.Join(trimmedSmallest, ", "))
	return text.String(), nil
}

func smallest(candidate float64, others ...float64) bool {
	for _, other := range others {
		if other <= candidate {
			return false
		}
	}
	return true
}

func relativeDifference(a, b float64) float64 {
	if a == b {
		return 0
	}
	return math.Abs(a-b) / math.Max(math.Abs(a), math.Abs(b))
}
