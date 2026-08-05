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
	"sort"
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
	discovered := make([]discoveredMod, 0, len(mods))
	for rootDir, filesystem := range mods {
		manifest, err := s.loadModManifest(rootDir, filesystem)
		if err != nil {
			s.Logger().Error("loading mod manifest", "error", err, "mod", rootDir)
			continue
		}
		discovered = append(discovered, discoveredMod{filesystem: filesystem, manifest: manifest})
	}
	ordered, err := orderMods(discovered)
	if err != nil {
		s.Logger().Error("ordering mods", "error", err)
		return
	}
	for _, mod := range ordered {
		if err := s.loadModManifestScripts(mod.manifest, mod.filesystem); err != nil {
			s.Logger().Error("loading mod", "error", err, "mod", mod.manifest.rootDir)
		}
	}
}

type discoveredMod struct {
	filesystem fs.FS
	manifest   *Manifest
}

func orderMods(mods []discoveredMod) ([]discoveredMod, error) {
	sort.Slice(mods, func(i, j int) bool { return mods[i].manifest.ID() < mods[j].manifest.ID() })
	byGlobal := make(map[string]int, len(mods))
	for idx, mod := range mods {
		if !mod.manifest.Enabled {
			continue
		}
		global := "api.mods." + mod.manifest.ApiKey()
		if _, exists := byGlobal[global]; exists {
			return nil, fmt.Errorf("duplicate enabled mod API global %q", global)
		}
		byGlobal[global] = idx
	}

	state := make([]uint8, len(mods))
	ordered := make([]discoveredMod, 0, len(mods))
	var visit func(int) error
	visit = func(idx int) error {
		if state[idx] == 2 {
			return nil
		}
		if !mods[idx].manifest.Enabled {
			state[idx] = 2
			ordered = append(ordered, mods[idx])
			return nil
		}
		if state[idx] == 1 {
			return fmt.Errorf("mod dependency cycle includes %s", mods[idx].manifest.ID())
		}
		state[idx] = 1
		for _, requirement := range mods[idx].manifest.Requires {
			if dependency, exists := byGlobal[requirement]; exists {
				if err := visit(dependency); err != nil {
					return err
				}
			} else if strings.HasPrefix(requirement, "api.mods.") {
				return fmt.Errorf("%s requires missing or disabled mod global %q", mods[idx].manifest.ID(), requirement)
			}
		}
		state[idx] = 2
		ordered = append(ordered, mods[idx])
		return nil
	}
	for idx := range mods {
		if err := visit(idx); err != nil {
			return nil, err
		}
	}
	return ordered, nil
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
	return s.loadModManifestScripts(manifest, mod)
}

func (s *Service) loadModManifestScripts(manifest *Manifest, mod fs.FS) (err error) {
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
	if err = manifest.Validate(); err != nil {
		return nil, fmt.Errorf("validating manifest: %w", err)
	}
	if enabled, configured := s.Config.EnabledMods[manifest.ID()]; configured {
		manifest.Enabled = enabled
	}

	return &manifest, nil
}

func (s *Service) loadModSources(manifest *Manifest) error {
	manifestID := manifest.ID()

	for srcGroupKey, sources := range manifest.Sources {
		for _, srcPath := range sources {
			srcPath = expandSourcePath(srcPath, manifest.rootDir)
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

func expandSourcePath(path, modRoot string) string {
	path = os.ExpandEnv(path)
	if mpqDirectory := os.Getenv("MPQ_DIRECTORY"); mpqDirectory != "" {
		path = strings.ReplaceAll(path, "{{MPQ_DIRECTORY}}", mpqDirectory)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(modRoot, path)
	}
	return filepath.Clean(path)
}

func (s *Service) runModScripts(manifest Manifest, mod fs.FS) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	requirementsMet := make(chan struct{}, 1)

	go func() {
		for !s.lua.GlobalsExist(manifest.Requires...) {
			time.Sleep(time.Second)
		}

		requirementsMet <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for required dependencies")
	case <-requirementsMet:
		if len(manifest.Requires) < 1 {
			s.Logger().Info("", "requirements met", true, "dir", manifest.rootDir)
		} else {
			s.Logger().Info("", "requirements met", true, "dir", manifest.rootDir, "requires", manifest.Requires)
		}
	}

	// exec a command in lua to import our mod as a table inside 'api.mods'
	// eg. 'api.mods.darkmagicterminal10'
	dirName := filepath.Base(manifest.rootDir)
	cmdRequire := fmt.Sprintf("api.mods[\"%s\"] = require(%q)", manifest.ApiKey(), dirName)
	// make a logger for the mod itself, assign to field "log" inside the
	// mod table (eg api.mods.darkmagicterminal10.log)
	logger := slog.New(prettylog.NewHandler(nil))
	logger = s.Logger().With("service", fmt.Sprintf("Lua Mod Loader->%s", manifest.Name))
	logger.Info("running mod scripts")
	cmdInit := fmt.Sprintf("api.mods[\"%s\"]:Init()", manifest.ApiKey())
	for {
		err := s.lua.WithState(func(state *lua.LState) error {
			if err := state.DoString(cmdRequire); err != nil {
				return fmt.Errorf("importing mod: %w", err)
			}

			candidate := state.GetGlobal("api")
			candidate = state.GetField(candidate, "mods")
			candidate = state.GetField(candidate, manifest.ApiKey())
			table, ok := candidate.(*lua.LTable)
			if !ok {
				return fmt.Errorf("api.mods.%s is not a table", manifest.ApiKey())
			}

			bindLoggerToLuaEnvironment(logger, state, table)
			return state.DoString(cmdInit)
		})
		if err == nil {
			break
		}
		logger.Error("initializing mod", "error", err)
		time.Sleep(time.Second)
	}

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
		logger.Info(arg0, args[1:]...)

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
