package headlessshell

import (
	"context"
	"os"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

// buildRuntime registers the requested modules and starts a Lua runtime.
func buildRuntime(
	ctx context.Context,
	policy shell.Policy,
	modules []modruntime.Module,
) (*modruntime.Runtime, *shell.Settings, shell.Policy, error) {
	runtime := modruntime.New()
	for _, module := range modules {
		if err := registerModule(runtime, module, &policy); err != nil {
			return nil, nil, policy, err
		}
	}

	settings, err := loadSettings()
	if err != nil {
		return nil, nil, policy, err
	}
	if err := registerShellModule(runtime, settings, &policy); err != nil {
		return nil, nil, policy, err
	}
	if err := runtime.Start(ctx); err != nil {
		return nil, nil, policy, err
	}
	return runtime, settings, policy, nil
}

// registerModule exposes one runtime module and records its capability.
func registerModule(runtime *modruntime.Runtime, module modruntime.Module, policy *shell.Policy) error {
	if err := runtime.RegisterModule(module); err != nil {
		return err
	}
	policy.Capabilities = append(policy.Capabilities, module.Name)
	return nil
}

// registerShellModule exposes settings operations to the administration shell.
func registerShellModule(runtime *modruntime.Runtime, settings *shell.Settings, policy *shell.Policy) error {
	if err := runtime.RegisterModule(modruntime.ShellModule(settings)); err != nil {
		return err
	}
	policy.Capabilities = append(policy.Capabilities, "engine.shell/v1")
	return nil
}

// loadSettings resolves the optional host path before opening shell settings.
func loadSettings() (*shell.Settings, error) {
	path, err := darkpaths.ExpandHost(os.Getenv("DARK_MAGIC_SHELL_CONFIG"))
	if err != nil {
		return nil, err
	}
	return shell.NewSettings(path)
}
