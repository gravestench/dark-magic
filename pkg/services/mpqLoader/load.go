package mpqLoader

import (
	"path/filepath"
	"time"
)

func (s *Service) loadArchivesFromFiles() {
	g := s.config.Group(s.Name())

	rootDir, err := expandHomeDirAlias(g.GetString("directory"))
	if err != nil {
		s.logger.Error("formatting root dir", "error", err)
		panic(err)
	}

	for _, fileName := range g.GetStrings("load order") {
		absPath := filepath.Join(rootDir, fileName)

		if err = s.AddArchive(absPath); err != nil {
			s.logger.Warn("adding MPQ", "error", err)
			continue
		}

		s.logger.Info("loaded MPQ", "path", absPath)
	}

	if len(s.archives) == 0 {
		time.Sleep(time.Second * 3)
		s.logger.Error("no MPQ files found")
		s.logger.Error("edit your config file: %s", filepath.Join(s.configManager.ConfigDirectory(), s.ConfigFileName()))
		s.mesh.Shutdown()
	}
}
