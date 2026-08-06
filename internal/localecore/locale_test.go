package localecore

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLocaleReportsMissingTablesAndPreservesKey(t *testing.T) {
	t.Parallel()

	locale := New(fstest.MapFS{}, "English")
	value, err := locale.Text("missing")
	if value != "missing" || err == nil || !strings.Contains(err.Error(), "no English string tables") {
		t.Fatalf("Text = %q/%v", value, err)
	}
}
