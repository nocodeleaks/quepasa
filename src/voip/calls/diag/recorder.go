// Package diag is a developer-only diagnostic recorder for the calls VoIP
// stack. It writes exact, per-category call diagnostics to newline-delimited JSON
// (one <stream>.jsonl file per category) so interop development can compare the
// Go stack's wire behavior and key schedule against a reference byte-for-byte.
//
// This is an explicit maintainer carve-out from the library's sanitized logging:
// a *Recorder MAY dump raw secrets, key material, IVs, ciphertext, plaintext and
// PCM, because it is a local opt-in dev tool, not production logging. It is never
// wired on by default — the top-level program opts in via Client.WithDiagnostics.
package diag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Recorder writes per-stream JSONL diagnostics into a directory. It is safe for
// concurrent use, and every method is nil-safe so callers can hold a nil *Recorder
// when diagnostics are off and emit unconditionally without a guard.
type Recorder struct {
	mu    sync.Mutex
	dir   string
	files map[string]*streamFile
}

// streamFile is one opened <stream>.jsonl and its line encoder.
type streamFile struct {
	f   *os.File
	enc *json.Encoder
}

// NewRecorder creates dir (and parents) and returns a Recorder that opens its
// per-stream files lazily on first Emit. It returns an error only if dir cannot
// be created.
func NewRecorder(dir string) (*Recorder, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("diag: create dir %q: %w", dir, err)
	}
	return &Recorder{dir: dir, files: make(map[string]*streamFile)}, nil
}

// Emit writes one JSON object as a single line to <dir>/<stream>.jsonl, injecting a
// ts_ms millisecond wall-clock timestamp. It is a no-op when r is nil (diagnostics
// off). Unknown stream names are tolerated: any non-empty stream opens its own file.
// A write or open failure is swallowed (diagnostics must never break a live call);
// it does not panic and does not propagate.
func (r *Recorder) Emit(stream string, fields map[string]any) {
	if r == nil || stream == "" {
		return
	}

	// Build the record before taking the lock, then serialize the file open + write.
	// ts_ms is always set by the recorder; a caller-supplied ts_ms is overwritten so
	// the timeline is the recorder's own clock.
	rec := make(map[string]any, len(fields)+1)
	for k, v := range fields {
		rec[k] = v
	}
	rec["ts_ms"] = time.Now().UnixMilli()

	r.mu.Lock()
	defer r.mu.Unlock()

	sf := r.files[stream]
	if sf == nil {
		var err error
		sf, err = r.openStream(stream)
		if err != nil {
			return
		}
		r.files[stream] = sf
	}
	// json.Encoder.Encode writes a trailing newline, giving us JSONL for free.
	_ = sf.enc.Encode(rec)
}

// openStream opens (creating/appending) <dir>/<stream>.jsonl. Caller holds r.mu.
func (r *Recorder) openStream(stream string) (*streamFile, error) {
	path := filepath.Join(r.dir, stream+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &streamFile{f: f, enc: json.NewEncoder(f)}, nil
}

// Close flushes and closes every opened stream file. It is a no-op on a nil
// Recorder and returns the first close error encountered.
func (r *Recorder) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error
	for name, sf := range r.files {
		if err := sf.f.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(r.files, name)
	}
	return firstErr
}
