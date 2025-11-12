//go:build !amd64

package hash

func xxh3Accumulate512Dispatch() xxh3AccumulateFunc {
	return xxh3Accumulate512_scalar
}

func xxh3ScrambleAccDispatch() xxh3ScrambleFunc {
	return xxh3ScrambleAcc_scalar
}
