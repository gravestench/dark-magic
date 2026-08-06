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

func TestLoadRecoversHistoricalGroupedCSVTags(t *testing.T) {
	t.Parallel()

	type record struct {
		Normal, Nightmare, Hell int `csv:"Value,Value(N),Value(H)"`
	}
	source := fstest.MapFS{"grouped.txt": &fstest.MapFile{Data: []byte("Value\tValue(N)\tValue(H)\n1\t2\t3\n")}}
	records, err := Load[record](recordstore.New(source), "grouped.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Normal != 1 || records[0].Nightmare != 2 || records[0].Hell != 3 {
		t.Fatalf("decoded grouped record = %#v", records)
	}
}

func TestLoadReportsSourceRowColumnAndField(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{"bad.txt": &fstest.MapFile{Data: []byte("class\tstr\nama\tnot-a-number\n")}}
	_, err := Load[models.CharStats](recordstore.New(source), "bad.txt")
	for _, fragment := range []string{"bad.txt", "line 2", "column 2", "not-a-number"} {
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

func TestObservedIndexPreservesRowsAndReportsDuplicate(t *testing.T) {
	t.Parallel()

	records := []models.CharStats{{Class: "unused", Strength: 1}, {Class: "unused", Strength: 2}}
	index, issues, err := ObservedIndex("charstats", records, func(record models.CharStats) string { return record.Class })
	if err != nil {
		t.Fatal(err)
	}
	if index["unused"].Strength != 1 || len(issues) != 1 || issues[0].Row != 3 || issues[0].Kind != "duplicate-key" {
		t.Fatalf("observed index = %#v, issues = %#v", index, issues)
	}
}
