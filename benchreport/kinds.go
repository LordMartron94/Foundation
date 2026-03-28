package benchreport

/*
MetricKind describes the unit / interpretation of a metric value.

Use cases:
- Selecting formatters in analysis tools without string heuristics on metric names.

Edge cases:
- Unknown kinds should be treated as scalar by consumers.
*/
type MetricKind string

const (
	MetricKindDurationNS  MetricKind = "duration_ns"
	MetricKindBytes       MetricKind = "bytes"
	MetricKindCount       MetricKind = "count"
	MetricKindRatePerSec  MetricKind = "rate_per_sec"
	MetricKindRatio       MetricKind = "ratio"
	MetricKindScalar      MetricKind = "scalar"
	MetricKindHWCounter   MetricKind = "hw_counter"
	MetricKindTimestamp   MetricKind = "timestamp"
	MetricKindHostGauge   MetricKind = "host_gauge"
)
