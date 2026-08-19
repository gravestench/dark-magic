package clientapp

import (
	"sort"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/networkclock"
)

// scheduledTransform describes one simulated network snapshot arrival.
type scheduledTransform struct {
	arrival time.Time
	tick    uint64
	x       float64
}

// remoteMotionResult summarizes visible behavior from the renderer-free harness.
type remoteMotionResult struct {
	sampledFrames    int
	stationaryFrames int
	maximumFrameStep float64
}

// TestRemoteMotionAcceptanceUnderLatencyJitterLossAndReordering exercises the
// production network clock and snapshot buffer under deterministic WAN noise.
func TestRemoteMotionAcceptanceUnderLatencyJitterLossAndReordering(t *testing.T) {
	result := runRemoteMotionHarness(t)

	if result.sampledFrames < 300 {
		t.Fatalf("sampled frames = %d, want at least 300", result.sampledFrames)
	}

	if result.maximumFrameStep > 0.55 {
		t.Fatalf(
			"maximum 60 Hz presentation step = %.4f subtiles, want <= 0.55",
			result.maximumFrameStep,
		)
	}

	if result.stationaryFrames > 10 {
		t.Fatalf("stationary frames during motion = %d, want <= 10", result.stationaryFrames)
	}
}

// runRemoteMotionHarness samples seven seconds of deterministic peer movement.
func runRemoteMotionHarness(t *testing.T) remoteMotionResult {
	t.Helper()

	const (
		step       = 40 * time.Millisecond
		frame      = time.Second / 60
		totalTicks = 150
	)

	start := time.Unix(1_000, 0)
	deliveries := scheduleRemoteTransforms(start, step, totalTicks)
	clock := remoteMotionClock(step)
	buffer := newPresentationBuffer()
	nextDelivery := 0
	lastX := 0.0
	result := remoteMotionResult{}

	for elapsed := time.Duration(0); elapsed <= 7*time.Second; elapsed += frame {
		now := start.Add(elapsed)
		nextDelivery = deliverReadyTransforms(deliveries, nextDelivery, now, step, clock, buffer)
		sampleRemoteFrame(t, clock, buffer, now, elapsed, &lastX, &result)
	}

	return result
}

// scheduleRemoteTransforms applies deterministic latency, jitter, loss, and reordering.
func scheduleRemoteTransforms(
	start time.Time,
	step time.Duration,
	totalTicks uint64,
) []scheduledTransform {
	jitter := []time.Duration{
		0,
		20 * time.Millisecond,
		75 * time.Millisecond,
		-15 * time.Millisecond,
		35 * time.Millisecond,
		-5 * time.Millisecond,
	}
	deliveries := make([]scheduledTransform, 0, totalTicks)

	for tick := uint64(1); tick <= totalTicks; tick++ {
		if tick%5 == 0 { // Deterministic 20% datagram loss.
			continue
		}

		sent := start.Add(time.Duration(tick) * step)
		deliveries = append(deliveries, scheduledTransform{
			arrival: sent.Add(80*time.Millisecond + jitter[int(tick)%len(jitter)]),
			tick:    tick,
			x:       float64(tick) * step.Seconds() * 10,
		})
	}

	sort.SliceStable(deliveries, func(i, j int) bool {
		return deliveries[i].arrival.Before(deliveries[j].arrival)
	})

	return deliveries
}

// remoteMotionClock returns the live interpolation policy used by the harness.
func remoteMotionClock(step time.Duration) *networkclock.Clock {
	return networkclock.New(networkclock.Config{
		UpdateInterval:   step,
		MinInterpolation: 2 * step,
		MaxInterpolation: 200 * time.Millisecond,
		MaxExtrapolation: 160 * time.Millisecond,
		JitterAllowance:  2,
	})
}

// deliverReadyTransforms feeds every snapshot that has arrived by this frame.
func deliverReadyTransforms(
	deliveries []scheduledTransform,
	next int,
	now time.Time,
	step time.Duration,
	clock *networkclock.Clock,
	buffer *presentationBuffer,
) int {
	for next < len(deliveries) && !deliveries[next].arrival.After(now) {
		packet := deliveries[next]
		sample := networkclock.Sample{
			Tick:         packet.tick,
			Step:         step,
			ReceivedAt:   packet.arrival,
			SmoothedRTT:  160 * time.Millisecond,
			RTTVariation: 12 * time.Millisecond,
		}

		if clock.Observe(sample) {
			buffer.Push(worldView(packet.tick, worldEntity("peer", packet.x, 0)))
		}

		next++
	}

	return next
}

// sampleRemoteFrame records visible progress for one ready interpolation frame.
func sampleRemoteFrame(
	t *testing.T,
	clock *networkclock.Clock,
	buffer *presentationBuffer,
	now time.Time,
	elapsed time.Duration,
	lastX *float64,
	result *remoteMotionResult,
) {
	t.Helper()

	timeline := clock.Timeline(now)
	if !timeline.Ready {
		return
	}

	sample, found := buffer.Sample(timeline.Interpolation)
	if !found || len(sample.entities) != 1 {
		return
	}

	x := sample.entities[0].Position.X
	if result.sampledFrames > 0 {
		recordRemoteFrameDelta(t, x-*lastX, elapsed, result)
	}

	*lastX = x
	result.sampledFrames++
}

// recordRemoteFrameDelta validates direction and accumulates smoothness metrics.
func recordRemoteFrameDelta(
	t *testing.T,
	delta float64,
	elapsed time.Duration,
	result *remoteMotionResult,
) {
	t.Helper()

	if delta < -1e-9 {
		t.Fatalf("remote presentation moved backward by %.6f at %s", delta, elapsed)
	}

	result.maximumFrameStep = max(result.maximumFrameStep, delta)

	duringMotion := elapsed > 500*time.Millisecond && elapsed < 6*time.Second
	if duringMotion && delta < 1e-6 {
		result.stationaryFrames++
	}
}

// TestRemoteMotionFreezesAfterBoundedOutageInsteadOfRunningAway verifies that
// extrapolation stops after its configured window during a prolonged outage.
func TestRemoteMotionFreezesAfterBoundedOutageInsteadOfRunningAway(t *testing.T) {
	const step = 40 * time.Millisecond

	start := time.Unix(2_000, 0)
	clock := networkclock.New(networkclock.Config{
		UpdateInterval:   step,
		MaxExtrapolation: 120 * time.Millisecond,
	})
	buffer := newPresentationBuffer()

	for tick := uint64(10); tick <= 12; tick++ {
		received := start.Add(time.Duration(tick-10) * step)
		clock.Observe(networkclock.Sample{Tick: tick, Step: step, ReceivedAt: received})
		buffer.Push(worldView(tick, worldEntity("peer", float64(tick), 0)))
	}

	firstTimeline := clock.Timeline(start.Add(500 * time.Millisecond))
	first, _ := buffer.Sample(firstTimeline.Interpolation)
	secondTimeline := clock.Timeline(start.Add(5 * time.Second))
	second, _ := buffer.Sample(secondTimeline.Interpolation)

	if !firstTimeline.Stale || !secondTimeline.Stale || first.entities[0].Position != second.entities[0].Position {
		t.Fatalf("outage projection kept advancing: first=%#v second=%#v", first, second)
	}
}
