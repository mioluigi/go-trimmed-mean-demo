package main

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

// headlineTrim is the assignment's comparison: R's mean(x, trim = 0.05).
const headlineTrim = 0.05

func isHeadline(row comparisonRow) bool {
	return row.method == "mean(x, trim)" && row.low == headlineTrim
}

// renderHeadline states the conclusions the README's summary rests on, generated so that they change
// when the results change (L-W6-08).
func renderHeadline(rows []comparisonRow, benchmarks []benchmark) string {
	passed, integerExact, integerRows, worstFloat := 0, 0, 0, 0.0
	for _, row := range rows {
		if row.verdict == "PASS" {
			passed++
		}
		if isHeadline(row) && row.kind == "integer" {
			integerRows++
			if row.ulps == 0 {
				integerExact++
			}
		}
		if isHeadline(row) && row.kind == "double" {
			worstFloat = max(worstFloat, math.Abs(row.goValue-row.rValue))
		}
	}
	slowest := slices.MaxFunc(benchmarks, func(a, b benchmark) int { return int(math.Copysign(1, a.nsPerOp-b.nsPerOp)) })
	return fmt.Sprintf("- **%d of %d** comparisons with R agree under the rule of spec §7.\n"+
		"- At `trim = 0.05`, Go equals R **exactly** on **%d of %d** integer samples; on the float samples the largest difference from R is **%.2g**.\n"+
		"- Slowest benchmark (`%s`): **%s** per call, **%s** allocated in **%d** allocation(s).\n",
		passed, len(rows), integerExact, integerRows, worstFloat,
		slowest.name, duration(slowest.nsPerOp), bytes(slowest.bytesPerOp), slowest.allocsPerOp)
}

func renderComparison(rows []comparisonRow) string {
	var text strings.Builder
	text.WriteString("| Sample | n | Kind | Go `Mean(x, 0.05)` | R `mean(x, trim = 0.05)` | Difference | ULPs | Rule | Verdict |\n")
	text.WriteString("|---|---:|---|---:|---:|---:|---:|---|---|\n")
	for _, row := range rows {
		if isHeadline(row) {
			fmt.Fprintf(&text, "| %s | %d | %s | %.17g | %.17g | %.2g | %d | %s | %s |\n", row.sample, row.n,
				row.kind, row.goValue, row.rValue, math.Abs(row.goValue-row.rValue), row.ulps, row.rule, row.verdict)
		}
	}
	return text.String()
}

// renderAgreement summarises all rows by rule: how many, how many passed, the largest ULP distance.
func renderAgreement(rows []comparisonRow) string {
	type summary struct {
		rows, passed int
		worstULPs    uint64
	}
	byRule := map[string]*summary{}
	var rules []string
	for _, row := range rows {
		key := row.rule + " · " + row.kind
		if byRule[key] == nil {
			byRule[key] = &summary{}
			rules = append(rules, key)
		}
		entry := byRule[key]
		entry.rows++
		if row.verdict == "PASS" {
			entry.passed++
		}
		entry.worstULPs = max(entry.worstULPs, row.ulps)
	}
	slices.Sort(rules)
	var text strings.Builder
	text.WriteString("| Rule · kind | Rows | Agree | Largest distance from R's built-in (ULPs) |\n|---|---:|---:|---:|\n")
	for _, rule := range rules {
		entry := byRule[rule]
		fmt.Fprintf(&text, "| %s | %d | %d | %d |\n", rule, entry.rows, entry.passed, entry.worstULPs)
	}
	return text.String()
}

func renderBenchmarks(benchmarks []benchmark) string {
	var text strings.Builder
	text.WriteString("| Benchmark (`trim = 0.05`) | Runs | Median time per call | Memory per call | Allocations |\n|---|---:|---:|---:|---:|\n")
	for _, b := range benchmarks {
		fmt.Fprintf(&text, "| %s | %d | %s | %s | %d |\n", b.name, b.runs, duration(b.nsPerOp), bytes(b.bytesPerOp), b.allocsPerOp)
	}
	return text.String()
}

// median of one column of repeated measurements.
func median(runs [][3]float64, column int) float64 {
	values := make([]float64, len(runs))
	for i, run := range runs {
		values[i] = run[column]
	}
	slices.Sort(values)
	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle]
	}
	return (values[middle-1] + values[middle]) / 2
}

func duration(nanoseconds float64) string {
	switch {
	case nanoseconds >= 1e6:
		return fmt.Sprintf("%.2f ms", nanoseconds/1e6)
	case nanoseconds >= 1e3:
		return fmt.Sprintf("%.2f µs", nanoseconds/1e3)
	default:
		return fmt.Sprintf("%.0f ns", nanoseconds)
	}
}

func bytes(count float64) string {
	switch {
	case count >= 1<<20:
		return fmt.Sprintf("%.2f MiB", count/(1<<20))
	case count >= 1<<10:
		return fmt.Sprintf("%.1f KiB", count/(1<<10))
	default:
		return fmt.Sprintf("%.0f B", count)
	}
}
