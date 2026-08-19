package assetdecode

import (
	"encoding/binary"
	"errors"
	"io"
	"io/fs"
	"time"
)

// readerAtFS owns immutable codec fixtures whose files expose ReaderAt while
// intentionally rejecting sequential access.
type readerAtFS map[string][]byte

// Open returns a file that supports random access but rejects sequential reads,
// making tests prove that capable archive entries take the efficient codec path.
func (f readerAtFS) Open(name string) (fs.File, error) {
	data, ok := f[name]
	if !ok {
		return nil, fs.ErrNotExist
	}

	return &readerAtFile{data: data, name: name}, nil
}

// readerAtFile models the capability combination that selects the random-access
// decoder without involving an operating-system file.
type readerAtFile struct {
	data []byte
	name string
}

// Read deliberately fails so a passing test cannot accidentally exercise the
// sequential fallback while claiming to cover random-access decoding.
func (f *readerAtFile) Read([]byte) (int, error) {
	return 0, errors.New("sequential read disabled")
}

// ReadAt exposes deterministic byte ranges and mirrors io.ReaderAt's EOF
// contract when a requested range extends beyond the fixture.
func (f *readerAtFile) ReadAt(buffer []byte, offset int64) (int, error) {
	if offset < 0 || offset >= int64(len(f.data)) {
		return 0, io.EOF
	}

	read := copy(buffer, f.data[offset:])
	if read != len(buffer) {
		return read, io.EOF
	}

	return read, nil
}

// Close is a no-op because readerAtFile owns only immutable in-memory fixture data.
func (f *readerAtFile) Close() error {
	return nil
}

// Stat reports the exact fixture size required by random-access codec constructors.
func (f *readerAtFile) Stat() (fs.FileInfo, error) {
	return readerAtInfo{file: f}, nil
}

// readerAtInfo exposes only the stable metadata needed to bound random reads.
type readerAtInfo struct {
	file *readerAtFile
}

// Name preserves the lookup key so diagnostics identify the same fixture the test opened.
func (i readerAtInfo) Name() string {
	return i.file.name
}

// Size exposes the complete immutable buffer and lets codecs bound ReaderAt requests.
func (i readerAtInfo) Size() int64 {
	return int64(len(i.file.data))
}

// Mode reports a regular fixture file because no permission semantics are under test.
func (i readerAtInfo) Mode() fs.FileMode {
	return 0
}

// ModTime returns a stable zero value so fixture metadata cannot make tests time-dependent.
func (i readerAtInfo) ModTime() time.Time {
	return time.Time{}
}

// IsDir keeps codec paths on regular-file behavior for this focused fixture.
func (i readerAtInfo) IsDir() bool {
	return false
}

// Sys omits platform metadata that has no bearing on codec capability selection.
func (i readerAtInfo) Sys() any {
	return nil
}

// onePixelDC6 constructs the smallest valid indexed frame, allowing tests to
// distinguish palette and access-path behavior without opaque binary fixtures.
func onePixelDC6(index byte) []byte {
	data := make([]byte, 16+8+4+32+3+3)
	putTestUint32(data, 0, 6)
	putTestUint32(data, 16, 1)
	putTestUint32(data, 20, 1)
	putTestUint32(data, 24, 28)
	putTestUint32(data, 32, 1)
	putTestUint32(data, 36, 1)
	putTestUint32(data, 56, 3)
	data[60], data[61], data[62] = 1, index, 0x80

	return data
}

// putTestUint32 makes binary fixture offsets visible at call sites without
// repeating endian conversion mechanics throughout scenario setup.
func putTestUint32(data []byte, offset int, value uint32) {
	binary.LittleEndian.PutUint32(data[offset:offset+4], value)
}
