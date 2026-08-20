package assetdecode

import (
	"encoding/binary"
	"testing"
)

// TestBIKMetadata verifies that validated container, timing, and packed audio
// fields retain their exact integer values for native decoder configuration.
func TestBIKMetadata(t *testing.T) {
	data := make([]byte, 60)
	copy(data, "BIKi")
	putTestUint32(data, 4, uint32(len(data)-8))
	putTestUint32(data, 8, 1)
	putTestUint32(data, 12, 16)
	putTestUint32(data, 16, 1)
	putTestUint32(data, 20, 640)
	putTestUint32(data, 24, 480)
	putTestUint32(data, 28, 24)
	putTestUint32(data, 32, 1)
	putTestUint32(data, 40, 1)
	putTestUint32(data, 44, 4096)
	binary.LittleEndian.PutUint16(data[48:50], 44100)
	binary.LittleEndian.PutUint16(data[50:52], 0xe000)
	putTestUint32(data, 52, 7)

	metadata, err := BIK(data)
	if err != nil {
		t.Fatal(err)
	}

	if metadata.Version != "BIKi" || metadata.Width != 640 || metadata.Height != 480 || metadata.Frames != 1 {
		t.Fatalf("metadata = %#v", metadata)
	}

	if len(metadata.AudioTracks) != 1 ||
		metadata.AudioTracks[0].SampleRate != 44100 ||
		metadata.AudioTracks[0].Channels != 2 ||
		metadata.AudioTracks[0].BitsPerSample != 16 ||
		metadata.AudioTracks[0].Codec != "rdft" {
		t.Fatalf("tracks = %#v", metadata.AudioTracks)
	}
}

// TestBIKRejectsMalformedPayloads mutates one validated field at a time so each
// trust-boundary check remains independently observable.
func TestBIKRejectsMalformedPayloads(t *testing.T) {
	valid := make([]byte, 48)
	copy(valid, "BIKi")
	putTestUint32(valid, 4, 40)
	putTestUint32(valid, 8, 1)
	putTestUint32(valid, 16, 1)
	putTestUint32(valid, 20, 640)
	putTestUint32(valid, 24, 480)
	putTestUint32(valid, 28, 24)
	putTestUint32(valid, 32, 1)

	for name, mutate := range map[string]func([]byte){
		"signature":  func(data []byte) { copy(data, "NOPE") },
		"size":       func(data []byte) { putTestUint32(data, 4, 1) },
		"dimensions": func(data []byte) { putTestUint32(data, 20, 0) },
		"rate":       func(data []byte) { putTestUint32(data, 28, 0) },
	} {
		t.Run(name, func(t *testing.T) {
			data := append([]byte(nil), valid...)
			mutate(data)

			if _, err := BIK(data); err == nil {
				t.Fatal("expected malformed BIK error")
			}
		})
	}
}
