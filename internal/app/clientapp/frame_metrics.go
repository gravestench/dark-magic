package clientapp

import (
	"sort"
	"sync"
	"time"
)

const frameMetricWindow = 512

// sceneFrameSamples is a fixed-size ring so long-running clients cannot grow diagnostics memory without bound.
type sceneFrameSamples struct {
	interval   [frameMetricWindow]time.Duration
	work       [frameMetricWindow]time.Duration
	simulation [frameMetricWindow]time.Duration
	lua        [frameMetricWindow]time.Duration
	count      int
	next       int
	maxWork    time.Duration
}

// frameMetrics serializes render-thread writes and diagnostic snapshots of per-scene timing history.
type frameMetrics struct {
	mu     sync.Mutex
	scenes map[string]*sceneFrameSamples
}

// frameTimingSnapshot is the stable diagnostic schema exposed to capture and profiling tools.
type frameTimingSnapshot struct {
	Samples       int           `json:"samples"`
	FrameP50      time.Duration `json:"frame_p50"`
	FrameP95      time.Duration `json:"frame_p95"`
	FrameP99      time.Duration `json:"frame_p99"`
	UpdateP50     time.Duration `json:"update_p50"`
	UpdateP95     time.Duration `json:"update_p95"`
	UpdateP99     time.Duration `json:"update_p99"`
	MaxUpdate     time.Duration `json:"max_update"`
	SimulationP95 time.Duration `json:"simulation_p95"`
	LuaP95        time.Duration `json:"lua_p95"`
}

// Record appends one frame to its scene ring and preserves the lifetime maximum update cost for spike diagnosis.
func (m *frameMetrics) Record(scene string, interval, work, simulation, luaWork time.Duration) {
	if scene == "" {
		scene = "none"
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scenes == nil {
		m.scenes = make(map[string]*sceneFrameSamples)
	}
	samples := m.scenes[scene]
	if samples == nil {
		samples = &sceneFrameSamples{}
		m.scenes[scene] = samples
	}

	samples.interval[samples.next] = interval
	samples.work[samples.next] = work
	samples.simulation[samples.next] = simulation
	samples.lua[samples.next] = luaWork
	samples.next = (samples.next + 1) % frameMetricWindow
	if samples.count < frameMetricWindow {
		samples.count++
	}

	if work > samples.maxWork {
		samples.maxWork = work
	}
}

// Snapshot copies samples while locked so sorting cannot reorder the live rings observed by future frames.
func (m *frameMetrics) Snapshot() map[string]frameTimingSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]frameTimingSnapshot, len(m.scenes))
	for scene, samples := range m.scenes {
		result[scene] = snapshotFrameSamples(samples)
	}

	return result
}

// snapshotFrameSamples derives percentiles from copies, leaving ring insertion order intact.
func snapshotFrameSamples(samples *sceneFrameSamples) frameTimingSnapshot {
	interval := sortedSampleCopy(samples.interval[:samples.count])
	work := sortedSampleCopy(samples.work[:samples.count])
	simulation := sortedSampleCopy(samples.simulation[:samples.count])
	luaWork := sortedSampleCopy(samples.lua[:samples.count])

	return frameTimingSnapshot{
		Samples:       samples.count,
		FrameP50:      percentile(interval, 50),
		FrameP95:      percentile(interval, 95),
		FrameP99:      percentile(interval, 99),
		UpdateP50:     percentile(work, 50),
		UpdateP95:     percentile(work, 95),
		UpdateP99:     percentile(work, 99),
		MaxUpdate:     samples.maxWork,
		SimulationP95: percentile(simulation, 95),
		LuaP95:        percentile(luaWork, 95),
	}
}

// sortedSampleCopy gives percentile calculation ordered private storage instead of mutating the live ring.
func sortedSampleCopy(samples []time.Duration) []time.Duration {
	result := append([]time.Duration(nil), samples...)
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })

	return result
}

// percentile uses nearest-rank selection so non-empty windows always return an observed duration.
func percentile(sorted []time.Duration, percent int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	index := (len(sorted)*percent + 99) / 100

	return sorted[max(1, index)-1]
}
