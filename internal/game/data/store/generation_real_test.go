package recordstore

import (
	"os"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

func TestOwnedTargetArchivesPinRequiredUnlistedTablesCaseInsensitively(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the expansion 1.14d MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	assets, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	defer assets.Close()
	pinned, generation, err := Pin(assets)
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"monlvl", "monpreset", "skilldesc"} {
		path := "data/global/excel/" + table + ".txt"
		if _, err := pinned.Load(path); err != nil {
			var matching []string
			for _, file := range generation.Files {
				if strings.Contains(strings.ToLower(file.Path), table) {
					matching = append(matching, file.Path)
				}
			}
			t.Fatalf("lowercase %s lookup: %v (pinned matches: %v)", table, err, matching)
		}
	}
}
