package localization

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

// TestD2LegacyLocalizationKeysLoadWithoutDiabloTables verifies that the JSON compatibility layer stands alone.
// This protects application-owned UI strings when a mounted content source has no original game tables.
func TestD2LegacyLocalizationKeysLoadWithoutDiabloTables(t *testing.T) {
	locale := New(content.D2Legacy(), "English")

	value, err := locale.Text("d2legacy.hud.title")
	if err != nil {
		t.Fatalf("Text returned an error: %v", err)
	}

	if value != "Dark Magic" {
		t.Fatalf("Text = %q, want %q", value, "Dark Magic")
	}

	if languages := locale.GetSupportedLanguages(); len(languages) != 1 || languages[0] != "English" {
		t.Fatalf("GetSupportedLanguages = %v, want [English]", languages)
	}
}

// TestLocaleReportsMissingTablesAndPreservesKey verifies the display-safe fallback for an empty content source.
// Returning the requested key alongside the error lets presentation code remain informative during asset failures.
func TestLocaleReportsMissingTablesAndPreservesKey(t *testing.T) {
	t.Parallel()

	locale := New(fstest.MapFS{}, "English")
	value, err := locale.Text("missing")

	if value != "missing" {
		t.Fatalf("Text value = %q, want %q", value, "missing")
	}

	if err == nil || !strings.Contains(err.Error(), "no English string tables") {
		t.Fatalf("Text error = %v, want missing-table error", err)
	}
}

// TestLocaleLoadsVersionOneTablesAndReportsWinningPatchSource verifies both overlay precedence and attribution.
// Source reporting must advance with the value so diagnostics never identify a shadowed lower-precedence table.
func TestLocaleLoadsVersionOneTablesAndReportsWinningPatchSource(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"data/local/lng/eng/string.tbl": {
			Data: encodeVersionOneStringTable(map[string]string{"skill": "base %s"}),
		},
		"data/local/lng/eng/patchstring.tbl": {
			Data: encodeVersionOneStringTable(map[string]string{"skill": "patch %+d"}),
		},
	}
	locale := New(source, "English")

	value, path, err := locale.Resolve("skill")
	if err != nil {
		t.Fatalf("Resolve returned an error: %v", err)
	}

	if value != "patch %+d" {
		t.Fatalf("Resolve value = %q, want %q", value, "patch %+d")
	}

	if path != "data/local/lng/eng/patchstring.tbl" {
		t.Fatalf("Resolve path = %q, want patch table", path)
	}
}

// TestLocaleBuffersReaderAtTablesBeforeDecoding verifies that archive-backed tables are decoded from memory.
// Avoiding decoder ReaderAt calls prevents small logical reads from multiplying compressed MPQ archive operations.
func TestLocaleBuffersReaderAtTablesBeforeDecoding(t *testing.T) {
	t.Parallel()

	source := &readerAtCountingFS{MapFS: fstest.MapFS{
		"data/local/lng/eng/string.tbl": {
			Data: encodeVersionOneStringTable(map[string]string{"skill": "buffered"}),
		},
	}}
	locale := New(source, "English")

	value, path, err := locale.Resolve("skill")
	if err != nil {
		t.Fatalf("Resolve returned an error: %v", err)
	}

	if value != "buffered" {
		t.Fatalf("Resolve value = %q, want %q", value, "buffered")
	}

	if path != "data/local/lng/eng/string.tbl" {
		t.Fatalf("Resolve path = %q, want base table", path)
	}

	if source.readAtCalls != 0 {
		t.Fatalf("ReaderAt calls = %d, want 0", source.readAtCalls)
	}
}

// TestLocaleCachesLoadedStringsUntilInvalidated verifies the lifecycle boundary around mutable content layers.
// Source changes remain invisible until invalidation, after which the next lookup must build a fresh generation.
func TestLocaleCachesLoadedStringsUntilInvalidated(t *testing.T) {
	t.Parallel()

	const path = "locales/English.json"

	source := fstest.MapFS{
		path: {Data: []byte(`{"title":"first"}`)},
	}
	locale := New(source, "English")

	assertLocalizedText(t, locale, "title", "first")

	source[path] = &fstest.MapFile{Data: []byte(`{"title":"second"}`)}

	assertLocalizedText(t, locale, "title", "first")

	locale.Invalidate()
	assertLocalizedText(t, locale, "title", "second")
}

// assertLocalizedText requires a successful text lookup and reports value mismatches at the calling test line.
// Centralizing this lifecycle assertion keeps the cache-generation test focused on its state transitions.
func assertLocalizedText(t *testing.T, locale *Locale, key, want string) {
	t.Helper()

	got, err := locale.Text(key)
	if err != nil {
		t.Fatalf("Text(%q) returned an error: %v", key, err)
	}

	if got != want {
		t.Fatalf("Text(%q) = %q, want %q", key, got, want)
	}
}
