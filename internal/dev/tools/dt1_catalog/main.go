// Command dt1_catalog prints the renderer-independent identities contained in
// one or more DT1 files. It is intentionally metadata-only: block pixels are
// never decoded, so it is safe to use while researching large production sets.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
)

type catalogFlags struct {
	assets string
	stamps string
}

// main preserves the command's exit-code split: missing input is usage error 2, while catalog failures use 1.
func main() {
	options := parseCatalogFlags()
	if !options.hasInput() {
		fmt.Fprintln(os.Stderr, "dt1_catalog: -assets or -stamps is required")
		os.Exit(2)
	}

	source, err := content.FromEnvironment()
	if err != nil {
		fatal(err)
	}

	if err := printCatalog(source, os.Stdout, options); err != nil {
		fatal(err)
	}
}

// parseCatalogFlags keeps the established flag names and defaults in one place so the CLI contract remains visible.
func parseCatalogFlags() catalogFlags {
	assets := flag.String("assets", "", "comma-separated DT1 asset paths")
	stamps := flag.String("stamps", "", "comma-separated DS1 stamps whose tile identities should be printed")

	flag.Parse()

	return catalogFlags{assets: *assets, stamps: *stamps}
}

// hasInput rejects lists containing only separators or whitespace before opening the configured content source.
func (options catalogFlags) hasInput() bool {
	return strings.TrimSpace(options.assets) != "" || strings.TrimSpace(options.stamps) != ""
}

// printCatalog preserves assets-before-stamps ordering and stops at the first inspection failure.
func printCatalog(source *content.FS, output io.Writer, options catalogFlags) error {
	if err := inspectAssetList(source, output, options.assets, inspectDT1); err != nil {
		return err
	}

	return inspectAssetList(source, output, options.stamps, inspectStamp)
}

// inspectAssetList normalizes each comma-separated path while retaining caller-defined order and duplicates.
func inspectAssetList(
	source *content.FS,
	output io.Writer,
	paths string,
	inspectAsset func(*content.FS, io.Writer, string) error,
) error {
	for _, path := range strings.Split(paths, ",") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		// Immediate failure prevents later output from making a partial catalog look complete.
		if err := inspectAsset(source, output, path); err != nil {
			return err
		}
	}

	return nil
}

// closeReadAsset preserves the command's historical choice to ignore close errors after read-only asset access.
func closeReadAsset(asset io.Closer) {
	_ = asset.Close()
}

// writeCatalogf keeps stdout best-effort so only asset access and decoding determine the command's error behavior.
func writeCatalogf(output io.Writer, format string, arguments ...any) {
	_, _ = fmt.Fprintf(output, format, arguments...)
}

// fatal applies the command prefix consistently and terminates with the established runtime-failure status.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, "dt1_catalog:", err)
	os.Exit(1)
}
