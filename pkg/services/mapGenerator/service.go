package mapGenerator

import (
	"fmt"
	"strings"

	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

type Service struct {
	logger *zerolog.Logger

	dt1     dt1Loader.Dependency
	ds1     ds1Loader.Dependency
	records recordManager.Dependency

	seed uint64
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
	return "WorldMap Generator"
}

func (s *Service) Seed() uint64 {
	return s.seed
}

func (s *Service) SetSeed(u uint64) {
	s.seed = u
}

func (s *Service) GenerateMap(act, difficulty uint) (*WorldMap, error) {
	m := NewWorldMap()

	if err := s.loadLevelTypeRecordsToWorldMap(act, m); err != nil {
		return nil, fmt.Errorf("loading level type records into map: %v", err)
	}

	if err := s.loadLevelPresetRecordsToWorldMap(m); err != nil {
		return nil, fmt.Errorf("loading level type records into map: %v", err)
	}

	return m, nil
}

func (s *Service) loadLevelTypeRecordsToWorldMap(act uint, m *WorldMap) error {
	for _, r := range s.records.LevelType() {
		if r.Act != act {
			continue
		}

		lvl := Level{}
		lvl.Name = r.Name
		lvl.Type = r

		for _, filePath := range []string{
			r.File1, r.File2, r.File3, r.File4,
			r.File5, r.File6, r.File7, r.File8,
			r.File9, r.File10, r.File11, r.File12,
			r.File13, r.File14, r.File15, r.File16,
			r.File17, r.File18, r.File19, r.File20,
			r.File21, r.File22, r.File23, r.File24,
			r.File25, r.File26, r.File27, r.File28,
			r.File29, r.File30, r.File31, r.File32,
		} {
			if filePath == "" {
				continue
			}

			if filePath == "0" {
				continue
			}

			if strings.HasSuffix(strings.ToLower(filePath), ".dt1") {
				filePath = fmt.Sprintf("data/global/tiles/%s", filePath)

				tileset, err := s.dt1.Load(filePath)
				if err != nil {
					continue
					//return fmt.Errorf("loading dt1 specified in level type record: %v", err)
				}

				lvl.TileSets = append(lvl.TileSets, *tileset)
			}

			if strings.HasSuffix(strings.ToLower(filePath), ".ds1") {
				filePath = fmt.Sprintf("data/global/tiles/%s", filePath)

				tileStamp, err := s.ds1.Load(filePath)
				if err != nil {
					continue
					//return fmt.Errorf("loading dt1 specified in level type record: %v", err)
				}

				lvl.TileStamps = append(lvl.TileStamps, *tileStamp)
			}
		}

		m.Levels[r.Act] = append(m.Levels[r.Act], lvl)
	}

	return nil
}

func (s *Service) loadLevelPresetRecordsToWorldMap(m *WorldMap) error {
	for _, levels := range m.Levels {
		for _, level := range levels {
			for _, record := range s.records.LevelPresets() {
				if level.Name == record.Name {
					level.Preset = record
					break
				}
			}

			for _, filePath := range []string{
				level.Preset.File1,
				level.Preset.File2,
				level.Preset.File3,
				level.Preset.File4,
				level.Preset.File5,
				level.Preset.File6,
			} {
				if !strings.HasSuffix(filePath, ".ds1") {
					continue
				}

				filePath = fmt.Sprintf("data/global/tiles/%s", filePath)

				stamp, err := s.ds1.Load(filePath)
				if err != nil {
					s.logger.Error().Msgf("loading ds1 %q for level %q: %v", filePath, level.Name, err)
					continue
				}

				level.TileStamps = append(level.TileStamps, *stamp)
			}
		}
	}

	return nil
}
