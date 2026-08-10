package clientapp

import (
	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/gravestench/dark-magic/internal/shell/luashell"
	raylibshell "github.com/gravestench/dark-magic/internal/shell/raylib"
)

// startDeveloperConsole builds the little terminal drawn over the game.
// The policy is its permission slip: it may only use the Lua doors listed here.
func (app *application) startDeveloperConsole() error {
	policy := shell.Policy{
		Name:         "local-developer",
		Capabilities: app.scripts.ModuleNames(),
		Mutable:      true,
	}
	evaluator, err := luashell.NewForPolicy(app.scripts, policy)
	if err != nil {
		return wrap("create developer shell evaluator", err)
	}
	app.shellSession, err = shell.NewSession("client-local", "client", policy, evaluator)
	if err != nil {
		return wrap("create developer shell session", err)
	}
	app.shellSession.AttachLogs(app.options.Logs)
	app.shellSession.AttachSettings(app.shellSettings)
	app.console = raylibshell.New(app.shellSession, app.shellSettings)
	return wrap("load developer console font", app.console.LoadFont())
}
