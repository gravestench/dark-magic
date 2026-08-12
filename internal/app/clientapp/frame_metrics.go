package clientapp

import (
	"sort"
	"sync"
	"time"
)

const frameMetricWindow = 512

type sceneFrameSamples struct {
	interval [frameMetricWindow]time.Duration
	work     [frameMetricWindow]time.Duration
	count    int
	next     int
	maxWork  time.Duration
}

type frameMetrics struct {
	mu     sync.Mutex
	scenes map[string]*sceneFrameSamples
}

type frameTimingSnapshot struct {
	Samples   int           `json:"samples"`
	FrameP50  time.Duration `json:"frame_p50"`
	FrameP95  time.Duration `json:"frame_p95"`
	FrameP99  time.Duration `json:"frame_p99"`
	UpdateP50 time.Duration `json:"update_p50"`
	UpdateP95 time.Duration `json:"update_p95"`
	UpdateP99 time.Duration `json:"update_p99"`
	MaxUpdate time.Duration `json:"max_update"`
}

func (m *frameMetrics) Record(scene string, interval, work time.Duration) {
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
	samples.next = (samples.next + 1) % frameMetricWindow
	if samples.count < frameMetricWindow {
		samples.count++
	}
	if work > samples.maxWork {
		samples.maxWork = work
	}
}

func (m *frameMetrics) Snapshot() map[string]frameTimingSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]frameTimingSnapshot, len(m.scenes))
	for scene, samples := range m.scenes {
		interval := append([]time.Duration(nil), samples.interval[:samples.count]...)
		work := append([]time.Duration(nil), samples.work[:samples.count]...)
		sort.Slice(interval, func(i, j int) bool { return interval[i] < interval[j] })
		sort.Slice(work, func(i, j int) bool { return work[i] < work[j] })
		result[scene] = frameTimingSnapshot{
			Samples: samples.count, FrameP50: percentile(interval, 50), FrameP95: percentile(interval, 95), FrameP99: percentile(interval, 99),
			UpdateP50: percentile(work, 50), UpdateP95: percentile(work, 95), UpdateP99: percentile(work, 99), MaxUpdate: samples.maxWork,
		}
	}
	return result
}

func percentile(sorted []time.Duration, percent int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	index := (len(sorted)*percent + 99) / 100
	return sorted[max(1, index)-1]
}
