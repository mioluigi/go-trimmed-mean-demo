// Package study reproduces the course's bootstrap study, run-bootstrap-robust.R (Thomas W. Miller, May 23,
// 2024), draw for draw: the same seed, the same random numbers in the same order (R's runif, rnorm and
// sample, from internal/rrand), and the same arithmetic, with the trimmed mean computed by the package
// github.com/mioluigi/trimmedmean (spec 002).
package study

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/mioluigi/go-trimmed-mean-demo/internal/rrand"
	"github.com/mioluigi/trimmedmean"
)

// ErrConfig marks an unusable configuration.
var ErrConfig = errors.New("invalid study configuration")

// Config holds the study's settings; Course returns the course file's.
type Config struct {
	Seed               uint32
	Iterations         int     // samples drawn per sample size
	SampleSizes        []int   // n
	Replicates         int     // B, bootstrap resamples per sample
	Trim               float64 // trimming proportion
	MainMean, MainSD   float64 // main normal population
	OutlierMean        float64 // outlier normal population
	OutlierSD          float64
	OutlierProbability float64
}

// Course returns the settings of run-bootstrap-robust.R.
func Course() Config {
	return Config{Seed: 9999, Iterations: 100, SampleSizes: []int{25, 100, 225, 400}, Replicates: 100,
		Trim: 0.05, MainMean: 100, MainSD: 10, OutlierMean: 200, OutlierSD: 1, OutlierProbability: 0.01}
}

// Row is one sample: the course's four values (its studyResults row) and, added by spec 002 FR-B4, the
// textbook bootstrap standard error of each statistic within this sample (the SD of its B values).
type Row struct {
	N                                             int
	SampleMean, BootMean, BootMedian, BootTrimmed float64
	WithinMean, WithinMedian, WithinTrimmed       float64
}

// Run draws every sample in the course's order: iterations outermost, then sample sizes.
func Run(config Config) ([]Row, error) {
	if config.Iterations < 2 || config.Replicates < 2 || len(config.SampleSizes) == 0 {
		return nil, fmt.Errorf("%w: need at least 2 iterations, 2 replicates and one sample size", ErrConfig)
	}
	generator := rrand.New(config.Seed)
	rows := make([]Row, 0, config.Iterations*len(config.SampleSizes))
	for iteration := 0; iteration < config.Iterations; iteration++ {
		for _, n := range config.SampleSizes {
			row, err := runSample(generator, config, n)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row)
		}
	}
	return rows, nil
}

// runSample draws one sample of size n and bootstraps it.
func runSample(generator *rrand.Generator, config Config, n int) (Row, error) {
	sample := make([]float64, n)
	for i := range sample {
		sample[i] = mixtureDraw(generator, config)
	}
	sampleMean, err := trimmedmean.Mean(sample)
	if err != nil {
		return Row{}, err
	}
	means, medians, trimmed, err := bootstrap(generator, config, sample)
	if err != nil {
		return Row{}, err
	}
	return Row{
		N: n, SampleMean: sampleMean,
		BootMean: averageReversed(means), BootMedian: averageReversed(medians), BootTrimmed: averageReversed(trimmed),
		WithinMean: SD(means), WithinMedian: SD(medians), WithinTrimmed: SD(trimmed),
	}, nil
}

// mixtureDraw is the course's rmixnorm(): one uniform decides the population, then one normal draw.
func mixtureDraw(generator *rrand.Generator, config Config) float64 {
	if generator.Float64() >= config.OutlierProbability {
		return generator.NormalWith(config.MainMean, config.MainSD)
	}
	return generator.NormalWith(config.OutlierMean, config.OutlierSD)
}

// bootstrap draws B resamples as sample(x, n, replace = TRUE) does and evaluates the three statistics.
func bootstrap(generator *rrand.Generator, config Config, sample []float64) (means, medians, trimmed []float64, err error) {
	n := len(sample)
	resample := make([]float64, n)
	for b := 0; b < config.Replicates; b++ {
		for i := range resample {
			resample[i] = sample[generator.Index(n)]
		}
		mean, err := trimmedmean.Mean(resample)
		if err != nil {
			return nil, nil, nil, err
		}
		trimmedMean, err := trimmedmean.Mean(resample, config.Trim)
		if err != nil {
			return nil, nil, nil, err
		}
		means = append(means, mean)
		medians = append(medians, Median(resample))
		trimmed = append(trimmed, trimmedMean)
	}
	return means, medians, trimmed, nil
}

// averageReversed is R's mean of a column built by rbind(new, old): the last replicate comes first, and
// a floating-point sum depends on its order.
func averageReversed(values []float64) float64 {
	reversed := slices.Clone(values)
	slices.Reverse(reversed)
	mean, _ := trimmedmean.Mean(reversed) // values are finite and non-empty here
	return mean
}

// Median is R's median(): the middle value, or R's mean of the two middle values.
func Median(values []float64) float64 {
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	half := (len(sorted) + 1) / 2
	if len(sorted)%2 == 1 {
		return sorted[half-1]
	}
	mean, _ := trimmedmean.Mean(sorted[half-1 : half+1])
	return mean
}

// SD is R's sd(): the square root of R's var(), which takes R's two-pass mean and then the sum of squared
// deviations divided by n − 1. Every product is rounded before it is added, as in R's C code (spec 002 §4).
func SD(values []float64) float64 {
	mean, _ := trimmedmean.Mean(values)
	sum := 0.0
	for _, value := range values {
		deviation := value - mean
		sum += float64(deviation * deviation)
	}
	return math.Sqrt(sum / float64(len(values)-1))
}
