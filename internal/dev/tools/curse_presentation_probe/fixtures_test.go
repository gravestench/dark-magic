package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// ownedCapture returns the complete two-skill target-band matrix accepted by validation. Each caller owns the result
// and may mutate individual cases without leaking state into another test.
func ownedCapture() capture {
	result := capture{
		Schema: probeSchema,
		Target: probeTarget,
		Source: "owned-runtime",
		Runtime: runtime{
			Patch:               "1.14d",
			Mode:                "expansion",
			Session:             "single-player",
			CharacterOrigin:     "probe-created",
			ExecutableSHA256:    strings.Repeat("a", 64),
			Observation:         "video-frame-log",
			AssetIdentification: "owned-mpq-dcc-comparison",
			CameraFixed:         true,
			ActorsStationary:    true,
		},
	}

	for _, skillID := range []int{66, 72} {
		result.Cases = append(result.Cases, ownedSkillCases(skillID)...)
	}

	return result
}

// ownedSkillCases builds empty, single, and multiple-target observations in coverage order. Keeping the record names
// explicit prevents the fixture from proving validation with the same lookup used by production.
func ownedSkillCases(skillID int) []probeCase {
	skillRecord, areaMissileRecord := "Amplify Damage", "curseamplifydamage"
	if skillID == 72 {
		skillRecord, areaMissileRecord = "Weaken", "curseweaken"
	}

	result := make([]probeCase, 0, 3)
	for _, targetCount := range []int{0, 1, 3} {
		result = append(result, ownedProbeCase(skillID, skillRecord, areaMissileRecord, targetCount))
	}

	return result
}

// ownedProbeCase creates one independently mutable observation with stable coordinates and frame ranges. Those fixed
// values make expected normalized offsets obvious in behavioral assertions.
func ownedProbeCase(skillID int, skillRecord, areaMissileRecord string, targetCount int) probeCase {
	observed := probeCase{
		ID:          probeCaseID(skillID, targetCount),
		SkillID:     skillID,
		SkillRecord: skillRecord,
		Difficulty:  "normal",
		Area:        "blood_moor",
		Caster:      point{X: 100, Y: 200},
		Cursor:      point{X: 300, Y: 220},
		Layers: []layer{
			{
				MissileRecord: areaMissileRecord,
				Present:       true,
				Instances: []instance{
					{
						FirstFrame: 10,
						LastFrame:  29,
						Anchor:     "cursor",
						Start:      point{X: 300, Y: 210},
						End:        point{X: 300, Y: 210},
					},
				},
			},
			{
				MissileRecord: "cursecast",
				Present:       true,
				Instances: []instance{
					{
						FirstFrame: 10,
						LastFrame:  34,
						Anchor:     "caster",
						Start:      point{X: 100, Y: 180},
						End:        point{X: 300, Y: 210},
					},
				},
			},
		},
	}

	for index := range targetCount {
		observed.TargetRecords = append(observed.TargetRecords, "fallen1")
		observed.Targets = append(observed.Targets, point{X: 300 + index*20, Y: 220})
	}

	return observed
}

// probeCaseID gives fixtures the same readable, collision-free identity scheme across skills and target bands.
func probeCaseID(skillID, targetCount int) string {
	return fmt.Sprintf("skill-%d-targets-%d", skillID, targetCount)
}

// marshalCapture serializes a fixture exactly once so analysis and fingerprinting consume identical bytes.
func marshalCapture(t *testing.T, captured capture) string {
	t.Helper()

	data, err := json.Marshal(captured)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

// requireValidationFailure keeps rejection assertions consistent while retaining a caller-specific failure reason.
func requireValidationFailure(t *testing.T, captured capture, acceptedDescription string) {
	t.Helper()

	if err := validate(captured); err == nil {
		t.Fatal(acceptedDescription)
	}
}
