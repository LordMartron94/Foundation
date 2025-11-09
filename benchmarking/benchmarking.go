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

func BenchmarkWithMetrics[data any](
	b *testing.B,
	prepareFn func(b *testing.B) data,
	testFn func(data data, b *testing.B),
	cleanupFn func(data data, b *testing.B),
) {
	name := b.Name()
	fmt.Fprintf(os.Stderr, "🔹 Running %s...\n", name)

	preparedData := prepareFn(b)
	b.ReportAllocs()

	debug.FreeOSMemory()
	runtime.GC()

	var panicValue any
	var before, after runtime.MemStats
	b.ResetTimer()

	runtime.ReadMemStats(&before)

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicValue = r
			}
		}()
		testFn(preparedData, b)
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
		os.Exit(1)
	}

	gcCount := float64(after.NumGC - before.NumGC)
	heapDelta := float64(int64(after.HeapAlloc) - int64(before.HeapAlloc))
	totalPauses := float64(after.PauseTotalNs - before.PauseTotalNs)

	b.ReportMetric(gcCount, "gc.count")
	b.ReportMetric(gcCount/float64(b.N), "gc.per.op")
	b.ReportMetric(heapDelta, "heap.delta.bytes")
	b.ReportMetric(totalPauses/float64(b.N), "ns/op.gc.pause")
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/sec")
	b.ReportMetric(float64(after.Sys), "sys.bytes")

	fmt.Fprintf(os.Stderr, "✅ Finished %s\n", name)
}

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
