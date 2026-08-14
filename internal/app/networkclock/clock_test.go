package networkclock

import (
	"math"
	"testing"
	"time"
)

func TestClockSeparatesPredictionAndInterpolationTimelines(t *testing.T) {
	clock := New(Config{UpdateInterval: 100 * time.Millisecond})
	received := time.Unix(100, 0)
	if !clock.Observe(Sample{Tick: 100, Step: 40 * time.Millisecond, ReceivedAt: received, SmoothedRTT: 40 * time.Millisecond}) {
		t.Fatal("initial authoritative sample was rejected")
	}
	timeline := clock.Timeline(received.Add(20 * time.Millisecond))
	if !timeline.Ready || timeline.LatestServerTick != 100 {
		t.Fatalf("timeline readiness = %#v", timeline)
	}
	assertMoment(t, timeline.Prediction, 101, 0)
	assertMoment(t, timeline.Interpolation, 96, 0)
	if timeline.InterpolationDelay != 200*time.Millisecond || timeline.SmoothedRTT != 40*time.Millisecond {
		t.Fatalf("timeline timing = %#v", timeline)
	}
}

func TestClockExpandsInterpolationDelayForArrivalJitter(t *testing.T) {
	clock := New(Config{UpdateInterval: 100 * time.Millisecond})
	start := time.Unix(200, 0)
	clock.Observe(Sample{Tick: 100, Step: 40 * time.Millisecond, ReceivedAt: start})
	clock.Observe(Sample{Tick: 103, Step: 40 * time.Millisecond, ReceivedAt: start.Add(100 * time.Millisecond)})
	clock.Observe(Sample{Tick: 105, Step: 40 * time.Millisecond, ReceivedAt: start.Add(250 * time.Millisecond)})
	timeline := clock.Timeline(start.Add(250 * time.Millisecond))
	if timeline.Jitter != 50*time.Millisecond {
		t.Fatalf("arrival jitter = %s, want 50ms", timeline.Jitter)
	}
	if timeline.InterpolationDelay != 300*time.Millisecond {
		t.Fatalf("interpolation delay = %s, want 300ms", timeline.InterpolationDelay)
	}
}

func TestClockRejectsReorderedAndMismatchedSamples(t *testing.T) {
	clock := New(Config{UpdateInterval: 100 * time.Millisecond})
	start := time.Unix(300, 0)
	clock.Observe(Sample{Tick: 20, Step: 40 * time.Millisecond, ReceivedAt: start})
	if clock.Observe(Sample{Tick: 19, Step: 40 * time.Millisecond, ReceivedAt: start.Add(time.Millisecond)}) {
		t.Fatal("reordered sample advanced the clock")
	}
	if clock.Observe(Sample{Tick: 21, Step: 50 * time.Millisecond, ReceivedAt: start.Add(100 * time.Millisecond)}) {
		t.Fatal("simulation step change advanced the clock")
	}
	if got := clock.Timeline(start.Add(100 * time.Millisecond)).LatestServerTick; got != 20 {
		t.Fatalf("latest tick = %d, want 20", got)
	}
}

func TestClockBoundsExtrapolationWhenCorrectionsStop(t *testing.T) {
	clock := New(Config{UpdateInterval: 100 * time.Millisecond, MaxExtrapolation: 120 * time.Millisecond})
	start := time.Unix(400, 0)
	clock.Observe(Sample{Tick: 40, Step: 40 * time.Millisecond, ReceivedAt: start})
	first := clock.Timeline(start.Add(500 * time.Millisecond))
	second := clock.Timeline(start.Add(2 * time.Second))
	if !first.Stale || !second.Stale || first.SampleAge != 500*time.Millisecond {
		t.Fatalf("stale timelines = first %#v second %#v", first, second)
	}
	if first.Prediction != second.Prediction || first.Interpolation != second.Interpolation {
		t.Fatalf("stale clock kept advancing: first %#v second %#v", first, second)
	}
	assertMoment(t, first.Prediction, 43, 0)
}

func TestClockNeverMovesInterpolationTimeBackward(t *testing.T) {
	clock := New(Config{UpdateInterval: 100 * time.Millisecond})
	start := time.Unix(500, 0)
	clock.Observe(Sample{Tick: 100, Step: 40 * time.Millisecond, ReceivedAt: start})
	before := clock.Timeline(start.Add(250 * time.Millisecond))
	// This delayed correction expands the adaptive interpolation delay. The
	// client slows its interpolation clock, but must neither pause nor reverse.
	clock.Observe(Sample{Tick: 104, Step: 40 * time.Millisecond, ReceivedAt: start.Add(500 * time.Millisecond)})
	after := clock.Timeline(start.Add(500 * time.Millisecond))
	if momentValue(after.Interpolation) <= momentValue(before.Interpolation) {
		t.Fatalf("interpolation paused or reversed from %#v to %#v", before.Interpolation, after.Interpolation)
	}
}

func TestClockSlewsTowardNewInterpolationTargetWithoutVisibleJump(t *testing.T) {
	clock := New(Config{UpdateInterval: 40 * time.Millisecond, InterpolationCatchUpRate: 1.25})
	start := time.Unix(600, 0)
	clock.Observe(Sample{Tick: 10, Step: 40 * time.Millisecond, ReceivedAt: start})
	before := clock.Timeline(start)
	clock.Observe(Sample{Tick: 20, Step: 40 * time.Millisecond, ReceivedAt: start.Add(40 * time.Millisecond)})
	after := clock.Timeline(start.Add(40 * time.Millisecond))
	advance := momentValue(after.Interpolation) - momentValue(before.Interpolation)
	if advance > 1.25+1e-9 {
		t.Fatalf("interpolation catch-up advanced %.4f ticks, want <= 1.25", advance)
	}
}

func assertMoment(t *testing.T, got Moment, tick uint64, fraction float64) {
	t.Helper()
	if got.Tick != tick || math.Abs(got.Fraction-fraction) > 0.000001 {
		t.Fatalf("moment = %#v, want tick=%d fraction=%f", got, tick, fraction)
	}
}

func momentValue(value Moment) float64 { return float64(value.Tick) + value.Fraction }
