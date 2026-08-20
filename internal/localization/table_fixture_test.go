package localization

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/fs"
	"sort"
	"testing/fstest"
)

// readerAtCountingFS exposes ReaderAt while recording whether a consumer bypasses ordinary buffered reads.
// Each test owns an instance, so the intentionally simple counter does not need synchronization.
type readerAtCountingFS struct {
	fstest.MapFS
	readAtCalls int
}

// readerAtCountingFile delegates ordinary file behavior while instrumenting its optional ReaderAt capability.
// The separate in-memory reader gives ReadAt stable random access without changing the wrapped file's read position.
type readerAtCountingFile struct {
	fs.File
	readerAt io.ReaderAt
	calls    *int
}

// Open wraps a fixture file with ReaderAt instrumentation while preserving MapFS open errors unchanged.
// Tests can therefore detect decoder access strategy without substituting different file contents or semantics.
func (source *readerAtCountingFS) Open(name string) (fs.File, error) {
	file, err := source.MapFS.Open(name)
	if err != nil {
		return nil, err
	}

	return &readerAtCountingFile{
		File:     file,
		readerAt: bytes.NewReader(source.MapFS[name].Data),
		calls:    &source.readAtCalls,
	}, nil
}

// ReadAt records each random-access request before delegating it to the immutable fixture bytes.
// A zero count proves production code buffered through fs.File instead of handing ReaderAt to the table decoder.
func (file *readerAtCountingFile) ReadAt(data []byte, offset int64) (int, error) {
	*file.calls++

	return file.readerAt.ReadAt(data, offset)
}

// encodeVersionOneStringTable builds the smallest deterministic Diablo table needed by localization tests.
// Sorted keys keep hash-slot references and resulting bytes stable despite randomized Go map iteration.
func encodeVersionOneStringTable(entries map[string]string) []byte {
	const (
		headerBytes    = 21
		hashEntryBytes = 17
	)

	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	indexBytes := len(entries) * 2
	hashBytes := len(entries) * hashEntryBytes
	data := make([]byte, headerBytes+indexBytes+hashBytes)

	binary.LittleEndian.PutUint16(data[2:4], uint16(len(entries)))
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(entries)))
	data[8] = 1

	for index, key := range keys {
		value := entries[key]

		binary.LittleEndian.PutUint16(data[headerBytes+index*2:], uint16(index))

		entryOffset := headerBytes + indexBytes + index*hashEntryBytes
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
