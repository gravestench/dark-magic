package d2legacy_test

import (
	"context"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	. "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func TestD2LegacyLuaOwnsActOneCaveMazePolicy(t *testing.T) {
	records := presetLuaRecords{
		"data/global/excel/Levels.txt": {{"Id": "9", "Act": "0", "DrlgType": "1", "LevelType": "3"}},
		"data/global/excel/LvlMaze.txt": {{
			"Level": "9", "Rooms": "12", "Rooms(N)": "14", "Rooms(H)": "16",
			"SizeX": "24", "SizeY": "24", "Merge": "500",
		}},
		"data/global/excel/LvlTypes.txt": {{}, {}, {}, {"File 1": "Act1/Caves/floor.dt1"}},
	}
	for definition := 53; definition <= 67; definition++ {
		records["data/global/excel/LvlPrest.txt"] = append(records["data/global/excel/LvlPrest.txt"], map[string]string{
			"Def": fmt.Sprint(definition), "SizeX": "24", "SizeY": "24", "Files": "1",
			"File1": "Act1/Caves/room.ds1", "Dt1Mask": "1",
		})
	}
	for definition := 83; definition <= 90; definition++ {
		records["data/global/excel/LvlPrest.txt"] = append(records["data/global/excel/LvlPrest.txt"], map[string]string{
			"Def": fmt.Sprint(definition), "SizeX": "24", "SizeY": "24", "Files": "1",
			"File1": "Act1/Caves/special.ds1", "Dt1Mask": "1",
		})
	}
	runtime := New()
	for _, module := range []Module{RecordsModule(records), DeterministicModule(), WorldgenModule()} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := fstest.MapFS{"test.lua": {Data: []byte(`
local maze=require("d2legacy.mapgen.maze")
local zone=maze.generate(9,42,0)
assert(zone.kind=="maze" and #zone.rooms==12 and #zone.stamps==12)
assert(#zone.links>=11)
local previous,next_level=0,0
local occupied={}
for _,stamp in ipairs(zone.stamps) do
  if stamp.role=="previous-level" then previous=previous+1 end
  if stamp.role=="next-level" then next_level=next_level+1 end
  local key=stamp.x..":"..stamp.y
  assert(not occupied[key]); occupied[key]=true
end
assert(previous==1 and next_level==1)
assert(zone.checksum==maze.generate(9,42,0).checksum)
assert(#maze.generate(9,42,1).rooms==14)
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
