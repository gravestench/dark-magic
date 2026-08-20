package assetcatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"slices"
)

// Coverage classifies presentation paths without requiring proprietary assets.
// Paths are normalized case-insensitively because the layered VFS is likewise
// case-insensitive for Diablo archive names.
type Coverage struct {
	Version                int      `json:"version"`
	PresentationPaths      []string `json:"presentation_paths"`
	VerifiedPresentation   []string `json:"verified_presentation"`
	UnverifiedPresentation []string `json:"unverified_presentation"`
	CodeOwnedPaths         []string `json:"code_owned_paths"`
	DynamicPrefixes        []string `json:"dynamic_prefixes"`
	CatalogOnlyPaths       []string `json:"catalog_only_paths"`
	CatalogFixtureGaps     []string `json:"catalog_fixture_gaps"`
	Fingerprint            string   `json:"fingerprint"`
}

// BuildCoverage joins presentation declarations, Lua references, the curated catalog, and its structural fixture.
// Discovery completes before classification so errors never return a partially populated coverage report.
func BuildCoverage(source fs.FS, manifest Manifest, fixture Fixture) (Coverage, error) {
	presentationPaths, err := readPresentationAssetPaths(source)
	if err != nil {
		return Coverage{}, err
	}

	codePaths, dynamicPrefixes, err := discoverLuaAssetPaths(source)
	if err != nil {
		return Coverage{}, err
	}

	catalogPaths, catalogIDs := indexCatalogAssets(manifest)
	fixtureIDs := indexFixtureAssets(fixture)
	result := classifyCoverage(
		presentationPaths,
		codePaths,
		dynamicPrefixes,
		catalogPaths,
		catalogIDs,
		fixtureIDs,
	)

	result.sort()
	result.Fingerprint = result.fingerprint()

	return result, nil
}

// indexCatalogAssets creates path and ID indexes with the same last-entry-wins semantics as the source manifest.
// Keeping both views avoids repeatedly normalizing paths during each classification pass.
func indexCatalogAssets(manifest Manifest) (map[string]bool, map[string]string) {
	paths := make(map[string]bool, len(manifest.Assets))
	ids := make(map[string]string, len(manifest.Assets))

	for _, asset := range manifest.Assets {
		normalized := normalizeAssetPath(asset.Path)
		paths[normalized] = true
		ids[asset.ID] = normalized
	}

	return paths, ids
}

// indexFixtureAssets normalizes fixture paths by ID so coverage can detect both absent and path-divergent fixtures.
func indexFixtureAssets(fixture Fixture) map[string]string {
	ids := make(map[string]string, len(fixture.Assets))
	for _, asset := range fixture.Assets {
		ids[asset.ID] = normalizeAssetPath(asset.Path)
	}

	return ids
}

// classifyCoverage assigns every discovered path to its review category and reports fixture ID/path discrepancies.
// Map iteration order is intentionally resolved by the final sort rather than influencing category membership.
func classifyCoverage(
	presentationPaths map[string]bool,
	codePaths map[string]bool,
	dynamicPrefixes map[string]bool,
	catalogPaths map[string]bool,
	catalogIDs map[string]string,
	fixtureIDs map[string]string,
) Coverage {
	result := Coverage{Version: 1}

	for name := range presentationPaths {
		result.PresentationPaths = append(result.PresentationPaths, name)

		if catalogPaths[name] {
			result.VerifiedPresentation = append(result.VerifiedPresentation, name)
		} else {
			result.UnverifiedPresentation = append(result.UnverifiedPresentation, name)
		}
	}

	for name := range codePaths {
		if !presentationPaths[name] {
			result.CodeOwnedPaths = append(result.CodeOwnedPaths, name)
		}
	}

	for name := range dynamicPrefixes {
		result.DynamicPrefixes = append(result.DynamicPrefixes, name)
	}

	for name := range catalogPaths {
		if !presentationPaths[name] {
			result.CatalogOnlyPaths = append(result.CatalogOnlyPaths, name)
		}
	}

	for id, catalogPath := range catalogIDs {
		if fixturePath, ok := fixtureIDs[id]; !ok || fixturePath != catalogPath {
			result.CatalogFixtureGaps = append(result.CatalogFixtureGaps, id)
		}
	}

	for id := range fixtureIDs {
		if _, ok := catalogIDs[id]; !ok {
			result.CatalogFixtureGaps = append(result.CatalogFixtureGaps, id)
		}
	}

	return result
}

// sort establishes deterministic order for JSON output and for the fingerprint calculated immediately afterward.
func (c *Coverage) sort() {
	for _, values := range c.pathGroups() {
		slices.Sort(values)
	}
}

// fingerprint hashes each ordered category with explicit item and category separators. The separators prevent distinct
// path/group combinations from producing the same concatenated byte stream.
func (c Coverage) fingerprint() string {
	hash := sha256.New()

	for _, group := range c.pathGroups() {
		for _, value := range group {
			_, _ = hash.Write([]byte(value))
			_, _ = hash.Write([]byte{0})
		}

		_, _ = hash.Write([]byte{0xff})
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// pathGroups defines the persisted category order once for sorting and hashing. New coverage fields must be added here
// so deterministic output and fingerprint compatibility cannot diverge accidentally.
func (c Coverage) pathGroups() [][]string {
	return [][]string{
		c.PresentationPaths,
		c.VerifiedPresentation,
		c.UnverifiedPresentation,
		c.CodeOwnedPaths,
		c.DynamicPrefixes,
		c.CatalogOnlyPaths,
		c.CatalogFixtureGaps,
	}
}
