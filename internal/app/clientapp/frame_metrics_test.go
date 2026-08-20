package clientapp

import (
	"testing"
	"time"
)

// TestFrameMetricsReportsScenePercentiles verifies nearest-rank diagnostics keep each work category distinct.
func TestFrameMetricsReportsScenePercentiles(t *testing.T) {
	var metrics frameMetrics
	for index := 1; index <= 100; index++ {
		metrics.Record(
			"game_world",
			time.Duration(index)*time.Millisecond,
			time.Duration(index/2)*time.Millisecond,
			time.Duration(index/4)*time.Millisecond,
			time.Duration(index/5)*time.Millisecond,
		)
	}

	got := metrics.Snapshot()["game_world"]
	if got.Samples != 100 || got.FrameP50 != 50*time.Millisecond ||
		got.FrameP95 != 95*time.Millisecond || got.FrameP99 != 99*time.Millisecond {
		t.Fatalf("frame timing = %#v", got)
	}

	if got.UpdateP95 != 47*time.Millisecond || got.MaxUpdate != 50*time.Millisecond {
		t.Fatalf("update timing = %#v", got)
	}

	if got.SimulationP95 != 23*time.Millisecond || got.LuaP95 != 19*time.Millisecond {
		t.Fatalf("stage timing = %#v", got)
	}
}

// TestFrameMetricsKeepsBoundedRollingWindow protects long-running clients from unbounded diagnostic storage.
func TestFrameMetricsKeepsBoundedRollingWindow(t *testing.T) {
	var metrics frameMetrics
	for index := range frameMetricWindow + 10 {
		metrics.Record("title", time.Duration(index)*time.Millisecond, time.Millisecond, 0, 0)
	}

	if got := metrics.Snapshot()["title"].Samples; got != frameMetricWindow {
		t.Fatalf("samples = %d, want %d", got, frameMetricWindow)
	}
}
