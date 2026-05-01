package benchreport

/*
CompareSemantics tells consumers how to interpret a positive median delta (new/base − 1)
when comparing benchmark runs.

Use cases:
- Coloring Δ%% in Anvil and other tools without hard-coding metric key lists.

Edge cases:
- Use CompareSemanticsNeutral when direction is ambiguous (e.g. generic ratio without domain).
- Empty value in a definition record means: derive via DefaultCompareSemantics(kind).
*/
type CompareSemantics string

const (
	CompareSemanticsLowerBetter  CompareSemantics = "lower_better"
	CompareSemanticsHigherBetter CompareSemantics = "higher_better"
	CompareSemanticsNeutral      CompareSemantics = "neutral"
)

/*
DefaultCompareSemantics returns the usual interpretation for a metric kind.

Time complexity: O(1)
Space complexity: O(1)

Edge cases:
- Unknown or empty kind is treated like scalar (lower is better).
*/
func DefaultCompareSemantics(kind MetricKind) CompareSemantics {
	switch kind {
	case MetricKindRatePerSec:
		return CompareSemanticsHigherBetter
	case MetricKindDurationNS, MetricKindBytes, MetricKindCount, MetricKindScalar:
		return CompareSemanticsLowerBetter
	case MetricKindRatio, MetricKindHWCounter, MetricKindTimestamp:
		return CompareSemanticsNeutral
	case MetricKindHostGauge:
		return CompareSemanticsLowerBetter
	default:
		return CompareSemanticsLowerBetter
	}
}

/*
CompareSemanticsResolve returns explicit semantics when set, otherwise DefaultCompareSemantics(kind).

Time complexity: O(1)
Space complexity: O(1)
*/
func CompareSemanticsResolve(explicit CompareSemantics, kind MetricKind) CompareSemantics {
	if explicit != "" {
		return explicit
	}
	return DefaultCompareSemantics(kind)
}
