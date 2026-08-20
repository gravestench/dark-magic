package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	cof "github.com/gravestench/cof"
	assetdecode "github.com/gravestench/dark-magic/internal/assets/decode"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	dcc "github.com/gravestench/dcc/pkg"
)

const rosterRendererVersion = "realm-roster-composite/v1"

var rosterClassTokens = map[string]string{
	"Amazon":      "AM",
	"Sorceress":   "SO",
	"Necromancer": "NE",
	"Paladin":     "PA",
	"Barbarian":   "BA",
	"Assassin":    "AI",
	"Druid":       "DZ",
}

type rosterRecipe struct {
	cof        string
	palette    string
	direction  int
	components map[string]string
}

type rosterFrame struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type rosterLayers struct {
	cofLayers  map[cof.CompositeType]cof.CofLayer
	directions map[cof.CompositeType]*dcc.Direction
	palettes   map[cof.CompositeType]color.Palette
}

// RosterDescriptor identifies one content-addressed browser animation. Its ID
// is based only on the resolved visual inputs and their source bytes, so two
// unrelated characters with the same appearance share the same cached files.
type RosterDescriptor struct {
	ID              string        `json:"id"`
	Image           string        `json:"image"`
	Metadata        string        `json:"metadata"`
	FrameDurationMS float64       `json:"frame_duration_ms"`
	Frames          []rosterFrame `json:"frames"`
}

// PrepareRoster resolves and renders one authoritative character appearance.
// Name, account, character ID, level, and stats deliberately do not enter the
// cache identity because they cannot change the pixels.
func (cache *Cache) PrepareRoster(character d2save.Character) (RosterDescriptor, error) {
	recipe, asset, err := cache.resolveRosterRecipe(character)
	if err != nil {
		return RosterDescriptor{}, err
	}

	id, err := cache.rosterID(recipe)
	if err != nil {
		return RosterDescriptor{}, err
	}

	prefix := filepath.Join(cache.directory, "roster-"+id)
	if descriptor, found := cachedRosterDescriptor(prefix); found {
		return descriptor, nil
	}

	// Rendering is expensive, so check without the lock first and repeat after acquiring it. The second check lets a
	// concurrent winner satisfy this request without duplicate decoding while keeping uncontended cache hits cheap.
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if descriptor, found := cachedRosterDescriptor(prefix); found {
		return descriptor, nil
	}

	descriptor, sheet, err := cache.composeRoster(id, recipe, asset)
	if err != nil {
		return RosterDescriptor{}, err
	}

	if err := writeRosterFiles(prefix, descriptor, sheet); err != nil {
		return RosterDescriptor{}, err
	}

	return descriptor, nil
}

// cachedRosterDescriptor returns a descriptor only when its image companion also exists. This prevents an interrupted
// two-file publication from being mistaken for a usable immutable roster entry.
func cachedRosterDescriptor(prefix string) (RosterDescriptor, bool) {
	descriptor, err := readRosterDescriptor(prefix + ".json")
	if err != nil {
		return RosterDescriptor{}, false
	}

	if _, err := os.Stat(prefix + ".png"); err != nil {
		return RosterDescriptor{}, false
	}

	return descriptor, true
}

// writeRosterFiles publishes pixels before the descriptor that advertises them. Each path is atomically replaced, and
// cachedRosterDescriptor requires the pair, so an interruption between publications cannot produce a cache hit.
func writeRosterFiles(prefix string, descriptor RosterDescriptor, sheet image.Image) error {
	if err := writePNG(prefix+".png", sheet); err != nil {
		return err
	}

	metadata, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}

	// writePrivateFile uses best-effort private permissions and replaces the descriptor path only after closing it.
	return writePrivateFile(prefix+".json", func(writer io.Writer) error {
		_, err := writer.Write(metadata)
		return err
	})
}

// resolveRosterRecipe prefers an authoritative saved appearance and otherwise derives the classic class preview. The
// returned COF is loaded from the same recipe so later composition cannot diverge from the inputs used for cache IDs.
func (cache *Cache) resolveRosterRecipe(character d2save.Character) (rosterRecipe, *cof.COF, error) {
	if character.Appearance != nil && strings.TrimSpace(character.Appearance.COF) != "" {
		return cache.resolveAuthoredRosterRecipe(character)
	}

	return cache.resolveDefaultRosterRecipe(character.Class)
}

// resolveAuthoredRosterRecipe normalizes untrusted saved appearance paths before loading anything. Validation confines
// every referenced file to the data namespace and prevents a character save from probing arbitrary filesystem paths.
func (cache *Cache) resolveAuthoredRosterRecipe(character d2save.Character) (rosterRecipe, *cof.COF, error) {
	recipe := rosterRecipe{
		cof:        strings.TrimSpace(character.Appearance.COF),
		palette:    strings.TrimSpace(character.Appearance.Palette),
		direction:  character.Appearance.Direction,
		components: cloneStrings(character.Appearance.Components),
	}
	if recipe.palette == "" {
		recipe.palette = unitsPalette
	}

	if err := validateRosterRecipe(recipe); err != nil {
		return rosterRecipe{}, nil, err
	}

	asset, err := assetdecode.COF(cache.source, recipe.cof)

	return recipe, asset, err
}

// resolveDefaultRosterRecipe recreates the classic town-neutral class preview when no saved appearance is available.
// Missing optional component files are skipped, matching the source archive rather than manufacturing invalid layers.
func (cache *Cache) resolveDefaultRosterRecipe(class string) (rosterRecipe, *cof.COF, error) {
	token := rosterClassTokens[class]
	if token == "" {
		return rosterRecipe{}, nil, fmt.Errorf("realm roster: unknown character class %q", class)
	}

	cofPath := fmt.Sprintf("data/global/chars/%s/COF/%sTNHTH.cof", token, token)

	asset, err := assetdecode.COF(cache.source, cofPath)
	if err != nil {
		return rosterRecipe{}, nil, err
	}

	direction, err := semanticCOFDirection(0, asset.NumberOfDirections)
	if err != nil {
		return rosterRecipe{}, nil, err
	}

	components := make(map[string]string)

	for _, layer := range asset.CofLayers {
		component := strings.ToUpper(layer.Type.String())

		candidate := defaultRosterComponentPath(token, component, layer)
		if _, err := fs.Stat(cache.source, candidate); err == nil {
			components[component] = candidate
		}
	}

	return rosterRecipe{
		cof:        cofPath,
		palette:    unitsPalette,
		direction:  direction,
		components: components,
	}, asset, nil
}

// defaultRosterComponentPath applies Diablo's component filename convention in one place. Keeping the convention
// explicit avoids subtle cache/render mismatches when new character classes or weapon classes are added.
func defaultRosterComponentPath(token, component string, layer cof.CofLayer) string {
	return fmt.Sprintf(
		"data/global/chars/%s/%s/%s%sLITTN%s.dcc",
		token,
		component,
		token,
		component,
		strings.ToUpper(layer.WeaponClass.String()),
	)
}

// validateRosterRecipe confines all saved appearance inputs to normalized archive-relative data paths. This is a trust
// boundary: a character-controlled recipe must not escape the content filesystem or address internal manifests.
func validateRosterRecipe(recipe rosterRecipe) error {
	paths := []string{recipe.cof, recipe.palette}
	for _, component := range recipe.components {
		paths = append(paths, component)
	}

	for _, name := range paths {
		clean := path.Clean(strings.TrimSpace(name))
		if !fs.ValidPath(clean) || !strings.HasPrefix(strings.ToLower(clean), "data/") {
			return errors.New("realm roster: invalid appearance asset path")
		}
	}

	return nil
}

// semanticCOFDirection translates user-facing compass order into the non-linear order stored by common COF assets.
// Unknown direction counts remain direct-indexed for compatibility with assets that already use semantic ordering.
func semanticCOFDirection(direction, count int) (int, error) {
	lookups := map[int][]int{
		8:  {1, 3, 5, 7, 0, 2, 4, 6},
		16: {2, 6, 10, 14, 0, 4, 8, 12, 1, 3, 5, 7, 9, 11, 13, 15},
	}

	lookup := lookups[count]
	if len(lookup) == 0 {
		if direction < 0 || direction >= count {
			return 0, errors.New("realm roster: semantic direction is out of range")
		}

		return direction, nil
	}

	if direction < 0 || direction >= len(lookup) {
		return 0, errors.New("realm roster: semantic direction is out of range")
	}

	return lookup[direction], nil
}

// rosterID hashes all visual inputs in deterministic order. Character identity and gameplay stats are intentionally
// excluded, allowing unrelated characters with identical resolved appearances to share immutable cache artifacts.
func (cache *Cache) rosterID(recipe rosterRecipe) (string, error) {
	hash := sha256.New()
	_, _ = io.WriteString(hash, rosterRendererVersion)
	paths := []string{recipe.cof, recipe.palette, "data/global/AnimData.d2"}

	components := make([]string, 0, len(recipe.components))
	for component, name := range recipe.components {
		components = append(components, component+"="+name)
		paths = append(paths, name)
	}
	// Map iteration order is unstable, so both the logical recipe and byte-source paths must be sorted before hashing.
	sort.Strings(components)
	sort.Strings(paths)

	_, _ = fmt.Fprintf(hash, "\x00%d\x00%s", recipe.direction, strings.Join(components, "\x00"))

	for _, name := range paths {
		data, err := fs.ReadFile(cache.source, name)
		if err != nil {
			return "", err
		}

		_, _ = io.WriteString(hash, "\x00"+name+"\x00")
		_, _ = hash.Write(data)
	}

	return hex.EncodeToString(hash.Sum(nil))[:40], nil
}

// composeRoster decodes component layers, composes each animation frame, and packs those frames into one browser
// sprite sheet. Its helpers preserve the authored COF layering and a shared coordinate system across every frame.
func (cache *Cache) composeRoster(
	id string,
	recipe rosterRecipe,
	asset *cof.COF,
) (RosterDescriptor, image.Image, error) {
	layers, err := cache.decodeRosterLayers(recipe, asset)
	if err != nil {
		return RosterDescriptor{}, nil, err
	}

	animationBounds, err := rosterAnimationBounds(layers.directions)
	if err != nil {
		return RosterDescriptor{}, nil, err
	}

	frames, err := composeRosterFrames(recipe.direction, asset, layers, animationBounds)
	if err != nil {
		return RosterDescriptor{}, nil, err
	}

	sheet, metadataFrames := packRosterFrames(frames)

	descriptor := RosterDescriptor{
		ID:              id,
		Image:           "/account/roster/" + id + ".png",
		Metadata:        "/account/roster/" + id + ".json",
		FrameDurationMS: cache.rosterFrameDuration(recipe.cof),
		Frames:          metadataFrames,
	}

	return descriptor, sheet, nil
}

// decodeRosterLayers loads only components named by the resolved recipe and selects their corresponding DCC direction.
// Palette slices are copied because the composed frames must not retain mutable storage owned by decoder internals.
func (cache *Cache) decodeRosterLayers(recipe rosterRecipe, asset *cof.COF) (rosterLayers, error) {
	dccDirection, err := assetdecode.DCCDirectionForCOF(recipe.direction, asset.NumberOfDirections)
	if err != nil {
		return rosterLayers{}, err
	}

	layers := rosterLayers{
		cofLayers:  make(map[cof.CompositeType]cof.CofLayer, len(asset.CofLayers)),
		directions: make(map[cof.CompositeType]*dcc.Direction),
		palettes:   make(map[cof.CompositeType]color.Palette),
	}
	for _, layer := range asset.CofLayers {
		name := recipe.components[layer.Type.String()]
		if name == "" {
			continue
		}

		component, err := assetdecode.DCC(cache.source, name, recipe.palette)
		if err != nil {
			return rosterLayers{}, fmt.Errorf("realm roster layer %s: %w", layer.Type, err)
		}

		directions := component.Directions()
		if dccDirection < 0 || dccDirection >= len(directions) {
			return rosterLayers{}, fmt.Errorf(
				"realm roster layer %s lacks direction %d",
				layer.Type,
				dccDirection,
			)
		}

		layers.cofLayers[layer.Type] = layer

		layers.directions[layer.Type] = directions[dccDirection]
		if palette := component.Palette(); palette != nil {
			layers.palettes[layer.Type] = append(color.Palette(nil), (*palette)...)
		}
	}

	return layers, nil
}

// rosterAnimationBounds unions every decoded frame so each composed frame shares one stable origin and size. Without
// this common canvas, frame-to-frame bounds changes would make the browser animation visibly jump.
func rosterAnimationBounds(directions map[cof.CompositeType]*dcc.Direction) (image.Rectangle, error) {
	var animationBounds image.Rectangle

	for _, direction := range directions {
		for _, frame := range direction.Frames() {
			if animationBounds.Empty() {
				animationBounds = frame.Bounds()
			} else {
				animationBounds = animationBounds.Union(frame.Bounds())
			}
		}
	}

	if animationBounds.Empty() {
		return image.Rectangle{}, errors.New("realm roster: no component animation bounds")
	}

	return animationBounds, nil
}

// composeRosterFrames assembles every COF frame from the decoded component directions. A missing component frame is a
// hard error because silently dropping that layer would produce an animation whose appearance changes unexpectedly.
func composeRosterFrames(
	direction int,
	asset *cof.COF,
	layers rosterLayers,
	animationBounds image.Rectangle,
) ([]image.Image, error) {
	frames := make([]image.Image, asset.FramesPerDirection)
	for frameIndex := range frames {
		components := make(map[cof.CompositeType]assetdecode.CompositeFrame, len(layers.directions))
		for componentType, componentDirection := range layers.directions {
			directionFrames := componentDirection.Frames()
			if frameIndex >= len(directionFrames) {
				return nil, fmt.Errorf("realm roster layer %s lacks frame %d", componentType, frameIndex)
			}

			frame := directionFrames[frameIndex]
			components[componentType] = assetdecode.CompositeFrame{
				Indices: frame.PixelData,
				Palette: layers.palettes[componentType],
				Bounds:  frame.Bounds(),
				Layer:   layers.cofLayers[componentType],
			}
		}

		composed, _, err := assetdecode.ComposeCOFFrame(
			asset,
			direction,
			frameIndex,
			components,
			animationBounds,
		)
		if err != nil {
			return nil, err
		}

		frames[frameIndex] = composed
	}

	return frames, nil
}

// packRosterFrames places composed frames left-to-right in rows no wider than 1024 pixels. Metadata records the exact
// placement order so the browser can animate the sheet without reproducing the packing algorithm.
func packRosterFrames(frames []image.Image) (image.Image, []rosterFrame) {
	frameWidth, frameHeight := frames[0].Bounds().Dx(), frames[0].Bounds().Dy()
	columns := max(1, min(len(frames), 1024/max(1, frameWidth)))
	rows := (len(frames) + columns - 1) / columns
	sheet := image.NewRGBA(image.Rect(0, 0, columns*frameWidth, rows*frameHeight))

	metadataFrames := make([]rosterFrame, len(frames))
	for index, frame := range frames {
		x, y := index%columns*frameWidth, index/columns*frameHeight
		draw.Draw(sheet, image.Rect(x, y, x+frameWidth, y+frameHeight), frame, frame.Bounds().Min, draw.Over)
		metadataFrames[index] = rosterFrame{X: x, Y: y, Width: frameWidth, Height: frameHeight}
	}

	return sheet, metadataFrames
}

// rosterFrameDuration uses the authored animation speed when available and falls back to Diablo's conventional
// 25-frames-per-second timing when animation data is missing or invalid, keeping browser playback usable.
func (cache *Cache) rosterFrameDuration(cofPath string) float64 {
	animation, err := assetdecode.AnimationData(cache.source, "data/global/AnimData.d2")
	if err == nil {
		name := strings.ToUpper(strings.TrimSuffix(path.Base(cofPath), path.Ext(cofPath)))
		if record := animation.GetRecord(name); record != nil && record.Speed() > 0 {
			return record.FrameDurationMS()
		}
	}

	return 256000.0 / (128.0 * 25.0)
}

// serveRoster exposes only 40-character content hashes with known extensions from the private cache directory. Strict
// validation prevents path traversal while immutable caching is safe because each URL names its exact visual inputs.
func (cache *Cache) serveRoster(writer http.ResponseWriter, request *http.Request) {
	name := strings.TrimPrefix(request.URL.Path, "/account/roster/")
	extension := filepath.Ext(name)

	id := strings.TrimSuffix(name, extension)
	if len(id) != 40 {
		http.NotFound(writer, request)
		return
	}

	if _, err := hex.DecodeString(id); err != nil {
		http.NotFound(writer, request)
		return
	}

	fileName := filepath.Join(cache.directory, "roster-"+id+extension)
	if extension != ".png" && extension != ".json" {
		http.NotFound(writer, request)
		return
	}

	if _, err := os.Stat(fileName); err != nil {
		http.NotFound(writer, request)
		return
	}

	writer.Header().Set("Cache-Control", "private, max-age=31536000, immutable")

	if extension == ".png" {
		writer.Header().Set("Content-Type", "image/png")
	} else {
		writer.Header().Set("Content-Type", "application/json")
	}

	http.ServeFile(writer, request, fileName)
}

// readRosterDescriptor decodes a previously published descriptor. Callers still verify its image companion before
// treating it as a cache hit because the two files cannot be renamed atomically as a pair.
func readRosterDescriptor(name string) (RosterDescriptor, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return RosterDescriptor{}, err
	}

	var descriptor RosterDescriptor
	if err := json.Unmarshal(data, &descriptor); err != nil {
		return RosterDescriptor{}, err
	}

	return descriptor, nil
}

// cloneStrings normalizes saved component keys and paths into the form expected by COF layer lookup. Copying also keeps
// later caller mutations from changing a recipe while it is being hashed or rendered.
func cloneStrings(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[strings.ToUpper(key)] = strings.TrimSpace(value)
	}

	return result
}
