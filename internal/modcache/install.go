package modcache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// InstallVerified accepts bytes only for an authenticated descriptor. Data
// remains in quarantine until archive, manifest, size, and digest validation
// all succeed, then becomes an immutable cache blob and indexed extension.
func (store *Store) InstallVerified(ctx context.Context, source io.Reader, expected Descriptor) (Manifest, error) {
	if ctx == nil || source == nil {
		return Manifest{}, errors.New("modcache: install context and source are required")
	}
	if !validID(expected.ID) || strings.TrimSpace(expected.Version) == "" || expected.Size <= 0 ||
		expected.Size > maxBundledPackageBytes || !validDigest(expected.Digest) {
		return Manifest{}, errors.New("modcache: invalid expected package descriptor")
	}
	temporary, err := os.CreateTemp(store.quarantinePath(), ".download-*.zip")
	if err != nil {
		return Manifest{}, fmt.Errorf("modcache: create quarantine download: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	written, copyErr := io.Copy(temporary, io.LimitReader(&contextReader{ctx: ctx, source: source}, maxBundledPackageBytes+1))
	syncErr := temporary.Sync()
	closeErr := temporary.Close()
	if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
		return Manifest{}, fmt.Errorf("modcache: receive quarantined package: %w", err)
	}
	if written != expected.Size {
		return Manifest{}, fmt.Errorf("modcache: downloaded package size %d differs from expected %d", written, expected.Size)
	}
	manifest, err := verifyPackageFile(temporaryPath, expected)
	if err != nil {
		return Manifest{}, err
	}
	if manifest.Kind != "extension" {
		return Manifest{}, fmt.Errorf("modcache: downloaded package %q is not an extension", manifest.ID)
	}
	if manifest.ID != expected.ID || manifest.Version != expected.Version || manifest.Redistributable != expected.Redistributable {
		return Manifest{}, fmt.Errorf("modcache: downloaded manifest differs from authenticated descriptor for %q", expected.ID)
	}
	err = store.withMutationLock(func() error {
		destination := store.blobPath(expected.Digest)
		if _, statErr := os.Stat(destination); statErr == nil {
			if _, verifyErr := verifyPackageFile(destination, expected); verifyErr != nil {
				return fmt.Errorf("modcache: existing immutable package differs: %w", verifyErr)
			}
		} else if !os.IsNotExist(statErr) {
			return fmt.Errorf("modcache: inspect immutable package: %w", statErr)
		} else if renameErr := os.Rename(temporaryPath, destination); renameErr != nil {
			return fmt.Errorf("modcache: promote verified package: %w", renameErr)
		}
		catalog, readErr := store.readIndex()
		if readErr != nil {
			return readErr
		}
		// The profile index chooses one default descriptor per package ID. Keep
		// that choice stable when a session downloads another exact version; the
		// content-addressed blob remains available to ResolveExact and may safely
		// coexist with other sessions using the same package ID.
		if selected, found := catalog.Packages[expected.ID]; !found || selected == expected {
			catalog.Packages[expected.ID] = expected
		}
		return writeJSONAtomic(store.indexPath(), catalog)
	})
	if err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (store *Store) quarantinePath() string { return filepath.Join(store.root, "quarantine") }

type contextReader struct {
	ctx    context.Context
	source io.Reader
}

func (reader *contextReader) Read(destination []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	return reader.source.Read(destination)
}
