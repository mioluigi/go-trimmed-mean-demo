package study

import (
	"encoding/csv"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/mioluigi/go-trimmed-mean-demo/internal/rrand"
)

const reference = "../../testdata/bootstrap-reference/"

// readTable reads a tab-separated file with a header into rows of named fields.
func readTable(t *testing.T, name string) []map[string]float64 {
	t.Helper()
	file, err := os.Open(reference + name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]map[string]float64, 0, len(records)-1)
	for _, record := range records[1:] {
		row := map[string]float64{}
		for i, name := range records[0] {
			value, err := strconv.ParseFloat(record[i], 64)
			if err != nil {
				t.Fatalf("%s %s: %v", name, record[i], err)
			}
			row[name] = value
		}
		rows = append(rows, row)
	}
	return rows
}

// TestMixtureDrawsMatchR checks 2,000 draws made as the course's rmixnorm() makes them (AC-B1).
func TestMixtureDrawsMatchR(t *testing.T) {
	file, err := os.ReadFile(reference + "mixture-seed-9999.txt")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Fields(string(file))[1:]
	generator, config := rrand.New(9999), Course()
	for i, line := range lines {
		want, _ := strconv.ParseFloat(line, 64)
		if got := mixtureDraw(generator, config); got != want {
			t.Fatalf("draw %d: Go %.17g, R %.17g", i+1, got, want)
		}
	}
	if len(lines) != 2000 {
		t.Fatalf("read %d draws, want 2000", len(lines))
	}
}

// TestStudyMatchesR is AC-B2: every one of the 400 rows of the course's studyResults, and the 16 standard
// errors, against R's values at full precision. The mean and median columns must be identical (same draws,
// same arithmetic); the trimmed-mean column may differ in the last bits, because R adds the kept values in
// the order of its partial sort (spec 001 §7), so it and the standard errors are held to 1e-9 relative.
func TestStudyMatchesR(t *testing.T) {
	config := Course()
	rows, err := Run(config)
	if err != nil {
		t.Fatal(err)
	}
	want := readTable(t, "study-results.tsv")
	if len(want) != len(rows) {
		t.Fatalf("R has %d rows, Go %d", len(want), len(rows))
	}
	worstTrimmed := 0.0
	for i, row := range rows {
		r := want[i]
		exact := int(r["n"]) == row.N && r["sampleMean"] == row.SampleMean && r["bootMean"] == row.BootMean &&
			r["bootMedian"] == row.BootMedian
		if !exact {
			t.Fatalf("row %d: Go %+v, R %v", i+1, row, r)
		}
		worstTrimmed = math.Max(worstTrimmed, relative(row.BootTrimmed, r["bootTrimmedMean"]))
	}
	t.Logf("largest relative difference in the trimmed-mean column: %.3g", worstTrimmed)
	if worstTrimmed > 1e-9 {
		t.Errorf("trimmed-mean column differs from R by up to %.3g (limit 1e-9)", worstTrimmed)
	}
	for i, summary := range Summarize(rows, config) {
		r := readTable(t, "standard-errors.tsv")[i]
		for name, got := range map[string]float64{"clt": summary.CLT, "sampleMean": summary.SampleMean,
			"bootMean": summary.BootMean, "bootMedian": summary.BootMedian, "bootTrimmedMean": summary.BootTrimmed} {
			if difference := relative(got, r[name]); difference > 1e-9 {
				t.Errorf("n = %d %s: Go %.17g, R %.17g (relative %.3g)", summary.N, name, got, r[name], difference)
			}
		}
	}
}

// TestListingMatchesR is AC-B3: the printed output equals the course file's own, character for character.
func TestListingMatchesR(t *testing.T) {
	want, err := os.ReadFile(reference + "course-listing.txt")
	if err != nil {
		t.Fatal(err)
	}
	config := Course()
	rows, err := Run(config)
	if err != nil {
		t.Fatal(err)
	}
	got := Listing(config, Summarize(rows, config))
	if strings.TrimRight(got, "\n") != strings.TrimRight(string(want), "\n") {
		t.Fatalf("listing differs from R's:\n--- Go ---\n%s\n--- R ---\n%s", got, want)
	}
}

func relative(a, b float64) float64 {
	if a == b {
		return 0
	}
	return math.Abs(a-b) / math.Max(math.Abs(a), math.Abs(b))
}
