package study

import (
	"math"
	"testing"
)

func TestMedianLikeR(t *testing.T) {
	cases := []struct {
		values []float64
		want   float64
	}{
		{[]float64{3, 1, 2}, 2},
		{[]float64{4, 1, 3, 2}, 2.5},
		{[]float64{7}, 7},
		{[]float64{1, 100}, 50.5},
	}
	for _, tc := range cases {
		if got := Median(tc.values); got != tc.want {
			t.Errorf("Median(%v) = %v, want %v", tc.values, got, tc.want)
		}
	}
}

func TestSDLikeR(t *testing.T) {
	// R: sd(c(2, 4, 4, 4, 5, 5, 7, 9)) = sqrt(32/7)
	if got, want := SD([]float64{2, 4, 4, 4, 5, 5, 7, 9}), math.Sqrt(32.0/7); math.Abs(got-want) > 1e-15 {
		t.Errorf("SD = %.17g, want %.17g", got, want)
	}
}

func TestRunShapeAndConfig(t *testing.T) {
	config := Course()
	config.Iterations, config.Replicates = 3, 5
	rows, err := Run(config)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3*len(config.SampleSizes) || rows[0].N != 25 || rows[3].N != 400 || rows[4].N != 25 {
		t.Fatalf("rows not in the course's order (iterations outermost): %d rows", len(rows))
	}
	if _, err := Run(Config{}); err == nil {
		t.Error("empty configuration accepted")
	}
}
