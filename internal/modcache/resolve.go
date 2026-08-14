package modcache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type Lock struct {
	Schema   string          `json:"schema"`
	Digest   string          `json:"digest"`
	Packages []LockedPackage `json:"packages"`
}

type LockedPackage struct {
	Descriptor Descriptor `json:"descriptor"`
	Manifest   Manifest   `json:"manifest"`
}

func (store *Store) Resolve(profile Profile) (Lock, error) {
	if err := ValidateProfile(profile); err != nil {
		return Lock{}, err
	}
	catalog, err := store.readIndex()
	if err != nil {
		return Lock{}, err
	}
	manifests := make(map[string]Manifest, len(catalog.Packages))
	load := func(id string) (Manifest, error) {
		if manifest, found := manifests[id]; found {
			return manifest, nil
		}
		descriptor, found := catalog.Packages[id]
		if !found {
			return Manifest{}, fmt.Errorf("modcache: enabled or required mod %q is not installed", id)
		}
		manifest, err := store.verifyDescriptor(descriptor)
		if err != nil {
			return Manifest{}, err
		}
		manifests[id] = manifest
		return manifest, nil
	}
	var packages []LockedPackage
	state := make(map[string]uint8)
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("modcache: dependency cycle includes %q", id)
		case 2:
			return nil
		}
		descriptor := catalog.Packages[id]
		manifest, err := load(id)
		if err != nil {
			return err
		}
		state[id] = 1
		for _, dependency := range manifest.Dependencies {
			if err := visit(dependency.ID); err != nil {
				return err
			}
			resolvedDependency := manifests[dependency.ID]
			if dependency.Version != "" && resolvedDependency.Version != dependency.Version {
				return fmt.Errorf("modcache: %s requires %s %s, installed %s", id, dependency.ID, dependency.Version, resolvedDependency.Version)
			}
		}
		state[id] = 2
		packages = append(packages, LockedPackage{Descriptor: descriptor, Manifest: manifest})
		return nil
	}
	for _, id := range profile.Enabled {
		if err := visit(id); err != nil {
			return Lock{}, err
		}
	}
	gameCount := 0
	for _, pkg := range packages {
		if pkg.Manifest.Kind == "game" {
			gameCount++
		}
	}
	if len(packages) > 0 && gameCount != 1 {
		return Lock{}, fmt.Errorf("modcache: resolved set contains %d game packages, want exactly one", gameCount)
	}
	lock := Lock{Schema: LockSchema, Packages: packages}
	digest, err := lockDigest(lock)
	if err != nil {
		return Lock{}, err
	}
	lock.Digest = digest
	return lock, nil
}

// ValidateLock proves that package metadata and activation order still match
// the exact identity emitted by Resolve. Blob verification remains Store's job
// because a lock intentionally contains no local filesystem paths.
func ValidateLock(lock Lock) error {
	if lock.Schema != LockSchema || !validDigest(lock.Digest) {
		return errors.New("modcache: invalid lock")
	}
	want, err := lockDigest(lock)
	if err != nil || want != lock.Digest {
		return errors.New("modcache: lock digest mismatch")
	}
	seen := make(map[string]struct{}, len(lock.Packages))
	gameCount := 0
	for _, pkg := range lock.Packages {
		if err := ValidateManifest(pkg.Manifest); err != nil {
			return fmt.Errorf("modcache: invalid lock manifest: %w", err)
		}
		if pkg.Descriptor.ID != pkg.Manifest.ID || pkg.Descriptor.Version != pkg.Manifest.Version ||
			pkg.Descriptor.Redistributable != pkg.Manifest.Redistributable || pkg.Descriptor.Size <= 0 ||
			!validDigest(pkg.Descriptor.Digest) {
			return fmt.Errorf("modcache: invalid locked package %q", pkg.Manifest.ID)
		}
		if _, duplicate := seen[pkg.Manifest.ID]; duplicate {
			return fmt.Errorf("modcache: duplicate locked package %q", pkg.Manifest.ID)
		}
		seen[pkg.Manifest.ID] = struct{}{}
		if pkg.Manifest.Kind == "game" {
			gameCount++
		}
	}
	if len(lock.Packages) > 0 && gameCount != 1 {
		return fmt.Errorf("modcache: lock contains %d game packages, want exactly one", gameCount)
	}
	return nil
}

func lockDigest(lock Lock) (string, error) {
	lock.Digest = ""
	data, err := json.Marshal(lock)
	if err != nil {
		return "", errors.New("modcache: encode resolved lock")
	}
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func (lock Lock) ClientComponents() []string {
	var result []string
	for _, pkg := range lock.Packages {
		result = append(result, pkg.Manifest.Entrypoints.ClientComponents...)
	}
	return result
}

func (lock Lock) AuthorityComponents() []string {
	var result []string
	for _, pkg := range lock.Packages {
		result = append(result, pkg.Manifest.Entrypoints.AuthorityComponents...)
	}
	return result
}
