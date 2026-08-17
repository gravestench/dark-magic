package behaviorcoverage

import (
	"strings"
	"testing"
)

func TestBuildGroupsSignaturesWithoutInferringCoverage(t *testing.T) {
	manifest := Manifest{
		Schema: Schema, Version: 1, Target: Target,
		Implementations: []Implementation{{
			SkillID: 36, Family: "missile.straight", EvidenceStatus: "owned-target-records-verified",
		}},
	}
	skills := []map[string]string{
		{"Id": "901", "skill": "Similar Fixture", "srvstfunc": "", "srvdofunc": "", "srvmissile": "similar"},
		{"Id": "36", "skill": "Reviewed Fixture", "srvstfunc": "", "srvdofunc": "", "srvmissile": "reviewed"},
		{"Id": "2", "skill": "Other Fixture", "srvstfunc": "7", "srvdofunc": "12"},
	}
	missiles := []map[string]string{
		{"Missile": "similar", "pSrvDoFunc": "1"},
		{"Missile": "reviewed", "pSrvDoFunc": "1"},
	}
	report, err := Build(manifest, skills, missiles, Sources{SkillsLayer: "patch", MissilesLayer: "patch"})
	if err != nil {
		t.Fatal(err)
	}
	if report.Summary.SkillRows != 3 || report.Summary.BehaviorGroups != 2 ||
		report.Summary.ImplementedSkills != 1 || report.Summary.MissingSkills != 2 {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	var shared BehaviorGroup
	for _, group := range report.Behaviors {
		if len(group.MissileServerDoFunctions) == 1 && group.MissileServerDoFunctions[0] == "1" {
			shared = group
		}
	}
	if len(shared.Consumers) != 2 || shared.Consumers[0].SkillID != 36 || shared.Consumers[1].SkillID != 901 {
		t.Fatalf("shared signature consumers are not deterministically sorted: %+v", shared.Consumers)
	}
	if shared.Consumers[0].MissingFamily || shared.Consumers[0].ImplementationFamily != "missile.straight" {
		t.Fatalf("reviewed skill lost declaration: %+v", shared.Consumers[0])
	}
	if !shared.Consumers[1].MissingFamily || shared.Consumers[1].ImplementationFamily != "" ||
		shared.Consumers[1].EvidenceStatus != "missing-implementation-family" {
		t.Fatalf("similar signature was implicitly admitted: %+v", shared.Consumers[1])
	}
}

func TestDecodeManifestRejectsNonTargetAndUnknownFields(t *testing.T) {
	for _, input := range []string{
		`{"schema":"d2legacy.skill_behavior_coverage/v1","version":1,"target":"classic","implementations":[]}`,
		`{"schema":"d2legacy.skill_behavior_coverage/v1","version":1,"target":"diablo-ii-lod-1.14d-expansion","implementations":[],"legacy":true}`,
	} {
		if _, err := DecodeManifest(strings.NewReader(input)); err == nil {
			t.Fatalf("DecodeManifest accepted unsupported input: %s", input)
		}
	}
}

func TestBuildRejectsMissingDeclaredSkill(t *testing.T) {
	manifest := Manifest{
		Schema: Schema, Version: 1, Target: Target,
		Implementations: []Implementation{{SkillID: 36, Family: "missile.straight", EvidenceStatus: "verified"}},
	}
	if _, err := Build(manifest, []map[string]string{{"Id": "40"}}, nil, Sources{}); err == nil {
		t.Fatal("Build accepted a declaration absent from mounted Skills.txt")
	}
}
