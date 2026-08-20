// Package networkclock maintains the client timelines derived from
// authoritative simulation ticks. It owns timing policy only; snapshot storage
// and entity interpolation belong to the presentation layer.
package networkclock

import (
	"math"
	"sync"
	"time"
)

const (
	defaultSimulationStep        = 40 * time.Millisecond
	defaultUpdateInterval        = 100 * time.Millisecond
	defaultJitterAllowance       = 2
	defaultInterpolationCatchUp  = 1.25
	defaultInterpolationSlowDown = 0.75
)

// sampleAnchor keeps a simulation tick and its local arrival time inseparable
// when authoritative history advances.
type sampleAnchor struct {
	tick       uint64
	receivedAt time.Time
}

// transportTiming groups the latency inputs that determine prediction lead and
// interpolation slack.
type transportTiming struct {
	roundTripTime      time.Duration
	roundTripVariation time.Duration
	arrivalJitter      time.Duration
}

// timelinePosition carries scalar tick positions through derivation and slew
// limiting before they become public Moments.
type timelinePosition struct {
	prediction    float64
	interpolation float64
	delay         time.Duration
}

// timelineProgress records the most recent chronological query so subsequent
// queries cannot create a visible reversal or correction jump.
type timelineProgress struct {
	queriedAt     time.Time
	prediction    float64
	interpolation float64
}

// Clock derives prediction and presentation time from authoritative samples.
// Its mutex keeps observation and query state ordered when reliable snapshots
// and low-latency datagrams arrive on different goroutines.
type Clock struct {
	mu sync.Mutex

	config    Config
	ready     bool
	step      time.Duration
	latest    sampleAnchor
	previous  sampleAnchor
	transport transportTiming
	progress  timelineProgress
}

// New creates an empty Clock with timing policies normalized before any
// concurrent observations can begin.
func New(config Config) *Clock {
	return &Clock{config: normalizedConfig(config)}
}

// normalizedConfig fills invalid policy values in dependency order so derived
// bounds always use the effective update interval and minimum delay.
func normalizedConfig(config Config) Config {
	if config.UpdateInterval <= 0 {
		config.UpdateInterval = defaultUpdateInterval
	}

	if config.MinInterpolation <= 0 {
		config.MinInterpolation = 2 * config.UpdateInterval
	}

	if config.MaxInterpolation < config.MinInterpolation {
		config.MaxInterpolation = max(5*config.UpdateInterval, config.MinInterpolation)
	}

	if config.MaxExtrapolation <= 0 {
		config.MaxExtrapolation = max(250*time.Millisecond, 2*config.UpdateInterval)
	}

	if config.JitterAllowance <= 0 {
		config.JitterAllowance = defaultJitterAllowance
	}

	if config.InterpolationCatchUpRate <= 1 {
		config.InterpolationCatchUpRate = defaultInterpolationCatchUp
	}

	if config.InterpolationSlowDownRate <= 0 || config.InterpolationSlowDownRate >= 1 {
		config.InterpolationSlowDownRate = defaultInterpolationSlowDown
	}

	return config
}

// Observe installs a newer authoritative tick sample. Duplicated or reordered
// samples may update transport statistics, but cannot move the clock anchor.
func (clock *Clock) Observe(sample Sample) bool {
	if sample.ReceivedAt.IsZero() {
		return false
	}

	if sample.Step <= 0 {
		sample.Step = defaultSimulationStep
	}

	clock.mu.Lock()
	defer clock.mu.Unlock()

	// Even a stale authority sample carries the transport's newest latency
	// estimate, so update these values before deciding whether to move the anchor.
	clock.transport.roundTripTime = max(0, sample.SmoothedRTT)

	clock.transport.roundTripVariation = max(0, sample.RTTVariation)
	if !clock.acceptsSample(sample) {
		return false
	}

	clock.installSample(sample)

	return true
}

// acceptsSample reports whether sample may advance the authoritative anchor. The
// caller holds mu so a concurrent observation cannot invalidate the decision.
func (clock *Clock) acceptsSample(sample Sample) bool {
	return !clock.ready || (sample.Tick > clock.latest.tick && sample.Step == clock.step)
}

// installSample advances the authoritative anchor and incorporates the arrival gap
// into jitter before replacing the prior sample. The caller holds mu.
func (clock *Clock) installSample(sample Sample) {
	if clock.ready {
		clock.previous = clock.latest
		clock.observeArrivalJitter(sample)
	}

	clock.ready = true
	clock.step = sample.Step
	clock.latest = sampleAnchor{tick: sample.Tick, receivedAt: sample.ReceivedAt}
}

// observeArrivalJitter updates the arrival-gap EWMA only for samples ordered in
// both simulation and wall-clock time. The caller holds mu and has preserved
// the former latest sample in previous.
func (clock *Clock) observeArrivalJitter(sample Sample) {
	if clock.previous.receivedAt.IsZero() ||
		!sample.ReceivedAt.After(clock.previous.receivedAt) ||
		sample.Tick <= clock.previous.tick {
		return
	}

	tickSpan := time.Duration(sample.Tick-clock.previous.tick) * clock.step
	periods := max(1, int(math.Round(float64(tickSpan)/float64(clock.config.UpdateInterval))))
	expected := time.Duration(periods) * clock.config.UpdateInterval

	variation := absDuration(sample.ReceivedAt.Sub(clock.previous.receivedAt) - expected)
	if clock.transport.arrivalJitter == 0 {
		clock.transport.arrivalJitter = variation

		return
	}

	// The same 1/8 EWMA weight used by common transport RTT estimators reacts
	// to persistent jitter without making one delayed packet reshape the clock.
	clock.transport.arrivalJitter += (variation - clock.transport.arrivalJitter) / 8
}

// Timeline returns prediction and interpolation time that advances monotonically
// across chronological queries. When corrections stop, both timelines advance
// only through the configured extrapolation window and then freeze with Stale set.
func (clock *Clock) Timeline(now time.Time) Timeline {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	if !clock.ready {
		return Timeline{}
	}

	age := now.Sub(clock.latest.receivedAt)
	if age < 0 {
		age = 0
	}

	position := clock.targetPosition(age)
	position = clock.applySlewLimits(now, age, position)

	return clock.timelineSnapshot(age, position)
}

// targetPosition derives the unconstrained client timelines for a sample age.
// Prediction includes half the RTT, while interpolation trails by the adaptive
// presentation delay.
func (clock *Clock) targetPosition(age time.Duration) timelinePosition {
	advance := min(age, clock.config.MaxExtrapolation)
	prediction := float64(clock.latest.tick) +
		float64(advance+clock.transport.roundTripTime/2)/float64(clock.step)
	delay := clock.interpolationDelay()
	interpolation := max(0, prediction-float64(delay)/float64(clock.step))

	return timelinePosition{
		prediction:    prediction,
		interpolation: interpolation,
		delay:         delay,
	}
}

// applySlewLimits prevents chronological queries from exposing prediction or
// interpolation that moves backward or jumps toward a changed target. Queries
// older than the recorded progress are derived without mutating that progress.
func (clock *Clock) applySlewLimits(
	now time.Time,
	age time.Duration,
	position timelinePosition,
) timelinePosition {
	if !clock.progress.queriedAt.IsZero() && now.Before(clock.progress.queriedAt) {
		return position
	}

	if !clock.progress.queriedAt.IsZero() && now.After(clock.progress.queriedAt) {
		elapsedTicks := float64(now.Sub(clock.progress.queriedAt)) / float64(clock.step)
		maximumAdvance := elapsedTicks * clock.config.InterpolationCatchUpRate
		position.interpolation = min(
			position.interpolation,
			clock.progress.interpolation+maximumAdvance,
		)

		if age <= clock.config.MaxExtrapolation {
			minimumAdvance := elapsedTicks * clock.config.InterpolationSlowDownRate
			position.interpolation = max(
				position.interpolation,
				min(position.prediction, clock.progress.interpolation+minimumAdvance),
			)
		}
	}

	position.prediction = max(position.prediction, clock.progress.prediction)
	position.interpolation = max(position.interpolation, clock.progress.interpolation)
	clock.progress = timelineProgress{
		queriedAt:     now,
		prediction:    position.prediction,
		interpolation: position.interpolation,
	}

	return position
}

// timelineSnapshot converts internal floating-point tick positions into the
// detached public value returned to callers. The caller holds mu so every field
// reflects one consistent observation state.
func (clock *Clock) timelineSnapshot(age time.Duration, position timelinePosition) Timeline {
	return Timeline{
		Ready:              true,
		LatestServerTick:   clock.latest.tick,
		Prediction:         splitMoment(position.prediction),
		Interpolation:      splitMoment(position.interpolation),
		InterpolationDelay: position.delay,
		SmoothedRTT:        clock.transport.roundTripTime,
		Jitter:             max(clock.transport.arrivalJitter, clock.transport.roundTripVariation),
		SampleAge:          age,
		Stale:              age > clock.config.MaxExtrapolation,
	}
}

// interpolationDelay converts the larger of transport and arrival variation
// into presentation slack, bounded by the configured interpolation window.
func (clock *Clock) interpolationDelay() time.Duration {
	variation := max(clock.transport.arrivalJitter, clock.transport.roundTripVariation)
	delay := clock.config.MinInterpolation + time.Duration(clock.config.JitterAllowance*float64(variation))

	return min(max(delay, clock.config.MinInterpolation), clock.config.MaxInterpolation)
}

// splitMoment converts a floating-point tick position into the public fixed
// tick and fractional representation, clamping non-positive positions to zero.
func splitMoment(value float64) Moment {
	if value <= 0 {
		return Moment{}
	}

	whole := math.Floor(value)

	return Moment{Tick: uint64(whole), Fraction: value - whole}
}

// absDuration returns the magnitude of a duration so early and late arrivals
// contribute equally to the jitter estimate.
func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}

	return value
}
