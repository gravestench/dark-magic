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
	"image/png"
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

// Cache serves only named, reviewed assets. It never exposes arbitrary MPQ
// paths or writes archive contents outside its private content-addressed cache.
type Cache struct {
	source    fs.FS
	directory string
	images    map[string]imageSpec
	mu        sync.Mutex
	cached    map[string]string
}

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
		source: source, directory: directory, images: images,
		cached: make(map[string]string),
	}, nil
}

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

	assetData, err := fs.ReadFile(cache.source, spec.path)
	if err != nil {
		return "", err
	}
	paletteData, err := fs.ReadFile(cache.source, spec.palette)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, _ = io.WriteString(hash, rendererVersion)
	_, _ = hash.Write(assetData)
	_, _ = hash.Write(paletteData)
	_, _ = fmt.Fprintf(hash, ":%d:%d:%t", spec.direction, spec.page, spec.combined)
	name := id + "-" + hex.EncodeToString(hash.Sum(nil))[:20] + ".png"
	path := filepath.Join(cache.directory, name)
	if _, err := os.Stat(path); err == nil {
		cache.cached[id] = path
		return path, nil
	}

	memory := fstest.MapFS{
		"asset.dc6":   {Data: assetData},
		"palette.dat": {Data: paletteData},
	}
	sheet, err := assetdecode.DC6(memory, "asset.dc6", "palette.dat")
	if err != nil {
		return "", err
	}
	var decoded image.Image
	if spec.combined {
		pages, err := assetdecode.CombinedDC6Pages(sheet, spec.direction)
		if err != nil {
			return "", err
		}
		if spec.page < 0 || spec.page >= len(pages) {
			return "", fmt.Errorf("realm portal asset %q page %d is out of range", id, spec.page)
		}
		decoded = pages[spec.page]
	} else {
		frame, err := assetdecode.Frame(sheet, spec.direction, spec.page)
		if err != nil {
			return "", err
		}
		decoded, err = assetdecode.FrameImage(sheet, frame)
		if err != nil {
			return "", err
		}
	}
	temporary, err := os.CreateTemp(cache.directory, ".portal-*.png")
	if err != nil {
		return "", err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err == nil {
		err = png.Encode(temporary, decoded)
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return "", err
	}
	cache.cached[id] = path
	return path, nil
}

func loadImageSpecs(source fs.FS) (map[string]imageSpec, error) {
	var presentation struct {
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
	if err := readJSON(source, "manifests/presentation.v1.json", &presentation); err != nil {
		return nil, err
	}
	if presentation.Schema != "d2legacy.presentation/v1" {
		return nil, fmt.Errorf("realm portal assets: unsupported presentation schema %q", presentation.Schema)
	}

	var catalog struct {
		Assets []struct {
			ID      string `json:"id"`
			Path    string `json:"path"`
			Palette string `json:"palette"`
		} `json:"assets"`
	}
	if err := readJSON(source, "manifests/asset-catalog.v1.json", &catalog); err != nil {
		return nil, err
	}
	catalogAssets := make(map[string]imageSpec, len(catalog.Assets))
	for _, asset := range catalog.Assets {
		catalogAssets[asset.ID] = imageSpec{path: asset.Path, palette: asset.Palette, combined: true}
	}

	palette := func(id string) string { return presentation.Palettes[id] }
	button := presentation.Screens.MainMenu.Controls["realm"]
	background := imageSpec{
		path:    presentation.Screens.MainMenu.Background,
		palette: palette(presentation.Screens.MainMenu.Palette), combined: true,
	}
	dialog := imageSpec{
		path:    presentation.Screens.RealmCommon.Popup.Sheet,
		palette: palette(presentation.Screens.RealmCommon.Popup.Palette), combined: true,
	}
	buttonUp := imageSpec{
		path: button.Sheet, palette: palette(button.Palette), combined: true,
	}
	for catalogID, authored := range map[string]imageSpec{
		"main-game-select-exp": background,
		"popup-340x224":        dialog,
		"button-wide":          buttonUp,
	} {
		cataloged := catalogAssets[catalogID]
		if cataloged.path != authored.path || cataloged.palette != authored.palette {
			return nil, fmt.Errorf(
				"realm portal assets: catalog %q (%s, %s) disagrees with presentation (%s, %s)",
				catalogID, cataloged.path, cataloged.palette, authored.path, authored.palette,
			)
		}
	}
	specs := map[string]imageSpec{
		"background": background,
		"dialog":     dialog,
		"textbox":    catalogAssets["text-box-wide"],
		"button":     buttonUp,
		"button-pressed": {
			path: button.Sheet, palette: palette(button.Palette), page: 1, combined: true,
		},
	}
	for id, spec := range specs {
		if spec.path == "" || spec.palette == "" {
			return nil, fmt.Errorf("realm portal assets: incomplete %q presentation recipe", id)
		}
	}
	return specs, nil
}

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

func writePrivateFile(path string, encode func(io.Writer) error) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".portal-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err == nil {
		err = encode(temporary)
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
