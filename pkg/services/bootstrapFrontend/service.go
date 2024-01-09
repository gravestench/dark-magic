package bootstrap_backend

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/cofLoader"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
	"github.com/gravestench/dark-magic/pkg/services/fontTableLoader"
	"github.com/gravestench/dark-magic/pkg/services/hero"
	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/lua"
	"github.com/gravestench/dark-magic/pkg/services/mapGenerator"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
	"github.com/gravestench/dark-magic/pkg/services/wavLoader"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
	"github.com/gravestench/dark-magic/pkg/services/webServer"
)

type Service struct {
	logger *slog.Logger
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	const (
		projectConfigDir = "~/.config/dark-magic"
	)

	// utility services
	//mesh.Add(&modalTui.Service{})
	//mesh.Add(&goscript.Service{}) // WIP
	mesh.Add(&lua.Service{})
	mesh.Add(&cacheManager.Service{})
	mesh.Add(&fileWatcher.Service{})
	mesh.Add(&configFile.Service{RootDirectory: projectConfigDir})
	mesh.Add(&webServer.Service{})
	mesh.Add(&webRouter.Service{})

	// file/record loaders
	mesh.Add(&fontTableLoader.Service{})
	mesh.Add(&dc6Loader.Service{})
	mesh.Add(&dccLoader.Service{})
	mesh.Add(&ds1Loader.Service{})
	mesh.Add(&dt1Loader.Service{})
	mesh.Add(&pl2Loader.Service{})
	mesh.Add(&tblLoader.Service{})
	mesh.Add(&tsvLoader.Service{})
	mesh.Add(&wavLoader.Service{})
	mesh.Add(&cofLoader.Service{})
	mesh.Add(&mpqLoader.Service{})

	// these all use the loaders and records
	mesh.Add(&assetLoader.Service{})
	mesh.Add(&recordManager.Service{})
	mesh.Add(&spriteManager.Service{})
	mesh.Add(&locale.Service{})
	mesh.Add(&mapGenerator.Service{})
	mesh.Add(&hero.Service{})
}

func (s *Service) Name() string {
	return "Backend Bootstrap"
}

func (s *Service) Ready() bool {
	return true
}

// the following methods are boilerplate, but they are used
// by the servicemesh to enforce a standard logging format.

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
