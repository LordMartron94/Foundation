//go:build linux && (amd64 || arm64)

package benchmarking

import "os"

/*
Callgrind request codes match valgrind/callgrind.h Vg_CallgrindClientRequest (ABI-stable).
*/
const (
	callgrindUserreqToolBase uint64 = 0x43540000 /* VG_USERREQ_TOOL_BASE('C','T') */

	callgrindReqStartInstrumentation = callgrindUserreqToolBase + 4 /* VG_USERREQ__START_INSTRUMENTATION */
	callgrindReqStopInstrumentation  = callgrindUserreqToolBase + 5 /* VG_USERREQ__STOP_INSTRUMENTATION */
)

func benchmarkingCallgrindClientRequest(req, a1, a2, a3, a4, a5 uint64) uint64

/*
BenchmarkingCallgrindInstrRegionMaybeBegin emits a Callgrind START_INSTRUMENTATION client request when
EnvAnvilCallgrindInstrRegion is "1". No-op otherwise. Pair with BenchmarkingCallgrindInstrRegionMaybeEnd
and run under Valgrind Callgrind with --instr-atstart=no for hot-path-only profiling.
*/
func BenchmarkingCallgrindInstrRegionMaybeBegin() {
	if os.Getenv(EnvAnvilCallgrindInstrRegion) != "1" {
		return
	}
	_ = benchmarkingCallgrindClientRequest(callgrindReqStartInstrumentation, 0, 0, 0, 0, 0)
}

/*
BenchmarkingCallgrindInstrRegionMaybeEnd emits a Callgrind STOP_INSTRUMENTATION client request when
EnvAnvilCallgrindInstrRegion is "1". No-op otherwise.
*/
func BenchmarkingCallgrindInstrRegionMaybeEnd() {
	if os.Getenv(EnvAnvilCallgrindInstrRegion) != "1" {
		return
	}
	_ = benchmarkingCallgrindClientRequest(callgrindReqStopInstrumentation, 0, 0, 0, 0, 0)
}
