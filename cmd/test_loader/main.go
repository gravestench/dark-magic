package main

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/config_file"
	"github.com/gravestench/dark-magic/pkg/services/loaders/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/fontTableLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/gplLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/wavLoader"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/record_manager"
	"github.com/gravestench/dark-magic/pkg/services/web_router"
	"github.com/gravestench/dark-magic/pkg/services/web_server"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	// utility services
	rt.Add(&lua.Service{})
	rt.Add(&web_server.Service{})
	rt.Add(&web_router.Service{})
	rt.Add(&config_file.Service{RootDirectory: projectConfigDir})

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
	rt.Add(&record_manager.Service{})

	rt.Run()
}
