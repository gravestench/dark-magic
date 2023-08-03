package d2_asset_loader

import (
	"io"
	"time"

	"github.com/gravestench/dc6"
	"github.com/gravestench/dcc"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	"github.com/gravestench/font_table"
	gpl "github.com/gravestench/gpl/pkg"
	"github.com/gravestench/mpq"
	"github.com/gravestench/pl2"
	"github.com/gravestench/runtime"
	tbl "github.com/gravestench/tbl_text"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/dc6_loader"
	"github.com/gravestench/dark-magic/pkg/services/dcc_loader"
	"github.com/gravestench/dark-magic/pkg/services/ds1_loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1_loader"
	"github.com/gravestench/dark-magic/pkg/services/font_table_loader"
	"github.com/gravestench/dark-magic/pkg/services/gpl_loader"
	"github.com/gravestench/dark-magic/pkg/services/mpq_loader"
	"github.com/gravestench/dark-magic/pkg/services/pl2_loader"
	"github.com/gravestench/dark-magic/pkg/services/tbl_loader"
	"github.com/gravestench/dark-magic/pkg/services/tsv_loader"
	"github.com/gravestench/dark-magic/pkg/services/wav_loader"
)

type Service struct {
	logger *zerolog.Logger

	mpq mpq_loader.Dependency

	dc6  dc6_loader.Dependency
	dcc  dcc_loader.Dependency
	ds1  ds1_loader.Dependency
	dt1  dt1_loader.Dependency
	font font_table_loader.Dependency
	gpl  gpl_loader.Dependency
	pl2  pl2_loader.Dependency
	tbl  tbl_loader.Dependency
	tsv  tsv_loader.Dependency
	wav  wav_loader.Dependency
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

func (s *Service) DependenciesResolved() bool {
	for _, dependency := range []any{
		s.mpq,
		s.dc6,
		s.dcc,
		s.ds1,
		s.dt1,
		s.gpl,
		s.pl2,
		s.tbl,
		s.tsv,
		s.wav,
	} {
		if dependency == nil {
			return false
		}
	}

	if len(s.mpq.Archives()) != 11 {
		time.Sleep(time.Second)
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case mpq_loader.Dependency:
			s.mpq = candidate
		case dc6_loader.Dependency:
			s.dc6 = candidate
		case dcc_loader.Dependency:
			s.dcc = candidate
		case ds1_loader.Dependency:
			s.ds1 = candidate
		case dt1_loader.Dependency:
			s.dt1 = candidate
		case gpl_loader.Dependency:
			s.gpl = candidate
		case pl2_loader.Dependency:
			s.pl2 = candidate
		case tbl_loader.Dependency:
			s.tbl = candidate
		case tsv_loader.Dependency:
			s.tsv = candidate
		case wav_loader.Dependency:
			s.wav = candidate
		}
	}
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

func (s *Service) LoadGpl(filepath string) (*gpl.GPL, error) {
	return s.gpl.Load(filepath)
}

func (s *Service) LoadFontTable(filepath string) (*font_table.Font, error) {
	return s.font.Load(filepath)
}

func (s *Service) LoadPl2(filepath string) (*pl2.PL2, error) {
	return s.pl2.Load(filepath)
}

func (s *Service) LoadTsv(filepath string, destination any) error {
	return s.tsv.Load(filepath, destination)
}

func (s *Service) LoadTbl(filepath string) (*tbl.TextTable, error) {
	return s.tbl.Load(filepath)
}

func (s *Service) LoadWav(filepath string) ([]byte, error) {
	return s.wav.Load(filepath)
}
