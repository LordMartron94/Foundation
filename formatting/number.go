package formatting

import (
	"fmt"
	"math"
)

/*
FormatNumberCompactF64 formats a numeric value using SI-style compact suffixes.

Suffixes:
- K: thousand (1e3)
- M: million (1e6)
- G: billion (1e9)
- T: trillion (1e12)
- P: quadrillion (1e15)

Use cases:
- Rendering dashboard and CLI metrics with large magnitudes
- Keeping tabular output readable under tight width constraints

Time complexity: O(1)
Space complexity: O(1)
*/
func FormatNumberCompactF64(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "N/A"
	}

	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	if value < 1000 {
		if value == math.Trunc(value) {
			return fmt.Sprintf("%s%.0f", sign, value)
		}
		return fmt.Sprintf("%s%.2f", sign, value)
	}

	suffixes := []string{"K", "M", "G", "T", "P"}
	scaled := value
	idx := 0
	for scaled >= 1000 && idx < len(suffixes)-1 {
		scaled /= 1000
		idx++
	}

	if scaled >= 100 {
		return fmt.Sprintf("%s%.0f%s", sign, scaled, suffixes[idx])
	}
	if scaled >= 10 {
		return fmt.Sprintf("%s%.1f%s", sign, scaled, suffixes[idx])
	}
	return fmt.Sprintf("%s%.2f%s", sign, scaled, suffixes[idx])
}

