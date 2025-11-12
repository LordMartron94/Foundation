//go:build !amd64

package hash

func xxh3Accumulate512Dispatch(accumulator *[8]uint64, stripe *[8]uint64, secret *[8]uint64) {
	xxh3Accumulate512_scalar(accumulator, stripe, secret)
}

func xxh3ScrambleAccDispatch(acc *[8]uint64, secret *[8]uint64) {
	xxh3ScrambleAcc_scalar(acc, secret)
}
