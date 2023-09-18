package main

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
)

const (
	projectName      = "Dark Magic Runtime"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := runtime.New(projectName)

	rt.Add(&configFile.Service{RootDirectory: projectConfigDir})

	rt.Add(&tsvLoader.Service{})
	rt.Add(&wavLoader.Service{})
	rt.Add(&mpqLoader.Service{})

	rt.Add(&audioFileTestService{})

	rt.Run()
}
