package main

import (
	"os"

	"github.com/faiface/mainthread"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	bootstrap_backend "github.com/gravestench/dark-magic/pkg/services/bootstrapBackend"
	bootstrap_frontend "github.com/gravestench/dark-magic/pkg/services/bootstrapFrontend"
	bootstrap_game_client "github.com/gravestench/dark-magic/pkg/services/bootstrapGameClient"
)

func main() {
	app := servicemesh.New("Dark Magic")

	app.SetLogHandler(prettylog.NewHandler(nil, os.Stdout))

	app.Add(&bootstrap_backend.Service{})
	app.Add(&bootstrap_frontend.Service{})
	app.Add(&bootstrap_game_client.Service{})

	mainthread.Run(app.Run) // renderer requires use of mainthread
}
