package gameScene

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

type Config struct {
	Enabled bool     `json:"enabled"`
	Source  string   `json:"source"`
	Map     string   `json:"map"`
	Tiles   []string `json:"tiles"`
	Palette string   `json:"palette"`
}

func (s *Service) ConfigFileName() string { return "game_scene.json" }

func (s *Service) DefaultConfigData() []byte {
	config := Config{
		Enabled: true,
		Source:  "$MPQ_DIRECTORY/d2data.mpq",
		Map:     "data/global/tiles/Act1/BARRACKS/barE.ds1",
		Tiles: []string{
			"data/global/tiles/Act1/BARRACKS/floor.dt1",
			"data/global/tiles/Act1/BARRACKS/basewall.dt1",
			"data/global/tiles/Act1/BARRACKS/barset.dt1",
		},
		Palette: "data/global/palette/ACT1/pal.pl2",
	}
	data, _ := json.MarshalIndent(config, "", "\t")
	return data
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("reading scene config: %w", err)
	}
	if err := json.Unmarshal(data, &s.Config); err != nil {
		return fmt.Errorf("decoding scene config: %w", err)
	}
	return nil
}
