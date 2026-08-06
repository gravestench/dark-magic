package rendercore

import (
	"testing"
	"time"
)

func TestAnimationPlayerModesAndControls(t *testing.T) {
	durations := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}
	loop := NewAnimationPlayer(durations, "loop")
	if frame, changed := loop.Advance(31 * time.Millisecond); frame != 2 || !changed {
		t.Fatalf("loop frame = %d, changed = %v", frame, changed)
	}
	loop.SetPaused(true)
	if frame, changed := loop.Advance(time.Second); frame != 2 || changed {
		t.Fatalf("paused frame = %d, changed = %v", frame, changed)
	}
	loop.SetPaused(false)
	if frame, _ := loop.Seek(61 * time.Millisecond); frame != 0 {
		t.Fatalf("loop seek frame = %d", frame)
	}

	once := NewAnimationPlayer(durations, "once")
	if frame, _ := once.Advance(time.Second); frame != 2 {
		t.Fatalf("once frame = %d", frame)
	}

	pingPong := NewAnimationPlayer([]time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}, "ping-pong")
	want := []int{1, 2, 1, 0, 1}
	for index, expected := range want {
		if frame, _ := pingPong.Advance(time.Millisecond); frame != expected {
			t.Fatalf("ping-pong step %d = %d, want %d", index, frame, expected)
		}
	}
}
