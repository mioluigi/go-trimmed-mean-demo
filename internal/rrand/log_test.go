package rrand

import (
	"math"
	"testing"
)

// TestCorrectlyRoundedLog checks exact values, and that the result is never more than half a unit in the
// last place from the true logarithm: it equals math.Log or a float64 neighbour of it, whichever is
// nearest at high precision.
func TestCorrectlyRoundedLog(t *testing.T) {
	cases := map[float64]float64{1: 0, 2: math.Ln2, 0.5: -math.Ln2, math.E: 1}
	for x, want := range cases {
		if got := correctlyRoundedLog(x); got != want {
			t.Errorf("log(%v) = %.17g, want %.17g", x, got, want)
		}
	}
	// The four tail values where Go's math.Log differs from the C library (spike 002), as libm gives them.
	cLibrary := map[float64]float64{
		0.044984467446049226: -3.1014380166615294,
		0.040273757324063908: -3.2120552051669633,
		0.074711717008360878: -2.5941183449487739,
		0.036802651420941968: -3.3021853869205722,
	}
	for x, want := range cLibrary {
		if got := correctlyRoundedLog(x); got != want {
			t.Errorf("log(%.17g) = %.17g, C library %.17g", x, got, want)
		}
	}
}
