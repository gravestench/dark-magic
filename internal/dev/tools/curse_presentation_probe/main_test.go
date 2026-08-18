package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func ownedCapture() capture {
	result := capture{
		Schema: probeSchema, Target: probeTarget, Source: "owned-runtime",
		Runtime: runtime{
			Patch: "1.14d", Mode: "expansion", Session: "single-player", CharacterOrigin: "probe-created",
			ExecutableSHA256: strings.Repeat("a", 64), Observation: "video-frame-log",
			AssetIdentification: "owned-mpq-dcc-comparison", CameraFixed: true, ActorsStationary: true,
		},
	}
	for _, skill := range []int{66, 72} {
		record, areaMissile := "Amplify Damage", "curseamplifydamage"
		if skill == 72 {
			record, areaMissile = "Weaken", "curseweaken"
		}
		for _, count := range []int{0, 1, 3} {
			observed := probeCase{
				ID: fmtID(skill, count), SkillID: skill, SkillRecord: record, Difficulty: "normal", Area: "blood_moor",
				Caster: point{X: 100, Y: 200}, Cursor: point{X: 300, Y: 220},
				Layers: []layer{
					{MissileRecord: areaMissile, Present: true, Instances: []instance{{FirstFrame: 10, LastFrame: 29, Anchor: "cursor", Start: point{X: 300, Y: 210}, End: point{X: 300, Y: 210}}}},
					{MissileRecord: "cursecast", Present: true, Instances: []instance{{FirstFrame: 10, LastFrame: 34, Anchor: "caster", Start: point{X: 100, Y: 180}, End: point{X: 300, Y: 210}}}},
				},
			}
			for index := range count {
				observed.TargetRecords = append(observed.TargetRecords, "fallen1")
				observed.Targets = append(observed.Targets, point{X: 300 + index*20, Y: 220})
			}
			result.Cases = append(result.Cases, observed)
		}
	}
	return result
}

func fmtID(skill, count int) string {
	return fmt.Sprintf("skill-%d-targets-%d", skill, count)
}

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
	if !got.Coverage.Complete || len(got.Coverage.Missing) != 0 || len(got.Cases) != 6 || len(got.CaptureFingerprint) != 64 {
		t.Fatalf("report coverage = %#v cases=%d", got.Coverage, len(got.Cases))
	}
	area, cast := got.Cases[0].Layers[0].Instances[0], got.Cases[0].Layers[1].Instances[0]
	if area.Frames != 20 || area.StartOffset != (point{X: 0, Y: -10}) || area.Translated {
		t.Fatalf("area instance = %#v", area)
	}
	if cast.Frames != 25 || cast.StartOffset != (point{X: 0, Y: -20}) || cast.EndOffset != (point{X: 200, Y: 10}) || !cast.Translated {
		t.Fatalf("cast instance = %#v", cast)
	}
}

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
	if err := validate(captured); err == nil {
		t.Fatal("non-target runtime and invented layer were accepted")
	}
}

func TestValidateRejectsContradictoryPresenceAndTargetAnchor(t *testing.T) {
	captured := ownedCapture()
	index := 0
	captured.Cases[0].Layers[0] = layer{
		MissileRecord: "curseamplifydamage", Present: false,
		Instances: []instance{{FirstFrame: 1, LastFrame: 2, Anchor: "target", TargetIndex: &index}},
	}
	if err := validate(captured); err == nil {
		t.Fatal("contradictory presence and absent target anchor were accepted")
	}
}

func TestCoverageReportsMissingTargetBandsWithoutPromotingRoles(t *testing.T) {
	captured := ownedCapture()
	captured.Cases = captured.Cases[:1]
	got := coverageFor(captured.Cases)
	if got.Complete || len(got.Missing) != 5 {
		t.Fatalf("coverage = %#v", got)
	}
}
