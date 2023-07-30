package mpq_loader

import (
	"path/filepath"
	"time"
)

func (s *Service) loadArchivesFromFiles() {
	cfg, err := s.Config()
	if err != nil {
		s.logger.Fatal().Msgf("getting config: %v", err)
	}

	g := cfg.Group(s.Name())

	rootDir, err := expandHomeDirAlias(g.GetString("directory"))
	if err != nil {
		s.logger.Fatal().Msgf("formatting root dir: %v", err)
	}

	for _, fileName := range g.GetStrings("load order") {
		absPath := filepath.Join(rootDir, fileName)

		if err = s.AddArchive(absPath); err != nil {
			s.logger.Warn().Msgf("adding MPQ: %v", err)
			continue
		}

		s.logger.Info().Msgf("loaded MPQ: %v", absPath)
	}

	if len(s.archives) == 0 {
		time.Sleep(time.Second * 3)
		s.logger.Error().Msg("no MPQ files found")
		s.logger.Fatal().Msgf("edit your config file: %s", filepath.Join(s.cfgManager.ConfigDirectory(), s.ConfigFilePath()))
	}
}
