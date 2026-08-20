// Package assetcatalog verifies documented Diablo II asset knowledge against a
// content filesystem. It never modifies the source archives.
package assetcatalog

// Source is the provenance needed by the catalog without coupling this public
// package to the internal content package.
type Source struct {
	Layer string `json:"layer"`
	Path  string `json:"path"`
}

// Hypothesis is one piece of community-derived knowledge to verify.
type Hypothesis struct {
	ID         string   `json:"id"`
	Screen     string   `json:"screen"`
	Path       string   `json:"path"`
	Palette    string   `json:"palette,omitempty"`
	Meaning    string   `json:"meaning"`
	References []string `json:"references,omitempty"`
}

// Manifest is a versioned set of hypotheses.
type Manifest struct {
	Version int          `json:"version"`
	Assets  []Hypothesis `json:"assets"`
}

// Frame describes stored DC6 placement metadata.
type Frame struct {
	Direction int `json:"direction"`
	Frame     int `json:"frame"`
	Width     int `json:"width"`
	Height    int `json:"height"`
	OffsetX   int `json:"offset_x"`
	OffsetY   int `json:"offset_y"`
}

// Result records whether a hypothesis resolved and what was observed.
type Result struct {
	Hypothesis
	Found      bool     `json:"found"`
	Source     *Source  `json:"source,omitempty"`
	Bytes      int      `json:"bytes,omitempty"`
	SHA256     string   `json:"sha256,omitempty"`
	Type       string   `json:"type,omitempty"`
	Directions int      `json:"directions,omitempty"`
	Frames     []Frame  `json:"frames,omitempty"`
	PaletteOK  bool     `json:"palette_found,omitempty"`
	Sheet      string   `json:"contact_sheet,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// Report is deterministic output suitable for review or mod generation.
type Report struct {
	ManifestVersion int      `json:"manifest_version"`
	Results         []Result `json:"results"`
	Aliases         []Alias  `json:"aliases,omitempty"`
}

// Alias groups different documented paths whose resolved bytes are identical.
type Alias struct {
	SHA256 string   `json:"sha256"`
	Paths  []string `json:"paths"`
}

// Options controls optional diagnostic artifacts.
type Options struct {
	SheetWriter func(name string, data []byte) (string, error)
	Resolve     func(name string) (Source, error)
}
