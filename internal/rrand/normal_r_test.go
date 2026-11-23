package rrand

import (
	"bufio"
	"os"
	"strconv"
	"testing"
)

// readValues reads a "value"-headed column saved by r/capture-bootstrap-reference.R.
func readValues(t *testing.T, path string) []float64 {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	var values []float64
	scanner := bufio.NewScanner(file)
	scanner.Scan() // header
	for scanner.Scan() {
		value, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			t.Fatal(err)
		}
		values = append(values, value)
	}
	return values
}

// TestNormalMatchesRnorm is spec 002 AC-B1: after set.seed(9999), R 4.4.2 on the Mac drew these 5,000
// values with rnorm(); Go must draw the same, bit for bit. It also reports how many draws the other
// arithmetic (fused or not) would have matched, which is how the setting was chosen.
func TestNormalMatchesRnorm(t *testing.T) {
	want := readValues(t, "../../testdata/bootstrap-reference/rnorm-seed-9999.txt")
	if len(want) != 5000 {
		t.Fatalf("read %d draws, want 5000", len(want))
	}
	matches := map[bool]int{}
	for _, fused := range []bool{false, true} {
		generator := New(9999)
		for _, value := range want {
			const big = 134217728
			u := generator.Float64()
			u = float64(int(big*u)) + generator.Float64()
			if quantile(u/big, fused) == value {
				matches[fused]++
			}
		}
	}
	t.Logf("draws equal to R's: unfused %d, fused %d (of %d)", matches[false], matches[true], len(want))
	if matches[fusedArithmetic] != len(want) {
		t.Fatalf("with fusedArithmetic = %v, %d of %d draws equal R's", fusedArithmetic, matches[fusedArithmetic], len(want))
	}
	generator := New(9999)
	for i, value := range want {
		if got := generator.Normal(); got != value {
			t.Fatalf("draw %d: Normal() = %.17g, R %.17g", i+1, got, value)
		}
	}
}
