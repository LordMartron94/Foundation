package formatting

import (
	"fmt"
	"math"
)

/*
FormatDurationNSF64 formats a duration represented in nanoseconds into a readable unit.

Unit selection:
- ns for values < 1e3
- us for values < 1e6
- ms for values < 1e9
- s  for values >= 1e9

Time complexity: O(1)
Space complexity: O(1)
*/
func FormatDurationNSF64(ns float64) string {
	if math.IsNaN(ns) || math.IsInf(ns, 0) {
		return "N/A"
	}

	switch {
	case ns < 1_000:
		return fmt.Sprintf("%.1f ns", ns)
	case ns < 1_000_000:
		return fmt.Sprintf("%.2f us", ns/1_000.0)
	case ns < 1_000_000_000:
		return fmt.Sprintf("%.2f ms", ns/1_000_000.0)
	default:
		return fmt.Sprintf("%.2f s", ns/1_000_000_000.0)
	}
}

/*
FormatMemoryBytesF64 formats a byte quantity using binary units with suffixes.

Time complexity: O(1)
Space complexity: O(1)
*/
func FormatMemoryBytesF64(bytes float64) string {
	if math.IsNaN(bytes) || math.IsInf(bytes, 0) {
		return "N/A"
	}
	if bytes < 0 {
		return "-" + FormatMemoryBytesF64(-bytes)
	}

	switch {
	case bytes < 1024:
		return fmt.Sprintf("%.0f B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.2f KiB", bytes/1024.0)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.2f MiB", bytes/(1024.0*1024.0))
	default:
		return fmt.Sprintf("%.2f GiB", bytes/(1024.0*1024.0*1024.0))
	}
}

/*
FormatThroughputOpsPerSecF64 formats an operation throughput value in ops/sec.

Time complexity: O(1)
Space complexity: O(1)
*/
func FormatThroughputOpsPerSecF64(opsPerSec float64) string {
	return fmt.Sprintf("%s ops/sec", FormatNumberCompactF64(opsPerSec))
}

/*
FormatFLOPSF64 formats floating-point throughput with compact suffixes.

Time complexity: O(1)
Space complexity: O(1)
*/
func FormatFLOPSF64(flops float64) string {
	return fmt.Sprintf("%s FLOPS", FormatNumberCompactF64(flops))
}

/*
FormatMemoryThroughputBytesPerSecF64 formats memory throughput in bytes per second.

Time complexity: O(1)
Space complexity: O(1)
*/
func FormatMemoryThroughputBytesPerSecF64(bytesPerSec float64) string {
	return fmt.Sprintf("%s/s", FormatMemoryBytesF64(bytesPerSec))
}

