package compare

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestULPDistance(t *testing.T) {
	one := 1.0
	next := math.Nextafter(one, 2)
	cases := []struct {
		a, b float64
		want uint64
	}{
		{one, one, 0},
		{one, next, 1},
		{next, one, 1},
		{0, math.Copysign(0, -1), 0},
		{math.SmallestNonzeroFloat64, -math.SmallestNonzeroFloat64, 2},
	}
	for _, tc := range cases {
		if got := ULPDistance(tc.a, tc.b); got != tc.want {
			t.Errorf("ULPDistance(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCheckExactRules(t *testing.T) {
	exactCases := []Reference{
		{Kind: "integer", Method: BuiltIn, Low: 0.05, High: 0.05, Value: 1},
		{Kind: "double", Method: BuiltIn, Low: 0, High: 0, Value: 1},
		{Kind: "double", Method: Defined, Low: 0.1, High: 0.2, Value: 1},
	}
	for _, reference := range exactCases {
		if !Check(reference, 1, []float64{1}).Pass {
			t.Errorf("%+v: equal values should pass", reference)
		}
		result := Check(reference, math.Nextafter(1, 2), []float64{1})
		if result.Pass || result.Rule != RuleExact || result.ULPs != 1 {
			t.Errorf("%+v: a 1-ULP difference should fail the exact rule, got %+v", reference, result)
		}
	}
}

func TestCheckToleranceRule(t *testing.T) {
	reference := Reference{Kind: "double", Method: BuiltIn, Low: 0.05, High: 0.05, Value: 1, FullSortValue: 1}
	kept := []float64{-2, 1, 4} // mean absolute value 7/3
	limit := Tolerance * 7.0 / 3

	if result := Check(reference, 1, kept); !result.Pass || result.Rule != RuleTolerance {
		t.Errorf("identical value: got %+v", result)
	}
	withinR := Reference{Kind: "double", Method: BuiltIn, Low: 0.05, High: 0.05, Value: 1 + limit/2, FullSortValue: 1}
	if result := Check(withinR, 1, kept); !result.Pass {
		t.Errorf("within tolerance of R's built-in, exact on full sort: got %+v", result)
	}
	beyondR := Reference{Kind: "double", Method: BuiltIn, Low: 0.05, High: 0.05, Value: 1 + 2*limit, FullSortValue: 1}
	if result := Check(beyondR, 1, kept); result.Pass {
		t.Errorf("beyond tolerance of R's built-in: got %+v", result)
	}
	fullSortOff := Reference{Kind: "double", Method: BuiltIn, Low: 0.05, High: 0.05, Value: 1, FullSortValue: math.Nextafter(1, 2)}
	if result := Check(fullSortOff, 1, kept); result.Pass {
		t.Errorf("1 ULP from R's full-sort mean must fail: got %+v", result)
	}
}

// writeReference writes a small reference file for the reader tests.
func writeReference(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "reference.tsv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestReadReferencesByColumnName checks that columns are found by their header name (review R2, [C]),
// so a reordered file still reads correctly, and a missing column is named in the error.
func TestReadReferencesByColumnName(t *testing.T) {
	reordered := "value\tsample\tkind\tn\tlow\thigh\tmethod\tdropped_low\tdropped_high\tfull_sort_value\textra\n" +
		"2.5\tfloats-4\tdouble\t4\t0.25\t0.25\tmean(x, trim)\t1\t1\t2.5\tignored\n"
	references, err := ReadReferences(writeReference(t, reordered))
	if err != nil {
		t.Fatal(err)
	}
	want := Reference{Sample: "floats-4", Kind: "double", Method: BuiltIn, N: 4, Low: 0.25, High: 0.25,
		Value: 2.5, FullSortValue: 2.5, DropLow: 1, DropHigh: 1}
	if len(references) != 1 || references[0] != want {
		t.Fatalf("got %+v, want %+v", references, want)
	}

	_, err = ReadReferences(writeReference(t, "sample\tkind\nfloats-4\tdouble\n"))
	if err == nil || !strings.Contains(err.Error(), `column "n"`) {
		t.Fatalf("missing column: got %v, want an error naming column \"n\"", err)
	}
	_, err = ReadReferences(writeReference(t, strings.Replace(reordered, "\t4\t0.25", "\tfour\t0.25", 1)))
	if err == nil || !strings.Contains(err.Error(), "line 2") || !strings.Contains(err.Error(), `"n"`) {
		t.Fatalf("bad number: got %v, want an error naming line 2 and column \"n\"", err)
	}
}

// TestCheckWithNoKeptValues checks the guard for an empty kept slice (review R2): the limit is 0, not NaN.
func TestCheckWithNoKeptValues(t *testing.T) {
	reference := Reference{Kind: "double", Method: BuiltIn, Low: 0.05, High: 0.05, Value: 1, FullSortValue: 1}
	if result := Check(reference, 1, nil); math.IsNaN(result.Limit) || result.Limit != 0 {
		t.Fatalf("limit with no kept values: got %v, want 0", result.Limit)
	}
}
