package hash

import (
	"encoding/binary"
	"fmt"
	"foundation/benchmarking"
	"hash/maphash"
	"testing"
)

// ────────────────────────────────────────────────────────────────
//   MAIN BENCHMARK SUITE
// ────────────────────────────────────────────────────────────────

func BenchmarkHashSuite(b *testing.B) {
	sizes := []int{
		0, 3, 7, 15, 64, 128, 240, 512, 1024, 4096, 16384,
	}
	seeds := []uint64{0, 12345, 0xDEADBEEFCAFEBABE}

	for _, n := range sizes {
		for _, seed := range seeds {
			runHash64Bench(b, n, seed)
			runHash128Bench(b, n, seed)
		}
	}
}

// ────────────────────────────────────────────────────────────────
//   BENCHMARK IMPLEMENTATIONS
// ────────────────────────────────────────────────────────────────

func runHash64Bench(b *testing.B, size int, seed uint64) {
	type benchData struct {
		hasher *XXH3Hasher
		data   []byte
	}

	config := fmt.Sprintf("N=%d/Seed=%d", size, seed)
	name := "XXH3_64"

	benchmarking.BenchmarkSetup(b, name, config,
		func(b *testing.B) benchData {
			data := randomBytes(size, 42)
			h := XXH3HasherCreateWithSeed(seed)
			return benchData{h, data}
		},
		func(d benchData, b *testing.B) {
			var res uint64
			for i := 0; i < b.N; i++ {
				res = XXH3HasherHash64(d.hasher, d.data)
			}
			if res == 0 {
				b.Logf("dummy: %x", res)
			}
		},
		func(d benchData, b *testing.B) {},
	)
}

func runHash128Bench(b *testing.B, size int, seed uint64) {
	type benchData struct {
		hasher *XXH3Hasher
		data   []byte
	}

	config := fmt.Sprintf("N=%d/Seed=%d", size, seed)
	name := "XXH3_128"

	benchmarking.BenchmarkSetup(b, name, config,
		func(b *testing.B) benchData {
			data := randomBytes(size, 42)
			h := XXH3HasherCreateWithSeed(seed)
			return benchData{h, data}
		},
		func(d benchData, b *testing.B) {
			var lo, hi uint64
			for i := 0; i < b.N; i++ {
				lo, hi = XXH3HasherHash128(d.hasher, d.data)
			}
			if lo == 0 || hi == 0 {
				b.Logf("dummy: %x, %x", lo, hi)
			}
		},
		func(d benchData, b *testing.B) {},
	)
}

// ────────────────────────────────────────────────────────────────
//   CROSS-COMPARISON: VS GOLANG MAPHASH
// ────────────────────────────────────────────────────────────────

func BenchmarkHash_Comparison_Maphash(b *testing.B) {
	sizes := []int{8, 64, 512, 4096, 16384}
	data := make([][]byte, len(sizes))
	for i, n := range sizes {
		data[i] = randomBytes(n, int64(n*17))
	}

	type benchData1 struct {
		hasher maphash.Hash
		data   []byte
	}

	type benchData2 struct {
		hasher *XXH3Hasher
		data   []byte
	}

	for _, input := range data {
		config := fmt.Sprintf("N=%d", len(input))

		benchmarking.BenchmarkSetup(b, "maphash.Hash64", config,
			func(b *testing.B) benchData1 {
				return benchData1{
					hasher: maphash.Hash{},
					data:   input,
				}
			},
			func(data benchData1, b *testing.B) {
				for i := 0; i < b.N; i++ {
					data.hasher.Reset()
					data.hasher.Write(input)
					_ = data.hasher.Sum64()
				}
			},
			func(data benchData1, b *testing.B) {

			},
		)

		benchmarking.BenchmarkSetup(b, "XXH3_64", config,
			func(b *testing.B) benchData2 {
				return benchData2{
					hasher: XXH3HasherCreateWithSeed(0),
					data:   input,
				}
			},
			func(data benchData2, b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = XXH3HasherHash64(data.hasher, input)
				}
			},
			func(data benchData2, b *testing.B) {

			},
		)
	}
}

type nilType struct{}

var nilVar nilType = struct{}{}

// ────────────────────────────────────────────────────────────────
//   MICROBENCHMARKS: AVALANCHE / UTILS
// ────────────────────────────────────────────────────────────────

func BenchmarkHash_Avalanche(b *testing.B) {
	var x uint64 = 0x123456789ABCDEF0
	config := fmt.Sprintf("X=%x", x)

	benchmarking.BenchmarkSetup(b, "avalanche", config,
		func(b *testing.B) nilType {
			return nilVar
		},
		func(data nilType, b *testing.B) {
			for i := 0; i < b.N; i++ {
				x = xxh3HasherAvalanche(x)
			}
			_ = x
		},
		func(data nilType, b *testing.B) {

		},
	)
}

func BenchmarkHash_AvalancheXXH64(b *testing.B) {
	var x uint64 = 0xFEDCBA9876543210
	config := fmt.Sprintf("X=%x", x)

	benchmarking.BenchmarkSetup(b, "avalanche_xxh64", config,
		func(b *testing.B) nilType {
			return nilVar
		},
		func(data nilType, b *testing.B) {
			for i := 0; i < b.N; i++ {
				x = xxh3HasherAvalancheXXH64(x)
			}
			_ = x
		},
		func(data nilType, b *testing.B) {

		},
	)
}

func BenchmarkHash_DeriveSecret(b *testing.B) {
	config := ""

	benchmarking.BenchmarkSetup(b, "derive_secret", config,
		func(b *testing.B) nilType {
			return nilVar
		},
		func(data nilType, b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = xxh3HasherDeriveSecret(uint64(i))
			}
		},
		func(data nilType, b *testing.B) {

		},
	)
}

func BenchmarkHash_ByteSwap64(b *testing.B) {
	val := binary.LittleEndian.Uint64([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88})
	config := ""

	benchmarking.BenchmarkSetup(b, "byte_swap_64", config,
		func(b *testing.B) nilType {
			return nilVar
		},
		func(data nilType, b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = byteSwap64(val)
			}
		},
		func(data nilType, b *testing.B) {

		},
	)
}
