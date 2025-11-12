//go:build amd64

package hash

import (
	"unsafe"

	"golang.org/x/sys/cpu"
)

func xxh3Accumulate512Dispatch() xxh3AccumulateFunc {
	if !cpu.X86.HasAVX2 {
		return xxh3Accumulate512_scalar
	}

	return xxh3Accumulate512_avx2
}

func xxh3ScrambleAccDispatch() xxh3ScrambleFunc {
	if !cpu.X86.HasAVX2 {
		return xxh3ScrambleAcc_scalar
	}
	return xxh3Scramble_avx2
}

//go:inline
//go:nosplit
func xxh3Accumulate512_avx2(acc, block, secret *[8]uint64) {
	xxh3Accumulate512AVX2(
		&acc[0],
		(*byte)(unsafe.Pointer(&block[0])),
		(*byte)(unsafe.Pointer(&secret[0])),
	)
}

//go:inline
//go:nosplit
func xxh3Scramble_avx2(acc, secret *[8]uint64) {
	xxh3ScrambleAccAVX2(
		&acc[0],
		(*byte)(unsafe.Pointer(&secret[0])),
	)
}

//go:noescape
func xxh3Accumulate512AVX2(acc *uint64, in *byte, secret *byte)

//go:noescape
func xxh3ScrambleAccAVX2(acc *uint64, secret *byte)
