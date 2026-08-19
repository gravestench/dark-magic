package localization

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	tbl "github.com/gravestench/tbl_text"
)

// localeComposer overlays every available localization format into a single cache generation.
// Keeping construction state private ensures callers receive either the complete generation or its terminal error.
type localeComposer struct {
	source   fs.FS
	language string
	strings  map[string]string
	sources  map[string]string
}

// newLocaleComposer allocates both result maps before any I/O so even failed generations have consistent state.
// Each cache generation owns these maps exclusively until Locale.load publishes them under the locale mutex.
func newLocaleComposer(source fs.FS, language string) *localeComposer {
	return &localeComposer{
		source:   source,
		language: language,
		strings:  make(map[string]string),
		sources:  make(map[string]string),
	}
}

// compose loads the JSON compatibility layer, then Diablo base, expansion, and patch tables in precedence order.
// Later layers overwrite both text and attribution, keeping the winning value and source path inseparable.
func (composer *localeComposer) compose() (map[string]string, map[string]string, error) {
	loaded, err := composer.loadJSONCompatibilityLayer()
	if err != nil {
		return composer.strings, composer.sources, err
	}

	for _, path := range diabloTablePaths(composer.language) {
		table, found, err := readDiabloTable(composer.source, path)
		if err != nil {
			return composer.strings, composer.sources, err
		}

		if !found {
			continue
		}

		loaded = true

		composer.overlayTable(path, table)
	}

	if !loaded {
		return composer.strings, composer.sources,
			fmt.Errorf("localization: no %s string tables found", composer.language)
	}

	return composer.strings, composer.sources, nil
}

// loadJSONCompatibilityLayer loads application-owned strings that may exist without Diablo table assets.
// Recording attribution only after a successful decode prevents malformed JSON from appearing partially usable.
func (composer *localeComposer) loadJSONCompatibilityLayer() (bool, error) {
	path := fmt.Sprintf("locales/%s.json", composer.language)

	data, err := fs.ReadFile(composer.source, path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("localization: read %q: %w", path, err)
	}

	if err := json.Unmarshal(data, &composer.strings); err != nil {
		return false, fmt.Errorf("localization: decode %q: %w", path, err)
	}

	for key := range composer.strings {
		composer.sources[key] = path
	}

	return len(composer.strings) > 0, nil
}

// overlayTable applies one complete Diablo layer without separating a value from its source attribution.
// Iteration order is irrelevant within a layer because each table contains one final value per key.
func (composer *localeComposer) overlayTable(path string, table tbl.TextTable) {
	for key, value := range table {
		composer.strings[key] = value
		composer.sources[key] = path
	}
}

// diabloTablePaths returns table paths from lowest to highest precedence so ordinary assignment performs overlaying.
// The fixed order is part of localization behavior: patch entries must win over expansion and base entries.
func diabloTablePaths(language string) [3]string {
	directory := diabloTableLanguage(language)

	return [3]string{
		fmt.Sprintf("data/local/lng/%s/string.tbl", directory),
		fmt.Sprintf("data/local/lng/%s/expansionstring.tbl", directory),
		fmt.Sprintf("data/local/lng/%s/patchstring.tbl", directory),
	}
}

// diabloTableLanguage maps the display name English to Diablo's on-disk abbreviation without altering other names.
// Equal-fold matching preserves compatibility with callers that vary English's capitalization.
func diabloTableLanguage(language string) string {
	if strings.EqualFold(language, "English") {
		return "eng"
	}

	return language
}

// readDiabloTable opens, buffers, decodes, and closes one table while distinguishing absence from a load failure.
// Decode failures take precedence over close failures, matching the original error contract after both operations run.
func readDiabloTable(source fs.FS, path string) (tbl.TextTable, bool, error) {
	file, err := source.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("localization: read %q: %w", path, err)
	}

	// Compressed MPQ files may implement ReaderAt with an archive read per call. Buffering these small tables once keeps
	// the decoder's fine-grained hash and key reads memory-local.
	data, decodeErr := io.ReadAll(file)

	var table tbl.TextTable
	if decodeErr == nil {
		table, decodeErr = tbl.Unmarshal(data)
	}

	closeErr := file.Close()

	if decodeErr != nil {
		return nil, false, fmt.Errorf("localization: decode %q: %w", path, decodeErr)
	}

	if closeErr != nil {
		return nil, false, fmt.Errorf("localization: close %q: %w", path, closeErr)
	}

	return table, true, nil
}
