package loot

import (
	"strings"
	"testing"
)

func TestParseSkillsForClassSelection(t *testing.T) {
	skills, err := ParseSkillsTSV(strings.NewReader("Id\tskill\tcharclass\n2\tFire Bolt\tsor\n1\tMagic Arrow\tama\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 2 || skills[0].ID != 1 || skills[1].Class != 1 {
		t.Fatalf("skills = %#v", skills)
	}
}
