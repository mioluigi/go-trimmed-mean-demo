package study

import "math"

// Summary is one sample size's results: the course's five lines, then the added textbook column.
type Summary struct {
	N                                             int
	CLT                                           float64 // MainSD / √n
	SampleMean, BootMean, BootMedian, BootTrimmed float64 // the course's SDs over the iterations
	WithinMean, WithinMedian, WithinTrimmed       float64 // averages over the iterations of the within-sample SDs
}

// Summarize computes, per sample size, what the course prints (SDs over iterations, R's sd) and the
// textbook bootstrap standard errors (means over iterations of each sample's own bootstrap SD).
func Summarize(rows []Row, config Config) []Summary {
	summaries := make([]Summary, 0, len(config.SampleSizes))
	for _, n := range config.SampleSizes {
		var columns [7][]float64
		for _, row := range rows {
			if row.N == n {
				for i, value := range []float64{row.SampleMean, row.BootMean, row.BootMedian, row.BootTrimmed,
					row.WithinMean, row.WithinMedian, row.WithinTrimmed} {
					columns[i] = append(columns[i], value)
				}
			}
		}
		summaries = append(summaries, Summary{
			N: n, CLT: config.MainSD / math.Sqrt(float64(n)),
			SampleMean: SD(columns[0]), BootMean: SD(columns[1]), BootMedian: SD(columns[2]), BootTrimmed: SD(columns[3]),
			WithinMean: average(columns[4]), WithinMedian: average(columns[5]), WithinTrimmed: average(columns[6]),
		})
	}
	return summaries
}

func average(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}
