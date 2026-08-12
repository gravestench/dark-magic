package mapgen

import "testing"

func TestPurposeStreamsDoNotPerturbEachOther(t *testing.T) {
	streams := NewStreams(42)
	topology := streams.For("topology")
	want := topology.Uint64()
	_ = streams.For("preset-variant").Uint64()
	if got := streams.For("topology").Uint64(); got != want {
		t.Fatalf("topology stream was perturbed: %d != %d", got, want)
	}
}
