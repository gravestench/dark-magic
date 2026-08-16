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

type assetSetFile struct {
	Root   int    `json:"root"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
}

type assetSetManifest struct {
	Schema string         `json:"schema"`
	Files  []assetSetFile `json:"files"`
}

type assetDigestCache struct {
	Schema  string                     `json:"schema"`
	Entries map[string]assetCacheEntry `json:"entries"`
}

type assetCacheEntry struct {
	Size       int64  `json:"size"`
	ModifiedNS int64  `json:"modified_ns"`
	Digest     string `json:"digest"`
}

// AssetSetIdentityFromEnvironment returns a path-independent digest of every
// regular file mounted from MPQ_DIRECTORY. File bytes are hashed on first use;
// an owner-local metadata cache makes subsequent clients and workers cheap.
// Absolute paths never enter the identity or network protocol.
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
		directory := strings.TrimSpace(entry)
		if directory == "" {
			return "", fmt.Errorf("content: MPQ_DIRECTORY entry %d is empty", rootIndex+1)
		}
		directory, err = darkpaths.ExpandHost(directory)
		if err != nil {
			return "", fmt.Errorf("content: expand asset-set directory %q: %w", entry, err)
		}
		directory, err = filepath.Abs(directory)
		if err != nil {
			return "", fmt.Errorf("content: resolve asset-set directory %q: %w", entry, err)
		}
		info, err := os.Stat(directory)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("content: asset-set path %q is not a directory", directory)
		}
		err = filepath.WalkDir(directory, func(name string, item os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if item.IsDir() {
				return nil
			}
			// Follow file symlinks because os.DirFS follows them when opening the
			// mounted path. Directory symlinks remain rejected rather than hiding
			// an unbounded tree outside the configured root.
			info, err := os.Stat(name)
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("content: asset set contains unsupported non-regular file %q", name)
			}
			relative, err := filepath.Rel(directory, name)
			if err != nil {
				return err
			}
			relative = strings.ToLower(filepath.ToSlash(relative))
			digest, err := cachedAssetDigest(cache, name, info)
			if err != nil {
				return err
			}
			manifest.Files = append(manifest.Files, assetSetFile{Root: rootIndex, Name: relative, Size: info.Size(), Digest: digest})
			return nil
		})
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

func cachedAssetDigest(cache *assetDigestCache, name string, info os.FileInfo) (string, error) {
	key, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	if os.Getenv("DARK_MAGIC_ASSET_SET_REHASH") != "1" {
		if value, found := cache.Entries[key]; found && value.Size == info.Size() && value.ModifiedNS == info.ModTime().UnixNano() && validAssetDigest(value.Digest) {
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

func digestAssetManifest(manifest assetSetManifest) (string, error) {
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

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
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err == nil {
		_, err = temporary.Write(data)
	}
	if err == nil {
		err = temporary.Sync()
	}
	err = errors.Join(err, temporary.Close())
	if err != nil {
		return fmt.Errorf("content: write asset-set cache: %w", err)
	}
	if err := os.Rename(temporaryName, name); err != nil {
		return fmt.Errorf("content: publish asset-set cache: %w", err)
	}
	return darkpaths.SyncDirectory(filepath.Dir(name))
}

func validAssetDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}
