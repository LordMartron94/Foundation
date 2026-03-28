package benchreport

import (
	"encoding/json"
	"os"
	"sync"
)

var fileStates sync.Map // path string -> *benchreportFileState

type benchreportFileState struct {
	mu            sync.Mutex
	headerWritten bool
}

func fileStateForPath(path string) *benchreportFileState {
	if v, ok := fileStates.Load(path); ok {
		return v.(*benchreportFileState)
	}
	s := &benchreportFileState{}
	actual, _ := fileStates.LoadOrStore(path, s)
	return actual.(*benchreportFileState)
}

/*
BenchreportWriteHeader writes the header line once per path (mutex + flag).

Use cases:
- Opening a sidecar run with schema version and environment.

Edge cases:
- Second call is a no-op; returns nil.
*/
func BenchreportWriteHeader(path string, hdr *BenchreportHeader) error {
	if path == "" || hdr == nil {
		return nil
	}
	st := fileStateForPath(path)
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.headerWritten {
		return nil
	}
	hdr.Type = BenchreportRecordTypeHeader
	line, err := json.Marshal(hdr)
	if err != nil {
		return err
	}
	if err := appendLine(path, line); err != nil {
		return err
	}
	st.headerWritten = true
	return nil
}

/*
BenchreportAppendDefinition writes a definition line (for new canonical keys).

Time complexity: O(1) amortized
*/
func BenchreportAppendDefinition(path string, def *BenchreportMetricDefinition) error {
	if path == "" || def == nil {
		return nil
	}
	def.Type = BenchreportRecordTypeDefinition
	line, err := json.Marshal(def)
	if err != nil {
		return err
	}
	st := fileStateForPath(path)
	st.mu.Lock()
	defer st.mu.Unlock()
	return appendLine(path, line)
}

/*
BenchreportAppendSample writes one sample record.

Prerequisites:
- Header should already be written for the same path (call BenchreportWriteHeader first).
*/
func BenchreportAppendSample(path string, sample *BenchreportSample) error {
	if path == "" || sample == nil {
		return nil
	}
	sample.Type = BenchreportRecordTypeSample
	line, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	st := fileStateForPath(path)
	st.mu.Lock()
	defer st.mu.Unlock()
	return appendLine(path, line)
}

func appendLine(path string, line []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(line); err != nil {
		return err
	}
	_, err = f.Write([]byte{'\n'})
	return err
}
