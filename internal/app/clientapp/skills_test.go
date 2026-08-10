package clientapp

import (
	"reflect"
	"testing"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/model"
	"github.com/gravestench/dark-magic/internal/persistence"
)

func TestLearnedSkillsUsesGeneralAndClassStartingActions(t *testing.T) {
	snapshot := gamedata.Snapshot{
		CharStatsByClass: map[string]models.CharStats{"Amazon": {Class: "Amazon", StartSkill: "Fire Arrow"}},
		Skills: []models.SkillData{
			{ID: "0", SkillName: "Attack", SkillDesc: "attack", General: "1", LeftSkill: "1"},
			{ID: "6", SkillName: "Fire Arrow", SkillDesc: "fire", CharClass: "ama", LeftSkill: "1"},
			{ID: "7", SkillName: "Inner Sight", SkillDesc: "sight", CharClass: "ama", LeftSkill: "0"},
			{ID: "8", SkillName: "Passive", SkillDesc: "passive", General: "1", Passive: "1"},
		},
		SkillDescByName: map[string]models.SkillDescData{
			"attack":  {SkillDesc: "attack", ListRow: 0},
			"fire":    {SkillDesc: "fire", ListRow: 1},
			"sight":   {SkillDesc: "sight", ListRow: 1},
			"passive": {SkillDesc: "passive", ListRow: 2},
		},
	}

	got := learnedSkills(snapshot, persistence.Character{Class: "amazon"})
	wantIDs := []int64{0, 6}
	gotIDs := make([]int64, len(got))
	for index, skill := range got {
		gotIDs[index] = skill.ID
		if !skill.RightAllowed || skill.Level != 1 {
			t.Fatalf("invalid admitted skill: %#v", skill)
		}
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("learned skill IDs = %v, want %v", gotIDs, wantIDs)
	}
}
