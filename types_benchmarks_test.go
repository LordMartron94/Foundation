package foundation

import (
	"fmt"
	"foundation/benchmarking"
	"math/rand"
	"testing"
)

// BenchmarkTypesSuite benchmarks the three Sqrt variants (Sqrt, Sqrt32, Sqrt64)
// across multiple numeric types and value magnitudes. Results are grouped by
// type and method for easy comparative analysis (fastest vs slowest).
func BenchmarkTypesSuite(b *testing.B) {
	rnd := rand.New(rand.NewSource(42))
	sizes := []float64{1, 10, 100, 1_000, 100_000, 1_000_000}

	// helper to generate random slices
	makeFloats32 := func(n int, scale float64) []float32 {
		arr := make([]float32, n)
		for i := range arr {
			arr[i] = float32(rnd.Float64() * scale)
		}
		return arr
	}
	makeFloats64 := func(n int, scale float64) []float64 {
		arr := make([]float64, n)
		for i := range arr {
			arr[i] = rnd.Float64() * scale
		}
		return arr
	}
	makeInts := func(n int, scale float64) []int64 {
		arr := make([]int64, n)
		for i := range arr {
			arr[i] = int64(rnd.Float64() * scale)
		}
		return arr
	}
	makeUints := func(n int, scale float64) []uint64 {
		arr := make([]uint64, n)
		for i := range arr {
			arr[i] = uint64(rnd.Float64() * scale)
		}
		return arr
	}

	// -------------------------------------------------------------
	// FLOAT32
	// -------------------------------------------------------------
	b.Run("Float32", func(b *testing.B) {
		for _, scale := range sizes {
			name := fmt.Sprintf("scale=%.0f", scale)
			b.Run(name+"/Sqrt", func(b *testing.B) {
				runFloat32Bench(b, makeFloats32(10_000, scale), Sqrt[float32])
			})
			b.Run(name+"/Sqrt32", func(b *testing.B) {
				runFloat32Bench(b, makeFloats32(10_000, scale), func(x float32) float32 { return Sqrt32(x) })
			})
			b.Run(name+"/Sqrt64", func(b *testing.B) {
				runFloat32Bench(b, makeFloats32(10_000, scale), func(x float32) float32 { return float32(Sqrt64(x)) })
			})
		}
	})

	// -------------------------------------------------------------
	// FLOAT64
	// -------------------------------------------------------------
	b.Run("Float64", func(b *testing.B) {
		for _, scale := range sizes {
			name := fmt.Sprintf("scale=%.0f", scale)
			b.Run(name+"/Sqrt", func(b *testing.B) {
				runFloat64Bench(b, makeFloats64(10_000, scale), Sqrt[float64])
			})
			b.Run(name+"/Sqrt32", func(b *testing.B) {
				runFloat64Bench(b, makeFloats64(10_000, scale), func(x float64) float64 { return float64(Sqrt32(x)) })
			})
			b.Run(name+"/Sqrt64", func(b *testing.B) {
				runFloat64Bench(b, makeFloats64(10_000, scale), Sqrt64[float64])
			})
		}
	})

	// -------------------------------------------------------------
	// SIGNED INTS
	// -------------------------------------------------------------
	b.Run("Int64", func(b *testing.B) {
		for _, scale := range sizes {
			name := fmt.Sprintf("scale=%.0f", scale)
			b.Run(name+"/Sqrt", func(b *testing.B) {
				runIntBench(b, makeInts(10_000, scale), Sqrt[int64])
			})
			b.Run(name+"/Sqrt32", func(b *testing.B) {
				runIntBench(b, makeInts(10_000, scale), func(x int64) int64 { return int64(Sqrt32(x)) })
			})
			b.Run(name+"/Sqrt64", func(b *testing.B) {
				runIntBench(b, makeInts(10_000, scale), func(x int64) int64 { return int64(Sqrt64(x)) })
			})
		}
	})

	// -------------------------------------------------------------
	// UNSIGNED INTS
	// -------------------------------------------------------------
	b.Run("Uint64", func(b *testing.B) {
		for _, scale := range sizes {
			name := fmt.Sprintf("scale=%.0f", scale)
			b.Run(name+"/Sqrt", func(b *testing.B) {
				runUintBench(b, makeUints(10_000, scale), Sqrt[uint64])
			})
			b.Run(name+"/Sqrt32", func(b *testing.B) {
				runUintBench(b, makeUints(10_000, scale), func(x uint64) uint64 { return uint64(Sqrt32(x)) })
			})
			b.Run(name+"/Sqrt64", func(b *testing.B) {
				runUintBench(b, makeUints(10_000, scale), func(x uint64) uint64 { return uint64(Sqrt64(x)) })
			})
		}
	})
}

// ────────────────────────────────────────────────────────────────
// INTERNAL BENCH HELPERS
// ────────────────────────────────────────────────────────────────

// runFloat32Bench benchmarks any sqrt-like function operating on float32.
func runFloat32Bench(b *testing.B, values []float32, fn func(float32) float32) {
	type benchData struct {
		values []float32
	}
	benchmarking.BenchmarkWithMetrics(b,
		func(b *testing.B) benchData {
			return benchData{values}
		},
		func(d benchData, b *testing.B) {
			idx := 0
			for i := 0; i < b.N; i++ {
				_ = fn(d.values[idx])
				idx++
				if idx >= len(d.values) {
					idx = 0
				}
			}
		},
		nil,
	)
}

// runFloat64Bench benchmarks any sqrt-like function operating on float64.
func runFloat64Bench(b *testing.B, values []float64, fn func(float64) float64) {
	type benchData struct {
		values []float64
	}
	benchmarking.BenchmarkWithMetrics(b,
		func(b *testing.B) benchData {
			return benchData{values}
		},
		func(d benchData, b *testing.B) {
			idx := 0
			for i := 0; i < b.N; i++ {
				_ = fn(d.values[idx])
				idx++
				if idx >= len(d.values) {
					idx = 0
				}
			}
		},
		nil,
	)
}

// runIntBench benchmarks any sqrt-like function operating on int64.
func runIntBench(b *testing.B, values []int64, fn func(int64) int64) {
	type benchData struct {
		values []int64
	}
	benchmarking.BenchmarkWithMetrics(b,
		func(b *testing.B) benchData {
			return benchData{values}
		},
		func(d benchData, b *testing.B) {
			idx := 0
			for i := 0; i < b.N; i++ {
				_ = fn(d.values[idx])
				idx++
				if idx >= len(d.values) {
					idx = 0
				}
			}
		},
		nil,
	)
}

// runUintBench benchmarks any sqrt-like function operating on uint64.
func runUintBench(b *testing.B, values []uint64, fn func(uint64) uint64) {
	type benchData struct {
		values []uint64
	}
	benchmarking.BenchmarkWithMetrics(b,
		func(b *testing.B) benchData {
			return benchData{values}
		},
		func(d benchData, b *testing.B) {
			idx := 0
			for i := 0; i < b.N; i++ {
				_ = fn(d.values[idx])
				idx++
				if idx >= len(d.values) {
					idx = 0
				}
			}
		},
		nil,
	)
}
