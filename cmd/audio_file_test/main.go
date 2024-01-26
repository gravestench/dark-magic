package main

import (
	"os"

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
	app := servicemesh.New(projectName)
	app.SetLogHandler(prettylog.NewHandler(nil, os.Stdout))

	app.Add(&configFile.Service{RootDirectory: projectConfigDir})
	app.Add(&fileWatcher.Service{})

	app.Add(&tsvLoader.Service{})
	app.Add(&wavLoader.Service{})
	app.Add(&mpqLoader.Service{})

	app.Add(&audioFileTestService{})

	app.Run()
}
