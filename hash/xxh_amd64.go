//go:build amd64

package hash

import (
	"unsafe"

	"golang.org/x/sys/cpu"
)

func xxh3Accumulate512Dispatch(accumulator *[8]uint64, stripe *[8]uint64, secret *[8]uint64) {
	if !cpu.X86.HasAVX2 {
		xxh3Accumulate512_scalar(accumulator, stripe, secret)
		return
	}
	xxh3Accumulate512AVX2(
		&accumulator[0],
		(*byte)(unsafe.Pointer(&stripe[0])),
		(*byte)(unsafe.Pointer(&secret[0])),
	)
}

func xxh3ScrambleAccDispatch(acc *[8]uint64, secret *[8]uint64) {
	if !cpu.X86.HasAVX2 {
		xxh3ScrambleAcc_scalar(acc, secret)
		return
	}
	xxh3ScrambleAccAVX2(
		&acc[0],
		(*byte)(unsafe.Pointer(&secret[0])),
	)
}

//go:noescape
func xxh3Accumulate512AVX2(acc *uint64, in *byte, secret *byte)

//go:noescape
func xxh3ScrambleAccAVX2(acc *uint64, secret *byte)
