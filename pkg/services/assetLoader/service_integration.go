package assetLoader

import (
	"io"

	dc6 "github.com/gravestench/dc6/pkg"
	"github.com/gravestench/dcc"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	"github.com/gravestench/pl2"
	"github.com/gravestench/servicemesh"
	tbl "github.com/gravestench/tbl_text"

	"github.com/gravestench/dark-magic/pkg/services/lua"
)

var (
	_ servicemesh.Service         = &Service{}
	_ servicemesh.HasLogger       = &Service{}
	_ servicemesh.HasDependencies = &Service{}
	_ lua.UsesLuaEnvironment      = &Service{}
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
	LoadTbl(filepath string) (tbl.TextTable, error)
	UnmarshalTsv(filepath string, destination any) error
	LoadTsv(filepath string) ([]byte, error)
	LoadWav(filepath string) ([]byte, error)
}
