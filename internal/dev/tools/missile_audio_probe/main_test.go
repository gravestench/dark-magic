package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// ownedCapture builds the complete supported matrix so validation tests start
// from evidence that is internally consistent and attributable to the target runtime.
func ownedCapture() capture {
	result := capture{
		Schema: probeSchema, Target: probeTarget, Source: "owned-runtime",
		Runtime: runtime{
			Patch: "1.14d", Mode: "expansion", Session: "single-player", CharacterOrigin: "probe-created",
			ExecutableSHA256: strings.Repeat("a", 64), RecordGenerationID: "sha256:" + strings.Repeat("b", 64),
			Observation: "isolated-audio-video-frame-log", SoundIdentification: "owned-mpq-waveform-comparison",
			GameFramesPerSecond: 25, AudioIsolated: true, CameraFixed: true, ActorsStationary: true,
		},
	}

	for _, spec := range requiredCases {
		observed := probeCase{
			ID: spec.id, SkillID: spec.skillID, SkillRecord: spec.skill, SkillLevel: 1, MissileRecord: spec.missile,
			Difficulty: "normal", Area: "blood_moor", Outcome: spec.outcome, TargetCount: spec.targets,
			ProjectileVisualCount: 1,
			CastEffectFrame:       100, MissileRemovedFrame: 130,
		}
		for range spec.targets {
			observed.TargetRecords = append(observed.TargetRecords, "fallen1")
		}

		if spec.outcome != "expired" {
			contact := 120
			observed.ContactFrame = &contact
		}

		for _, expected := range spec.sounds {
			sound := soundObservation{Record: expected.record, Role: expected.role}
			if expected.role == "travel" {
				sound.Present = true
				sound.Intervals = []frameInterval{{FirstFrame: 100, LastFrame: 130}}
			} else if observed.ContactFrame != nil {
				sound.Present = true
				sound.Intervals = []frameInterval{{FirstFrame: 120, LastFrame: 125}}
			}

			observed.Sounds = append(observed.Sounds, sound)
		}

		result.Cases = append(result.Cases, observed)
	}

	return result
}

// TestAnalyzeNormalizesCompleteOwned114dMatrix protects report ordering,
// fingerprints, and frame-relative timing derived from a complete capture.
func TestAnalyzeNormalizesCompleteOwned114dMatrix(t *testing.T) {
	captured := ownedCapture()

	data, err := json.Marshal(captured)
	if err != nil {
		t.Fatal(err)
	}

	got, err := analyze(strings.NewReader(string(data)))
	if err != nil {
		t.Fatal(err)
	}

	if !got.Coverage.Complete || len(got.Coverage.Missing) != 0 || len(got.Cases) != len(requiredCases) ||
		len(got.CaptureFingerprint) != 64 {
		t.Fatalf("report coverage=%#v cases=%d fingerprint=%q", got.Coverage, len(got.Cases), got.CaptureFingerprint)
	}

	var fireBolt caseReport

	for _, observed := range got.Cases {
		if observed.ID == "fire-bolt-hit" {
			fireBolt = observed
		}
	}

	validFireBoltTiming := fireBolt.LifetimeFrames == 30 && fireBolt.ContactFromEffect != nil &&
		*fireBolt.ContactFromEffect == 20
	if !validFireBoltTiming || len(fireBolt.Sounds) != 2 {
		t.Fatalf("normalized fire-bolt hit = %#v", fireBolt)
	}

	for _, sound := range fireBolt.Sounds {
		if sound.Role == "travel" {
			validTiming := sound.RecordLoop && sound.Intervals[0].FirstFromEffect == 0 &&
				sound.Intervals[0].LastFromRemoval == 0
			if !validTiming {
				t.Fatalf("normalized travel sound = %#v", sound)
			}
		}

		if sound.Role == "hit" {
			validTiming := !sound.RecordLoop && sound.Intervals[0].FirstFromContact != nil &&
				*sound.Intervals[0].FirstFromContact == 0
			if !validTiming {
				t.Fatalf("normalized hit sound = %#v", sound)
			}
		}
	}
}

// TestValidateRejectsUnsupportedRuntimeAndMismatchedMatrixRow ensures community
// or non-target observations cannot be promoted as owned 1.14d evidence.
func TestValidateRejectsUnsupportedRuntimeAndMismatchedMatrixRow(t *testing.T) {
	captured := ownedCapture()
	captured.Target = "classic"
	captured.Source = "community-tool"
	captured.Runtime.Patch = "1.13c"
	captured.Runtime.Mode = "classic"
	captured.Runtime.Session = "vanilla-server"
	captured.Runtime.CharacterOrigin = "imported-save"
	captured.Runtime.ExecutableSHA256 = "bad"
	captured.Runtime.RecordGenerationID = "mutable"
	captured.Runtime.Observation = "memory-tool"
	captured.Runtime.SoundIdentification = "community-table"
	captured.Runtime.AudioIsolated = false
	captured.Runtime.CameraFixed = false
	captured.Runtime.ActorsStationary = false

	captured.Cases[0].SkillID = 999
	if err := validate(captured); err == nil {
		t.Fatal("unsupported runtime and mismatched matrix row were accepted")
	}
}

// TestValidateRejectsContradictorySoundAndContactTimelines prevents impossible
// frame relationships from entering normalized timing reports.
func TestValidateRejectsContradictorySoundAndContactTimelines(t *testing.T) {
	captured := ownedCapture()
	captured.Cases[0].Sounds[0].Present = false
	captured.Cases[0].Sounds[0].Intervals = []frameInterval{{FirstFrame: 10, LastFrame: 20}}
	contact := 90

	captured.Cases[1].ContactFrame = &contact
	if err := validate(captured); err == nil {
		t.Fatal("contradictory sound presence and out-of-lifetime contact were accepted")
	}
}

// TestValidateRejectsPreCastOverlappingAndInventedSoundObservations protects the
// authored sound matrix from observations that could not come from the capture.
func TestValidateRejectsPreCastOverlappingAndInventedSoundObservations(t *testing.T) {
	captured := ownedCapture()
	captured.Cases[0].Sounds[0].Intervals = []frameInterval{
		{FirstFrame: 99, LastFrame: 110},
		{FirstFrame: 110, LastFrame: 120},
	}

	captured.Cases[1].Sounds = append(captured.Cases[1].Sounds, soundObservation{
		Record: "invented_sound", Role: "hit", Present: true,
		Intervals: []frameInterval{{FirstFrame: 120, LastFrame: 121}},
	})
	if err := validate(captured); err == nil {
		t.Fatal("pre-cast, overlapping, and invented sound observations were accepted")
	}
}

// TestCoverageReportsMissingCasesWithoutPromotingTiming keeps incomplete
// evidence visible without accidentally marking the matrix complete.
func TestCoverageReportsMissingCasesWithoutPromotingTiming(t *testing.T) {
	captured := ownedCapture()
	captured.Cases = captured.Cases[:1]

	got := coverageFor(captured.Cases)
	if got.Complete || len(got.Missing) != len(requiredCases)-1 {
		t.Fatalf("coverage = %#v", got)
	}
}
