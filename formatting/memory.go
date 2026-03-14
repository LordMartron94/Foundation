package formatting

import (
	"fmt"
	"strconv"
)

const (
	formatMemoryBytesKiB = 1024
	formatMemoryBytesMiB = formatMemoryBytesKiB * 1024
	formatMemoryBytesGiB = formatMemoryBytesMiB * 1024
	formatMemoryBytesTiB = formatMemoryBytesGiB * 1024
)

/*
FormatMemoryBytes formats a byte count as a human-readable string using binary units (B, KiB, MiB, GiB, TiB).

Use cases:
- Allocation overflow and limit error messages
- Logging and diagnostics for memory usage
- Configuration and reporting UIs

Time complexity: O(1)
Space complexity: O(1) for fixed-size output

Edge cases:
- 0 returns "0 B"
- Values >= 1024 use the largest unit that fits (e.g. 1536 → "1.5 KiB")
- Exact multiples render without a decimal (e.g. 1024 → "1 KiB")
*/
func FormatMemoryBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	if bytes < formatMemoryBytesKiB {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	if bytes < formatMemoryBytesMiB {
		return formatMemoryBytesUnit(bytes, formatMemoryBytesKiB, "KiB")
	}
	if bytes < formatMemoryBytesGiB {
		return formatMemoryBytesUnit(bytes, formatMemoryBytesMiB, "MiB")
	}
	if bytes < formatMemoryBytesTiB {
		return formatMemoryBytesUnit(bytes, formatMemoryBytesGiB, "GiB")
	}
	return formatMemoryBytesUnit(bytes, formatMemoryBytesTiB, "TiB")
}

func formatMemoryBytesUnit(n, unit uint64, suffix string) string {
	whole := n / unit
	frac  := n % unit
	if frac == 0 {
		return strconv.FormatUint(whole, 10) + " " + suffix
	}
	tenths := frac * 10 / unit
	return fmt.Sprintf("%d.%d %s", whole, tenths, suffix)
}
