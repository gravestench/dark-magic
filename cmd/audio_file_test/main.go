package main

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/config_file"
	"github.com/gravestench/dark-magic/pkg/services/loaders/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/wavLoader"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	rt.Add(&config_file.Service{RootDirectory: projectConfigDir})

	rt.Add(&tsvLoader.Service{})
	rt.Add(&wavLoader.Service{})
	rt.Add(&mpqLoader.Service{})

	rt.Add(&audioFileTestService{})

	rt.Run()
}
