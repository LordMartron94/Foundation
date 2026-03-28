/*
Package benchmarking provides helpers for benchmarks and optional JSONL export.

When the environment variable BENCHMARK_METRICS_JSON is set to a file path,
BenchmarkWithMetricsConfig appends one NDJSON sample per benchmark completion
(foundation/benchreport format) for tools such as Anvil. Export never affects
benchmark success when perf counters or host probes are unavailable.
*/
package benchmarking
