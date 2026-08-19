package main

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

// TestCatalogFlagsHasInput locks down whitespace-only rejection without requiring process-level exit assertions.
func TestCatalogFlagsHasInput(t *testing.T) {
	tests := []struct {
		name    string
		options catalogFlags
		want    bool
	}{
		{name: "empty", options: catalogFlags{}, want: false},
		{name: "whitespace assets", options: catalogFlags{assets: " \t"}, want: false},
		{name: "whitespace stamps", options: catalogFlags{stamps: "\n"}, want: false},
		{name: "asset", options: catalogFlags{assets: "data/global/tiles/foo.dt1"}, want: true},
		{name: "stamp", options: catalogFlags{stamps: "data/global/tiles/foo.ds1"}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.options.hasInput(); got != test.want {
				t.Fatalf("hasInput() = %t, want %t", got, test.want)
			}
		})
	}
}

// TestInspectAssetListPreservesMeaningfulPathOrder verifies trimming and empty-entry removal without deduplication.
func TestInspectAssetListPreservesMeaningfulPathOrder(t *testing.T) {
	var inspected []string

	inspect := func(_ *content.FS, _ io.Writer, path string) error {
		inspected = append(inspected, path)
		return nil
	}

	err := inspectAssetList(nil, io.Discard, " first.dt1, ,second.dt1,first.dt1 ", inspect)
	if err != nil {
		t.Fatalf("inspectAssetList() error = %v", err)
	}

	want := []string{"first.dt1", "second.dt1", "first.dt1"}
	if !reflect.DeepEqual(inspected, want) {
		t.Fatalf("inspected paths = %q, want %q", inspected, want)
	}
}

// TestInspectAssetListStopsAfterFailure ensures later paths cannot produce misleading output after a partial failure.
func TestInspectAssetListStopsAfterFailure(t *testing.T) {
	wantErr := errors.New("inspection failed")

	var inspected []string

	inspect := func(_ *content.FS, _ io.Writer, path string) error {
		inspected = append(inspected, path)
		if path == "second.dt1" {
			return wantErr
		}

		return nil
	}

	err := inspectAssetList(nil, io.Discard, "first.dt1,second.dt1,third.dt1", inspect)
	if !errors.Is(err, wantErr) {
		t.Fatalf("inspectAssetList() error = %v, want %v", err, wantErr)
	}

	want := []string{"first.dt1", "second.dt1"}
	if !reflect.DeepEqual(inspected, want) {
		t.Fatalf("inspected paths = %q, want %q", inspected, want)
	}
}
