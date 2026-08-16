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
	defer os.Remove(temporaryPath)
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
	descriptor := Descriptor{ID: manifest.ID, Version: manifest.Version, Digest: "sha256:" + digestHex,
		Size: limited.written, Redistributable: manifest.Redistributable}
	destination := store.blobPath(descriptor.Digest)
	if info, statErr := os.Stat(destination); statErr == nil {
		if info.Size() != descriptor.Size {
			corrupt := filepath.Join(store.root, "quarantine", ".corrupt-"+digestHex+"-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".zip")
			if err := os.Rename(destination, corrupt); err != nil {
				return Descriptor{}, fmt.Errorf("modcache: quarantine corrupt bundled blob: %w", err)
			}
			if err := os.Rename(temporaryPath, destination); err != nil {
				return Descriptor{}, fmt.Errorf("modcache: restore bundled package: %w", err)
			}
		} else if _, verifyErr := verifyPackageFile(destination, descriptor); verifyErr != nil {
			corrupt := filepath.Join(store.root, "quarantine", ".corrupt-"+digestHex+"-"+strconv.FormatInt(time.Now().UnixNano(), 10)+".zip")
			if err := os.Rename(destination, corrupt); err != nil {
				return Descriptor{}, fmt.Errorf("modcache: quarantine corrupt bundled blob: %w", err)
			}
			if err := os.Rename(temporaryPath, destination); err != nil {
				return Descriptor{}, fmt.Errorf("modcache: restore bundled package: %w", err)
			}
		}
	} else if !os.IsNotExist(statErr) {
		return Descriptor{}, fmt.Errorf("modcache: inspect immutable blob: %w", statErr)
	} else if err := os.Rename(temporaryPath, destination); err != nil {
		return Descriptor{}, fmt.Errorf("modcache: promote bundled package: %w", err)
	}
	if _, err := store.verifyDescriptor(descriptor); err != nil {
		return Descriptor{}, err
	}
	return descriptor, nil
}

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
		header.SetModTime(archiveTimestamp)
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
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) || result.Schema != IndexSchema || result.Packages == nil {
		return index{}, errors.New("modcache: invalid index")
	}
	for id, descriptor := range result.Packages {
		if id != descriptor.ID || !validID(id) || descriptor.Size <= 0 || !validDigest(descriptor.Digest) {
			return index{}, errors.New("modcache: invalid index descriptor")
		}
	}
	return result, nil
}

func (store *Store) verifyDescriptor(descriptor Descriptor) (Manifest, error) {
	if !validID(descriptor.ID) || descriptor.Size <= 0 || descriptor.Size > maxBundledPackageBytes || !validDigest(descriptor.Digest) {
		return Manifest{}, errors.New("modcache: invalid package descriptor")
	}
	return verifyPackageFile(store.blobPath(descriptor.Digest), descriptor)
}

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
	if manifest.ID != descriptor.ID || manifest.Version != descriptor.Version || manifest.Redistributable != descriptor.Redistributable {
		return Manifest{}, fmt.Errorf("modcache: package %s descriptor differs from its manifest", descriptor.ID)
	}
	return manifest, nil
}

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
		if file.UncompressedSize64 > uint64(maxPackageFileBytes) || total > uint64(maxPackageUncompressedBytes)-file.UncompressedSize64 {
			return errors.New("archive exceeds uncompressed size limit")
		}
		total += file.UncompressedSize64
	}
	return nil
}

func (store *Store) indexPath() string { return filepath.Join(store.root, "index.json") }

func (store *Store) blobPath(digest string) string {
	return filepath.Join(store.root, "blobs", "sha256", strings.TrimPrefix(digest, "sha256:")+".zip")
}

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

func (writer *limitedWriter) Write(data []byte) (int, error) {
	if int64(len(data)) > writer.remaining {
		return 0, errors.New("modcache: bundled package exceeds size limit")
	}
	n, err := writer.writer.Write(data)
	writer.remaining -= int64(n)
	writer.written += int64(n)
	return n, err
}

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
	defer os.Remove(temporaryPath)
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
