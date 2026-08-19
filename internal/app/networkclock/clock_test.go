package networkclock

import (
	"math"
	"testing"
	"time"
)

const (
	testSimulationStep = 40 * time.Millisecond
	testUpdateInterval = 100 * time.Millisecond
)

// TestNewNormalizesTimingPolicy verifies the zero-value configuration and the
// dependency order used to derive interpolation and extrapolation bounds.
func TestNewNormalizesTimingPolicy(t *testing.T) {
	clock := New(Config{})
	want := Config{
		UpdateInterval:            100 * time.Millisecond,
		MinInterpolation:          200 * time.Millisecond,
		MaxInterpolation:          500 * time.Millisecond,
		MaxExtrapolation:          250 * time.Millisecond,
		JitterAllowance:           2,
		InterpolationCatchUpRate:  1.25,
		InterpolationSlowDownRate: 0.75,
	}

	if clock.config != want {
		t.Fatalf("normalized config = %#v, want %#v", clock.config, want)
	}
}

// TestClockRejectsSampleWithoutArrivalTime verifies that an unusable timestamp
// cannot make the clock ready or establish an authority anchor.
func TestClockRejectsSampleWithoutArrivalTime(t *testing.T) {
	clock := New(Config{})

	if clock.Observe(Sample{Tick: 10, Step: testSimulationStep}) {
		t.Fatal("sample without arrival time was accepted")
	}

	if timeline := clock.Timeline(time.Unix(50, 0)); timeline.Ready {
		t.Fatalf("clock became ready after rejected sample: %#v", timeline)
	}
}

// TestClockSeparatesPredictionAndInterpolationTimelines verifies that latency
// advances prediction while presentation remains behind by the configured delay.
func TestClockSeparatesPredictionAndInterpolationTimelines(t *testing.T) {
	clock := New(Config{UpdateInterval: testUpdateInterval})
	received := time.Unix(100, 0)
	mustObserve(t, clock, Sample{
		Tick:        100,
		Step:        testSimulationStep,
		ReceivedAt:  received,
		SmoothedRTT: 40 * time.Millisecond,
	})

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

// TestClockExpandsInterpolationDelayForArrivalJitter verifies that irregular
// sample timing becomes bounded presentation slack.
func TestClockExpandsInterpolationDelayForArrivalJitter(t *testing.T) {
	clock := New(Config{UpdateInterval: testUpdateInterval})
	start := time.Unix(200, 0)
	mustObserve(t, clock, Sample{Tick: 100, Step: testSimulationStep, ReceivedAt: start})
	mustObserve(t, clock, Sample{
		Tick:       103,
		Step:       testSimulationStep,
		ReceivedAt: start.Add(100 * time.Millisecond),
	})
	mustObserve(t, clock, Sample{
		Tick:       105,
		Step:       testSimulationStep,
		ReceivedAt: start.Add(250 * time.Millisecond),
	})

	timeline := clock.Timeline(start.Add(250 * time.Millisecond))

	if timeline.Jitter != 50*time.Millisecond {
		t.Fatalf("arrival jitter = %s, want 50ms", timeline.Jitter)
	}

	if timeline.InterpolationDelay != 300*time.Millisecond {
		t.Fatalf("interpolation delay = %s, want 300ms", timeline.InterpolationDelay)
	}
}

// TestClockRejectsReorderedAndMismatchedSamples verifies that only a newer tick
// at the established simulation step may replace the authority anchor.
func TestClockRejectsReorderedAndMismatchedSamples(t *testing.T) {
	clock := New(Config{UpdateInterval: testUpdateInterval})
	start := time.Unix(300, 0)
	mustObserve(t, clock, Sample{Tick: 20, Step: testSimulationStep, ReceivedAt: start})

	if clock.Observe(Sample{Tick: 19, Step: testSimulationStep, ReceivedAt: start.Add(time.Millisecond)}) {
		t.Fatal("reordered sample advanced the clock")
	}

	if clock.Observe(Sample{Tick: 21, Step: 50 * time.Millisecond, ReceivedAt: start.Add(100 * time.Millisecond)}) {
		t.Fatal("simulation step change advanced the clock")
	}

	if got := clock.Timeline(start.Add(100 * time.Millisecond)).LatestServerTick; got != 20 {
		t.Fatalf("latest tick = %d, want 20", got)
	}
}

// TestClockUpdatesTransportStatsFromRejectedSample verifies that stale authority
// data cannot move the anchor but can carry a fresher latency estimate.
func TestClockUpdatesTransportStatsFromRejectedSample(t *testing.T) {
	clock := New(Config{UpdateInterval: testUpdateInterval})
	start := time.Unix(350, 0)
	mustObserve(t, clock, Sample{Tick: 20, Step: testSimulationStep, ReceivedAt: start})

	accepted := clock.Observe(Sample{
		Tick:         20,
		Step:         testSimulationStep,
		ReceivedAt:   start.Add(10 * time.Millisecond),
		SmoothedRTT:  60 * time.Millisecond,
		RTTVariation: 15 * time.Millisecond,
	})
	timeline := clock.Timeline(start.Add(10 * time.Millisecond))

	if accepted {
		t.Fatal("duplicate sample advanced the clock")
	}

	if timeline.LatestServerTick != 20 || timeline.SmoothedRTT != 60*time.Millisecond {
		t.Fatalf("timeline transport update = %#v", timeline)
	}

	if timeline.Jitter != 15*time.Millisecond {
		t.Fatalf("timeline jitter = %s, want 15ms", timeline.Jitter)
	}
}

// TestClockUsesDefaultSimulationStep verifies that incomplete samples retain
// the package's established simulation cadence.
func TestClockUsesDefaultSimulationStep(t *testing.T) {
	clock := New(Config{UpdateInterval: testUpdateInterval})
	start := time.Unix(375, 0)
	mustObserve(t, clock, Sample{Tick: 10, ReceivedAt: start})

	timeline := clock.Timeline(start.Add(testSimulationStep))

	assertMoment(t, timeline.Prediction, 11, 0)
}

// TestClockBoundsExtrapolationWhenCorrectionsStop verifies that a stalled
// authority eventually marks and freezes both client timelines.
func TestClockBoundsExtrapolationWhenCorrectionsStop(t *testing.T) {
	clock := New(Config{UpdateInterval: testUpdateInterval, MaxExtrapolation: 120 * time.Millisecond})
	start := time.Unix(400, 0)
	mustObserve(t, clock, Sample{Tick: 40, Step: testSimulationStep, ReceivedAt: start})

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

// TestClockNeverMovesInterpolationTimeBackward verifies that an increased
// adaptive delay slows presentation without pausing or reversing it.
func TestClockNeverMovesInterpolationTimeBackward(t *testing.T) {
	clock := New(Config{UpdateInterval: testUpdateInterval})
	start := time.Unix(500, 0)
	mustObserve(t, clock, Sample{Tick: 100, Step: testSimulationStep, ReceivedAt: start})

	before := clock.Timeline(start.Add(250 * time.Millisecond))
	mustObserve(t, clock, Sample{
		Tick:       104,
		Step:       testSimulationStep,
		ReceivedAt: start.Add(500 * time.Millisecond),
	})
	after := clock.Timeline(start.Add(500 * time.Millisecond))

	if momentValue(after.Interpolation) <= momentValue(before.Interpolation) {
		t.Fatalf("interpolation paused or reversed from %#v to %#v", before.Interpolation, after.Interpolation)
	}
}

// TestClockSlewsTowardNewInterpolationTargetWithoutVisibleJump verifies that a
// large authority correction respects the configured presentation catch-up cap.
func TestClockSlewsTowardNewInterpolationTargetWithoutVisibleJump(t *testing.T) {
	clock := New(Config{UpdateInterval: testSimulationStep, InterpolationCatchUpRate: 1.25})
	start := time.Unix(600, 0)
	mustObserve(t, clock, Sample{Tick: 10, Step: testSimulationStep, ReceivedAt: start})

	before := clock.Timeline(start)
	mustObserve(t, clock, Sample{
		Tick:       20,
		Step:       testSimulationStep,
		ReceivedAt: start.Add(testSimulationStep),
	})
	after := clock.Timeline(start.Add(testSimulationStep))
	advance := momentValue(after.Interpolation) - momentValue(before.Interpolation)

	if advance > 1.25+1e-9 {
		t.Fatalf("interpolation catch-up advanced %.4f ticks, want <= 1.25", advance)
	}
}

// mustObserve installs fixture data and fails immediately when the scenario
// violates the authoritative sample ordering it intends to establish.
func mustObserve(t *testing.T, clock *Clock, sample Sample) {
	t.Helper()

	if !clock.Observe(sample) {
		t.Fatalf("authoritative sample was rejected: %#v", sample)
	}
}

// assertMoment compares tick positions with tolerance for floating-point
// fractions while keeping failures expressed in the public representation.
func assertMoment(t *testing.T, got Moment, tick uint64, fraction float64) {
	t.Helper()

	if got.Tick != tick || math.Abs(got.Fraction-fraction) > 0.000001 {
		t.Fatalf("moment = %#v, want tick=%d fraction=%f", got, tick, fraction)
	}
}

// momentValue converts a public Moment back to its scalar form for relative
// movement assertions.
func momentValue(value Moment) float64 {
	return float64(value.Tick) + value.Fraction
}
