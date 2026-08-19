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

// TestRetiredPublicPackagesCannotReturn prevents convenience imports from recreating packages whose
// ownership was deliberately moved inward. It also keeps passive game-data models independent of
// engine and Lua runtime behavior.
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

			engineRuntime := strings.HasPrefix(
				name,
				"github.com/gravestench/dark-magic/internal/",
			)
			if pathWithin(path, modelRoot) && (name == "github.com/yuin/gopher-lua" || engineRuntime) {
				t.Errorf("%s couples typed game data to engine/runtime package %s", path, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestDependencyDirection enforces the repository's inward dependency arrows by inspecting imports.
// Without this ratchet, a small convenience import can quietly make mechanisms depend on apps,
// developer tools, presentation, or runtime composition.
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

// TestGameDataHasNoGlobalD2Catalog prevents the host from hardcoding one eager
// list of every Diablo table again. The engine may decode a caller-selected
// schema; d2legacy owns which records form its game and which are required.
func TestGameDataHasNoGlobalD2Catalog(t *testing.T) {
	root := repositoryRoot(t)
	retired := filepath.Join(root, "internal", "game", "data", "catalog")
	if entries, err := os.ReadDir(retired); err == nil && len(entries) != 0 {
		t.Fatalf("retired global game-data catalog returned: %s", retired)
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	for _, relative := range []string{"internal/game/data/typed", "internal/game/data/store"} {
		searchRoot := filepath.Join(root, filepath.FromSlash(relative))

		err := filepath.WalkDir(searchRoot, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			for _, declaration := range file.Decls {
				typeDecl, ok := declaration.(*ast.GenDecl)
				if !ok || typeDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range typeDecl.Specs {
					if named, ok := spec.(*ast.TypeSpec); ok && (named.Name.Name == "Catalog" || named.Name.Name == "Snapshot") {
						t.Errorf("%s restores global %s composition", path, named.Name.Name)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestD2ModelsRemainPassiveSchemas keeps typed records as inert decoded data. Constants, variables,
// or methods would let Diablo-specific interpretation leak out of d2legacy and into generic storage.
func TestD2ModelsRemainPassiveSchemas(t *testing.T) {
	root := filepath.Join(repositoryRoot(t), "internal", "game", "data", "model")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok {
				t.Errorf("%s contains behavior %s; move D2 interpretation to d2legacy", path, function.Name.Name)
			}
			if values, ok := declaration.(*ast.GenDecl); ok && (values.Tok == token.CONST || values.Tok == token.VAR) {
				t.Errorf("%s contains interpreted values; keep raw schemas here and move D2 vocabulary to d2legacy", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// forbiddenLayerImport encodes the permitted direction between repository layers. Returning true
// means the dependency points from an inner mechanism toward a more contextual outer owner.
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
		return hasAnyPrefix(
			dependency,
			"internal/app/",
			"internal/dev/",
			"internal/platform/",
			"internal/presentation/",
			"internal/runtime/",
		)
	}
	if strings.HasPrefix(packagePath, "internal/presentation/") {
		return hasAnyPrefix(dependency, "internal/app/", "internal/dev/", "internal/platform/", "internal/runtime/")
	}
	if strings.HasPrefix(packagePath, "internal/platform/") {
		return hasAnyPrefix(dependency, "internal/app/", "internal/dev/", "internal/runtime/")
	}
	return false
}

// hasAnyPrefix keeps layer checks readable while preserving path-segment prefixes supplied by the
// caller.
func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

// TestCommandRemainsCompositionOnly keeps cmd packages as documented, private wiring code. Exported
// helpers would invite application behavior to depend on executable entry points.
func TestCommandRemainsCompositionOnly(t *testing.T) {
	root := repositoryRoot(t)
	err := filepath.WalkDir(filepath.Join(root, "cmd"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if function.Name.Name != "main" && ast.IsExported(function.Name.Name) {
				t.Errorf("command exposes %s; composition helpers must remain private", function.Name.Name)
			}

			if function.Doc == nil {
				t.Errorf("command function %s lacks a documentation comment", function.Name.Name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// pathWithin performs a segment-aware containment check so similarly prefixed sibling directories do
// not accidentally inherit architectural restrictions intended for root.
func pathWithin(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

// TestRetiredDeveloperDirectoriesCannotReturn prevents old miscellaneous tool buckets from becoming
// an easy home for code whose ownership should be explicit.
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

// TestLegacyRendererObjectAPIAndWorldAdapterCannotReturn protects the ECS presentation boundary from
// the retired direct-renderable object model and its parallel world ownership.
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
	forbidden := map[string]bool{
		"Renderable":          true,
		"NewRenderable":       true,
		"ProvidesRenderables": true,
		"ProvidesTextures":    true,
		"ManagesCameras":      true,
	}
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

// TestNoAccidentalPublicGoPackages requires new reusable code to remain internal unless the project
// deliberately establishes a supported external API under pkg.
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

// repositoryRoot derives the checkout from this source file rather than the process working
// directory, allowing architecture tests to run consistently from any package or CI command.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
