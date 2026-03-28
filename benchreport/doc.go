/*
Package benchreport defines the JSONL sidecar format for benchmark metrics export.

Each run uses one file (path from BENCHMARK_METRICS_JSON): line-oriented JSON (NDJSON).
The first line is a header record; further lines are definition and sample records.

MetricKind values describe how consumers (e.g. Anvil) should format and interpret values.
*/
package benchreport
