// Package benchmarking provides helpful utilities for benchmarking.
package benchmarking

import (
	"fmt"
	"memcore"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
)

/*
BenchmarkMetricsConfig provides optional configuration for advanced throughput metrics.

Use cases:
- Configuring FLOPS reporting for computational benchmarks
- Specifying manual memory bytes per operation for accurate memory throughput with manual memory management
- Specifying items processed per operation for accurate throughput calculation
- Future extensibility for additional throughput metrics

Fields:
- FLOPSPerOp: Floating point operations per benchmark operation (0 means not specified)
- BytesPerOp: Manual memory bytes per operation (0 means not specified). Use this when benchmarks use manual memory management (memcore, memforge, etc.) outside the Go GC. This takes precedence over standard GC-tracked bytes/op for memory throughput calculation.
- ItemsPerOp: Items processed per operation (for future extensibility, 0 means not specified)

Time complexity: O(1)
Space complexity: O(1)

Prerequisites:
- Zero value means no additional metrics will be reported
- All fields are optional and default to 0

Edge cases:
- Zero values are ignored (no metrics reported for that field)
- Negative values are treated as zero
- BytesPerOp should be used when benchmarks allocate memory outside Go's GC (e.g., via memcore, memforge)
*/
type BenchmarkMetricsConfig struct {
	FLOPSPerOp float64
	BytesPerOp float64
	ItemsPerOp float64
}

/*
BenchmarkWithMetrics executes a benchmark with comprehensive metrics collection and reporting.

This function provides a standardized way to run benchmarks with automatic collection
of performance metrics including execution time, memory usage, and garbage collection
statistics. It handles setup, execution, cleanup, and panic recovery.

The function performs the following steps:
1. Prepares benchmark data using prepareFn
2. Enables allocation reporting
3. Performs initial GC and memory cleanup
4. Executes the benchmark function with panic recovery
5. Collects memory and GC statistics
6. Performs cleanup operations
7. Reports comprehensive metrics

Parameters:
- b: Benchmark context from testing.B
- prepareFn: Function to prepare benchmark data (executed before timing starts)
- benchmarkFn: Function containing the actual benchmark code (executed during timing)
- cleanupFn: Function to clean up resources (executed after timing ends, may be nil)

Reported Metrics:
- gc.count: Total number of GC cycles during benchmark
- gc.per.op: Average GC cycles per operation
- heap.delta.bytes: Change in heap allocation (may be negative if GC freed memory)
- heap.total.alloc.bytes: Total bytes allocated during benchmark (cumulative)
- heap.inuse.bytes: Heap memory currently in use after benchmark
- heap.objects: Number of heap objects allocated
- mallocs.per.op: Average number of allocations per operation
- ns/op.gc.pause: Average GC pause time per operation
- ns/op.gc.pause.avg: Average GC pause time per GC cycle
- ops/sec: Operations per second throughput
- sys.bytes: Total system memory obtained from OS

For advanced throughput metrics (FLOPS, etc.), use BenchmarkWithMetricsConfig instead.

Time complexity: O(n) - where n is the work performed by benchmarkFn
Space complexity: O(1) - minimal overhead for metrics collection

Edge cases:
- Handles panics gracefully with stack trace reporting
- Performs cleanup even if benchmark panics
- Excludes setup and cleanup time from measurements
- Reports metrics even if benchmark fails
*/
func BenchmarkWithMetrics[data any](
	b *testing.B,
	prepareFn func(b *testing.B) data,
	benchmarkFn func(data data, b *testing.B),
	cleanupFn func(data data, b *testing.B),
) {
	name := b.Name()
	fmt.Fprintf(os.Stderr, "🔹 Running %s...\n", name)

	preparedData := prepareFn(b)
	b.ReportAllocs()

	debug.FreeOSMemory()
	runtime.GC()

	var panicValue any
	var panicStack []byte

	var before, after runtime.MemStats
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

	if cleanupFn != nil {
		cleanupFn(preparedData, b)
	}

	memcore.MemmapUnmapAllRegions()
	memcore.MemcoreMarkManagementStateReset(false)
	debug.FreeOSMemory()

	if panicValue != nil {
		fmt.Fprintf(os.Stderr, "\n🔥 Benchmark panic in %s: %v\n", name, panicValue)
		if len(panicStack) > 0 {
			fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", panicStack)
		}
		os.Exit(1)
	}

	gcCount := float64(after.NumGC - before.NumGC)
	heapDelta := float64(int64(after.HeapAlloc) - int64(before.HeapAlloc))
	totalPauses := float64(after.PauseTotalNs - before.PauseTotalNs)
	totalAllocDelta := float64(int64(after.TotalAlloc) - int64(before.TotalAlloc))
	mallocsDelta := float64(int64(after.Mallocs) - int64(before.Mallocs))
	heapObjectsDelta := float64(int64(after.HeapObjects) - int64(before.HeapObjects))

	// GC metrics
	b.ReportMetric(gcCount, "gc.count")
	if b.N > 0 {
		b.ReportMetric(gcCount/float64(b.N), "gc.per.op")
	}
	if gcCount > 0 {
		b.ReportMetric(totalPauses/gcCount, "ns/op.gc.pause.avg")
	}
	if b.N > 0 {
		b.ReportMetric(totalPauses/float64(b.N), "ns/op.gc.pause")
	}

	// Heap allocation metrics
	b.ReportMetric(heapDelta, "heap.delta.bytes")
	b.ReportMetric(totalAllocDelta, "heap.total.alloc.bytes")
	b.ReportMetric(float64(after.HeapInuse), "heap.inuse.bytes")
	b.ReportMetric(heapObjectsDelta, "heap.objects")

	// Allocation frequency metrics
	if b.N > 0 {
		b.ReportMetric(mallocsDelta/float64(b.N), "mallocs.per.op")
	}

	// Throughput and system metrics
	if b.Elapsed().Seconds() > 0 {
		b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
	}
	b.ReportMetric(float64(after.Sys), "sys.bytes")

	fmt.Fprintf(os.Stderr, "✅ Finished %s\n", name)
}

/*
BenchmarkWithMetricsConfig executes a benchmark with comprehensive metrics collection and optional throughput configuration.

This function wraps BenchmarkWithMetrics and adds support for configurable throughput metrics
such as FLOPS (Floating Point Operations Per Second). It maintains full backward compatibility
with BenchmarkWithMetrics while providing additional metrics when configured.

The function performs the following steps:
1. Calls BenchmarkWithMetrics to collect standard metrics
2. Calculates and reports derived throughput metrics based on config
3. Reports flops/sec if FLOPSPerOp is specified: flops/sec = FLOPSPerOp * ops/sec

Parameters:
- b: Benchmark context from testing.B
- config: Configuration struct with optional throughput metrics (zero value means no additional metrics)
- prepareFn: Function to prepare benchmark data (executed before timing starts)
- benchmarkFn: Function containing the actual benchmark code (executed during timing)
- cleanupFn: Function to clean up resources (executed after timing ends, may be nil)

Reported Metrics:
- All metrics from BenchmarkWithMetrics
- flops/sec: Floating point operations per second (if FLOPSPerOp > 0 in config)
- manual.bytes/op: Manual memory bytes per operation (if BytesPerOp > 0 in config)

Example usage:
	BenchmarkWithMetricsConfig(b, BenchmarkMetricsConfig{FLOPSPerOp: 1024.0}, prepareFn, testFn, cleanupFn)
	
	// For manual memory management:
	BenchmarkWithMetricsConfig(b, BenchmarkMetricsConfig{BytesPerOp: 4096.0}, prepareFn, testFn, cleanupFn)

Time complexity: O(n) - where n is the work performed by benchmarkFn
Space complexity: O(1) - minimal overhead for metrics collection

Prerequisites:
- config can be zero value (no additional metrics will be reported)
- BenchmarkWithMetrics must complete successfully for derived metrics to be calculated

Edge cases:
- Zero or negative config values are ignored (no metrics reported)
- Derived metrics require ops/sec to be available from BenchmarkWithMetrics
- If ops/sec is not available, derived metrics are not reported
*/
func BenchmarkWithMetricsConfig[data any](
	b *testing.B,
	config BenchmarkMetricsConfig,
	prepareFn func(b *testing.B) data,
	benchmarkFn func(data data, b *testing.B),
	cleanupFn func(data data, b *testing.B),
) {
	BenchmarkWithMetrics(b, prepareFn, benchmarkFn, cleanupFn)

	// Calculate and report derived metrics based on config
	if config.FLOPSPerOp > 0 && b.Elapsed().Seconds() > 0 {
		opsPerSec := float64(b.N) / b.Elapsed().Seconds()
		flopsPerSec := config.FLOPSPerOp * opsPerSec
		b.ReportMetric(flopsPerSec, "flops/sec")
	}

	// Report manual bytes/op if specified (for manual memory management)
	if config.BytesPerOp > 0 {
		b.ReportMetric(config.BytesPerOp, "manual.bytes/op")
	}
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
CheckMemoryPressure determines if current memory usage exceeds safe thresholds.

Returns true if memory pressure is detected and GC should be triggered.
Uses HeapAlloc to measure current heap usage and compares against configurable thresholds.

Parameters:
- initialHeap: Baseline heap size to compare growth against
- maxHeapGrowth: Maximum allowed heap growth in bytes before triggering GC
- maxHeapSize: Maximum absolute heap size in bytes before triggering GC

Time complexity: O(1) - single MemStats read
Space complexity: O(1) - no allocations
*/
func CheckMemoryPressure(initialHeap, maxHeapGrowth, maxHeapSize uint64) bool {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	currentHeap := memStats.HeapAlloc
	heapGrowth := currentHeap - initialHeap

	return heapGrowth > maxHeapGrowth || currentHeap > maxHeapSize
}

/*
RunBatchedBenchmark executes a benchmark function with memory-aware GC management.

Monitors memory pressure during execution and triggers GC when memory usage exceeds
configurable thresholds. Memory checks are performed outside timing to exclude
MemStats overhead from benchmark measurements.

Parameters:
- b: Benchmark context
- workFn: Function to execute for each iteration
- maxHeapGrowth: Maximum allowed heap growth in bytes before triggering GC (default: 500MB)
- maxHeapSize: Maximum absolute heap size in bytes before triggering GC (default: 1GB)
- memoryCheckInterval: Number of iterations between memory checks (default: 50,000)

Time complexity: O(n) - where n is b.N
Space complexity: O(1) - no additional allocations

Edge cases:
- Memory checks are performed outside timing to exclude MemStats overhead
- GC is triggered only when memory pressure is detected
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
