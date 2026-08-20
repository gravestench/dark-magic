package raylibRenderer

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestDecodePCM16PreservesLittleEndianAndIgnoresOddTail protects the wire format and its intentional odd-byte rule.
func TestDecodePCM16PreservesLittleEndianAndIgnoresOddTail(t *testing.T) {
	t.Parallel()

	got := decodePCM16([]byte{0x34, 0x12, 0xfe, 0xff, 0x7f})
	want := []int16{0x1234, -2}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded samples = %v, want %v", got, want)
	}
}

// TestPCMPlaybackQueuesCompleteAndPaddedFinalBlocks verifies the device always receives fixed-size sample blocks.
func TestPCMPlaybackQueuesCompleteAndPaddedFinalBlocks(t *testing.T) {
	t.Parallel()

	channels := 2
	blockSamples := pcmBlockFrames * channels
	playback := pcmPlayback{
		channels: channels,
		pending:  make([]int16, blockSamples+2),
	}
	playback.pending[blockSamples] = 17
	playback.pending[blockSamples+1] = 23

	playback.queueCompleteBlocks()

	if len(playback.queued) != 1 || len(playback.queued[0]) != blockSamples {
		t.Fatalf("complete blocks = %d with first size %d, want 1 block of %d", len(playback.queued),
			len(playback.queued[0]), blockSamples)
	}

	if len(playback.pending) != 2 {
		t.Fatalf("pending samples = %d, want 2", len(playback.pending))
	}

	playback.queueFinalBlock()

	if len(playback.queued) != 2 || len(playback.queued[1]) != blockSamples {
		t.Fatalf("final blocks = %d with final size %d, want 2 blocks with final size %d", len(playback.queued),
			len(playback.queued[1]), blockSamples)
	}

	if playback.queued[1][0] != 17 || playback.queued[1][1] != 23 || playback.queued[1][2] != 0 {
		t.Fatalf("padded final block prefix = %v, want [17 23 0]", playback.queued[1][:3])
	}

	if playback.pending != nil {
		t.Fatalf("pending samples remain after final block: %v", playback.pending)
	}
}

// TestStageMusicDataPreservesExtensionAndPayload verifies that streamed music retains a decoder-compatible suffix and
// byte-for-byte contents while staged in the filesystem.
func TestStageMusicDataPreservesExtensionAndPayload(t *testing.T) {
	path, err := stageMusicData(".wav", []byte("music"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove staged music: %v", err)
		}
	})

	if filepath.Ext(path) != ".wav" {
		t.Fatalf("staged extension = %q", filepath.Ext(path))
	}

	data, err := os.ReadFile(path)
	if err != nil || string(data) != "music" {
		t.Fatalf("staged data = %q, error = %v", data, err)
	}
}
