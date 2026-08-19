package main

import "testing"

// TestLoadMapPreviewSkipsUnrequestedContent ensures default startup performs no filesystem work for the generated grid.
func TestLoadMapPreviewSkipsUnrequestedContent(t *testing.T) {
	preview, err := loadMapPreview(demoConfig{})
	if err != nil {
		t.Fatalf("load unrequested preview: %v", err)
	}

	if preview != nil {
		t.Fatalf("unrequested preview = %d bytes, want nil", len(preview))
	}
}

// TestLoadMapPreviewRequiresPairedInputs protects the exact startup error for incomplete source and map selections.
func TestLoadMapPreviewRequiresPairedInputs(t *testing.T) {
	configs := []demoConfig{
		{sourcePath: "maps.mpq"},
		{mapPath: "data/global/tiles/map.ds1"},
	}

	for _, config := range configs {
		preview, err := loadMapPreview(config)
		if err == nil {
			t.Fatalf("load preview with config %#v succeeded, want error", config)
		}

		if err.Error() != "both -source and -map are required" {
			t.Fatalf("load preview error = %q, want paired-input error", err)
		}

		if preview != nil {
			t.Fatalf("failed preview = %d bytes, want nil", len(preview))
		}
	}
}

// TestDemoConfigMapLabel preserves the generated fallback and the verbatim DS1 label shown in the HUD.
func TestDemoConfigMapLabel(t *testing.T) {
	if label := (demoConfig{}).mapLabel(); label != "generated grid" {
		t.Fatalf("default map label = %q, want %q", label, "generated grid")
	}

	config := demoConfig{mapPath: "data/global/tiles/map.ds1"}
	if label := config.mapLabel(); label != config.mapPath {
		t.Fatalf("selected map label = %q, want %q", label, config.mapPath)
	}
}
