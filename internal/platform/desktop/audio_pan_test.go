package desktop

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func TestStereoPanReaderPreservesPartialReadsAndTail(t *testing.T) {
	for _, pan := range []float32{-1, 0, 1} {
		reader := &stereoPanReader{source: bytes.NewReader([]byte{0xe8, 3, 0x18, 0xfc})}
		reader.setPan(pan)
		var result []byte
		for {
			var sample [1]byte
			n, err := reader.Read(sample[:])
			result = append(result, sample[:n]...)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		left, right := int16(binary.LittleEndian.Uint16(result)), int16(binary.LittleEndian.Uint16(result[2:]))
		wantLeft, wantRight := int16(1000), int16(-1000)
		if pan < 0 {
			wantRight = 0
		}
		if pan > 0 {
			wantLeft = 0
		}
		if left != wantLeft || right != wantRight {
			t.Fatalf("pan %v: got %d,%d want %d,%d", pan, left, right, wantLeft, wantRight)
		}
	}
}

func TestStereoPanReaderRejectsIncompleteFrame(t *testing.T) {
	reader := &stereoPanReader{source: bytes.NewReader([]byte{1, 2, 3})}
	if _, err := io.ReadAll(reader); err != io.ErrUnexpectedEOF {
		t.Fatalf("got %v", err)
	}
}
