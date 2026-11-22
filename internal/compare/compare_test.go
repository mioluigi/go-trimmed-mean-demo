package compare

import (
	"math"
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
