package hash

import (
	"encoding/binary"
	"fmt"
	"foundation"
	"math/bits"
	"memcore"
	"unsafe"
)

// Implemented according to spec: https://www.ietf.org/archive/id/draft-josefsson-xxhash-00.html#name-xxh3-algorithm-overview

const (
	prime32_1 uint64 = 0x9E3779B1
	prime32_2 uint64 = 0x85EBCA77
	prime32_3 uint64 = 0xC2B2AE3D

	prime64_1 uint64 = 0x9E3779B185EBCA87
	prime64_2 uint64 = 0xC2B2AE3D27D4EB4F
	prime64_3 uint64 = 0x165667B19E3779F9
	prime64_4 uint64 = 0x85EBCA77C2B2AE63
	prime64_5 uint64 = 0x27D4EB2F165667C5

	prime_mx1 uint64 = 0x165667919E3779F9
	prime_mx2 uint64 = 0x9FB21C651E98DF25
)

var (
	defaultHashSecret []uint8 = []uint8{
		0xb8, 0xfe, 0x6c, 0x39, 0x23, 0xa4, 0x4b, 0xbe,
		0x7c, 0x01, 0x81, 0x2c, 0xf7, 0x21, 0xad, 0x1c,
		0xde, 0xd4, 0x6d, 0xe9, 0x83, 0x90, 0x97, 0xdb,
		0x72, 0x40, 0xa4, 0xa4, 0xb7, 0xb3, 0x67, 0x1f,
		0xcb, 0x79, 0xe6, 0x4e, 0xcc, 0xc0, 0xe5, 0x78,
		0x82, 0x5a, 0xd0, 0x7d, 0xcc, 0xff, 0x72, 0x21,
		0xb8, 0x08, 0x46, 0x74, 0xf7, 0x43, 0x24, 0x8e,
		0xe0, 0x35, 0x90, 0xe6, 0x81, 0x3a, 0x26, 0x4c,
		0x3c, 0x28, 0x52, 0xbb, 0x91, 0xc3, 0x00, 0xcb,
		0x88, 0xd0, 0x65, 0x8b, 0x1b, 0x53, 0x2e, 0xa3,
		0x71, 0x64, 0x48, 0x97, 0xa2, 0x0d, 0xf9, 0x4e,
		0x38, 0x19, 0xef, 0x46, 0xa9, 0xde, 0xac, 0xd8,
		0xa8, 0xfa, 0x76, 0x3f, 0xe3, 0x9c, 0x34, 0x3f,
		0xf9, 0xdc, 0xbb, 0xc7, 0xc7, 0x0b, 0x4f, 0x1d,
		0x8a, 0x51, 0xe0, 0x4b, 0xcd, 0xb4, 0x59, 0x31,
		0xc8, 0x9f, 0x7e, 0xc9, 0xd9, 0x78, 0x73, 0x64,
		0xea, 0xc5, 0xac, 0x83, 0x34, 0xd3, 0xeb, 0xc3,
		0xc5, 0x81, 0xa0, 0xff, 0xfa, 0x13, 0x63, 0xeb,
		0x17, 0x0d, 0xdd, 0x51, 0xb7, 0xf0, 0xda, 0x49,
		0xd3, 0x16, 0x55, 0x26, 0x29, 0xd4, 0x68, 0x9e,
		0x2b, 0x16, 0xbe, 0x58, 0x7d, 0x47, 0xa1, 0xfc,
		0x8f, 0xf8, 0xb8, 0xd1, 0x7a, 0xd0, 0x31, 0xce,
		0x45, 0xcb, 0x3a, 0x8f, 0x95, 0x16, 0x04, 0x28,
		0xaf, 0xd7, 0xfb, 0xca, 0xbb, 0x4b, 0x40, 0x7e,
	}

	defaultHashSeed uint64 = 0
)

// XXH3Hasher is an hasher using XXH3 algorithm according to spec:
// https://www.ietf.org/archive/id/draft-josefsson-xxhash-00.html#name-xxh3-algorithm-overview
type XXH3Hasher struct {
	seed          uint64
	secret        []uint8
	derivedSecret []uint8
}

// XXH3HasherCreateWithSeed creates an XXH3 hasher with the specified seed.
// As according to spec, in this instance the default secret is used.
func XXH3HasherCreateWithSeed(seed uint64) *XXH3Hasher {
	return &XXH3Hasher{
		seed:          seed,
		secret:        defaultHashSecret,
		derivedSecret: xxh3HasherDeriveSecret(seed),
	}
}

// XXH3HasherCreateWithSecret creates an XXH3 hasher with the specified secret.
// The secret must be a minimum of 136 elements (136 bytes)
// As according to spec, in this instance the default seed is used.
func XXH3HasherCreateWithSecret(secret []uint8) *XXH3Hasher {
	if len(secret) < 136 {
		panic(fmt.Errorf("secret must be 136 bytes or bigger; got: %d", len(secret)))
	}

	return &XXH3Hasher{
		seed:          defaultHashSeed,
		secret:        secret,
		derivedSecret: nil,
	}
}

func XXH3HasherHash64(hasher *XXH3Hasher, content []byte) uint64 {
	contentLength := len(content)

	secret := hasher.secret
	if contentLength > 240 && hasher.derivedSecret != nil {
		secret = hasher.derivedSecret
	}

	switch {
	case contentLength == 0:
		return xxh3Hasher64Empty(hasher, secret)
	case contentLength < 4:
		// LSB                      8        16           24                    MSB
		combined := uint32(content[contentLength-1]) | (uint32(contentLength) << 8) |
			(uint32(content[0]) << 16) | (uint32(content[contentLength>>1]) << 24)
		return xxh3Hasher64Length1To3(hasher, combined, secret)
	case contentLength < 9:
		inputFirst := littleEndianWordGet[uint32](content, 0)
		inputLast := littleEndianWordGet[uint32](content, uint64(contentLength-4))
		modifiedSeed := hasher.seed ^ (uint64(byteSwap32(uint32(hasher.seed))) << 32)
		return xxh3Hasher64Length4To8(hasher, inputFirst, inputLast, modifiedSeed, uint64(contentLength), secret)
	case contentLength < 17:
		inputFirst := littleEndianWordGet[uint64](content, 0)
		inputLast := littleEndianWordGet[uint64](content, uint64(contentLength-8))
		return xxh3Hasher64Length9To16(hasher, inputFirst, inputLast, uint64(contentLength), secret)
	case contentLength < 129:
		accumulator := uint64(contentLength) * prime64_1
		xxh3Hasher64Length17To128(hasher, &accumulator, content, uint64(contentLength), secret)
		return xxh3HasherAvalanche(accumulator)
	case contentLength < 241:
		accumulator := uint64(contentLength) * prime64_1
		xxh3Hasher64Length129To240(hasher, &accumulator, content, uint64(contentLength), secret)
		return xxh3HasherAvalanche(accumulator)
	default:
		accumulator := &[8]uint64{
			prime32_3, prime64_1, prime64_2, prime64_3,
			prime64_4, prime32_2, prime64_5, prime32_1,
		}

		xxh3HasherAccumulateAll(hasher, accumulator, content, uint64(contentLength), secret)
		return xxh3HasherFinalize(hasher, accumulator, uint64(contentLength)*prime64_1, 11, secret)
	}
}

//go:nosplit
//go:inline
func XXH3HasherHash64View[T any](hasher *XXH3Hasher, mark memcore.MarkRaw) uint64 {
	ptr, length := memcore.MemcoreView[T](mark)
	data := unsafe.Slice((*byte)(ptr), length)
	return XXH3HasherHash64(hasher, data)
}

func XXH3HasherHash128(hasher *XXH3Hasher, content []byte) (uint64, uint64) {
	contentLength := len(content)

	secret := hasher.secret
	if contentLength > 240 && hasher.derivedSecret != nil {
		secret = hasher.derivedSecret
	}

	switch {
	case contentLength == 0:
		return xxh3Hasher128Empty(hasher, secret)
	case contentLength < 4:
		// LSB                      8        16           24                    MSB
		combined := uint32(content[contentLength-1]) | (uint32(contentLength) << 8) |
			(uint32(content[0]) << 16) | (uint32(content[contentLength>>1]) << 24)
		return xxh3Hasher128Length1To3(hasher, combined, secret)
	case contentLength < 9:
		inputFirst := littleEndianWordGet[uint32](content, 0)
		inputLast := littleEndianWordGet[uint32](content, uint64(contentLength-4))
		modifiedSeed := hasher.seed ^ (uint64(byteSwap32(uint32(hasher.seed))) << 32)
		return xxh3Hasher128Length4To8(hasher, inputFirst, inputLast, modifiedSeed, uint64(contentLength), secret)
	case contentLength < 17:
		inputFirst := littleEndianWordGet[uint64](content, 0)
		inputLast := littleEndianWordGet[uint64](content, uint64(contentLength-8))
		return xxh3Hasher128Length9To16(hasher, inputFirst, inputLast, uint64(contentLength), secret)
	case contentLength < 129:
		accumulator := &[2]uint64{uint64(contentLength) * prime64_1, 0}
		xxh3Hasher128Length17To128(hasher, accumulator, content, uint64(contentLength), secret)
		low := accumulator[0] + accumulator[1]
		high := (accumulator[0] * prime64_1) + (accumulator[1] * prime64_4) + ((uint64(contentLength) - hasher.seed) * prime64_2)
		return xxh3HasherAvalanche(low), 0 - xxh3HasherAvalanche(high)
	case contentLength < 241:
		accumulator := &[2]uint64{uint64(contentLength) * prime64_1, 0}
		xxh3Hasher128Length129To240(hasher, accumulator, content, uint64(contentLength), secret)
		low := accumulator[0] + accumulator[1]
		high := (accumulator[0] * prime64_1) + (accumulator[1] * prime64_4) + ((uint64(contentLength) - hasher.seed) * prime64_2)
		return xxh3HasherAvalanche(low), 0 - xxh3HasherAvalanche(high)
	default:
		accumulator := &[8]uint64{
			prime32_3, prime64_1, prime64_2, prime64_3,
			prime64_4, prime32_2, prime64_5, prime32_1,
		}
		secretLength := uint64(len(secret))

		xxh3HasherAccumulateAll(hasher, accumulator, content, uint64(contentLength), secret)
		return xxh3HasherFinalize(hasher, accumulator, uint64(contentLength)*prime64_1, 11, secret), xxh3HasherFinalize(hasher, accumulator, ^(uint64(contentLength) * prime64_2), secretLength-75, secret)
	}
}

//go:nosplit
//go:inline
func XXH3HasherHash128View[T any](hasher *XXH3Hasher, mark memcore.MarkRaw) (uint64, uint64) {
	ptr, length := memcore.MemcoreView[T](mark)
	data := unsafe.Slice((*byte)(ptr), length)
	return XXH3HasherHash128(hasher, data)
}

// ------------------------------------------------------- PRIVATE HELPERS

//go:inline
//go:nosplit
func xxh3Hasher64Empty(hasher *XXH3Hasher, secret []byte) uint64 {
	words := *(*[2]uint64)(unsafe.Pointer(&secret[56]))

	return xxh3HasherAvalancheXXH64(
		hasher.seed ^
			words[0] ^
			words[1],
	)
}

//go:inline
//go:nosplit
func xxh3Hasher128Empty(hasher *XXH3Hasher, secret []byte) (lo, hi uint64) {
	seed := hasher.seed
	words := *(*[4]uint64)(unsafe.Pointer(&secret[64]))

	lo = xxh3HasherAvalancheXXH64(seed ^
		words[0] ^
		words[1])
	hi = xxh3HasherAvalancheXXH64(seed ^
		words[2] ^
		words[3])
	return
}

//go:inline
//go:nosplit
func xxh3Hasher64Length1To3(hasher *XXH3Hasher, combined uint32, secret []byte) uint64 {
	seed := hasher.seed
	words := *(*[2]uint32)(unsafe.Pointer(&secret[0]))

	return xxh3HasherAvalancheXXH64(
		(uint64(words[0]^words[1]) + seed) ^ uint64(combined),
	)
}

//go:inline
//go:nosplit
func xxh3Hasher128Length1To3(hasher *XXH3Hasher, combined uint32, secret []byte) (uint64, uint64) {
	seed := hasher.seed
	words := *(*[4]uint32)(unsafe.Pointer(&secret[0]))

	low := (uint64(words[0]^words[1]) + seed) ^ uint64(combined)
	high := (uint64(words[2]^words[3]) - seed) ^ uint64(rotateBits32(byteSwap32(combined), 13))

	return xxh3HasherAvalancheXXH64(low), xxh3HasherAvalancheXXH64(high)
}

//go:inline
//go:nosplit
func xxh3Hasher64Length4To8(hasher *XXH3Hasher, inputFirst, inputLast uint32, modifiedSeed, inputLength uint64, secret []byte) uint64 {
	secretWords := *(*[2]uint64)(unsafe.Pointer(&secret[8]))

	combined := uint64(inputLast) | (uint64(inputFirst) << 32)
	value := ((secretWords[0] ^ secretWords[1]) - modifiedSeed) ^ combined
	value = value ^ rotateBits64(value, 49) ^ rotateBits64(value, 24)
	value = value * prime_mx2
	value = value ^ ((value >> 35) + inputLength)
	value = value * prime_mx2
	value = value ^ (value >> 28)
	return value
}

//go:inline
//go:nosplit
func xxh3Hasher128Length4To8(hasher *XXH3Hasher, inputFirst, inputLast uint32, modifiedSeed, inputLength uint64, secret []byte) (uint64, uint64) {
	secretWords := *(*[2]uint64)(unsafe.Pointer(&secret[16]))

	combined := uint64(inputFirst) | (uint64(inputLast) << 32)
	value := ((secretWords[0] ^ secretWords[1]) + modifiedSeed) ^ combined

	mulResult := foundation.Uint128{Lo: value, Hi: 0}.Mul(foundation.Uint128{Lo: prime64_1 + (inputLength << 2), Hi: 0})

	hi := mulResult.Hi
	lo := mulResult.Lo

	hi = hi + (lo << 1)
	lo = lo ^ (hi >> 3)
	lo = lo ^ (lo >> 35)
	lo = lo * prime_mx2
	lo = lo ^ (lo >> 28)
	hi = xxh3HasherAvalanche(hi)

	return lo, hi
}

//go:inline
//go:nosplit
func xxh3Hasher64Length9To16(hasher *XXH3Hasher, inputFirst, inputLast, inputLength uint64, secret []byte) uint64 {
	secretWords := *(*[4]uint64)(unsafe.Pointer(&secret[24]))
	low := ((secretWords[0] ^ secretWords[1]) + hasher.seed) ^ inputFirst
	high := ((secretWords[2] ^ secretWords[3]) - hasher.seed) ^ inputLast
	hi, lo := bits.Mul64(low, high)
	value := inputLength + byteSwap64(low) + high + (lo ^ hi)
	return xxh3HasherAvalanche(value)
}

//go:inline
//go:nosplit
func xxh3Hasher128Length9To16(hasher *XXH3Hasher, inputFirst, inputLast, inputLength uint64, secret []byte) (uint64, uint64) {
	secretWords := *(*[4]uint64)(unsafe.Pointer(&secret[32]))
	val1 := ((secretWords[0] ^ secretWords[1]) - hasher.seed) ^ inputFirst ^ inputLast
	val2 := ((secretWords[2] ^ secretWords[3]) + hasher.seed) ^ inputLast
	hi, lo := bits.Mul64(val1, prime64_1)
	low := lo + (uint64(inputLength-1) << 54)
	high := hi + (uint64(uint32(val2>>32)) << 32) + uint64(uint32(val2))*prime32_2
	low = low ^ byteSwap64(high)
	hi2, lo2 := bits.Mul64(low, prime64_2)
	low = lo2
	high = hi2 + high*prime64_2
	return xxh3HasherAvalanche(low), xxh3HasherAvalanche(high)
}

//go:nosplit
//go:inline
func xxh3Hasher64Length17To128(hasher *XXH3Hasher, accumulator *uint64, input []byte, inputLength uint64, secret []byte) {
	numRounds := (inputLength-1)>>5 + 1
	for i := int64(numRounds) - 1; i >= 0; i-- {
		idx := uint64(i)
		offsetStart := idx * 16
		offsetEnd := inputLength - idx*16 - 16

		block1 := *(*[16]byte)(unsafe.Pointer(&input[offsetStart]))
		block2 := *(*[16]byte)(unsafe.Pointer(&input[offsetEnd]))

		{
			sw := *(*[2]uint64)(unsafe.Pointer(&secret[idx*32]))
			dw := *(*[2]uint64)(unsafe.Pointer(&block1[0]))
			hi, lo := bits.Mul64(dw[0]^(sw[0]+hasher.seed), dw[1]^(sw[1]-hasher.seed))
			*accumulator += lo ^ hi
		}
		{
			sw := *(*[2]uint64)(unsafe.Pointer(&secret[idx*32+16]))
			dw := *(*[2]uint64)(unsafe.Pointer(&block2[0]))
			hi, lo := bits.Mul64(dw[0]^(sw[0]+hasher.seed), dw[1]^(sw[1]-hasher.seed))
			*accumulator += lo ^ hi
		}
	}
}

//go:nosplit
//go:inline
func xxh3Hasher128Length17To128(hasher *XXH3Hasher, accumulator *[2]uint64, input []byte, inputLength uint64, secret []byte) {
	numRounds := (inputLength-1)>>5 + 1
	for i := int64(numRounds) - 1; i >= 0; i-- {
		idx := uint64(i)
		offsetStart := idx * 16
		offsetEnd := inputLength - idx*16 - 16

		block1 := *(*[16]byte)(unsafe.Pointer(&input[offsetStart]))
		block2 := *(*[16]byte)(unsafe.Pointer(&input[offsetEnd]))

		{
			sw1 := *(*[2]uint64)(unsafe.Pointer(&secret[idx*32]))
			sw2 := *(*[2]uint64)(unsafe.Pointer(&secret[idx*32+16]))
			dw1 := *(*[2]uint64)(unsafe.Pointer(&block1[0]))
			dw2 := *(*[2]uint64)(unsafe.Pointer(&block2[0]))

			// Mix 1
			hi, lo := bits.Mul64(dw1[0]^(sw1[0]+hasher.seed), dw1[1]^(sw1[1]-hasher.seed))
			accumulator[0] += lo ^ hi

			// Mix 2
			hi2, lo2 := bits.Mul64(dw2[0]^(sw2[0]+hasher.seed), dw2[1]^(sw2[1]-hasher.seed))
			accumulator[1] += lo2 ^ hi2

			// Cross-lane XOR (same as MixTwoChunks did)
			accumulator[0] ^= dw2[0] + dw2[1]
			accumulator[1] ^= dw1[0] + dw1[1]
		}
	}
}

//go:inline
//go:nosplit
func xxh3Hasher64Length129To240(
	hasher *XXH3Hasher,
	accumulator *uint64,
	input []byte,
	inputLength uint64,
	secret []byte,
) {
	numChunks := inputLength >> 4

	for i := uint64(0); i < 8; i++ {
		block := *(*[16]byte)(unsafe.Pointer(&input[i*16]))
		sw := *(*[2]uint64)(unsafe.Pointer(&secret[i*16]))
		dw := *(*[2]uint64)(unsafe.Pointer(&block[0]))

		hi, lo := bits.Mul64(dw[0]^(sw[0]+hasher.seed), dw[1]^(sw[1]-hasher.seed))
		*accumulator += lo ^ hi
	}

	*accumulator = xxh3HasherAvalanche(*accumulator)

	for i := uint64(8); i < numChunks; i++ {
		block := *(*[16]byte)(unsafe.Pointer(&input[i*16]))
		sw := *(*[2]uint64)(unsafe.Pointer(&secret[(i-8)*16+3]))
		dw := *(*[2]uint64)(unsafe.Pointer(&block[0]))
		hi, lo := bits.Mul64(dw[0]^(sw[0]+hasher.seed), dw[1]^(sw[1]-hasher.seed))
		*accumulator += lo ^ hi
	}

	block := *(*[16]byte)(unsafe.Pointer(&input[inputLength-16]))
	sw := *(*[2]uint64)(unsafe.Pointer(&secret[119]))
	dw := *(*[2]uint64)(unsafe.Pointer(&block[0]))
	hi, lo := bits.Mul64(dw[0]^(sw[0]+hasher.seed), dw[1]^(sw[1]-hasher.seed))
	*accumulator += lo ^ hi
}

//go:inline
//go:nosplit
func xxh3Hasher128Length129To240(
	hasher *XXH3Hasher,
	accumulator *[2]uint64,
	input []byte,
	inputLength uint64,
	secret []byte,
) {
	numChunks := inputLength >> 5 // 32 bytes per chunk pair

	// ─── First 4 × 32-byte pairs ───
	for i := uint64(0); i < 4; i++ {
		block1 := *(*[16]byte)(unsafe.Pointer(&input[i*32]))
		block2 := *(*[16]byte)(unsafe.Pointer(&input[i*32+16]))
		sw1 := *(*[2]uint64)(unsafe.Pointer(&secret[i*32]))
		sw2 := *(*[2]uint64)(unsafe.Pointer(&secret[i*32+16]))

		dw1 := *(*[2]uint64)(unsafe.Pointer(&block1[0]))
		dw2 := *(*[2]uint64)(unsafe.Pointer(&block2[0]))

		// Mix 1
		hi, lo := bits.Mul64(dw1[0]^(sw1[0]+hasher.seed), dw1[1]^(sw1[1]-hasher.seed))
		accumulator[0] += lo ^ hi

		// Mix 2
		hi2, lo2 := bits.Mul64(dw2[0]^(sw2[0]+hasher.seed), dw2[1]^(sw2[1]-hasher.seed))
		accumulator[1] += lo2 ^ hi2

		accumulator[0] ^= dw2[0] + dw2[1]
		accumulator[1] ^= dw1[0] + dw1[1]
	}

	// ─── Mid avalanche ───
	accumulator[0] = xxh3HasherAvalanche(accumulator[0])
	accumulator[1] = xxh3HasherAvalanche(accumulator[1])

	// ─── Remaining chunk pairs ───
	for i := uint64(4); i < numChunks; i++ {
		block1 := *(*[16]byte)(unsafe.Pointer(&input[i*32]))
		block2 := *(*[16]byte)(unsafe.Pointer(&input[i*32+16]))
		sw1 := *(*[2]uint64)(unsafe.Pointer(&secret[(i-4)*32+3]))
		sw2 := *(*[2]uint64)(unsafe.Pointer(&secret[(i-4)*32+19]))

		dw1 := *(*[2]uint64)(unsafe.Pointer(&block1[0]))
		dw2 := *(*[2]uint64)(unsafe.Pointer(&block2[0]))

		// Mix 1
		hi, lo := bits.Mul64(dw1[0]^(sw1[0]+hasher.seed), dw1[1]^(sw1[1]-hasher.seed))
		accumulator[0] += lo ^ hi

		// Mix 2
		hi2, lo2 := bits.Mul64(dw2[0]^(sw2[0]+hasher.seed), dw2[1]^(sw2[1]-hasher.seed))
		accumulator[1] += lo2 ^ hi2

		accumulator[0] ^= dw2[0] + dw2[1]
		accumulator[1] ^= dw1[0] + dw1[1]
	}

	// ─── Final two blocks ───
	block1 := *(*[16]byte)(unsafe.Pointer(&input[inputLength-16]))
	block2 := *(*[16]byte)(unsafe.Pointer(&input[inputLength-32]))
	sw1 := *(*[2]uint64)(unsafe.Pointer(&secret[103]))
	sw2 := *(*[2]uint64)(unsafe.Pointer(&secret[119]))

	dw1 := *(*[2]uint64)(unsafe.Pointer(&block1[0]))
	dw2 := *(*[2]uint64)(unsafe.Pointer(&block2[0]))

	negatedSeed := -hasher.seed

	hi, lo := bits.Mul64(dw1[0]^(sw1[0]+negatedSeed), dw1[1]^(sw1[1]-negatedSeed))
	hi2, lo2 := bits.Mul64(dw2[0]^(sw2[0]+negatedSeed), dw2[1]^(sw2[1]-negatedSeed))

	accumulator[0] += lo ^ hi
	accumulator[1] += lo2 ^ hi2

	accumulator[0] ^= dw2[0] + dw2[1]
	accumulator[1] ^= dw1[0] + dw1[1]
}

//go:inline
//go:nosplit
func xxh3HasherAccumulate(hasher *XXH3Hasher, accumulator *[8]uint64, stripe [8]uint64, secretOffset uint64, secret []byte) {
	sw := *(*[8]uint64)(unsafe.Pointer(&secret[secretOffset]))

	{
		v := stripe[0] ^ sw[0]
		accumulator[1] += stripe[0]
		accumulator[0] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
	{
		v := stripe[1] ^ sw[1]
		accumulator[0] += stripe[1]
		accumulator[1] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
	{
		v := stripe[2] ^ sw[2]
		accumulator[3] += stripe[2]
		accumulator[2] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
	{
		v := stripe[3] ^ sw[3]
		accumulator[2] += stripe[3]
		accumulator[3] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
	{
		v := stripe[4] ^ sw[4]
		accumulator[5] += stripe[4]
		accumulator[4] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
	{
		v := stripe[5] ^ sw[5]
		accumulator[4] += stripe[5]
		accumulator[5] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
	{
		v := stripe[6] ^ sw[6]
		accumulator[7] += stripe[6]
		accumulator[6] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
	{
		v := stripe[7] ^ sw[7]
		accumulator[6] += stripe[7]
		accumulator[7] += uint64(uint32(v)) * uint64(uint32(v>>32))
	}
}

//go:nosplit
//go:inline
func xxh3HasherAccumulateAll(
	hasher *XXH3Hasher,
	acc *[8]uint64,
	content []byte,
	contentLength uint64,
	secret []byte,
) {
	secretLength := uint64(len(secret))
	stripesPerBlock := (secretLength - 64) / 8
	blockSize := 64 * stripesPerBlock
	sw := *(*[8]uint64)(unsafe.Pointer(&secret[secretLength-64]))

	idx := uint64(0)
	for ; idx+blockSize < contentLength; idx += blockSize {
		block := content[idx : idx+blockSize]

		// ─── process 2 stripes (128 B) per iteration ───
		for n := uint64(0); n+1 < stripesPerBlock; n += 2 {
			s0 := (*[8]uint64)(unsafe.Pointer(&secret[n*8]))
			s1 := (*[8]uint64)(unsafe.Pointer(&secret[(n+1)*8]))
			b0 := (*[8]uint64)(unsafe.Pointer(&block[n*64]))
			b1 := (*[8]uint64)(unsafe.Pointer(&block[(n+1)*64]))

			// Stripe 0
			v0 := b0[0] ^ s0[0]
			v1 := b0[1] ^ s0[1]
			v2 := b0[2] ^ s0[2]
			v3 := b0[3] ^ s0[3]
			v4 := b0[4] ^ s0[4]
			v5 := b0[5] ^ s0[5]
			v6 := b0[6] ^ s0[6]
			v7 := b0[7] ^ s0[7]

			acc[1] += b0[0]
			acc[0] += uint64(uint32(v0)) * uint64(uint32(v0>>32))
			acc[0] += b0[1]
			acc[1] += uint64(uint32(v1)) * uint64(uint32(v1>>32))
			acc[3] += b0[2]
			acc[2] += uint64(uint32(v2)) * uint64(uint32(v2>>32))
			acc[2] += b0[3]
			acc[3] += uint64(uint32(v3)) * uint64(uint32(v3>>32))
			acc[5] += b0[4]
			acc[4] += uint64(uint32(v4)) * uint64(uint32(v4>>32))
			acc[4] += b0[5]
			acc[5] += uint64(uint32(v5)) * uint64(uint32(v5>>32))
			acc[7] += b0[6]
			acc[6] += uint64(uint32(v6)) * uint64(uint32(v6>>32))
			acc[6] += b0[7]
			acc[7] += uint64(uint32(v7)) * uint64(uint32(v7>>32))

			// Stripe 1
			v8 := b1[0] ^ s1[0]
			v9 := b1[1] ^ s1[1]
			v10 := b1[2] ^ s1[2]
			v11 := b1[3] ^ s1[3]
			v12 := b1[4] ^ s1[4]
			v13 := b1[5] ^ s1[5]
			v14 := b1[6] ^ s1[6]
			v15 := b1[7] ^ s1[7]

			acc[1] += b1[0]
			acc[0] += uint64(uint32(v8)) * uint64(uint32(v8>>32))
			acc[0] += b1[1]
			acc[1] += uint64(uint32(v9)) * uint64(uint32(v9>>32))
			acc[3] += b1[2]
			acc[2] += uint64(uint32(v10)) * uint64(uint32(v10>>32))
			acc[2] += b1[3]
			acc[3] += uint64(uint32(v11)) * uint64(uint32(v11>>32))
			acc[5] += b1[4]
			acc[4] += uint64(uint32(v12)) * uint64(uint32(v12>>32))
			acc[4] += b1[5]
			acc[5] += uint64(uint32(v13)) * uint64(uint32(v13>>32))
			acc[7] += b1[6]
			acc[6] += uint64(uint32(v14)) * uint64(uint32(v14>>32))
			acc[6] += b1[7]
			acc[7] += uint64(uint32(v15)) * uint64(uint32(v15>>32))
		}

		// ─── Handle tail stripe if odd count ───
		if stripesPerBlock&1 != 0 {
			n := stripesPerBlock - 1
			s := (*[8]uint64)(unsafe.Pointer(&secret[n*8]))
			b := (*[8]uint64)(unsafe.Pointer(&block[n*64]))
			v0 := b[0] ^ s[0]
			v1 := b[1] ^ s[1]
			v2 := b[2] ^ s[2]
			v3 := b[3] ^ s[3]
			v4 := b[4] ^ s[4]
			v5 := b[5] ^ s[5]
			v6 := b[6] ^ s[6]
			v7 := b[7] ^ s[7]
			acc[1] += b[0]
			acc[0] += uint64(uint32(v0)) * uint64(uint32(v0>>32))
			acc[0] += b[1]
			acc[1] += uint64(uint32(v1)) * uint64(uint32(v1>>32))
			acc[3] += b[2]
			acc[2] += uint64(uint32(v2)) * uint64(uint32(v2>>32))
			acc[2] += b[3]
			acc[3] += uint64(uint32(v3)) * uint64(uint32(v3>>32))
			acc[5] += b[4]
			acc[4] += uint64(uint32(v4)) * uint64(uint32(v4>>32))
			acc[4] += b[5]
			acc[5] += uint64(uint32(v5)) * uint64(uint32(v5>>32))
			acc[7] += b[6]
			acc[6] += uint64(uint32(v6)) * uint64(uint32(v6>>32))
			acc[6] += b[7]
			acc[7] += uint64(uint32(v7)) * uint64(uint32(v7>>32))
		}

		// Scramble accumulators after each block
		acc[0] = (acc[0] ^ (acc[0] >> 47)) ^ sw[0]
		acc[0] *= prime32_1
		acc[1] = (acc[1] ^ (acc[1] >> 47)) ^ sw[1]
		acc[1] *= prime32_1
		acc[2] = (acc[2] ^ (acc[2] >> 47)) ^ sw[2]
		acc[2] *= prime32_1
		acc[3] = (acc[3] ^ (acc[3] >> 47)) ^ sw[3]
		acc[3] *= prime32_1
		acc[4] = (acc[4] ^ (acc[4] >> 47)) ^ sw[4]
		acc[4] *= prime32_1
		acc[5] = (acc[5] ^ (acc[5] >> 47)) ^ sw[5]
		acc[5] *= prime32_1
		acc[6] = (acc[6] ^ (acc[6] >> 47)) ^ sw[6]
		acc[6] *= prime32_1
		acc[7] = (acc[7] ^ (acc[7] >> 47)) ^ sw[7]
		acc[7] *= prime32_1
	}

	// ─── Phase 2: last block ───
	lastBlock := content[idx:contentLength]
	lastBlockSize := uint64(len(lastBlock))
	numFullStripes := (lastBlockSize - 1) / 64

	for n := uint64(0); n < numFullStripes; n++ {
		s := (*[8]uint64)(unsafe.Pointer(&secret[n*8]))
		b := (*[8]uint64)(unsafe.Pointer(&lastBlock[n*64]))

		v0 := b[0] ^ s[0]
		acc[1] += b[0]
		acc[0] += uint64(uint32(v0)) * uint64(uint32(v0>>32))
		v1 := b[1] ^ s[1]
		acc[0] += b[1]
		acc[1] += uint64(uint32(v1)) * uint64(uint32(v1>>32))
		v2 := b[2] ^ s[2]
		acc[3] += b[2]
		acc[2] += uint64(uint32(v2)) * uint64(uint32(v2>>32))
		v3 := b[3] ^ s[3]
		acc[2] += b[3]
		acc[3] += uint64(uint32(v3)) * uint64(uint32(v3>>32))
		v4 := b[4] ^ s[4]
		acc[5] += b[4]
		acc[4] += uint64(uint32(v4)) * uint64(uint32(v4>>32))
		v5 := b[5] ^ s[5]
		acc[4] += b[5]
		acc[5] += uint64(uint32(v5)) * uint64(uint32(v5>>32))
		v6 := b[6] ^ s[6]
		acc[7] += b[6]
		acc[6] += uint64(uint32(v6)) * uint64(uint32(v6>>32))
		v7 := b[7] ^ s[7]
		acc[6] += b[7]
		acc[7] += uint64(uint32(v7)) * uint64(uint32(v7>>32))
	}

	// ─── Phase 3: final overlapping stripe ───
	lastStripe := *(*[8]uint64)(unsafe.Pointer(&content[contentLength-64]))
	sFinal := (*[8]uint64)(unsafe.Pointer(&secret[secretLength-71]))

	v0 := lastStripe[0] ^ sFinal[0]
	acc[1] += lastStripe[0]
	acc[0] += uint64(uint32(v0)) * uint64(uint32(v0>>32))
	v1 := lastStripe[1] ^ sFinal[1]
	acc[0] += lastStripe[1]
	acc[1] += uint64(uint32(v1)) * uint64(uint32(v1>>32))
	v2 := lastStripe[2] ^ sFinal[2]
	acc[3] += lastStripe[2]
	acc[2] += uint64(uint32(v2)) * uint64(uint32(v2>>32))
	v3 := lastStripe[3] ^ sFinal[3]
	acc[2] += lastStripe[3]
	acc[3] += uint64(uint32(v3)) * uint64(uint32(v3>>32))
	v4 := lastStripe[4] ^ sFinal[4]
	acc[5] += lastStripe[4]
	acc[4] += uint64(uint32(v4)) * uint64(uint32(v4>>32))
	v5 := lastStripe[5] ^ sFinal[5]
	acc[4] += lastStripe[5]
	acc[5] += uint64(uint32(v5)) * uint64(uint32(v5>>32))
	v6 := lastStripe[6] ^ sFinal[6]
	acc[7] += lastStripe[6]
	acc[6] += uint64(uint32(v6)) * uint64(uint32(v6>>32))
	v7 := lastStripe[7] ^ sFinal[7]
	acc[6] += lastStripe[7]
	acc[7] += uint64(uint32(v7)) * uint64(uint32(v7>>32))
}

//go:inline
//go:nosplit
func xxh3HasherFinalize(
	hasher *XXH3Hasher,
	accumulator *[8]uint64,
	initValue, secretOffset uint64,
	secret []byte,
) uint64 {
	sw := *(*[8]uint64)(unsafe.Pointer(&secret[secretOffset]))
	result := initValue

	// Round 0
	{
		a0 := accumulator[0] ^ sw[0]
		a1 := accumulator[1] ^ sw[1]
		hi, lo := bits.Mul64(a0, a1)
		result += lo ^ hi
	}

	// Round 1
	{
		a0 := accumulator[2] ^ sw[2]
		a1 := accumulator[3] ^ sw[3]
		hi, lo := bits.Mul64(a0, a1)
		result += lo ^ hi
	}

	// Round 2
	{
		a0 := accumulator[4] ^ sw[4]
		a1 := accumulator[5] ^ sw[5]
		hi, lo := bits.Mul64(a0, a1)
		result += lo ^ hi
	}

	// Round 3
	{
		a0 := accumulator[6] ^ sw[6]
		a1 := accumulator[7] ^ sw[7]
		hi, lo := bits.Mul64(a0, a1)
		result += lo ^ hi
	}

	return xxh3HasherAvalanche(result)
}

// xxh3HasherAvalanche avalanches the value according to the first avalanche variant in spec:
// https://www.ietf.org/archive/id/draft-josefsson-xxhash-00.html#name-final-mixing-step-avalanche
func xxh3HasherAvalanche(x uint64) uint64 {
	avalanched := x ^ (x >> 37)
	avalanched = avalanched * prime_mx1
	avalanched = avalanched ^ (avalanched >> 32)
	return avalanched
}

// xxh3HasherAvalancheXXH64 avalanches the value according to the second avalanche variant in spec:
// https://www.ietf.org/archive/id/draft-josefsson-xxhash-00.html#name-final-mixing-step-avalanche
func xxh3HasherAvalancheXXH64(x uint64) uint64 {
	avalanched := x ^ (x >> 33)
	avalanched = avalanched * prime64_2
	avalanched = avalanched ^ (avalanched >> 29)
	avalanched = avalanched * prime64_3
	avalanched = avalanched ^ (avalanched >> 32)
	return avalanched
}

func xxh3HasherDeriveSecret(seed uint64) []byte {
	derivedSecret := make([]byte, 192)
	copy(derivedSecret, defaultHashSecret[:192])

	for i := 0; i < 12; i++ {
		loOff := i * 16
		hiOff := loOff + 8

		lo := binary.LittleEndian.Uint64(derivedSecret[loOff:])
		hi := binary.LittleEndian.Uint64(derivedSecret[hiOff:])

		lo += seed
		hi -= seed

		binary.LittleEndian.PutUint64(derivedSecret[loOff:], lo)
		binary.LittleEndian.PutUint64(derivedSecret[hiOff:], hi)
	}

	return derivedSecret
}

//go:inline
//go:nosplit
func littleEndianWordGet[T foundation.Unsigned](s []byte, from uint64) T {
	wordSize := memcore.SizeOf[T]()

	switch wordSize {
	case 1:
		return T(uint8(s[from]))
	case 2:
		return T(uint16(binary.LittleEndian.Uint16(s[from:])))
	case 4:
		return T(uint32(binary.LittleEndian.Uint32(s[from:])))
	case 8:
		return T(uint64(binary.LittleEndian.Uint64(s[from:])))
	default:
		panic("unsupported word size")
	}
}

//go:inline
//go:nosplit
func byteSwap32(x uint32) uint32 {
	return (x >> 24) | ((x >> 8) & 0x0000FF00) | ((x << 8) & 0x00FF0000) | (x << 24)
}

//go:inline
//go:nosplit
func byteSwap64(x uint64) uint64 {
	return (x >> 56) |
		((x >> 40) & 0x000000000000FF00) |
		((x >> 24) & 0x0000000000FF0000) |
		((x >> 8) & 0x00000000FF000000) |
		((x << 8) & 0x000000FF00000000) |
		((x << 24) & 0x0000FF0000000000) |
		((x << 40) & 0x00FF000000000000) |
		(x << 56)
}

//go:inline
//go:nosplit
func rotateBits32(x uint32, n uint) uint32 {
	n &= 31
	return (x << n) | (x >> (32 - n))
}

//go:inline
//go:nosplit
func rotateBits64(x uint64, n uint) uint64 {
	n &= 63
	return (x << n) | (x >> (64 - n))
}
