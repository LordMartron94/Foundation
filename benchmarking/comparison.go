package benchmarking

import (
	"fmt"
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
type Measurer[TData any] func(fn BenchFunction[TData], data TData, targetRuntime time.Duration) float64

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
	targetRuntime time.Duration,
) CrossoverResult {

	low := cfg.MinParam
	high := cfg.MaxParam
	var lastMid int
	var rawA, rawB []float64

	for low <= high {
		mid := low + (high-low)/2
		lastMid = mid

		dataA := prepA(Parameter(mid))
		rawA = collectData(benchA, dataA, measure, cfg.Samples, targetRuntime)
		cleanA(dataA)

		dataB := prepB(Parameter(mid))
		rawB = collectData(benchB, dataB, measure, cfg.Samples, targetRuntime)
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
	targetRuntime time.Duration,
) []float64 {
	results := make([]float64, samples)

	fn(data)

	for i := 0; i < samples; i++ {
		results[i] = measure(fn, data, targetRuntime)
	}
	return results
}

func MeasureTimeAdaptive[TData any](fn BenchFunction[TData], data TData, targetRuntime time.Duration, enableLogging bool) float64 {
	// 1. Calibration Phase
	const calibrationThreshold = 100 * time.Microsecond
	iterations := 1
	var duration time.Duration

	for {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			fn(data)
		}
		duration = time.Since(start)

		if duration >= calibrationThreshold {
			break
		}
		if duration >= targetRuntime {
			break
		}

		iterations *= 2
	}

	// 2. Extrapolation
	var targetCount int
	if duration >= targetRuntime {
		targetCount = iterations
	} else {
		projectedOps := float64(targetRuntime) / float64(duration) * float64(iterations)
		targetCount = int(projectedOps)
		if targetCount == 0 {
			targetCount = 1
		}
	}

	if enableLogging {
		fmt.Printf("   [Adaptive] Target: %v | Warmup: %v (%d ops) -> Sprint: %d ops\n",
			targetRuntime, duration, iterations, targetCount)
	}

	// 3. Measurement Sprint
	start := time.Now()
	for i := 0; i < targetCount; i++ {
		fn(data)
	}
	elapsed := time.Since(start)

	return float64(elapsed.Nanoseconds()) / float64(targetCount)
}
