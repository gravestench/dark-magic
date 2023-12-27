package main

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
)

const (
	projectName      = "Dark Magic"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	rt := servicemesh.New(projectName)
	rt.SetLogHandler(prettylog.NewHandler(nil))

	rt.Add(&configFile.Service{RootDirectory: projectConfigDir})
	rt.Add(&fileWatcher.Service{})

	rt.Add(&tsvLoader.Service{})
	rt.Add(&wavLoader.Service{})
	rt.Add(&mpqLoader.Service{})

	rt.Add(&audioFileTestService{})

	rt.Run()
}
