package foundation

import (
	"math"
	"testing"

	foundationtesting "foundation/testing"
)

// TestSqrt validates Sqrt, Sqrt32, and Sqrt64 across
// all numeric types and edge cases using a unified suite.
func TestSqrt(t *testing.T) {
	type testCase struct {
		name     string
		value    any
		expected float64
	}

	cases := []testCase{
		// Signed integers
		{"int/zero", int(0), 0},
		{"int/one", int(1), 1},
		{"int/sixtyfour", int(64), 8},
		{"int/large", int(1024), 32},
		{"int8/small", int8(9), 3},
		{"int16/small", int16(81), 9},
		{"int32/medium", int32(225), 15},
		{"int64/large", int64(4096), 64},

		// Unsigned integers
		{"uint8/basic", uint8(16), 4},
		{"uint16/basic", uint16(25), 5},
		{"uint32/basic", uint32(36), 6},
		{"uint64/basic", uint64(49), 7},

		// Floats
		{"float32/basic", float32(64.0), 8},
		{"float32/decimal", float32(2.25), 1.5},
		{"float64/basic", float64(100.0), 10},
		{"float64/decimal", float64(2.25), 1.5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			switch v := tc.value.(type) {
			case int:
				assertIntSqrt(v, tc.expected, t)
			case int8:
				assertIntSqrt(v, tc.expected, t)
			case int16:
				assertIntSqrt(v, tc.expected, t)
			case int32:
				assertIntSqrt(v, tc.expected, t)
			case int64:
				assertIntSqrt(v, tc.expected, t)

			case uint:
				assertUintSqrt(v, tc.expected, t)
			case uint8:
				assertUintSqrt(v, tc.expected, t)
			case uint16:
				assertUintSqrt(v, tc.expected, t)
			case uint32:
				assertUintSqrt(v, tc.expected, t)
			case uint64:
				assertUintSqrt(v, tc.expected, t)
			case uintptr:
				assertUintSqrt(v, tc.expected, t)

			case float32:
				assertFloatSqrt(v, tc.expected, t)
			case float64:
				assertFloatSqrt(v, tc.expected, t)

			default:
				t.Fatalf("unsupported type in test: %T", v)
			}
		})
	}

	// Negative input tests (float only)
	t.Run("float64/negative", func(t *testing.T) {
		res := Sqrt64(-9.0)
		foundationtesting.Assert(math.IsNaN(res), "expected NaN for negative input", "handled negative float correctly", t)
	})
	t.Run("float32/negative", func(t *testing.T) {
		res := Sqrt32(float32(-9.0))
		foundationtesting.Assert(math.IsNaN(float64(res)), "expected NaN for negative input", "handled negative float correctly", t)
	})
}

// ────────────────────────────────────────────────
// HELPERS
// ────────────────────────────────────────────────

func assertFloatSqrt[T ~float32 | ~float64](v T, expected float64, t *testing.T) {
	result := Sqrt(v)
	foundationtesting.Assert(almostEqual(float64(result), expected), "Sqrt incorrect", "Sqrt correct", t)
	result32 := Sqrt32(v)
	foundationtesting.Assert(almostEqual(float64(result32), expected), "Sqrt32 incorrect", "Sqrt32 correct", t)
	result64 := Sqrt64(v)
	foundationtesting.Assert(almostEqual(result64, expected), "Sqrt64 incorrect", "Sqrt64 correct", t)
}

func assertIntSqrt[T ~int | ~int8 | ~int16 | ~int32 | ~int64](v T, expected float64, t *testing.T) {
	result := Sqrt(v)
	foundationtesting.Assert(float64(result) == expected, "integer sqrt incorrect", "integer sqrt correct", t)
	result32 := Sqrt32(v)
	foundationtesting.Assert(almostEqual(float64(result32), expected), "Sqrt32 incorrect", "Sqrt32 correct", t)
	result64 := Sqrt64(v)
	foundationtesting.Assert(almostEqual(result64, expected), "Sqrt64 incorrect", "Sqrt64 correct", t)
}

func assertUintSqrt[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr](v T, expected float64, t *testing.T) {
	result := Sqrt(v)
	foundationtesting.Assert(float64(result) == expected, "unsigned sqrt incorrect", "unsigned sqrt correct", t)
	result32 := Sqrt32(v)
	foundationtesting.Assert(almostEqual(float64(result32), expected), "Sqrt32 incorrect", "Sqrt32 correct", t)
	result64 := Sqrt64(v)
	foundationtesting.Assert(almostEqual(result64, expected), "Sqrt64 incorrect", "Sqrt64 correct", t)
}

func almostEqual(a, b float64) bool {
	const eps = 1e-7
	return math.Abs(a-b) <= eps
}
