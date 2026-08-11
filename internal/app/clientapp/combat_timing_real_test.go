package clientapp

import (
	"os"
	"testing"

	assetdecode "github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/content"
)

// TestLegacyUnarmedAttackTimings is an opt-in production-asset acceptance
// probe. Normal CI has no proprietary MPQs and skips it; local milestone work
// runs it with MPQ_DIRECTORY to prove every playable class has an authored A1
// impact event before the adapter is changed.
func TestLegacyUnarmedAttackTimings(t *testing.T) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	data, err := assetdecode.AnimationData(assets, "data/global/AnimData.d2")
	if err != nil {
		t.Fatal(err)
	}
	adapter := newCombatTimingAdapter(data)
	for _, token := range []string{"AM", "SO", "NE", "PA", "BA", "DZ", "AI"} {
		t.Run(token, func(t *testing.T) {
			timing, found := adapter.AttackTiming(token, "HTH")
			if !found {
				t.Fatalf("%sA1HTH has no valid attack timing", token)
			}
			t.Logf("%sA1HTH frames=%d speed=%d impact=%d", token, timing.Frames, timing.Speed, timing.ImpactFrame)
		})
	}
}
