package main

import (
	"log/slog"

	"github.com/faiface/mainthread"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
	"github.com/gravestench/dark-magic/pkg/services/luaModLoader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
	"github.com/gravestench/dark-magic/pkg/services/tweens"
	"github.com/gravestench/dark-magic/pkg/services/ui"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
	"github.com/gravestench/dark-magic/pkg/services/webServer"
)

const (
	projectName      = "Dark Magic"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	app := servicemesh.New(projectName)

	app.SetLogHandler(prettylog.NewHandler(nil))
	app.SetLogLevel(slog.LevelDebug)

	// utility services
	//rt.Add(&modalTui.Service{})
	//app.Add(&goscript.Service{}) // WIP
	app.Add(&luaManager.Service{})
	app.Add(&cacheManager.Service{})
	app.Add(&fileLoader.Service{})
	app.Add(&fileWatcher.Service{})
	app.Add(&configManager.Service{RootDirectory: projectConfigDir})
	app.Add(&webServer.Service{})
	app.Add(&webRouter.Service{})
	app.Add(&tweens.Service{})

	// these all use the loaders and records
	app.Add(&assetLoader.Service{})
	app.Add(&recordManager.Service{})
	app.Add(&spriteManager.Service{})
	app.Add(&locale.Service{})
	//app.Add(&mapGenerator.Service{})
	//app.Add(&hero.Service{})

	// rendering-dependant services
	app.Add(&raylibRenderer.Service{})
	app.Add(&ui.Service{})
	app.Add(&input.Service{}) // rendering backend also handles input
	//app.Add(&backgroundMusic.Service{}) // rendering backend also handles audio
	//app.Add(&guiManager.Service{})
	//app.Add(&modalGameUI.Service{})
	//app.Add(&loading.Screen{})
	app.Add(&luaModLoader.Service{})

	// renderer requires use of mainthread
	mainthread.Run(app.Run)
}
