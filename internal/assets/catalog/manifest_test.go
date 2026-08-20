package assetcatalog

import "testing"

// TestManifestValidate confirms a complete manifest passes and duplicate IDs fail before archive inspection. Unique IDs
// are required because reports and fixtures use them as stable identities across installations.
func TestManifestValidate(t *testing.T) {
	valid := Manifest{Version: 1, Assets: []Hypothesis{{
		ID:      "one",
		Screen:  "menu",
		Path:    "menu.dc6",
		Meaning: "background",
	}}}

	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}

	valid.Assets = append(valid.Assets, valid.Assets[0])
	if err := valid.Validate(); err == nil {
		t.Fatal("expected duplicate id to fail")
	}
}
