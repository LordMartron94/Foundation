package foundation

import "math"

// Numeric specifies all numeric value types.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// MaxValue returns the maximum representable value for any numeric type.
func MaxValue[T Numeric]() T {
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
	default:
		panic("foundation.MaxValue: unsupported type")
	}
}

// MinValue returns the minimum representable value for any numeric type.
func MinValue[T Numeric]() T {
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
	case uint:
		v := uint(0)
		return any(v).(T)
	case uint8:
		v := uint8(0)
		return any(v).(T)
	case uint16:
		v := uint16(0)
		return any(v).(T)
	case uint32:
		v := uint32(0)
		return any(v).(T)
	case uint64:
		v := uint64(0)
		return any(v).(T)
	case uintptr:
		v := uintptr(0)
		return any(v).(T)
	case float32:
		v := float32(-math.MaxFloat32)
		return any(v).(T)
	case float64:
		v := -math.MaxFloat64
		return any(v).(T)
	default:
		panic("foundation.MinValue: unsupported type")
	}
}

// Sqrt returns the square root of any numeric, automatically choosing
// the most efficient method of computation. It keeps the exact precision
// and type of the numeric.
func Sqrt[T Numeric](x T) T {
	switch v := any(x).(type) {
	// ───── FLOATS ─────
	case float32:
		return T(math.Sqrt(float64(v)))
	case float64:
		return T(math.Sqrt(v))

	// ───── UNSIGNED INTS (native precision, no cast) ─────
	case uint:
		return T(intSqrt(v))
	case uint8:
		return T(intSqrt(v))
	case uint16:
		return T(intSqrt(v))
	case uint32:
		return T(intSqrt(v))
	case uint64:
		return T(intSqrt(v))
	case uintptr:
		return T(intSqrt(v))

	// ───── SIGNED INTS (checked, then reuse unsigned path) ─────
	case int:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return T(intSqrt(uint(v)))
	case int8:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return T(intSqrt(uint8(v)))
	case int16:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return T(intSqrt(uint16(v)))
	case int32:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return T(intSqrt(uint32(v)))
	case int64:
		if v < 0 {
			panic("foundation.Sqrt: negative integer")
		}
		return T(intSqrt(uint64(v)))

	default:
		panic("foundation.Sqrt: unsupported type")
	}
}

// Sqrt32 computes the square root of any numeric with float32 precision.
// Negative integers return NaN, mirroring float behavior.
func Sqrt32[T Numeric](x T) float32 {
	f := toFloat64(x)
	if f < 0 {
		return float32(math.NaN())
	}
	return float32(math.Sqrt(f))
}

// Sqrt64 computes the square root of any numeric with float64 precision.
// Negative integers return NaN, mirroring float behavior.
func Sqrt64[T Numeric](x T) float64 {
	f := toFloat64(x)
	if f < 0 {
		return math.NaN()
	}
	return math.Sqrt(f)
}

// ────────────────────────────────────────────────────────────────
// INTERNAL UTILITIES
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
func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	default:
		return int64(toUint64(v))
	}
}

//go:inline
func toUint64(v any) uint64 {
	switch n := v.(type) {
	case uint:
		return uint64(n)
	case uint8:
		return uint64(n)
	case uint16:
		return uint64(n)
	case uint32:
		return uint64(n)
	case uint64:
		return n
	case uintptr:
		return uint64(n)
	default:
		return 0
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
