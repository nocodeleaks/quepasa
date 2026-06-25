package calls

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// wavRecorder writes incoming 16 kHz mono frames to a canonical 16-bit PCM WAV file.
// The RIFF and data chunk sizes are placeholders until Close rewrites them with the
// true byte count.
type wavRecorder struct {
	mu      sync.Mutex
	f       *os.File
	w       *bufio.Writer
	dataLen uint32
	closed  bool
	scratch []byte
}

// WAVRecorder creates an AudioSink that records the decoded 16 kHz mono peer audio to a
// 16-bit PCM WAV file at path. Close finalizes the header size fields.
func WAVRecorder(path string) (AudioSink, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	r := &wavRecorder{f: f, w: bufio.NewWriter(f)}
	if err := r.writeHeader(0); err != nil {
		f.Close()
		return nil, err
	}
	return r, nil
}

// writeHeader writes the 44-byte canonical PCM WAV header for the given data byte count.
func (r *wavRecorder) writeHeader(dataLen uint32) error {
	var h [44]byte
	copy(h[0:4], "RIFF")
	binary.LittleEndian.PutUint32(h[4:8], 36+dataLen)
	copy(h[8:12], "WAVE")
	copy(h[12:16], "fmt ")
	binary.LittleEndian.PutUint32(h[16:20], 16) // PCM fmt chunk size
	binary.LittleEndian.PutUint16(h[20:22], 1)  // PCM
	binary.LittleEndian.PutUint16(h[22:24], 1)  // mono
	binary.LittleEndian.PutUint32(h[24:28], SampleRate)
	binary.LittleEndian.PutUint32(h[28:32], SampleRate*2) // byte rate (mono * 2 bytes)
	binary.LittleEndian.PutUint16(h[32:34], 2)            // block align
	binary.LittleEndian.PutUint16(h[34:36], 16)           // bits per sample
	copy(h[36:40], "data")
	binary.LittleEndian.PutUint32(h[40:44], dataLen)
	_, err := r.w.Write(h[:])
	return err
}

// WriteFrame appends one mono frame as s16le PCM.
func (r *wavRecorder) WriteFrame(frame []float32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	if cap(r.scratch) < len(frame)*2 {
		r.scratch = make([]byte, len(frame)*2)
	}
	b := r.scratch[:len(frame)*2]
	for i, s := range frame {
		v := s * 32768.0
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		binary.LittleEndian.PutUint16(b[2*i:], uint16(int16(v)))
	}
	if _, err := r.w.Write(b); err != nil {
		return err
	}
	r.dataLen += uint32(len(b))
	return nil
}

// Close flushes buffered samples and rewrites the RIFF/data size fields. Safe to call
// more than once.
func (r *wavRecorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	if err := r.w.Flush(); err != nil {
		r.f.Close()
		return err
	}
	if _, err := r.f.Seek(0, 0); err != nil {
		r.f.Close()
		return err
	}
	// Rewrite the header in place with the final data length.
	bw := bufio.NewWriter(r.f)
	old := r.w
	r.w = bw
	if err := r.writeHeader(r.dataLen); err != nil {
		r.w = old
		r.f.Close()
		return err
	}
	if err := bw.Flush(); err != nil {
		r.f.Close()
		return err
	}
	r.w = old
	return r.f.Close()
}
