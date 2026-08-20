package assetcatalog

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// addFrames records dimension bounds and an order-sensitive hash of every frame field. The hash preserves direction and
// frame order, offsets, and signed values without retaining decoded pixels or proprietary source bytes.
func (f *AssetFixture) addFrames(frames []Frame) {
	f.FrameCount = len(frames)
	if len(frames) == 0 {
		return
	}

	f.MinWidth, f.MaxWidth = frames[0].Width, frames[0].Width
	f.MinHeight, f.MaxHeight = frames[0].Height, frames[0].Height

	hash := sha256.New()

	var encoded [8]byte

	for _, frame := range frames {
		values := [...]int{
			frame.Direction,
			frame.Frame,
			frame.Width,
			frame.Height,
			frame.OffsetX,
			frame.OffsetY,
		}

		for _, value := range values {
			// Converting through int64 preserves negative offsets as their stable two's-complement representation.
			binary.LittleEndian.PutUint64(encoded[:], uint64(int64(value)))
			_, _ = hash.Write(encoded[:])
		}

		f.MinWidth = min(f.MinWidth, frame.Width)
		f.MaxWidth = max(f.MaxWidth, frame.Width)
		f.MinHeight = min(f.MinHeight, frame.Height)
		f.MaxHeight = max(f.MaxHeight, frame.Height)
	}

	f.FramesHash = hex.EncodeToString(hash.Sum(nil))
}
