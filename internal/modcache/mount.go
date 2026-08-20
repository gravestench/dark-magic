package modcache

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"reflect"
)

type MountedPackage struct {
	ID string
	FS fs.FS
}

type MountedSet struct {
	Packages []MountedPackage
	closers  []*zip.ReadCloser
}

// Mount opens packages in activation order: dependencies first, followed by
// dependents and then later profile entries. LookupOrder reverses that sequence
// so the most specific package wins resource conflicts.
func (store *Store) Mount(lock Lock, base LockedPackage) (*MountedSet, error) {
	if err := ValidateLock(lock, base); err != nil {
		return nil, err
	}

	set := &MountedSet{}

	for _, pkg := range lock.Packages {
		manifest, err := store.verifyDescriptor(pkg.Descriptor)
		if err != nil {
			_ = set.Close()
			return nil, err
		}

		if !reflect.DeepEqual(manifest, pkg.Manifest) {
			_ = set.Close()
			return nil, fmt.Errorf("modcache: mounted package %s differs from resolved lock", pkg.Manifest.ID)
		}

		archive, err := zip.OpenReader(store.blobPath(pkg.Descriptor.Digest))
		if err != nil {
			_ = set.Close()
			return nil, fmt.Errorf("modcache: mount %s: %w", pkg.Manifest.ID, err)
		}

		set.closers = append(set.closers, archive)
		set.Packages = append(set.Packages, MountedPackage{ID: pkg.Manifest.ID, FS: &archive.Reader})
	}

	return set, nil
}

// LookupOrder returns a reversed copy so dependents and later profile entries
// override dependencies without exposing the mutable activation-order slice.
func (set *MountedSet) LookupOrder() []MountedPackage {
	result := make([]MountedPackage, len(set.Packages))
	for index := range set.Packages {
		result[len(result)-1-index] = set.Packages[index]
	}

	return result
}

// Close releases archives in reverse open order and clears mounted references.
// Joining errors guarantees every handle receives a close attempt.
func (set *MountedSet) Close() error {
	var result error
	for index := len(set.closers) - 1; index >= 0; index-- {
		result = errors.Join(result, set.closers[index].Close())
	}

	set.closers = nil
	set.Packages = nil

	return result
}
