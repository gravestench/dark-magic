package modruntime

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"time"

	cof "github.com/gravestench/cof"
	"github.com/gravestench/dark-magic/internal/assets/decode"
	assetinspect "github.com/gravestench/dark-magic/internal/assets/inspect"
	cachepkg "github.com/gravestench/dark-magic/internal/cache"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/presentation/maprender"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	dc6 "github.com/gravestench/dc6/pkg"
	dcc "github.com/gravestench/dcc/pkg"
	dt1 "github.com/gravestench/dt1"
	lua "github.com/yuin/gopher-lua"
)

const renderNodeType = "engine.render.node/v1"

type ownedRenderNode struct {
	composer *render.Composer
	id       render.NodeID
	resource render.ResourceID
	// resourceRelease is non-nil when resource is borrowed from the capability's
	// shared immutable texture pool rather than owned exclusively by this node.
	resourceRelease func() error
	palette         render.ResourceID
	owned           []render.ResourceID
	assets          fs.FS
	cache           *renderAssetCache
	pool            *renderResourcePool
	once            sync.Once
	err             error
}

type renderAssetCache struct {
	mu          sync.Mutex
	generation  uint64
	encoded     *cachepkg.Cache
	decoded     *cachepkg.Cache
	composed    *cachepkg.Cache
	world       *cachepkg.Cache
	inflight    map[string]*assetDecodeFlight
	decodeCalls uint64
	decodeTime  time.Duration
	stages      map[string]DecodeStageDiagnostics
}

// assetDecodeFlight lets every caller interested in the same cold asset share
// one decode. The expensive work happens without holding the cache mutex, so
// unrelated assets can be prepared by different preload workers at once.
type assetDecodeFlight struct {
	done  chan struct{}
	value any
	err   error
}

// RenderDiagnostics is a stable profiling snapshot of decoded and retained
// renderer state. Cache weight estimates retained decoded bytes, while retained
// texture bytes estimates expanded RGBA residency.
type RenderDiagnostics struct {
	Decoded         cachepkg.Stats
	Encoded         cachepkg.Stats
	Composed        cachepkg.Stats
	World           cachepkg.Stats
	Retained        render.Diagnostics
	DecodeCalls     uint64
	DecodeTime      time.Duration
	PreloadsPending int
	Stages          map[string]DecodeStageDiagnostics
}

// DecodeStageDiagnostics separates file parsing, direction decoding, and
// composition. Without this split, a single cumulative number cannot tell us
// whether a codec, the compositor, or native upload deserves attention.
type DecodeStageDiagnostics struct {
	Calls uint64
	Time  time.Duration
}

// RenderCapability owns the shared asset cache behind engine.render/v1.
type RenderCapability struct {
	runtime  *Runtime
	composer *render.Composer
	assets   fs.FS
	cache    *renderAssetCache
	preloads *assetPreloader
	pool     *renderResourcePool
}

// NewRenderCapability creates the shared decode/preload cache used by every Lua
// render module instance. The composer still owns semantic retained resources,
// and the platform backend alone owns native uploads.
func NewRenderCapability(runtime *Runtime, composer *render.Composer, assets fs.FS) *RenderCapability {
	// Character-creation animations expand into hundreds of RGBA frames. A
	// 64 MiB cache discarded the first prepared states before preload finished,
	// forcing the Lua interaction path to decode them again on first hover.
	const (
		encodedAssetBudget  = 32 * 1024 * 1024
		decodedAssetBudget  = 128 * 1024 * 1024
		composedAssetBudget = 352 * 1024 * 1024
		worldChunkBudget    = 96 * 1024 * 1024
	)
	cache := &renderAssetCache{
		encoded:  cachepkg.New(encodedAssetBudget),
		decoded:  cachepkg.New(decodedAssetBudget),
		composed: cachepkg.New(composedAssetBudget),
		world:    cachepkg.New(worldChunkBudget),
		inflight: make(map[string]*assetDecodeFlight),
		stages:   make(map[string]DecodeStageDiagnostics),
	}
	return &RenderCapability{
		runtime: runtime, composer: composer, assets: assets, cache: cache,
		preloads: newAssetPreloader(assets, cache, composer),
		pool:     newRenderResourcePool(composer),
	}
}

func (c *renderAssetCache) loadImage(assets fs.FS, name string) (image.Image, error) {
	value, err := c.load(assets, "image\x00"+name, func() (any, int, error) {
		file, err := assets.Open(name)
		if err != nil {
			return nil, 0, err
		}
		defer file.Close()
		decoded, _, err := image.Decode(file)
		if err != nil {
			return nil, 0, err
		}
		return decoded, decoded.Bounds().Dx() * decoded.Bounds().Dy() * 4, nil
	})
	if err != nil {
		return nil, err
	}
	return value.(image.Image), nil
}

// Diagnostics returns copied cache/composer counters without exposing mutable
// resource tables to scripts or developer tooling.
func (r *RenderCapability) Diagnostics() RenderDiagnostics {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()
	stages := make(map[string]DecodeStageDiagnostics, len(r.cache.stages))
	for name, stage := range r.cache.stages {
		stages[name] = stage
	}
	encoded, decoded, composed, world := r.cache.encoded.Diagnostics(), r.cache.decoded.Diagnostics(), r.cache.composed.Diagnostics(), r.cache.world.Diagnostics()
	return RenderDiagnostics{Decoded: combinedCacheStats(encoded, decoded, composed, world), Encoded: encoded, Composed: composed, World: world, Retained: r.composer.Diagnostics(), DecodeCalls: r.cache.decodeCalls, DecodeTime: r.cache.decodeTime, PreloadsPending: r.preloads.Pending(), Stages: stages}
}

func combinedCacheStats(stats ...cachepkg.Stats) cachepkg.Stats {
	var result cachepkg.Stats
	for _, current := range stats {
		result.Entries += current.Entries
		result.Weight += current.Weight
		result.Budget += current.Budget
		result.Hits += current.Hits
		result.Misses += current.Misses
		result.Evictions += current.Evictions
	}
	return result
}

type generationSource interface{ Generation() uint64 }

type compositeFrame struct {
	image   image.Image
	indices []byte
	palette color.Palette
	bounds  image.Rectangle
	layer   cof.CofLayer
}

type rgbaFrameDigest struct {
	width, height int
	pixels        [32]byte
}

type preparedDC6Frame struct {
	image image.Image
	frame *dc6.Frame
}

type preparedDC6File struct {
	file    *dc6.File
	palette color.Palette
}

type preparedDC6Direction struct {
	asset *dc6.DC6
}

type preparedDT1File struct {
	file *dt1.File
}

type preparedDT1Tile struct {
	image image.Image
	tile  *dt1.Tile
	total int
}

type preparedDC6Animation struct {
	frames []image.Image
	bounds image.Rectangle
}

type preparedCOFAnimation struct {
	frames []image.Image
	keys   []string
	asset  *cof.COF
	origin image.Point
}

type preparedDCCFile struct {
	file    *dcc.File
	palette color.Palette
}

type preparedDCCDirection struct {
	direction *dcc.Direction
	palette   color.Palette
}

func imageWeight(value image.Image) int {
	if value == nil {
		return 1
	}
	return max(value.Bounds().Dx()*value.Bounds().Dy()*4, 1)
}

func imagesWeight(values []image.Image) int {
	weight := 0
	for _, value := range values {
		weight += imageWeight(value)
	}
	return max(weight, 1)
}

func blendRGBA(destination *image.RGBA, x, y int, source color.RGBA, opacity uint8) {
	if !image.Pt(x, y).In(destination.Bounds()) || source.A == 0 || opacity == 0 {
		return
	}
	alpha := uint32(source.A) * uint32(opacity) / 255
	offset := destination.PixOffset(x, y)
	if alpha == 255 {
		destination.Pix[offset], destination.Pix[offset+1] = source.R, source.G
		destination.Pix[offset+2], destination.Pix[offset+3] = source.B, 255
		return
	}
	inverse := 255 - alpha
	destination.Pix[offset] = uint8((uint32(source.R)*alpha + uint32(destination.Pix[offset])*inverse) / 255)
	destination.Pix[offset+1] = uint8((uint32(source.G)*alpha + uint32(destination.Pix[offset+1])*inverse) / 255)
	destination.Pix[offset+2] = uint8((uint32(source.B)*alpha + uint32(destination.Pix[offset+2])*inverse) / 255)
	destination.Pix[offset+3] = uint8(alpha + uint32(destination.Pix[offset+3])*inverse/255)
}

// drawCompositeComponent consumes DCC palette indexes directly. This avoids a
// temporary RGBA allocation and palette expansion for every layer/frame before
// immediately copying those pixels into the final composite.
func drawCompositeComponent(output *image.RGBA, component compositeFrame, destination image.Point, opacity uint8) {
	if len(component.indices) > 0 && len(component.palette) > 0 {
		width, height := component.bounds.Dx(), component.bounds.Dy()
		for y := 0; y < height; y++ {
			row := y * width
			for x := 0; x < width && row+x < len(component.indices); x++ {
				index := int(component.indices[row+x])
				if index == 0 || index >= len(component.palette) {
					continue
				}
				value := color.RGBAModel.Convert(component.palette[index]).(color.RGBA)
				blendRGBA(output, destination.X+x, destination.Y+y, value, opacity)
			}
		}
		return
	}
	if component.image == nil {
		return
	}
	if opacity == 255 {
		draw.Draw(output, component.image.Bounds().Add(destination), component.image, component.image.Bounds().Min, draw.Over)
		return
	}
	mask := image.NewUniform(color.Alpha{A: opacity})
	draw.DrawMask(output, component.image.Bounds().Add(destination), component.image, component.image.Bounds().Min, mask, image.Point{}, draw.Over)
}

// projectedShadowBounds mirrors Riiablo's legacy character-shadow transform:
// keep the feet on their shared baseline, compress vertical distance by one
// half, and shear upper pixels left by that same distance.
func projectedShadowBounds(bounds image.Rectangle) image.Rectangle {
	if bounds.Empty() {
		return image.Rectangle{}
	}
	shift := bounds.Dy() / 2
	baseline := bounds.Max.Y - 1
	return image.Rect(bounds.Min.X-shift, baseline-shift, bounds.Max.X, baseline+1)
}

func shadowCanvasBounds(visible image.Rectangle, components map[cof.CompositeType]compositeFrame) image.Rectangle {
	projected := visible
	for _, component := range components {
		if component.layer.Shadow != 0 {
			// One eligible component means the assembled silhouette is projected
			// from the shared character bounds. Do not size the canvas from each
			// component's private baseline or the far-left shadow gets clipped.
			projected = projected.Union(projectedShadowBounds(visible))
			break
		}
	}
	// Retained animation nodes are center-anchored. Expand both sides equally so
	// adding a long shadow never slides the visible body away from the hero point.
	horizontal := max(visible.Min.X-projected.Min.X, projected.Max.X-visible.Max.X, 0)
	vertical := max(visible.Min.Y-projected.Min.Y, projected.Max.Y-visible.Max.Y, 0)
	return image.Rect(visible.Min.X-horizontal, visible.Min.Y-vertical, visible.Max.X+horizontal, visible.Max.Y+vertical)
}

// compositeShadowMask assembles every shadow-enabled body part at the shared
// character origin before projection. Projecting parts independently gives a
// head, arms, and torso different ground baselines—the floating fragments the
// player sees beside the character instead of one contiguous silhouette.
func compositeShadowMask(bounds image.Rectangle, priority []cof.CompositeType, components map[cof.CompositeType]compositeFrame) *image.RGBA {
	mask := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for _, componentType := range priority {
		component, ok := components[componentType]
		if !ok || component.layer.Shadow == 0 {
			continue
		}
		width, height := component.bounds.Dx(), component.bounds.Dy()
		origin := component.bounds.Min.Sub(bounds.Min)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var alpha uint8
				if len(component.indices) > 0 && y*width+x < len(component.indices) {
					index := int(component.indices[y*width+x])
					if index > 0 && index < len(component.palette) {
						alpha = color.RGBAModel.Convert(component.palette[index]).(color.RGBA).A
					}
				} else if component.image != nil {
					point := component.image.Bounds().Min.Add(image.Pt(x, y))
					_, _, _, value := component.image.At(point.X, point.Y).RGBA()
					alpha = uint8(value >> 8)
				}
				if alpha > mask.RGBAAt(origin.X+x, origin.Y+y).A {
					mask.SetRGBA(origin.X+x, origin.Y+y, color.RGBA{A: alpha})
				}
			}
		}
	}
	return mask
}

func drawCompositeShadow(output *image.RGBA, mask *image.RGBA, bounds, canvas image.Rectangle, opacity uint8) {
	width, height := mask.Bounds().Dx(), mask.Bounds().Dy()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := mask.RGBAAt(x, y).A
			if alpha == 0 {
				continue
			}
			// Integer rounding deliberately matches nearest-pixel rasterization of
			// the half-scale/shear transform instead of accumulating float drift.
			distance := height - 1 - y
			shift := (distance + 1) / 2
			absoluteX := bounds.Min.X + x - shift
			absoluteY := bounds.Max.Y - 1 - shift
			blendRGBA(output, absoluteX-canvas.Min.X, absoluteY-canvas.Min.Y, color.RGBA{A: alpha}, opacity)
		}
	}
}

// rgbaDCCFrame expands palette indexes with tight slice writes. Using
// image/draw against dcc.Frame would call the image.Image interface once per
// pixel (Bounds, At, ColorIndexAt, palette conversion). Character composites
// perform this operation for every layer and frame, so that generic path makes
// cold preparation needlessly expensive.
func rgbaDCCFrame(frame *dcc.Frame, palette *color.Palette) *image.RGBA {
	bounds := frame.Bounds()
	result := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	if palette == nil {
		return result
	}
	var colors [256]color.RGBA
	for index, value := range *palette {
		if index >= len(colors) {
			break
		}
		colors[index] = color.RGBAModel.Convert(value).(color.RGBA)
	}
	for pixel, paletteIndex := range frame.PixelData {
		if pixel*4+3 >= len(result.Pix) {
			break
		}
		value := colors[paletteIndex]
		offset := pixel * 4
		result.Pix[offset], result.Pix[offset+1] = value.R, value.G
		result.Pix[offset+2], result.Pix[offset+3] = value.B, value.A
	}
	return result
}

// dccDirectionForCOF translates a spatial COF priority direction into the
// separately interleaved direction order stored by DCC. OpenDiablo2 has two
// distinct Dir64ToCof and Dir64ToDcc tables; using one index for both pairs a
// correctly facing sprite with another facing's arm/head priority.
func dccDirectionForCOF(direction, count int) (int, error) {
	lookups := map[int][]int{
		8:  {4, 0, 5, 1, 6, 2, 7, 3},
		16: {4, 8, 0, 9, 5, 10, 1, 11, 6, 12, 2, 13, 7, 14, 3, 15},
	}
	lookup, ok := lookups[count]
	if !ok {
		if direction < 0 || direction >= count {
			return 0, fmt.Errorf("COF direction %d is out of range for %d directions", direction, count)
		}
		return direction, nil
	}
	if direction < 0 || direction >= len(lookup) {
		return 0, fmt.Errorf("COF direction %d is out of range for %d directions", direction, count)
	}
	return lookup[direction], nil
}

func composeCOFFrame(asset *cof.COF, direction, frame int, components map[cof.CompositeType]compositeFrame, animationBounds ...image.Rectangle) (image.Image, error) {
	if direction < 0 || direction >= len(asset.Priority) {
		return nil, fmt.Errorf("COF direction %d is out of range", direction)
	}
	if frame < 0 || frame >= len(asset.Priority[direction]) {
		return nil, fmt.Errorf("COF frame %d is out of range", frame)
	}
	var bounds image.Rectangle
	if len(animationBounds) > 0 {
		bounds = animationBounds[0]
	} else {
		for _, component := range components {
			if bounds.Empty() {
				bounds = component.bounds
			} else {
				bounds = bounds.Union(component.bounds)
			}
		}
	}
	if bounds.Empty() {
		return nil, errors.New("COF composition has no component frames")
	}
	canvas := shadowCanvasBounds(bounds, components)
	output := image.NewRGBA(image.Rect(0, 0, canvas.Dx(), canvas.Dy()))
	// Diablo's composite renderer makes two complete passes over the COF order:
	// every projected shadow first, then every visible body/equipment layer.
	// Drawing a layer's shadow immediately before that layer lets later shadows
	// darken earlier limbs and looks like broken arm priority on thin characters.
	priority := asset.Priority[direction][frame]
	shadow := compositeShadowMask(bounds, priority, components)
	drawCompositeShadow(output, shadow, bounds, canvas, 96)
	for _, componentType := range priority {
		component, ok := components[componentType]
		if !ok {
			continue
		}
		destination := component.bounds.Min.Sub(canvas.Min)
		alpha := uint8(255)
		// DrawEffect is meaningful only when COF marks the layer transparent.
		// Most body layers contain zero in this byte, which otherwise looks like
		// the 25-percent effect despite being explicitly opaque.
		if component.layer.Transparent {
			switch component.layer.DrawEffect {
			case cof.DrawEffect(0):
				alpha = 191
			case cof.DrawEffect(1):
				alpha = 128
			case cof.DrawEffect(2):
				alpha = 64
			default:
				alpha = 128
			}
		}
		drawCompositeComponent(output, component, destination, alpha)
	}
	return output, nil
}

func (c *renderAssetCache) refresh(assets fs.FS) {
	source, ok := assets.(generationSource)
	if !ok || source.Generation() == c.generation {
		return
	}
	c.generation = source.Generation()
	c.encoded.InvalidateNamespace("encoded", c.generation)
	c.decoded.InvalidateNamespace("decoded", c.generation)
	c.composed.InvalidateNamespace("composed", c.generation)
	c.world.InvalidateNamespace("world", c.generation)
}

func (c *renderAssetCache) tier(key string) (*cachepkg.Cache, string) {
	switch {
	case strings.HasPrefix(key, "world-chunk\x00"):
		return c.world, "world"
	case strings.HasPrefix(key, "dcc-file\x00"), strings.HasPrefix(key, "dc6-file\x00"), strings.HasPrefix(key, "dt1-file\x00"), strings.HasPrefix(key, "cof\x00"), strings.HasPrefix(key, "animdata\x00"):
		return c.encoded, "encoded"
	case strings.HasPrefix(key, "cof-animation\x00"), strings.HasPrefix(key, "dc6-animation\x00"), strings.HasPrefix(key, "dc6-combined\x00"), strings.HasPrefix(key, "ds1\x00"), strings.HasPrefix(key, "ds1-chunks\x00"):
		return c.composed, "composed"
	default:
		return c.decoded, "decoded"
	}
}

func assetWeight(assets fs.FS, names ...string) int {
	weight := 0
	for _, name := range names {
		file, err := assets.Open(name)
		if err != nil {
			continue
		}
		count, readErr := io.Copy(io.Discard, file)
		_ = file.Close()
		if readErr == nil {
			weight += int(count)
		}
	}
	if weight < 1 {
		return 1
	}
	return weight
}

func dccDirectionWeight(direction *dcc.Direction) int {
	if direction == nil {
		return 1
	}
	weight := len(direction.PixelData)
	for _, frame := range direction.Frames() {
		weight += len(frame.PixelData)
	}
	return max(weight, 1)
}

func (c *renderAssetCache) load(assets fs.FS, key string, decode func() (any, int, error)) (any, error) {
	c.mu.Lock()
	c.refresh(assets)
	generation := c.generation
	tier, namespace := c.tier(key)
	if cached, ok := tier.RetrieveVersioned(namespace, key, generation); ok {
		c.mu.Unlock()
		return cached, nil
	}
	flightKey := fmt.Sprintf("%d\x00%s", generation, key)
	if flight, ok := c.inflight[flightKey]; ok {
		c.mu.Unlock()
		<-flight.done
		return flight.value, flight.err
	}
	flight := &assetDecodeFlight{done: make(chan struct{})}
	c.inflight[flightKey] = flight
	c.mu.Unlock()

	started := time.Now()
	value, weight, err := decode()
	elapsed := time.Since(started)

	c.mu.Lock()
	c.decodeCalls++
	c.decodeTime += elapsed
	stageName := key
	if separator := strings.IndexByte(stageName, 0); separator >= 0 {
		stageName = stageName[:separator]
	}
	stage := c.stages[stageName]
	stage.Calls++
	stage.Time += elapsed
	c.stages[stageName] = stage
	if err == nil && generation == c.generation {
		err = tier.InsertVersioned(namespace, key, generation, value, weight)
	}
	flight.value, flight.err = value, err
	delete(c.inflight, flightKey)
	close(flight.done)
	c.mu.Unlock()
	return value, err
}

func (c *renderAssetCache) currentGeneration(assets fs.FS) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.refresh(assets)
	return c.generation
}

func (c *renderAssetCache) loadFont(assets fs.FS, table, sheet, palette, transform string) (*assetdecode.BitmapFont, error) {
	key := table + "\x00" + sheet + "\x00" + palette + "\x00" + transform
	value, err := c.load(assets, "font\x00"+key, func() (any, int, error) {
		font, err := assetdecode.LoadBitmapFontWithTransform(assets, table, sheet, palette, transform)
		return font, assetWeight(assets, table, sheet, palette), err
	})
	if err != nil {
		return nil, err
	}
	return value.(*assetdecode.BitmapFont), nil
}

func (c *renderAssetCache) loadCOF(assets fs.FS, name string) (*cof.COF, error) {
	value, err := c.load(assets, "cof\x00"+name, func() (any, int, error) {
		asset, err := assetdecode.COF(assets, name)
		return asset, assetWeight(assets, name), err
	})
	if err != nil {
		return nil, err
	}
	return value.(*cof.COF), nil
}

func (c *renderAssetCache) loadAnimationData(assets fs.FS, name string) (*cof.AnimationData, error) {
	value, err := c.load(assets, "animdata\x00"+name, func() (any, int, error) {
		asset, err := assetdecode.AnimationData(assets, name)
		return asset, assetWeight(assets, name), err
	})
	if err != nil {
		return nil, err
	}
	return value.(*cof.AnimationData), nil
}

func (c *renderAssetCache) loadDCCFile(assets fs.FS, name, palette string) (preparedDCCFile, error) {
	key := name + "\x00" + palette
	value, err := c.load(assets, "dcc-file\x00"+key, func() (any, int, error) {
		asset, err := assets.Open(name)
		if err != nil {
			return nil, 0, err
		}
		defer asset.Close()
		encoded, err := io.ReadAll(asset)
		if err != nil {
			return nil, 0, err
		}
		opened, err := dcc.OpenBytes(encoded)
		if err != nil {
			return nil, 0, err
		}
		colors := append(color.Palette(nil), (*dcc.DefaultPalette())...)
		if palette != "" {
			colors, err = assetdecode.Palette(assets, palette)
			if err != nil {
				return nil, 0, err
			}
		}
		opened.SetPalette(colors)
		return preparedDCCFile{file: opened, palette: colors}, max(len(encoded)+len(colors)*4, 1), nil
	})
	if err != nil {
		return preparedDCCFile{}, err
	}
	return value.(preparedDCCFile), nil
}

func (c *renderAssetCache) loadDCCDirection(assets fs.FS, name, palette string, direction int) (preparedDCCDirection, error) {
	key := fmt.Sprintf("dcc-direction\x00%s\x00%s\x00%d", name, palette, direction)
	value, err := c.load(assets, key, func() (any, int, error) {
		opened, err := c.loadDCCFile(assets, name, palette)
		if err != nil {
			return nil, 0, err
		}
		decoded, err := opened.file.DecodeDirection(direction)
		if err != nil {
			return nil, 0, err
		}
		return preparedDCCDirection{direction: decoded, palette: opened.palette}, dccDirectionWeight(decoded), nil
	})
	if err != nil {
		return preparedDCCDirection{}, err
	}
	return value.(preparedDCCDirection), nil
}

func (c *renderAssetCache) loadDC6File(assets fs.FS, name, palette string) (preparedDC6File, error) {
	key := name + "\x00" + palette
	value, err := c.load(assets, "dc6-file\x00"+key, func() (any, int, error) {
		asset, err := assets.Open(name)
		if err != nil {
			return nil, 0, err
		}
		defer asset.Close()
		encoded, err := io.ReadAll(asset)
		if err != nil {
			return nil, 0, err
		}
		opened, err := dc6.OpenBytes(encoded)
		if err != nil {
			return nil, 0, err
		}
		defaults := &dc6.DC6{}
		defaults.SetPalette(nil)
		colors := append(color.Palette(nil), defaults.Palette()...)
		if palette != "" {
			colors, err = assetdecode.Palette(assets, palette)
			if err != nil {
				return nil, 0, err
			}
		}
		opened.SetPalette(colors)
		return preparedDC6File{file: opened, palette: colors}, max(len(encoded)+len(colors)*4, 1), nil
	})
	if err != nil {
		return preparedDC6File{}, err
	}
	return value.(preparedDC6File), nil
}

func (c *renderAssetCache) loadDC6Direction(assets fs.FS, name, palette string, direction int) (preparedDC6Direction, error) {
	key := fmt.Sprintf("dc6-direction\x00%s\x00%s\x00%d", name, palette, direction)
	value, err := c.load(assets, key, func() (any, int, error) {
		opened, err := c.loadDC6File(assets, name, palette)
		if err != nil {
			return nil, 0, err
		}
		if direction < 0 || direction >= opened.file.Directions() {
			return nil, 0, fmt.Errorf("dc6: direction %d out of range [0,%d)", direction, opened.file.Directions())
		}
		decoded := &dc6.DC6{Directions: []*dc6.Direction{{Frames: make([]*dc6.Frame, opened.file.FramesPerDirection())}}}
		decoded.SetPalette(opened.palette)
		weight := 1
		for frameIndex := range decoded.Directions[0].Frames {
			frame, err := opened.file.DecodeFrame(direction, frameIndex)
			if err != nil {
				return nil, 0, err
			}
			decoded.Directions[0].Frames[frameIndex] = frame
			weight += len(frame.FrameData) + len(frame.IndexData)
		}
		return preparedDC6Direction{asset: decoded}, weight, nil
	})
	if err != nil {
		return preparedDC6Direction{}, err
	}
	return value.(preparedDC6Direction), nil
}

func (c *renderAssetCache) loadDC6Frame(assets fs.FS, name, palette string, direction, frameIndex int) (preparedDC6Frame, error) {
	key := fmt.Sprintf("dc6-frame\x00%s\x00%s\x00%d\x00%d", name, palette, direction, frameIndex)
	value, err := c.load(assets, key, func() (any, int, error) {
		opened, err := c.loadDC6File(assets, name, palette)
		if err != nil {
			return nil, 0, err
		}
		frame, err := opened.file.DecodeFrame(direction, frameIndex)
		if err != nil {
			return nil, 0, err
		}
		asset := &dc6.DC6{}
		asset.SetPalette(opened.palette)
		decoded, err := assetdecode.FrameImage(asset, frame)
		if err != nil {
			return nil, 0, err
		}
		// The prepared image owns the pixels now. Keep only placement metadata on
		// the frame rather than retaining a second indexed/encoded copy.
		frame.FrameData = nil
		frame.IndexData = nil
		frame.Terminator = nil
		return preparedDC6Frame{image: decoded, frame: frame}, imageWeight(decoded), nil
	})
	if err != nil {
		return preparedDC6Frame{}, err
	}
	return value.(preparedDC6Frame), nil
}

func (c *renderAssetCache) loadDC6Combined(assets fs.FS, name, palette string, direction int) ([]image.Image, error) {
	key := fmt.Sprintf("dc6-combined\x00%s\x00%s\x00%d", name, palette, direction)
	value, err := c.load(assets, key, func() (any, int, error) {
		prepared, err := c.loadDC6Direction(assets, name, palette, direction)
		if err != nil {
			return nil, 0, err
		}
		pages, err := combinedDC6Pages(prepared.asset, 0)
		return pages, imagesWeight(pages), err
	})
	if err != nil {
		return nil, err
	}
	return value.([]image.Image), nil
}

func (c *renderAssetCache) loadDC6Animation(assets fs.FS, name, palette string, direction int, anchorMode string, sharedBounds ...image.Rectangle) (preparedDC6Animation, error) {
	boundsKey := ""
	if len(sharedBounds) > 0 {
		boundsKey = fmt.Sprint(sharedBounds[0])
	}
	key := fmt.Sprintf("dc6-animation\x00%s\x00%s\x00%d\x00%s\x00%s", name, palette, direction, anchorMode, boundsKey)
	value, err := c.load(assets, key, func() (any, int, error) {
		prepared, err := c.loadDC6Direction(assets, name, palette, direction)
		if err != nil {
			return nil, 0, err
		}
		frames, bounds, err := normalizedDC6Frames(prepared.asset, 0, anchorMode, sharedBounds...)
		return preparedDC6Animation{frames: frames, bounds: bounds}, imagesWeight(frames), err
	})
	if err != nil {
		return preparedDC6Animation{}, err
	}
	return value.(preparedDC6Animation), nil
}

func (c *renderAssetCache) loadDT1File(assets fs.FS, name, palette string) (preparedDT1File, error) {
	key := name + "\x00" + palette
	value, err := c.load(assets, "dt1-file\x00"+key, func() (any, int, error) {
		asset, err := assets.Open(name)
		if err != nil {
			return nil, 0, err
		}
		defer asset.Close()
		encoded, err := io.ReadAll(asset)
		if err != nil {
			return nil, 0, err
		}
		opened, err := dt1.OpenBytes(encoded)
		if err != nil {
			return nil, 0, err
		}
		if palette != "" {
			colors, paletteErr := assetdecode.Palette(assets, palette)
			if paletteErr != nil {
				return nil, 0, paletteErr
			}
			opened.SetPalette(colors)
		}
		return preparedDT1File{file: opened}, max(len(encoded), 1), nil
	})
	if err != nil {
		return preparedDT1File{}, err
	}
	return value.(preparedDT1File), nil
}

func (c *renderAssetCache) loadDT1Tile(assets fs.FS, name, palette string, index int, view string) (preparedDT1Tile, error) {
	key := fmt.Sprintf("dt1-tile\x00%s\x00%s\x00%d\x00%s", name, palette, index, view)
	value, err := c.load(assets, key, func() (any, int, error) {
		opened, err := c.loadDT1File(assets, name, palette)
		if err != nil {
			return nil, 0, err
		}
		tile, err := opened.file.DecodeTile(index)
		if err != nil {
			return nil, 0, err
		}
		var pixels image.Image
		var imageErr error
		switch view {
		case "floor":
			pixels, imageErr = tile.FloorImageE()
		case "wall":
			pixels, imageErr = tile.WallImageE()
		case "composite", "":
			pixels, imageErr = tile.ImageE()
		default:
			return nil, 0, fmt.Errorf("dt1: unknown view %q", view)
		}
		if imageErr != nil {
			return nil, 0, fmt.Errorf("dt1: decode tile %d graphics: %w", index, imageErr)
		}
		if pixels == nil {
			height := int(tile.Height)
			if height < 0 {
				height = -height
			}
			pixels = image.NewRGBA(image.Rect(0, 0, max(1, int(tile.Width)), max(1, height)))
		}
		return preparedDT1Tile{image: pixels, tile: tile, total: opened.file.NumTiles()}, imageWeight(pixels), nil
	})
	if err != nil {
		return preparedDT1Tile{}, err
	}
	return value.(preparedDT1Tile), nil
}

func (c *renderAssetCache) loadDS1(assets fs.FS, name string, tiles []string, palette string) (image.Image, error) {
	key := name + "\x00" + strings.Join(tiles, "\x00") + "\x00" + palette
	value, err := c.load(assets, "ds1\x00"+key, func() (any, int, error) {
		preview, err := assetinspect.TexturedDS1Image(assets, name, tiles, palette)
		if err != nil {
			return nil, 0, err
		}
		return preview, imageWeight(preview), nil
	})
	if err != nil {
		return nil, err
	}
	return value.(image.Image), nil
}

func (c *renderAssetCache) loadDS1Chunks(assets fs.FS, name string, tiles []string, palette string, chunkSize int) (*maprender.Set, error) {
	key := fmt.Sprintf("ds1-chunks\x00%s\x00%s\x00%s\x00%d", name, strings.Join(tiles, "\x00"), palette, chunkSize)
	value, err := c.load(assets, key, func() (any, int, error) {
		mapData, err := gameworld.Load(assets, name, tiles)
		if err != nil {
			return nil, 0, err
		}
		chunks, err := maprender.Compose(assets, mapData, palette, chunkSize)
		if err != nil {
			return nil, 0, err
		}
		weight := 0
		for _, chunk := range chunks.Chunks {
			weight += imageWeight(chunk.Pixels)
		}
		return chunks, max(weight, 1), nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*maprender.Set), nil
}

func (c *renderAssetCache) loadWorldChunks(assets fs.FS, world *gameworld.Map, palette string, chunkSize int) (*maprender.Set, error) {
	if world == nil {
		return nil, errors.New("world map is required")
	}
	key := fmt.Sprintf("world-chunks\x00%p\x00%s\x00%d", world, palette, chunkSize)
	value, err := c.load(assets, key, func() (any, int, error) {
		chunks, err := maprender.Index(assets, world, palette, chunkSize)
		if err != nil {
			return nil, 0, err
		}
		// The index retains placement recipes and decoded DT1 block metadata, not
		// expanded RGBA. This estimate accounts for spatial/depth bookkeeping;
		// each visible raster is weighed separately in the composed cache.
		return chunks, max(len(chunks.Chunks)*256, 1), nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*maprender.Set), nil
}

func (c *renderAssetCache) loadWorldTiles(assets fs.FS, world *gameworld.Map, palette string) (*maprender.TileSet, error) {
	if world == nil {
		return nil, errors.New("world map is required")
	}
	key := fmt.Sprintf("world-tiles\x00%p\x00%s", world, palette)
	value, err := c.load(assets, key, func() (any, int, error) {
		set, placeErr := maprender.Place(assets, world, palette)
		if placeErr != nil {
			return nil, 0, placeErr
		}
		// The placement index retains DT1 block metadata, not expanded RGBA.
		// Individual unique pictures enter the composed cache on viewport demand.
		weight := len(set.Draws)*64 + len(set.Graphics)*256
		return set, max(weight, 1), nil
	})
	if err != nil {
		return nil, err
	}
	return value.(*maprender.TileSet), nil
}

func (c *renderAssetCache) loadWorldTileGraphic(assets fs.FS, world *gameworld.Map, palette string, graphicIndex int) (image.Image, error) {
	set, err := c.loadWorldTiles(assets, world, palette)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("world-tile-graphic\x00%p\x00%s\x00%d", world, palette, graphicIndex)
	value, err := c.load(assets, key, func() (any, int, error) {
		pixels, materializeErr := set.MaterializeGraphic(graphicIndex)
		return pixels, max(imageWeight(pixels), 1), materializeErr
	})
	if err != nil {
		return nil, err
	}
	return value.(image.Image), nil
}

func (c *renderAssetCache) loadWorldChunk(assets fs.FS, world *gameworld.Map, palette string, chunkSize, index int) (maprender.Chunk, error) {
	set, err := c.loadWorldChunks(assets, world, palette, chunkSize)
	if err != nil {
		return maprender.Chunk{}, err
	}
	key := fmt.Sprintf("world-chunk\x00%p\x00%s\x00%d\x00%d", world, palette, chunkSize, index)
	value, err := c.load(assets, key, func() (any, int, error) {
		chunk, materializeErr := set.Materialize(index)
		if materializeErr != nil {
			return nil, 0, materializeErr
		}
		return chunk, max(imageWeight(chunk.Pixels), 1), nil
	})
	if err != nil {
		return maprender.Chunk{}, err
	}
	return value.(maprender.Chunk), nil
}

func pushChunkSet(state *lua.LState, chunks *maprender.Set) {
	result := state.NewTable()
	result.RawSetString("width", lua.LNumber(chunks.Width))
	result.RawSetString("height", lua.LNumber(chunks.Height))
	result.RawSetString("chunk_size", lua.LNumber(chunks.ChunkSize))
	entries := state.NewTable()
	for index, chunk := range chunks.Chunks {
		entry := state.NewTable()
		entry.RawSetString("index", lua.LNumber(index))
		entry.RawSetString("x", lua.LNumber(chunk.X))
		entry.RawSetString("y", lua.LNumber(chunk.Y))
		entry.RawSetString("width", lua.LNumber(chunk.Width))
		entry.RawSetString("height", lua.LNumber(chunk.Height))
		entry.RawSetString("layer", lua.LNumber(chunk.Layer))
		entry.RawSetString("layer_name", lua.LString(chunk.Layer.String()))
		entry.RawSetString("depth", lua.LNumber(chunk.Depth))
		entries.RawSetInt(index+1, entry)
	}
	result.RawSetString("chunks", entries)
	state.Push(result)
}

func pushTileSet(state *lua.LState, set *maprender.TileSet) {
	result := state.NewTable()
	result.RawSetString("width", lua.LNumber(set.Width))
	result.RawSetString("height", lua.LNumber(set.Height))
	result.RawSetString("graphic_count", lua.LNumber(len(set.Graphics)))
	result.RawSetString("bucket_size", lua.LNumber(set.BucketSize))
	entries := state.NewTable()
	for index, draw := range set.Draws {
		entry := state.NewTable()
		entry.RawSetString("index", lua.LNumber(index))
		entry.RawSetString("graphic", lua.LNumber(draw.Graphic))
		entry.RawSetString("x", lua.LNumber(draw.Bounds.Min.X))
		entry.RawSetString("y", lua.LNumber(draw.Bounds.Min.Y))
		entry.RawSetString("width", lua.LNumber(draw.Bounds.Dx()))
		entry.RawSetString("height", lua.LNumber(draw.Bounds.Dy()))
		entry.RawSetString("layer", lua.LNumber(draw.Layer))
		entry.RawSetString("layer_name", lua.LString(draw.Layer.String()))
		entry.RawSetString("depth", lua.LNumber(draw.Depth))
		entries.RawSetInt(index+1, entry)
	}
	result.RawSetString("draws", entries)
	buckets := state.NewTable()
	for _, bucket := range set.Buckets {
		indexes := state.NewTable()
		for _, drawIndex := range bucket.Draws {
			// Lua sequences are one-based; draw.index remains the zero-based API
			// identity passed back into set_world_tile.
			indexes.Append(lua.LNumber(drawIndex + 1))
		}
		buckets.RawSetString(fmt.Sprintf("%d:%d", bucket.Column, bucket.Row), indexes)
	}
	result.RawSetString("buckets", buckets)
	state.Push(result)
}

func (c *renderAssetCache) loadDS1Collision(assets fs.FS, name string, tiles []string) (image.Image, error) {
	key := "ds1-collision\x00" + name + "\x00" + strings.Join(tiles, "\x00")
	value, err := c.load(assets, key, func() (any, int, error) {
		mapData, err := gameworld.Load(assets, name, tiles)
		if err != nil {
			return nil, 0, err
		}
		overlay := maprender.CollisionImage(mapData)
		return overlay, imageWeight(overlay), nil
	})
	if err != nil {
		return nil, err
	}
	return value.(image.Image), nil
}

func (n *ownedRenderNode) release() error {
	n.once.Do(func() {
		// Destroying a retained parent recursively invalidates every child node.
		// Scopes still own those child wrappers and release them independently;
		// an already-gone generation is successful cleanup, not a stale-handle
		// failure. Child-owned resources remain explicitly released below.
		if n.composer.Exists(n.id) {
			n.err = n.composer.Destroy(n.id)
		}
		if n.resource != (render.ResourceID{}) {
			if n.resourceRelease != nil {
				n.err = errors.Join(n.err, n.resourceRelease())
			} else {
				n.err = errors.Join(n.err, n.composer.DestroyResource(n.resource))
			}
			n.resource = render.ResourceID{}
			n.resourceRelease = nil
		}
		if n.palette != (render.ResourceID{}) {
			n.err = errors.Join(n.err, n.composer.DestroyResource(n.palette))
			n.palette = render.ResourceID{}
		}
		for _, resource := range n.owned {
			n.err = errors.Join(n.err, n.composer.DestroyResource(resource))
		}
		n.owned = nil
	})
	return n.err
}

func (n *ownedRenderNode) setPaletteQuantization(palette color.Palette) error {
	resource, err := n.composer.CreateResource(render.ResourcePalette, palette)
	if err != nil {
		return err
	}
	previous := n.palette
	if err := n.composer.Update(n.id, func(current *render.Node) { current.Palette = resource }); err != nil {
		_ = n.composer.DestroyResource(resource)
		return err
	}
	n.palette = resource
	if previous != (render.ResourceID{}) {
		return n.composer.DestroyResource(previous)
	}
	return nil
}

func (n *ownedRenderNode) clearPaletteQuantization() error {
	previous := n.palette
	if err := n.composer.Update(n.id, func(current *render.Node) { current.Palette = render.ResourceID{} }); err != nil {
		return err
	}
	n.palette = render.ResourceID{}
	if previous != (render.ResourceID{}) {
		return n.composer.DestroyResource(previous)
	}
	return nil
}

func (n *ownedRenderNode) setImage(decoded image.Image) error {
	resource, err := n.composer.CreateResource(render.ResourceTexture, decoded)
	if err != nil {
		return err
	}
	previous, previousRelease := n.resource, n.resourceRelease
	previousOwned := n.owned
	if err := n.composer.Update(n.id, func(current *render.Node) { current.Resource = resource }); err != nil {
		_ = n.composer.DestroyResource(resource)
		return err
	}
	n.resource = resource
	n.resourceRelease = nil
	n.owned = nil
	if previous != (render.ResourceID{}) {
		if previousRelease != nil {
			err = previousRelease()
		} else {
			err = n.composer.DestroyResource(previous)
		}
	}
	for _, owned := range previousOwned {
		err = errors.Join(err, n.composer.DestroyResource(owned))
	}
	return err
}

func (n *ownedRenderNode) setSharedImage(key string, decoded image.Image) error {
	resource, release, err := n.pool.acquire(key, decoded)
	if err != nil {
		return err
	}
	previous, previousRelease, previousOwned := n.resource, n.resourceRelease, n.owned
	if err := n.composer.Update(n.id, func(current *render.Node) { current.Resource = resource }); err != nil {
		_ = release()
		return err
	}
	n.resource, n.resourceRelease, n.owned = resource, release, nil
	if previous != (render.ResourceID{}) {
		if previousRelease != nil {
			err = previousRelease()
		} else {
			err = n.composer.DestroyResource(previous)
		}
	}
	for _, owned := range previousOwned {
		err = errors.Join(err, n.composer.DestroyResource(owned))
	}
	return err
}

func (n *ownedRenderNode) setAnimation(frames []image.Image, duration time.Duration, loop string, initialSeek ...time.Duration) error {
	return n.setAnimationKeyed(frames, nil, duration, loop, initialSeek...)
}

func (n *ownedRenderNode) setAnimationKeyed(frames []image.Image, textureKeys []string, duration time.Duration, loop string, initialSeek ...time.Duration) error {
	textures := make([]render.ResourceID, 0, len(frames))
	owned := make([]render.ResourceID, 0, len(frames))
	duplicates := make(map[rgbaFrameDigest]render.ResourceID)
	cleanup := func() {
		for _, texture := range owned {
			_ = n.composer.DestroyResource(texture)
		}
	}
	for frameIndex, frame := range frames {
		key, shareable := rgbaFrameKey(frame)
		if shareable {
			if texture, exists := duplicates[key]; exists {
				textures = append(textures, texture)
				continue
			}
		}
		var texture render.ResourceID
		var err error
		if frameIndex < len(textureKeys) && textureKeys[frameIndex] != "" {
			texture, err = n.composer.CreateTexture(frame, textureKeys[frameIndex])
		} else {
			texture, err = n.composer.CreateResource(render.ResourceTexture, frame)
		}
		if err != nil {
			cleanup()
			return err
		}
		textures = append(textures, texture)
		owned = append(owned, texture)
		if shareable {
			duplicates[key] = texture
		}
	}
	durations := make([]time.Duration, len(textures))
	for index := range durations {
		durations[index] = duration
	}
	animation, err := n.composer.CreateResource(render.ResourceAnimation, render.AnimationData{Frames: textures, Durations: durations, Loop: loop})
	if err != nil {
		cleanup()
		return err
	}
	previous, previousRelease, previousOwned := n.resource, n.resourceRelease, n.owned
	seek := time.Duration(0)
	if len(initialSeek) > 0 && initialSeek[0] > 0 {
		seek = initialSeek[0]
	}
	if err := n.composer.Update(n.id, func(current *render.Node) {
		current.Resource = animation
		current.AnimationPaused = false
		current.AnimationSeek = seek
		current.AnimationSeekRevision++
	}); err != nil {
		_ = n.composer.DestroyResource(animation)
		cleanup()
		return err
	}
	n.resource, n.resourceRelease, n.owned = animation, nil, owned
	if previous != (render.ResourceID{}) {
		if previousRelease != nil {
			err = previousRelease()
		} else {
			err = n.composer.DestroyResource(previous)
		}
	}
	for _, owned := range previousOwned {
		err = errors.Join(err, n.composer.DestroyResource(owned))
	}
	return err
}

func rgbaFrameKey(frame image.Image) (rgbaFrameDigest, bool) {
	rgba, ok := frame.(*image.RGBA)
	if !ok || rgba.Stride != rgba.Bounds().Dx()*4 {
		return rgbaFrameDigest{}, false
	}
	bounds := rgba.Bounds()
	size := bounds.Dx() * bounds.Dy() * 4
	start := rgba.PixOffset(bounds.Min.X, bounds.Min.Y)
	if size <= 0 || start < 0 || start+size > len(rgba.Pix) {
		return rgbaFrameDigest{}, false
	}
	return rgbaFrameDigest{width: bounds.Dx(), height: bounds.Dy(), pixels: sha256.Sum256(rgba.Pix[start : start+size])}, true
}

func (n *ownedRenderNode) requireAnimation() error {
	if n.resource == (render.ResourceID{}) {
		return errors.New("render node has no animation")
	}
	resource, err := n.composer.ResourceSnapshot(n.resource)
	if err != nil {
		return err
	}
	if resource.Kind != render.ResourceAnimation {
		return errors.New("render node resource is not an animation")
	}
	return nil
}

func (n *ownedRenderNode) cofFrames(cofName, palette string, direction int, paths map[string]string) ([]image.Image, *cof.COF, image.Rectangle, error) {
	asset, err := n.cache.loadCOF(n.assets, cofName)
	if err != nil {
		return nil, nil, image.Rectangle{}, err
	}
	if direction < 0 || direction >= asset.NumberOfDirections {
		return nil, nil, image.Rectangle{}, fmt.Errorf("COF direction %d is out of range", direction)
	}
	dccDirection, err := dccDirectionForCOF(direction, asset.NumberOfDirections)
	if err != nil {
		return nil, nil, image.Rectangle{}, err
	}
	layers := make(map[cof.CompositeType]cof.CofLayer, len(asset.CofLayers))
	decoded := make(map[cof.CompositeType]preparedDCCDirection)
	for _, layer := range asset.CofLayers {
		name, ok := paths[layer.Type.String()]
		if !ok || name == "" {
			continue
		}
		component, err := n.cache.loadDCCDirection(n.assets, name, palette, dccDirection)
		if err != nil {
			return nil, nil, image.Rectangle{}, fmt.Errorf("COF layer %s: %w", layer.Type, err)
		}
		layers[layer.Type], decoded[layer.Type] = layer, component
	}
	// A retained animation node has one anchor and native quad. Every composite
	// frame must therefore use one shared canvas even though individual DCC
	// frame rectangles move and change size. Per-frame canvases make the quad
	// resize around its center (visible jitter) and are unsafe for backends that
	// retain texture geometry from the first animation frame.
	var animationBounds image.Rectangle
	for _, component := range decoded {
		for _, frame := range component.direction.Frames() {
			if animationBounds.Empty() {
				animationBounds = frame.Bounds()
			} else {
				animationBounds = animationBounds.Union(frame.Bounds())
			}
		}
	}
	if animationBounds.Empty() {
		return nil, nil, image.Rectangle{}, errors.New("COF composition has no component animation bounds")
	}

	frames := make([]image.Image, asset.FramesPerDirection)
	var canvas image.Rectangle
	for frameIndex := range frames {
		components := make(map[cof.CompositeType]compositeFrame, len(decoded))
		for componentType, component := range decoded {
			directionFrames := component.direction.Frames()
			if frameIndex >= len(directionFrames) {
				return nil, nil, image.Rectangle{}, fmt.Errorf("COF layer %s lacks frame %d", componentType, frameIndex)
			}
			frame := directionFrames[frameIndex]
			components[componentType] = compositeFrame{indices: frame.PixelData, palette: component.palette, bounds: frame.Bounds(), layer: layers[componentType]}
		}
		frames[frameIndex], err = composeCOFFrame(asset, direction, frameIndex, components, animationBounds)
		if err != nil {
			return nil, nil, image.Rectangle{}, err
		}
		canvas = shadowCanvasBounds(animationBounds, components)
	}
	return frames, asset, canvas, nil
}

func compositeCacheKey(cofName, palette string, direction int, paths map[string]string) string {
	components := make([]string, 0, len(paths))
	for component, path := range paths {
		components = append(components, component+"="+path)
	}
	sort.Strings(components)
	return fmt.Sprintf("cof-animation\x00%s\x00%s\x00%d\x00%s", cofName, palette, direction, strings.Join(components, "\x00"))
}

// cachedCOFAnimation retains the expensive decoded-and-composed RGBA frames.
// Changing direction may still replace a tiny semantic animation resource, but
// it no longer decodes DCCs or recomposites every layer on the input path.
func (n *ownedRenderNode) cachedCOFAnimation(cofName, palette string, direction int, paths map[string]string) (preparedCOFAnimation, error) {
	key := compositeCacheKey(cofName, palette, direction, paths)
	value, err := n.cache.load(n.assets, key, func() (any, int, error) {
		frames, asset, canvas, err := n.cofFrames(cofName, palette, direction, paths)
		return preparedCOFAnimation{frames: frames, asset: asset, origin: image.Pt(-canvas.Min.X, -canvas.Min.Y)}, imagesWeight(frames), err
	})
	if err != nil {
		return preparedCOFAnimation{}, err
	}
	prepared := value.(preparedCOFAnimation)
	prepared.keys = make([]string, len(prepared.frames))
	generation := n.cache.currentGeneration(n.assets)
	for index := range prepared.keys {
		prepared.keys[index] = fmt.Sprintf("cof:%d:%s:frame:%d", generation, key, index)
	}
	return prepared, nil
}

func luaComponentPaths(state *lua.LState, index int) map[string]string {
	result := make(map[string]string)
	table := state.CheckTable(index)
	table.ForEach(func(key, value lua.LValue) {
		if key.Type() == lua.LTString && value.Type() == lua.LTString {
			result[key.String()] = value.String()
		}
	})
	return result
}

// normalizedDC6Frames places every cropped frame on one shared canvas using
// the DC6 anchor offsets. The retained node can then animate at a fixed world
// position without jitter when individual frame bounds change.
func dc6FrameTop(frame *dc6.Frame) int {
	if frame.Flipped > 0 {
		return int(frame.OffsetY)
	}
	return int(frame.OffsetY) - int(frame.Height) + 1
}

func dc6AnimationBounds(asset *dc6.DC6, direction int) image.Rectangle {
	var bounds image.Rectangle
	for index, frame := range asset.Directions[direction].Frames {
		top := dc6FrameTop(frame)
		frameBounds := image.Rect(int(frame.OffsetX), top, int(frame.OffsetX+int32(frame.Width)), top+int(frame.Height))
		if index == 0 {
			bounds = frameBounds
		} else {
			bounds = bounds.Union(frameBounds)
		}
	}
	return bounds
}

func dc6FixedAnimationBounds(asset *dc6.DC6, direction int) image.Rectangle {
	frames := asset.Directions[direction].Frames
	if len(frames) == 0 {
		return image.Rectangle{}
	}
	var width, height int
	for _, frame := range frames {
		width = max(width, int(frame.Width))
		height = max(height, int(frame.Height))
	}
	top := int(frames[0].OffsetY) - height + 1
	if frames[0].Flipped > 0 {
		top = int(frames[0].OffsetY)
	}
	return image.Rect(int(frames[0].OffsetX), top,
		int(frames[0].OffsetX)+width, top+height)
}

func normalizedDC6Frames(asset *dc6.DC6, direction int, anchorMode string, sharedBounds ...image.Rectangle) ([]image.Image, image.Rectangle, error) {
	frames := asset.Directions[direction].Frames
	var bounds image.Rectangle
	if len(sharedBounds) > 0 {
		bounds = sharedBounds[0]
	} else if anchorMode == "first-frame" && len(frames) > 0 {
		var width, height int
		for _, frame := range frames {
			width = max(width, int(frame.Width))
			height = max(height, int(frame.Height))
		}
		top := int(frames[0].OffsetY) - height + 1
		if frames[0].Flipped > 0 {
			top = int(frames[0].OffsetY)
		}
		bounds = image.Rect(int(frames[0].OffsetX), top, int(frames[0].OffsetX)+width, top+height)
	}
	for index, frame := range frames {
		if len(sharedBounds) > 0 {
			continue
		}
		if anchorMode == "first-frame" {
			continue
		}
		top := dc6FrameTop(frame)
		frameBounds := image.Rect(int(frame.OffsetX), top, int(frame.OffsetX+int32(frame.Width)), top+int(frame.Height))
		if index == 0 {
			bounds = frameBounds
		} else {
			bounds = bounds.Union(frameBounds)
		}
	}
	if bounds.Empty() {
		return nil, image.Rectangle{}, errors.New("DC6 animation has no visible frames")
	}
	result := make([]image.Image, len(frames))
	for index, frame := range frames {
		decoded, err := assetdecode.FrameImage(asset, frame)
		if err != nil {
			return nil, image.Rectangle{}, err
		}
		canvas := image.NewRGBA(image.Rectangle{Max: bounds.Size()})
		position := image.Point{}
		if anchorMode == "first-frame" && len(sharedBounds) > 0 {
			position = image.Pt(int(frames[0].OffsetX)-bounds.Min.X, dc6FrameTop(frames[0])-bounds.Min.Y)
		} else if anchorMode != "first-frame" {
			position = image.Pt(int(frame.OffsetX)-bounds.Min.X, dc6FrameTop(frame)-bounds.Min.Y)
		}
		draw.Draw(canvas, decoded.Bounds().Add(position), decoded, decoded.Bounds().Min, draw.Src)
		result[index] = canvas
	}
	return result, bounds, nil
}

// combinedDC6Pages reconstructs tiled logical images using Diablo II's 256px
// DC6 page convention. This matches Riiablo's DC6Parameters.COMBINE behavior:
// a short frame terminates the column/row scan and repeated grids are pages.
func combinedDC6Pages(asset *dc6.DC6, direction int) ([]image.Image, error) {
	if asset == nil || direction < 0 || direction >= len(asset.Directions) {
		return nil, fmt.Errorf("DC6 combined direction %d is out of range", direction)
	}
	frames := asset.Directions[direction].Frames
	if len(frames) == 0 {
		return nil, errors.New("DC6 combined direction has no frames")
	}
	const pageSize = 256
	columns, width := 0, 0
	for _, frame := range frames {
		columns++
		width += int(frame.Width)
		if frame.Width < pageSize {
			break
		}
	}
	rows, height := 0, 0
	for index := 0; index < len(frames); index += columns {
		rows++
		height += int(frames[index].Height)
		if frames[index].Height < pageSize {
			break
		}
	}
	framesPerPage := rows * columns
	if framesPerPage <= 0 || len(frames)%framesPerPage != 0 {
		return nil, fmt.Errorf("DC6 combined grid %dx%d does not divide %d frames", columns, rows, len(frames))
	}
	pages := make([]image.Image, 0, len(frames)/framesPerPage)
	frameIndex := 0
	for page := 0; page < len(frames)/framesPerPage; page++ {
		canvas := image.NewRGBA(image.Rect(0, 0, width, height))
		y := 0
		for row := 0; row < rows; row++ {
			x := 0
			for column := 0; column < columns; column++ {
				frame := frames[frameIndex]
				frameIndex++
				decoded, err := assetdecode.FrameImage(asset, frame)
				if err != nil {
					return nil, err
				}
				position := image.Pt(x, y)
				draw.Draw(canvas, decoded.Bounds().Add(position), decoded, decoded.Bounds().Min, draw.Over)
				x += int(frame.Width)
			}
			y += pageSize
		}
		pages = append(pages, canvas)
	}
	return pages, nil
}

func horizontalDC6Strip(asset *dc6.DC6, direction int) (image.Image, error) {
	if asset == nil || direction < 0 || direction >= len(asset.Directions) {
		return nil, fmt.Errorf("DC6 strip direction %d is out of range", direction)
	}
	frames := asset.Directions[direction].Frames
	if len(frames) == 0 {
		return nil, errors.New("DC6 strip has no frames")
	}
	width, height := 0, 0
	for _, frame := range frames {
		width += int(frame.Width)
		height = max(height, int(frame.Height))
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	x := 0
	for _, frame := range frames {
		decoded, err := assetdecode.FrameImage(asset, frame)
		if err != nil {
			return nil, err
		}
		position := image.Pt(x, height-int(frame.Height))
		draw.Draw(canvas, decoded.Bounds().Add(position), decoded, decoded.Bounds().Min, draw.Over)
		x += int(frame.Width)
	}
	return canvas, nil
}

// RenderModule exposes backend-neutral retained composition to scoped Lua
// components. Nodes are automatically destroyed with their component scope.
func RenderModule(runtime *Runtime, composer *render.Composer) Module {
	return RenderModuleWithAssets(runtime, composer, nil)
}

// RenderModuleWithAssets additionally lets nodes decode standard image assets
// from the layered content filesystem. Scene-demanded decoding occurs on the
// Lua owner, while explicit preloads use bounded background workers. Renderer
// upload always remains queued for the renderer thread.
func RenderModuleWithAssets(runtime *Runtime, composer *render.Composer, assets fs.FS) Module {
	return NewRenderCapability(runtime, composer, assets).Module()
}

// Module returns the versioned Lua render capability.
func (r *RenderCapability) Module() Module {
	runtime, composer, assets, cache, preloads, pool := r.runtime, r.composer, r.assets, r.cache, r.preloads, r.pool
	return Module{Name: "engine.render/v1", Help: documentedModule("Create and inspect retained presentation nodes.", map[string]CommandHelp{
		"diagnostics":          commandHelp("engine.render.diagnostics()", "Return decoded-asset cache and retained-renderer diagnostics."),
		"preload":              commandHelp("engine.render.preload(requests)", "Decode assets asynchronously and return a preload job identifier."),
		"preload_status":       commandHelp("engine.render.preload_status(job)", "Return progress and errors for a preload job."),
		"preload_release":      commandHelp("engine.render.preload_release(job)", "Release bookkeeping for a completed preload job."),
		"ds1_dependencies":     commandHelp("engine.render.ds1_dependencies(map)", "Resolve the mounted DT1 libraries declared by a DS1 stamp."),
		"ds1_chunks":           commandHelp("engine.render.ds1_chunks(map, tiles, palette [, chunk_size])", "Return sparse DS1 chunk geometry after CPU composition."),
		"world_chunks":         commandHelp("engine.render.world_chunks(world, palette [, chunk_size])", "Return sparse chunk geometry for an assembled authoritative world map."),
		"world_tiles":          commandHelp("engine.render.world_tiles(world, palette)", "Return shared DT1 graphic placement geometry for an authoritative world map."),
		"assets_available":     commandHelp("engine.render.assets_available()", "Report whether asset-backed rendering is available."),
		"asset_exists":         commandHelp("engine.render.asset_exists(path)", "Report whether a render asset exists."),
		"dc6_animation_bounds": commandHelp("engine.render.dc6_animation_bounds(path)", "Inspect the normalized bounds of a DC6 animation."),
		"cof_info":             commandHelp("engine.render.cof_info(path)", "Inspect COF layer and animation metadata."),
		"animdata_info":        commandHelp("engine.render.animdata_info(key)", "Read typed rate and frame-event metadata from an AnimData record source."),
		"create":               commandHelp("engine.render.create()", "Create a scoped retained presentation node."),
	}, map[string]TypeHelp{renderNodeType: {Summary: "A scoped retained presentation node.", Methods: map[string]CommandHelp{
		"set_position":               commandHelp("node:set_position(x, y)", "Set the node position."),
		"set_scale":                  commandHelp("node:set_scale(x, y)", "Set the node scale."),
		"set_z":                      commandHelp("node:set_z(z)", "Set the node draw priority."),
		"set_rotation":               commandHelp("node:set_rotation(degrees)", "Set the node rotation."),
		"set_blend":                  commandHelp("node:set_blend(mode)", "Set the node blend mode."),
		"set_palette_quantization":   commandHelp("node:set_palette_quantization(path)", "Quantize the node through a display palette."),
		"clear_palette_quantization": commandHelp("node:clear_palette_quantization()", "Disable node palette quantization."),
		"set_visible":                commandHelp("node:set_visible(visible)", "Show or hide the node."),
		"set_origin":                 commandHelp("node:set_origin(x, y)", "Set the normalized texture pivot used at the node position."),
		"set_clip":                   commandHelp("node:set_clip(x, y, width, height)", "Set the node clip rectangle."),
		"clear_clip":                 commandHelp("node:clear_clip()", "Remove the node clip rectangle."),
		"set_image":                  commandHelp("node:set_image(path)", "Render a decoded image asset."),
		"set_ds1":                    commandHelp("node:set_ds1(map, tiles, palette)", "Render a DS1 map using mounted DT1 tiles and a palette."),
		"set_ds1_chunk":              commandHelp("node:set_ds1_chunk(map, tiles, palette, chunk_index [, chunk_size])", "Render one sparse DS1 map chunk and return its map-space geometry."),
		"set_world_chunk":            commandHelp("node:set_world_chunk(world, palette, chunk_index [, chunk_size])", "Render one sparse chunk from an assembled authoritative world map."),
		"set_world_tile":             commandHelp("node:set_world_tile(world, palette, draw_index)", "Render one placement borrowing a shared immutable DT1 texture."),
		"set_world_collision_region": commandHelp("node:set_world_collision_region(world, left, top, right, bottom)", "Render a bounded authoritative subtile collision diagnostic and return map-pixel geometry."),
		"set_world_tile_region":      commandHelp("node:set_world_tile_region(world, left, top, right, bottom)", "Render bounded authoritative tile/subtile projection geometry and return map-pixel geometry."),
		"set_ds1_collision":          commandHelp("node:set_ds1_collision(map, tiles)", "Render a diagnostic DT1 subtile collision overlay for a DS1 map."),
		"set_dt1":                    commandHelp("node:set_dt1(path, palette, tile_index[, view])", "Render one lazy-decoded DT1 tile and return its dimensions and metadata."),
		"set_dc6":                    commandHelp("node:set_dc6(path, frame [, options])", "Render one DC6 frame."),
		"set_dc6_combined":           commandHelp("node:set_dc6_combined(path [, options])", "Reconstruct and render one tiled DC6 page."),
		"set_dc6_strip":              commandHelp("node:set_dc6_strip(path [, options])", "Join every frame in one DC6 direction horizontally."),
		"set_dc6_animation":          commandHelp("node:set_dc6_animation(path [, options])", "Render a DC6 animation."),
		"set_dcc":                    commandHelp("node:set_dcc(path [, options])", "Render a DCC asset."),
		"set_dcc_animation":          commandHelp("node:set_dcc_animation(path [, options])", "Render a DCC animation."),
		"set_cof":                    commandHelp("node:set_cof(path [, options])", "Render a COF composite."),
		"set_cof_animation":          commandHelp("node:set_cof_animation(path, palette, direction, components [, loop, rate, seek_seconds])", "Render a cached COF composite with resolved rate and optional preserved playback position."),
		"set_text":                   commandHelp("node:set_text(text [, options])", "Render bitmap-font text."),
		"animation_pause":            commandHelp("node:animation_pause()", "Pause the node animation."),
		"animation_resume":           commandHelp("node:animation_resume()", "Resume the node animation."),
		"animation_seek":             commandHelp("node:animation_seek(frame)", "Seek the node animation."),
		"fill_rect":                  commandHelp("node:fill_rect(width, height, color)", "Render a filled rectangle."),
		"destroy":                    commandHelp("node:destroy()", "Destroy and release the node."),
		"exists":                     commandHelp("node:exists()", "Report whether this retained node handle is still live."),
	}}}), Loader: func(state *lua.LState) int {
		registerRenderNodeType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"ds1_dependencies": func(state *lua.LState) int {
				if assets == nil {
					state.RaiseError("render asset filesystem is unavailable")
					return 0
				}
				paths, err := assetinspect.DS1TilePaths(assets, state.CheckString(1))
				if err != nil {
					state.RaiseError("resolving DS1 dependencies: %v", err)
					return 0
				}
				result := state.NewTable()
				for index, path := range paths {
					result.RawSetInt(index+1, lua.LString(path))
				}
				state.Push(result)
				return 1
			},
			"world_tiles": func(state *lua.LState) int {
				if assets == nil {
					state.RaiseError("render asset filesystem is unavailable")
					return 0
				}
				set, err := cache.loadWorldTiles(assets, checkWorldMap(state, 1), state.CheckString(2))
				if err != nil {
					state.RaiseError("indexing shared world tiles: %v", err)
					return 0
				}
				pushTileSet(state, set)
				return 1
			},
			"world_chunks": func(state *lua.LState) int {
				if assets == nil {
					state.RaiseError("render asset filesystem is unavailable")
					return 0
				}
				world := checkWorldMap(state, 1)
				chunks, err := cache.loadWorldChunks(assets, world, state.CheckString(2), state.OptInt(3, maprender.DefaultChunkSize))
				if err != nil {
					state.RaiseError("rendering world chunks: %v", err)
					return 0
				}
				pushChunkSet(state, chunks)
				return 1
			},
			"ds1_chunks": func(state *lua.LState) int {
				if assets == nil {
					state.RaiseError("render asset filesystem is unavailable")
					return 0
				}
				mapName := state.CheckString(1)
				tileTable := state.CheckTable(2)
				tiles := make([]string, 0, tileTable.Len())
				for index := 1; index <= tileTable.Len(); index++ {
					value, ok := tileTable.RawGetInt(index).(lua.LString)
					if !ok || value == "" {
						state.RaiseError("DS1 tile %d must be a non-empty path", index)
						return 0
					}
					tiles = append(tiles, string(value))
				}
				chunks, err := cache.loadDS1Chunks(assets, mapName, tiles, state.CheckString(3), state.OptInt(4, maprender.DefaultChunkSize))
				if err != nil {
					state.RaiseError("inspecting DS1 chunks %q: %v", mapName, err)
					return 0
				}
				result := state.NewTable()
				result.RawSetString("width", lua.LNumber(chunks.Width))
				result.RawSetString("height", lua.LNumber(chunks.Height))
				result.RawSetString("chunk_size", lua.LNumber(chunks.ChunkSize))
				entries := state.NewTable()
				for index, chunk := range chunks.Chunks {
					entry := state.NewTable()
					entry.RawSetString("index", lua.LNumber(index))
					entry.RawSetString("x", lua.LNumber(chunk.X))
					entry.RawSetString("y", lua.LNumber(chunk.Y))
					entry.RawSetString("width", lua.LNumber(chunk.Pixels.Bounds().Dx()))
					entry.RawSetString("height", lua.LNumber(chunk.Pixels.Bounds().Dy()))
					entry.RawSetString("layer", lua.LNumber(chunk.Layer))
					entry.RawSetString("layer_name", lua.LString(chunk.Layer.String()))
					entry.RawSetString("depth", lua.LNumber(chunk.Depth))
					entries.RawSetInt(index+1, entry)
				}
				result.RawSetString("chunks", entries)
				objects := state.NewTable()
				for index, object := range chunks.Objects {
					entry := state.NewTable()
					entry.RawSetString("type", lua.LNumber(object.Type))
					entry.RawSetString("id", lua.LNumber(object.ID))
					entry.RawSetString("x", lua.LNumber(object.X))
					entry.RawSetString("y", lua.LNumber(object.Y))
					entry.RawSetString("flags", lua.LNumber(object.Flags))
					objects.RawSetInt(index+1, entry)
				}
				result.RawSetString("objects", objects)
				state.Push(result)
				return 1
			},
			"preload": func(state *lua.LState) int {
				if assets == nil {
					state.RaiseError("render asset filesystem is unavailable")
					return 0
				}
				requests, err := luaPreloadRequests(state, 1)
				if err != nil {
					state.ArgError(1, err.Error())
					return 0
				}
				state.Push(lua.LNumber(preloads.Start(requests)))
				return 1
			},
			"preload_status": func(state *lua.LState) int {
				status, ok := preloads.Status(uint64(state.CheckInt(1)))
				if !ok {
					state.Push(lua.LNil)
					return 1
				}
				result := state.NewTable()
				result.RawSetString("total", lua.LNumber(status.Total))
				result.RawSetString("completed", lua.LNumber(status.Completed))
				result.RawSetString("failed", lua.LNumber(status.Failed))
				result.RawSetString("done", lua.LBool(status.Done))
				errors := state.NewTable()
				for _, message := range status.Errors {
					errors.Append(lua.LString(message))
				}
				result.RawSetString("errors", errors)
				state.Push(result)
				return 1
			},
			"preload_release": func(state *lua.LState) int {
				state.Push(lua.LBool(preloads.Forget(uint64(state.CheckInt(1)))))
				return 1
			},
			"diagnostics": func(state *lua.LState) int {
				cache.mu.Lock()
				encoded := cache.encoded.Diagnostics()
				decodedTier := cache.decoded.Diagnostics()
				composed := cache.composed.Diagnostics()
				worldTier := cache.world.Diagnostics()
				decoded := combinedCacheStats(encoded, decodedTier, composed, worldTier)
				decodeCalls, decodeTime := cache.decodeCalls, cache.decodeTime
				stages := make(map[string]DecodeStageDiagnostics, len(cache.stages))
				for name, stage := range cache.stages {
					stages[name] = stage
				}
				cache.mu.Unlock()
				retained := composer.Diagnostics()
				result := state.NewTable()
				result.RawSetString("decoded_entries", lua.LNumber(decoded.Entries))
				result.RawSetString("decoded_weight", lua.LNumber(decoded.Weight))
				result.RawSetString("decoded_budget", lua.LNumber(decoded.Budget))
				result.RawSetString("cache_hits", lua.LNumber(decoded.Hits))
				result.RawSetString("cache_misses", lua.LNumber(decoded.Misses))
				result.RawSetString("cache_evictions", lua.LNumber(decoded.Evictions))
				result.RawSetString("encoded_weight", lua.LNumber(encoded.Weight))
				result.RawSetString("direction_weight", lua.LNumber(decodedTier.Weight))
				result.RawSetString("composed_weight", lua.LNumber(composed.Weight))
				result.RawSetString("world_chunk_weight", lua.LNumber(worldTier.Weight))
				result.RawSetString("world_chunk_budget", lua.LNumber(worldTier.Budget))
				result.RawSetString("encoded_evictions", lua.LNumber(encoded.Evictions))
				result.RawSetString("direction_evictions", lua.LNumber(decodedTier.Evictions))
				result.RawSetString("composed_evictions", lua.LNumber(composed.Evictions))
				result.RawSetString("world_chunk_evictions", lua.LNumber(worldTier.Evictions))
				result.RawSetString("active_nodes", lua.LNumber(retained.ActiveNodes))
				result.RawSetString("active_resources", lua.LNumber(retained.ActiveResources))
				result.RawSetString("pending_commands", lua.LNumber(retained.Pending))
				result.RawSetString("pending_warm_textures", lua.LNumber(retained.WarmPending))
				result.RawSetString("pending_warm_bytes", lua.LNumber(retained.WarmPendingBytes))
				result.RawSetString("decode_calls", lua.LNumber(decodeCalls))
				result.RawSetString("decode_time_ms", lua.LNumber(float64(decodeTime)/float64(time.Millisecond)))
				stageTable := state.NewTable()
				for name, stage := range stages {
					values := state.NewTable()
					values.RawSetString("calls", lua.LNumber(stage.Calls))
					values.RawSetString("time_ms", lua.LNumber(float64(stage.Time)/float64(time.Millisecond)))
					stageTable.RawSetString(name, values)
				}
				result.RawSetString("decode_stages", stageTable)
				state.Push(result)
				return 1
			},
			"assets_available": func(state *lua.LState) int {
				state.Push(lua.LBool(assets != nil))
				return 1
			},
			"asset_exists": func(state *lua.LState) int {
				if assets == nil {
					state.Push(lua.LFalse)
					return 1
				}
				_, err := fs.Stat(assets, state.CheckString(1))
				state.Push(lua.LBool(err == nil))
				return 1
			},
			"dc6_animation_bounds": func(state *lua.LState) int {
				if assets == nil {
					state.RaiseError("render asset filesystem is unavailable")
					return 0
				}
				direction := state.OptInt(3, 0)
				prepared, err := cache.loadDC6Direction(assets, state.CheckString(1), state.OptString(2, ""), direction)
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				var bounds image.Rectangle
				if state.OptString(4, "offsets") == "first-frame" {
					bounds = dc6FixedAnimationBounds(prepared.asset, 0)
				} else {
					bounds = dc6AnimationBounds(prepared.asset, 0)
				}
				state.Push(lua.LNumber(bounds.Min.X))
				state.Push(lua.LNumber(bounds.Min.Y))
				state.Push(lua.LNumber(bounds.Max.X))
				state.Push(lua.LNumber(bounds.Max.Y))
				return 4
			},
			"cof_info": func(state *lua.LState) int {
				if assets == nil {
					state.Push(lua.LNil)
					state.Push(lua.LString("render asset filesystem is unavailable"))
					return 2
				}
				asset, err := cache.loadCOF(assets, state.CheckString(1))
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				result := state.NewTable()
				result.RawSetString("directions", lua.LNumber(asset.NumberOfDirections))
				result.RawSetString("frames", lua.LNumber(asset.FramesPerDirection))
				result.RawSetString("speed", lua.LNumber(asset.Speed))
				layers := state.NewTable()
				for _, layer := range asset.CofLayers {
					entry := state.NewTable()
					entry.RawSetString("type", lua.LString(layer.Type.String()))
					entry.RawSetString("shadow", lua.LNumber(layer.Shadow))
					entry.RawSetString("selectable", lua.LBool(layer.Selectable))
					entry.RawSetString("transparent", lua.LBool(layer.Transparent))
					entry.RawSetString("draw_effect", lua.LNumber(layer.DrawEffect))
					entry.RawSetString("weapon_class", lua.LString(layer.WeaponClass.String()))
					layers.Append(entry)
				}
				result.RawSetString("layers", layers)
				events := state.NewTable()
				for _, event := range asset.AnimationFrames {
					events.Append(lua.LNumber(event))
				}
				result.RawSetString("events", events)
				priority := state.NewTable()
				for _, direction := range asset.Priority {
					directionTable := state.NewTable()
					for _, frame := range direction {
						frameTable := state.NewTable()
						for _, layer := range frame {
							frameTable.Append(lua.LString(layer.String()))
						}
						directionTable.Append(frameTable)
					}
					priority.Append(directionTable)
				}
				result.RawSetString("priority", priority)
				state.Push(result)
				return 1
			},
			"animdata_info": func(state *lua.LState) int {
				if assets == nil {
					state.Push(lua.LNil)
					state.Push(lua.LString("render asset filesystem is unavailable"))
					return 2
				}
				catalog, err := cache.loadAnimationData(assets, "data/global/AnimData.d2")
				if err != nil {
					state.Push(lua.LNil)
					state.Push(lua.LString(err.Error()))
					return 2
				}
				record := catalog.GetRecord(strings.ToUpper(state.CheckString(1)))
				if record == nil {
					state.Push(lua.LNil)
					return 1
				}
				result := state.NewTable()
				result.RawSetString("name", lua.LString(record.Name()))
				result.RawSetString("frames", lua.LNumber(record.FramesPerDirection()))
				result.RawSetString("speed", lua.LNumber(record.Speed()))
				events := state.NewTable()
				for frame, event := range record.Events() {
					events.RawSetInt(frame+1, lua.LNumber(event))
				}
				result.RawSetString("events", events)
				state.Push(result)
				return 1
			},
			"create": func(state *lua.LState) int {
				scope, err := runtime.requireActiveScope()
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				layer, err := parseLayer(state.CheckString(1))
				if err != nil {
					state.RaiseError("%v", err)
					return 0
				}
				var parent render.NodeID
				if state.GetTop() >= 2 && state.Get(2) != lua.LNil {
					parent = checkRenderNode(state, 2).id
				}
				id, err := composer.Create(parent, layer)
				if err != nil {
					state.RaiseError("creating render node: %v", err)
					return 0
				}
				node := &ownedRenderNode{composer: composer, id: id, assets: assets, cache: cache, pool: pool}
				if err := scope.Add(node.release); err != nil {
					_ = node.release()
					state.RaiseError("owning render node: %v", err)
					return 0
				}
				userData := state.NewUserData()
				userData.Value = node
				state.SetMetatable(userData, state.GetTypeMetatable(renderNodeType))
				state.Push(userData)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func registerRenderNodeType(state *lua.LState) {
	meta := state.NewTypeMetatable(renderNodeType)
	state.SetField(meta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"exists": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			state.Push(lua.LBool(node.composer.Exists(node.id)))
			return 1
		},
		"set_position": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			if err := node.composer.Update(node.id, func(current *render.Node) { current.X, current.Y = x, y }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_scale": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			if err := node.composer.Update(node.id, func(current *render.Node) { current.ScaleX, current.ScaleY = x, y }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_z": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			z := state.CheckInt(2)
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Z = z }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_rotation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			rotation := float64(state.CheckNumber(2))
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Rotation = rotation }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_blend": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			blend := state.CheckString(2)
			switch blend {
			case "alpha", "additive", "multiply", "add-colors", "subtract-colors", "screen":
			default:
				state.ArgError(2, "blend must be alpha, additive, multiply, add-colors, subtract-colors, or screen")
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Blend = blend }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_palette_quantization": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			name := state.CheckString(2)
			palette, err := assetdecode.DisplayPalette(node.assets, name)
			if err != nil {
				state.RaiseError("loading display palette %q: %v", name, err)
				return 0
			}
			if err := node.setPaletteQuantization(palette); err != nil {
				state.RaiseError("applying display palette: %v", err)
			}
			return 0
		},
		"clear_palette_quantization": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.clearPaletteQuantization(); err != nil {
				state.RaiseError("clearing display palette: %v", err)
			}
			return 0
		},
		"set_visible": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			visible := state.CheckBool(2)
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Visible = visible }); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_origin": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			if err := node.composer.Update(node.id, func(current *render.Node) { current.OriginX, current.OriginY = x, y }); err != nil {
				state.RaiseError("updating render node origin: %v", err)
			}
			return 0
		},
		"set_clip": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			x, y := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			width, height := float64(state.CheckNumber(4)), float64(state.CheckNumber(5))
			if width <= 0 || height <= 0 {
				state.ArgError(4, "clip width and height must be positive")
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) {
				current.Clip = &render.Rect{X: x, Y: y, Width: width, Height: height}
			}); err != nil {
				state.RaiseError("updating render clip: %v", err)
			}
			return 0
		},
		"clear_clip": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.composer.Update(node.id, func(current *render.Node) { current.Clip = nil }); err != nil {
				state.RaiseError("clearing render clip: %v", err)
			}
			return 0
		},
		"set_image": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			decoded, err := node.cache.loadImage(node.assets, fileName)
			if err != nil {
				state.RaiseError("decoding image %q: %v", fileName, err)
				return 0
			}
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"set_ds1": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			mapName := state.CheckString(2)
			tileTable := state.CheckTable(3)
			tiles := make([]string, 0, tileTable.Len())
			for index := 1; index <= tileTable.Len(); index++ {
				value, ok := tileTable.RawGetInt(index).(lua.LString)
				if !ok || value == "" {
					state.RaiseError("DS1 tile %d must be a non-empty path", index)
					return 0
				}
				tiles = append(tiles, string(value))
			}
			if len(tiles) == 0 {
				state.RaiseError("DS1 rendering requires at least one DT1 tile path")
				return 0
			}
			palette := state.CheckString(4)
			decoded, err := node.cache.loadDS1(node.assets, mapName, tiles, palette)
			if err != nil {
				state.RaiseError("rendering DS1 %q: %v", mapName, err)
				return 0
			}
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating DS1 render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(decoded.Bounds().Dx()))
			state.Push(lua.LNumber(decoded.Bounds().Dy()))
			return 2
		},
		"set_ds1_chunk": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			mapName := state.CheckString(2)
			tileTable := state.CheckTable(3)
			tiles := make([]string, 0, tileTable.Len())
			for index := 1; index <= tileTable.Len(); index++ {
				value, ok := tileTable.RawGetInt(index).(lua.LString)
				if !ok || value == "" {
					state.RaiseError("DS1 tile %d must be a non-empty path", index)
					return 0
				}
				tiles = append(tiles, string(value))
			}
			chunkIndex := state.CheckInt(5)
			chunkSize := state.OptInt(6, maprender.DefaultChunkSize)
			chunks, err := node.cache.loadDS1Chunks(node.assets, mapName, tiles, state.CheckString(4), chunkSize)
			if err != nil {
				state.RaiseError("rendering DS1 chunks %q: %v", mapName, err)
				return 0
			}
			if chunkIndex < 0 || chunkIndex >= len(chunks.Chunks) {
				state.RaiseError("DS1 chunk %d out of range [0,%d)", chunkIndex, len(chunks.Chunks))
				return 0
			}
			chunk := chunks.Chunks[chunkIndex]
			if err := node.setImage(chunk.Pixels); err != nil {
				state.RaiseError("updating DS1 chunk node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(chunk.X))
			state.Push(lua.LNumber(chunk.Y))
			state.Push(lua.LNumber(chunk.Pixels.Bounds().Dx()))
			state.Push(lua.LNumber(chunk.Pixels.Bounds().Dy()))
			state.Push(lua.LNumber(chunks.Width))
			state.Push(lua.LNumber(chunks.Height))
			state.Push(lua.LNumber(len(chunks.Chunks)))
			return 7
		},
		"set_world_chunk": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			world := checkWorldMap(state, 2)
			chunkIndex := state.CheckInt(4)
			palette := state.CheckString(3)
			chunkSize := state.OptInt(5, maprender.DefaultChunkSize)
			chunks, err := node.cache.loadWorldChunks(node.assets, world, palette, chunkSize)
			if err != nil {
				state.RaiseError("rendering world chunks: %v", err)
				return 0
			}
			if chunkIndex < 0 || chunkIndex >= len(chunks.Chunks) {
				state.RaiseError("world chunk %d out of range [0,%d)", chunkIndex, len(chunks.Chunks))
				return 0
			}
			chunk, err := node.cache.loadWorldChunk(node.assets, world, palette, chunkSize, chunkIndex)
			if err != nil {
				state.RaiseError("rendering world chunk %d: %v", chunkIndex, err)
				return 0
			}
			if err := node.setImage(chunk.Pixels); err != nil {
				state.RaiseError("updating world chunk node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(chunk.X))
			state.Push(lua.LNumber(chunk.Y))
			state.Push(lua.LNumber(chunk.Pixels.Bounds().Dx()))
			state.Push(lua.LNumber(chunk.Pixels.Bounds().Dy()))
			state.Push(lua.LNumber(chunks.Width))
			state.Push(lua.LNumber(chunks.Height))
			state.Push(lua.LNumber(len(chunks.Chunks)))
			return 7
		},
		"set_world_tile": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			world := checkWorldMap(state, 2)
			palette := state.CheckString(3)
			drawIndex := state.CheckInt(4)
			set, err := node.cache.loadWorldTiles(node.assets, world, palette)
			if err != nil {
				state.RaiseError("indexing shared world tiles: %v", err)
				return 0
			}
			if drawIndex < 0 || drawIndex >= len(set.Draws) {
				state.RaiseError("world tile draw %d out of range [0,%d)", drawIndex, len(set.Draws))
				return 0
			}
			draw := set.Draws[drawIndex]
			if draw.Graphic < 0 || draw.Graphic >= len(set.Graphics) {
				state.RaiseError("world tile graphic %d out of range [0,%d)", draw.Graphic, len(set.Graphics))
				return 0
			}
			graphic := set.Graphics[draw.Graphic]
			pixels, err := node.cache.loadWorldTileGraphic(node.assets, world, palette, draw.Graphic)
			if err != nil {
				state.RaiseError("decoding shared world tile graphic: %v", err)
				return 0
			}
			key := fmt.Sprintf("world-dt1:%p:%s:%s:%d", world, palette, graphic.Path, graphic.Index)
			if err := node.setSharedImage(key, pixels); err != nil {
				state.RaiseError("updating shared world tile node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(draw.Bounds.Min.X))
			state.Push(lua.LNumber(draw.Bounds.Min.Y))
			state.Push(lua.LNumber(draw.Bounds.Dx()))
			state.Push(lua.LNumber(draw.Bounds.Dy()))
			state.Push(lua.LNumber(set.Width))
			state.Push(lua.LNumber(set.Height))
			state.Push(lua.LNumber(len(set.Draws)))
			return 7
		},
		"set_world_collision_region": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			world := checkWorldMap(state, 2)
			region := image.Rect(state.CheckInt(3), state.CheckInt(4), state.CheckInt(5), state.CheckInt(6))
			decoded, bounds := maprender.CollisionRegionImage(world, region)
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating world collision diagnostic: %v", err)
				return 0
			}
			state.Push(lua.LNumber(bounds.Min.X))
			state.Push(lua.LNumber(bounds.Min.Y))
			state.Push(lua.LNumber(bounds.Dx()))
			state.Push(lua.LNumber(bounds.Dy()))
			return 4
		},
		"set_world_tile_region": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			world := checkWorldMap(state, 2)
			region := image.Rect(state.CheckInt(3), state.CheckInt(4), state.CheckInt(5), state.CheckInt(6))
			decoded, bounds := maprender.TileRegionImage(world, region)
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating world tile diagnostic: %v", err)
				return 0
			}
			state.Push(lua.LNumber(bounds.Min.X))
			state.Push(lua.LNumber(bounds.Min.Y))
			state.Push(lua.LNumber(bounds.Dx()))
			state.Push(lua.LNumber(bounds.Dy()))
			return 4
		},
		"set_ds1_collision": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			mapName := state.CheckString(2)
			tileTable := state.CheckTable(3)
			tiles := make([]string, 0, tileTable.Len())
			for index := 1; index <= tileTable.Len(); index++ {
				value, ok := tileTable.RawGetInt(index).(lua.LString)
				if !ok || value == "" {
					state.RaiseError("DS1 tile %d must be a non-empty path", index)
					return 0
				}
				tiles = append(tiles, string(value))
			}
			decoded, err := node.cache.loadDS1Collision(node.assets, mapName, tiles)
			if err != nil {
				state.RaiseError("rendering DS1 collision %q: %v", mapName, err)
				return 0
			}
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating DS1 collision node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(decoded.Bounds().Dx()))
			state.Push(lua.LNumber(decoded.Bounds().Dy()))
			return 2
		},
		"set_dt1": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			prepared, err := node.cache.loadDT1Tile(node.assets, state.CheckString(2), state.OptString(3, ""), state.OptInt(4, 0), state.OptString(5, "composite"))
			if err != nil {
				state.RaiseError("rendering DT1 tile: %v", err)
				return 0
			}
			if err := node.setImage(prepared.image); err != nil {
				state.RaiseError("updating DT1 render node: %v", err)
				return 0
			}
			metadata := state.NewTable()
			metadata.RawSetString("total", lua.LNumber(prepared.total))
			metadata.RawSetString("type", lua.LNumber(prepared.tile.Type))
			metadata.RawSetString("style", lua.LNumber(prepared.tile.Style))
			metadata.RawSetString("sequence", lua.LNumber(prepared.tile.Sequence))
			// These are the names used by the map-selection algorithm and by
			// modern reference implementations. Keep the historical aliases
			// above because existing mod scripts may still use them.
			metadata.RawSetString("orientation", lua.LNumber(prepared.tile.Type))
			metadata.RawSetString("main_index", lua.LNumber(prepared.tile.Style))
			metadata.RawSetString("sub_index", lua.LNumber(prepared.tile.Sequence))
			metadata.RawSetString("direction", lua.LNumber(prepared.tile.Direction))
			metadata.RawSetString("rarity", lua.LNumber(prepared.tile.RarityFrameIndex))
			metadata.RawSetString("blocks", lua.LNumber(len(prepared.tile.Blocks)))
			metadata.RawSetString("tile_width", lua.LNumber(prepared.tile.Width))
			metadata.RawSetString("tile_height", lua.LNumber(prepared.tile.Height))
			metadata.RawSetString("roof_height", lua.LNumber(prepared.tile.RoofHeight))
			state.Push(lua.LNumber(prepared.image.Bounds().Dx()))
			state.Push(lua.LNumber(prepared.image.Bounds().Dy()))
			state.Push(metadata)
			return 3
		},
		"set_dc6": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			paletteName := state.OptString(3, "")
			direction := state.OptInt(4, 0)
			frameIndex := state.OptInt(5, 0)
			prepared, err := node.cache.loadDC6Frame(node.assets, fileName, paletteName, direction, frameIndex)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if err := node.setImage(prepared.image); err != nil {
				state.RaiseError("updating render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(prepared.frame.Width))
			state.Push(lua.LNumber(prepared.frame.Height))
			state.Push(lua.LNumber(prepared.frame.OffsetX))
			state.Push(lua.LNumber(dc6FrameTop(prepared.frame)))
			return 4
		},
		"set_dc6_combined": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			paletteName := state.OptString(3, "")
			direction := state.OptInt(4, 0)
			page := state.OptInt(5, 0)
			pages, err := node.cache.loadDC6Combined(node.assets, fileName, paletteName, direction)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if page < 0 || page >= len(pages) {
				state.ArgError(5, "combined page is out of range")
				return 0
			}
			decoded := pages[page]
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(decoded.Bounds().Dx()))
			state.Push(lua.LNumber(decoded.Bounds().Dy()))
			state.Push(lua.LNumber(len(pages)))
			return 3
		},
		"set_dc6_strip": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			paletteName := state.OptString(3, "")
			direction := state.OptInt(4, 0)
			prepared, err := node.cache.loadDC6Direction(node.assets, fileName, paletteName, direction)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			decoded, err := horizontalDC6Strip(prepared.asset, 0)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if err := node.setImage(decoded); err != nil {
				state.RaiseError("updating render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(decoded.Bounds().Dx()))
			state.Push(lua.LNumber(decoded.Bounds().Dy()))
			return 2
		},
		"set_dc6_animation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName := state.CheckString(2)
			paletteName := state.OptString(3, "")
			direction := state.OptInt(4, 0)
			framesPerSecond := float64(state.OptNumber(5, 15))
			loop := state.OptString(6, "loop")
			anchorMode := state.OptString(7, "offsets")
			if framesPerSecond <= 0 {
				state.ArgError(5, "frames per second must be positive")
				return 0
			}
			if loop != "loop" && loop != "once" && loop != "ping-pong" {
				state.ArgError(6, "loop mode must be loop, once, or ping-pong")
				return 0
			}
			if anchorMode != "offsets" && anchorMode != "first-frame" {
				state.ArgError(7, "anchor mode must be offsets or first-frame")
				return 0
			}
			var sharedBounds []image.Rectangle
			if state.GetTop() >= 11 {
				bounds := image.Rect(state.CheckInt(8), state.CheckInt(9), state.CheckInt(10), state.CheckInt(11))
				if bounds.Empty() {
					state.ArgError(8, "shared animation bounds must not be empty")
					return 0
				}
				sharedBounds = append(sharedBounds, bounds)
			}
			prepared, err := node.cache.loadDC6Animation(node.assets, fileName, paletteName, direction, anchorMode, sharedBounds...)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			if err := node.setAnimation(prepared.frames, time.Duration(float64(time.Second)/framesPerSecond), loop); err != nil {
				state.RaiseError("updating render animation: %v", err)
				return 0
			}
			state.Push(lua.LNumber(len(prepared.frames)))
			state.Push(lua.LNumber(prepared.bounds.Dx()))
			state.Push(lua.LNumber(prepared.bounds.Dy()))
			state.Push(lua.LNumber(prepared.bounds.Min.X))
			state.Push(lua.LNumber(prepared.bounds.Min.Y))
			return 5
		},
		"set_dcc": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName, paletteName := state.CheckString(2), state.OptString(3, "")
			direction, frameIndex := state.OptInt(4, 0), state.OptInt(5, 0)
			asset, err := node.cache.loadDCCDirection(node.assets, fileName, paletteName, direction)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			directionFrames := asset.direction.Frames()
			frames := make([]image.Image, len(directionFrames))
			for index, frame := range directionFrames {
				frames[index] = rgbaDCCFrame(frame, &asset.palette)
			}
			if frameIndex < 0 || frameIndex >= len(frames) {
				state.ArgError(5, "frame is out of range")
				return 0
			}
			if err := node.setImage(frames[frameIndex]); err != nil {
				state.RaiseError("updating DCC render node: %v", err)
			}
			bounds := asset.direction.Frames()[frameIndex].Bounds()
			state.Push(lua.LNumber(bounds.Dx()))
			state.Push(lua.LNumber(bounds.Dy()))
			state.Push(lua.LNumber(bounds.Min.X))
			state.Push(lua.LNumber(bounds.Min.Y))
			return 4
		},
		"set_dcc_animation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			fileName, paletteName := state.CheckString(2), state.OptString(3, "")
			direction := state.OptInt(4, 0)
			framesPerSecond := float64(state.OptNumber(5, 25))
			loop := state.OptString(6, "loop")
			if framesPerSecond <= 0 {
				state.ArgError(5, "frames per second must be positive")
				return 0
			}
			if loop != "loop" && loop != "once" && loop != "ping-pong" {
				state.ArgError(6, "loop mode must be loop, once, or ping-pong")
				return 0
			}
			asset, err := node.cache.loadDCCDirection(node.assets, fileName, paletteName, direction)
			if err != nil {
				state.RaiseError("%v", err)
				return 0
			}
			directionFrames := asset.direction.Frames()
			frames := make([]image.Image, len(directionFrames))
			for index, frame := range directionFrames {
				frames[index] = rgbaDCCFrame(frame, &asset.palette)
			}
			if err := node.setAnimation(frames, time.Duration(float64(time.Second)/framesPerSecond), loop); err != nil {
				state.RaiseError("updating DCC animation: %v", err)
				return 0
			}
			state.Push(lua.LNumber(len(frames)))
			return 1
		},
		"set_cof": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			cofName, palette := state.CheckString(2), state.OptString(3, "")
			direction, frameIndex := state.OptInt(4, 0), state.OptInt(5, 0)
			frames, asset, canvas, err := node.cofFrames(cofName, palette, direction, luaComponentPaths(state, 6))
			if err != nil {
				state.RaiseError("composing COF: %v", err)
				return 0
			}
			if frameIndex < 0 || frameIndex >= len(frames) {
				state.ArgError(5, "frame is out of range")
				return 0
			}
			if err := node.setImage(frames[frameIndex]); err != nil {
				state.RaiseError("updating COF render node: %v", err)
				return 0
			}
			width, height := frames[frameIndex].Bounds().Dx(), frames[frameIndex].Bounds().Dy()
			if err := node.composer.Update(node.id, func(current *render.Node) {
				current.OriginX = float64(-canvas.Min.X) / float64(max(width, 1))
				current.OriginY = float64(-canvas.Min.Y) / float64(max(height, 1))
			}); err != nil {
				state.RaiseError("updating COF ground origin: %v", err)
				return 0
			}
			state.Push(lua.LNumber(frames[frameIndex].Bounds().Dx()))
			state.Push(lua.LNumber(frames[frameIndex].Bounds().Dy()))
			state.Push(lua.LNumber(asset.AnimationFrames[frameIndex]))
			return 3
		},
		"set_cof_animation": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			cofName, palette := state.CheckString(2), state.OptString(3, "")
			direction := state.OptInt(4, 0)
			paths := luaComponentPaths(state, 5)
			loop := state.OptString(6, "loop")
			prepared, err := node.cachedCOFAnimation(cofName, palette, direction, paths)
			if err != nil {
				state.RaiseError("composing COF animation: %v", err)
				return 0
			}
			// Player COFs commonly leave speed at zero because AnimData.d2 owns the
			// runtime rate. Callers that already resolved that authority may supply
			// the classic 1/256-rate value without teaching the renderer gameplay.
			rate := state.OptInt(7, prepared.asset.Speed)
			if rate <= 0 {
				state.RaiseError("COF speed must be positive")
				return 0
			}
			duration := time.Duration(float64(time.Second) * 256 / (float64(rate) * 25))
			seek := time.Duration(float64(time.Second) * float64(state.OptNumber(8, 0)))
			if err := node.setAnimationKeyed(prepared.frames, prepared.keys, duration, loop, seek); err != nil {
				state.RaiseError("updating COF animation: %v", err)
				return 0
			}
			// DCC bounds are authored around logical character origin (0,0), which
			// is the feet/ground contact. The composed RGBA canvas may extend in
			// every direction for limbs and shadows, so its visual center is not a
			// valid world pivot.
			width, height := prepared.frames[0].Bounds().Dx(), prepared.frames[0].Bounds().Dy()
			if err := node.composer.Update(node.id, func(current *render.Node) {
				current.OriginX = float64(prepared.origin.X) / float64(max(width, 1))
				current.OriginY = float64(prepared.origin.Y) / float64(max(height, 1))
			}); err != nil {
				state.RaiseError("updating COF ground origin: %v", err)
				return 0
			}
			events := state.NewTable()
			for _, event := range prepared.asset.AnimationFrames {
				events.Append(lua.LNumber(event))
			}
			state.Push(lua.LNumber(len(prepared.frames)))
			state.Push(events)
			return 2
		},
		"set_text": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if node.assets == nil {
				state.RaiseError("render asset filesystem is unavailable")
				return 0
			}
			tableName, sheetName := state.CheckString(2), state.CheckString(3)
			paletteName, text := state.OptString(4, ""), state.CheckString(5)
			red, green, blue, alpha := 255, 255, 255, 255
			transform := ""
			maxWidth, align := 0, "left"
			if state.GetTop() >= 6 && state.Get(6) != lua.LNil {
				options := state.CheckTable(6)
				integer := func(name string, fallback int) int {
					value := options.RawGetString(name)
					if value == lua.LNil {
						return fallback
					}
					return int(lua.LVAsNumber(value))
				}
				red, green, blue, alpha = integer("red", red), integer("green", green), integer("blue", blue), integer("alpha", alpha)
				if value := options.RawGetString("transform"); value != lua.LNil {
					transform = lua.LVAsString(value)
				}
				maxWidth = integer("max_width", 0)
				if value := options.RawGetString("align"); value != lua.LNil {
					align = lua.LVAsString(value)
				}
			}
			for _, channel := range []int{red, green, blue, alpha} {
				if channel < 0 || channel > 255 {
					state.ArgError(6, "text color channels must be between 0 and 255")
					return 0
				}
			}
			if maxWidth < 0 {
				state.ArgError(6, "max_width cannot be negative")
				return 0
			}
			font, err := node.cache.loadFont(node.assets, tableName, sheetName, paletteName, transform)
			if err != nil {
				state.RaiseError("loading bitmap font: %v", err)
				return 0
			}
			rendered, err := font.Render(text, color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: uint8(alpha)}, maxWidth, align)
			if err != nil {
				state.RaiseError("rendering bitmap text: %v", err)
				return 0
			}
			if err := node.setImage(rendered); err != nil {
				state.RaiseError("updating text render node: %v", err)
				return 0
			}
			state.Push(lua.LNumber(rendered.Bounds().Dx()))
			state.Push(lua.LNumber(rendered.Bounds().Dy()))
			return 2
		},
		"animation_pause": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.requireAnimation(); err != nil {
				state.RaiseError("pausing animation: %v", err)
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) { current.AnimationPaused = true }); err != nil {
				state.RaiseError("pausing animation: %v", err)
			}
			return 0
		},
		"animation_resume": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			if err := node.requireAnimation(); err != nil {
				state.RaiseError("resuming animation: %v", err)
				return 0
			}
			if err := node.composer.Update(node.id, func(current *render.Node) { current.AnimationPaused = false }); err != nil {
				state.RaiseError("resuming animation: %v", err)
			}
			return 0
		},
		"animation_seek": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			seconds := float64(state.CheckNumber(2))
			if seconds < 0 {
				state.ArgError(2, "seek position cannot be negative")
				return 0
			}
			if err := node.requireAnimation(); err != nil {
				state.RaiseError("seeking animation: %v", err)
				return 0
			}
			position := time.Duration(seconds * float64(time.Second))
			if err := node.composer.Update(node.id, func(current *render.Node) {
				current.AnimationSeek = position
				current.AnimationSeekRevision++
			}); err != nil {
				state.RaiseError("seeking animation: %v", err)
			}
			return 0
		},
		"fill_rect": func(state *lua.LState) int {
			node := checkRenderNode(state, 1)
			width, height := state.CheckInt(2), state.CheckInt(3)
			if width <= 0 || height <= 0 {
				state.ArgError(2, "positive width and height required")
				return 0
			}
			channel := func(index int, fallback int) uint8 {
				value := fallback
				if state.GetTop() >= index {
					value = state.CheckInt(index)
				}
				if value < 0 || value > 255 {
					state.ArgError(index, "color channel must be between 0 and 255")
				}
				return uint8(value)
			}
			fill := image.NewRGBA(image.Rect(0, 0, width, height))
			draw.Draw(fill, fill.Bounds(), &image.Uniform{C: color.RGBA{R: channel(4, 0), G: channel(5, 0), B: channel(6, 0), A: channel(7, 255)}}, image.Point{}, draw.Src)
			if err := node.setImage(fill); err != nil {
				state.RaiseError("updating render node: %v", err)
			}
			return 0
		},
		"destroy": func(state *lua.LState) int {
			if err := checkRenderNode(state, 1).release(); err != nil {
				state.RaiseError("destroying render node: %v", err)
			}
			return 0
		},
	}))
}

func checkRenderNode(state *lua.LState, index int) *ownedRenderNode {
	userData := state.CheckUserData(index)
	node, ok := userData.Value.(*ownedRenderNode)
	if !ok {
		state.ArgError(index, "engine.render/v1 node expected")
		return nil
	}
	return node
}

func parseLayer(name string) (render.Layer, error) {
	switch name {
	case "world":
		return render.LayerWorld, nil
	case "hud":
		return render.LayerHUD, nil
	case "modal":
		return render.LayerModal, nil
	case "cursor":
		return render.LayerCursor, nil
	case "debug":
		return render.LayerDebug, nil
	case "transition":
		return render.LayerTransition, nil
	default:
		return 0, fmt.Errorf("unknown render layer %q", name)
	}
}
