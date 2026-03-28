package benchreport

/*
BenchreportFormatting carries display preferences for consumers.

Use cases:
- Mirroring foundation/formatting modes in Anvil or other tools.
*/
type BenchreportFormatting struct {
	NumberStyle string `json:"number_style,omitempty"`
}

/*
BenchreportEnvironment is a best-effort snapshot written once in the header.

Use cases:
- Correlating benchmark runs with host state and toolchain.

Edge cases:
- All fields optional; omit when unavailable.
*/
type BenchreportEnvironment struct {
	GoVersion        string  `json:"go_version,omitempty"`
	GOOS             string  `json:"goos,omitempty"`
	GOARCH           string  `json:"goarch,omitempty"`
	PID              int     `json:"pid,omitempty"`
	Hostname         string  `json:"hostname,omitempty"`
	CPUMHzNominal    float64 `json:"cpu_mhz_nominal,omitempty"`
	Load1            float64 `json:"load1,omitempty"`
	Load5            float64 `json:"load5,omitempty"`
	Load15           float64 `json:"load15,omitempty"`
	HWCountersStatus string  `json:"hw_counters_status,omitempty"`
}

/*
BenchreportMetricDefinition declares one metric for tooling registries.

Use cases:
- Streaming new canonical keys before first sample that references them.
*/
type BenchreportMetricDefinition struct {
	Type          string     `json:"type"`
	CanonicalKey  string     `json:"canonical_key"`
	ReportLabel   string     `json:"report_label,omitempty"`
	Kind          MetricKind `json:"kind"`
}

/*
BenchreportHeader is the first line of a metrics sidecar file.

Use cases:
- Versioning and run-wide environment context.
*/
type BenchreportHeader struct {
	Type           string                 `json:"type"`
	SchemaVersion  int                    `json:"schema_version"`
	Formatting     *BenchreportFormatting `json:"formatting,omitempty"`
	Environment    *BenchreportEnvironment `json:"environment,omitempty"`
	HWCountersStatus string               `json:"hw_counters_status,omitempty"`
}

/*
BenchreportWall holds wall-clock bounds for one benchmark sample (-count repetition).

Use cases:
- Correlating samples with OS noise and external events.
*/
type BenchreportWall struct {
	SampleStartRFC3339Nano string `json:"sample_start_rfc3339nano,omitempty"`
	SampleEndRFC3339Nano   string `json:"sample_end_rfc3339nano,omitempty"`
	SampleWallNs           int64  `json:"sample_wall_ns,omitempty"`
}

/*
BenchreportLoadSample repeats load averages at sample time (optional).

Use cases:
- Spotting neighbor load drift across -count iterations.
*/
type BenchreportLoadSample struct {
	Load1  float64 `json:"load1,omitempty"`
	Load5  float64 `json:"load5,omitempty"`
	Load15 float64 `json:"load15,omitempty"`
}

/*
BenchreportSample is one benchmark completion (one line of go test -count output).

Use cases:
- Feeding Anvil without parsing benchmark text for values/kinds.
*/
type BenchreportSample struct {
	Type            string                 `json:"type"`
	BenchmarkName   string                 `json:"benchmark_name"`
	Values          map[string]float64     `json:"values"`
	Wall            *BenchreportWall       `json:"wall,omitempty"`
	Load            *BenchreportLoadSample `json:"load,omitempty"`
}

const (
	BenchreportRecordTypeHeader      = "header"
	BenchreportRecordTypeDefinition  = "definition"
	BenchreportRecordTypeSample      = "sample"
)
