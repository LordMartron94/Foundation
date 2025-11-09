package foundation

import (
	"math"
	"math/bits"
)

// ────────────────────────────────────────────────────────────────
//  Uint128 Implementation
// ────────────────────────────────────────────────────────────────

// Uint128 represents a 128-bit unsigned integer using two uint64 words.
type Uint128 struct {
	Lo uint64 // Least significant 64 bits
	Hi uint64 // Most significant 64 bits
}

// Uint128New constructs a new 128-bit value from the given high and low words.
//
// The conventional ordering is:
//
//	Uint128New(lo, hi) — where lo is the least significant part.
//
// But to match your signature (value1, value2), we’ll assume:
//
//	value1 → lower 64 bits (Lo)
//	value2 → upper 64 bits (Hi)
//
// So Uint128New(0x0123456789ABCDEF, 0xFEDCBA9876543210)
// represents the 128-bit integer 0xFEDCBA98765432100123456789ABCDEF.
func Uint128New(lo, hi uint64) Uint128 {
	return Uint128{
		Lo: lo,
		Hi: hi,
	}
}

// Add returns x + y.
func (x Uint128) Add(y Uint128) Uint128 {
	lo, carry := bits.Add64(x.Lo, y.Lo, 0)
	hi, _ := bits.Add64(x.Hi, y.Hi, carry)
	return Uint128{Lo: lo, Hi: hi}
}

// Xor returns x ^ y.
func (x Uint128) Xor(y Uint128) Uint128 {
	return Uint128{Lo: x.Lo ^ y.Lo, Hi: x.Hi ^ y.Hi}
}

// Mul returns (x * y) mod 2^128 (low 128 bits only).
func (x Uint128) Mul(y Uint128) Uint128 {
	hi, lo := bits.Mul64(x.Lo, y.Lo)
	hi += x.Hi*y.Lo + x.Lo*y.Hi
	return Uint128{Lo: lo, Hi: hi}
}

// ShiftRight returns x >> n.
func (x Uint128) ShiftRight(n uint) Uint128 {
	if n == 0 {
		return x
	}
	if n < 64 {
		return Uint128{
			Lo: (x.Lo >> n) | (x.Hi << (64 - n)),
			Hi: x.Hi >> n,
		}
	}
	if n < 128 {
		return Uint128{
			Lo: x.Hi >> (n - 64),
			Hi: 0,
		}
	}
	return Uint128{}
}

// Zero and Max constants
var (
	Uint128Zero = Uint128{0, 0}
	Uint128Max  = Uint128{Lo: ^uint64(0), Hi: ^uint64(0)}
)

// ────────────────────────────────────────────────────────────────
//  Type Constraints
// ────────────────────────────────────────────────────────────────

type Unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Float interface {
	~float32 | ~float64
}

// Numeric = only built-in types that support arithmetic
type Numeric interface {
	Integer | Float
}

// ExtendedNumeric = Numeric + Uint128 (for min/max, type handling)
type ExtendedNumeric interface {
	Numeric | Uint128
}

// ────────────────────────────────────────────────────────────────
//  Max / Min Value
// ────────────────────────────────────────────────────────────────

func MaxValue[T ExtendedNumeric]() T {
	switch any(*new(T)).(type) {
	case int:
		v := int(^uint(0) >> 1)
		return any(v).(T)
	case int8:
		v := int8(math.MaxInt8)
		return any(v).(T)
	case int16:
		v := int16(math.MaxInt16)
		return any(v).(T)
	case int32:
		v := int32(math.MaxInt32)
		return any(v).(T)
	case int64:
		v := int64(math.MaxInt64)
		return any(v).(T)
	case uint:
		v := ^uint(0)
		return any(v).(T)
	case uint8:
		v := uint8(math.MaxUint8)
		return any(v).(T)
	case uint16:
		v := uint16(math.MaxUint16)
		return any(v).(T)
	case uint32:
		v := uint32(math.MaxUint32)
		return any(v).(T)
	case uint64:
		v := uint64(math.MaxUint64)
		return any(v).(T)
	case uintptr:
		v := ^uintptr(0)
		return any(v).(T)
	case float32:
		v := float32(math.MaxFloat32)
		return any(v).(T)
	case float64:
		v := float64(math.MaxFloat64)
		return any(v).(T)
	case Uint128:
		return any(Uint128Max).(T)
	default:
		panic("foundation.MaxValue: unsupported type")
	}
}

func MinValue[T ExtendedNumeric]() T {
	switch any(*new(T)).(type) {
	case int:
		v := -int(^uint(0)>>1) - 1
		return any(v).(T)
	case int8:
		v := int8(math.MinInt8)
		return any(v).(T)
	case int16:
		v := int16(math.MinInt16)
		return any(v).(T)
	case int32:
		v := int32(math.MinInt32)
		return any(v).(T)
	case int64:
		v := int64(math.MinInt64)
		return any(v).(T)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		v := 0
		return any(v).(T)
	case float32:
		v := float32(-math.MaxFloat32)
		return any(v).(T)
	case float64:
		v := -math.MaxFloat64
		return any(v).(T)
	case Uint128:
		return any(Uint128Zero).(T)
	default:
		panic("foundation.MinValue: unsupported type")
	}
}

// ────────────────────────────────────────────────────────────────
//  Sqrt (Numeric Only — excludes Uint128)
// ────────────────────────────────────────────────────────────────

func Sqrt[T Numeric](x T) T {
	switch v := any(x).(type) {
	// ───── FLOATS ─────
	case float32:
		return T(math.Sqrt(float64(v)))
	case float64:
		return T(math.Sqrt(v))

	// ───── UNSIGNED INTS ─────
	case uint:
		return any(intSqrt(v)).(T)
	case uint8:
		return any(intSqrt(v)).(T)
	case uint16:
		return any(intSqrt(v)).(T)
	case uint32:
		return any(intSqrt(v)).(T)
	case uint64:
		return any(intSqrt(v)).(T)
	case uintptr:
		return any(intSqrt(v)).(T)

	// ───── SIGNED INTS ─────
	case int:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return any(intSqrt(uint(v))).(T)
	case int8:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return any(intSqrt(uint8(v))).(T)
	case int16:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return any(intSqrt(uint16(v))).(T)
	case int32:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return any(intSqrt(uint32(v))).(T)
	case int64:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return any(intSqrt(uint64(v))).(T)

	default:
		panic("foundation.Sqrt: unsupported type")
	}
}

func Sqrt32[T Numeric](x T) float32 {
	f := toFloat64(x)
	if f < 0 {
		return float32(math.NaN())
	}
	return float32(math.Sqrt(f))
}

func Sqrt64[T Numeric](x T) float64 {
	f := toFloat64(x)
	if f < 0 {
		return math.NaN()
	}
	return math.Sqrt(f)
}

// ────────────────────────────────────────────────────────────────
//  Internal Utilities
// ────────────────────────────────────────────────────────────────

//go:inline
//go:nosplit
func intSqrt[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr](x T) T {
	var res T
	var bit T = 1 << (bitSize[T]() - 2)

	for bit > x {
		bit >>= 2
	}
	for bit != 0 {
		if x >= res+bit {
			x -= res + bit
			res = (res >> 1) + bit
		} else {
			res >>= 1
		}
		bit >>= 2
	}
	return res
}

//go:inline
//go:nosplit
func bitSize[T any]() int {
	switch any(*new(T)).(type) {
	case uint8:
		return 8
	case uint16:
		return 16
	case uint32:
		return 32
	case uint64, uintptr:
		return 64
	default:
		return 32
	}
}

//go:inline
func toFloat64(v any) float64 {
	switch n := v.(type) {
	case float32:
		return float64(n)
	case float64:
		return n
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	case uintptr:
		return float64(n)
	default:
		return 0
	}
}
