// Package assets derives the small, allowlisted set of Realm portal images
// from the user's game archives and stores the results in a private cache.
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing/fstest"

	assetdecode "github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/content"
)

const rendererVersion = "realm-portal-png/v3"

type imageSpec struct {
	path      string
	palette   string
	direction int
	page      int
	combined  bool
}

type presentationManifest struct {
	Schema   string            `json:"schema"`
	Palettes map[string]string `json:"palettes"`
	Screens  struct {
		MainMenu struct {
			Background string `json:"background"`
			Palette    string `json:"palette"`
			Controls   map[string]struct {
				Sheet   string `json:"sheet"`
				Palette string `json:"palette"`
			} `json:"controls"`
		} `json:"main_menu"`
		RealmCommon struct {
			Popup struct {
				Sheet   string `json:"sheet"`
				Palette string `json:"palette"`
			} `json:"popup"`
		} `json:"realm_common"`
	} `json:"screens"`
}

type assetCatalogManifest struct {
	Assets []struct {
		ID      string `json:"id"`
		Path    string `json:"path"`
		Palette string `json:"palette"`
	} `json:"assets"`
}

// Cache serves only named, reviewed assets. It never exposes arbitrary MPQ
// paths or writes archive contents outside its private content-addressed cache.
type Cache struct {
	source    fs.FS
	directory string
	images    map[string]imageSpec
	mu        sync.Mutex
	cached    map[string]string
}

// New creates an allowlisted asset cache rooted in a private directory. The constructor validates the embedded
// presentation recipes up front so configuration errors fail startup instead of surfacing as unrelated HTTP 404s.
func New(source fs.FS, directory string) (*Cache, error) {
	if source == nil || strings.TrimSpace(directory) == "" {
		return nil, errors.New("realm portal assets require content and a cache directory")
	}

	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}

	images, err := loadImageSpecs(content.D2Legacy())
	if err != nil {
		return nil, err
	}

	return &Cache{
		source:    source,
		directory: directory,
		images:    images,
		cached:    make(map[string]string),
	}, nil
}

// ServeHTTP routes only the three reviewed asset namespaces. Refusing all other paths keeps the archive-backed source
// from becoming a general-purpose file server.
func (cache *Cache) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch {
	case strings.HasPrefix(request.URL.Path, "/account/assets/"):
		cache.serveImage(writer, request)
	case strings.HasPrefix(request.URL.Path, "/account/fonts/"):
		cache.serveFont(writer, request)
	case strings.HasPrefix(request.URL.Path, "/account/roster/"):
		cache.serveRoster(writer, request)
	default:
		http.NotFound(writer, request)
	}
}

// serveImage resolves a stable browser-facing alias to a versioned PNG. Render failures intentionally appear as 404s
// so callers cannot distinguish missing allowlist entries from unreadable archive content.
func (cache *Cache) serveImage(writer http.ResponseWriter, request *http.Request) {
	id := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/account/assets/"), ".png")

	path, err := cache.renderImage(id)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	// These names are stable aliases whose content changes when the renderer or
	// presentation manifest changes. Force revalidation instead of pinning stale
	// pixels in the browser. Roster URLs are content-addressed and stay immutable.
	writer.Header().Set("Cache-Control", "private, no-cache")
	writer.Header().Set("Content-Type", "image/png")
	http.ServeFile(writer, request, path)
}

// renderImage materializes an allowlisted DC6 recipe once and memoizes its path. The cache lock covers lookup,
// decoding, and publication so concurrent requests cannot race while replacing the same content-addressed file.
func (cache *Cache) renderImage(id string) (string, error) {
	spec, found := cache.images[id]
	if !found {
		return "", fs.ErrNotExist
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	if path := cache.cached[id]; path != "" {
		return path, nil
	}

	assetData, paletteData, err := cache.readImageInputs(spec)
	if err != nil {
		return "", err
	}

	path := cache.imageCachePath(id, spec, assetData, paletteData)
	if _, err := os.Stat(path); err == nil {
		cache.cached[id] = path
		return path, nil
	}

	decoded, err := decodeImage(id, spec, assetData, paletteData)
	if err != nil {
		return "", err
	}

	if err := writePNG(path, decoded); err != nil {
		return "", err
	}

	cache.cached[id] = path

	return path, nil
}

// readImageInputs reads both authored inputs before hashing or decoding. Keeping the paired read in one helper makes
// it explicit that a rendered file represents an asset and its palette together.
func (cache *Cache) readImageInputs(spec imageSpec) ([]byte, []byte, error) {
	assetData, err := fs.ReadFile(cache.source, spec.path)
	if err != nil {
		return nil, nil, err
	}

	paletteData, err := fs.ReadFile(cache.source, spec.palette)
	if err != nil {
		return nil, nil, err
	}

	return assetData, paletteData, nil
}

// imageCachePath incorporates every pixel-affecting input into the filename. Renderer changes therefore publish a
// new file instead of silently reusing pixels generated by an older algorithm.
func (cache *Cache) imageCachePath(id string, spec imageSpec, assetData, paletteData []byte) string {
	hash := sha256.New()
	_, _ = io.WriteString(hash, rendererVersion)
	_, _ = hash.Write(assetData)
	_, _ = hash.Write(paletteData)
	_, _ = fmt.Fprintf(hash, ":%d:%d:%t", spec.direction, spec.page, spec.combined)
	name := id + "-" + hex.EncodeToString(hash.Sum(nil))[:20] + ".png"

	return filepath.Join(cache.directory, name)
}

// decodeImage converts an in-memory DC6 recipe into the requested page. The temporary filesystem prevents the decoder
// from gaining access to any archive path beyond the two inputs already selected by the allowlist.
func decodeImage(id string, spec imageSpec, assetData, paletteData []byte) (image.Image, error) {
	memory := fstest.MapFS{
		"asset.dc6":   {Data: assetData},
		"palette.dat": {Data: paletteData},
	}

	sheet, err := assetdecode.DC6(memory, "asset.dc6", "palette.dat")
	if err != nil {
		return nil, err
	}

	if spec.combined {
		pages, err := assetdecode.CombinedDC6Pages(sheet, spec.direction)
		if err != nil {
			return nil, err
		}

		if spec.page < 0 || spec.page >= len(pages) {
			return nil, fmt.Errorf("realm portal asset %q page %d is out of range", id, spec.page)
		}

		return pages[spec.page], nil
	}

	frame, err := assetdecode.Frame(sheet, spec.direction, spec.page)
	if err != nil {
		return nil, err
	}

	return assetdecode.FrameImage(sheet, frame)
}

// loadImageSpecs joins the authored presentation manifest with the asset catalog. Cross-checking duplicated recipes
// protects the portal from drifting away from the presentation data used by the game client.
func loadImageSpecs(source fs.FS) (map[string]imageSpec, error) {
	var presentation presentationManifest
	if err := readJSON(source, "manifests/presentation.v1.json", &presentation); err != nil {
		return nil, err
	}

	if presentation.Schema != "d2legacy.presentation/v1" {
		return nil, fmt.Errorf("realm portal assets: unsupported presentation schema %q", presentation.Schema)
	}

	var catalog assetCatalogManifest
	if err := readJSON(source, "manifests/asset-catalog.v1.json", &catalog); err != nil {
		return nil, err
	}

	catalogAssets := make(map[string]imageSpec, len(catalog.Assets))
	for _, asset := range catalog.Assets {
		catalogAssets[asset.ID] = imageSpec{path: asset.Path, palette: asset.Palette, combined: true}
	}

	button := presentation.Screens.MainMenu.Controls["realm"]
	background := imageSpec{
		path:     presentation.Screens.MainMenu.Background,
		palette:  presentation.Palettes[presentation.Screens.MainMenu.Palette],
		combined: true,
	}
	dialog := imageSpec{
		path:     presentation.Screens.RealmCommon.Popup.Sheet,
		palette:  presentation.Palettes[presentation.Screens.RealmCommon.Popup.Palette],
		combined: true,
	}
	buttonUp := imageSpec{
		path:     button.Sheet,
		palette:  presentation.Palettes[button.Palette],
		combined: true,
	}

	authoredCatalogSpecs := map[string]imageSpec{
		"main-game-select-exp": background,
		"popup-340x224":        dialog,
		"button-wide":          buttonUp,
	}
	if err := validateCatalogSpecs(catalogAssets, authoredCatalogSpecs); err != nil {
		return nil, err
	}

	specs := map[string]imageSpec{
		"background": background,
		"dialog":     dialog,
		"textbox":    catalogAssets["text-box-wide"],
		"button":     buttonUp,
		"button-pressed": {
			path:     button.Sheet,
			palette:  presentation.Palettes[button.Palette],
			page:     1,
			combined: true,
		},
	}
	for id, spec := range specs {
		if spec.path == "" || spec.palette == "" {
			return nil, fmt.Errorf("realm portal assets: incomplete %q presentation recipe", id)
		}
	}

	return specs, nil
}

// validateCatalogSpecs confirms that duplicated presentation and catalog entries resolve to identical source files.
// A mismatch is rejected because choosing either source implicitly would make the rendered UI depend on load order.
func validateCatalogSpecs(catalog, authored map[string]imageSpec) error {
	for catalogID, authoredSpec := range authored {
		catalogSpec := catalog[catalogID]
		if catalogSpec.path == authoredSpec.path && catalogSpec.palette == authoredSpec.palette {
			continue
		}

		return fmt.Errorf(
			"realm portal assets: catalog %q (%s, %s) disagrees with presentation (%s, %s)",
			catalogID,
			catalogSpec.path,
			catalogSpec.palette,
			authoredSpec.path,
			authoredSpec.palette,
		)
	}

	return nil
}

// readJSON adds the manifest path to read and decode failures so startup errors identify the authored input that must
// be repaired rather than exposing a context-free filesystem or JSON error.
func readJSON(source fs.FS, path string, destination any) error {
	data, err := fs.ReadFile(source, path)
	if err != nil {
		return fmt.Errorf("realm portal assets: read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, destination); err != nil {
		return fmt.Errorf("realm portal assets: decode %s: %w", path, err)
	}

	return nil
}

// writePrivateFile encodes into a private temporary path and renames only after closing it. Atomic replacement keeps
// readers away from an active write; encoder and chmod failures remain best-effort for historical compatibility.
func writePrivateFile(path string, encode func(io.Writer) error) error {
	directory := filepath.Dir(path)

	temporary, err := os.CreateTemp(directory, ".portal-*")
	if err != nil {
		return err
	}

	temporaryPath := temporary.Name()
	// Best-effort removal clears failed writes; a successful rename already removes the temporary pathname.
	defer func() {
		_ = os.Remove(temporaryPath)
	}()

	if temporary.Chmod(0o600) == nil {
		// Existing callers rely on close and rename errors; preserve their historical best-effort encoder handling.
		_ = encode(temporary)
	}

	if err := temporary.Close(); err != nil {
		return err
	}

	return os.Rename(temporaryPath, path)
}
