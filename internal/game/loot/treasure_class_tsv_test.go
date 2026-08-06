package loot

import (
	"strings"
	"testing"
)

func TestParseTreasureClassTSV(t *testing.T) {
	input := "\ufeffTreasure Class\tgroup\tPicks\tUnique\tSet\tRare\tMagic\tNoDrop\tItem1\tProb1\tItem2\tProb2\r\n" +
		"Root\t1\t2\t900\t800\t700\t600\t3\tNested\t4\tr01\t5\r\n" +
		"Nested\t1\t-1\t0\t0\t0\t0\t0\tamu\t1\t\t\r\n"
	catalog, err := ParseTreasureClassTSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	root := catalog["Root"]
	if root.Name != "Root" || root.Picks != 2 || root.NoDrop != 3 || len(root.Entries) != 2 || root.Entries[1].Code != "r01" || root.Entries[1].Weight != 5 {
		t.Fatalf("root = %#v", root)
	}
	if root.Quality != (QualityModifiers{Unique: 900, Set: 800, Rare: 700, Magic: 600}) {
		t.Fatalf("quality = %#v", root.Quality)
	}
	drops, err := New(catalog, 4).Roll("Nested")
	if err != nil || len(drops) != 1 || drops[0].Code != "amu" {
		t.Fatalf("drops = %#v, error = %v", drops, err)
	}
}

func TestParseTreasureClassTSVErrorsAreActionable(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "missing header", input: "Treasure Class\tPicks\n", want: "NoDrop"},
		{name: "bad integer", input: "Treasure Class\tPicks\tNoDrop\nRoot\tmany\t0\n", want: "row 2 column \"Picks\""},
		{name: "duplicate", input: "Treasure Class\tPicks\tNoDrop\nRoot\t1\t0\nRoot\t1\t0\n", want: "duplicate"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseTreasureClassTSV(strings.NewReader(test.input))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}
