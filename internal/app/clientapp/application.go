package clientapp

import (
	"context"
	"image"
	"sync"
	"time"

	"github.com/gravestench/akara"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	"github.com/gravestench/dark-magic/internal/inputstate"
	loadcore "github.com/gravestench/dark-magic/internal/loading"
	"github.com/gravestench/dark-magic/internal/localization"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/worldobjects"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	"github.com/gravestench/dark-magic/internal/modcache"
	"github.com/gravestench/dark-magic/internal/platform/desktop"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

// Options contains the choices made outside the client composition root.
type Options struct {
	Content               *content.FS
	Mods                  *modcache.ResolvedSet
	Packages              simulation.RuntimePackageSet
	AssetSetID            string
	ModCache              *modcache.Store
	Profile               Profile
	NewCapture            CaptureFactory
	CaptureDirectory      string
	CaptureScenes         string
	CaptureSettle         int
	StartScene            string
	StartOverlays         string
	FixtureCharacters     int
	PlayerProfilePath     string
	FixtureWorldLevel     int
	FixtureWorldSpawn     string
	FixturePointerMove    bool
	OutputPalette         string
	ViewportFit           string
	BorderlessFullscreen  bool
	DisableNativeAudio    bool
	PresentationProfileID string
	Logs                  *shell.LogBuffer
}

// Profile describes the optional developer profiler used by the client.
type Profile interface {
	CaptureSceneHeap(string) error
	SetDiagnostics(func() any)
}

// Capture observes completed frames and writes requested screenshots.
type Capture interface {
	Observe([]string, uint64, bool)
	Complete() bool
	Close() error
}

// Screenshotter describes the renderer operation required by a capture.
type Screenshotter interface {
	CaptureScreenshot(string) error
}

// CaptureFactory constructs optional developer captures outside production code.
type CaptureFactory func(string, string, int, Screenshotter) (Capture, error)

// application owns the dependencies and mutable state of one client process.
type application struct {
	options Options
	ctx     context.Context
	stop    context.CancelFunc

	// User configuration and presentation foundations.
	shellSettings *shell.Settings
	gameSettings  *preferences.Settings
	profile       content.PresentationProfile
	presentation  content.PresentationBootstrap

	// Renderer-independent presentation services.
	renderer         desktop.Renderer
	renderWindow     image.Point
	input            desktop.Input
	inputState       *inputstate.Store
	locale           *localization.Locale
	scripts          *modruntime.Runtime
	composer         *render.Composer
	mixer            *audio.Mixer
	navigator        *navigation.Manager
	scenes           *modruntime.Scenes
	renderCapability *modruntime.RenderCapability

	// Game data, authoritative simulation, and player state.
	records              *recordstore.Store
	questCatalog         *recovered.Catalog
	worldObjectResolver  *worldobjects.Resolver
	remoteMirrors        map[string]akara.Entity
	remoteMirrorKeys     map[string]string
	networkRosterLogKey  string
	privateProjectionKey string
	clientWorld          *clientWorld
	saves                *d2save.Store
	networkTrust         *networktrust.Store
	playerProfilePath    string
	entitySimulation     *gameecs.Engine
	clientSimulation     *gameecs.Engine
	ecsCapability        *modruntime.ECSCapability
	offlineSession       *gamesession.Session
	authoritativeState   *simulation.StateStore
	authoritativeRandom  *simulation.RandomStreams
	playerControl        *d2movement.MovementController
	movementSource       *d2movement.MovementSource
	movementCatalog      d2movement.Catalog
	transitionSeam       gametransition.Seam
	commandIntents       *gamesession.IntentController
	commandIntentSource  *gamesession.IntentSource
	commandSource        func(uint64) []simulation.Command

	// World state shared by local and remote presentation paths.
	worldMu           sync.RWMutex
	gameWorlds        map[int]*gameworld.Map
	gameWorldZones    map[int]*worldgen.Zone
	gameWorldSpawns   map[int][2]float64
	activeWorldLevel  int
	loading           *loadcore.Coordinator
	pointerAcceptance *pointerMovementAcceptance
	network           *networkController
	realm             *realmController

	// Runtime composition and developer tooling.
	components      *host.Manager
	packageRegistry *modruntime.PackageRegistry
	packageDigests  map[string]string
	configuredMods  simulation.RuntimePackageSet
	componentIDs    map[string]bool
	networkMounted  *modcache.MountedSet
	recomposeMu     sync.Mutex
	engineHost      *host.Host
	shellSession    *shell.Session
	console         desktop.Console

	// Frame-loop subscriptions and metrics.
	sceneErrors  chan error
	capture      Capture
	stopScene    func()
	stopOverlay  func()
	stopCapture  func()
	hostStopped  bool
	lastFrame    time.Time
	frameMetrics frameMetrics
}

// noCleanup is the safe default for optional subscription cleanup callbacks.
func noCleanup() {}
