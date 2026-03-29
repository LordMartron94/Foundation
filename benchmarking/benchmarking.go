package benchmarking

import (
	"fmt"
	"foundation/benchhost"
	"foundation/benchreport"
	"memcore"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/trace"
	"strconv"
	"strings"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------
// Configuration & Types
// -----------------------------------------------------------------------------

/*
BenchmarkMetricsConfig provides optional configuration for advanced throughput metrics.

Use cases:
- Configuring FLOPS reporting for computational benchmarks
- Specifying manual memory bytes per operation for accurate memory throughput
- Specifying items processed per operation
- Telemetry controls JSON sidecar host/wall/HW context (see BenchmarkTelemetryConfig)

Fields:
- FLOPSPerOp: Floating point operations per benchmark operation.
- BytesPerOp: Manual memory bytes accessed per operation (for bandwidth calculation).
- ItemsPerOp: Logical items processed per operation.
- Telemetry: Host and wall-clock export options (zero value = defaults-on).
*/
type BenchmarkMetricsConfig struct {
	FLOPSPerOp float64
	BytesPerOp float64
	ItemsPerOp float64
	/*
		WarmupIterations runs warmupFn that many times after prepare/GC and before b.ResetTimer().
		When zero, BENCHMARK_WARMUP_ITERATIONS may still supply a positive count. Warmup is skipped when warmupFn is nil.
	*/
	WarmupIterations int
	Telemetry        BenchmarkTelemetryConfig
}

// -----------------------------------------------------------------------------
// Benchmark Runners
// -----------------------------------------------------------------------------

/*
BenchmarkWithMetricsConfig executes a benchmark with comprehensive metrics and throughput reporting.

It wraps the standard benchmark execution with:
1. Memory cleanup (GC/FreeOSMemory) before execution.
2. Panic recovery.
3. Reporting of GC, Heap, and Allocation metrics.
4. Reporting of Throughput (ops/sec), FLOPS (flops/sec), and Bandwidth (manual.bytes/op).
5. Optional JSONL export when BENCHMARK_METRICS_JSON is set.
6. Optional runtime/trace when BENCHMARK_TRACE_OUT is set.
7. Optional warmupFn iterations (config.WarmupIterations or BENCHMARK_WARMUP_ITERATIONS) before the timed region.
*/
func BenchmarkWithMetricsConfig[data any](
	b *testing.B,
	config BenchmarkMetricsConfig,
	prepareFn func(b *testing.B) data,
	warmupFn func(data data),
	benchmarkFn func(data data, b *testing.B),
	cleanupFn func(data data, b *testing.B),
) {
	name := b.Name()
	fmt.Fprintf(os.Stderr, "🔹 Running %s...\n", name)

	path := exportPath()
	scratch := newMetricExportScratch(path)
	exportRegisterScratch(b, scratch)
	defer exportUnregisterScratch(b)

	var hwStatus string
	var hwSess *benchhost.BenchhostHWCounterSession
	if path != "" {
		if config.Telemetry.OmitHWCounters {
			hwStatus = "omitted"
		} else {
			var msg string
			hwSess, msg = benchhost.BenchhostHWCounterOpen()
			if hwSess == nil {
				hwStatus = msg
			} else {
				hwStatus = "ok"
			}
		}
	}

	// 1. Preparation
	preparedData := prepareFn(b)
	b.ReportAllocs()

	debug.FreeOSMemory()
	runtime.GC()

	warmN := config.WarmupIterations
	if warmN <= 0 {
		if v := strings.TrimSpace(os.Getenv(EnvBenchmarkWarmupIterations)); v != "" {
			if x, err := strconv.Atoi(v); err == nil && x > 0 {
				warmN = x
			}
		}
	}
	if warmN > 0 && warmupFn != nil {
		for i := 0; i < warmN; i++ {
			warmupFn(preparedData)
		}
	}

	var panicValue any
	var panicStack []byte

	var before, after runtime.MemStats

	var wallStart, wallEnd time.Time
	if path != "" && !config.Telemetry.OmitWallTimestamps {
		wallStart = time.Now()
	}

	// 2. Execution
	b.ResetTimer()
	runtime.ReadMemStats(&before)

	if hwSess != nil {
		benchhost.BenchhostHWCounterResetEnable(hwSess)
	}

	tracePath := os.Getenv(EnvBenchmarkTraceOut)
	var traceFile *os.File
	traceOn := false
	if tracePath != "" {
		if err := os.MkdirAll(filepath.Dir(tracePath), 0755); err == nil {
			traceFile, _ = os.Create(tracePath)
		}
		if traceFile != nil {
			if err := trace.Start(traceFile); err == nil {
				traceOn = true
			}
		}
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicValue = r
				panicStack = debug.Stack()
			}
		}()
		benchmarkFn(preparedData, b)
	}()

	if traceOn {
		trace.Stop()
	}
	if traceFile != nil {
		_ = traceFile.Close()
	}

	var hwCycles, hwIns, hwMiss uint64
	hwReadOK := false
	if hwSess != nil {
		hwCycles, hwIns, hwMiss, hwReadOK = benchhost.BenchhostHWCounterDisableRead(hwSess)
	}
	benchhost.BenchhostHWCounterClose(hwSess)
	hwSess = nil

	b.StopTimer()

	if path != "" && !config.Telemetry.OmitWallTimestamps {
		wallEnd = time.Now()
	}

	runtime.ReadMemStats(&after)

	// 3. Cleanup
	if cleanupFn != nil {
		cleanupFn(preparedData, b)
	}

	memcore.MemmapUnmapAllRegions()
	memcore.MemcoreMarkManagementStateReset(false)
	debug.FreeOSMemory()

	// 4. Panic Reporting
	if panicValue != nil {
		fmt.Fprintf(os.Stderr, "\n🔥 Benchmark panic in %s: %v\n", name, panicValue)
		if len(panicStack) > 0 {
			fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", panicStack)
		}
		os.Exit(1)
	}

	// 5. Metric Reporting
	reportStandardMetrics(b, before, after, scratch)
	reportThroughputMetrics(b, config, scratch)

	if path != "" && b.N > 0 {
		if d := b.Elapsed(); d > 0 {
			exportAddSynthetic(scratch, float64(d.Nanoseconds())/float64(b.N), "ns/op", benchreport.MetricKindDurationNS)
		}
		exportAddSynthetic(scratch, float64(after.Mallocs-before.Mallocs)/float64(b.N), "allocs/op", benchreport.MetricKindCount)
		bAlloc := int64(after.TotalAlloc) - int64(before.TotalAlloc)
		exportAddSynthetic(scratch, float64(bAlloc)/float64(b.N), "B/op", benchreport.MetricKindBytes)
	}

	exportFlush(path, b, scratch, config.Telemetry, wallStart, wallEnd, hwCycles, hwIns, hwMiss, hwReadOK, hwStatus)

	fmt.Fprintf(os.Stderr, "✅ Finished %s\n", name)
}

/*
BenchmarkSetup creates a benchmark subtest with a structured naming convention.

This is a convenience wrapper around BenchmarkWithMetrics that creates a properly
named sub-benchmark using the test configuration and variant. The naming format
is "testConfiguration/testVariant", which helps organize related benchmarks.

Use this function when you want to run multiple variants of the same benchmark
with different configurations (e.g., different data sizes, algorithms, etc.).

Parameters:
- b: Benchmark context from testing.B
- testVariant: Variant identifier (e.g., "insert", "access", "small", "large")
- testConfiguration: Configuration identifier (e.g., "hashmap", "vector", "allocator")
- prepareFn: Function to prepare benchmark data
- testFn: Function containing the actual benchmark code
- cleanupFn: Function to clean up resources (may be nil)

Example usage:

	BenchmarkSetup(b, "insert", "hashmap",
		func(b *testing.B) hashMap { return createHashMap() },
		func(m hashMap, b *testing.B) {
			// benchmark code here
		},
		nil,
	)
	This creates a benchmark named "hashmap/insert"

Time complexity: O(n) - where n is the work performed by testFn
Space complexity: O(1) - minimal overhead
*/
func BenchmarkSetup[data any](
	b *testing.B,
	testVariant, testConfiguration string,
	prepareFn func(b *testing.B) data,
	testFn func(data data, b *testing.B),
	cleanupFn func(data data, b *testing.B),
) {
	b.Run(testConfiguration+"/"+testVariant, func(b *testing.B) {
		BenchmarkWithMetrics(b, prepareFn, testFn, cleanupFn)
	})
}

/*
BenchmarkWithMetrics is a convenience wrapper around BenchmarkWithMetricsConfig
for cases where no specific throughput configuration is needed.
*/
func BenchmarkWithMetrics[data any](
	b *testing.B,
	prepareFn func(b *testing.B) data,
	benchmarkFn func(data data, b *testing.B),
	cleanupFn func(data data, b *testing.B),
) {
	BenchmarkWithMetricsConfig(b, BenchmarkMetricsConfig{}, prepareFn, nil, benchmarkFn, cleanupFn)
}

// -----------------------------------------------------------------------------
// Batched Execution Helpers
// -----------------------------------------------------------------------------

/*
RunBatchedBenchmark executes a benchmark function with memory-aware GC management.

It triggers GC manually if heap growth exceeds thresholds, ensuring that long-running
benchmarks don't crash due to OOM while excluding GC time from the timer.
*/
func RunBatchedBenchmark(b *testing.B, workFn func(i int), maxHeapGrowth, maxHeapSize uint64, memoryCheckInterval int) {
	b.StopTimer()
	var initialMemStats runtime.MemStats
	runtime.ReadMemStats(&initialMemStats)
	initialHeap := initialMemStats.HeapAlloc
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		workFn(i)

		if (i+1)%memoryCheckInterval == 0 {
			b.StopTimer()
			if CheckMemoryPressure(initialHeap, maxHeapGrowth, maxHeapSize) {
				runtime.GC()
				runtime.ReadMemStats(&initialMemStats)
				initialHeap = initialMemStats.HeapAlloc
			}
			b.StartTimer()
		}
	}
}

// -----------------------------------------------------------------------------
// Internal Reporting Helpers
// -----------------------------------------------------------------------------

func reportStandardMetrics(b *testing.B, before, after runtime.MemStats, scratch *metricExportScratch) {
	gcCount := float64(after.NumGC - before.NumGC)
	heapDelta := float64(int64(after.HeapAlloc) - int64(before.HeapAlloc))
	totalPauses := float64(after.PauseTotalNs - before.PauseTotalNs)
	totalAllocDelta := float64(int64(after.TotalAlloc) - int64(before.TotalAlloc))
	mallocsDelta := float64(int64(after.Mallocs) - int64(before.Mallocs))
	heapObjectsDelta := float64(int64(after.HeapObjects) - int64(before.HeapObjects))

	exportAccumulateMetric(b, scratch, gcCount, "gc.count", benchreport.MetricKindCount)
	exportAccumulateMetric(b, scratch, heapDelta, "heap.delta.bytes", benchreport.MetricKindBytes)
	exportAccumulateMetric(b, scratch, totalAllocDelta, "heap.total.alloc.bytes", benchreport.MetricKindBytes)
	exportAccumulateMetric(b, scratch, float64(after.HeapInuse), "heap.inuse.bytes", benchreport.MetricKindBytes)
	exportAccumulateMetric(b, scratch, heapObjectsDelta, "heap.objects", benchreport.MetricKindCount)
	exportAccumulateMetric(b, scratch, float64(after.Sys), "sys.bytes", benchreport.MetricKindBytes)

	if b.N > 0 {
		exportAccumulateMetric(b, scratch, gcCount/float64(b.N), "gc.per.op", benchreport.MetricKindScalar)
		exportAccumulateMetric(b, scratch, mallocsDelta/float64(b.N), "mallocs.per.op", benchreport.MetricKindScalar)
		if gcCount > 0 {
			exportAccumulateMetric(b, scratch, totalPauses/gcCount, "ns/op.gc.pause.avg", benchreport.MetricKindDurationNS)
		}
		exportAccumulateMetric(b, scratch, totalPauses/float64(b.N), "ns/op.gc.pause", benchreport.MetricKindDurationNS)
	}

	if b.Elapsed().Seconds() > 0 {
		exportAccumulateMetric(b, scratch, float64(b.N)/b.Elapsed().Seconds(), "ops/sec", benchreport.MetricKindRatePerSec)
	}
}

func reportThroughputMetrics(b *testing.B, config BenchmarkMetricsConfig, scratch *metricExportScratch) {
	if b.Elapsed().Seconds() <= 0 {
		return
	}

	opsPerSec := float64(b.N) / b.Elapsed().Seconds()

	if config.FLOPSPerOp > 0 {
		exportAccumulateMetric(b, scratch, config.FLOPSPerOp*opsPerSec, "flops/sec", benchreport.MetricKindRatePerSec)
	}

	if config.BytesPerOp > 0 {
		exportAccumulateMetric(b, scratch, config.BytesPerOp, "manual.bytes/op", benchreport.MetricKindBytes)
	}

	if config.ItemsPerOp > 0 {
		exportAccumulateMetric(b, scratch, config.ItemsPerOp, "items/op", benchreport.MetricKindCount)
	}
}

func CheckMemoryPressure(initialHeap, maxHeapGrowth, maxHeapSize uint64) bool {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	currentHeap := memStats.HeapAlloc
	return (currentHeap-initialHeap) > maxHeapGrowth || currentHeap > maxHeapSize
}
