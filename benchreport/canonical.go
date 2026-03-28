package benchreport

/*
BenchreportMetricCanonicalKey maps a testing.B ReportMetric unit label to the canonical
key used in Anvil aggregation (must stay aligned with tools/anvil normalizeMetricLabel).

Time complexity: O(1)
Space complexity: O(1)
*/
func BenchreportMetricCanonicalKey(reportLabel string) string {
	switch reportLabel {
	case "ns/op":
		return "ns"
	case "B/op":
		return "bytes"
	case "allocs/op":
		return "allocs"
	case "gc.count":
		return "gcCount"
	case "gc.per.op":
		return "gcPerOp"
	case "heap.delta.bytes":
		return "heapDelta"
	case "ns/op.gc.pause":
		return "nsGcPerOp"
	case "ops/sec":
		return "throughput"
	case "sys.bytes":
		return "peakMem"
	case "heap.total.alloc.bytes":
		return "heapTotalAlloc"
	case "heap.inuse.bytes":
		return "heapInuse"
	case "heap.objects":
		return "heapObjects"
	case "mallocs.per.op":
		return "mallocsPerOp"
	case "ns/op.gc.pause.avg":
		return "gcPauseAvg"
	case "flops/sec":
		return "flops"
	case "flops/op":
		return "flopsPerOp"
	case "manual.bytes/op":
		return "manualBytes"
	default:
		return reportLabel
	}
}
