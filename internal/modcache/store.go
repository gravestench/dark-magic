package modcache

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxBundledPackageBytes      int64 = 256 << 20
	maxPackageUncompressedBytes int64 = 1 << 30
	maxPackageFileBytes         int64 = 256 << 20
	maxPackageFiles                   = 32768
)

var archiveTimestamp = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

type Store struct {
	root     string
	mutation sync.Mutex
}

type Bundle struct {
	Source         fs.FS
	DefaultEnabled bool
}

type Descriptor struct {
	ID              string `json:"id"`
	Version         string `json:"version"`
	Digest          string `json:"digest"`
	Size            int64  `json:"size"`
	Redistributable bool   `json:"redistributable"`
}

type index struct {
	Schema   string                `json:"schema"`
	Packages map[string]Descriptor `json:"packages"`
}

// New prepares one cache root with blob and quarantine directories, requesting
// owner-only permissions for directories it creates. A successful Store has
// every directory required by later mutations.
func New(root string) (*Store, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("modcache: root is required")
	}

	if err := os.MkdirAll(filepath.Join(root, "blobs", "sha256"), 0o700); err != nil {
		return nil, fmt.Errorf("modcache: create blob directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(root, "quarantine"), 0o700); err != nil {
		return nil, fmt.Errorf("modcache: create quarantine directory: %w", err)
	}

	return &Store{root: root}, nil
}

// ReconcileBundled installs distribution packages and atomically updates their
// selected descriptors. Defaults are returned in bundle order so first-run
// profile creation remains deterministic.
func (store *Store) ReconcileBundled(bundles []Bundle) ([]string, error) {
	var defaults []string

	err := store.withMutationLock(func() error {
		catalog, err := store.readIndex()
		if err != nil {
			return err
		}

		defaults = make([]string, 0, len(bundles))
		for _, bundle := range bundles {
			if bundle.Source == nil {
				return errors.New("modcache: bundled source is required")
			}

			descriptor, err := store.installBundled(bundle.Source)
			if err != nil {
				return err
			}

			catalog.Packages[descriptor.ID] = descriptor
			if bundle.DefaultEnabled {
				defaults = append(defaults, descriptor.ID)
			}
		}

		if err := writeJSONAtomic(store.indexPath(), catalog); err != nil {
			return fmt.Errorf("modcache: write index: %w", err)
		}

		return nil
	})

	return defaults, err
}

// installBundled creates deterministic archive bytes, repairs any corrupt blob
// already stored at that digest, and verifies the promoted immutable package.
// The caller must hold the mutation lock while this function runs.
func (store *Store) installBundled(source fs.FS) (Descriptor, error) {
	manifest, err := ReadManifest(source)
	if err != nil {
		return Descriptor{}, err
	}

	temporary, err := os.CreateTemp(filepath.Join(store.root, "quarantine"), ".bundle-*.zip")
	if err != nil {
		return Descriptor{}, fmt.Errorf("modcache: create quarantine package: %w", err)
	}

	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()

	hash := sha256.New()

	limited := &limitedWriter{writer: io.MultiWriter(temporary, hash), remaining: maxBundledPackageBytes}
	if err := writeArchive(limited, source); err != nil {
		_ = temporary.Close()
		return Descriptor{}, err
	}

	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return Descriptor{}, fmt.Errorf("modcache: sync bundled package: %w", err)
	}

	if err := temporary.Close(); err != nil {
		return Descriptor{}, fmt.Errorf("modcache: close bundled package: %w", err)
	}

	digestHex := hex.EncodeToString(hash.Sum(nil))
	descriptor := Descriptor{
		ID:              manifest.ID,
		Version:         manifest.Version,
		Digest:          "sha256:" + digestHex,
		Size:            limited.written,
		Redistributable: manifest.Redistributable,
	}

	destination := store.blobPath(descriptor.Digest)
	if err := store.promoteBundledPackage(temporaryPath, destination, digestHex, descriptor); err != nil {
		return Descriptor{}, err
	}

	if _, err := store.verifyDescriptor(descriptor); err != nil {
		return Descriptor{}, err
	}

	return descriptor, nil
}

// promoteBundledPackage preserves a valid immutable blob and replaces only a
// missing or corrupt one. Corrupt bytes are retained in quarantine for diagnosis
// instead of being destroyed during automatic repair.
func (store *Store) promoteBundledPackage(
	temporaryPath string,
	destination string,
	digestHex string,
	descriptor Descriptor,
) error {
	info, statErr := os.Stat(destination)
	if os.IsNotExist(statErr) {
		if err := os.Rename(temporaryPath, destination); err != nil {
			return fmt.Errorf("modcache: promote bundled package: %w", err)
		}

		return nil
	}

	if statErr != nil {
		return fmt.Errorf("modcache: inspect immutable blob: %w", statErr)
	}

	if info.Size() == descriptor.Size {
		if _, verifyErr := verifyPackageFile(destination, descriptor); verifyErr == nil {
			return nil
		}
	}

	corrupt := store.corruptBlobPath(digestHex)
	if err := os.Rename(destination, corrupt); err != nil {
		return fmt.Errorf("modcache: quarantine corrupt bundled blob: %w", err)
	}

	if err := os.Rename(temporaryPath, destination); err != nil {
		return fmt.Errorf("modcache: restore bundled package: %w", err)
	}

	return nil
}

// corruptBlobPath gives every quarantined repair a collision-resistant name so
// concurrent or repeated diagnostics do not overwrite prior evidence.
func (store *Store) corruptBlobPath(digestHex string) string {
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	name := ".corrupt-" + digestHex + "-" + timestamp + ".zip"

	return filepath.Join(store.root, "quarantine", name)
}

// writeArchive emits package files in filesystem traversal order using stable
// metadata. The caller hashes these exact bytes as the content-addressed blob.
func writeArchive(destination io.Writer, source fs.FS) error {
	writer := zip.NewWriter(destination)

	err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}

		data, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}

		header := &zip.FileHeader{Name: path.Clean(name), Method: zip.Deflate}
		// SetModTime also populates legacy DOS timestamp fields. Replacing it with
		// Modified would change archive bytes and every existing package digest.
		header.SetModTime(archiveTimestamp) //nolint:staticcheck
		header.SetMode(0o644)

		file, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = file.Write(data)

		return err
	})
	if err != nil {
		_ = writer.Close()
		return fmt.Errorf("modcache: archive bundled package: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("modcache: close bundled archive: %w", err)
	}

	return nil
}

// readIndex strictly decodes the selected-descriptor catalog. A missing file is
// the one valid empty-cache representation; malformed contents fail closed.
func (store *Store) readIndex() (index, error) {
	result := index{Schema: IndexSchema, Packages: make(map[string]Descriptor)}

	data, err := os.ReadFile(store.indexPath())
	if os.IsNotExist(err) {
		return result, nil
	}

	if err != nil {
		return index{}, fmt.Errorf("modcache: read index: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&result); err != nil {
		return index{}, errors.New("modcache: invalid index")
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) ||
		result.Schema != IndexSchema ||
		result.Packages == nil {
		return index{}, errors.New("modcache: invalid index")
	}

	for id, descriptor := range result.Packages {
		if id != descriptor.ID || !validID(id) || descriptor.Size <= 0 || !validDigest(descriptor.Digest) {
			return index{}, errors.New("modcache: invalid index descriptor")
		}
	}

	return result, nil
}

// verifyDescriptor validates metadata before opening the content-addressed blob,
// preventing malformed digests from influencing cache paths.
func (store *Store) verifyDescriptor(descriptor Descriptor) (Manifest, error) {
	if !validID(descriptor.ID) ||
		descriptor.Size <= 0 ||
		descriptor.Size > maxBundledPackageBytes ||
		!validDigest(descriptor.Digest) {
		return Manifest{}, errors.New("modcache: invalid package descriptor")
	}

	return verifyPackageFile(store.blobPath(descriptor.Digest), descriptor)
}

// verifyPackageFile authenticates size and digest before parsing archive or
// manifest contents, then proves the embedded metadata matches the descriptor.
func verifyPackageFile(fileName string, descriptor Descriptor) (Manifest, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return Manifest{}, fmt.Errorf("modcache: open %s: %w", descriptor.ID, err)
	}

	hash := sha256.New()
	size, copyErr := io.Copy(hash, io.LimitReader(file, maxBundledPackageBytes+1))

	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		return Manifest{}, fmt.Errorf("modcache: verify %s: %w", descriptor.ID, errors.Join(copyErr, closeErr))
	}

	actual := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	if size != descriptor.Size || actual != descriptor.Digest {
		return Manifest{}, fmt.Errorf("modcache: package %s failed digest or size verification", descriptor.ID)
	}

	archive, err := zip.OpenReader(fileName)
	if err != nil {
		return Manifest{}, fmt.Errorf("modcache: open package %s: %w", descriptor.ID, err)
	}

	if err := validateArchive(&archive.Reader); err != nil {
		_ = archive.Close()
		return Manifest{}, fmt.Errorf("modcache: validate package %s: %w", descriptor.ID, err)
	}

	manifest, manifestErr := ReadManifest(&archive.Reader)

	closeErr = archive.Close()
	if manifestErr != nil || closeErr != nil {
		return Manifest{}, errors.Join(manifestErr, closeErr)
	}

	if manifest.ID != descriptor.ID ||
		manifest.Version != descriptor.Version ||
		manifest.Redistributable != descriptor.Redistributable {
		return Manifest{}, fmt.Errorf("modcache: package %s descriptor differs from its manifest", descriptor.ID)
	}

	return manifest, nil
}

// validateArchive rejects traversal, duplicate names, special files, unsupported
// compression, and decompression bombs before any archive is mounted.
func validateArchive(archive *zip.Reader) error {
	if len(archive.File) == 0 || len(archive.File) > maxPackageFiles {
		return errors.New("invalid archive file count")
	}

	seen := make(map[string]struct{}, len(archive.File))

	var total uint64

	for _, file := range archive.File {
		name := strings.TrimSuffix(file.Name, "/")
		if name == "" || name != file.Name || !fs.ValidPath(name) || file.FileInfo().Mode()&fs.ModeType != 0 {
			return fmt.Errorf("invalid archive entry %q", file.Name)
		}

		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("duplicate archive entry %q", name)
		}

		seen[name] = struct{}{}
		if file.Method != zip.Store && file.Method != zip.Deflate {
			return fmt.Errorf("unsupported compression for %q", name)
		}

		tooLarge := file.UncompressedSize64 > uint64(maxPackageFileBytes)

		exceedsTotalLimit := file.UncompressedSize64 > uint64(maxPackageUncompressedBytes) ||
			total > uint64(maxPackageUncompressedBytes)-file.UncompressedSize64
		if tooLarge || exceedsTotalLimit {
			return errors.New("archive exceeds uncompressed size limit")
		}

		total += file.UncompressedSize64
	}

	return nil
}

// indexPath locates the mutable selected-descriptor catalog beneath one store.
func (store *Store) indexPath() string { return filepath.Join(store.root, "index.json") }

// blobPath maps an authenticated digest to its immutable content-addressed file.
func (store *Store) blobPath(digest string) string {
	return filepath.Join(store.root, "blobs", "sha256", strings.TrimPrefix(digest, "sha256:")+".zip")
}

// validDigest accepts only lowercase/uppercase hexadecimal SHA-256 text with the
// required algorithm prefix, preventing path-shaped digest values.
func validDigest(digest string) bool {
	value := strings.TrimPrefix(digest, "sha256:")
	if value == digest || len(value) != sha256.Size*2 {
		return false
	}

	_, err := hex.DecodeString(value)

	return err == nil
}

type limitedWriter struct {
	writer    io.Writer
	remaining int64
	written   int64
}

// Write enforces the compressed archive byte ceiling while tracking the exact
// promoted size used in the descriptor.
func (writer *limitedWriter) Write(data []byte) (int, error) {
	if int64(len(data)) > writer.remaining {
		return 0, errors.New("modcache: bundled package exceeds size limit")
	}

	n, err := writer.writer.Write(data)
	writer.remaining -= int64(n)
	writer.written += int64(n)

	return n, err
}

// writeJSONAtomic flushes a private same-directory temporary file before rename,
// avoiding publication of partially written JSON. Rename replacement behavior
// remains subject to the host filesystem's guarantees.
func writeJSONAtomic(fileName string, value any) error {
	if err := os.MkdirAll(filepath.Dir(fileName), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	temporary, err := os.CreateTemp(filepath.Dir(fileName), ".modcache-*.json")
	if err != nil {
		return err
	}

	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()

	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}

	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}

	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}

	if err := temporary.Close(); err != nil {
		return err
	}

	return os.Rename(temporaryPath, fileName)
}
