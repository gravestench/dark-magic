package typed

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/game/data/model"
	"github.com/gravestench/dark-magic/internal/game/data/store"
)

// TestLoadDecodesSurvivingTypedSchema verifies that known fields are decoded
// while a mod-owned column remains harmless to the engine schema.
func TestLoadDecodesSurvivingTypedSchema(t *testing.T) {
	t.Parallel()

	const (
		path     = "data/global/excel/charstats.txt"
		contents = "class\tstr\tdex\tint\tvit\tStartSkill\textra\n" +
			"ama\t20\t25\t15\t20\tAttack\tmod-owned\n"
	)

	records := loadFixture[models.CharStats](t, path, contents)
	if len(records) != 1 {
		t.Fatalf("decoded charstats count = %d, want 1: %#v", len(records), records)
	}

	record := records[0]
	if record.Class != "ama" || record.Strength != 20 || record.Dexterity != 25 || record.StartSkill != "Attack" {
		t.Fatalf("decoded charstats = %#v", records)
	}
}

// TestLoadSupportsAliasesPointersAndDiabloBooleans protects the destination
// conversions that differ from the strict decoder's default zero-value behavior.
func TestLoadSupportsAliasesPointersAndDiabloBooleans(t *testing.T) {
	t.Parallel()

	type alias int

	type record struct {
		Name     string `csv:"name"`
		Mode     alias  `csv:"mode"`
		Enabled  bool   `csv:"enabled"`
		Optional *int   `csv:"optional"`
	}

	const contents = "name\tmode\tenabled\toptional\n" +
		"first\t7\t1\t\n" +
		"second\t8\tfalse\t42\n"

	records := loadFixture[record](t, "test.txt", contents)
	if len(records) != 2 {
		t.Fatalf("decoded record count = %d, want 2: %#v", len(records), records)
	}

	first, second := records[0], records[1]
	if first.Mode != 7 || !first.Enabled || first.Optional != nil {
		t.Fatalf("first decoded record = %#v", first)
	}

	if second.Optional == nil || *second.Optional != 42 {
		t.Fatalf("second decoded record = %#v", second)
	}
}

// TestLoadRecoversHistoricalGroupedCSVTags ensures consecutive fields sharing
// one tag still receive their corresponding difficulty-specific columns.
func TestLoadRecoversHistoricalGroupedCSVTags(t *testing.T) {
	t.Parallel()

	type record struct {
		Normal, Nightmare, Hell int `csv:"Value,Value(N),Value(H)"`
	}

	const contents = "Value\tValue(N)\tValue(H)\n1\t2\t3\n"

	records := loadFixture[record](t, "grouped.txt", contents)
	if len(records) != 1 {
		t.Fatalf("decoded grouped record count = %d, want 1: %#v", len(records), records)
	}

	decoded := records[0]
	if decoded.Normal != 1 || decoded.Nightmare != 2 || decoded.Hell != 3 {
		t.Fatalf("decoded grouped record = %#v", records)
	}
}

// TestLoadRecoversHistoricalArrayCSVTags ensures fixed-size array fields retain
// the authored column order used by existing schemas.
func TestLoadRecoversHistoricalArrayCSVTags(t *testing.T) {
	t.Parallel()

	type record struct {
		Values [3]int `csv:"value1,value2,value3"`
	}

	const contents = "value1\tvalue2\tvalue3\n4\t5\t6\n"

	records := loadFixture[record](t, "array.txt", contents)
	if len(records) != 1 {
		t.Fatalf("decoded array record count = %d, want 1: %#v", len(records), records)
	}

	if records[0].Values != [3]int{4, 5, 6} {
		t.Fatalf("array values = %#v", records[0].Values)
	}
}

// TestLoadPreservesShippedBareQuotes exercises the narrow generic-store
// fallback without making the strict decoder globally tolerant.
func TestLoadPreservesShippedBareQuotes(t *testing.T) {
	t.Parallel()

	type record struct {
		Name   string `csv:"name"`
		Effect string `csv:"effect"`
		Code   int    `csv:"code"`
	}

	const contents = "name\teffect\tcode\nShrine\tHero's \"gift\t7\n"

	records := loadFixture[record](t, "quoted.txt", contents)
	if len(records) != 1 {
		t.Fatalf("decoded bare-quote record count = %d, want 1: %#v", len(records), records)
	}

	if records[0].Effect != `Hero's "gift` || records[0].Code != 7 {
		t.Fatalf("decoded bare quote = %#v", records)
	}
}

// TestLoadReportsSourceRowColumnAndField protects actionable source coordinates
// in strict-decoder failures so authors can repair the exact cell.
func TestLoadReportsSourceRowColumnAndField(t *testing.T) {
	t.Parallel()

	const contents = "class\tstr\nama\tnot-a-number\n"

	store := fixtureStore("bad.txt", contents)

	_, err := Load[models.CharStats](store, "bad.txt")
	if err == nil {
		t.Fatal("Load succeeded for an invalid numeric cell")
	}

	for _, fragment := range []string{"bad.txt", "line 2", "column 2", "not-a-number"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("error %q does not contain %q", err, fragment)
		}
	}
}

// loadFixture decodes one in-memory table and fails at the helper boundary so
// each scenario can present its behavior without repetitive store plumbing.
func loadFixture[T any](t *testing.T, path string, contents string) []T {
	t.Helper()

	records, err := Load[T](fixtureStore(path, contents), path)
	if err != nil {
		t.Fatal(err)
	}

	return records
}

// fixtureStore builds an immutable one-file source; each test owns its store,
// allowing parallel scenarios without shared state or cleanup obligations.
func fixtureStore(path string, contents string) *recordstore.Store {
	source := fstest.MapFS{
		path: &fstest.MapFile{Data: []byte(contents)},
	}

	return recordstore.New(source)
}
