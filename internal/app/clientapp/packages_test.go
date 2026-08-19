package clientapp

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// TestCloneRuntimePackagesCopiesExtensions verifies that saved startup state
// cannot be mutated through a later recipe's extension slice.
func TestCloneRuntimePackagesCopiesExtensions(t *testing.T) {
	original := simulation.RuntimePackageSet{
		Base:       runtimePackage("base", "base-digest"),
		Extensions: []simulation.RuntimePackage{runtimePackage("extension", "old-digest")},
	}

	cloned := cloneRuntimePackages(original)
	cloned.Extensions[0].Digest = "new-digest"

	if original.Extensions[0].Digest != "old-digest" {
		t.Fatal("clone retained the caller's extension backing array")
	}
}

// TestResolvedMatchesPackagesRequiresExactOrderedMetadata verifies that cache
// resolution cannot substitute a package with merely the same namespace.
func TestResolvedMatchesPackagesRequiresExactOrderedMetadata(t *testing.T) {
	base := lockedPackage("base", "base-digest")
	extension := lockedPackage("extension", "extension-digest")
	resolved := modcache.ResolvedSet{
		Base: base,
		Extensions: modcache.Lock{
			Packages: []modcache.LockedPackage{extension},
		},
	}
	wanted := simulation.RuntimePackageSet{
		Base:       runtimePackage("base", "base-digest"),
		Extensions: []simulation.RuntimePackage{runtimePackage("extension", "extension-digest")},
	}

	if !resolvedMatchesPackages(resolved, wanted) {
		t.Fatal("exact resolved package set did not match")
	}

	wanted.Extensions[0].Digest = "different-digest"
	if resolvedMatchesPackages(resolved, wanted) {
		t.Fatal("resolved package with a different digest matched")
	}
}

// TestChangedPackageIDsIncludesAddedRemovedAndReplaced verifies that module
// invalidation covers every form of package-set transition without duplicates.
func TestChangedPackageIDsIncludesAddedRemovedAndReplaced(t *testing.T) {
	previous := map[string]string{
		"base":    "old-base",
		"removed": "removed-digest",
	}
	next := modcache.ResolvedSet{
		Base: lockedPackage("base", "new-base"),
		Extensions: modcache.Lock{
			Packages: []modcache.LockedPackage{lockedPackage("added", "added-digest")},
		},
	}

	changed := changedPackageIDs(previous, next)
	sort.Strings(changed)

	want := []string{"added", "base", "removed"}

	if !reflect.DeepEqual(changed, want) {
		t.Fatalf("changed packages = %v, want %v", changed, want)
	}
}

// TestNetworkPackagePlanAbortReleasesUntransferredMount verifies that failures
// before the ownership handoff close archives and preserve the original error.
func TestNetworkPackagePlanAbortReleasesUntransferredMount(t *testing.T) {
	mounted := &modcache.MountedSet{
		Packages: []modcache.MountedPackage{{ID: "extension"}},
	}
	plan := &networkPackagePlan{mounted: mounted}
	want := errors.New("failed")

	if err := plan.abort(want); !errors.Is(err, want) {
		t.Fatalf("abort error = %v, want %v", err, want)
	}

	if plan.mounted != nil || len(mounted.Packages) != 0 {
		t.Fatal("abort retained an untransferred mounted package set")
	}
}

// runtimePackage creates transport-neutral metadata for package tests.
func runtimePackage(id, digest string) simulation.RuntimePackage {
	return simulation.RuntimePackage{
		ID:              id,
		Version:         "1.0.0",
		Digest:          digest,
		Size:            42,
		Redistributable: true,
	}
}

// lockedPackage creates cache metadata equivalent to runtimePackage.
func lockedPackage(id, digest string) modcache.LockedPackage {
	return modcache.LockedPackage{
		Descriptor: modcache.Descriptor{
			ID:              id,
			Version:         "1.0.0",
			Digest:          digest,
			Size:            42,
			Redistributable: true,
		},
		Manifest: modcache.Manifest{
			ID:      id,
			Version: "1.0.0",
		},
	}
}
