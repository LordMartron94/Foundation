/*
Package benchhost collects best-effort host telemetry for benchmarks.

Hardware performance counters use Linux perf_event_open when permitted
(see kernel.perf_event_paranoid). Failure is non-fatal: callers receive
a status string and nil counters.
*/
package benchhost
