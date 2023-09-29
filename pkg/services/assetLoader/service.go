package assetLoader

import (
	"io"

	"github.com/gravestench/dc6"
	"github.com/gravestench/dcc"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	"github.com/gravestench/font_table"
	"github.com/gravestench/mpq"
	"github.com/gravestench/pl2"
	"github.com/gravestench/runtime"
	tbl "github.com/gravestench/tbl_text"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/cofLoader"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/fontTableLoader"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
)

type Service struct {
	logger *zerolog.Logger

	mpq  mpqLoader.Dependency
	dc6  dc6Loader.Dependency
	dcc  dccLoader.Dependency
	ds1  ds1Loader.Dependency
	dt1  dt1Loader.Dependency
	cof  cofLoader.Dependency
	font fontTableLoader.Dependency
	pl2  pl2Loader.Dependency
	tbl  tblLoader.Dependency
	tsv  tsvLoader.Dependency
	wav  wavLoader.Dependency
}

func (s *Service) Archives() map[string]*mpq.MPQ {
	return s.mpq.Archives()
}

func (s *Service) AddArchive(filepath string) error {
	return s.mpq.AddArchive(filepath)
}

func (s *Service) Load(filepath string) (io.Reader, error) {
	s.logger.Info().Msgf("loading file: %v", filepath)
	return s.mpq.Load(filepath)
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) Init(rt runtime.R) {
}

func (s *Service) Name() string {
	return "Diablo II Unified Asset Loader"
}

func (s *Service) LoadDc6(filepath string) (*dc6.DC6, error) {
	return s.dc6.Load(filepath)
}

func (s *Service) LoadDcc(filepath string) (*dcc.DCC, error) {
	return s.dcc.Load(filepath)
}

func (s *Service) LoadDs1(filepath string) (*ds1.DS1, error) {
	return s.ds1.Load(filepath)
}

func (s *Service) LoadDt1(filepath string) (*dt1.DT1, error) {
	return s.dt1.Load(filepath)
}

func (s *Service) LoadFontTable(filepath string) (*font_table.Font, error) {
	return s.font.Load(filepath)
}

func (s *Service) LoadPl2(filepath string) (*pl2.PL2, error) {
	return s.pl2.Load(filepath)
}

func (s *Service) UnmarshalTsv(filepath string, destination any) error {
	return s.tsv.Unmarshal(filepath, destination)
}

func (s *Service) LoadTsv(filepath string) ([]byte, error) {
	return s.tsv.Load(filepath)
}

func (s *Service) LoadTbl(filepath string) (tbl.TextTable, error) {
	return s.tbl.Load(filepath)
}

func (s *Service) LoadWav(filepath string) ([]byte, error) {
	return s.wav.Load(filepath)
}
