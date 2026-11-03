package sessionquic

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

// Reliable streams may return several frames in one read, or fragment a frame
// across reads. Neither delivery shape may lose the next message's bytes.
func TestReadFramePreservesConsecutiveMessages(t *testing.T) {
	type message struct {
		Text string `json:"text"`
	}

	for _, size := range []struct {
		name          string
		first, second int
	}{
		{"small_small", 5, 9},
		{"small_large", 5, 8192},
		{"large_large", 8192, 8192},
	} {
		for _, fragmented := range []bool{false, true} {
			name := size.name + "/coalesced"
			if fragmented {
				name = size.name + "/fragmented"
			}

			t.Run(name, func(t *testing.T) {
				want := []message{{strings.Repeat("a", size.first)}, {strings.Repeat("b", size.second)}, {"last"}}

				var wire bytes.Buffer
				for _, value := range want {
					if err := writeFrame(&wire, value); err != nil {
						t.Fatal(err)
					}
				}

				var reader io.Reader = &wire
				if fragmented {
					reader = iotest.OneByteReader(reader)
				}

				for index, expected := range want {
					var got message
					if err := readFrame(reader, &got); err != nil {
						t.Fatalf("frame %d: %v", index+1, err)
					}

					if got != expected {
						t.Fatalf("frame %d contents differ", index+1)
					}
				}

				if err := readFrame(reader, &message{}); !errors.Is(err, io.EOF) {
					t.Fatalf("end of stream: got %v, want EOF", err)
				}
			})
		}
	}
}

// A valid frame must not hide a malformed successor in a discarded buffer.
func TestReadFrameRejectsMalformedSuccessor(t *testing.T) {
	for _, suffix := range []struct {
		name string
		wire []byte
		want error
	}{
		{"zero_length", []byte{0, 0, 0, 0}, ErrWire},
		{"truncated_header", []byte{0, 0}, io.ErrUnexpectedEOF},
		{"truncated_body", []byte{0, 0, 0, 2, '{'}, ErrWire},
	} {
		t.Run(suffix.name, func(t *testing.T) {
			var wire bytes.Buffer
			if err := writeFrame(&wire, request{Operation: operationLeave}); err != nil {
				t.Fatal(err)
			}

			wire.Write(suffix.wire)

			if err := readFrame(&wire, &request{}); err != nil {
				t.Fatal(err)
			}

			if err := readFrame(&wire, &request{}); !errors.Is(err, suffix.want) {
				t.Fatalf("successor: got %v, want %v", err, suffix.want)
			}
		})
	}
}
