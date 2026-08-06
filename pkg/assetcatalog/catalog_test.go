package assetcatalog

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestVerifyReportsFoundAndMissingAssetsWithoutStopping(t *testing.T) {
	manifest := Manifest{Version: 1, Assets: []Hypothesis{
		{ID: "found", Screen: "test", Path: "data/found.txt", Meaning: "fixture"},
		{ID: "missing", Screen: "test", Path: "data/missing.txt", Meaning: "fixture"},
	}}
	source := fstest.MapFS{"data/found.txt": {Data: []byte("screen knowledge")}}
	report := Verify(source, manifest, Options{Resolve: func(name string) (Source, error) {
		if name == "data/found.txt" {
			return Source{Layer: "fixture.mpq", Path: name}, nil
		}
		return Source{}, errors.New("missing")
	}})
	if len(report.Results) != 2 {
		t.Fatalf("got %d results", len(report.Results))
	}
	if !report.Results[0].Found || report.Results[0].Source == nil || report.Results[0].Source.Layer != "fixture.mpq" {
		t.Fatalf("unexpected found result: %+v", report.Results[0])
	}
	if report.Results[1].Found || report.Results[1].Error == "" {
		t.Fatalf("unexpected missing result: %+v", report.Results[1])
	}
}

func TestVerifyReportsByteIdenticalAliases(t *testing.T) {
	manifest := Manifest{Version: 1, Assets: []Hypothesis{
		{ID: "one", Screen: "test", Path: "one.bin", Meaning: "fixture"},
		{ID: "two", Screen: "test", Path: "two.bin", Meaning: "fixture alias"},
	}}
	report := Verify(fstest.MapFS{
		"one.bin": {Data: []byte("same")},
		"two.bin": {Data: []byte("same")},
	}, manifest, Options{})
	if len(report.Aliases) != 1 || len(report.Aliases[0].Paths) != 2 {
		t.Fatalf("unexpected aliases: %+v", report.Aliases)
	}
}

func TestManifestValidate(t *testing.T) {
	valid := Manifest{Version: 1, Assets: []Hypothesis{{ID: "one", Screen: "menu", Path: "menu.dc6", Meaning: "background"}}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	valid.Assets = append(valid.Assets, valid.Assets[0])
	if err := valid.Validate(); err == nil {
		t.Fatal("expected duplicate id to fail")
	}
}

func TestDC6ContactSheetRejectsNil(t *testing.T) {
	if _, err := DC6ContactSheet(nil); err == nil {
		t.Fatal("expected nil DC6 to fail")
	}
}

func TestReadPalette(t *testing.T) {
	data := make([]byte, 256*3)
	data[3], data[4], data[5] = 10, 20, 30
	palette, err := readPalette(fstest.MapFS{"pal.dat": {Data: data}}, "pal.dat")
	if err != nil {
		t.Fatal(err)
	}
	r, g, b, a := palette[1].RGBA()
	if r>>8 != 10 || g>>8 != 20 || b>>8 != 30 || a>>8 != 255 {
		t.Fatalf("unexpected color: %d %d %d %d", r>>8, g>>8, b>>8, a>>8)
	}
}
