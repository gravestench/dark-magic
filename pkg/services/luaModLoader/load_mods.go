package luaModLoader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
)

func (s *Service) ensureModDirectoryExists() error {
	info, err := os.Stat(s.Config.ModDirectory)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(s.Config.ModDirectory, 0o755); err != nil {
			return fmt.Errorf("making new mod directory: %v", err)
		}
		info, err = os.Stat(s.Config.ModDirectory)
		if err != nil {
			return fmt.Errorf("making new mod directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("checking file info: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("configured mod directory path is not a directory")
	}

	return nil
}

func (s *Service) loadMods(mods map[string]fs.FS) {
	for rootDir, mod := range mods {
		if err := s.loadMod(rootDir, mod); err != nil {
			s.Logger().Error("loading mod", "error", err, "mod", rootDir)
		}
	}
}

func (s *Service) getModManifestPaths(modDirPath string) (map[string]fs.FS, error) {
	modMap := make(map[string]fs.FS)
	modDirs, err := findManifestDirectories(modDirPath)
	if err != nil {
		return nil, fmt.Errorf("getting mod directories with manifest files: %v", err)
	}

	for _, dir := range modDirs {
		modMap[dir] = os.DirFS(dir)
	}

	return modMap, nil
}

// FindManifestDirectories takes a root directory path and finds all subdirectories
// containing a 'manifest.json' file. It returns a slice of fs.FS representing those directories.
func findManifestDirectories(root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Check if this directory contains a manifest.json
		manifestPath := filepath.Join(path, "manifest.json")
		if _, err := os.Stat(manifestPath); err == nil {
			dirs = append(dirs, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dirs, nil
}

func (s *Service) loadMod(rootDir string, mod fs.FS) (err error) {
	manifest, err := s.loadModManifest(rootDir, mod)
	if err != nil {
		return fmt.Errorf("loading manifest: %v", err)
	}

	s.Logger().Info("loading mod", "mod", manifest.ID(), "enabled", manifest.Enabled)

	if !manifest.Enabled {
		return
	}

	modRootSource := fileLoader.NewSource(manifest.rootDir)

	if err = s.loader.AddSourceToGroup(modRootSource, manifest.ID()); err != nil {
		return fmt.Errorf("adding mod root directory as source: %v", err)
	}

	if err = s.loadModSources(manifest); err != nil {
		return fmt.Errorf("loading mod sources: %v", err)
	}

	s.Logger().Debug("loaded mod sources from manifest", "mod", manifest.String())

	if err = s.runModScripts(*manifest, mod); err != nil {
		return fmt.Errorf("running mod scripts: %v", err)
	}

	s.Logger().Info("loaded mod", "name", manifest.String())

	return nil
}

func (s *Service) loadModManifest(rootDir string, mod fs.FS) (*Manifest, error) {
	const manifestFileName = "manifest.json"

	f, err := mod.Open(manifestFileName)
	if err != nil {
		return nil, fmt.Errorf("opening file: %v", err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading data: %v", err)
	}

	manifest := Manifest{
		rootDir: rootDir,
		Sources: make(map[string][]string),
	}

	if err = json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("unmarshalling data: %v", err)
	}

	return &manifest, nil
}

func (s *Service) loadModSources(manifest *Manifest) error {
	manifestID := manifest.ID()

	for srcGroupKey, sources := range manifest.Sources {
		for _, srcPath := range sources {
			src := fileLoader.NewSource(srcPath)
			if err := s.loader.AddSourceToGroup(src, srcGroupKey); err != nil {
				return fmt.Errorf("")
			}

			if err := s.loader.AddSourceToGroup(src, manifestID); err != nil {
				return fmt.Errorf("")
			}
		}
	}

	s.Logger().Info("loaded sources", "mod", manifest.Name, "version", manifest.Version)

	return nil
}

func (s *Service) runModScripts(manifest Manifest, mod fs.FS) error {
	// exec a command in lua to import our mod as a table inside 'api.mods'
	// eg. 'api.mods.darkmagicterminal10'
	dirName := filepath.Base(manifest.rootDir)

	cmdRequire := fmt.Sprintf("api.mods[\"%s\"] = require(%q)", manifest.ApiKey(), dirName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	requirementsMet := make(chan bool)
	defer close(requirementsMet)

	go func() {
		s.lua.WaitForGlobals(manifest.Requires...)

		requirementsMet <- true
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for required dependencies")
	case <-requirementsMet:
		s.Logger().Info("requirements met")
	}

	go func() {
		if err := s.lua.WithState(func(state *lua.LState) error {
			err := state.DoString(cmdRequire)
			if err != nil {
				return fmt.Errorf("importing mode: %+v", err)
			}

			// get the target table
			candidate := state.GetGlobal("api")
			candidate = state.GetField(candidate, "mods")
			candidate = state.GetField(candidate, manifest.ApiKey())

			table, ok := candidate.(*lua.LTable)
			if !ok {
				s.Logger().Warn("got non-table entry", "global", fmt.Sprintf("api.mods.%s", manifest.ApiKey()))
				return fmt.Errorf("not a table: %v", candidate)
			}

			// make a logger for the mod itself, assign to field "log" inside the
			// mod table (eg api.mods.darkmagicterminal10.log)
			logger := slog.New(prettylog.NewHandler(nil))
			logger = s.Logger().With("service", fmt.Sprintf("[Lua Mod] %s", manifest.Name))
			logger.Info("running mod scripts")
			bindLoggerToLuaEnvironment(logger, state, table)

			logger.Info("initializing")

			cmdInit := fmt.Sprintf("api.mods[\"%s\"]:Init()", manifest.ApiKey())

			for err = state.DoString(cmdInit); err != nil; {
				logger.Error("initializing", "error", err)
				time.Sleep(time.Second)
			}

			return nil
		}); err != nil {
			s.Logger().Error("running mod scripts", "error", err)
		}
	}()

	return nil
}

// this makes it so we can use 'require' and the lua module
// will look for lua scripts in the mod root dir
func setupPackagePath(L *lua.LState, modsDirectory string) {
	// Assuming mods are in a directory relative to the current directory
	path := L.GetGlobal("package").(*lua.LTable).RawGetString("path").String()
	path = fmt.Sprintf("%s;%s/?/init.lua", path, modsDirectory) // in root
	//path = fmt.Sprintf("%s;%s", path, modRootPath) // one level down
	L.GetGlobal("package").(*lua.LTable).RawSetString("path", lua.LString(path))
}

func bindLoggerToLuaEnvironment(logger *slog.Logger, state *lua.LState, table *lua.LTable) {
	fnPrint := state.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		var args []any

		for idx := 0; idx < numArgs; idx++ {
			args = append(args, fmt.Sprintf("%s", L.CheckAny(idx+1)))
		}

		arg0 := fmt.Sprintf("%s", args[0])

		if len(args) > 1 {
			logger.Info(arg0, args[1:]...)
		} else {
			logger.Info(arg0)
		}

		return 0
	})

	state.SetField(table, "log", fnPrint)
}

func luaExists(L *lua.LState, path string) bool {
	keys := strings.Split(path, ".")
	value := L.GetGlobal(keys[0])
	if value == lua.LNil {
		return false
	}

	for _, key := range keys[1:] {
		if tbl, ok := value.(*lua.LTable); ok {
			if L.GetField(tbl, key) == lua.LNil {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
