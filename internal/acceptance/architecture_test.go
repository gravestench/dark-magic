package acceptance

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestRetiredPublicPackagesCannotReturn(t *testing.T) {
	root := repositoryRoot(t)
	forbidden := map[string]struct{}{
		"github.com/gravestench/dark-magic/pkg/paths":                 {},
		"github.com/gravestench/dark-magic/pkg/prettylog":             {},
		"github.com/gravestench/dark-magic/pkg/cache":                 {},
		"github.com/gravestench/dark-magic/pkg/easing":                {},
		"github.com/gravestench/dark-magic/pkg/scene":                 {},
		"github.com/gravestench/dark-magic/pkg/assetdecode":           {},
		"github.com/gravestench/dark-magic/pkg/assetcatalog":          {},
		"github.com/gravestench/dark-magic/pkg/assetinspect":          {},
		"github.com/gravestench/dark-magic/pkg/loot":                  {},
		"github.com/gravestench/dark-magic/pkg/models":                {},
		"github.com/gravestench/dark-magic/internal/service_template": {},
		"github.com/gravestench/dark-magic/internal/recordstore":      {},
		"github.com/gravestench/dark-magic/internal/gamedata":         {},
		"github.com/gravestench/dark-magic/internal/inputcore":        {},
		"github.com/gravestench/dark-magic/internal/loadcore":         {},
		"github.com/gravestench/dark-magic/internal/localecore":       {},
		"github.com/gravestench/dark-magic/internal/savecore":         {},
		"github.com/gravestench/dark-magic/internal/audiocore":        {},
		"github.com/gravestench/dark-magic/internal/rendercore":       {},
		"github.com/gravestench/dark-magic/internal/videocore":        {},
		"github.com/gravestench/dark-magic/internal/raylib/common":    {},
		"github.com/gravestench/dark-magic/internal/raylib/input":     {},
		"github.com/gravestench/dark-magic/internal/raylib/renderer":  {},
		"github.com/gravestench/dark-magic/internal/raylib/world":     {},
		"github.com/gravestench/dark-magic/internal/host":             {},
		"github.com/gravestench/dark-magic/internal/filewatch":        {},
		"github.com/gravestench/dark-magic/internal/hotreload":        {},
		"github.com/gravestench/dark-magic/internal/runtimeapi":       {},
		"github.com/gravestench/dark-magic/internal/modruntime":       {},
		"github.com/gravestench/dark-magic/internal/navigation":       {},
		"github.com/gravestench/dark-magic/internal/capture":          {},
		"github.com/gravestench/dark-magic/internal/profiling":        {},
	}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "dist") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			name, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if _, rejected := forbidden[name]; rejected {
				t.Errorf("%s imports retired public package %s", path, name)
			}
			if strings.Contains(name, "servicemesh") {
				t.Errorf("%s imports retired service-mesh package %s", path, name)
			}
			modelRoot := filepath.Join(root, "internal", "game", "data", "model")
			if pathWithin(path, modelRoot) && (name == "github.com/yuin/gopher-lua" || strings.HasPrefix(name, "github.com/gravestench/dark-magic/internal/")) {
				t.Errorf("%s couples typed game data to engine/runtime package %s", path, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDependencyDirection(t *testing.T) {
	root := repositoryRoot(t)
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		packagePath := filepath.ToSlash(filepath.Dir(relative))
		for _, imported := range file.Imports {
			name, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			const projectInternal = "github.com/gravestench/dark-magic/internal/"
			if !strings.HasPrefix(name, projectInternal) {
				continue
			}
			dependency := strings.TrimPrefix(name, "github.com/gravestench/dark-magic/")
			if !strings.HasPrefix(packagePath, "internal/dev") && strings.HasPrefix(dependency, "internal/dev/") {
				t.Errorf("%s imports developer-only package %s", relative, dependency)
			}
			if forbiddenLayerImport(packagePath, dependency) {
				t.Errorf("%s points outward from %s to %s", relative, packagePath, dependency)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func forbiddenLayerImport(packagePath, dependency string) bool {
	for _, root := range []string{"internal/cache", "internal/paths", "internal/logging", "internal/game/data/model"} {
		if packagePath == root || strings.HasPrefix(packagePath, root+"/") {
			return strings.HasPrefix(dependency, "internal/")
		}
	}
	if packagePath == "internal/content" || strings.HasPrefix(packagePath, "internal/content/") {
		return strings.HasPrefix(dependency, "internal/") && dependency != "internal/paths"
	}
	if strings.HasPrefix(packagePath, "internal/game/") {
		return hasAnyPrefix(dependency, "internal/app/", "internal/dev/", "internal/platform/", "internal/presentation/", "internal/runtime/")
	}
	if strings.HasPrefix(packagePath, "internal/presentation/") {
		return hasAnyPrefix(dependency, "internal/app/", "internal/dev/", "internal/platform/", "internal/runtime/")
	}
	if strings.HasPrefix(packagePath, "internal/platform/") {
		return hasAnyPrefix(dependency, "internal/app/", "internal/dev/", "internal/runtime/")
	}
	return false
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func TestCommandRemainsCompositionOnly(t *testing.T) {
	root := repositoryRoot(t)
	allowedFunctions := map[string]struct{}{
		"main": {}, "environmentDefault": {}, "parseLogLevel": {}, "run": {},
		"developmentCharacters": {}, "buildVersion": {}, "stopHost": {},
	}
	err := filepath.WalkDir(filepath.Join(root, "cmd"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if function.Recv != nil {
				t.Errorf("command contains method %s; move behavior under internal", function.Name.Name)
				continue
			}
			if _, allowed := allowedFunctions[function.Name.Name]; !allowed {
				t.Errorf("command contains unreviewed function %s; keep commands to composition and move behavior under internal", function.Name.Name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func pathWithin(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func TestRetiredDeveloperDirectoriesCannotReturn(t *testing.T) {
	root := repositoryRoot(t)
	for _, relative := range []string{"internal/tools", "internal/testapps"} {
		_, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative)))
		if err == nil {
			t.Errorf("retired developer directory exists: %s", relative)
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("inspect retired developer directory %s: %v", relative, err)
		}
	}
}

func TestLegacyRendererObjectAPIAndWorldAdapterCannotReturn(t *testing.T) {
	root := repositoryRoot(t)
	world := filepath.Join(root, "internal", "platform", "raylib", "world")
	files, err := filepath.Glob(filepath.Join(world, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("retired direct-renderable world adapter has Go files: %v", files)
	}

	renderer := filepath.Join(root, "internal", "platform", "raylib", "renderer")
	forbidden := map[string]bool{"Renderable": true, "NewRenderable": true, "ProvidesRenderables": true, "ProvidesTextures": true, "ManagesCameras": true}
	err = filepath.WalkDir(renderer, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(candidate ast.Node) bool {
			switch candidate := candidate.(type) {
			case *ast.TypeSpec:
				if forbidden[candidate.Name.Name] {
					t.Errorf("%s restores retired renderer type %s", path, candidate.Name.Name)
				}
			case *ast.FuncDecl:
				if forbidden[candidate.Name.Name] {
					t.Errorf("%s restores retired renderer constructor %s", path, candidate.Name.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoAccidentalPublicGoPackages(t *testing.T) {
	pkgRoot := filepath.Join(repositoryRoot(t), "pkg")
	if _, err := os.Stat(pkgRoot); errors.Is(err, fs.ErrNotExist) {
		return
	} else if err != nil {
		t.Fatal(err)
	}
	err := filepath.WalkDir(pkgRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			t.Errorf("public Go source requires an explicit compatibility commitment: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
