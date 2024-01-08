package main

import (
	"github.com/faiface/mainthread"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/backgroundMusic"
	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/cofLoader"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
	"github.com/gravestench/dark-magic/pkg/services/fontTableLoader"
	"github.com/gravestench/dark-magic/pkg/services/guiManager"
	"github.com/gravestench/dark-magic/pkg/services/hero"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/mapGenerator"
	"github.com/gravestench/dark-magic/pkg/services/modalGameUI"
	"github.com/gravestench/dark-magic/pkg/services/modalGameUI/ui/loading"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
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

	// utility services
	//rt.Add(&modalTui.Service{})
	//app.Add(&goscript.Service{}) // WIP
	app.Add(&lua.Service{})
	app.Add(&cacheManager.Service{})
	app.Add(&fileWatcher.Service{})
	app.Add(&configFile.Service{RootDirectory: projectConfigDir})
	app.Add(&webServer.Service{})
	app.Add(&webRouter.Service{})

	// file/record loaders
	app.Add(&fontTableLoader.Service{})
	app.Add(&dc6Loader.Service{})
	app.Add(&dccLoader.Service{})
	app.Add(&ds1Loader.Service{})
	app.Add(&dt1Loader.Service{})
	app.Add(&pl2Loader.Service{})
	app.Add(&tblLoader.Service{})
	app.Add(&tsvLoader.Service{})
	app.Add(&wavLoader.Service{})
	app.Add(&cofLoader.Service{})
	app.Add(&mpqLoader.Service{})

	// these all use the loaders and records
	app.Add(&assetLoader.Service{})
	app.Add(&recordManager.Service{})
	app.Add(&spriteManager.Service{})
	app.Add(&locale.Service{})
	app.Add(&mapGenerator.Service{})
	app.Add(&hero.Service{})

	// rendering-dependant services
	app.Add(&raylibRenderer.Service{})
	app.Add(&input.Service{})           // rendering backend also handles input
	app.Add(&backgroundMusic.Service{}) // rendering backend also handles audio

	app.Add(&guiManager.Service{})
	app.Add(&modalGameUI.Service{})
	app.Add(&loading.Screen{})
	//app.Add(&trademark.Screen{})

	// renderer requires use of mainthread
	mainthread.Run(app.Run)
}
