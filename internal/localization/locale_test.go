package localization

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

func TestShimLocalizationKeysLoadWithoutDiabloTables(t *testing.T) {
	locale := New(content.Shim(), "English")
	value, err := locale.Text("d2.hud.title")
	if err != nil || value != "Dark Magic" {
		t.Fatalf("title = %q, %v", value, err)
	}
	if languages := locale.GetSupportedLanguages(); len(languages) != 1 || languages[0] != "English" {
		t.Fatalf("languages = %v", languages)
	}
}

func TestLocaleReportsMissingTablesAndPreservesKey(t *testing.T) {
	t.Parallel()

	locale := New(fstest.MapFS{}, "English")
	value, err := locale.Text("missing")
	if value != "missing" || err == nil || !strings.Contains(err.Error(), "no English string tables") {
		t.Fatalf("Text = %q/%v", value, err)
	}
}
