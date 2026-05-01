/*
Package benchreport defines the JSONL sidecar format for benchmark metrics export.

Each run uses one file (path from BENCHMARK_RESULT_JSON, or legacy BENCHMARK_METRICS_JSON): NDJSON.
The first line is a run envelope (type run, schema_version 2) or legacy header (type header); further lines are definition and sample records.

MetricKind values describe how consumers (e.g. Anvil) should format values.
Definition records may include compare_semantics (or consumers derive it via DefaultCompareSemantics(kind))
for interpreting median deltas between runs.
*/
package benchreport
