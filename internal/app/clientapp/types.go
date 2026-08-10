// Package clientapp assembles the interactive Dark Magic client.
//
// A composition root is the place where independent parts are plugged
// together. It should describe the plugs; it should not hide game rules. This
// package therefore contains small assembly steps and no reusable domain logic.
package clientapp

import (
	"context"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/recovered"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/game/data/worldobjects"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
	loadcore "github.com/gravestench/dark-magic/internal/loading"
	"github.com/gravestench/dark-magic/internal/localization"
	"github.com/gravestench/dark-magic/internal/persistence"
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
	FixtureCharacters     int
	OutputPalette         string
	ViewportFit           string
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
	Observe([]string)
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

	renderer       *raylibrenderer.Service
	rendererConfig raylibrenderer.Config
	input          *raylibinput.Service
	inputState     *inputstate.Store
	locale         *localization.Locale
	scripts        *modruntime.Runtime
	composer       *render.Composer
	mixer          *audio.Mixer
	navigator      *navigation.Manager
	scenes         *modruntime.Scenes

	records             *recordstore.Store
	gameData            *gamedata.Catalog
	questCatalog        *recovered.Catalog
	worldObjectResolver *worldobjects.Resolver
	saves               *persistence.Store
	entitySimulation    *gameecs.Engine
	offlineSession      *gamesession.Session
	playerControl       *gamesession.MovementController
	itemAuthority       *gameitem.Authority
	itemControl         *gameitem.Controller
	itemSource          *gameitem.Source
	commandSource       func(uint64) []simulation.Command
	loading             *loadcore.Coordinator

	components   *host.Manager
	engineHost   *host.Host
	shellSession *shell.Session
	console      *raylibshell.Overlay

	sceneErrors chan error
	capture     Capture
	stopScene   func()
	stopOverlay func()
	stopCapture func()
	hostStopped bool
	lastFrame   time.Time
}

func noCleanup() {}
