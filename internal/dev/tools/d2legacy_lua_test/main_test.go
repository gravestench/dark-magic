package main

import "testing"

// TestModuleName covers both accepted lua roots and rejection so generated require paths stay checkout-independent.
func TestModuleName(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantModule string
		wantOK     bool
	}{
		{
			name:       "repository path",
			path:       "internal/content/d2legacy/lua/d2legacy/policy/damage.lua",
			wantModule: "d2legacy.policy.damage",
			wantOK:     true,
		},
		{
			name:       "relative lua root",
			path:       "lua/d2legacy/policy/damage.lua",
			wantModule: "d2legacy.policy.damage",
			wantOK:     true,
		},
		{
			name:       "last lua root",
			path:       "checkout/lua/archive/lua/d2legacy/policy/damage.lua",
			wantModule: "d2legacy.policy.damage",
			wantOK:     true,
		},
		{
			name: "missing lua root",
			path: "internal/content/d2legacy/damage.lua",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotModule, gotOK := moduleName(test.path)
			if gotModule != test.wantModule || gotOK != test.wantOK {
				t.Fatalf("moduleName(%q) = %q, %v; want %q, %v", test.path, gotModule, gotOK, test.wantModule, test.wantOK)
			}
		})
	}
}

// TestParseSidecarSpecification keeps validation precedence, path cleaning, and derived output names observable.
func TestParseSidecarSpecification(t *testing.T) {
	tests := []struct {
		name              string
		arguments         []string
		wantSpecification sidecarSpecification
		wantError         string
	}{
		{
			name:      "missing input",
			arguments: []string{"d2legacy_lua_test"},
			wantError: usageMessage,
		},
		{
			name:      "extra input",
			arguments: []string{"d2legacy_lua_test", "lua/first.lua", "lua/second.lua"},
			wantError: usageMessage,
		},
		{
			name:      "non-lua input",
			arguments: []string{"d2legacy_lua_test", "lua/d2legacy/policy/damage.txt"},
			wantError: productionFileError,
		},
		{
			name:      "existing test input",
			arguments: []string{"d2legacy_lua_test", "lua/d2legacy/policy/damage_test.lua"},
			wantError: productionFileError,
		},
		{
			name:      "missing lua directory",
			arguments: []string{"d2legacy_lua_test", "internal/content/d2legacy/policy/damage.lua"},
			wantError: luaDirectoryError,
		},
		{
			name:      "cleaned production path",
			arguments: []string{"d2legacy_lua_test", "lua/d2legacy/systems/../policy/damage.lua"},
			wantSpecification: sidecarSpecification{
				outputPath: "lua/d2legacy/policy/damage_test.lua",
				moduleName: "d2legacy.policy.damage",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotSpecification, err := parseSidecarSpecification(test.arguments)
			if gotSpecification != test.wantSpecification {
				t.Errorf(
					"parseSidecarSpecification(%q) = %+v; want %+v",
					test.arguments,
					gotSpecification,
					test.wantSpecification,
				)
			}

			if gotError := commandErrorText(err); gotError != test.wantError {
				t.Errorf(
					"parseSidecarSpecification(%q) error = %q; want %q",
					test.arguments,
					gotError,
					test.wantError,
				)
			}
		})
	}
}

// TestSidecarTemplate locks the authoring profile, starter assertion, module quoting, and final newline as one format.
func TestSidecarTemplate(t *testing.T) {
	const expected = `local test = require("d2legacy.tests/v1")

return test.suite({
    profile = "authority",
    tier = "fast",
    cases = {
        test.case("describes expected behavior", function(t)
            t:check(function()
                local subject = require("d2legacy.policy.damage")
                test.expect(type(subject), "module type"):equals("table")
            end)
        end),
    },
})
`

	got := sidecarTemplate("d2legacy.policy.damage")
	if got != expected {
		t.Fatalf("sidecarTemplate() mismatch (-want +got):\nwant:\n%s\ngot:\n%s", expected, got)
	}
}

// commandErrorText gives table tests a comparable empty value while retaining exact command-facing error text.
func commandErrorText(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
