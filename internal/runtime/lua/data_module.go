package modruntime

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
	lua "github.com/yuin/gopher-lua"
)

// DataModule exposes immutable JSON shim data as native Lua values. It keeps
// data ownership in the layered VFS while avoiding a second JSON parser in Lua.
func DataModule(source fs.FS, presentationProfiles ...string) Module {
	presentationProfile := ""
	if len(presentationProfiles) > 0 {
		presentationProfile = presentationProfiles[0]
	}
	return Module{Name: "engine.data/v1", Help: documentedModule("Load structured data and Lua manifests from mounted content.", map[string]CommandHelp{
		"load":          commandHelp("engine.data.load(path)", "Decode a structured data asset into Lua values."),
		"load_manifest": commandHelp("engine.data.load_manifest(path)", "Load and validate a Lua-facing content manifest."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"load": func(state *lua.LState) int {
				name := state.CheckString(1)
				decoded, err := readDataJSON(source, name)
				if err != nil {
					return pushLuaError(state, err)
				}
				value, err := dataToLua(state, decoded)
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(value)
				return 1
			},
			"load_manifest": func(state *lua.LState) int {
				name, expectedSchema := state.CheckString(1), state.CheckString(2)
				decoded, err := readDataJSON(source, name)
				if err == nil {
					err = validateManifest(decoded, expectedSchema)
				}
				if err == nil && expectedSchema == "d2legacy.presentation/v1" {
					document := decoded.(map[string]any)
					decoded, _, err = content.ApplyPresentationProfile(document, presentationProfile)
				}
				if err != nil {
					return pushLuaError(state, err)
				}
				value, err := dataToLua(state, decoded)
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(value)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func readDataJSON(source fs.FS, name string) (any, error) {
	if strings.ToLower(path.Ext(name)) != ".json" {
		return nil, fmt.Errorf("engine.data/v1: only JSON documents are supported")
	}
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return nil, fmt.Errorf("engine.data/v1: read %q: %w", name, err)
	}
	var decoded any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("engine.data/v1: decode %q: %w", name, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("engine.data/v1: decode %q: trailing JSON value", name)
	}
	return decoded, nil
}

func validateManifest(decoded any, expectedSchema string) error {
	document, ok := decoded.(map[string]any)
	if !ok {
		return fmt.Errorf("engine.data/v1: manifest root must be an object")
	}
	schema, ok := document["schema"].(string)
	if !ok || schema != expectedSchema {
		return fmt.Errorf("engine.data/v1: manifest schema is %q, want %q", schema, expectedSchema)
	}
	version, ok := document["version"].(json.Number)
	parsedVersion, err := version.Int64()
	if !ok || err != nil || parsedVersion < 1 {
		return fmt.Errorf("engine.data/v1: manifest version must be a positive integer")
	}
	for _, field := range []string{"game_version", "language", "confidence"} {
		value, ok := document[field].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return fmt.Errorf("engine.data/v1: manifest field %q must be a non-empty string", field)
		}
	}
	resolution, ok := document["resolution"].(map[string]any)
	if !ok || !positiveJSONNumber(resolution["width"]) || !positiveJSONNumber(resolution["height"]) {
		return fmt.Errorf("engine.data/v1: manifest resolution requires positive width and height")
	}
	if profilesValue, exists := document["supported_profiles"]; exists {
		profiles, ok := profilesValue.([]any)
		if !ok || len(profiles) == 0 {
			return fmt.Errorf("engine.data/v1: supported_profiles must be a non-empty array")
		}
		for index, value := range profiles {
			profile, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("engine.data/v1: supported_profiles[%d] must be an object", index)
			}
			for _, field := range []string{"id", "game_version", "language"} {
				text, ok := profile[field].(string)
				if !ok || strings.TrimSpace(text) == "" {
					return fmt.Errorf("engine.data/v1: supported_profiles[%d].%s must be a non-empty string", index, field)
				}
			}
			profileResolution, ok := profile["resolution"].(map[string]any)
			if !ok || !positiveJSONNumber(profileResolution["width"]) || !positiveJSONNumber(profileResolution["height"]) {
				return fmt.Errorf("engine.data/v1: supported_profiles[%d] resolution requires positive width and height", index)
			}
			if scope, exists := profile["screens"]; exists {
				screens, ok := scope.([]any)
				if !ok || len(screens) == 0 {
					return fmt.Errorf("engine.data/v1: supported_profiles[%d].screens must be a non-empty array", index)
				}
				for _, value := range screens {
					if name, ok := value.(string); !ok || strings.TrimSpace(name) == "" {
						return fmt.Errorf("engine.data/v1: supported_profiles[%d].screens contains an invalid ID", index)
					}
				}
			}
			if overrides, exists := profile["overrides"]; exists {
				if _, ok := overrides.(map[string]any); !ok {
					return fmt.Errorf("engine.data/v1: supported_profiles[%d].overrides must be an object", index)
				}
			}
		}
	}
	return nil
}

func positiveJSONNumber(value any) bool {
	number, ok := value.(json.Number)
	if !ok {
		return false
	}
	parsed, err := number.Int64()
	return err == nil && parsed > 0
}

func pushLuaError(state *lua.LState, err error) int {
	state.Push(lua.LNil)
	state.Push(lua.LString(err.Error()))
	return 2
}

func dataToLua(state *lua.LState, input any) (lua.LValue, error) {
	switch value := input.(type) {
	case nil:
		return lua.LNil, nil
	case bool:
		return lua.LBool(value), nil
	case string:
		return lua.LString(value), nil
	case json.Number:
		number, err := value.Float64()
		if err != nil {
			return nil, fmt.Errorf("engine.data/v1: invalid number %q: %w", value, err)
		}
		return lua.LNumber(number), nil
	case []any:
		table := state.NewTable()
		for _, item := range value {
			converted, err := dataToLua(state, item)
			if err != nil {
				return nil, err
			}
			table.Append(converted)
		}
		return table, nil
	case map[string]any:
		table := state.NewTable()
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			converted, err := dataToLua(state, value[key])
			if err != nil {
				return nil, err
			}
			table.RawSetString(key, converted)
		}
		return table, nil
	default:
		return nil, fmt.Errorf("engine.data/v1: unsupported decoded value %T", input)
	}
}
