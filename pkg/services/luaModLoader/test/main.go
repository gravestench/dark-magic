package main

import (
	"log/slog"

	"github.com/faiface/mainthread"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
	"github.com/gravestench/dark-magic/pkg/services/luaModLoader"
)

func main() {
	app := servicemesh.New("mod loader test")

	app.SetLogHandler(prettylog.NewHandler(&slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	app.Add(&configManager.Service{RootDirectory: "~/.config/dark-magic"})
	app.Add(&luaManager.Service{})
	app.Add(&fileLoader.Service{})
	app.Add(&luaModLoader.Service{})

	// renderer requires use of mainthread
	mainthread.Run(app.Run)
}
