//go:build amd64

package foundation

import (
	"math"

	"golang.org/x/sys/cpu"
)

// Declare assembly functions
//
//go:noescape
func foundationSqrt32SSE2(x float32) float32

//go:noescape
func foundationSqrt64SSE2(x float64) float64

// sqrtDispatchTable holds function pointers for sqrt operations
// These are initialized once at package init based on CPU features
type sqrtDispatchTable struct {
	sqrt32 func(float32) float32
	sqrt64 func(float64) float64
}

var defaultSqrtDispatchTable sqrtDispatchTable

// init initializes the dispatch table based on CPU features
func init() {
	initSqrtDispatchTable(&defaultSqrtDispatchTable)
}

// initSqrtDispatchTable initializes a dispatch table based on CPU features
func initSqrtDispatchTable(table *sqrtDispatchTable) {
	if cpu.X86.HasSSE2 {
		table.sqrt32 = foundationSqrt32Go // float32 in assembly has issues for some reason
		table.sqrt64 = foundationSqrt64SSE2
	} else {
		table.sqrt32 = foundationSqrt32Go
		table.sqrt64 = foundationSqrt64Go
	}
}

// foundationSqrt32Go is the Go fallback implementation for Sqrt32
func foundationSqrt32Go(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

// foundationSqrt64Go is the Go fallback implementation for Sqrt64
func foundationSqrt64Go(x float64) float64 {
	return math.Sqrt(x)
}
