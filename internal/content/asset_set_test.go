package content

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAssetSetIdentityPinsBytesOrderAndNotAbsolutePaths proves identity follows content bytes rather than the host
// location, which keeps copied installations compatible.
func TestAssetSetIdentityPinsBytesOrderAndNotAbsolutePaths(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	for _, directory := range []string{first, second} {
		if err := os.WriteFile(filepath.Join(directory, "d2data.mpq"), []byte("same bytes"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("DARK_MAGIC_ASSET_SET_CACHE", filepath.Join(t.TempDir(), "cache.json"))
	t.Setenv("MPQ_DIRECTORY", first)

	firstID, err := AssetSetIdentityFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("MPQ_DIRECTORY", second)

	secondID, err := AssetSetIdentityFromEnvironment()
	if err != nil || secondID != firstID {
		t.Fatalf("copied asset identity = %q, %v; want %q", secondID, err, firstID)
	}

	if err := os.WriteFile(filepath.Join(second, "d2data.mpq"), []byte("different bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	changedID, err := AssetSetIdentityFromEnvironment()
	if err != nil || changedID == firstID {
		t.Fatalf("changed asset identity = %q, %v; want a different digest", changedID, err)
	}
}

// TestAssetSetIdentityPinsRootOrderAndHasEmptyIdentity preserves root priority and a valid empty configuration digest.
func TestAssetSetIdentityPinsRootOrderAndHasEmptyIdentity(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(first, "one"), []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(second, "two"), []byte("two"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DARK_MAGIC_ASSET_SET_CACHE", filepath.Join(t.TempDir(), "cache.json"))
	t.Setenv("MPQ_DIRECTORY", first+","+second)

	forward, err := AssetSetIdentityFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("MPQ_DIRECTORY", second+","+first)

	reverse, err := AssetSetIdentityFromEnvironment()
	if err != nil || reverse == forward {
		t.Fatalf("reordered identity = %q, %v; want different from %q", reverse, err, forward)
	}

	t.Setenv("MPQ_DIRECTORY", "")

	empty, err := AssetSetIdentityFromEnvironment()
	if err != nil || !validAssetDigest(empty) {
		t.Fatalf("empty asset identity = %q, %v", empty, err)
	}
}
