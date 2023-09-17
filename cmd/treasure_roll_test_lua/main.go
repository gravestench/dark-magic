package main

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/config_file"
	"github.com/gravestench/dark-magic/pkg/services/loaders/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/record_manager"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	// utility services
	rt.Add(&lua.Service{})
	rt.Add(&config_file.Service{RootDirectory: projectConfigDir})

	// d2 file loaders
	rt.Add(&tsvLoader.Service{})
	rt.Add(&mpqLoader.Service{})

	// high level d2 services
	rt.Add(&record_manager.Service{})

	rt.Run()
}
