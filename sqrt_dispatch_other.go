//go:build !amd64

package foundation

import "math"

// sqrtDispatchTable holds function pointers for sqrt operations
// On non-amd64 platforms, all function pointers use Go fallback
type sqrtDispatchTable struct {
	sqrt32 func(float32) float32
	sqrt64 func(float64) float64
}

var defaultSqrtDispatchTable sqrtDispatchTable

// init initializes the dispatch table with Go fallback implementations
func init() {
	defaultSqrtDispatchTable.sqrt32 = foundationSqrt32Go
	defaultSqrtDispatchTable.sqrt64 = foundationSqrt64Go
}

// foundationSqrt32Go is the Go fallback implementation for Sqrt32
func foundationSqrt32Go(x float32) float32 {
	if x < 0 {
		return float32(math.NaN())
	}
	return float32(math.Sqrt(float64(x)))
}

// foundationSqrt64Go is the Go fallback implementation for Sqrt64
func foundationSqrt64Go(x float64) float64 {
	if x < 0 {
		return math.NaN()
	}
	return math.Sqrt(x)
}


