package desktop

import (
	"encoding/binary"
	"io"
	"math"
	"sync/atomic"
)

// stereoPanReader attenuates interleaved PCM16 stereo without changing the stream length.
type stereoPanReader struct {
	source  io.Reader
	pan     atomic.Uint32
	buffer  [4096]byte
	pending []byte
	err     error
}

func (r *stereoPanReader) setPan(pan float32) { r.pan.Store(math.Float32bits(pan)) }

func (r *stereoPanReader) Read(output []byte) (int, error) {
	if len(output) == 0 {
		return 0, nil
	}
	if len(r.pending) == 0 {
		if r.err != nil {
			return 0, r.err
		}
		n, err := io.ReadFull(r.source, r.buffer[:])
		if err == io.ErrUnexpectedEOF && n%4 == 0 {
			err = io.EOF
		}
		r.err = err
		if n%4 != 0 {
			return 0, io.ErrUnexpectedEOF
		}
		pan := math.Float32frombits(r.pan.Load())
		left, right := float32(1), float32(1)
		if pan < 0 {
			right += pan
		} else {
			left -= pan
		}
		for offset := 0; offset < n; offset += 4 {
			for channel, gain := range [2]float32{left, right} {
				data := r.buffer[offset+channel*2:]
				sample := int16(binary.LittleEndian.Uint16(data))
				binary.LittleEndian.PutUint16(data, uint16(int16(float32(sample)*gain)))
			}
		}
		r.pending = r.buffer[:n]
		if n == 0 {
			return 0, err
		}
	}
	n := copy(output, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}
