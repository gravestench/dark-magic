package main

import (
	"github.com/faiface/mainthread"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	bootstrap_backend "github.com/gravestench/dark-magic/pkg/services/bootstrapBackend"
	bootstrap_frontend "github.com/gravestench/dark-magic/pkg/services/bootstrapFrontend"
)

const (
	projectName = "Dark Magic"
)

func main() {
	app := servicemesh.New(projectName)

	app.SetLogHandler(prettylog.NewHandler(nil))
	
	app.Add(&bootstrap_backend.Service{})
	app.Add(&bootstrap_frontend.Service{})

	mainthread.Run(app.Run) // renderer requires use of mainthread
}
