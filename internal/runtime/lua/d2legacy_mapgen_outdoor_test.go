package modruntime

import (
	"context"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
)

func TestD2LegacyLuaOwnsBloodMoorRecipeAndStructurePolicy(t *testing.T) {
	records := presetLuaRecords{
		"data/global/excel/Levels.txt": {{
			"Id": "2", "Act": "0", "DrlgType": "3", "LevelType": "2", "SizeX": "80", "SizeY": "80",
		}},
		"data/global/excel/LvlTypes.txt": {{}, {}, {"File 1": "Act1/Outdoors/Outdoor1.dt1"}},
	}
	for _, definition := range []int{17, 26, 27, 28, 29, 30, 35} {
		row := map[string]string{
			"Def": fmt.Sprint(definition), "SizeX": "8", "SizeY": "8", "Files": "1",
			"File1": fmt.Sprintf("Act1/Outdoors/fill%d.ds1", definition), "Dt1Mask": "1", "Populate": "1",
		}
		if definition == 26 || definition == 27 || definition == 28 {
			row["Files"] = "4"
			for variant := 2; variant <= 4; variant++ {
				row[fmt.Sprintf("File%d", variant)] = fmt.Sprintf("Act1/Outdoors/structure%d-%d.ds1", definition, variant)
			}
		}
		records["data/global/excel/LvlPrest.txt"] = append(records["data/global/excel/LvlPrest.txt"], row)
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
local outdoor=require("d2legacy.mapgen.outdoor")
for _,direction in ipairs({"north","east","south","west"}) do
  local zone=outdoor.generate(2,42,direction,0)
  assert(zone.kind=="outdoor" and #zone.rooms==100 and #zone.links==180)
  assert(#zone.stamps>=120 and zone.checksum==outdoor.generate(2,42,direction,0).checksum)
  local path={}; for _,tile in ipairs(zone.paths) do path[tile.x..":"..tile.y]=true end
  assert(path[zone.warps[1].x..":"..zone.warps[1].y])
  assert(path[zone.warps[2].x..":"..zone.warps[2].y])
  local bridge,passable=0,0
  for _,tile in ipairs(zone.structures) do
    if path[tile.x..":"..tile.y] then assert(tile.passable, "route intersects blocked structure") end
    if tile.kind=="bridge" then bridge=bridge+1; if tile.passable then passable=passable+1 end end
  end
  assert(bridge==64 and passable==48)
end
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
