package assetcatalog

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

// staticAssetPath intentionally recognizes only the asset formats covered by the catalog contract.
var staticAssetPath = regexp.MustCompile(
	`(?i)data/[a-z0-9_./ -]+\.(?:bik|cof|dat|dc6|dcc|ds1|dt1|pl2|tbl|txt|wav)`,
)

// luaAssetDiscovery owns path sets across both Lua roots so duplicate references collapse before classification.
type luaAssetDiscovery struct {
	source          fs.FS
	staticPaths     map[string]bool
	dynamicPrefixes map[string]bool
}

// readPresentationAssetPaths extracts asset-looking strings from arbitrary manifest JSON structure. JSON decode errors
// retain catalog context so callers can distinguish malformed declarations from filesystem failures.
func readPresentationAssetPaths(source fs.FS) (map[string]bool, error) {
	presentation, err := fs.ReadFile(source, "manifests/presentation.v1.json")
	if err != nil {
		return nil, err
	}

	var document any
	if err := json.Unmarshal(presentation, &document); err != nil {
		return nil, fmt.Errorf("asset catalog: decode presentation manifest: %w", err)
	}

	paths := make(map[string]bool)
	collectJSONPaths(document, paths)

	return paths, nil
}

// collectJSONPaths recursively visits JSON containers but records only complete strings that match a supported asset
// extension. Object key order is irrelevant because BuildCoverage sorts every result category before returning it.
func collectJSONPaths(value any, paths map[string]bool) {
	switch current := value.(type) {
	case []any:
		for _, item := range current {
			collectJSONPaths(item, paths)
		}
	case map[string]any:
		for _, item := range current {
			collectJSONPaths(item, paths)
		}
	case string:
		if staticAssetPath.MatchString(current) {
			paths[normalizeAssetPath(current)] = true
		}
	}
}

// discoverLuaAssetPaths walks the boot script before the Lua tree, preserving the original failure order. Static
// paths and string-concatenated prefixes are separated because only the former identify individual assets.
func discoverLuaAssetPaths(source fs.FS) (map[string]bool, map[string]bool, error) {
	discovery := luaAssetDiscovery{
		source:          source,
		staticPaths:     make(map[string]bool),
		dynamicPrefixes: make(map[string]bool),
	}

	for _, root := range []string{"boot.lua", "lua"} {
		if err := fs.WalkDir(source, root, discovery.inspectLuaFile); err != nil {
			return nil, nil, err
		}
	}

	return discovery.staticPaths, discovery.dynamicPrefixes, nil
}

// inspectLuaFile collects literal asset references from one Lua file. Read or walk failures stop discovery so coverage
// never implies that an unreadable subtree was fully examined.
func (d *luaAssetDiscovery) inspectLuaFile(name string, entry fs.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}

	if entry.IsDir() || path.Ext(name) != ".lua" {
		return nil
	}

	data, err := fs.ReadFile(d.source, name)
	if err != nil {
		return err
	}

	contents := string(data)
	for _, match := range staticAssetPath.FindAllString(contents, -1) {
		d.staticPaths[normalizeAssetPath(match)] = true
	}

	// Compile this per file, as the original scanner did, to avoid shifting work into package initialization.
	dynamicAssetPrefix := regexp.MustCompile(`(?i)data/[a-z0-9_./ -]+/`)
	for _, match := range dynamicAssetPrefix.FindAllString(contents, -1) {
		// Only concatenated prefixes are dynamic asset references; ordinary directory text remains unclassified.
		if strings.Contains(contents, match+`" ..`) {
			d.dynamicPrefixes[normalizeAssetPath(match)] = true
		}
	}

	return nil
}

// normalizeAssetPath mirrors the case-insensitive layered VFS and treats archive-style backslashes as separators.
// Trimming surrounding space prevents presentation formatting from creating distinct coverage identities.
func normalizeAssetPath(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), `\`, "/"))
}
