package benchmarking

import (
	"fmt"
	"memcore"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
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

Fields:
- FLOPSPerOp: Floating point operations per benchmark operation.
- BytesPerOp: Manual memory bytes accessed per operation (for bandwidth calculation).
- ItemsPerOp: Logical items processed per operation.
*/
type BenchmarkMetricsConfig struct {
	FLOPSPerOp float64
	BytesPerOp float64
	ItemsPerOp float64
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
*/
func BenchmarkWithMetricsConfig[data any](
	b *testing.B,
	config BenchmarkMetricsConfig,
	prepareFn func(b *testing.B) data,
	benchmarkFn func(data data, b *testing.B),
	cleanupFn func(data data, b *testing.B),
) {
	name := b.Name()
	fmt.Fprintf(os.Stderr, "🔹 Running %s...\n", name)

	// 1. Preparation
	preparedData := prepareFn(b)
	b.ReportAllocs()

	debug.FreeOSMemory()
	runtime.GC()

	var panicValue any
	var panicStack []byte

	var before, after runtime.MemStats

	// 2. Execution
	b.ResetTimer()
	runtime.ReadMemStats(&before)

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicValue = r
				panicStack = debug.Stack()
			}
		}()
		benchmarkFn(preparedData, b)
	}()

	b.StopTimer()
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
	reportStandardMetrics(b, before, after)
	reportThroughputMetrics(b, config)

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
	BenchmarkWithMetricsConfig(b, BenchmarkMetricsConfig{}, prepareFn, benchmarkFn, cleanupFn)
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

func reportStandardMetrics(b *testing.B, before, after runtime.MemStats) {
	gcCount := float64(after.NumGC - before.NumGC)
	heapDelta := float64(int64(after.HeapAlloc) - int64(before.HeapAlloc))
	totalPauses := float64(after.PauseTotalNs - before.PauseTotalNs)
	totalAllocDelta := float64(int64(after.TotalAlloc) - int64(before.TotalAlloc))
	mallocsDelta := float64(int64(after.Mallocs) - int64(before.Mallocs))
	heapObjectsDelta := float64(int64(after.HeapObjects) - int64(before.HeapObjects))

	b.ReportMetric(gcCount, "gc.count")
	b.ReportMetric(heapDelta, "heap.delta.bytes")
	b.ReportMetric(totalAllocDelta, "heap.total.alloc.bytes")
	b.ReportMetric(float64(after.HeapInuse), "heap.inuse.bytes")
	b.ReportMetric(heapObjectsDelta, "heap.objects")
	b.ReportMetric(float64(after.Sys), "sys.bytes")

	if b.N > 0 {
		b.ReportMetric(gcCount/float64(b.N), "gc.per.op")
		b.ReportMetric(mallocsDelta/float64(b.N), "mallocs.per.op")
		if gcCount > 0 {
			b.ReportMetric(totalPauses/gcCount, "ns/op.gc.pause.avg")
		}
		b.ReportMetric(totalPauses/float64(b.N), "ns/op.gc.pause")
	}

	if b.Elapsed().Seconds() > 0 {
		b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
	}
}

func reportThroughputMetrics(b *testing.B, config BenchmarkMetricsConfig) {
	if b.Elapsed().Seconds() <= 0 {
		return
	}

	opsPerSec := float64(b.N) / b.Elapsed().Seconds()

	if config.FLOPSPerOp > 0 {
		b.ReportMetric(config.FLOPSPerOp*opsPerSec, "flops/sec")
	}

	if config.BytesPerOp > 0 {
		b.ReportMetric(config.BytesPerOp, "manual.bytes/op")
	}

	if config.ItemsPerOp > 0 {
		b.ReportMetric(config.ItemsPerOp, "items/op")
	}
}

func CheckMemoryPressure(initialHeap, maxHeapGrowth, maxHeapSize uint64) bool {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	currentHeap := memStats.HeapAlloc
	return (currentHeap-initialHeap) > maxHeapGrowth || currentHeap > maxHeapSize
}
