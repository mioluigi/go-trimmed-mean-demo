package study

import (
	"fmt"
	"strconv"
	"strings"
)

// Listing returns, character for character, what run-bootstrap-robust.R prints with cat(): R prints each
// number with up to 7 significant digits, and the standard errors after round(x, 2).
func Listing(config Config, summaries []Summary) string {
	var text strings.Builder
	fmt.Fprintf(&text, "\nRunning study with %s bootstrap samples\n", number(float64(config.Replicates)))
	text.WriteString("\nStudy conditions:")
	fmt.Fprintf(&text, "\n  Main population mean: %s SD: %s", number(config.MainMean), number(config.MainSD))
	fmt.Fprintf(&text, "\n  Outlier mean: %s SD: %s", number(config.OutlierMean), number(config.OutlierSD))
	fmt.Fprintf(&text, "\n  Probability of outliers: %s \n", number(config.OutlierProbability))
	fmt.Fprintf(&text, "\nTrimmed mean computed with trimming proportion %s \n", number(config.Trim))
	fmt.Fprintf(&text, "\nEstimated standard errors using %d bootstrap samples\n", config.Replicates)
	for _, s := range summaries {
		fmt.Fprintf(&text, "\nSamples of size n = %d", s.N)
		fmt.Fprintf(&text, "\n  SE Mean from Central Limit Theorem (assuming no outliers): %s", rounded(s.CLT))
		fmt.Fprintf(&text, "\n  SE Mean from Samples (with outliers): %s", rounded(s.SampleMean))
		fmt.Fprintf(&text, "\n  SE Mean from Bootstrap Samples (with outliers): %s", rounded(s.BootMean))
		fmt.Fprintf(&text, "\n  SE Median from Bootstrap Samples (with outliers): %s", rounded(s.BootMedian))
		fmt.Fprintf(&text, "\n  SE Trimmed Mean from Bootstrap Samples (with outliers): %s", rounded(s.BootTrimmed))
		text.WriteString("\n")
	}
	text.WriteString("\n----- Run Complete -----\n")
	return text.String()
}

// TextbookListing returns the added column (spec 002 FR-B4), which R's course file does not compute.
func TextbookListing(summaries []Summary) string {
	var text strings.Builder
	text.WriteString("\nAdded by this program (spec 002, Go only): the textbook bootstrap standard error,\n")
	text.WriteString("the SD of each sample's own bootstrap values, averaged over the samples\n")
	for _, s := range summaries {
		fmt.Fprintf(&text, "  n = %d: mean %s, median %s, trimmed mean %s\n", s.N,
			rounded(s.WithinMean), rounded(s.WithinMedian), rounded(s.WithinTrimmed))
	}
	return text.String()
}

// rounded is R's cat(round(x, 2)): the value rounded to 2 decimals, printed without trailing zeros.
func rounded(x float64) string {
	value, _ := strconv.ParseFloat(strconv.FormatFloat(x, 'f', 2, 64), 64)
	return number(value)
}

// number prints like R's cat(): up to 7 significant digits, no trailing zeros.
func number(x float64) string {
	return strconv.FormatFloat(x, 'g', 7, 64)
}
