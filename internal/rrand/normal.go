package rrand

import "math"

// Normal returns the next standard normal draw exactly as R's rnorm(1) makes it with its default
// normal.kind "Inversion" (R's src/nmath/snorm.c): one uniform is not precise enough, so two are combined
// into u = floor(2^27 * u1) + u2, and qnorm(u / 2^27) is returned.
func (generator *Generator) Normal() float64 {
	const big = 134217728 // 2^27
	u := generator.Float64()
	u = float64(int(big*u)) + generator.Float64()
	return quantile(u/big, fusedArithmetic)
}

// NormalWith returns mean + sd × Normal(), as R's rnorm(1, mean, sd) does.
func (generator *Generator) NormalWith(mean, sd float64) float64 {
	return multiplyAdd(sd, generator.Normal(), mean, fusedArithmetic)
}

// fusedArithmetic says whether R's compiled qnorm and rnorm use fused multiply-adds (one rounding for
// a*b + c). Go fuses only where told to, so both forms are written out explicitly and the saved R draws
// decide (spec 002 §4; spike addendum C).
var fusedArithmetic = false

// multiplyAdd returns a*b + c, rounded once (fused) or twice.
func multiplyAdd(a, b, c float64, fused bool) float64 {
	if fused {
		return math.FMA(a, b, c)
	}
	return float64(a*b) + c // the conversion stops Go from fusing on arm64
}

// horner evaluates a polynomial with coefficients from the highest power down, as R writes them:
// (((c0*x + c1)*x + c2)*x ...).
func horner(x float64, coefficients []float64, fused bool) float64 {
	value := coefficients[0]
	for _, coefficient := range coefficients[1:] {
		value = multiplyAdd(value, x, coefficient, fused)
	}
	return value
}

// Coefficients of Wichura's algorithm AS 241 (1988), as in R's src/nmath/qnorm.c, highest power first.
var (
	centralNumerator = []float64{2509.0809287301226727, 33430.575583588128105, 67265.770927008700853,
		45921.953931549871457, 13731.693765509461125, 1971.5909503065514427, 133.14166789178437745,
		3.387132872796366608}
	centralDenominator = []float64{5226.4952788528545610, 28729.085735721942674, 39307.89580009271061,
		21213.794301586595867, 5394.1960214247511077, 687.1870074920579083, 42.313330701600911252, 1}
	tailNumerator = []float64{7.7454501427834140764e-4, .0227238449892691845833, .24178072517745061177,
		1.27045825245236838258, 3.64784832476320460504, 5.7694972214606914055, 4.6303378461565452959,
		1.42343711074968357734}
	tailDenominator = []float64{1.05075007164441684324e-9, 5.475938084995344946e-4, .0151986665636164571966,
		.14810397642748007459, .68976733498510000455, 1.6763848301838038494, 2.05319162663775882187, 1}
	farNumerator = []float64{2.01033439929228813265e-7, 2.71155556874348757815e-5, .0012426609473880784386,
		.026532189526576123093, .29656057182850489123, 1.7848265399172913358, 5.4637849111641143699,
		6.6579046435011037772}
	farDenominator = []float64{2.04426310338993978564e-15, 1.4215117583164458887e-7,
		1.8463183175100546818e-5, 7.868691311456132591e-4, .0148753612908506148525, .13692988092273580531,
		.59983220655588793769, 1}
)

// quantile is R's qnorm(p, 0, 1) for 0 < p < 1, lower tail, not log: AS 241 in three regions.
func quantile(p float64, fused bool) float64 {
	q := p - 0.5
	if math.Abs(q) <= 0.425 {
		r := 0.180625 - q*q
		return q * horner(r, centralNumerator, fused) / horner(r, centralDenominator, fused)
	}
	tailProbability := p
	if q > 0 {
		tailProbability = 0.5 - p + 0.5 // R's R_DT_CIv(p): 1 - p, written as R writes it
	}
	r := math.Sqrt(-correctlyRoundedLog(tailProbability))
	var value float64
	if r <= 5 {
		r -= 1.6
		value = horner(r, tailNumerator, fused) / horner(r, tailDenominator, fused)
	} else {
		r -= 5
		value = horner(r, farNumerator, fused) / horner(r, farDenominator, fused)
	}
	if q < 0 {
		return -value
	}
	return value
}
