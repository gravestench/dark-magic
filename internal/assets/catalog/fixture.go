package assetcatalog

import (
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

// FixtureFromReport extracts only successfully decoded observations. Rejecting the entire report on the first failed
// asset prevents a partial installation from silently becoming the new golden structural fixture.
func FixtureFromReport(report Report) (Fixture, error) {
	fixture := Fixture{Version: 1, ManifestVersion: report.ManifestVersion}

	for _, result := range report.Results {
		if !result.Found || result.Error != "" {
			return Fixture{}, fmt.Errorf("asset catalog: cannot fixture %q: %s", result.ID, result.Error)
		}

		fixture.Assets = append(fixture.Assets, assetFixtureFromResult(result))
	}

	return fixture, fixture.Validate()
}

// assetFixtureFromResult copies only redistributable structural observations and summarizes frame metadata. Keeping
// source bytes and decoded pixels out of this boundary makes the resulting fixture safe to distribute.
func assetFixtureFromResult(result Result) AssetFixture {
	fixture := AssetFixture{
		ID:         result.ID,
		Path:       result.Path,
		SHA256:     result.SHA256,
		Bytes:      result.Bytes,
		Type:       result.Type,
		Directions: result.Directions,
	}
	fixture.addFrames(result.Frames)

	return fixture
}

// Validate checks a fixture independently of a game installation. It preserves first-failure ordering so malformed
// field errors remain more useful than later duplicate-ID errors for the same entry.
func (f Fixture) Validate() error {
	if f.Version != 1 || f.ManifestVersion < 1 {
		return errors.New("asset catalog: unsupported fixture contract")
	}

	if len(f.Assets) == 0 {
		return errors.New("asset catalog: fixture has no assets")
	}

	seen := make(map[string]struct{}, len(f.Assets))
	for index, asset := range f.Assets {
		if err := validateFixtureAsset(index, asset); err != nil {
			return err
		}

		if _, exists := seen[asset.ID]; exists {
			return fmt.Errorf("asset catalog: duplicate fixture id %q", asset.ID)
		}

		seen[asset.ID] = struct{}{}
	}

	return nil
}

// validateFixtureAsset enforces fields needed for byte and frame comparison before duplicate detection runs.
func validateFixtureAsset(index int, asset AssetFixture) error {
	if asset.ID == "" || asset.Path == "" || len(asset.SHA256) != 64 || asset.Bytes <= 0 || asset.Type == "" {
		return fmt.Errorf("asset catalog: fixture asset %d is incomplete", index)
	}

	if asset.Directions > 0 && (asset.FrameCount == 0 || len(asset.FramesHash) != 64) {
		return fmt.Errorf("asset catalog: fixture asset %q has incomplete frame metadata", asset.ID)
	}

	return nil
}
