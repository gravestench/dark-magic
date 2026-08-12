package d2legacy_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	. "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// This test deliberately boots the policy headlessly. It proves the Go client
// no longer needs typed item structs to interpret fixture footprints, melee
// values, graphics, placements, or vendor multipliers.
func TestD2LegacyBuildsDevelopmentItemsFromRawRecords(t *testing.T) {
	records := presetLuaRecords{
		"data/global/excel/weapons.txt": {{
			"code": "ssd", "invwidth": "1", "invheight": "3", "cost": "466",
			"invfile": "invssd", "flippyfile": "flpssd", "component": "RH",
			"alternategfx": "SSD", "wclass": "1hs", "rangeadder": "1",
			"mindam": "2", "maxdam": "7",
		}},
		"data/global/excel/armor.txt": {{
			"code": "cap", "invwidth": "2", "invheight": "2", "cost": "64",
			"invfile": "invcap", "flippyfile": "flpcap", "component": "0",
			"alternategfx": "CAP",
		}},
		"data/global/excel/misc.txt": {
			{"code": "hp1", "invwidth": "1", "invheight": "1", "cost": "30", "invfile": "invhp1", "flippyfile": "flphp1"},
			{"code": "mp1", "invwidth": "1", "invheight": "1", "cost": "30", "invfile": "invmp1", "flippyfile": "flpmp1"},
		},
		"data/global/excel/Npc.txt": {{
			"npc": "Akara", "buy mult": "2048", "sell mult": "1024", "max buy": "5000",
		}},
	}
	runtime := New()
	if err := runtime.RegisterModule(RecordsModule(records)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())

	script := fstest.MapFS{"test.lua": {Data: []byte(`
local fixtures=require("d2legacy.items.development_fixtures").build(true)
assert(fixtures.owner=="local-player" and fixtures.inventory_width==10)
assert(fixtures.trade_terms.Akara.buy_multiplier==2048)
local by_id={}
for _,item in ipairs(fixtures.items) do by_id[item.id]=item end
local sword=assert(by_id["fixture-short-sword"])
assert(sword.width==1 and sword.height==3 and sword.container=="inventory")
assert(sword.body_slots=="rarm,larm" and sword.composite=="RH=SSD")
assert(sword.weapon_class=="1HS" and sword.melee_range==2)
assert(sword.physical_min==512 and sword.physical_max==1792)
assert(by_id["fixture-vendor-short-sword"].slot=="weap")
assert(by_id["fixture-hireling-cap"].slot=="head")
assert(by_id["fixture-mp1"].container=="belt")
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
