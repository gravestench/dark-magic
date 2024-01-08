package ui

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"

	luaService "github.com/gravestench/dark-magic/pkg/services/luaManager"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

const LuaApiKey = "ui"

var _ luaService.LuaPlugin = &Service{}

// these methods are automatically invoked
// by the lua service to export stuff into the
// lua environment for use in scripts.

func (s *Service) ExportToLua(state *lua.LState, rootTable *lua.LTable) {
	table := state.NewTable()

	s.bindMethods(state, table)

	state.SetField(rootTable, LuaApiKey, table)
}

func (s *Service) UnexportFromLua(state *lua.LState, rootTable *lua.LTable) {
	state.SetField(rootTable, LuaApiKey, lua.LNil)
}

func (s *Service) bindMethods(state *lua.LState, table *lua.LTable) {
	fnMap := map[string]lua.LGFunction{
		"FillRect": s.luaFillRect,
	}

	for key, fn := range fnMap {
		state.SetField(table, key, state.NewFunction(fn))
	}
}

type providesLuaRenderable interface {
	MakeLuaRenderable(r raylibRenderer.Renderable, L *lua.LState) *lua.LTable
}

func (s *Service) luaFillRect(L *lua.LState) int {
	numArgs := L.GetTop()

	if numArgs < 7 {
		return 0
	}

	x := L.CheckInt(1)
	y := L.CheckInt(2)
	w := L.CheckInt(3)
	h := L.CheckInt(4)
	strokeWidth := L.CheckInt(5)

	fillColor, err := ParseHexColor(L.CheckString(6))
	if err != nil {
		fillColor = image.White
	}

	strokeColor, err := ParseHexColor(L.CheckString(7))
	if err != nil {
		strokeColor = image.White
	}

	r := s.FillRect(x, y, w, h, strokeWidth, fillColor, strokeColor)

	table := s.renderer.(providesLuaRenderable).MakeLuaRenderable(r, L)

	L.Push(table)

	return 1
}

// ParseHexColor parses a string representing a color in hex format and returns a color.Color.
func ParseHexColor(s string) (color.Color, error) {
	var r, g, b uint64
	var err error

	// Remove the leading '#' or '0x' if present
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	} else if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}

	// Handle 3-digit hex colors (e.g., "#333")
	if len(s) == 3 {
		r, err = strconv.ParseUint(strings.Repeat(string(s[0]), 2), 16, 8)
		if err != nil {
			return nil, err
		}
		g, err = strconv.ParseUint(strings.Repeat(string(s[1]), 2), 16, 8)
		if err != nil {
			return nil, err
		}
		b, err = strconv.ParseUint(strings.Repeat(string(s[2]), 2), 16, 8)
		if err != nil {
			return nil, err
		}
	} else if len(s) == 6 { // Handle 6-digit hex colors (e.g., "#FF00FF" or "0xFF00FF")
		r, err = strconv.ParseUint(s[0:2], 16, 8)
		if err != nil {
			return nil, err
		}
		g, err = strconv.ParseUint(s[2:4], 16, 8)
		if err != nil {
			return nil, err
		}
		b, err = strconv.ParseUint(s[4:6], 16, 8)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid color format")
	}

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
}
