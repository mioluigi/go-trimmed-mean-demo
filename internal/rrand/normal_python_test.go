package rrand

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestQuantileMatchesIndependentAS241 compares quantile (unfused) with Python's
// statistics.NormalDist.inv_cdf, an independent implementation of the same algorithm AS 241 that rounds
// every operation: the coefficients and the order of operations must agree bit for bit.
func TestQuantileMatchesIndependentAS241(t *testing.T) {
	file, err := os.Open("testdata/python-as241-quantiles.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	rows, mismatches := 0, 0
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		p, _ := strconv.ParseFloat(fields[0], 64)
		want, _ := strconv.ParseFloat(fields[1], 64)
		rows++
		if got := quantile(p, false); got != want {
			mismatches++
			if mismatches <= 5 {
				t.Errorf("quantile(%.17g) = %.17g, Python %.17g", p, got, want)
			}
		}
	}
	if rows < 3000 {
		t.Fatalf("read %d rows", rows)
	}
	if mismatches > 0 {
		t.Errorf("%d of %d quantiles differ", mismatches, rows)
	}
}
