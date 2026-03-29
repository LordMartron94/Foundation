// Valgrind client request: VALGRIND_DO_CLIENT_REQUEST_EXPR preamble for amd64-linux (valgrind/include/valgrind.h).

#include "textflag.h"

// func benchmarkingCallgrindClientRequest(req, a1, a2, a3, a4, a5 uint64) uint64
TEXT ·benchmarkingCallgrindClientRequest(SB), NOSPLIT, $0-56
	LEAQ	args+0(FP), AX

	XORL	DX, DX

	ROLQ	$3, DI; ROLQ	$13, DI
	ROLQ	$61, DI; ROLQ	$51, DI

	XCHGQ	BX, BX

	MOVQ	DX, ret+48(FP)
	RET
