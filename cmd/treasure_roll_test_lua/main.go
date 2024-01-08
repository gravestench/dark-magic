package main

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
)

const (
	projectName      = "Dark Magic"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	app := servicemesh.New(projectName)

	// utility services
	app.Add(&lua.Service{})
	app.Add(&configFile.Service{RootDirectory: projectConfigDir})

	// d2 file loaders
	app.Add(&tsvLoader.Service{})
	app.Add(&mpqLoader.Service{})

	// high level d2 services
	app.Add(&recordManager.Service{})

	app.Run()
}
