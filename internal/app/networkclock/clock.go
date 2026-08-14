// Package networkclock maintains the client timelines derived from
// authoritative simulation ticks. It owns timing policy only; snapshot storage
// and entity interpolation belong to the presentation layer.
package networkclock

import (
	"math"
	"sync"
	"time"
)

const defaultStep = 40 * time.Millisecond

type Config struct {
	UpdateInterval            time.Duration
	MinInterpolation          time.Duration
	MaxInterpolation          time.Duration
	MaxExtrapolation          time.Duration
	JitterAllowance           float64
	InterpolationCatchUpRate  float64
	InterpolationSlowDownRate float64
}

type Sample struct {
	Tick         uint64
	Step         time.Duration
	ReceivedAt   time.Time
	SmoothedRTT  time.Duration
	RTTVariation time.Duration
}

// Moment identifies a fixed simulation tick plus the fractional progress to
// the next tick. Fraction is always in [0, 1).
type Moment struct {
	Tick     uint64
	Fraction float64
}

type Timeline struct {
	Ready              bool
	LatestServerTick   uint64
	Prediction         Moment
	Interpolation      Moment
	InterpolationDelay time.Duration
	SmoothedRTT        time.Duration
	Jitter             time.Duration
	SampleAge          time.Duration
	Stale              bool
}

type Clock struct {
	mu sync.Mutex

	config       Config
	ready        bool
	step         time.Duration
	latestTick   uint64
	latestAt     time.Time
	previousTick uint64
	previousAt   time.Time
	rtt          time.Duration
	rttVariation time.Duration
	jitter       time.Duration

	lastQuery         time.Time
	lastPrediction    float64
	lastInterpolation float64
}

func New(config Config) *Clock {
	if config.UpdateInterval <= 0 {
		config.UpdateInterval = 100 * time.Millisecond
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
		config.JitterAllowance = 2
	}
	if config.InterpolationCatchUpRate <= 1 {
		config.InterpolationCatchUpRate = 1.25
	}
	if config.InterpolationSlowDownRate <= 0 || config.InterpolationSlowDownRate >= 1 {
		config.InterpolationSlowDownRate = 0.75
	}
	return &Clock{config: config}
}

// Observe installs a newer authoritative tick sample. Duplicated or reordered
// samples may update transport statistics, but cannot move the clock anchor.
func (clock *Clock) Observe(sample Sample) bool {
	if sample.ReceivedAt.IsZero() {
		return false
	}
	if sample.Step <= 0 {
		sample.Step = defaultStep
	}
	clock.mu.Lock()
	defer clock.mu.Unlock()
	clock.rtt = max(0, sample.SmoothedRTT)
	clock.rttVariation = max(0, sample.RTTVariation)
	if clock.ready && (sample.Tick <= clock.latestTick || sample.Step != clock.step) {
		return false
	}
	if clock.ready {
		clock.previousTick, clock.previousAt = clock.latestTick, clock.latestAt
		clock.observeArrivalJitter(sample.Tick, sample.ReceivedAt)
	}
	clock.ready = true
	clock.step = sample.Step
	clock.latestTick = sample.Tick
	clock.latestAt = sample.ReceivedAt
	return true
}

func (clock *Clock) observeArrivalJitter(tick uint64, receivedAt time.Time) {
	if clock.previousAt.IsZero() || !receivedAt.After(clock.previousAt) || tick <= clock.previousTick {
		return
	}
	tickSpan := time.Duration(tick-clock.previousTick) * clock.step
	periods := max(1, int(math.Round(float64(tickSpan)/float64(clock.config.UpdateInterval))))
	expected := time.Duration(periods) * clock.config.UpdateInterval
	variation := absDuration(receivedAt.Sub(clock.previousAt) - expected)
	if clock.jitter == 0 {
		clock.jitter = variation
		return
	}
	// The same 1/8 EWMA weight used by common transport RTT estimators reacts
	// to persistent jitter without making one delayed packet reshape the clock.
	clock.jitter += (variation - clock.jitter) / 8
}

// Timeline returns monotonically advancing prediction and interpolation time.
// When corrections stop, both timelines advance only through the configured
// extrapolation window and then freeze with Stale set.
func (clock *Clock) Timeline(now time.Time) Timeline {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	if !clock.ready {
		return Timeline{}
	}
	age := now.Sub(clock.latestAt)
	if age < 0 {
		age = 0
	}
	advance := min(age, clock.config.MaxExtrapolation)
	prediction := float64(clock.latestTick) + float64(advance+clock.rtt/2)/float64(clock.step)
	delay := clock.interpolationDelay()
	interpolation := max(0, prediction-float64(delay)/float64(clock.step))
	if clock.lastQuery.IsZero() || !now.Before(clock.lastQuery) {
		if !clock.lastQuery.IsZero() && now.After(clock.lastQuery) {
			elapsedTicks := float64(now.Sub(clock.lastQuery)) / float64(clock.step)
			maximumAdvance := elapsedTicks * clock.config.InterpolationCatchUpRate
			interpolation = min(interpolation, clock.lastInterpolation+maximumAdvance)
			if age <= clock.config.MaxExtrapolation {
				minimumAdvance := elapsedTicks * clock.config.InterpolationSlowDownRate
				interpolation = max(interpolation, min(prediction, clock.lastInterpolation+minimumAdvance))
			}
		}
		prediction = max(prediction, clock.lastPrediction)
		interpolation = max(interpolation, clock.lastInterpolation)
		clock.lastQuery = now
		clock.lastPrediction = prediction
		clock.lastInterpolation = interpolation
	}
	return Timeline{
		Ready: true, LatestServerTick: clock.latestTick,
		Prediction: splitMoment(prediction), Interpolation: splitMoment(interpolation),
		InterpolationDelay: delay, SmoothedRTT: clock.rtt,
		Jitter: max(clock.jitter, clock.rttVariation), SampleAge: age,
		Stale: age > clock.config.MaxExtrapolation,
	}
}

func (clock *Clock) interpolationDelay() time.Duration {
	variation := max(clock.jitter, clock.rttVariation)
	delay := clock.config.MinInterpolation + time.Duration(clock.config.JitterAllowance*float64(variation))
	return min(max(delay, clock.config.MinInterpolation), clock.config.MaxInterpolation)
}

func splitMoment(value float64) Moment {
	if value <= 0 {
		return Moment{}
	}
	whole := math.Floor(value)
	return Moment{Tick: uint64(whole), Fraction: value - whole}
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}
