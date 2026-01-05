package benchmarking

import (
	"time"
)

// ComparisonResult represents the statistical verdict.
type ComparisonResult int

const (
	ResultIndistinguishable ComparisonResult = 0
	ResultAFaster           ComparisonResult = -1
	ResultASlower           ComparisonResult = 1
)

type Parameter int

// PreparationFn sets up the environment and data for a specific parameter step.
type PreparationFn[TData any] func(p Parameter) TData

// BenchFunction performs the action to be measured.
type BenchFunction[TData any] func(data TData)

// CleanupFn tears down resources (memory, global state resets).
type CleanupFn[TData any] func(data TData)

// Measurer captures the metric (time, memory) from the function.
type Measurer[TData any] func(fn BenchFunction[TData], data TData) float64

// StatisticalStrategy is the injected judge (Dependency Inversion).
type StatisticalStrategy func(samplesA, samplesB []float64) ComparisonResult

type SearchConfig struct {
	MinParam int
	MaxParam int
	Samples  int
}

type CrossoverResult struct {
	LowIndex          int
	HighIndex         int
	Indistinguishable bool
	FinalStatsA       []float64
	FinalStatsB       []float64
}

func FindCrossover[TData any](
	benchA, benchB BenchFunction[TData],
	prepA, prepB PreparationFn[TData],
	cleanA, cleanB CleanupFn[TData],
	measure Measurer[TData],
	judge StatisticalStrategy,
	cfg SearchConfig,
) CrossoverResult {

	low := cfg.MinParam
	high := cfg.MaxParam
	var lastMid int
	var rawA, rawB []float64

	for low <= high {
		mid := low + (high-low)/2
		lastMid = mid

		dataA := prepA(Parameter(mid))
		rawA = collectData(benchA, dataA, measure, cfg.Samples)
		cleanA(dataA)

		dataB := prepB(Parameter(mid))
		rawB = collectData(benchB, dataB, measure, cfg.Samples)
		cleanB(dataB)

		verdict := judge(rawA, rawB)

		switch verdict {
		case ResultIndistinguishable:
			return CrossoverResult{
				LowIndex:          low,
				HighIndex:         high,
				Indistinguishable: true,
				FinalStatsA:       rawA,
				FinalStatsB:       rawB,
			}
		case ResultASlower:
			high = mid - 1
		case ResultAFaster:
			low = mid + 1
		}
	}

	return CrossoverResult{
		LowIndex:          lastMid,
		HighIndex:         lastMid,
		Indistinguishable: false,
		FinalStatsA:       rawA,
		FinalStatsB:       rawB,
	}
}

func collectData[TData any](
	fn BenchFunction[TData],
	data TData,
	measure Measurer[TData],
	samples int,
) []float64 {
	results := make([]float64, samples)

	fn(data)

	for i := 0; i < samples; i++ {
		results[i] = measure(fn, data)
	}
	return results
}

func MeasureTime[TData any](fn BenchFunction[TData], data TData) float64 {
	start := time.Now()
	fn(data)
	return float64(time.Since(start).Nanoseconds())
}
