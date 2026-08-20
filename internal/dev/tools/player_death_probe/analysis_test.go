package main

import (
	"reflect"
	"strings"
	"testing"
)

// TestAnalyzeNormalizesOwned114dPlayerDeath protects both provenance fields and every measured death consequence.
func TestAnalyzeNormalizesOwned114dPlayerDeath(t *testing.T) {
	got, err := analyze(strings.NewReader(normalizationCaptureJSON))
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Cases) != 1 || len(got.Cases[0].Deaths) != 1 {
		t.Fatalf("case/death counts = %#v", got.Cases)
	}

	wantDeath := deathReport{
		DeathIndex:                  1,
		DeathAnimationFrames:        20,
		RespawnObserved:             true,
		RespawnInputToControlFrames: 20,
		ExperienceLoss:              1000,
		CarriedGoldLoss:             500,
		GroundGoldAfterRespawn:      500,
		EquipmentRemovedAtRespawn:   1,
		CorpseCountAfterRespawn:     1,
		RecoveryObserved:            true,
		ExperienceRecovered:         750,
		EquipmentRestored:           1,
	}
	if gotDeath := got.Cases[0].Deaths[0]; !reflect.DeepEqual(gotDeath, wantDeath) {
		t.Fatalf("death report = %#v, want %#v", gotDeath, wantDeath)
	}

	if got.ExecutableSHA256 == "" || len(got.CaptureFingerprint) != 64 || !got.Cases[0].StashedGoldUnchanged {
		t.Fatalf("provenance/case = %#v", got)
	}
}

// TestAnalyzeRejectsClassicCommunityServerAndImportedSave ensures provenance failures cannot produce target evidence.
func TestAnalyzeRejectsClassicCommunityServerAndImportedSave(t *testing.T) {
	if _, err := analyze(strings.NewReader(invalidContextCaptureJSON)); err == nil {
		t.Fatal("non-target/community/server/imported-save capture was accepted")
	}
}

// TestAnalyzeRejectsContradictoryTimelineAndStashMutation guards temporal, control, and stash invariants together.
func TestAnalyzeRejectsContradictoryTimelineAndStashMutation(t *testing.T) {
	if _, err := analyze(strings.NewReader(contradictoryTimelineCaptureJSON)); err == nil {
		t.Fatal("contradictory timeline was accepted")
	}
}
