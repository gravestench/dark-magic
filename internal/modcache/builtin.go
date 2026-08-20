package modcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// DescribeBuiltin gives an immutable distribution-owned package the same
// content identity shape used by cached extensions without serializing it into
// the user's cache. The digest is over canonical package-relative file names
// and distribution source bytes, not an archive compressor's output.
func DescribeBuiltin(source fs.FS) (LockedPackage, error) {
	if source == nil {
		return LockedPackage{}, fmt.Errorf("modcache: built-in package filesystem is required")
	}

	manifest, err := ReadManifest(source)
	if err != nil {
		return LockedPackage{}, err
	}

	if manifest.Kind != "game" {
		return LockedPackage{}, fmt.Errorf("modcache: built-in package %q must be a game", manifest.ID)
	}

	names := make([]string, 0, 256)

	err = fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !entry.IsDir() {
			names = append(names, name)
		}

		return nil
	})
	if err != nil {
		return LockedPackage{}, fmt.Errorf("modcache: walk built-in package %q: %w", manifest.ID, err)
	}

	sort.Strings(names)

	hash := sha256.New()

	var (
		total   int64
		encoded [8]byte
	)

	for _, name := range names {
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return LockedPackage{}, fmt.Errorf("modcache: read built-in package %q: %w", name, err)
		}

		data = CanonicalBuiltinSource(name, data)
		binary.BigEndian.PutUint64(encoded[:], uint64(len(name)))
		_, _ = hash.Write(encoded[:])
		_, _ = hash.Write([]byte(name))

		binary.BigEndian.PutUint64(encoded[:], uint64(len(data)))
		_, _ = hash.Write(encoded[:])
		_, _ = hash.Write(data)
		total += int64(len(name) + len(data) + 16)
	}

	descriptor := Descriptor{
		ID:              manifest.ID,
		Version:         manifest.Version,
		Digest:          "sha256:" + hex.EncodeToString(hash.Sum(nil)),
		Size:            total,
		Redistributable: manifest.Redistributable,
	}

	return LockedPackage{Descriptor: descriptor, Manifest: manifest}, nil
}

// CanonicalBuiltinSource returns the bytes used to identify distribution-owned
// source trees. Git may check text out with CRLF on Windows and LF elsewhere;
// that transport detail must not make otherwise identical d2legacy builds
// network-incompatible. Cached extension archives remain byte-exact and do not
// use this normalization.
func CanonicalBuiltinSource(name string, data []byte) []byte {
	switch strings.ToLower(path.Ext(name)) {
	case ".json", ".lua", ".md", ".txt":
		return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	default:
		return data
	}
}
