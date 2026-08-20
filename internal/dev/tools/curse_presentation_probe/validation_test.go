package main

import "testing"

// TestValidateRejectsNonTargetRuntimeAndInventedMissile verifies that provenance failures and unsupported visual
// records cannot enter normalization as owned 1.14d evidence.
func TestValidateRejectsNonTargetRuntimeAndInventedMissile(t *testing.T) {
	captured := ownedCapture()

	captured.Target = "classic"
	captured.Source = "community-tool"
	captured.Runtime.Patch = "1.13c"
	captured.Runtime.Mode = "classic"
	captured.Runtime.Session = "vanilla-server"
	captured.Runtime.CharacterOrigin = "imported-save"
	captured.Runtime.ExecutableSHA256 = "bad"
	captured.Runtime.Observation = "memory-tool"
	captured.Runtime.AssetIdentification = "community-tool"
	captured.Runtime.CameraFixed = false
	captured.Cases[0].Layers = append(captured.Cases[0].Layers, layer{MissileRecord: "invented"})

	requireValidationFailure(t, captured, "non-target runtime and invented layer were accepted")
}

// TestValidateRejectsContradictoryPresenceAndTargetAnchor verifies that a layer cannot claim absence while supplying
// instances, even when an instance also refers to a target missing from the observation.
func TestValidateRejectsContradictoryPresenceAndTargetAnchor(t *testing.T) {
	captured := ownedCapture()
	targetIndex := 0
	captured.Cases[0].Layers[0] = layer{
		MissileRecord: "curseamplifydamage",
		Present:       false,
		Instances: []instance{
			{
				FirstFrame:  1,
				LastFrame:   2,
				Anchor:      "target",
				TargetIndex: &targetIndex,
			},
		},
	}

	requireValidationFailure(t, captured, "contradictory presence and absent target anchor were accepted")
}
