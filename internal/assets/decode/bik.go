package assetdecode

import (
	"encoding/binary"
	"fmt"
)

// BIKAudioTrack describes one Bink audio stream declared by the container.
type BIKAudioTrack struct {
	ID            uint32 `json:"id"`
	SampleRate    uint16 `json:"sample_rate"`
	Channels      int    `json:"channels"`
	BitsPerSample int    `json:"bits_per_sample"`
	Codec         string `json:"codec"`
	MaxPacketSize uint32 `json:"max_packet_size"`
}

// BIKMetadata contains the bounded container information needed to configure a
// decoder and reject malformed payloads before native code sees them.
type BIKMetadata struct {
	Version              string          `json:"version"`
	FileSize             uint32          `json:"file_size"`
	Frames               uint32          `json:"frames"`
	LargestFrameSize     uint32          `json:"largest_frame_size"`
	Width                uint32          `json:"width"`
	Height               uint32          `json:"height"`
	FrameRateNumerator   uint32          `json:"frame_rate_numerator"`
	FrameRateDenominator uint32          `json:"frame_rate_denominator"`
	DurationMillis       uint64          `json:"duration_millis"`
	Flags                uint32          `json:"flags"`
	AudioTracks          []BIKAudioTrack `json:"audio_tracks"`
}

// BIK parses and validates the Bink 1 container header and frame index bounds.
func BIK(data []byte) (BIKMetadata, error) {
	if len(data) < 44 {
		return BIKMetadata{}, fmt.Errorf("assetdecode: BIK header is truncated: %d bytes", len(data))
	}
	if string(data[:3]) != "BIK" || data[3] < 'b' || data[3] > 'i' {
		return BIKMetadata{}, fmt.Errorf("assetdecode: unsupported BIK signature %q", data[:4])
	}
	u32 := func(offset int) uint32 { return binary.LittleEndian.Uint32(data[offset : offset+4]) }
	declaredSize := u32(4)
	if uint64(declaredSize)+8 != uint64(len(data)) {
		return BIKMetadata{}, fmt.Errorf("assetdecode: BIK size %d does not match payload %d", declaredSize+8, len(data))
	}
	frames, repeatedFrames := u32(8), u32(16)
	if frames == 0 || repeatedFrames != frames {
		return BIKMetadata{}, fmt.Errorf("assetdecode: invalid BIK frame counts %d and %d", frames, repeatedFrames)
	}
	width, height := u32(20), u32(24)
	if width == 0 || height == 0 || width > 7680 || height > 4800 {
		return BIKMetadata{}, fmt.Errorf("assetdecode: invalid BIK dimensions %dx%d", width, height)
	}
	fpsNumerator, fpsDenominator := u32(28), u32(32)
	if fpsNumerator == 0 || fpsDenominator == 0 {
		return BIKMetadata{}, fmt.Errorf("assetdecode: invalid BIK frame rate %d/%d", fpsNumerator, fpsDenominator)
	}
	trackCount := u32(40)
	if trackCount > 256 {
		return BIKMetadata{}, fmt.Errorf("assetdecode: too many BIK audio tracks: %d", trackCount)
	}
	headerSize := uint64(44) + uint64(trackCount)*12 + uint64(frames)*4
	if headerSize > uint64(len(data)) {
		return BIKMetadata{}, fmt.Errorf("assetdecode: BIK track/frame index is truncated")
	}

	tracks := make([]BIKAudioTrack, trackCount)
	packetOffset := 44
	formatOffset := packetOffset + int(trackCount)*4
	idOffset := formatOffset + int(trackCount)*4
	for index := range tracks {
		flags := binary.LittleEndian.Uint16(data[formatOffset+index*4+2 : formatOffset+index*4+4])
		codec := "rdft"
		if flags&0x1000 != 0 {
			codec = "dct"
		}
		channels := 1
		if flags&0x2000 != 0 {
			channels = 2
		}
		bits := 8
		if flags&0x4000 != 0 {
			bits = 16
		}
		tracks[index] = BIKAudioTrack{
			ID:         binary.LittleEndian.Uint32(data[idOffset+index*4 : idOffset+index*4+4]),
			SampleRate: binary.LittleEndian.Uint16(data[formatOffset+index*4 : formatOffset+index*4+2]),
			Channels:   channels, BitsPerSample: bits, Codec: codec,
			MaxPacketSize: binary.LittleEndian.Uint32(data[packetOffset+index*4 : packetOffset+index*4+4]),
		}
	}
	return BIKMetadata{
		Version: string(data[:4]), FileSize: declaredSize + 8, Frames: frames,
		LargestFrameSize: u32(12), Width: width, Height: height,
		FrameRateNumerator: fpsNumerator, FrameRateDenominator: fpsDenominator,
		DurationMillis: uint64(frames) * uint64(fpsDenominator) * 1000 / uint64(fpsNumerator),
		Flags:          u32(36), AudioTracks: tracks,
	}, nil
}
