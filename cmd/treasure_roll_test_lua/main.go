package main

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

const (
	projectName      = "Dark Magic"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	app := servicemesh.New(projectName)
	app.SetLogHandler(prettylog.NewHandler(nil))

	// utility services
	app.Add(&configManager.Service{RootDirectory: projectConfigDir})
	app.Add(&fileWatcher.Service{})

	// d2 file loaders
	//app.Add(&tsvLoader.Service{})
	//app.Add(&mpqLoader.Service{})

	// high level d2 services
	app.Add(&recordManager.Service{})
	app.Add(&luaManager.Service{})

	app.Run()
}
