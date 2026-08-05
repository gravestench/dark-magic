package luaModLoader

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gravestench/servicemesh"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

//go:embed internal/mods
var internalMods embed.FS

var upgradeableBuiltinHashes = map[string][]string{
	"terminal/init.lua": {
		"8e6dc3dc2b04e62bf7b3e99301a727ccaa3cb6943c717f8f9df5fc44e6f1bcb7",
		"8b61623fc668f82c48e9ca0d87607c792ea53198b423b40c97d00ac9de72a0e2",
	},
}

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
	lua     luaManager.Dependency
	loader  fileLoader.Dependency
	watcher fileWatcher.Dependency
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	if err := s.lua.WithState(func(state *lua.LState) error {
		apiTable := state.GetGlobal("api").(*lua.LTable)
		state.SetField(apiTable, "mods", state.NewTable())
		state.SetField(apiTable, "services", luar.New(state, map[string]any{
			"modloader": luar.New(state, s),
		}))
		setupPackagePath(state, s.Config.ModDirectory)
		return nil
	}); err != nil {
		s.Logger().Error("initializing Lua mod API", "error", err)
		mesh.Shutdown()
		return
	}

	if err := s.ensureModDirectoryExists(); err != nil {
		s.Logger().Error("resolving mods directory", "error", err)
		mesh.Shutdown()
	}
	if err := s.installBuiltinMods(); err != nil {
		s.Logger().Error("installing built-in mods", "error", err)
		mesh.Shutdown()
		return
	}

	// 2) look for mods in mod folder
	mods, err := s.getModManifestPaths(s.Config.ModDirectory)
	if err != nil {
		s.Logger().Error("discovering mods", "error", err)
		mesh.Shutdown()
	}

	s.Logger().Info("init", "mod directory", s.Config.ModDirectory, "mods found", len(mods))

	go func() {
		for !s.lua.GlobalsExist("api.ui", "api.renderer", "api.tweens", "api.input", "api.records") {
			s.Logger().Info("waiting for api to become populated")
			time.Sleep(time.Second * 2)
		}

		// 3) load each enabled mod
		s.loadMods(mods)
	}()
}

// installBuiltinMods materializes embedded examples so Lua's standard require
// loader can execute them. Existing user files are never overwritten.
func (s *Service) installBuiltinMods() error {
	const root = "internal/mods"
	return fs.WalkDir(internalMods, root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." {
			return err
		}
		target := filepath.Join(s.Config.ModDirectory, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, err := os.Stat(target); err == nil {
			existing, readErr := os.ReadFile(target)
			if readErr != nil {
				return readErr
			}
			if !matchesBuiltinHash(existing, upgradeableBuiltinHashes[filepath.ToSlash(relative)]) {
				return nil
			}
			data, readErr := internalMods.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			return os.WriteFile(target, data, 0o644)
		} else if !os.IsNotExist(err) {
			return err
		}
		data, err := internalMods.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func matchesBuiltinHash(data []byte, hashes []string) bool {
	sum := sha256.Sum256(data)
	encoded := hex.EncodeToString(sum[:])
	for _, candidate := range hashes {
		if encoded == candidate {
			return true
		}
	}
	return false
}

func (s *Service) Name() string {
	return "Lua Mod Loader"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) OnServiceAdded(service servicemesh.Service) {

}
