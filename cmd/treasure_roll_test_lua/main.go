package main

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	// utility services
	rt.Add(&lua.Service{})
	rt.Add(&configFile.Service{RootDirectory: projectConfigDir})

	// d2 file loaders
	rt.Add(&tsvLoader.Service{})
	rt.Add(&mpqLoader.Service{})

	// high level d2 services
	rt.Add(&recordManager.Service{})

	rt.Run()
}
