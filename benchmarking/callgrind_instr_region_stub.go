//go:build !linux || !(amd64 || arm64)

package benchmarking

/*
BenchmarkingCallgrindInstrRegionMaybeBegin is a no-op when Valgrind client-request assembly is unavailable
(GOOS/GOARCH other than linux on amd64 or arm64).
*/
func BenchmarkingCallgrindInstrRegionMaybeBegin() {}

/*
BenchmarkingCallgrindInstrRegionMaybeEnd is a no-op when Valgrind client-request assembly is unavailable
(GOOS/GOARCH other than linux on amd64 or arm64).
*/
func BenchmarkingCallgrindInstrRegionMaybeEnd() {}
