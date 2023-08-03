package main

import (
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"

	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/mpq_loader"
	"github.com/gravestench/dark-magic/pkg/services/record_manager"
	"github.com/gravestench/dark-magic/pkg/services/tsv_loader"
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
	rt.Add(&tsv_loader.Service{})
	rt.Add(&mpq_loader.Service{})

	// high level d2 services
	rt.Add(&record_manager.Service{})

	rt.Run()
}
