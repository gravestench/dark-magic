package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCompileMatchesGameLua pins the cases where system luac 5.4/5.5 and the
// game's gopher-lua disagree, so the check fails exactly when the game would.
func TestCompileMatchesGameLua(t *testing.T) {
	t.Parallel()

	for name, source := range map[string]string{
		"loop variable assignment": "for i = 1, 3 do\n  i = i + 1\nend\n",
		"goto":                     "for i = 1, 2 do\n  goto skip\n  ::skip::\nend\n",
	} {
		if err := compile(name+".lua", source); err != nil {
			t.Errorf("%s rejected: %v", name, err)
		}
	}

	for name, source := range map[string]string{
		"floor division":   "local a = 7 // 2\n",
		"right shift":      "local a = 1 >> 2\n",
		"bitwise and":      "local a = 1 & 2\n",
		"undefined label":  "goto missing\n",
		"unclosed context": "local function f()\n",
	} {
		if err := compile(name+".lua", source); err == nil {
			t.Errorf("%s accepted", name)
		} else if !strings.Contains(err.Error(), name+".lua") {
			t.Errorf("%s error does not name the file: %v", name, err)
		}
	}
}

// TestCheckPathsReportsEveryFailingLuaFile keeps one run useful: every broken
// file is listed, and non-Lua files are ignored.
func TestCheckPathsReportsEveryFailingLuaFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	files := map[string]string{
		"good.lua":        "return 1\n",
		"nested/bad.lua":  "return 7 // 2\n",
		"nested/also.lua": "return 1 >> 2\n",
		"notes.txt":       "return 7 // 2\n",
	}

	for name, source := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	failures, err := checkPaths([]string{root})
	if err != nil {
		t.Fatal(err)
	}

	if len(failures) != 2 {
		t.Fatalf("failures = %v, want the two broken .lua files", failures)
	}

	if _, err := checkPaths([]string{filepath.Join(root, "absent")}); err == nil {
		t.Fatal("missing path was not reported")
	}
}
