package headlessshell

import (
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/gravestench/dark-magic/internal/shell/luashell"
)

// buildSession connects the Lua evaluator, captured logs, and persisted settings.
func buildSession(
	target string,
	policy shell.Policy,
	runtime *modruntime.Runtime,
	settings *shell.Settings,
	logs *shell.LogBuffer,
) (*shell.Session, error) {
	evaluator, err := luashell.NewForPolicy(runtime, policy)
	if err != nil {
		return nil, err
	}

	session, err := shell.NewSession(target+"-local", target, policy, evaluator)
	if err != nil {
		return nil, err
	}

	// The terminal reads these attachments while rendering its auxiliary views.
	session.AttachLogs(logs)
	session.AttachSettings(settings)

	return session, nil
}
