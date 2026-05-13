package entropy

// TODO - optimize performance by using custom assembly (see xxh3 hashing and Blaze lib for reference)

import "foundation"

/*
EntropyProvider is a mutable pseudo-random stream provider.

[Context]
The struct stores generator operations as function fields so call sites can
pass and swap providers explicitly without relying on package-global RNG
state.

[Contract]
- NextUint64 returns the next 64-bit value from the provider stream.
- NextFloat64 returns a value in [0, 1) derived from the provider stream.

[Use Cases]
  - Reproducible test fixtures that require deterministic pseudo-random values.
  - Simulation and sampling pipelines that need explicit seed control.
  - Procedural content generation where stream ownership must be local to a
    component.

[Side Effects]
Calls mutate internal closure-captured state for the concrete provider
instance. The type is not synchronized and is not safe for concurrent access
without external locking.
*/
type EntropyProvider struct {
	NextUint64  func() uint64
	NextFloat64 func() float64
}

/*
EntropyProviderCreateMixSplit64 creates a provider backed by SplitMix64.

[Context]
SplitMix64 is a small, fast generator useful for reproducible non-crypto
randomness where simple seeding and low overhead are preferred.

[Algorithmic Approach]
The algorithm increments a 64-bit state by a fixed odd Weyl constant each call,
then applies a sequence of xor-shift and multiplication mix rounds. This
stateless mixing permutation decorrelates nearby states into well-distributed
output values while preserving deterministic replay for a given seed.

[Use Cases]
- Fast deterministic sequence generation for tests and fuzz scaffolding.
- Seed expansion step when initializing other PRNG families.
- Lightweight randomization for non-security-sensitive utilities.

[Parameters]
seed initializes the internal 64-bit state.

[Returns]
A new provider with independent mutable state.

[Side Effects]
The returned provider mutates only its own internal state when called.

[Complexity]
Time: O(1) per generated value.
Space: O(1) provider state.
*/
func EntropyProviderCreateMixSplit64(seed uint64) *EntropyProvider {
	// source: https://rosettacode.org/wiki/Pseudo-random_numbers/Splitmix64

	state := seed

	nextUint64 := func() uint64 {
		state += 0x9e3779b97f4a7c15
		z := state
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		return z ^ (z >> 31)
	}

	return &EntropyProvider{
		NextUint64: nextUint64,
		NextFloat64: func() float64 {
			return nextFloat64(nextUint64)
		},
	}
}

/*
EntropyProviderCreateMixSplit128 creates a SplitMix64-backed provider from a
128-bit seed.

[Context]
Some systems maintain seeds as 128-bit tuples (for example composite run id +
stream id). This constructor accepts that wider seed form directly while
reusing the SplitMix64 execution core for fast deterministic generation.

[Algorithmic Approach]
The Uint128 seed is folded into a single 64-bit value via fold128To64, then
delegated to EntropyProviderCreateMixSplit64. The folding step provides a
deterministic reduction from 128-bit input space to the 64-bit SplitMix state.

[Use Cases]
- Pipelines that store seeds as two-word identifiers.
- Stream partitioning where high/low seed words carry different semantics.
- Compatibility layers migrating from 128-bit seed APIs to SplitMix64.

[Parameters]
seed is the 128-bit input seed.

[Returns]
A new provider with independent mutable state.

[Side Effects]
The returned provider mutates only provider-local internal state.

[Edge Cases]
Different 128-bit seeds can fold to the same 64-bit value; those seeds will
produce identical output streams after folding.

[Complexity]
Initialization: O(1)
Generation: O(1) per value.
Space: O(1)
*/
func EntropyProviderCreateMixSplit128(seed foundation.Uint128) *EntropyProvider {
	foldedSeed := fold128To64(seed)
	return EntropyProviderCreateMixSplit64(foldedSeed)
}

/*
EntropyProviderCreateXoroshiro128 creates a provider backed by xoroshiro128.

[Context]
xoroshiro128 offers high-throughput deterministic generation using a 128-bit
internal state represented by foundation.Uint128.

[Algorithmic Approach]
The generator derives output from bitwise combinations of two 64-bit state
words, then advances state with xor, shifts, and fixed-bit rotations. The
rotation schedule and transition function provide strong speed/quality tradeoffs
for non-cryptographic workloads while keeping deterministic replay.

[Use Cases]
- High-volume Monte Carlo style simulation loops.
- Reproducible procedural systems that consume many random values per frame.
- Benchmarks where PRNG overhead should stay minimal and explicit.

[Parameters]
seed initializes the two 64-bit internal state words.

[Returns]
A new provider with independent mutable state.

[Side Effects]
The returned provider mutates only its own internal state when called.

[Complexity]
Time: O(1) per generated value.
Space: O(1) provider state.
*/
func EntropyProviderCreateXoroshiro128(seed foundation.Uint128) *EntropyProvider {
	// source: https://arxiv.org/pdf/2203.04058

	state0 := seed.Hi
	state1 := seed.Lo

	nextUint64 := func() uint64 {
		stateX := state0 ^ state1
		stateA := state0 & state1
		res := stateX ^ (rotateLeftU64(stateA, 1) | rotateLeftU64(stateA, 2))
		state0 = rotateLeftU64(state0, 55) ^ stateX ^ (stateX << 14)
		state1 = rotateLeftU64(stateX, 36)
		return res
	}

	return &EntropyProvider{
		NextUint64: nextUint64,
		NextFloat64: func() float64 {
			return nextFloat64(nextUint64)
		},
	}
}

/*
EntropyProviderCreateChaCha_Small creates a ChaCha-backed provider from a
compact 64-bit seed.

[Context]
Some call sites only have a small scalar seed available (for example a test id
or run number). This constructor expands that seed into a full 256-bit ChaCha
key while preserving deterministic replay semantics.

[Algorithmic Approach]
The seed is first expanded through SplitMix64 into eight 32-bit words that form
the ChaCha key. The generated key is then delegated to EntropyProviderCreateChaCha_Raw,
which executes the ChaCha quarter-round schedule for the caller-provided round
count (applied as rounds/2 double-rounds) on each 512-bit block.

[Use Cases]
- Deterministic fuzz/test runs keyed by a single integer.
- Procedural systems that store only small seed values.
- Replay pipelines where compact seed serialization is preferred.
- Stream partitioning where one seed is reused with distinct nonces.

[Parameters]
seed initializes deterministic key expansion.
nonce selects a distinct ChaCha stream for the derived key.
rounds configures ChaCha diffusion depth and must be an even number.

[Returns]
A new provider with independent mutable state.

[Side Effects]
The returned provider mutates only provider-local internal state.

[Complexity]
Initialization: O(1)
Generation: O(1) amortized per 64-bit value.
Space: O(1)
*/
func EntropyProviderCreateChaCha_Small(seed uint64, nonce uint64, rounds int) *EntropyProvider {
	splitMixer := EntropyProviderCreateMixSplit64(seed)

	key := [8]uint32{}

	for i := 0; i < 8; i += 2 {
		key[i], key[i+1] = unpackLittleEndian(splitMixer.NextUint64())
	}

	return EntropyProviderCreateChaCha_Raw(key, nonce, rounds)
}

/*
EntropyProviderCreateChaCha_Big creates a ChaCha-backed provider from a
128-bit seed.

[Context]
When a wider seed space is required, this constructor expands Uint128 input
into a full 256-bit ChaCha key while preserving deterministic replay and
explicit provider ownership.

[Algorithmic Approach]
The 128-bit seed initializes xoroshiro128, which emits words used to fill the
8-word ChaCha key. The resulting key is passed to EntropyProviderCreateChaCha_Raw,
which performs ChaCha diffusion using the caller-provided even round count.

[Use Cases]
- Distributed simulations that need lower seed-collision probability.
- Workloads where a 64-bit seed domain is too small.
- Reproducible generation pipelines with structured 128-bit seeds.
- Multiple independent streams generated from one seed via nonce partitioning.

[Parameters]
seed initializes deterministic key expansion.
nonce selects a distinct ChaCha stream for the derived key.
rounds configures ChaCha diffusion depth and must be an even number.

[Returns]
A new provider with independent mutable state.

[Side Effects]
The returned provider mutates only provider-local internal state.

[Complexity]
Initialization: O(1)
Generation: O(1) amortized per 64-bit value.
Space: O(1)
*/
func EntropyProviderCreateChaCha_Big(seed foundation.Uint128, nonce uint64, rounds int) *EntropyProvider {
	xoroshiro := EntropyProviderCreateXoroshiro128(seed)

	key := [8]uint32{}

	for i := 0; i < 8; i += 2 {
		key[i], key[i+1] = unpackLittleEndian(xoroshiro.NextUint64())
	}

	return EntropyProviderCreateChaCha_Raw(key, nonce, rounds)
}

/*
EntropyProviderCreateChaCha_Raw creates a provider backed directly by a
user-specified ChaCha key.

[Context]
This is the lowest-level ChaCha constructor and is intended for callers that
already own key derivation and want exact control over seed material.

[Algorithmic Approach]
The function initializes the ChaCha 4x4 state matrix using sigma constants,
the provided 256-bit key, and a 64-bit block counter plus caller-provided
64-bit nonce (counter starts at zero). For each block:
1) copy the original matrix into a working matrix,
2) run rounds/2 double-rounds (column round + diagonal round),
3) add the original matrix back (feed-forward),
4) increment the block counter.
Generated 32-bit words are buffered and emitted in little-endian packed
64-bit pairs.

[Use Cases]
- Deterministic replay where the full key is persisted externally.
- Interop testing against other ChaCha implementations at reduced round count.
- Controlled pseudo-random streams for simulation and benchmarking.

[Parameters]
key is the 256-bit ChaCha key represented as 8 little-endian uint32 words.
nonce is encoded into the nonce lane of the ChaCha state and separates streams
for the same key.
rounds configures ChaCha diffusion depth and must be an even number.

[Returns]
A new provider with independent mutable state.

[Side Effects]
The returned provider mutates only provider-local internal state (counter,
buffer cursor, and working state). No global state or I/O is touched.

[Thread Safety]
Not safe for concurrent access without external synchronization.

[Security Notes]
This API is designed for deterministic entropy generation, not cryptographic
security boundaries. Callers needing cryptographic guarantees should use a
dedicated crypto package and nonce/key management strategy.

[Exceptions & Errors]
Panics if rounds is not even.

[Complexity]
Time: O(1) amortized per 64-bit value.
Space: O(1) provider state.
*/
func EntropyProviderCreateChaCha_Raw(key [8]uint32, nonce uint64, rounds int) *EntropyProvider {
	// source: https://cr.yp.to/chacha/chacha-20080128.pdf

	if rounds%2 != 0 {
		panic("rounds must be multiple of 2")
	}

	doubleRounds := rounds / 2

	type stateMatrix [16]uint32
	originalMatrix := stateMatrix{}

	originalMatrix[0] = 0x61707865
	originalMatrix[1] = 0x3320646e
	originalMatrix[2] = 0x79622d32
	originalMatrix[3] = 0x6b206574

	originalMatrix[4] = key[0]
	originalMatrix[5] = key[1]
	originalMatrix[6] = key[2]
	originalMatrix[7] = key[3]

	originalMatrix[8] = key[4]
	originalMatrix[9] = key[5]
	originalMatrix[10] = key[6]
	originalMatrix[11] = key[7]

	originalMatrix[12] = 0 // lower block counter
	originalMatrix[13] = 0 // upper block counter
	originalMatrix[14], originalMatrix[15] = unpackLittleEndian(nonce)

	quarterRound := func(stateA, stateB, stateC, stateD uint32) (uint32, uint32, uint32, uint32) {
		// 1.
		stateA += stateB
		stateD ^= stateA
		stateD = rotateLeftU32(stateD, 16)

		// 2.
		stateC += stateD
		stateB ^= stateC
		stateB = rotateLeftU32(stateB, 12)

		// 3.
		stateA += stateB
		stateD ^= stateA
		stateD = rotateLeftU32(stateD, 8)

		// 4.
		stateC += stateD
		stateB ^= stateC
		stateB = rotateLeftU32(stateB, 7)

		return stateA, stateB, stateC, stateD
	}

	buffer := stateMatrix{}
	index := 16

	diffuse := func(workMatrix *stateMatrix) {
		// Column Rounds
		workMatrix[0], workMatrix[4], workMatrix[8], workMatrix[12] = quarterRound(workMatrix[0], workMatrix[4], workMatrix[8], workMatrix[12])
		workMatrix[1], workMatrix[5], workMatrix[9], workMatrix[13] = quarterRound(workMatrix[1], workMatrix[5], workMatrix[9], workMatrix[13])
		workMatrix[2], workMatrix[6], workMatrix[10], workMatrix[14] = quarterRound(workMatrix[2], workMatrix[6], workMatrix[10], workMatrix[14])
		workMatrix[3], workMatrix[7], workMatrix[11], workMatrix[15] = quarterRound(workMatrix[3], workMatrix[7], workMatrix[11], workMatrix[15])

		// Diagonal Rounds
		workMatrix[0], workMatrix[5], workMatrix[10], workMatrix[15] = quarterRound(workMatrix[0], workMatrix[5], workMatrix[10], workMatrix[15])
		workMatrix[1], workMatrix[6], workMatrix[11], workMatrix[12] = quarterRound(workMatrix[1], workMatrix[6], workMatrix[11], workMatrix[12])
		workMatrix[2], workMatrix[7], workMatrix[8], workMatrix[13] = quarterRound(workMatrix[2], workMatrix[7], workMatrix[8], workMatrix[13])
		workMatrix[3], workMatrix[4], workMatrix[9], workMatrix[14] = quarterRound(workMatrix[3], workMatrix[4], workMatrix[9], workMatrix[14])
	}

	generateNextBlock := func() stateMatrix {
		blockMatrix := stateMatrix{}
		copy(blockMatrix[:], originalMatrix[:])

		for i := 0; i < doubleRounds; i++ {
			diffuse(&blockMatrix)
		}

		for i := 0; i < 16; i++ {
			blockMatrix[i] += originalMatrix[i]
		}

		originalMatrix[12] += 1 // increment block counter
		if originalMatrix[12] == 0 {
			originalMatrix[13] += 1
		}

		return blockMatrix
	}

	nextUint64 := func() uint64 {
		if index >= 16 {
			block := generateNextBlock()
			buffer = block
			index = 0
		}

		word1, word2 := buffer[index], buffer[index+1]
		index += 2

		return packLittleEndian(word1, word2)
	}

	return &EntropyProvider{
		NextUint64: nextUint64,
		NextFloat64: func() float64 {
			return nextFloat64(nextUint64)
		},
	}
}

/*
EntropyProviderPCGCreate_Small creates a PCG-backed provider from a compact
64-bit seed.

[Context]
This constructor is used when callers only have a small scalar seed but still
want access to a wider-state PCG stream for deterministic replay and better
state space than single-word generators.

[Algorithmic Approach]
The 64-bit seed is expanded via SplitMix64 into two 64-bit words, which are
packed into a Uint128 initial state. The expanded state is then delegated to
EntropyProviderPCGCreate_Big.

[Use Cases]
- Deterministic test runs keyed by one integer.
- Simple seed persistence for CI and fuzz reproductions.
- Lightweight reproducible entropy where setup ergonomics matter.

[Parameters]
seed initializes deterministic state expansion into Uint128.

[Returns]
A new provider with independent mutable state.

[Side Effects]
The returned provider mutates only provider-local internal state.

[Complexity]
Initialization: O(1)
Generation: O(1) per value.
Space: O(1)
*/
func EntropyProviderPCGCreate_Small(seed uint64) *EntropyProvider {
	splitMixer := EntropyProviderCreateMixSplit64(seed)

	hi := splitMixer.NextUint64()
	lo := splitMixer.NextUint64()

	return EntropyProviderPCGCreate_Big(foundation.Uint128{Hi: hi, Lo: lo})
}

/*
EntropyProviderPCGCreate_Big creates a PCG-XSL-RR provider with a 128-bit
internal state and 64-bit output values.

[Context]
This constructor provides deterministic non-cryptographic entropy with explicit
state ownership, making it suitable for simulations and reproducible sampling
workloads that need stable streams across runs.

[Algorithmic Approach]
The generator uses a linear congruential state transition on Uint128:
state = state * mConst + aConst. Output is produced with an XSL-RR style
permutation by xoring the high and low 64-bit halves and rotating right by a
rotation value derived from the top state bits. The stream output is computed
from the current state, then state is advanced for the next call.

[Use Cases]
- Deterministic Monte Carlo and simulation pipelines.
- Reproducible benchmark input generation.
- Stable pseudo-random streams for property-based and fuzz-style tests.

[Parameters]
seed is the full 128-bit initial state for the generator.

[Returns]
A new provider with independent mutable state.

[Side Effects]
Each call mutates provider-local state. No global state or I/O is touched.

[Thread Safety]
Not safe for concurrent access without external synchronization.

[Security Notes]
PCG here is intended for deterministic non-cryptographic entropy. Do not use
it as a cryptographic RNG.

[Complexity]
Time: O(1) per generated value.
Space: O(1) provider state.
*/
func EntropyProviderPCGCreate_Big(seed foundation.Uint128) *EntropyProvider {
	// source: https://www.pcg-random.org/pdf/hmc-cs-2014-0905.pdf -> PCG-XSL-RR (128-bit state, 64-bit output)
	state := seed

	mConst := foundation.Uint128{
		Hi: 2549297995355413924,
		Lo: 4865540595714422341,
	}

	aConst := foundation.Uint128{
		Hi: 42344321231232,
		Lo: 1,
	}

	updateState := func() {
		state = foundation.Uint128Mul(mConst, state)
		state = foundation.Uint128Add(state, aConst)
	}

	scramble := func() uint64 {
		top := state.Hi
		bottom := state.Lo
		folded := top ^ bottom

		shiftedTop := top >> 58 // extract top 6 bits
		return rotateRightU64(folded, int(shiftedTop))
	}

	nextUint64 := func() uint64 {
		scrambled := scramble()
		updateState()
		return scrambled
	}

	return &EntropyProvider{
		NextUint64: nextUint64,
		NextFloat64: func() float64 {
			return nextFloat64(nextUint64)
		},
	}
}

/*
EntropyProviderCreateTapeReader creates a deterministic provider from a fixed slice.

[Context]
When a fuzzer finds a catastrophic failure, the exact sequence of entropy used
can be saved. This provider replays that exact sequence, guaranteeing the failure
can be reproduced even if the underlying PRNG algorithms change in the future.
*/
func EntropyProviderCreateTapeReader(tape []uint64) *EntropyProvider {
	index := 0
	length := len(tape)

	nextUint64 := func() uint64 {
		if index >= length {
			panic("TapeReader: entropy tape exhausted")
		}
		val := tape[index]
		index++
		return val
	}

	return &EntropyProvider{
		NextUint64: nextUint64,
		NextFloat64: func() float64 {
			return nextFloat64(nextUint64)
		},
	}
}

/*
EntropyProviderCreateCounter creates a completely predictable linear sequence.

[Context]
Used strictly for debugging custom input generators. It allows the developer
to step through their generator logic knowing exactly what the entropy provider
will yield (e.g., 0, 1, 2, 3...).
*/
func EntropyProviderCreateCounter(start uint64, step uint64) *EntropyProvider {
	state := start

	nextUint64 := func() uint64 {
		val := state
		state += step
		return val
	}

	return &EntropyProvider{
		NextUint64: nextUint64,
		NextFloat64: func() float64 {
			return nextFloat64(nextUint64)
		},
	}
}

// ----------------------------------------------------------------- PRIVATE HELPERS

func nextFloat64(nextUint64 func() uint64) float64 {
	// We use the top 53 bits because float64 has 53 bits of mantissa.
	// 1 << 53 is the largest power of 2 that can be represented exactly.
	return float64(nextUint64()>>11) / (1 << 53)
}

//go:nosplit
//go:inline
func packLittleEndian(lsb, msb uint32) uint64 {
	return uint64(lsb) | (uint64(msb) << 32)
}

//go:nosplit
//go:inline
func unpackLittleEndian(val uint64) (uint32, uint32) {
	a := uint32(val)
	b := uint32(val >> 32)
	return a, b
}

//go:nosplit
//go:inline
func rotateLeftU64(val uint64, delta int) uint64 {
	return (val << delta) | (val >> (64 - delta))
}

//go:nosplit
//go:inline
func rotateLeftU32(val uint32, delta int) uint32 {
	return (val << delta) | (val >> (32 - delta))
}

//go:nosplit
//go:inline
func rotateRightU64(val uint64, delta int) uint64 {
	return (val >> delta) | (val << (64 - delta))
}

//go:nosplit
//go:inline
func rotateRightU32(val uint32, delta int) uint32 {
	return (val >> delta) | (val << (32 - delta))
}

//go:nosplit
//go:inline
func fold128To64(val foundation.Uint128) uint64 {
	return val.Hi ^ val.Lo
}
