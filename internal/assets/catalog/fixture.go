package assetcatalog

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

// Fixture is a redistributable structural fingerprint of one verified asset
// catalog. It deliberately contains no decoded pixels or source bytes.
type Fixture struct {
	Version         int            `json:"version"`
	ManifestVersion int            `json:"manifest_version"`
	Assets          []AssetFixture `json:"assets"`
}

// AssetFixture records stable source and decoder observations for one asset.
type AssetFixture struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Bytes      int    `json:"bytes"`
	Type       string `json:"type"`
	Directions int    `json:"directions,omitempty"`
	FrameCount int    `json:"frame_count,omitempty"`
	MinWidth   int    `json:"min_width,omitempty"`
	MaxWidth   int    `json:"max_width,omitempty"`
	MinHeight  int    `json:"min_height,omitempty"`
	MaxHeight  int    `json:"max_height,omitempty"`
	FramesHash string `json:"frames_sha256,omitempty"`
}

// FixtureFromReport extracts only successfully decoded observations. A
// partially resolved installation cannot accidentally become a golden fixture.
func FixtureFromReport(report Report) (Fixture, error) {
	fixture := Fixture{Version: 1, ManifestVersion: report.ManifestVersion}
	for _, result := range report.Results {
		if !result.Found || result.Error != "" {
			return Fixture{}, fmt.Errorf("asset catalog: cannot fixture %q: %s", result.ID, result.Error)
		}
		assetFixture := AssetFixture{
			ID:         result.ID,
			Path:       result.Path,
			SHA256:     result.SHA256,
			Bytes:      result.Bytes,
			Type:       result.Type,
			Directions: result.Directions,
		}
		assetFixture.addFrames(result.Frames)
		fixture.Assets = append(fixture.Assets, assetFixture)
	}
	return fixture, fixture.Validate()
}

// Validate checks a fixture independently of a game installation.
func (f Fixture) Validate() error {
	if f.Version != 1 || f.ManifestVersion < 1 {
		return errors.New("asset catalog: unsupported fixture contract")
	}
	if len(f.Assets) == 0 {
		return errors.New("asset catalog: fixture has no assets")
	}
	seen := make(map[string]struct{}, len(f.Assets))
	for index, asset := range f.Assets {
		if asset.ID == "" || asset.Path == "" || len(asset.SHA256) != 64 || asset.Bytes <= 0 || asset.Type == "" {
			return fmt.Errorf("asset catalog: fixture asset %d is incomplete", index)
		}
		if asset.Directions > 0 && (asset.FrameCount == 0 || len(asset.FramesHash) != 64) {
			return fmt.Errorf("asset catalog: fixture asset %q has incomplete frame metadata", asset.ID)
		}
		if _, exists := seen[asset.ID]; exists {
			return fmt.Errorf("asset catalog: duplicate fixture id %q", asset.ID)
		}
		seen[asset.ID] = struct{}{}
	}
	return nil
}

// CompareFixture returns every structural mismatch instead of stopping at the
// first one, making archive-version differences practical to review.
func CompareFixture(report Report, fixture Fixture) []string {
	if err := fixture.Validate(); err != nil {
		return []string{err.Error()}
	}
	actual := make(map[string]Result, len(report.Results))
	for _, result := range report.Results {
		actual[result.ID] = result
	}
	var mismatches []string
	for _, expected := range fixture.Assets {
		result, exists := actual[expected.ID]
		if !exists || !result.Found {
			mismatches = append(mismatches, fmt.Sprintf("%s: missing", expected.ID))
			continue
		}
		if result.Error != "" {
			mismatches = append(mismatches, fmt.Sprintf("%s: %s", expected.ID, result.Error))
			continue
		}
		observed := AssetFixture{Directions: result.Directions}
		observed.addFrames(result.Frames)
		if result.Path != expected.Path || result.SHA256 != expected.SHA256 || result.Bytes != expected.Bytes ||
			result.Type != expected.Type || observed.Directions != expected.Directions ||
			observed.FrameCount != expected.FrameCount || observed.MinWidth != expected.MinWidth ||
			observed.MaxWidth != expected.MaxWidth || observed.MinHeight != expected.MinHeight ||
			observed.MaxHeight != expected.MaxHeight || observed.FramesHash != expected.FramesHash {
			mismatches = append(mismatches, fmt.Sprintf("%s: structural fingerprint differs", expected.ID))
		}
	}
	return mismatches
}

func (f *AssetFixture) addFrames(frames []Frame) {
	f.FrameCount = len(frames)
	if len(frames) == 0 {
		return
	}
	f.MinWidth, f.MaxWidth = frames[0].Width, frames[0].Width
	f.MinHeight, f.MaxHeight = frames[0].Height, frames[0].Height
	hash := sha256.New()
	var encoded [8]byte
	for _, frame := range frames {
		values := [...]int{frame.Direction, frame.Frame, frame.Width, frame.Height, frame.OffsetX, frame.OffsetY}
		for _, value := range values {
			binary.LittleEndian.PutUint64(encoded[:], uint64(int64(value)))
			_, _ = hash.Write(encoded[:])
		}
		f.MinWidth = min(f.MinWidth, frame.Width)
		f.MaxWidth = max(f.MaxWidth, frame.Width)
		f.MinHeight = min(f.MinHeight, frame.Height)
		f.MaxHeight = max(f.MaxHeight, frame.Height)
	}
	f.FramesHash = hex.EncodeToString(hash.Sum(nil))
}
