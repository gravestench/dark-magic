package assetinspect

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/gravestench/ds1"
)

// DS1TilePaths returns the mounted DT1 libraries declared by one DS1 stamp.
//
// DS1 records retain Blizzard's old TG1 development names even though shipped
// libraries use the DT1 extension. The declaration is authoritative: loading
// every DT1 beside a stamp makes one unrelated or unsupported library prevent
// otherwise valid stamps from rendering.
func DS1TilePaths(source fs.FS, ds1Path string) ([]string, error) {
	file, err := source.Open(ds1Path)
	if err != nil {
		return nil, fmt.Errorf("opening DS1 asset %q: %w", ds1Path, err)
	}

	data, readErr := io.ReadAll(file)
	closeErr := file.Close()

	if readErr != nil {
		return nil, fmt.Errorf("reading DS1 asset %q: %w", ds1Path, readErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("closing DS1 asset %q: %w", ds1Path, closeErr)
	}

	stamp, err := ds1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding DS1 asset %q: %w", ds1Path, err)
	}

	paths := resolveDS1TilePaths(stamp.Files)
	for _, tilePath := range paths {
		if _, err := fs.Stat(source, tilePath); err != nil {
			return nil, fmt.Errorf("DS1 asset %q declares unavailable tile library %q: %w", ds1Path, tilePath, err)
		}
	}

	return paths, nil
}

// resolveDS1TilePaths normalizes declarations in source order and removes
// case-insensitive duplicates so callers load each shipped library only once.
func resolveDS1TilePaths(files []string) []string {
	result := make([]string, 0, len(files))

	seen := make(map[string]struct{}, len(files))
	for _, declared := range files {
		normalized := normalizeDS1TilePath(declared)
		if normalized == "." || normalized == "" {
			continue
		}

		// MPQ lookups are case-insensitive, but the first declaration's spelling
		// remains part of the returned path and therefore must be preserved.
		key := strings.ToLower(normalized)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}

		result = append(result, normalized)
	}

	return result
}

// normalizeDS1TilePath converts development-era DS1 declarations into mounted
// archive paths while preserving the original spelling of every path segment.
func normalizeDS1TilePath(declared string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(declared), "\\", "/")

	normalized = strings.TrimLeft(normalized, "/")
	if strings.HasPrefix(strings.ToLower(normalized), "d2/") {
		normalized = normalized[len("d2/"):]
	}

	// DS1 stamps retain the internal TG1 name, but shipped tile libraries use
	// the DT1 extension regardless of the declaration's original casing.
	extension := path.Ext(normalized)
	if strings.EqualFold(extension, ".tg1") {
		normalized = strings.TrimSuffix(normalized, extension) + ".dt1"
	}

	return path.Clean(normalized)
}
