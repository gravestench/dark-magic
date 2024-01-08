package assetLoader

import (
	"image/color"
	"io"

	"github.com/gravestench/cof"
	dc6 "github.com/gravestench/dc6/pkg"
	"github.com/gravestench/dcc"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	"github.com/gravestench/font_table"
	"github.com/gravestench/pl2"
	"github.com/gravestench/servicemesh"
	tbl "github.com/gravestench/tbl_text"

	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ luaManager.LuaPlugin        = &Service{}
	_ LoadsDiabloFiles            = &Service{}
)

type Dependency = LoadsDiabloFiles

type LoadsDiabloFiles interface {
	Load(filepath string) (io.Reader, error)
	LoadDc6(filepath string) (*dc6.DC6, error)
	LoadDcc(filepath string) (*dcc.DCC, error)
	LoadDs1(filepath string) (*ds1.DS1, error)
	LoadDt1(filepath string) (*dt1.DT1, error)
	LoadPl2(filepath string) (*pl2.PL2, error)
	LoadTsv(filepath string) ([]byte, error)
	LoadTbl(filepath string) (tbl.TextTable, error)
	LoadWav(filepath string) ([]byte, error)
	LoadCOF(filepath string) (*cof.COF, error)
	LoadFontTable(filepath string) (*font_table.Font, error)
	LoadAnimationData(filepath string) (*cof.AnimationData, error)
	UnmarshalTsv(filepath string, destination any) error
	ExtractPaletteFromPl2(pathPL2 string) (color.Palette, error)
}
