package main

import (
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"

	"github.com/gravestench/dark-magic/pkg/services/mpq_loader"
	"github.com/gravestench/dark-magic/pkg/services/tsv_loader"
	"github.com/gravestench/dark-magic/pkg/services/wav_loader"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	rt.Add(&config_file.Service{RootDirectory: projectConfigDir})
	rt.Add(&tsv_loader.Service{})
	rt.Add(&wav_loader.Service{})
	rt.Add(&mpq_loader.Service{})

	rt.Add(&audioFileTestService{})

	rt.Run()
}
