// Valgrind client request: VALGRIND_DO_CLIENT_REQUEST_EXPR preamble for arm64-linux (valgrind/include/valgrind.h).

#include "textflag.h"

// func benchmarkingCallgrindClientRequest(req, a1, a2, a3, a4, a5 uint64) uint64
TEXT ·benchmarkingCallgrindClientRequest(SB), NOSPLIT, $0-56
	MOVD	$args+0(FP), R4

	MOVD	ZR, R3

	ROR	$3, R12; ROR	$13, R12
	ROR	$51, R12; ROR	$61, R12

	ORR	R10, R10

	MOVD	R3, ret+48(FP)
	RET
