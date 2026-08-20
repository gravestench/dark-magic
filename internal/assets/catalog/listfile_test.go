package assetcatalog

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestParseAndAuditListfilePreservesKnowledgeAndLocalStatus protects deduplication and traversal rejection.
// Availability stays in entry order, and missing local files remain useful knowledge rather than parse errors.
func TestParseAndAuditListfilePreservesKnowledgeAndLocalStatus(t *testing.T) {
	listfile := "data\\global\\ui\\One.dc6\r\n" +
		"DATA/global/UI/one.dc6\r\n" +
		"missing.wav\r\n" +
		"../escape\r\n"

	entries, err := ParseListfile(strings.NewReader(listfile))
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 || entries[0].Normalized != "data/global/ui/One.dc6" {
		t.Fatalf("entries = %#v", entries)
	}

	source := fstest.MapFS{
		"data/global/ui/One.dc6": {Data: []byte("one")},
	}
	report := AuditListfile(source, entries)

	countsDiffer := report.Listed != 2 || report.Found != 1 || report.Missing != 1

	statusDiffers := !report.Entries[0].Found || report.Entries[1].Found
	if countsDiffer || statusDiffers {
		t.Fatalf("report = %#v", report)
	}
}
