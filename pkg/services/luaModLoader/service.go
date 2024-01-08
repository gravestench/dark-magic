package luaModLoader

import (
	"embed"
	"time"

	"github.com/gravestench/servicemesh"
	lua "github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

//go:embed internal/mods
var internalMods embed.FS

type recipe interface {
	servicemesh.Service
	servicemesh.HasLogger
	servicemesh.HasDependencies
	servicemesh.EventHandlerServiceAdded
	configManager.HasConfiguration
}

var _ recipe = &Service{}

type Service struct {
	common.Service
	*Config
	lua    luaManager.Dependency
	loader fileLoader.Dependency
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.lua.WithState(func(state *lua.LState) error {
		apiTable := state.GetGlobal("api").(*lua.LTable)
		state.SetField(apiTable, "mods", state.NewTable())

		if err := s.ensureModDirectoryExists(); err != nil {
			s.Logger().Error("resolving mods directory", "error", err)
			mesh.Shutdown()
		}
		s.Logger().Info("init", "mod directory", s.Config.ModDirectory)

		// 2) look for mods in mod folder
		mods, err := s.getModManifestPaths(s.Config.ModDirectory)
		if err != nil {
			s.Logger().Error("discovering mods", "error", err)
			mesh.Shutdown()
		}

		s.Logger().Info("init", "mods found", len(mods))

		setupPackagePath(state, s.Config.ModDirectory)

		go func() {
			for !s.lua.GlobalsExist("api.ui", "api.renderer", "api.tweens") {
				time.Sleep(time.Second)
			}

			// 3) load each enabled mod
			s.loadMods(mods)
		}()

		return nil
	})
}

func (s *Service) Name() string {
	return "Lua Mod Loader"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) OnServiceAdded(service servicemesh.Service) {

}
