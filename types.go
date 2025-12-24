package foundation

import (
	"math"
	"math/bits"
)

const (
	Pi    = math.Pi
	E     = math.E
	Phi   = 1.61803398874989484820
	Sqrt2 = math.Sqrt2
	Ln2   = math.Ln2
	Ln10  = math.Ln10
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
// But to match your signature (value1, value2), we'll assume:
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
//  Type Conversion (SRP: Conversion Responsibility)
// ────────────────────────────────────────────────────────────────

// FromInt converts an int value to any numeric type T.
// This function handles type conversion from int to all supported numeric types.
func FromInt[T Numeric](val int) T {
	switch any(*new(T)).(type) {
	case int:
		return any(val).(T)
	case int8:
		return any(int8(val)).(T)
	case int16:
		return any(int16(val)).(T)
	case int32:
		return any(int32(val)).(T)
	case int64:
		return any(int64(val)).(T)
	case uint:
		if val < 0 {
			panic("foundation.FromInt: negative value cannot be converted to unsigned type")
		}
		return any(uint(val)).(T)
	case uint8:
		if val < 0 {
			panic("foundation.FromInt: negative value cannot be converted to unsigned type")
		}
		return any(uint8(val)).(T)
	case uint16:
		if val < 0 {
			panic("foundation.FromInt: negative value cannot be converted to unsigned type")
		}
		return any(uint16(val)).(T)
	case uint32:
		if val < 0 {
			panic("foundation.FromInt: negative value cannot be converted to unsigned type")
		}
		return any(uint32(val)).(T)
	case uint64:
		if val < 0 {
			panic("foundation.FromInt: negative value cannot be converted to unsigned type")
		}
		return any(uint64(val)).(T)
	case uintptr:
		if val < 0 {
			panic("foundation.FromInt: negative value cannot be converted to unsigned type")
		}
		return any(uintptr(val)).(T)
	case float32:
		return any(float32(val)).(T)
	case float64:
		return any(float64(val)).(T)
	default:
		panic("foundation.FromInt: unsupported type")
	}
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
//  Absolute Value (SRP: Absolute Value Operation)
// ────────────────────────────────────────────────────────────────

// Abs returns the absolute value of x.
// For unsigned types, returns x unchanged.
// For signed types, returns -x if x < 0, otherwise x.
// For floating types, uses math.Abs.
func Abs[T Numeric](x T) T {
	switch v := any(x).(type) {
	case float32:
		return T(math.Abs(float64(v)))
	case float64:
		return T(math.Abs(v))
	case int:
		if v < 0 {
			return T(-v)
		}
		return T(v)
	case int8:
		if v < 0 {
			return T(-v)
		}
		return T(v)
	case int16:
		if v < 0 {
			return T(-v)
		}
		return T(v)
	case int32:
		if v < 0 {
			return T(-v)
		}
		return T(v)
	case int64:
		if v < 0 {
			return T(-v)
		}
		return T(v)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return x
	default:
		panic("foundation.Abs: unsupported type")
	}
}

// ────────────────────────────────────────────────────────────────
//  Sign Function (SRP: Sign Determination)
// ────────────────────────────────────────────────────────────────

// Sign returns the sign of x: -1 if x < 0, 0 if x == 0, 1 if x > 0.
// For unsigned types, returns 0 if x == 0, 1 if x > 0.
func Sign[T Numeric](x T) int {
	switch v := any(x).(type) {
	case float32:
		if v < 0 {
			return -1
		}
		if v > 0 {
			return 1
		}
		return 0
	case float64:
		if v < 0 {
			return -1
		}
		if v > 0 {
			return 1
		}
		return 0
	case int:
		if v < 0 {
			return -1
		}
		if v > 0 {
			return 1
		}
		return 0
	case int8:
		if v < 0 {
			return -1
		}
		if v > 0 {
			return 1
		}
		return 0
	case int16:
		if v < 0 {
			return -1
		}
		if v > 0 {
			return 1
		}
		return 0
	case int32:
		if v < 0 {
			return -1
		}
		if v > 0 {
			return 1
		}
		return 0
	case int64:
		if v < 0 {
			return -1
		}
		if v > 0 {
			return 1
		}
		return 0
	case uint, uint8, uint16, uint32, uint64, uintptr:
		if v == 0 {
			return 0
		}
		return 1
	default:
		panic("foundation.Sign: unsupported type")
	}
}

// ────────────────────────────────────────────────────────────────
//  Logarithmic Functions (SRP: Logarithmic Operations)
// ────────────────────────────────────────────────────────────────

// Log returns the natural logarithm of x.
// For integer types, converts to float64, computes log, and returns as the same type.
// For floating types, uses math.Log.
func Log[T Numeric](x T) T {
	f := toFloat64(x)
	if f <= 0 {
		panic("foundation.Log: argument must be positive")
	}
	result := math.Log(f)
	return fromFloat64[T](result)
}

// Log32 returns the natural logarithm of x as float32.
func Log32[T Numeric](x T) float32 {
	f := toFloat64(x)
	if f <= 0 {
		return float32(math.NaN())
	}
	return float32(math.Log(f))
}

// Log64 returns the natural logarithm of x as float64.
func Log64[T Numeric](x T) float64 {
	f := toFloat64(x)
	if f <= 0 {
		return math.NaN()
	}
	return math.Log(f)
}

// ────────────────────────────────────────────────────────────────
//  Exponential Functions (SRP: Exponential Operations)
// ────────────────────────────────────────────────────────────────

// Exp returns e^x, the base-e exponential of x.
// For integer types, converts to float64, computes exp, and returns as the same type.
// For floating types, uses math.Exp.
func Exp[T Numeric](x T) T {
	f := toFloat64(x)
	result := math.Exp(f)
	return fromFloat64[T](result)
}

// Exp32 returns e^x as float32.
func Exp32[T Numeric](x T) float32 {
	return float32(math.Exp(toFloat64(x)))
}

// Exp64 returns e^x as float64.
func Exp64[T Numeric](x T) float64 {
	return math.Exp(toFloat64(x))
}

// ────────────────────────────────────────────────────────────────
//  Square Root (SRP: Square Root Operation)
// ────────────────────────────────────────────────────────────────

func Sqrt[T Numeric](x T) T {
	switch v := any(x).(type) {
	case float32:
		return T(math.Sqrt(float64(v)))
	case float64:
		return T(math.Sqrt(v))
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
//  Power Functions (SRP: Power Operations)
// ────────────────────────────────────────────────────────────────

// Pow returns x^y.
// For floating point types, it uses math.Pow.
// For integer types, it uses the efficient "Exponentiation by Squaring" algorithm.
func Pow[T Numeric](x T, y T) T {
	switch v := any(x).(type) {
	case float32:
		return T(math.Pow(float64(v), float64(any(y).(float32))))
	case float64:
		return T(math.Pow(v, any(y).(float64)))
	case int:
		return any(intPow(v, any(y).(int))).(T)
	case int8:
		return any(intPow(v, any(y).(int8))).(T)
	case int16:
		return any(intPow(v, any(y).(int16))).(T)
	case int32:
		return any(intPow(v, any(y).(int32))).(T)
	case int64:
		return any(intPow(v, any(y).(int64))).(T)
	case uint:
		return any(intPow(v, any(y).(uint))).(T)
	case uint8:
		return any(intPow(v, any(y).(uint8))).(T)
	case uint16:
		return any(intPow(v, any(y).(uint16))).(T)
	case uint32:
		return any(intPow(v, any(y).(uint32))).(T)
	case uint64:
		return any(intPow(v, any(y).(uint64))).(T)
	case uintptr:
		return any(intPow(v, any(y).(uintptr))).(T)
	default:
		panic("foundation.Pow: unsupported type")
	}
}

// Pow32 returns x^y as a float32.
func Pow32[T Numeric](x T, y T) float32 {
	return float32(math.Pow(toFloat64(x), toFloat64(y)))
}

// Pow64 returns x^y as a float64.
func Pow64[T Numeric](x T, y T) float64 {
	return math.Pow(toFloat64(x), toFloat64(y))
}

// ────────────────────────────────────────────────────────────────
//  Rounding Functions (SRP: Rounding Operations)
// ────────────────────────────────────────────────────────────────

// Ceil returns the least integer value greater than or equal to x.
// For integer types, returns x unchanged.
// For floating types, uses math.Ceil.
func Ceil[T Numeric](x T) T {
	switch v := any(x).(type) {
	case float32:
		return T(math.Ceil(float64(v)))
	case float64:
		return T(math.Ceil(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return x
	default:
		panic("foundation.Ceil: unsupported type")
	}
}

// Floor returns the greatest integer value less than or equal to x.
// For integer types, returns x unchanged.
// For floating types, uses math.Floor.
func Floor[T Numeric](x T) T {
	switch v := any(x).(type) {
	case float32:
		return T(math.Floor(float64(v)))
	case float64:
		return T(math.Floor(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return x
	default:
		panic("foundation.Floor: unsupported type")
	}
}

// Round returns the nearest integer, rounding half away from zero.
// For integer types, returns x unchanged.
// For floating types, uses math.Round.
func Round[T Numeric](x T) T {
	switch v := any(x).(type) {
	case float32:
		return T(math.Round(float64(v)))
	case float64:
		return T(math.Round(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return x
	default:
		panic("foundation.Round: unsupported type")
	}
}

// ────────────────────────────────────────────────────────────────
//  Trigonometric Functions (SRP: Trigonometric Operations)
// ────────────────────────────────────────────────────────────────

// Sin returns the sine of the radian argument x.
// For integer types, converts to float64, computes sin, and returns as the same type.
// For floating types, uses math.Sin.
func Sin[T Numeric](x T) T {
	f := toFloat64(x)
	result := math.Sin(f)
	return fromFloat64[T](result)
}

// Cos returns the cosine of the radian argument x.
// For integer types, converts to float64, computes cos, and returns as the same type.
// For floating types, uses math.Cos.
func Cos[T Numeric](x T) T {
	f := toFloat64(x)
	result := math.Cos(f)
	return fromFloat64[T](result)
}

// Tan returns the tangent of the radian argument x.
// For integer types, converts to float64, computes tan, and returns as the same type.
// For floating types, uses math.Tan.
func Tan[T Numeric](x T) T {
	f := toFloat64(x)
	result := math.Tan(f)
	return fromFloat64[T](result)
}

// ────────────────────────────────────────────────────────────────
//  Internal Utilities
// ────────────────────────────────────────────────────────────────

//go:inline
//go:nosplit
func intPow[T Integer](base, exp T) T {
	var zero = FromInt[T](0)
	var one = FromInt[T](1)

	if exp < 0 {
		if base == zero {
			panic("foundation.intPow: division by zero")
		}
		if base == one {
			return one
		}
		var negOne = FromInt[T](-1)
		if base == negOne {
			if exp%2 == zero {
				return one
			}
			return negOne
		}
		return zero
	}

	var res = one
	for exp > zero {
		if exp%2 == one {
			res *= base
		}
		base *= base
		exp /= 2
	}
	return res
}

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

//go:inline
func fromFloat64[T Numeric](f float64) T {
	switch any(*new(T)).(type) {
	case float32:
		return any(float32(f)).(T)
	case float64:
		return any(f).(T)
	case int:
		return any(int(f)).(T)
	case int8:
		return any(int8(f)).(T)
	case int16:
		return any(int16(f)).(T)
	case int32:
		return any(int32(f)).(T)
	case int64:
		return any(int64(f)).(T)
	case uint:
		if f < 0 {
			panic("foundation.fromFloat64: negative value cannot be converted to unsigned type")
		}
		return any(uint(f)).(T)
	case uint8:
		if f < 0 {
			panic("foundation.fromFloat64: negative value cannot be converted to unsigned type")
		}
		return any(uint8(f)).(T)
	case uint16:
		if f < 0 {
			panic("foundation.fromFloat64: negative value cannot be converted to unsigned type")
		}
		return any(uint16(f)).(T)
	case uint32:
		if f < 0 {
			panic("foundation.fromFloat64: negative value cannot be converted to unsigned type")
		}
		return any(uint32(f)).(T)
	case uint64:
		if f < 0 {
			panic("foundation.fromFloat64: negative value cannot be converted to unsigned type")
		}
		return any(uint64(f)).(T)
	case uintptr:
		if f < 0 {
			panic("foundation.fromFloat64: negative value cannot be converted to unsigned type")
		}
		return any(uintptr(f)).(T)
	default:
		panic("foundation.fromFloat64: unsupported type")
	}
}
