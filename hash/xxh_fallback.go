//go:build !amd64

package hash

/*
xxh3Accumulate512Dispatch returns the scalar XXH3 accumulate-512 implementation for any
architecture other than amd64.

[Context]
amd64 builds use xxh_amd64.go to dispatch to an AVX2-accelerated variant. All other architectures
fall back to the portable scalar implementation defined in xxh_generic.go. The filename
deliberately avoids the `_amd64.go` suffix so Go's implicit GOARCH filename constraint does not
combine with the explicit `!amd64` build tag to exclude this file from every build.
*/
func xxh3Accumulate512Dispatch() xxh3AccumulateFunc {
	return xxh3Accumulate512_scalar
}

/*
xxh3ScrambleAccDispatch returns the scalar XXH3 scramble accumulator for any architecture other
than amd64.
*/
func xxh3ScrambleAccDispatch() xxh3ScrambleFunc {
	return xxh3ScrambleAcc_scalar
}
