package gameScene

import (
	"encoding/json"
)

type Config struct {
	Enabled bool     `json:"enabled"`
	Source  string   `json:"source"`
	Map     string   `json:"map"`
	Tiles   []string `json:"tiles"`
	Palette string   `json:"palette"`
}

func (s *Service) DefaultConfigData() []byte {
	config := DefaultConfig()
	data, _ := json.MarshalIndent(config, "", "\t")
	return data
}

// DefaultConfig returns the production scene defaults.
func DefaultConfig() Config {
	return Config{
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
}
