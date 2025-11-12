package hash

//go:inline
//go:nosplit
func xxh3Accumulate512_scalar(acc *[8]uint64, block *[8]uint64, secret *[8]uint64) {
	v0 := block[0] ^ secret[0]
	v1 := block[1] ^ secret[1]
	v2 := block[2] ^ secret[2]
	v3 := block[3] ^ secret[3]
	v4 := block[4] ^ secret[4]
	v5 := block[5] ^ secret[5]
	v6 := block[6] ^ secret[6]
	v7 := block[7] ^ secret[7]

	acc[1] += block[0]
	acc[0] += uint64(uint32(v0)) * uint64(uint32(v0>>32))
	acc[0] += block[1]
	acc[1] += uint64(uint32(v1)) * uint64(uint32(v1>>32))
	acc[3] += block[2]
	acc[2] += uint64(uint32(v2)) * uint64(uint32(v2>>32))
	acc[2] += block[3]
	acc[3] += uint64(uint32(v3)) * uint64(uint32(v3>>32))
	acc[5] += block[4]
	acc[4] += uint64(uint32(v4)) * uint64(uint32(v4>>32))
	acc[4] += block[5]
	acc[5] += uint64(uint32(v5)) * uint64(uint32(v5>>32))
	acc[7] += block[6]
	acc[6] += uint64(uint32(v6)) * uint64(uint32(v6>>32))
	acc[6] += block[7]
	acc[7] += uint64(uint32(v7)) * uint64(uint32(v7>>32))
}

//go:inline
//go:nosplit
func xxh3ScrambleAcc_scalar(accumulator *[8]uint64, sw *[8]uint64) {
	accumulator[0] = (accumulator[0] ^ (accumulator[0] >> 47)) ^ sw[0]
	accumulator[0] *= prime32_1
	accumulator[1] = (accumulator[1] ^ (accumulator[1] >> 47)) ^ sw[1]
	accumulator[1] *= prime32_1
	accumulator[2] = (accumulator[2] ^ (accumulator[2] >> 47)) ^ sw[2]
	accumulator[2] *= prime32_1
	accumulator[3] = (accumulator[3] ^ (accumulator[3] >> 47)) ^ sw[3]
	accumulator[3] *= prime32_1
	accumulator[4] = (accumulator[4] ^ (accumulator[4] >> 47)) ^ sw[4]
	accumulator[4] *= prime32_1
	accumulator[5] = (accumulator[5] ^ (accumulator[5] >> 47)) ^ sw[5]
	accumulator[5] *= prime32_1
	accumulator[6] = (accumulator[6] ^ (accumulator[6] >> 47)) ^ sw[6]
	accumulator[6] *= prime32_1
	accumulator[7] = (accumulator[7] ^ (accumulator[7] >> 47)) ^ sw[7]
	accumulator[7] *= prime32_1
}
