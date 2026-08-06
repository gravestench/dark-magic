package gamedata

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/recordstore"
	"github.com/gravestench/dark-magic/pkg/models"
)

func TestLoadDecodesSurvivingTypedSchema(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{CharStatsTable: &fstest.MapFile{Data: []byte("class\tstr\tdex\tint\tvit\tStartSkill\textra\nama\t20\t25\t15\t20\tAttack\tmod-owned\n")}}
	records, err := Load[models.CharStats](recordstore.New(source), CharStatsTable)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Class != "ama" || records[0].Strength != 20 || records[0].Dexterity != 25 || records[0].StartSkill != "Attack" {
		t.Fatalf("decoded charstats = %#v", records)
	}
}

func TestLoadSupportsAliasesPointersAndDiabloBooleans(t *testing.T) {
	t.Parallel()

	type alias int
	type record struct {
		Name     string `csv:"name"`
		Mode     alias  `csv:"mode"`
		Enabled  bool   `csv:"enabled"`
		Optional *int   `csv:"optional"`
	}
	source := fstest.MapFS{"test.txt": &fstest.MapFile{Data: []byte("name\tmode\tenabled\toptional\nfirst\t7\t1\t\nsecond\t8\tfalse\t42\n")}}
	records, err := Load[record](recordstore.New(source), "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Mode != 7 || !records[0].Enabled || records[0].Optional != nil || records[1].Optional == nil || *records[1].Optional != 42 {
		t.Fatalf("decoded records = %#v", records)
	}
}

func TestLoadReportsSourceRowColumnAndField(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{"bad.txt": &fstest.MapFile{Data: []byte("class\tstr\nama\tnot-a-number\n")}}
	_, err := Load[models.CharStats](recordstore.New(source), "bad.txt")
	for _, fragment := range []string{"bad.txt", "row 2", `column "str"`, "field Strength"} {
		if err == nil || !strings.Contains(err.Error(), fragment) {
			t.Fatalf("error %q does not contain %q", err, fragment)
		}
	}
}

func TestIndexRejectsDuplicatePrimaryKeys(t *testing.T) {
	t.Parallel()

	_, err := Index([]models.CharStats{{Class: "ama"}, {Class: "ama"}}, func(record models.CharStats) string { return record.Class })
	if err == nil || !strings.Contains(err.Error(), "duplicate key ama") {
		t.Fatalf("duplicate index error = %v", err)
	}
}
