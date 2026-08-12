// Command d2legacy_lua_test creates a readable sidecar suite for a d2legacy Lua module.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./internal/dev/tools/d2legacy_lua_test path/to/module.lua")
		os.Exit(2)
	}
	production := filepath.Clean(os.Args[1])
	if filepath.Ext(production) != ".lua" || strings.HasSuffix(production, "_test.lua") {
		fmt.Fprintln(os.Stderr, "input must be a production .lua file")
		os.Exit(2)
	}
	output := strings.TrimSuffix(production, ".lua") + "_test.lua"
	module, ok := moduleName(production)
	if !ok {
		fmt.Fprintln(os.Stderr, "input must be below a lua/ directory")
		os.Exit(2)
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create %s: %v\n", output, err)
		os.Exit(1)
	}
	defer file.Close()
	template := fmt.Sprintf(`local test = require("d2legacy.tests/v1")

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
`, module)
	if _, err := file.WriteString(template); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", output, err)
		os.Exit(1)
	}
	fmt.Println(output)
}

func moduleName(path string) (string, bool) {
	slash := filepath.ToSlash(path)
	marker := "/lua/"
	index := strings.LastIndex(slash, marker)
	if index < 0 {
		if strings.HasPrefix(slash, "lua/") {
			module := strings.TrimSuffix(strings.TrimPrefix(slash, "lua/"), ".lua")
			return strings.ReplaceAll(module, "/", "."), true
		} else {
			return "", false
		}
	}
	module := strings.TrimSuffix(slash[index+len(marker):], ".lua")
	return strings.ReplaceAll(module, "/", "."), true
}
