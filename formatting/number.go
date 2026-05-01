package formatting

import (
	"fmt"
	"math"
)

/*
FormatNumberCompactF64 formats a float64 value using decimal compact notation.

Use cases:
- Displaying count-like benchmark metrics (allocs/op, mallocs/op, heap.objects)
- Showing large scalar values in dense table layouts
- Keeping CLI output readable without losing order-of-magnitude meaning

Time complexity: O(1)
Space complexity: O(1)

Edge cases:
- Returns "0" for exactly zero values
- Preserves sign for negative values
- Uses suffixes "", "k", "M", "G", "T", "P", "E" in powers of 1000
- Avoids compact suffixes for values below 1000 in absolute value
*/
func FormatNumberCompactF64(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "N/A"
	}

	if value == 0 {
		return "0"
	}

	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	if value < 1000 {
		return sign + formatCompactMantissa(value)
	}

	units := []string{"", "k", "M", "G", "T", "P", "E"}
	idx := 0
	for value >= 1000 && idx < len(units)-1 {
		value /= 1000
		idx++
	}

	return sign + formatCompactMantissa(value) + units[idx]
}

func formatCompactMantissa(value float64) string {
	switch {
	case value >= 100:
		return fmt.Sprintf("%.0f", math.Round(value))
	case value >= 10:
		return fmt.Sprintf("%.1f", value)
	default:
		return fmt.Sprintf("%.2f", value)
	}
}
