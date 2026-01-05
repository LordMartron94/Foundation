//go:build amd64

#include "textflag.h"

// // foundationSqrt32SSE2 computes square root of float32 using SSE2
// // func foundationSqrt32SSE2(x float32) float32
TEXT ·foundationSqrt32SSE2(SB), NOSPLIT, $0-8
	MOVSS x+0(FP), X0        // Load float32 into X0
	SQRTSS X0, X0            // Compute sqrt: X0 = sqrt(X0)
	MOVSS X0, ret+4(FP)      // Store result
	RET

// foundationSqrt64SSE2 computes square root of float64 using SSE2
// func foundationSqrt64SSE2(x float64) float64
TEXT ·foundationSqrt64SSE2(SB), NOSPLIT, $0-16
	MOVSD x+0(FP), X0        // Load float64 into X0
	SQRTSD X0, X0            // Compute sqrt: X0 = sqrt(X0)
	MOVSD X0, ret+8(FP)      // Store result
	RET

