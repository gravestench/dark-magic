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
	"github.com/gravestench/dark-magic/pkg/services/goscript"
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
	rt := servicemesh.New(projectName)

	rt.SetLogHandler(prettylog.NewHandler(nil))

	// utility services
	//rt.Add(&modalTui.Service{})
	rt.Add(&lua.Service{})
	rt.Add(&goscript.Service{})
	rt.Add(&cacheManager.Service{})
	rt.Add(&fileWatcher.Service{})
	rt.Add(&configFile.Service{RootDirectory: projectConfigDir})
	rt.Add(&webServer.Service{})
	rt.Add(&webRouter.Service{})
	rt.Add(&raylibRenderer.Service{})
	rt.Add(&input.Service{})
	rt.Add(&fontTableLoader.Service{})
	rt.Add(&dc6Loader.Service{})
	rt.Add(&dccLoader.Service{})
	rt.Add(&ds1Loader.Service{})
	rt.Add(&dt1Loader.Service{})
	rt.Add(&pl2Loader.Service{})
	rt.Add(&tblLoader.Service{})
	rt.Add(&tsvLoader.Service{})
	rt.Add(&wavLoader.Service{})
	rt.Add(&cofLoader.Service{})
	rt.Add(&mpqLoader.Service{})

	// high level services
	rt.Add(&assetLoader.Service{})
	rt.Add(&recordManager.Service{})
	rt.Add(&spriteManager.Service{})
	rt.Add(&locale.Service{})
	rt.Add(&hero.Service{})
	rt.Add(&mapGenerator.Service{})
	rt.Add(&guiManager.Service{})
	rt.Add(&modalGameUI.Service{})
	rt.Add(&backgroundMusic.Service{})

	// game ui screens
	rt.Add(&loading.Screen{})

	mainthread.Run(rt.Run)
}
