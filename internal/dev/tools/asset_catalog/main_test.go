package main

import "testing"

func TestDefaultManifestComesFromEmbeddedD2Legacy(t *testing.T) {
	manifest, err := loadManifest("")
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	if manifest.Version != 1 || len(manifest.Assets) != 113 {
		t.Fatalf("default manifest = version %d with %d assets", manifest.Version, len(manifest.Assets))
	}
}
