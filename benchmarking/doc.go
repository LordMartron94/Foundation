/*
Package benchmarking provides helpers for benchmarks and optional JSONL export.

When the environment variable BENCHMARK_METRICS_JSON is set to a file path,
BenchmarkWithMetricsConfig appends one NDJSON sample per benchmark completion
(foundation/benchreport format) for tools such as Anvil. Export never affects
benchmark success when perf counters or host probes are unavailable.

Warmup iterations (BenchmarkMetricsConfig.WarmupIterations and/or BENCHMARK_WARMUP_ITERATIONS) run an optional
warmupFn after prepare/GC and before b.ResetTimer(), so steady-state timing excludes them. Valgrind-oriented
fixed workloads use TestProfile_* tests with ANVIL_PROFILE_WARMUP_ITERATIONS / ANVIL_PROFILE_WORK_ITERATIONS instead
of the testing.B driver.
*/
package benchmarking
