package modcache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

type Paths struct {
	Cache   string
	Profile string
}

// DefaultPaths resolves environment overrides or platform defaults and expands
// host aliases consistently. Callers receive alias-free paths ready for host
// filesystem I/O; relative overrides intentionally remain relative.
func DefaultPaths() (Paths, error) {
	cacheRoot := strings.TrimSpace(os.Getenv("DARK_MAGIC_MOD_CACHE"))
	if cacheRoot == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return Paths{}, fmt.Errorf("modcache: resolve user cache directory: %w", err)
		}

		cacheRoot = filepath.Join(base, "dark-magic", "modcache")
	}

	profilePath := strings.TrimSpace(os.Getenv("DARK_MAGIC_MOD_PROFILE"))
	if profilePath == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return Paths{}, fmt.Errorf("modcache: resolve user configuration directory: %w", err)
		}

		profilePath = filepath.Join(base, "dark-magic", "mods.json")
	}

	cacheRoot, err := darkpaths.ExpandHost(cacheRoot)
	if err != nil {
		return Paths{}, fmt.Errorf("modcache: expand cache path: %w", err)
	}

	profilePath, err = darkpaths.ExpandHost(profilePath)
	if err != nil {
		return Paths{}, fmt.Errorf("modcache: expand profile path: %w", err)
	}

	return Paths{Cache: cacheRoot, Profile: profilePath}, nil
}
