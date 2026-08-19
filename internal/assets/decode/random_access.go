package assetdecode

import (
	"io"
	"io/fs"
)

// randomAccess reports the capabilities required by codecs that can decode
// directly from an archive entry, avoiding a full-file allocation when safe.
func randomAccess(file fs.File) (io.ReaderAt, int64, bool) {
	reader, ok := file.(io.ReaderAt)
	if !ok {
		return nil, 0, false
	}

	info, err := file.Stat()
	if err != nil || info.Size() < 0 {
		return nil, 0, false
	}

	return reader, info.Size(), true
}
