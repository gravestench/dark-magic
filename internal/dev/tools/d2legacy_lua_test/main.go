// Command d2legacy_lua_test creates a readable sidecar suite for a d2legacy Lua module.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	usageMessage         = "usage: go run ./internal/dev/tools/d2legacy_lua_test path/to/module.lua"
	productionFileError  = "input must be a production .lua file"
	luaDirectoryError    = "input must be below a lua/ directory"
	luaDirectoryMarker   = "/lua/"
	luaDirectoryPrefix   = "lua/"
	luaFileExtension     = ".lua"
	luaTestFileExtension = "_test.lua"
)

// sidecarSpecification keeps the derived output and require paths together so they cannot diverge after validation.
type sidecarSpecification struct {
	outputPath string
	moduleName string
}

// main validates the command before touching disk and preserves distinct exit codes for usage and filesystem errors.
func main() {
	specification, err := parseSidecarSpecification(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// Exclusive creation protects an existing hand-edited suite from accidental replacement.
	file, err := os.OpenFile(specification.outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create %s: %v\n", specification.outputPath, err)
		os.Exit(1)
	}

	// Close errors remain non-authoritative after the write succeeds, preserving the command's prior exit behavior.
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.WriteString(sidecarTemplate(specification.moduleName)); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", specification.outputPath, err)
		os.Exit(1)
	}

	fmt.Println(specification.outputPath)
}

// parseSidecarSpecification normalizes the input once so the printed path, created file, and module stay consistent.
func parseSidecarSpecification(arguments []string) (sidecarSpecification, error) {
	if len(arguments) != 2 {
		return sidecarSpecification{}, errors.New(usageMessage)
	}

	productionPath := filepath.Clean(arguments[1])
	if filepath.Ext(productionPath) != luaFileExtension || strings.HasSuffix(productionPath, luaTestFileExtension) {
		return sidecarSpecification{}, errors.New(productionFileError)
	}

	module, ok := moduleName(productionPath)
	if !ok {
		return sidecarSpecification{}, errors.New(luaDirectoryError)
	}

	outputPath := strings.TrimSuffix(productionPath, luaFileExtension) + luaTestFileExtension

	return sidecarSpecification{
		outputPath: outputPath,
		moduleName: module,
	}, nil
}

// moduleName derives the final lua-relative namespace so checkout prefixes never leak into generated require calls.
func moduleName(path string) (string, bool) {
	slashPath := filepath.ToSlash(path)

	markerIndex := strings.LastIndex(slashPath, luaDirectoryMarker)
	if markerIndex >= 0 {
		luaRelativePath := slashPath[markerIndex+len(luaDirectoryMarker):]

		return dottedModuleName(luaRelativePath), true
	}

	if !strings.HasPrefix(slashPath, luaDirectoryPrefix) {
		return "", false
	}

	luaRelativePath := strings.TrimPrefix(slashPath, luaDirectoryPrefix)

	return dottedModuleName(luaRelativePath), true
}

// dottedModuleName converts a lua-relative filename into the dotted module identifier consumed by Lua require.
func dottedModuleName(luaRelativePath string) string {
	withoutExtension := strings.TrimSuffix(luaRelativePath, luaFileExtension)

	return strings.ReplaceAll(withoutExtension, "/", ".")
}

// sidecarTemplate renders the stable starter suite separately so its exact generated format remains testable.
func sidecarTemplate(moduleName string) string {
	return fmt.Sprintf(`local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("describes expected behavior", function(t)
            t:check(function()
                local subject = require(%q)
                test.expect(type(subject), "module type"):equals("table")
            end)
        end),
    },
})
`, moduleName)
}
