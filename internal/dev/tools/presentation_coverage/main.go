package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"github.com/gravestench/dark-magic/internal/assets/catalog"
	"github.com/gravestench/dark-magic/internal/content"
)

func main() {
	source := content.Shim()
	manifest := readManifest[assetcatalog.Manifest](source, "manifests/asset-catalog.v1.json")
	fixture := readManifest[assetcatalog.Fixture](source, "manifests/asset-fixture.v1.json")
	coverage, err := assetcatalog.BuildCoverage(source, manifest, fixture)
	if err != nil {
		fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(coverage); err != nil {
		fatal(err)
	}
}

func readManifest[T any](source fs.FS, name string) T {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		fatal(err)
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		fatal(err)
	}
	return value
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
