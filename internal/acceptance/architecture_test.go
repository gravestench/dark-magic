package acceptance

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestRetiredPublicPackagesCannotReturn(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
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
		"github.com/gravestench/dark-magic/internal/service_template": {},
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
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
