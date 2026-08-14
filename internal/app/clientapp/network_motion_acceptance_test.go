package clientapp

import (
	"sort"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/networkclock"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type scheduledTransform struct {
	arrival time.Time
	tick    uint64
	x       float64
}

// This deterministic harness is deliberately renderer-free. It measures the
// exact network clock and snapshot-buffer path used by live peer transforms
// under ordinary residential-network latency, jitter, loss, and reordering.
func TestRemoteMotionAcceptanceUnderLatencyJitterLossAndReordering(t *testing.T) {
	const (
		step       = 40 * time.Millisecond
		frame      = time.Second / 60
		speed      = 10.0
		totalTicks = 150
	)
	start := time.Unix(1_000, 0)
	jitter := []time.Duration{0, 20 * time.Millisecond, 75 * time.Millisecond, -15 * time.Millisecond, 35 * time.Millisecond, -5 * time.Millisecond}
	deliveries := make([]scheduledTransform, 0, totalTicks)
	for tick := uint64(1); tick <= totalTicks; tick++ {
		if tick%5 == 0 { // deterministic 20% datagram loss
			continue
		}
		sent := start.Add(time.Duration(tick) * step)
		deliveries = append(deliveries, scheduledTransform{
			arrival: sent.Add(80*time.Millisecond + jitter[int(tick)%len(jitter)]),
			tick:    tick, x: float64(tick) * step.Seconds() * speed,
		})
	}
	sort.SliceStable(deliveries, func(i, j int) bool { return deliveries[i].arrival.Before(deliveries[j].arrival) })

	clock := networkclock.New(networkclock.Config{
		UpdateInterval: step, MinInterpolation: 2 * step, MaxInterpolation: 200 * time.Millisecond,
		MaxExtrapolation: 160 * time.Millisecond, JitterAllowance: 2,
	})
	buffer := newPresentationBuffer()
	nextDelivery := 0
	lastX, maxFrameStep := 0.0, 0.0
	sampledFrames, stationaryFrames := 0, 0
	for elapsed := time.Duration(0); elapsed <= 7*time.Second; elapsed += frame {
		now := start.Add(elapsed)
		for nextDelivery < len(deliveries) && !deliveries[nextDelivery].arrival.After(now) {
			packet := deliveries[nextDelivery]
			if clock.Observe(networkclock.Sample{Tick: packet.tick, Step: step, ReceivedAt: packet.arrival, SmoothedRTT: 160 * time.Millisecond, RTTVariation: 12 * time.Millisecond}) {
				buffer.Push(worldView(packet.tick, worldEntity("peer", packet.x, 0)))
			}
			nextDelivery++
		}
		timeline := clock.Timeline(now)
		if !timeline.Ready {
			continue
		}
		sample, found := buffer.Sample(timeline.Interpolation)
		if !found || len(sample.entities) != 1 {
			continue
		}
		x := sample.entities[0].Position.X
		if sampledFrames > 0 {
			delta := x - lastX
			if delta < -1e-9 {
				t.Fatalf("remote presentation moved backward by %.6f at %s", delta, elapsed)
			}
			maxFrameStep = max(maxFrameStep, delta)
			if elapsed > 500*time.Millisecond && elapsed < 6*time.Second && delta < 1e-6 {
				stationaryFrames++
			}
		}
		lastX = x
		sampledFrames++
	}
	if sampledFrames < 300 {
		t.Fatalf("sampled frames = %d, want at least 300", sampledFrames)
	}
	if maxFrameStep > .55 {
		t.Fatalf("maximum 60 Hz presentation step = %.4f subtiles, want <= 0.55", maxFrameStep)
	}
	if stationaryFrames > 10 {
		t.Fatalf("stationary frames during motion = %d, want <= 10", stationaryFrames)
	}
}

func TestRemoteMotionFreezesAfterBoundedOutageInsteadOfRunningAway(t *testing.T) {
	const step = 40 * time.Millisecond
	start := time.Unix(2_000, 0)
	clock := networkclock.New(networkclock.Config{UpdateInterval: step, MaxExtrapolation: 120 * time.Millisecond})
	buffer := newPresentationBuffer()
	for tick := uint64(10); tick <= 12; tick++ {
		received := start.Add(time.Duration(tick-10) * step)
		clock.Observe(networkclock.Sample{Tick: tick, Step: step, ReceivedAt: received})
		buffer.Push(playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: tick, Entities: []playeradapter.WorldEntity{worldEntity("peer", float64(tick), 0)}})
	}
	firstTimeline := clock.Timeline(start.Add(500 * time.Millisecond))
	first, _ := buffer.Sample(firstTimeline.Interpolation)
	secondTimeline := clock.Timeline(start.Add(5 * time.Second))
	second, _ := buffer.Sample(secondTimeline.Interpolation)
	if !firstTimeline.Stale || !secondTimeline.Stale || first.entities[0].Position != second.entities[0].Position {
		t.Fatalf("outage projection kept advancing: first=%#v second=%#v", first, second)
	}
}
