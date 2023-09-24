package main

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/fontTableLoader"
	"github.com/gravestench/dark-magic/pkg/services/gplLoader"
	"github.com/gravestench/dark-magic/pkg/services/hero"
	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
	"github.com/gravestench/dark-magic/pkg/services/webServer"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	// utility services
	rt.Add(&lua.Service{})
	rt.Add(&webServer.Service{})
	rt.Add(&webRouter.Service{})
	rt.Add(&configFile.Service{RootDirectory: projectConfigDir})

	// d2 file loaders
	rt.Add(&fontTableLoader.Service{})
	rt.Add(&dc6Loader.Service{})
	rt.Add(&dccLoader.Service{})
	rt.Add(&ds1Loader.Service{})
	rt.Add(&dt1Loader.Service{})
	rt.Add(&gplLoader.Service{})
	rt.Add(&pl2Loader.Service{})
	rt.Add(&tblLoader.Service{})
	rt.Add(&tsvLoader.Service{})
	rt.Add(&wavLoader.Service{})
	rt.Add(&mpqLoader.Service{})

	// high level d2 services
	rt.Add(&assetLoader.Service{})
	rt.Add(&recordManager.Service{})
	rt.Add(&locale.Service{})
	rt.Add(&hero.Service{})

	rt.Run()
}
