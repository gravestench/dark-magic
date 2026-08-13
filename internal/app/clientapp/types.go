// Package clientapp assembles the interactive Dark Magic client.
//
// A composition root is the place where independent parts are plugged
// together. It should describe the plugs; it should not hide game rules. This
// package therefore contains small assembly steps and no reusable domain logic.
package clientapp

import (
	"context"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/store"
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
	raylibinput "github.com/gravestench/dark-magic/internal/platform/raylib/input"
	raylibrenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	raylibshell "github.com/gravestench/dark-magic/internal/shell/raylib"
)

// Options are the few choices made outside the client composition root.
// Everything else is created here so ownership stays obvious.
type Options struct {
	Content               *content.FS
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
	PresentationProfileID string
	Logs                  *shell.LogBuffer
}

// Profile is the tiny piece of the optional developer profiler the client uses.
// An interface keeps developer tooling outside the production dependency tree.
type Profile interface {
	CaptureSceneHeap(string) error
	SetDiagnostics(func() any)
}

// Capture observes completed frames and writes the requested screenshots.
type Capture interface {
	Observe([]string, uint64, bool)
	Complete() bool
	Close() error
}

// Screenshotter is what a capture needs from the renderer.
type Screenshotter interface {
	CaptureScreenshot(string) error
}

// CaptureFactory keeps optional developer capture construction outside the
// production client package while making frame observation an explicit plug.
type CaptureFactory func(string, string, int, Screenshotter) (Capture, error)

type application struct {
	options Options
	ctx     context.Context
	stop    context.CancelFunc

	shellSettings *shell.Settings
	gameSettings  *preferences.Settings
	profile       content.PresentationProfile
	presentation  content.PresentationBootstrap

	renderer         *raylibrenderer.Service
	rendererConfig   raylibrenderer.Config
	input            *raylibinput.Service
	inputState       *inputstate.Store
	locale           *localization.Locale
	scripts          *modruntime.Runtime
	composer         *render.Composer
	mixer            *audio.Mixer
	navigator        *navigation.Manager
	scenes           *modruntime.Scenes
	renderCapability *modruntime.RenderCapability

	records             *recordstore.Store
	questCatalog        *recovered.Catalog
	worldObjectResolver *worldobjects.Resolver
	saves               *d2save.Store
	networkTrust        *networktrust.Store
	playerProfilePath   string
	entitySimulation    *gameecs.Engine
	offlineSession      *gamesession.Session
	authoritativeState  *simulation.StateStore
	authoritativeRandom *simulation.RandomStreams
	playerControl       *d2movement.MovementController
	movementSource      *d2movement.MovementSource
	transitionSeam      gametransition.Seam
	commandIntents      *gamesession.IntentController
	commandIntentSource *gamesession.IntentSource
	commandSource       func(uint64) []simulation.Command
	worldMu             sync.RWMutex
	gameWorlds          map[int]*gameworld.Map
	gameWorldZones      map[int]*worldgen.Zone
	gameWorldSpawns     map[int][2]float64
	activeWorldLevel    int
	loading             *loadcore.Coordinator
	pointerAcceptance   *pointerMovementAcceptance
	network             *networkController

	components   *host.Manager
	engineHost   *host.Host
	shellSession *shell.Session
	console      *raylibshell.Overlay

	sceneErrors  chan error
	capture      Capture
	stopScene    func()
	stopOverlay  func()
	stopCapture  func()
	hostStopped  bool
	lastFrame    time.Time
	frameMetrics frameMetrics
}

func noCleanup() {}
