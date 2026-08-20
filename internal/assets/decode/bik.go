package assetdecode

import (
	"encoding/binary"
	"fmt"
)

const (
	bikFixedHeaderBytes   = 44
	bikTrackMetadataBytes = 12
	bikFrameIndexBytes    = 4
	maxBIKAudioTracks     = 256
	maxBIKWidth           = 7680
	maxBIKHeight          = 4800
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

// bikHeader holds only fields used after validation, making the trust boundary
// between raw container bytes and metadata construction explicit.
type bikHeader struct {
	version          string
	declaredSize     uint32
	frames           uint32
	largestFrameSize uint32
	width            uint32
	height           uint32
	fpsNumerator     uint32
	fpsDenominator   uint32
	flags            uint32
	trackCount       uint32
}

// BIK parses and validates the Bink 1 container header and frame index bounds
// before returning metadata that is safe to pass to a native decoder.
func BIK(data []byte) (BIKMetadata, error) {
	header, err := parseBIKHeader(data)
	if err != nil {
		return BIKMetadata{}, err
	}

	tracks := parseBIKAudioTracks(data, header.trackCount)

	return BIKMetadata{
		Version:              header.version,
		FileSize:             header.declaredSize + 8,
		Frames:               header.frames,
		LargestFrameSize:     header.largestFrameSize,
		Width:                header.width,
		Height:               header.height,
		FrameRateNumerator:   header.fpsNumerator,
		FrameRateDenominator: header.fpsDenominator,
		DurationMillis:       bikDurationMillis(header),
		Flags:                header.flags,
		AudioTracks:          tracks,
	}, nil
}

// parseBIKHeader validates fields in file order so malformed payloads retain
// the same first failure and error text relied upon by callers and diagnostics.
func parseBIKHeader(data []byte) (bikHeader, error) {
	if len(data) < bikFixedHeaderBytes {
		return bikHeader{}, fmt.Errorf("assetdecode: BIK header is truncated: %d bytes", len(data))
	}

	if string(data[:3]) != "BIK" || data[3] < 'b' || data[3] > 'i' {
		return bikHeader{}, fmt.Errorf("assetdecode: unsupported BIK signature %q", data[:4])
	}

	header := bikHeader{
		version:          string(data[:4]),
		declaredSize:     bikUint32(data, 4),
		frames:           bikUint32(data, 8),
		largestFrameSize: bikUint32(data, 12),
		width:            bikUint32(data, 20),
		height:           bikUint32(data, 24),
		fpsNumerator:     bikUint32(data, 28),
		fpsDenominator:   bikUint32(data, 32),
		flags:            bikUint32(data, 36),
		trackCount:       bikUint32(data, 40),
	}

	declaredSize := header.declaredSize
	if uint64(declaredSize)+8 != uint64(len(data)) {
		return bikHeader{}, fmt.Errorf(
			"assetdecode: BIK size %d does not match payload %d",
			declaredSize+8,
			len(data),
		)
	}

	repeatedFrames := bikUint32(data, 16)
	if header.frames == 0 || repeatedFrames != header.frames {
		return bikHeader{}, fmt.Errorf(
			"assetdecode: invalid BIK frame counts %d and %d",
			header.frames,
			repeatedFrames,
		)
	}

	if header.width == 0 || header.height == 0 || header.width > maxBIKWidth || header.height > maxBIKHeight {
		return bikHeader{}, fmt.Errorf(
			"assetdecode: invalid BIK dimensions %dx%d",
			header.width,
			header.height,
		)
	}

	if header.fpsNumerator == 0 || header.fpsDenominator == 0 {
		return bikHeader{}, fmt.Errorf(
			"assetdecode: invalid BIK frame rate %d/%d",
			header.fpsNumerator,
			header.fpsDenominator,
		)
	}

	if header.trackCount > maxBIKAudioTracks {
		return bikHeader{}, fmt.Errorf("assetdecode: too many BIK audio tracks: %d", header.trackCount)
	}

	headerSize := uint64(bikFixedHeaderBytes) +
		uint64(header.trackCount)*bikTrackMetadataBytes +
		uint64(header.frames)*bikFrameIndexBytes
	if headerSize > uint64(len(data)) {
		return bikHeader{}, fmt.Errorf("assetdecode: BIK track/frame index is truncated")
	}

	return header, nil
}

// parseBIKAudioTracks reads the three parallel track tables only after the
// header has proved that all indexed ranges fit within the payload.
func parseBIKAudioTracks(data []byte, trackCount uint32) []BIKAudioTrack {
	tracks := make([]BIKAudioTrack, trackCount)
	packetOffset := bikFixedHeaderBytes
	formatOffset := packetOffset + int(trackCount)*4

	idOffset := formatOffset + int(trackCount)*4
	for index := range tracks {
		flags := binary.LittleEndian.Uint16(data[formatOffset+index*4+2 : formatOffset+index*4+4])
		codec, channels, bits := bikAudioFormat(flags)

		tracks[index] = BIKAudioTrack{
			ID:            binary.LittleEndian.Uint32(data[idOffset+index*4 : idOffset+index*4+4]),
			SampleRate:    binary.LittleEndian.Uint16(data[formatOffset+index*4 : formatOffset+index*4+2]),
			Channels:      channels,
			BitsPerSample: bits,
			Codec:         codec,
			MaxPacketSize: binary.LittleEndian.Uint32(data[packetOffset+index*4 : packetOffset+index*4+4]),
		}
	}

	return tracks
}

// bikAudioFormat translates the packed track flags into the decoder settings
// while retaining Bink's mono, 8-bit, RDFT defaults.
func bikAudioFormat(flags uint16) (codec string, channels, bits int) {
	codec = "rdft"
	if flags&0x1000 != 0 {
		codec = "dct"
	}

	channels = 1
	if flags&0x2000 != 0 {
		channels = 2
	}

	bits = 8
	if flags&0x4000 != 0 {
		bits = 16
	}

	return codec, channels, bits
}

// bikDurationMillis computes the established integer-truncated duration without
// introducing floating-point rounding into metadata or diagnostic output.
func bikDurationMillis(header bikHeader) uint64 {
	return uint64(header.frames) * uint64(header.fpsDenominator) * 1000 / uint64(header.fpsNumerator)
}

// bikUint32 centralizes little-endian header access after the fixed-size header
// check has made every referenced four-byte field safe.
func bikUint32(data []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(data[offset : offset+4])
}
