package localization

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/fs"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

type countingReaderAtFS struct {
	fstest.MapFS
	readAtCalls int
}

type countingReaderAtFile struct {
	fs.File
	readerAt io.ReaderAt
	calls    *int
}

func (source *countingReaderAtFS) Open(name string) (fs.File, error) {
	file, err := source.MapFS.Open(name)
	if err != nil {
		return nil, err
	}
	return &countingReaderAtFile{
		File:     file,
		readerAt: bytes.NewReader(source.MapFS[name].Data),
		calls:    &source.readAtCalls,
	}, nil
}

func (file *countingReaderAtFile) ReadAt(data []byte, offset int64) (int, error) {
	*file.calls = *file.calls + 1
	return file.readerAt.ReadAt(data, offset)
}

func stringTable(entries map[string]string) []byte {
	const headerBytes, hashEntryBytes = 21, 17
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	data := make([]byte, headerBytes+len(entries)*2+len(entries)*hashEntryBytes)
	binary.LittleEndian.PutUint16(data[2:4], uint16(len(entries)))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(entries)))
	data[8] = 1
	for index, key := range keys {
		value := entries[key]
		binary.LittleEndian.PutUint16(data[headerBytes+index*2:], uint16(index))
		entryOffset := headerBytes + len(entries)*2 + index*hashEntryBytes
		entry := data[entryOffset : entryOffset+hashEntryBytes]
		entry[0] = 1
		keyOffset := len(data)
		valueOffset := keyOffset + len(key) + 1
		binary.LittleEndian.PutUint32(entry[7:11], uint32(keyOffset))
		binary.LittleEndian.PutUint32(entry[11:15], uint32(valueOffset))
		binary.LittleEndian.PutUint16(entry[15:17], uint16(len(value)+1))
		data = append(data, key...)
		data = append(data, 0)
		data = append(data, value...)
		data = append(data, 0)
	}
	return data
}

func TestD2LegacyLocalizationKeysLoadWithoutDiabloTables(t *testing.T) {
	locale := New(content.D2Legacy(), "English")
	value, err := locale.Text("d2legacy.hud.title")
	if err != nil || value != "Dark Magic" {
		t.Fatalf("title = %q, %v", value, err)
	}
	if languages := locale.GetSupportedLanguages(); len(languages) != 1 || languages[0] != "English" {
		t.Fatalf("languages = %v", languages)
	}
}

func TestLocaleReportsMissingTablesAndPreservesKey(t *testing.T) {
	t.Parallel()

	locale := New(fstest.MapFS{}, "English")
	value, err := locale.Text("missing")
	if value != "missing" || err == nil || !strings.Contains(err.Error(), "no English string tables") {
		t.Fatalf("Text = %q/%v", value, err)
	}
}

func TestLocaleLoadsVersionOneTablesAndReportsWinningPatchSource(t *testing.T) {
	t.Parallel()
	source := fstest.MapFS{
		"data/local/lng/eng/string.tbl":      {Data: stringTable(map[string]string{"skill": "base %s"})},
		"data/local/lng/eng/patchstring.tbl": {Data: stringTable(map[string]string{"skill": "patch %+d"})},
	}
	locale := New(source, "English")
	value, path, err := locale.Resolve("skill")
	if err != nil || value != "patch %+d" || path != "data/local/lng/eng/patchstring.tbl" {
		t.Fatalf("Resolve = %q/%q/%v", value, path, err)
	}
}

func TestLocaleBuffersReaderAtTablesBeforeDecoding(t *testing.T) {
	t.Parallel()
	source := &countingReaderAtFS{MapFS: fstest.MapFS{
		"data/local/lng/eng/string.tbl": {Data: stringTable(map[string]string{"skill": "buffered"})},
	}}
	locale := New(source, "English")
	value, path, err := locale.Resolve("skill")
	if err != nil || value != "buffered" || path != "data/local/lng/eng/string.tbl" {
		t.Fatalf("Resolve = %q/%q/%v", value, path, err)
	}
	if source.readAtCalls != 0 {
		t.Fatalf("ReaderAt calls = %d, want 0", source.readAtCalls)
	}
}
