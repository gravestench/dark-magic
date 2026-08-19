package networkclock

import "time"

// Config controls how Clock converts authoritative samples into client-facing
// prediction and interpolation timelines. New supplies safe defaults for
// missing or invalid values so callers can configure only the policies they
// need to override.
type Config struct {
	// UpdateInterval is the expected cadence of authoritative samples.
	UpdateInterval time.Duration
	// MinInterpolation is the smallest amount by which presentation trails prediction.
	MinInterpolation time.Duration
	// MaxInterpolation caps the adaptive presentation delay.
	MaxInterpolation time.Duration
	// MaxExtrapolation limits how long timelines advance without a new authoritative sample.
	MaxExtrapolation time.Duration
	// JitterAllowance scales observed timing variation into additional interpolation delay.
	JitterAllowance float64
	// InterpolationCatchUpRate caps forward slew as a multiple of real-time tick progress.
	InterpolationCatchUpRate float64
	// InterpolationSlowDownRate preserves forward progress while presentation absorbs added delay.
	InterpolationSlowDownRate float64
}

// Sample combines an authoritative simulation position with the transport
// timing observed when it arrived. ReceivedAt must be non-zero; Step defaults
// to the package simulation step when it is not positive.
type Sample struct {
	// Tick is the authority's completed simulation tick.
	Tick uint64
	// Step is the authority's fixed simulation duration.
	Step time.Duration
	// ReceivedAt records when the client observed this authority state.
	ReceivedAt time.Time
	// SmoothedRTT is the transport's current round-trip estimate.
	SmoothedRTT time.Duration
	// RTTVariation is the transport's current round-trip variation estimate.
	RTTVariation time.Duration
}

// Moment identifies a fixed simulation tick plus the fractional progress to
// the next tick. Fraction is always in [0, 1).
type Moment struct {
	// Tick is the completed simulation tick.
	Tick uint64
	// Fraction is progress from Tick toward the next simulation tick.
	Fraction float64
}

// Timeline is a detached value view of the client timing state at one query
// time. Prediction leads presentation, while Interpolation intentionally trails
// it to absorb network variation.
type Timeline struct {
	// Ready reports whether the clock has accepted an authoritative sample.
	Ready bool
	// LatestServerTick is the most recent accepted authority anchor.
	LatestServerTick uint64
	// Prediction is the client's latency-adjusted simulation position.
	Prediction Moment
	// Interpolation is the intentionally delayed presentation position.
	Interpolation Moment
	// InterpolationDelay is the delay used to derive Interpolation from Prediction.
	InterpolationDelay time.Duration
	// SmoothedRTT is the latest non-negative round-trip estimate.
	SmoothedRTT time.Duration
	// Jitter is the larger of arrival-gap and transport RTT variation.
	Jitter time.Duration
	// SampleAge is elapsed query time since the latest accepted anchor.
	SampleAge time.Duration
	// Stale reports whether SampleAge exceeds the extrapolation window.
	Stale bool
}
