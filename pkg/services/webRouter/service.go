package webRouter

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/webRouter/middleware/static_assets"
)

type Service struct {
	log        *slog.Logger
	cfgManager configFile.Dependency

	root *gin.Engine

	boundServices map[string]*struct{} // holds 0-size entries

	config struct {
		debug bool
	}

	reloadDebounce time.Time
}

func (s *Service) SetLogger(l *slog.Logger) {
	s.log = l
}

func (s *Service) Logger() *slog.Logger {
	return s.log
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	gin.SetMode("release")
	mesh.Add(&static_assets.Middleware{})
	s.root = gin.New()
	go s.beginDynamicRouteBinding(mesh)
}

func (s *Service) Name() string {
	return "Web Router"
}
