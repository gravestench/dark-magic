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

	//modalTui        modalTui.Service
	//goscript        goscript.Service
	lua             lua.Service
	cacheManager    cacheManager.Service
	fileWatcher     fileWatcher.Service
	configFile      configFile.Service
	webServer       webServer.Service
	webRouter       webRouter.Service
	fontTableLoader fontTableLoader.Service
	dc6Loader       dc6Loader.Service
	dccLoader       dccLoader.Service
	ds1Loader       ds1Loader.Service
	dt1Loader       dt1Loader.Service
	pl2Loader       pl2Loader.Service
	tblLoader       tblLoader.Service
	tsvLoader       tsvLoader.Service
	wavLoader       wavLoader.Service
	cofLoader       cofLoader.Service
	mpqLoader       mpqLoader.Service
	assetLoader     assetLoader.Service
	recordManager   recordManager.Service
	spriteManager   spriteManager.Service
	locale          locale.Service
	mapGenerator    mapGenerator.Service
	hero            hero.Service
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	const (
		projectConfigDir = "~/.config/dark-magic"
	)

	s.configFile.RootDirectory = projectConfigDir

	// utility services
	//mesh.Add(&s.modalTui)
	//mesh.Add(goscript.&s.ServicP
	mesh.Add(&s.lua)
	mesh.Add(&s.cacheManager)
	mesh.Add(&s.fileWatcher)
	mesh.Add(&s.configFile)
	mesh.Add(&s.webServer)
	mesh.Add(&s.webRouter)

	// file/record loaders
	mesh.Add(&s.fontTableLoader)
	mesh.Add(&s.dc6Loader)
	mesh.Add(&s.dccLoader)
	mesh.Add(&s.ds1Loader)
	mesh.Add(&s.dt1Loader)
	mesh.Add(&s.pl2Loader)
	mesh.Add(&s.tblLoader)
	mesh.Add(&s.tsvLoader)
	mesh.Add(&s.wavLoader)
	mesh.Add(&s.cofLoader)
	mesh.Add(&s.mpqLoader)

	// these all use the loaders and records
	mesh.Add(&s.assetLoader)
	mesh.Add(&s.recordManager)
	mesh.Add(&s.spriteManager)
	mesh.Add(&s.locale)
	mesh.Add(&s.mapGenerator)
	mesh.Add(&s.hero)
}

func (s *Service) Name() string {
	return "Backend Bootstrap"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) BackendReady() bool {
	//if !s.modalTui.Ready() {
	//	return false
	//}
	//
	//if !s.goscript.Ready() {
	//	return false
	//}

	if !s.lua.Ready() {
		return false
	}

	if !s.cacheManager.Ready() {
		return false
	}

	if !s.fileWatcher.Ready() {
		return false
	}

	if !s.configFile.Ready() {
		return false
	}

	if !s.webServer.Ready() {
		return false
	}

	if !s.webRouter.Ready() {
		return false
	}

	if !s.fontTableLoader.Ready() {
		return false
	}

	if !s.dc6Loader.Ready() {
		return false
	}

	if !s.dccLoader.Ready() {
		return false
	}

	if !s.ds1Loader.Ready() {
		return false
	}

	if !s.dt1Loader.Ready() {
		return false
	}

	if !s.pl2Loader.Ready() {
		return false
	}

	if !s.tblLoader.Ready() {
		return false
	}

	if !s.tsvLoader.Ready() {
		return false
	}

	if !s.wavLoader.Ready() {
		return false
	}

	if !s.cofLoader.Ready() {
		return false
	}

	if !s.mpqLoader.Ready() {
		return false
	}

	if !s.assetLoader.Ready() {
		return false
	}

	if !s.recordManager.Ready() {
		return false
	}

	if !s.spriteManager.Ready() {
		return false
	}

	if !s.locale.Ready() {
		return false
	}

	if !s.mapGenerator.Ready() {
		return false
	}

	if !s.hero.Ready() {
		return false
	}

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
