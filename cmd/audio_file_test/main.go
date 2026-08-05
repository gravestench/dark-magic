package main

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
)

const (
	projectName      = "Dark Magic"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	app := servicemesh.New(projectName)
	app.SetLogHandler(prettylog.NewHandler(nil))

	app.Add(&configManager.Service{RootDirectory: projectConfigDir})
	app.Add(&fileWatcher.Service{})

	//app.Add(&tsvLoader.Service{})
	//app.Add(&wavLoader.Service{})
	//app.Add(&mpqLoader.Service{})

	app.Add(&audioFileTestService{})

	app.Run()
}
