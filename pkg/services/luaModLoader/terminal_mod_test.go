package luaModLoader

import (
	"bytes"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

const terminalLuaStubs = `
local function renderable()
  local r = { enabled = true, opacity = 1, x = 0, y = 0 }
  r.UUID = function() return "test-renderable" end
  r.Opacity = function(value) if value ~= nil then r.opacity = value end return r.opacity end
  r.Position = function(x, y) if x ~= nil then r.x, r.y = x, y end return r.x, r.y end
  r.Origin = function() end
  r.Enable = function(value) if value ~= nil then r.enabled = value end return r.enabled end
  r.Parent = function(parent) r.parent = parent end
  r.ZIndex = function(value) r.z = value end
  return r
end
local function tween()
  local t = {}
  t.OnUpdate = function(fn) t.update = fn end
  t.OnStart = function(fn) t.start = fn end
  t.Time = function() end
  t.Stop = function() end
  t.Play = function() end
  t.Start = function()
    if t.start then t.start() end
    if t.update then t.update(0.5) end
  end
  return t
end
api = {
  renderer = { NewRenderable = renderable, window = { Size = function() return 800, 600 end } },
  ui = { FillRect = function() return renderable() end },
  tweens = { New = tween, Add = function() end },
}
`

func TestTerminalModInitializesAndRunsTweenCallbacks(t *testing.T) {
	state := lua.NewState()
	defer state.Close()
	if err := state.DoString(terminalLuaStubs); err != nil {
		t.Fatal(err)
	}
	data, err := internalMods.ReadFile("internal/mods/terminal/init.lua")
	if err != nil {
		t.Fatal(err)
	}
	function, err := state.Load(bytes.NewReader(data), "@terminal/init.lua")
	if err != nil {
		t.Fatal(err)
	}
	if err := state.CallByParam(lua.P{Fn: function, NRet: 1, Protect: true}); err != nil {
		t.Fatal(err)
	}
	terminal := state.Get(-1)
	state.Pop(1)
	init := state.GetField(terminal, "Init")
	if err := state.CallByParam(lua.P{Fn: init, NRet: 0, Protect: true}, terminal); err != nil {
		t.Fatal(err)
	}
	table := terminal.(*lua.LTable)
	if state.GetField(table, "enabled") != lua.LTrue {
		t.Fatal("terminal was not enabled")
	}
	root := state.GetField(table, "root").(*lua.LTable)
	if state.GetField(root, "enabled") != lua.LTrue {
		t.Fatal("terminal root was not enabled")
	}
}
