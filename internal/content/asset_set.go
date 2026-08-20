package content

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

const assetSetSchema = "dark-magic.asset-set/v1"

// assetSetFile is the path-independent identity record for one regular file in a configured root.
type assetSetFile struct {
	Root   int    `json:"root"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
}

// assetSetManifest preserves configured root order while making file order deterministic within the final digest.
type assetSetManifest struct {
	Schema string         `json:"schema"`
	Files  []assetSetFile `json:"files"`
}

// assetDigestCache stores owner-local metadata shortcuts that never enter the shared asset-set identity.
type assetDigestCache struct {
	Schema  string                     `json:"schema"`
	Entries map[string]assetCacheEntry `json:"entries"`
}

// assetCacheEntry reuses a digest only while the file's size and nanosecond modification time still match.
type assetCacheEntry struct {
	Size       int64  `json:"size"`
	ModifiedNS int64  `json:"modified_ns"`
	Digest     string `json:"digest"`
}

// assetManifestCollector appends regular files from one configured root while retaining its identity-order index.
type assetManifestCollector struct {
	manifest  *assetSetManifest
	cache     *assetDigestCache
	rootIndex int
	directory string
}

// AssetSetIdentityFromEnvironment returns a path-independent digest of every regular file mounted from MPQ_DIRECTORY.
// Absolute paths remain confined to the owner-local metadata cache and never enter the identity or network protocol.
func AssetSetIdentityFromEnvironment() (string, error) {
	configured := strings.TrimSpace(os.Getenv("MPQ_DIRECTORY"))
	if configured == "" {
		return digestAssetManifest(assetSetManifest{Schema: assetSetSchema})
	}

	cachePath, err := assetSetCachePath()
	if err != nil {
		return "", err
	}

	cache := readAssetDigestCache(cachePath)
	manifest := assetSetManifest{Schema: assetSetSchema}

	for rootIndex, entry := range strings.Split(configured, ",") {
		directory, err := resolveAssetSetDirectory(entry, rootIndex)
		if err != nil {
			return "", err
		}

		collector := assetManifestCollector{
			manifest:  &manifest,
			cache:     cache,
			rootIndex: rootIndex,
			directory: directory,
		}

		err = filepath.WalkDir(directory, collector.visit)
		if err != nil {
			return "", fmt.Errorf("content: identify asset-set directory %q: %w", directory, err)
		}
	}

	sort.Slice(manifest.Files, func(i, j int) bool {
		if manifest.Files[i].Root != manifest.Files[j].Root {
			return manifest.Files[i].Root < manifest.Files[j].Root
		}

		return manifest.Files[i].Name < manifest.Files[j].Name
	})

	if err := writeAssetDigestCache(cachePath, cache); err != nil {
		return "", err
	}

	return digestAssetManifest(manifest)
}

// resolveAssetSetDirectory expands and validates one configured root without leaking its absolute path into identity.
func resolveAssetSetDirectory(entry string, rootIndex int) (string, error) {
	directory := strings.TrimSpace(entry)
	if directory == "" {
		return "", fmt.Errorf("content: MPQ_DIRECTORY entry %d is empty", rootIndex+1)
	}

	expanded, err := darkpaths.ExpandHost(directory)
	if err != nil {
		return "", fmt.Errorf("content: expand asset-set directory %q: %w", entry, err)
	}

	directory, err = filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("content: resolve asset-set directory %q: %w", entry, err)
	}

	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("content: asset-set path %q is not a directory", directory)
	}

	return directory, nil
}

// visit records one walked regular file and rejects special files that the mounted filesystem cannot hash reliably.
func (c *assetManifestCollector) visit(name string, item os.DirEntry, walkErr error) error {
	if walkErr != nil {
		return walkErr
	}

	if item.IsDir() {
		return nil
	}

	// Follow file symlinks because os.DirFS follows them when opening a mounted path. Directory symlinks remain
	// untraversed, which prevents a configured root from silently admitting an unbounded external tree.
	info, err := os.Stat(name)
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("content: asset set contains unsupported non-regular file %q", name)
	}

	relative, err := filepath.Rel(c.directory, name)
	if err != nil {
		return err
	}

	relative = strings.ToLower(filepath.ToSlash(relative))

	digest, err := cachedAssetDigest(c.cache, name, info)
	if err != nil {
		return err
	}

	c.manifest.Files = append(c.manifest.Files, assetSetFile{
		Root:   c.rootIndex,
		Name:   relative,
		Size:   info.Size(),
		Digest: digest,
	})

	return nil
}

// cachedAssetDigest returns a reusable metadata-keyed digest or hashes the file and updates the owner-local cache.
func cachedAssetDigest(cache *assetDigestCache, name string, info os.FileInfo) (string, error) {
	key, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}

	if os.Getenv("DARK_MAGIC_ASSET_SET_REHASH") != "1" {
		if value, found := cache.Entries[key]; found && assetCacheEntryMatches(value, info) {
			return value.Digest, nil
		}
	}

	file, err := os.Open(name)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	_, copyErr := io.Copy(hash, file)

	closeErr := file.Close()
	if err := errors.Join(copyErr, closeErr); err != nil {
		return "", err
	}

	digest := "sha256:" + hex.EncodeToString(hash.Sum(nil))
	cache.Entries[key] = assetCacheEntry{Size: info.Size(), ModifiedNS: info.ModTime().UnixNano(), Digest: digest}

	return digest, nil
}

// assetCacheEntryMatches verifies all metadata and syntax required before trusting a cached content digest.
func assetCacheEntryMatches(entry assetCacheEntry, info os.FileInfo) bool {
	return entry.Size == info.Size() && entry.ModifiedNS == info.ModTime().UnixNano() && validAssetDigest(entry.Digest)
}

// digestAssetManifest hashes the stable JSON wire form so every process derives the same path-independent identity.
func digestAssetManifest(manifest assetSetManifest) (string, error) {
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}

	digest := sha256.Sum256(encoded)

	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

// assetSetCachePath resolves an explicit cache override or the owner-local default cache location.
func assetSetCachePath() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("DARK_MAGIC_ASSET_SET_CACHE")); configured != "" {
		return darkpaths.ExpandHost(configured)
	}

	directory, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("content: resolve asset-set cache directory: %w", err)
	}

	return filepath.Join(directory, "dark-magic", "asset-set-v1.json"), nil
}

// readAssetDigestCache treats missing, malformed, and schema-incompatible cache data as an empty performance hint.
// Cache damage therefore cannot prevent authoritative asset bytes from being identified.
func readAssetDigestCache(name string) *assetDigestCache {
	cache := &assetDigestCache{Schema: assetSetSchema, Entries: make(map[string]assetCacheEntry)}

	data, err := os.ReadFile(name)
	if err != nil {
		return cache
	}

	var decoded assetDigestCache
	if json.Unmarshal(data, &decoded) == nil && decoded.Schema == assetSetSchema && decoded.Entries != nil {
		return &decoded
	}

	return cache
}

// writeAssetDigestCache publishes the owner-local hint atomically with private permissions and durable directory state.
func writeAssetDigestCache(name string, cache *assetDigestCache) error {
	if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
		return fmt.Errorf("content: create asset-set cache directory: %w", err)
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(name), ".asset-set-*.tmp")
	if err != nil {
		return err
	}

	temporaryName := temporary.Name()
	defer func() {
		_ = os.Remove(temporaryName)
	}()

	// Cache bytes are only a performance hint, so the existing contract ignores chmod and write failures. Invalid or
	// incomplete cache data is rejected on its next read; sync, close, and publication failures remain reportable.
	if temporary.Chmod(0o600) == nil {
		_, _ = temporary.Write(data)
	}

	err = temporary.Sync()

	err = errors.Join(err, temporary.Close())
	if err != nil {
		return fmt.Errorf("content: write asset-set cache: %w", err)
	}

	if err := os.Rename(temporaryName, name); err != nil {
		return fmt.Errorf("content: publish asset-set cache: %w", err)
	}

	return darkpaths.SyncDirectory(filepath.Dir(name))
}

// validAssetDigest accepts only a complete SHA-256 identity in the versioned textual form used by network contracts.
func validAssetDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}

	_, err := hex.DecodeString(value[len("sha256:"):])

	return err == nil
}
