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
	if descriptor, err := readRosterDescriptor(prefix + ".json"); err == nil {
		if _, imageErr := os.Stat(prefix + ".png"); imageErr == nil {
			return descriptor, nil
		}
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if descriptor, err := readRosterDescriptor(prefix + ".json"); err == nil {
		if _, imageErr := os.Stat(prefix + ".png"); imageErr == nil {
			return descriptor, nil
		}
	}
	descriptor, sheet, err := cache.composeRoster(id, recipe, asset)
	if err != nil {
		return RosterDescriptor{}, err
	}
	if err := writePNG(prefix+".png", sheet); err != nil {
		return RosterDescriptor{}, err
	}
	metadata, err := json.Marshal(descriptor)
	if err != nil {
		return RosterDescriptor{}, err
	}
	if err := writePrivateFile(prefix+".json", func(writer io.Writer) error {
		_, err := writer.Write(metadata)
		return err
	}); err != nil {
		return RosterDescriptor{}, err
	}
	return descriptor, nil
}

func (cache *Cache) resolveRosterRecipe(character d2save.Character) (rosterRecipe, *cof.COF, error) {
	if character.Appearance != nil && strings.TrimSpace(character.Appearance.COF) != "" {
		recipe := rosterRecipe{
			cof:        strings.TrimSpace(character.Appearance.COF),
			palette:    strings.TrimSpace(character.Appearance.Palette),
			direction:  character.Appearance.Direction,
			components: cloneStrings(character.Appearance.Components),
		}
		if recipe.palette == "" {
			recipe.palette = "data/global/Palette/units/pal.dat"
		}
		if err := validateRosterRecipe(recipe); err != nil {
			return rosterRecipe{}, nil, err
		}
		asset, err := assetdecode.COF(cache.source, recipe.cof)
		return recipe, asset, err
	}

	token := rosterClassTokens[character.Class]
	if token == "" {
		return rosterRecipe{}, nil, fmt.Errorf("realm roster: unknown character class %q", character.Class)
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
		candidate := fmt.Sprintf(
			"data/global/chars/%s/%s/%s%sLITTN%s.dcc",
			token,
			component,
			token,
			component,
			strings.ToUpper(layer.WeaponClass.String()),
		)
		if _, err := fs.Stat(cache.source, candidate); err == nil {
			components[component] = candidate
		}
	}
	return rosterRecipe{
		cof: cofPath, palette: "data/global/Palette/units/pal.dat",
		direction: direction, components: components,
	}, asset, nil
}

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

func (cache *Cache) rosterID(recipe rosterRecipe) (string, error) {
	hash := sha256.New()
	_, _ = io.WriteString(hash, rosterRendererVersion)
	paths := []string{recipe.cof, recipe.palette, "data/global/AnimData.d2"}
	components := make([]string, 0, len(recipe.components))
	for component, name := range recipe.components {
		components = append(components, component+"="+name)
		paths = append(paths, recipe.components[component])
	}
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

func (cache *Cache) composeRoster(id string, recipe rosterRecipe, asset *cof.COF) (RosterDescriptor, image.Image, error) {
	dccDirection, err := assetdecode.DCCDirectionForCOF(recipe.direction, asset.NumberOfDirections)
	if err != nil {
		return RosterDescriptor{}, nil, err
	}
	layers := make(map[cof.CompositeType]cof.CofLayer, len(asset.CofLayers))
	decoded := make(map[cof.CompositeType]*dcc.Direction)
	palettes := make(map[cof.CompositeType]color.Palette)
	for _, layer := range asset.CofLayers {
		name := recipe.components[layer.Type.String()]
		if name == "" {
			continue
		}
		component, err := assetdecode.DCC(cache.source, name, recipe.palette)
		if err != nil {
			return RosterDescriptor{}, nil, fmt.Errorf("realm roster layer %s: %w", layer.Type, err)
		}
		directions := component.Directions()
		if dccDirection < 0 || dccDirection >= len(directions) {
			return RosterDescriptor{}, nil, fmt.Errorf("realm roster layer %s lacks direction %d", layer.Type, dccDirection)
		}
		layers[layer.Type] = layer
		decoded[layer.Type] = directions[dccDirection]
		if palette := component.Palette(); palette != nil {
			palettes[layer.Type] = append(color.Palette(nil), (*palette)...)
		}
	}

	var animationBounds image.Rectangle
	for _, direction := range decoded {
		for _, frame := range direction.Frames() {
			if animationBounds.Empty() {
				animationBounds = frame.Bounds()
			} else {
				animationBounds = animationBounds.Union(frame.Bounds())
			}
		}
	}
	if animationBounds.Empty() {
		return RosterDescriptor{}, nil, errors.New("realm roster: no component animation bounds")
	}

	frames := make([]image.Image, asset.FramesPerDirection)
	for frameIndex := range frames {
		components := make(map[cof.CompositeType]assetdecode.CompositeFrame, len(decoded))
		for componentType, direction := range decoded {
			directionFrames := direction.Frames()
			if frameIndex >= len(directionFrames) {
				return RosterDescriptor{}, nil, fmt.Errorf("realm roster layer %s lacks frame %d", componentType, frameIndex)
			}
			frame := directionFrames[frameIndex]
			components[componentType] = assetdecode.CompositeFrame{
				Indices: frame.PixelData,
				Palette: palettes[componentType],
				Bounds:  frame.Bounds(),
				Layer:   layers[componentType],
			}
		}
		frames[frameIndex], _, err = assetdecode.ComposeCOFFrame(
			asset,
			recipe.direction,
			frameIndex,
			components,
			animationBounds,
		)
		if err != nil {
			return RosterDescriptor{}, nil, err
		}
	}

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
	descriptor := RosterDescriptor{
		ID:              id,
		Image:           "/account/roster/" + id + ".png",
		Metadata:        "/account/roster/" + id + ".json",
		FrameDurationMS: cache.rosterFrameDuration(recipe.cof),
		Frames:          metadataFrames,
	}
	return descriptor, sheet, nil
}

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

func cloneStrings(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[strings.ToUpper(key)] = strings.TrimSpace(value)
	}
	return result
}
