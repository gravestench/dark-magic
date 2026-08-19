package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// ownedCapture builds the complete target matrix so tests share one valid provenance and coverage baseline.
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
			RecordGenerationID:  "sha256:" + strings.Repeat("b", 64),
			Observation:         "video-frame-log",
			StatIdentification:  "owned-itemstatcost-properties-tbl",
			Locale:              "eng",
			GameFramesPerSecond: 25,
		},
	}

	for _, rate := range []int{0, 8, 9, 19, 20, 36, 37, 62, 63, 104, 105, 199, 200} {
		appendOwnedProfile(&result, 36, "HTH", rate)
	}

	for _, weapon := range []string{"1HS", "STF"} {
		for _, rate := range []int{0, 105} {
			appendOwnedProfile(&result, 36, weapon, rate)
		}
	}

	for _, rate := range []int{0, 105} {
		appendOwnedProfile(&result, 49, "HTH", rate)
	}

	return result
}

// appendOwnedProfile adds one internally consistent observation, including item provenance for every nonzero rate.
func appendOwnedProfile(captured *capture, skillID int, weapon string, rate int) {
	record, mode, sequence, _ := expectedSkill(skillID)

	observed := probeCase{
		ID:                 fmt.Sprintf("skill-%d-%s-%d", skillID, strings.ToLower(weapon), rate),
		SkillID:            skillID,
		SkillRecord:        record,
		CharacterClass:     "sor",
		AnimationMode:      mode,
		SequenceTransition: "SC",
		SequenceNumber:     sequence,
		WeaponClass:        weapon,
		RawFasterCastRate:  rate,
		ModifierKey:        "ModStr4v",
		ModifierText:       "Faster Cast Rate",
		StartFrame:         100,
		EffectFrame:        107,
		NeutralFrame:       114,
	}
	if rate > 0 {
		observed.ModifierSources = []modifierSource{
			{
				ItemRecord:   "probe-equipped-item",
				PropertyCode: "cast1",
				Value:        rate,
			},
		}
	}

	captured.Cases = append(captured.Cases, observed)
}

// TestAnalyzeNormalizesCompleteOwned114dMatrix verifies complete evidence yields sorted delays and a fingerprint.
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

	if !got.Coverage.Complete ||
		len(got.Coverage.Missing) != 0 ||
		len(got.Profiles) != 19 ||
		len(got.CaptureFingerprint) != 64 {
		t.Fatalf("report coverage=%#v profiles=%d", got.Coverage, len(got.Profiles))
	}

	if got.Profiles[0].SkillID != 36 ||
		got.Profiles[0].WeaponClass != "1HS" ||
		got.Profiles[0].EffectDelay != 7 ||
		got.Profiles[0].CompletionDelay != 14 {
		t.Fatalf("first normalized profile = %#v", got.Profiles[0])
	}
}

// TestValidateRejectsNonTargetRuntimeAndUnownedStatIdentification protects the owned-runtime trust boundary.
func TestValidateRejectsNonTargetRuntimeAndUnownedStatIdentification(t *testing.T) {
	captured := ownedCapture()
	captured.Target = "classic"
	captured.Source = "community-tool"
	captured.Runtime.Patch = "1.13c"
	captured.Runtime.Mode = "classic"
	captured.Runtime.Session = "vanilla-server"
	captured.Runtime.CharacterOrigin = "imported-save"
	captured.Runtime.ExecutableSHA256 = "bad"
	captured.Runtime.RecordGenerationID = "bad"
	captured.Runtime.Observation = "memory-inspection"
	captured.Runtime.StatIdentification = "community-table"
	captured.Runtime.Locale = ""
	captured.Runtime.GameFramesPerSecond = 60

	if err := validate(captured); err == nil {
		t.Fatal("non-target runtime was accepted")
	}
}

// TestValidateRejectsUnknownSkillDuplicateProfileAndBadBoundaries keeps independent case failures cumulative.
func TestValidateRejectsUnknownSkillDuplicateProfileAndBadBoundaries(t *testing.T) {
	captured := ownedCapture()
	captured.Cases[0].SkillID = 99
	captured.Cases[0].EffectFrame = captured.Cases[0].StartFrame
	captured.Cases[1].ModifierSources[0].PropertyCode = "swing1"
	captured.Cases = append(captured.Cases, captured.Cases[1])

	if err := validate(captured); err == nil {
		t.Fatal("invalid case or duplicate profile was accepted")
	}
}

// TestCoverageReportsMissingProfilesWithoutPromotingAFormula verifies sparse evidence reports every absent profile.
func TestCoverageReportsMissingProfilesWithoutPromotingAFormula(t *testing.T) {
	captured := ownedCapture()
	got := coverageFor(captured.Cases[:1])

	if got.Complete || len(got.Missing) != 18 {
		t.Fatalf("coverage = %#v", got)
	}
}
