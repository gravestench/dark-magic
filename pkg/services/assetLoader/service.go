package assetLoader

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/gravestench/cof"
	dc6 "github.com/gravestench/dc6/pkg"
	"github.com/gravestench/dcc"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	"github.com/gravestench/font_table"
	"github.com/gravestench/pl2"
	"github.com/gravestench/servicemesh"
	tbl "github.com/gravestench/tbl_text"
	"github.com/gravestench/tsv"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
)

type Service struct {
	logger *slog.Logger

	Config

	file fileLoader.Dependency

	cache struct {
		dc6  *cache.Cache
		dcc  *cache.Cache
		ds1  *cache.Cache
		dt1  *cache.Cache
		cof  *cache.Cache
		font *cache.Cache
		pl2  *cache.Cache
		tbl  *cache.Cache
		tsv  *cache.Cache
		wav  *cache.Cache
	}
}

type cacheClosure struct {
	cache  **cache.Cache
	Budget int
}

func (c *cacheClosure) CacheBudget() int {
	return c.Budget
}

func (c *cacheClosure) FlushCache(newCache *cache.Cache) {
	*c.cache = newCache
}

func (s *Service) Caches() []cacheManager.HasCache {
	return []cacheManager.HasCache{
		&cacheClosure{
			cache:  &s.cache.dc6,
			Budget: s.Config.Dc6CacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.dcc,
			Budget: s.Config.DccCacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.ds1,
			Budget: s.Config.Ds1CacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.dt1,
			Budget: s.Config.Dt1CacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.cof,
			Budget: s.Config.CofCacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.font,
			Budget: s.Config.FontCacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.pl2,
			Budget: s.Config.Pl2CacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.tbl,
			Budget: s.Config.TblCacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.tsv,
			Budget: s.Config.TsvCacheMB,
		},
		&cacheClosure{
			cache:  &s.cache.wav,
			Budget: s.Config.WavCacheMB,
		},
	}
}

func (s *Service) Load(filepath string) (io.Reader, error) {
	s.logger.Info("loading file", "path", filepath)
	return s.file.Load(filepath)
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) Init(mesh servicemesh.Mesh) {
}

func (s *Service) Name() string {
	return "Unified Asset Loader"
}

func (s *Service) Ready() bool {
	if s.file == nil {
		return false
	}

	return true
}

func (s *Service) LoadDc6(filepath string) (*dc6.DC6, error) {
	cachedData, isCached := s.cache.dc6.Retrieve(filepath)
	if isCached {
		return cachedData.(*dc6.DC6), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	dc6Image, err := dc6.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dc6 %q: %v", filepath, err)
	}

	if err = s.cache.dc6.Insert(filepath, dc6Image, len(data)); err != nil {
		s.Logger().Error("caching file", "error", err)
	}

	s.Logger().Info("loaded DC6 file", "path", filepath)

	return dc6Image, nil
}

func (s *Service) LoadDcc(filepath string) (*dcc.DCC, error) {
	cachedData, isCached := s.cache.dcc.Retrieve(filepath)
	if isCached {
		return cachedData.(*dcc.DCC), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	dccImage, err := dcc.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dcc %q: %v", filepath, err)
	}

	if err = s.cache.dcc.Insert(filepath, dccImage, len(data)); err != nil {
		s.Logger().Error("caching file", "error", err)
	}

	s.Logger().Info("loaded DCC file", "path", filepath)

	return dccImage, nil
}

func (s *Service) LoadDs1(filepath string) (*ds1.DS1, error) {
	cachedData, isCached := s.cache.ds1.Retrieve(filepath)
	if isCached {
		return cachedData.(*ds1.DS1), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	ds1Object, err := ds1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing ds1 %q: %v", filepath, err)
	}

	if err = s.cache.ds1.Insert(filepath, ds1Object, len(data)); err != nil {
		s.logger.Error("caching file", "error", err)
	}

	s.logger.Info("loaded DS1 file", "path", filepath)

	return ds1Object, nil
}

func (s *Service) LoadDt1(filepath string) (*dt1.DT1, error) {
	cachedData, isCached := s.cache.dt1.Retrieve(filepath)
	if isCached {
		return cachedData.(*dt1.DT1), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	dt1Object, err := dt1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing dt1 %q: %v", filepath, err)
	}

	if err = s.cache.dt1.Insert(filepath, dt1Object, len(data)); err != nil {
		s.Logger().Error("caching file", "error", err)
	}

	s.Logger().Info("loaded DT1 file", "path", filepath)

	return dt1Object, nil
}

func (s *Service) LoadFontTable(filepath string) (*font_table.Font, error) {
	cachedData, isCached := s.cache.font.Retrieve(filepath)
	if isCached {
		return cachedData.(*font_table.Font), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	font, err := font_table.Load(data)
	if err != nil {
		return nil, fmt.Errorf("parsing font table %q: %v", filepath, err)
	}

	if err = s.cache.font.Insert(filepath, font, len(data)); err != nil {
		s.Logger().Error("caching file", "error", err)
	}

	s.Logger().Info("loaded font table", "path", filepath)

	return font, nil
}

func (s *Service) LoadPl2(filepath string) (*pl2.PL2, error) {
	cachedData, isCached := s.cache.pl2.Retrieve(filepath)
	if isCached {
		return cachedData.(*pl2.PL2), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	paletteTransform, err := pl2.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing pl2 %q: %v", filepath, err)
	}

	if err = s.cache.pl2.Insert(filepath, paletteTransform, len(data)); err != nil {
		s.logger.Error("caching file", "error", err)
	}

	s.logger.Info("loaded PL2 file", "path", filepath)

	return paletteTransform, nil
}

func (s *Service) UnmarshalTsv(filepath string, destination any) error {
	stream, err := s.file.Load(filepath)
	if err != nil {
		return fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("reading data %q: %v", filepath, err)
	}

	if err = tsv.Unmarshal([]byte(strings.ReplaceAll(string(data), "\"", "")), destination); err != nil {
		return fmt.Errorf("parsing data: %v", err)
	}

	return nil
}

func (s *Service) LoadTsv(filepath string) ([]byte, error) {
	cacheKey := fmt.Sprintf("tsv: %s %s", filepath)

	cacheData, isCached := s.cache.tsv.Retrieve(cacheKey)
	if isCached {
		return cacheData.([]byte), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	if err = s.cache.tsv.Insert(cacheKey, data, len(data)); err != nil {
		s.logger.Warn("caching data", "error", err)
	}

	s.logger.Info("loaded TSV", "path", filepath)

	return data, nil
}

func (s *Service) LoadTbl(filepath string) (tbl.TextTable, error) {
	cachedData, isCached := s.cache.tbl.Retrieve(filepath)
	if isCached {
		return *cachedData.(*tbl.TextTable), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	table, err := tbl.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("parsing TBL %q: %v", filepath, err)
	}

	if err = s.cache.tbl.Insert(filepath, table, len(data)); err != nil {
		s.logger.Error("caching file", "error", err)
	}

	s.logger.Info("loaded", "path", filepath)

	return table, nil
}

func (s *Service) LoadWav(filepath string) ([]byte, error) {
	for s.cache.wav == nil {
		time.Sleep(time.Second)
	}

	cachedData, isCached := s.cache.wav.Retrieve(filepath)
	if isCached {
		return cachedData.([]byte), nil
	}

	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	audioData, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data %q: %v", filepath, err)
	}

	if err = s.cache.wav.Insert(filepath, audioData, len(audioData)); err != nil {
		s.logger.Error("caching file", "error", err)
	}

	s.logger.Info("loaded", "path", filepath)

	return audioData, nil
}

func (s *Service) LoadCOF(filepath string) (*cof.COF, error) {
	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data", "error", err)
	}

	cofInstance, err := cof.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("parsing cof", "error", err)
	}

	s.logger.Info("loaded COF", "path", filepath)

	return cofInstance, nil
}

func (s *Service) LoadAnimationData(filepath string) (*cof.AnimationData, error) {
	stream, err := s.file.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("loading file %q: %v", filepath, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("reading data", "error", err)
	}

	animData, err := cof.Load(data)
	if err != nil {
		return nil, fmt.Errorf("parsing anim data", "error", err)
	}

	s.logger.Info("loaded Animation Data", "path", filepath)

	return animData, nil
}

func (s *Service) ExtractPaletteFromPl2(pathPL2 string) (color.Palette, error) {
	paletteStream, err := s.file.Load(pathPL2)
	if err != nil {
		return nil, fmt.Errorf("loading pl2", "error", err)
	}

	const (
		numColors    = 256
		numBytesRGBA = numColors * 4
		opaque       = 255
	)

	paletteData := make([]byte, numBytesRGBA)
	numRead, err := paletteStream.Read(paletteData)
	if err != nil {
		return nil, fmt.Errorf("reading from PL2 stream", "error", err)
	} else if numRead != numBytesRGBA {
		return nil, fmt.Errorf("couldn't read all palette bytes")
	}

	p := make(color.Palette, numColors)
	for idx := range p {
		if idx == 0 {
			p[idx] = image.Transparent

			continue
		}

		p[idx] = color.RGBA{
			R: paletteData[(idx*4)+0],
			G: paletteData[(idx*4)+1],
			B: paletteData[(idx*4)+2],
			A: opaque,
		}
	}

	return p, nil
}
