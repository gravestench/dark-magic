package assetcatalog

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseAndAuditListfilePreservesKnowledgeAndLocalStatus(t *testing.T) {
	entries, err := ParseListfile(strings.NewReader("data\\global\\ui\\One.dc6\r\nDATA/global/UI/one.dc6\r\nmissing.wav\r\n../escape\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Normalized != "data/global/ui/One.dc6" {
		t.Fatalf("entries = %#v", entries)
	}
	report := AuditListfile(fstest.MapFS{"data/global/ui/One.dc6": &fstest.MapFile{Data: []byte("one")}}, entries)
	if report.Listed != 2 || report.Found != 1 || report.Missing != 1 || !report.Entries[0].Found || report.Entries[1].Found {
		t.Fatalf("report = %#v", report)
	}
}
