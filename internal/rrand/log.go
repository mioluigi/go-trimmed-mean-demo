package rrand

import (
	"math"
	"math/big"
	"sync"
)

// logPrecision is the working precision, in bits, of correctlyRoundedLog: far more than float64's 53, so
// rounding the result to float64 gives the correctly rounded logarithm (except in pathological cases
// vanishingly rarer than any input this package sees).
const logPrecision = 160

var (
	ln2Once sync.Once
	ln2     *big.Float
)

// correctlyRoundedLog returns ln(x) rounded to the nearest float64, for x > 0. R's qnorm calls the C
// library's log, which is correctly rounded on these platforms, while Go's math.Log can be one unit in
// the last place away (spike 002: 4 of 4 sampled tail values), and that unit would change R's normal draws.
func correctlyRoundedLog(x float64) float64 {
	ln2Once.Do(func() { ln2 = atanhTimesTwo(new(big.Float).SetPrec(logPrecision).Quo(one(), big.NewFloat(3))) })
	fraction, exponent := math.Frexp(x) // x = fraction × 2^exponent, fraction in [0.5, 1)
	m := new(big.Float).SetPrec(logPrecision).SetFloat64(fraction)
	numerator := new(big.Float).SetPrec(logPrecision).Sub(m, one())
	denominator := new(big.Float).SetPrec(logPrecision).Add(m, one())
	logFraction := atanhTimesTwo(numerator.Quo(numerator, denominator)) // ln m = 2 atanh((m−1)/(m+1))
	scaled := new(big.Float).SetPrec(logPrecision).Mul(ln2, big.NewFloat(float64(exponent)))
	result, _ := scaled.Add(scaled, logFraction).Float64()
	return result
}

// atanhTimesTwo returns 2 atanh(t) = 2 (t + t³/3 + t⁵/5 + …) for |t| ≤ 1/3, summed until the terms fall
// below the working precision.
func atanhTimesTwo(t *big.Float) *big.Float {
	square := new(big.Float).SetPrec(logPrecision).Mul(t, t)
	power := new(big.Float).SetPrec(logPrecision).Set(t)
	sum := new(big.Float).SetPrec(logPrecision).Set(t)
	term := new(big.Float).SetPrec(logPrecision)
	for k := int64(3); ; k += 2 {
		power.Mul(power, square)
		term.Quo(power, new(big.Float).SetInt64(k))
		if term.Sign() == 0 || term.MantExp(nil)-sum.MantExp(nil) < -logPrecision {
			break
		}
		sum.Add(sum, term)
	}
	return sum.Mul(sum, big.NewFloat(2))
}

func one() *big.Float { return new(big.Float).SetPrec(logPrecision).SetInt64(1) }
