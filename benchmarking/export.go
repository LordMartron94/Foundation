package benchmarking

import (
	"foundation/benchhost"
	"foundation/benchreport"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

/*
EnvBenchmarkResultJSON names the environment variable pointing at the authoritative NDJSON result path.

When unset, export checks EnvBenchmarkMetricsJSON for backward compatibility.
*/
const EnvBenchmarkResultJSON = "BENCHMARK_RESULT_JSON"

/*
EnvBenchmarkMetricsJSON is deprecated; use EnvBenchmarkResultJSON. Same file role.
*/
const EnvBenchmarkMetricsJSON = "BENCHMARK_METRICS_JSON"

/*
EnvBenchmarkTraceOut names the environment variable for runtime/trace output.

When set (e.g. by Anvil observability go_trace), the benchmark timed region is traced to that path.
*/
const EnvBenchmarkTraceOut = "BENCHMARK_TRACE_OUT"

/*
EnvBenchmarkWarmupIterations sets extra untimed iterations before b.ResetTimer() when BenchmarkWithMetricsConfig
is given a non-nil warmupFn. Ignored when unset or non-positive. Anvil may set this from [benchmarking].warmup_iterations.
*/
const EnvBenchmarkWarmupIterations = "BENCHMARK_WARMUP_ITERATIONS"

/*
EnvAnvilProfileWarmupIterations / EnvAnvilProfileWorkIterations are set by Anvil for TestProfile_* scenarios (fixed
workloads under Valgrind). Tests read these in plain loops — not via testing.B.
*/
const EnvAnvilProfileWarmupIterations = "ANVIL_PROFILE_WARMUP_ITERATIONS"

const EnvAnvilProfileWorkIterations = "ANVIL_PROFILE_WORK_ITERATIONS"

/*
EnvAnvilCallgrindInstrRegion is set to "1" by Anvil when [profiling.valgrind] callgrind_client_instr_region is true
and the effective Valgrind tool is callgrind. Benchmark code may use BenchmarkingCallgrindInstrRegionMaybeBegin/End
around hot paths together with Valgrind --instr-atstart=no.
*/
const EnvAnvilCallgrindInstrRegion = "ANVIL_CALLGRIND_INSTR_REGION"

/*
BenchmarkTelemetryConfig controls optional host and wall-clock telemetry.

Zero value enables cheap probes (load, wall timestamps, CPU MHz) and best-effort
hardware counters. Use Omit* fields to disable subsystems.

Edge cases:
- Hardware counters never fail the benchmark; they are omitted when unavailable.
- Hostname is omitted unless IncludeHostname is true (CI / sensitive environments).
*/
type BenchmarkTelemetryConfig struct {
	OmitHostLoad       bool
	OmitCPUFrequency   bool
	OmitHWCounters     bool
	OmitWallTimestamps bool
	OmitPerSampleLoad  bool
	IncludeHostname    bool
}

type metricExportScratch struct {
	path       string
	values     map[string]float64
	defKind    map[string]benchreport.MetricKind
	defLabel   map[string]string
	defCompare map[string]benchreport.CompareSemantics // non-empty overrides DefaultCompareSemantics(kind)
}

var exportRegistry struct {
	mu sync.Mutex
	m  map[*testing.B]*metricExportScratch
}

var exportDefSeen sync.Map // key path+\x00+canonical -> struct{}

var exportHeaderOnce sync.Map // path -> *sync.Once

func exportPath() string {
	if p := os.Getenv(EnvBenchmarkResultJSON); p != "" {
		return p
	}
	return os.Getenv(EnvBenchmarkMetricsJSON)
}

func exportRegisterScratch(b *testing.B, s *metricExportScratch) {
	if b == nil || s == nil || s.path == "" {
		return
	}
	exportRegistry.mu.Lock()
	defer exportRegistry.mu.Unlock()
	if exportRegistry.m == nil {
		exportRegistry.m = make(map[*testing.B]*metricExportScratch)
	}
	exportRegistry.m[b] = s
}

func exportUnregisterScratch(b *testing.B) {
	if b == nil {
		return
	}
	exportRegistry.mu.Lock()
	defer exportRegistry.mu.Unlock()
	delete(exportRegistry.m, b)
}

func exportLookupScratch(b *testing.B) *metricExportScratch {
	if b == nil {
		return nil
	}
	exportRegistry.mu.Lock()
	defer exportRegistry.mu.Unlock()
	return exportRegistry.m[b]
}

func newMetricExportScratch(path string) *metricExportScratch {
	if path == "" {
		return &metricExportScratch{}
	}
	return &metricExportScratch{
		path:     path,
		values:   make(map[string]float64),
		defKind:  make(map[string]benchreport.MetricKind),
		defLabel: make(map[string]string),
	}
}

/*
BenchmarkingReportMetric reports a metric to testing.B and records it for JSON export
when BENCHMARK_RESULT_JSON or BENCHMARK_METRICS_JSON is set (inside BenchmarkWithMetricsConfig).

Use this instead of b.ReportMetric for custom metrics so Anvil receives explicit kinds.

Prerequisites:
  - b must be the same *testing.B passed to BenchmarkWithMetricsConfig (or its sub-bench
    during nested b.Run — only the outer registration is supported; register is per wrapper b).

Edge cases:
- If called outside BenchmarkWithMetricsConfig export registration, behaves like b.ReportMetric only.
*/
func BenchmarkingReportMetric(b *testing.B, value float64, reportLabel string, kind benchreport.MetricKind) {
	scratch := exportLookupScratch(b)
	exportAccumulateMetric(b, scratch, value, reportLabel, kind, "")
}

/*
BenchmarkingReportMetricWithCompare is like BenchmarkingReportMetric but sets compare semantics
explicitly when kind defaults are wrong (e.g. a ratio where higher is better).

Edge cases:
- Pass empty compare to derive semantics from kind (same as BenchmarkingReportMetric).
*/
func BenchmarkingReportMetricWithCompare(b *testing.B, value float64, reportLabel string, kind benchreport.MetricKind, compare benchreport.CompareSemantics) {
	scratch := exportLookupScratch(b)
	exportAccumulateMetric(b, scratch, value, reportLabel, kind, compare)
}

func exportAccumulateMetric(b *testing.B, scratch *metricExportScratch, value float64, reportLabel string, kind benchreport.MetricKind, compareOverride benchreport.CompareSemantics) {
	b.ReportMetric(value, reportLabel)
	exportAddSynthetic(scratch, value, reportLabel, kind, compareOverride)
}

/*
exportAddSynthetic records a metric for JSON export only (no b.ReportMetric).

Use cases:
- Primary ns/op and mem lines emitted by the testing package, mirrored into the sidecar.
*/
func exportAddSynthetic(scratch *metricExportScratch, value float64, reportLabel string, kind benchreport.MetricKind, compareOverride benchreport.CompareSemantics) {
	if scratch == nil || scratch.path == "" || scratch.values == nil {
		return
	}
	canonical := benchreport.BenchreportMetricCanonicalKey(reportLabel)
	scratch.values[canonical] = value
	scratch.defKind[canonical] = kind
	scratch.defLabel[canonical] = reportLabel
	if compareOverride != "" {
		if scratch.defCompare == nil {
			scratch.defCompare = make(map[string]benchreport.CompareSemantics)
		}
		scratch.defCompare[canonical] = compareOverride
	}
}

func exportEnsureHeader(path string, telemetry BenchmarkTelemetryConfig, hwCountersStatus string) {
	if path == "" {
		return
	}
	v, _ := exportHeaderOnce.LoadOrStore(path, &sync.Once{})
	once := v.(*sync.Once)
	once.Do(func() {
		env := &benchreport.BenchreportEnvironment{
			GoVersion:        runtime.Version(),
			GOOS:             runtime.GOOS,
			GOARCH:           runtime.GOARCH,
			PID:              os.Getpid(),
			HWCountersStatus: hwCountersStatus,
		}
		if telemetry.IncludeHostname {
			if h, err := os.Hostname(); err == nil {
				env.Hostname = h
			}
		}
		if !telemetry.OmitCPUFrequency {
			env.CPUMHzNominal = benchhost.BenchhostReadCPUMHzNominal()
		}
		if !telemetry.OmitHostLoad {
			if l1, l5, l15, ok := benchhost.BenchhostReadLoadAvg(); ok {
				env.Load1, env.Load5, env.Load15 = l1, l5, l15
			}
		}
		hdr := &benchreport.BenchreportHeader{
			SchemaVersion:    benchreport.BenchreportSchemaVersionRunEnvelope,
			Formatting:       &benchreport.BenchreportFormatting{NumberStyle: "compact"},
			Environment:      env,
			HWCountersStatus: hwCountersStatus,
		}
		_ = benchreport.BenchreportWriteHeader(path, hdr)
	})
}

func exportAppendDefinitionOnce(path, canonical, reportLabel string, kind benchreport.MetricKind, compareOverride benchreport.CompareSemantics) {
	if path == "" || canonical == "" {
		return
	}
	key := path + "\x00" + canonical
	if _, loaded := exportDefSeen.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	sem := benchreport.CompareSemanticsResolve(compareOverride, kind)
	_ = benchreport.BenchreportAppendDefinition(path, &benchreport.BenchreportMetricDefinition{
		CanonicalKey:     canonical,
		ReportLabel:      reportLabel,
		Kind:             kind,
		CompareSemantics: sem,
	})
}

func exportFlush(
	path string,
	b *testing.B,
	scratch *metricExportScratch,
	telemetry BenchmarkTelemetryConfig,
	wallStart, wallEnd time.Time,
	hwCycles, hwIns, hwMiss uint64,
	hwOK bool,
	hwStatus string,
) {
	if path == "" || scratch == nil || scratch.values == nil {
		return
	}

	exportEnsureHeader(path, telemetry, hwStatus)

	kinds := make(map[string]benchreport.MetricKind, len(scratch.defKind)+8)
	labels := make(map[string]string, len(scratch.defLabel)+8)
	for k, v := range scratch.defKind {
		kinds[k] = v
	}
	for k, v := range scratch.defLabel {
		labels[k] = v
	}
	vals := make(map[string]float64, len(scratch.values)+8)
	for k, v := range scratch.values {
		vals[k] = v
	}

	var wall *benchreport.BenchreportWall
	if !telemetry.OmitWallTimestamps && !wallStart.IsZero() && !wallEnd.IsZero() {
		d := wallEnd.Sub(wallStart)
		wall = &benchreport.BenchreportWall{
			SampleStartRFC3339Nano: wallStart.Format(time.RFC3339Nano),
			SampleEndRFC3339Nano:   wallEnd.Format(time.RFC3339Nano),
			SampleWallNs:           d.Nanoseconds(),
		}
		vals["sample.wall.ns"] = float64(d.Nanoseconds())
		kinds["sample.wall.ns"] = benchreport.MetricKindDurationNS
		labels["sample.wall.ns"] = "sample.wall.ns"
	}

	if hwOK {
		vals["hw.cpu_cycles"] = float64(hwCycles)
		vals["hw.instructions"] = float64(hwIns)
		vals["hw.cache_misses"] = float64(hwMiss)
		kinds["hw.cpu_cycles"] = benchreport.MetricKindHWCounter
		kinds["hw.instructions"] = benchreport.MetricKindHWCounter
		kinds["hw.cache_misses"] = benchreport.MetricKindHWCounter
		labels["hw.cpu_cycles"] = "hw.cpu_cycles"
		labels["hw.instructions"] = "hw.instructions"
		labels["hw.cache_misses"] = "hw.cache_misses"
	}

	var load *benchreport.BenchreportLoadSample
	if !telemetry.OmitPerSampleLoad {
		if l1, l5, l15, ok := benchhost.BenchhostReadLoadAvg(); ok {
			load = &benchreport.BenchreportLoadSample{Load1: l1, Load5: l5, Load15: l15}
			vals["host.load1"] = l1
			vals["host.load5"] = l5
			vals["host.load15"] = l15
			kinds["host.load1"] = benchreport.MetricKindHostGauge
			kinds["host.load5"] = benchreport.MetricKindHostGauge
			kinds["host.load15"] = benchreport.MetricKindHostGauge
			labels["host.load1"] = "host.load1"
			labels["host.load5"] = "host.load5"
			labels["host.load15"] = "host.load15"
		}
	}

	for canonical, kind := range kinds {
		var override benchreport.CompareSemantics
		if scratch.defCompare != nil {
			override = scratch.defCompare[canonical]
		}
		exportAppendDefinitionOnce(path, canonical, labels[canonical], kind, override)
	}

	sample := &benchreport.BenchreportSample{
		BenchmarkName: b.Name(),
		Values:        vals,
		Wall:          wall,
		Load:          load,
	}
	_ = benchreport.BenchreportAppendSample(path, sample)
}
