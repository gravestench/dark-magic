package assetdecode

import (
	"encoding/binary"
	"testing"
)

func TestBIKMetadata(t *testing.T) {
	data := make([]byte, 60)
	copy(data, "BIKi")
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(data[offset:offset+4], value) }
	put(4, uint32(len(data)-8))
	put(8, 1)
	put(12, 16)
	put(16, 1)
	put(20, 640)
	put(24, 480)
	put(28, 24)
	put(32, 1)
	put(40, 1)
	put(44, 4096)
	binary.LittleEndian.PutUint16(data[48:50], 44100)
	binary.LittleEndian.PutUint16(data[50:52], 0xe000)
	put(52, 7)
	metadata, err := BIK(data)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Version != "BIKi" || metadata.Width != 640 || metadata.Height != 480 || metadata.Frames != 1 {
		t.Fatalf("metadata = %#v", metadata)
	}
	if len(metadata.AudioTracks) != 1 || metadata.AudioTracks[0].SampleRate != 44100 || metadata.AudioTracks[0].Channels != 2 || metadata.AudioTracks[0].BitsPerSample != 16 || metadata.AudioTracks[0].Codec != "rdft" {
		t.Fatalf("tracks = %#v", metadata.AudioTracks)
	}
}

func TestBIKRejectsMalformedPayloads(t *testing.T) {
	valid := make([]byte, 48)
	copy(valid, "BIKi")
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(valid[offset:offset+4], value) }
	put(4, 40)
	put(8, 1)
	put(16, 1)
	put(20, 640)
	put(24, 480)
	put(28, 24)
	put(32, 1)
	for name, mutate := range map[string]func([]byte){
		"signature":  func(data []byte) { copy(data, "NOPE") },
		"size":       func(data []byte) { binary.LittleEndian.PutUint32(data[4:8], 1) },
		"dimensions": func(data []byte) { binary.LittleEndian.PutUint32(data[20:24], 0) },
		"rate":       func(data []byte) { binary.LittleEndian.PutUint32(data[28:32], 0) },
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
